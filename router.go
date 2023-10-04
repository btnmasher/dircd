package dircd

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

type MessageContext struct {
	Conn    *Conn
	Msg     *Message
	handler string
	handled bool
	abort   bool
	err     error
}

// Handled signals to the router to not call the next MessageHandler in the chain if applicable
func (c *MessageContext) Handled() {
	c.handled = true
}

// AbortWithError signals to the router to not call the next MessageHandler in the chain
// if applicable, and to log the error reported
func (c *MessageContext) AbortWithError(err error) {
	c.abort = true
	c.err = err
}

// MessageHandler defines the function signature of a handler used to process IRC messages.
type MessageHandler func(*MessageContext)

// IRouter defines all router handle interface includes single and group router.
type IRouter interface {
	IRoutes
	Group(...MessageHandler) *RouterGroup
}

// IRoutes defines all router handle interface.
type IRoutes interface {
	Use(...MessageHandler) IRoutes
	Handle(string, ...MessageHandler) IRoutes
}

// HandlersChain defines a HandlerFunc slice.
type HandlersChain []MessageHandler

// Last returns the last handler in the chain. i.e. the last handler is the main one.
func (c HandlersChain) Last() MessageHandler {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

type Router struct {
	logger *logrus.Entry
	RouterGroup
	HandlerMap map[string]HandlersChain
}

func NewRouter(logger *logrus.Entry) *Router {
	if logger == nil {
		panic("must provide a logger to NewRouter")
	}

	log := logger.WithField("sub-component", "router")
	r := &Router{
		logger:     log,
		HandlerMap: make(map[string]HandlersChain),
	}
	r.root = true
	r.router = r
	return r
}
func (router *Router) addHandler(command string, handlers HandlersChain) {
	if command == "" {
		panic("command must not be an empty string")
	}

	if len(handlers) == 0 {
		panic("there must be at least one handler")
	}

	if _, exists := router.HandlerMap[command]; exists {
		panic(fmt.Sprintf("handler(s) already registered for command: %s", command))
	}

	router.HandlerMap[command] = handlers
}

// Use attaches a global middleware to the router. i.e. the middleware attached through Use() will be
// included in the handlers chain for every single command.
// For example, this is the right place for a logger or error management middleware.
func (router *Router) Use(middleware ...MessageHandler) IRoutes {
	router.RouterGroup.Use(middleware...)
	return router
}

// Handle registers a new request handle and middleware with the given name and name.
// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
func (router *Router) Handle(command string, handlers ...MessageHandler) IRoutes {
	handlers = router.combineHandlers(handlers)
	router.router.addHandler(command, handlers)
	return router.returnRouter()
}

// HandlerInfo represents a request route's specification which contains the command and its handler.
type HandlerInfo struct {
	Command  string
	Handlers []string
}

// HandlersInfo defines a HandlerInfo slice.
type HandlersInfo []HandlerInfo

// RouterGroup is used internally to configure router, a RouterGroup is associated with
// a GroupCondition and an array of handlers (middleware).
type RouterGroup struct {
	root     bool
	router   *Router
	Handlers HandlersChain
}

func (group *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(group.Handlers) + len(handlers)
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

// Handle registers a new request handle and middleware with the given name and name.
// The last handler should be the real handler, the other ones should be middleware that can
// and should be shared among different routes.
func (group *RouterGroup) Handle(command string, handlers ...MessageHandler) IRoutes {
	handlers = group.combineHandlers(handlers)
	group.router.addHandler(command, handlers)
	return group.returnRouter()
}

// Use adds middleware to the group
func (group *RouterGroup) Use(middleware ...MessageHandler) IRoutes {
	group.Handlers = append(group.Handlers, middleware...)
	return group.returnRouter()
}

func (group *RouterGroup) returnRouter() IRouter {
	if group.root {
		return group.router
	}
	return group
}

// Group creates a new router group. You should add all the routes that have common middlewares.
// For example, all the routes that use a common middleware for authorization could be grouped.
func (group *RouterGroup) Group(handlers ...MessageHandler) *RouterGroup {
	if len(handlers) == 0 {
		panic("a group must have at least one handler")
	}

	newGroup := &RouterGroup{
		Handlers: group.combineHandlers(handlers),
		router:   group.router,
	}

	return newGroup
}

// Handlers returns a slice of registered routes, including some useful information, such as:
// the http name, name and the handler name.
func (router *Router) Handlers() HandlersInfo {
	info := make(HandlersInfo, 0, len(router.HandlerMap))
	for command, handlers := range router.HandlerMap {
		info = append(info, HandlerInfo{
			Command:  command,
			Handlers: getHandlerChain(handlers),
		})
	}
	return info
}

func (router *Router) PrintHandlers() {
	logger := router.logger.WithField("sub-component", "Router")
	logger.Debug("Registered Handlers:")
	handlers := router.Handlers()
	chains := make([]string, 0)
	for i := range handlers {
		if len(handlers[i].Handlers) > 1 {
			chains = append(chains, fmt.Sprintf("| Command: %s \tHandlers: %s", handlers[i].Command, strings.Join(handlers[i].Handlers, " -> ")))
			continue
		}
		router.logger.Debugf("| Command: %s \tHandler: %s", handlers[i].Command, handlers[i].Handlers[0])
	}

	for i := range chains {
		router.logger.Debug(chains[i])
	}
}

func getHandlerChain(handlers HandlersChain) []string {
	chain := make([]string, 0, len(handlers))
	for i := range handlers {
		chain = append(chain, nameOfFunction(handlers[i]))
	}
	return chain
}

func enoughParams(msg *Message, expected int) bool {
	return !(len(msg.Params) < expected)
}

func nameOfFunction(f any) string {
	return path.Base(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name())
}

// RouteCommand accepts an IRC message and routes it to a function
// in which is designed to process the command.
func (router *Router) RouteCommand(conn *Conn, msg *Message) {
	defer msgPool.Recycle(msg)
	log := router.logger.WithField("command", msg.Command)
	handlers, exists := router.HandlerMap[msg.Command]
	if !exists {
		conn.ReplyNotImplemented(msg.Command)
		defer msgPool.Recycle(msg)
		log.Warnf("command not implemented encountered for: %s", msg.Command)
		return
	}

	ctx := &MessageContext{Conn: conn, Msg: msg}

	for i := range handlers {
		ctx.handler = nameOfFunction(handlers[i])
		handlers[i](ctx)
		if ctx.handled {
			return
		}
		if ctx.err != nil {
			log.Warn(fmt.Errorf("error encounterd handling command with handler [%s]: %w", ctx.handler, ctx.err))
		}
		if ctx.abort && len(handlers) > 1 {
			log.Debugf("command handler chain aborted at: %s", ctx.handler)
			return
		}
	}
}
