/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/btnmasher/util"
	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/conc"

	"github.com/btnmasher/dircd/shared/concurrentmap"
	"github.com/btnmasher/dircd/shared/itempool"
)

// keepAliveTimeout sets the connection timeout duration on the client IRC connections.
const keepAliveTimeout = 2 * time.Minute

// messagePoolMax sets the message pool buffer length
const messagePoolMax = 1000

// bufferPoolMax sets the bytes.Buffer pool length
const bufferPoolMax = 1000

// Reference to the global Message object pool.
var msgPool = itempool.New(messagePoolMax, NewMessage)

// Reference to the global bytes.Buffer object pool.
var bufPool = util.NewBufferPool(bufferPoolMax)

// Server holds the state of an IRC server instance.
type Server struct {
	mu  sync.Mutex
	rwm sync.RWMutex

	logger       *logrus.Entry
	logLevel     logrus.Level
	logFormatter logrus.Formatter
	listener     net.Listener

	// Configuration
	listenAddr *net.TCPAddr
	hostname   string
	motd       string
	welcome    string
	support    concurrentmap.ConcurrentMap[string, string]
	tlsConfig  *tls.Config

	// Active State
	Users    UserMap
	Nicks    UserMap
	Channels ChanMap

	listeners       map[*net.Listener]struct{}
	activeConn      map[*Conn]struct{}
	onShutdown      []func()
	inShutdown      atomic.Bool // true when server is in shutdown
	listenerGroup   sync.WaitGroup
	connectionGroup conc.WaitGroup
}

type ServerOption interface {
	apply(*Server) error
}

type option func(*Server) error

func (o option) apply(s *Server) error {
	return o(s)
}

// NewServer initializes and returns a new instance of a Server.
func NewServer(options ...ServerOption) (*Server, error) {
	server := &Server{
		logLevel: logrus.InfoLevel,
		Users:    concurrentmap.New[string, *User](),
		Nicks:    concurrentmap.New[string, *User](),
		Channels: concurrentmap.New[string, *Channel](),
		support:  concurrentmap.New[string, string](),
	}

	var optionErrors error
	for i := range options {
		err := options[i].apply(server)
		if err != nil {
			optionErrors = errors.Join(optionErrors, err)
		}
	}

	if optionErrors != nil {
		return nil, optionErrors
	}

	if server.logger == nil {
		_ = WithDefaultLogger().apply(server)
	}

	if server.logFormatter != nil {
		server.logger.Logger.SetFormatter(server.logFormatter)
	}

	if server.logLevel != logrus.InfoLevel {
		server.logger.Logger.SetLevel(server.logLevel)
	}

	return server, nil
}

func (srv *Server) warmup() {
	logger := srv.logger.WithField("operation", "warmup")
	logger.Info("registering message handlers")
	registerHandlers()

	logger.Info("populating ISupport")
	srv.populateISupport()

	logger.Info("warming up message pool")
	msgPool.Warmup(messagePoolMax)
}

// Network returns the configured network name of the server in a concurrent safe manner
func (srv *Server) Network() string {
	val, ok := srv.support.Get("network")
	if ok {
		return srv.Hostname()
	}
	return val
}

func WithDefaultLogger() ServerOption {
	return option(func(s *Server) error {
		logger := logrus.New()
		logger.SetFormatter(&nested.Formatter{
			HideKeys:      true,
			FieldsOrder:   []string{"component", "category", "operation", "handler"},
			ShowFullLevel: true,
		})
		logger.SetReportCaller(true)
		s.logger = logger.WithField("component", "irc-server")
		return nil
	})
}

func WithLogLevel(level logrus.Level) ServerOption {
	return option(func(s *Server) error {
		s.logLevel = level
		return nil
	})
}

func WithLogFormatter(formatter logrus.Formatter) ServerOption {
	return option(func(s *Server) error {
		s.logFormatter = formatter
		return nil
	})
}

func WithDefaultLogFormatter() ServerOption {
	return option(func(s *Server) error {
		s.logFormatter = &nested.Formatter{
			HideKeys:      true,
			FieldsOrder:   []string{"component", "category", "operation", "handler"},
			ShowFullLevel: true,
			CallerFirst:   true,
		}
		return nil
	})
}

// WithLogger sets the configured logger for the server.
// If logger is nil, then the default logger will be used at level INFO
func WithLogger(logger *logrus.Logger) ServerOption {
	return option(func(s *Server) error {
		if logger != nil {
			s.logger = logger.WithField("component", "irc-server")
		}
		return nil
	})
}

// WithNetwork sets the configured network name of the server
func WithNetwork(network string) ServerOption {
	return option(func(s *Server) error {
		s.support.Set("network", network)
		return nil
	})
}

// Address returns the configured address string of the server
func (srv *Server) Address() string {
	srv.rwm.RLock()
	defer srv.rwm.RUnlock()

	if srv.listenAddr == nil {
		return ""
	}
	return srv.listenAddr.String()
}

// WithAddress sets the configured address of the server
func WithAddress(address string) ServerOption {
	return option(func(s *Server) error {
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			return err
		}
		s.listenAddr = addr
		return nil
	})
}

// Hostname returns the configured hostname of the server in a concurrency safe manner
func (srv *Server) Hostname() string {
	srv.rwm.RLock()
	defer srv.rwm.RUnlock()

	if len(srv.hostname) == 0 {
		return fmt.Sprint(srv.listenAddr.IP)
	}
	return srv.hostname
}

// WithHostname sets the configured hostname of the server
func WithHostname(host string) ServerOption {
	return option(func(s *Server) error {
		s.hostname = host
		return nil
	})
}

// WithMOTD sets the configured MOTD of the server
func WithMOTD(motd string) ServerOption {
	return option(func(s *Server) error {
		s.motd = motd
		return nil
	})
}

// MOTD returns the configured MOTD of the server in a currency safe manner
func (srv *Server) MOTD() string {
	srv.rwm.RLock()
	defer srv.rwm.RUnlock()

	if len(srv.motd) == 0 {
		return "Server has no message of the day set."
	}
	return srv.motd
}

// Welcome returns the configured welcome message of the server in a concurrency safe manner
func (srv *Server) Welcome() string {
	srv.rwm.RLock()
	defer srv.rwm.RUnlock()

	if len(srv.welcome) == 0 {
		return "Server has no welcome message set."
	}
	return srv.welcome
}

// WithWelcome sets the configured welcome message of the server
func WithWelcome(msg string) ServerOption {
	return option(func(s *Server) error {
		s.welcome = msg
		return nil
	})
}

func WithGracefulShutdown(ctx context.Context, shutdownTimeout time.Duration) ServerOption {
	return option(func(s *Server) error {
		go func() {
			<-ctx.Done()

			shutdownCtx, shutdown := context.WithTimeout(context.Background(), shutdownTimeout)
			defer shutdown()

			start := time.Now()
			s.logger.Infof("gracefully shutting down server within the next %v", shutdownTimeout)
			if err := s.Shutdown(shutdownCtx); err != nil {
				s.logger.Error(fmt.Errorf("failed to gracefully shutdown server: %w", err))
			} else {
				s.logger.Info("server has initiated termination of all connections successfully")
			}

			diff := time.Now().Sub(start)
			if diff < shutdownTimeout { // still time to wait for connections to flush
				if !waitTimeout(&s.connectionGroup, shutdownTimeout-diff) {
					s.logger.Info("goodbye! <3")
					return
				}
			}

			s.logger.Info("connection termination exceeded graceful shutdown timeout, force closing connections")
			_ = s.Close()
		}()
		return nil
	})
}

func waitTimeout(wg *conc.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

// ISupport returns a slice of formatted ISupport key=value pairs.
func (srv *Server) ISupport() []string {
	support := make([]string, 0, srv.support.Length())
	var buffer bytes.Buffer

	srv.support.ForEach(func(config, setting string) error {
		buffer.WriteString(strings.ToUpper(config))

		if len(setting) > 0 {
			buffer.WriteString("=")
			buffer.WriteString(setting)
		}

		support = append(support, buffer.String())
		buffer.Reset()
		return nil
	})

	return support
}

func (srv *Server) populateISupport() {
	srv.support.Set("chanmodes", "bhoOv,p,LMT,AacEeFHIimNnPqRrstV")
	srv.support.Set("prefix", "(Oohv)~@%+")
	srv.support.Set("maxpara", fmt.Sprint(MaxMsgParams))
	srv.support.Set("modes", fmt.Sprint(MaxModeChange))
	srv.support.Set("chanlimit", fmt.Sprintf("#!:%v", MaxJoinedChans))
	srv.support.Set("nicklen", fmt.Sprint(MaxNickLength))
	srv.support.Set("maxlist", fmt.Sprintf("bhov:%v,O:1", MaxListItems))
	srv.support.Set("casemapping", "ascii")
	srv.support.Set("topiclen", fmt.Sprint(MaxTopicLength))
	srv.support.Set("kicklen", fmt.Sprint(MaxKickLength))
	srv.support.Set("chanlen", fmt.Sprint(MaxChanLength))
	srv.support.Set("awaylen", fmt.Sprint(MaxAwayLength))
}

func (srv *Server) shuttingDown() bool {
	return srv.inShutdown.Load()
}

// Close immediately closes all active net.Listeners and any open connections.
//
// Returns any error returned from closing the Server's underlying Listener(s).
func (srv *Server) Close() error {
	srv.inShutdown.Store(true)
	srv.mu.Lock()
	defer srv.mu.Unlock()
	err := srv.closeListenersLocked()

	// Unlock srv.mu while waiting for listenerGroup.
	// The group Add and Done calls are made with srv.mu held,
	// to avoid adding a new listener in the window between
	// us setting inShutdown above and waiting here.
	srv.mu.Unlock()
	srv.listenerGroup.Wait()
	srv.mu.Lock()

	for c := range srv.activeConn {
		c.cancel(errors.New("server forcibly closed"))
		delete(srv.activeConn, c)
	}
	return err
}

func (srv *Server) closeListenersLocked() error {
	var err error
	for ln := range srv.listeners {
		if closeErr := (*ln).Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

// shutdownPollIntervalMax is the max polling interval when checking
// quiescence during Server.Shutdown. Polling starts with a small
// interval and backs off to the max.
const shutdownPollIntervalMax = 500 * time.Millisecond

// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing all open
// listeners, then closing all connections, and then waiting
// indefinitely for connections to shut down.
// If the provided context expires before the shutdown is complete,
// Shutdown returns the context's error, otherwise it returns any
// error returned from closing the Server's underlying Listener(s).
//
// When Shutdown is called, Serve, ListenAndServe, and
// ListenAndServeTLS immediately return ErrServerClosed. Make sure the
// program doesn't exit and waits instead for Shutdown to return.
//
// Once Shutdown has been called on a server, it may not be reused;
// future calls to methods such as Serve will return ErrServerClosed.
func (srv *Server) Shutdown(ctx context.Context) error {
	srv.inShutdown.Store(true)

	srv.mu.Lock()
	listenErr := srv.closeListenersLocked()
	for _, f := range srv.onShutdown {
		go f()
	}
	srv.mu.Unlock()
	srv.listenerGroup.Wait()

	pollIntervalBase := time.Millisecond
	nextPollInterval := func() time.Duration {
		// Add 10% jitter.
		interval := pollIntervalBase + time.Duration(rand.Intn(int(pollIntervalBase/10)))
		// Double and clamp for next time.
		pollIntervalBase *= 2
		if pollIntervalBase > shutdownPollIntervalMax {
			pollIntervalBase = shutdownPollIntervalMax
		}
		return interval
	}

	timer := time.NewTimer(nextPollInterval())
	defer timer.Stop()
	for {
		if srv.closeConns() {
			return listenErr
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(nextPollInterval())
		}
	}
}

// closeConns closes all idle connections and reports whether the
// server is quiescent.
func (srv *Server) closeConns() bool {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	quiescent := true
	for conn := range srv.activeConn {
		state, unixSec := conn.getState()
		if state&(StateNew|StateHandshake) != 0 || unixSec == 0 {
			// Assume unixSec == 0 means it's a very new
			// connection, without state set yet.
			quiescent = false
			continue
		}
		conn.shutdown()
		delete(srv.activeConn, conn)
	}
	return quiescent
}

// trackListener adds or removes a net.Listener to the set of tracked
// listeners.
//
// We store a pointer to interface in the map set, in case the
// net.Listener is not comparable. This is safe because we only call
// trackListener via Serve and can track+defer untrack the same
// pointer to local variable there. We never need to compare a
// Listener from another caller.
//
// It reports whether the server is still up (not Shutdown or Closed).
func (srv *Server) trackListener(ln *net.Listener, add bool) bool {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.listeners == nil {
		srv.listeners = make(map[*net.Listener]struct{})
	}
	if add {
		if srv.shuttingDown() {
			return false
		}
		srv.listeners[ln] = struct{}{}
		srv.listenerGroup.Add(1)
	} else {
		delete(srv.listeners, ln)
		srv.listenerGroup.Done()
	}
	return true
}

func (srv *Server) trackConn(conn *Conn, add bool) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.activeConn == nil {
		srv.activeConn = make(map[*Conn]struct{})
	}
	if add {
		srv.activeConn[conn] = struct{}{}
	} else {
		delete(srv.activeConn, conn)
	}
}

// ListenAndServe listens on the TCP network address srv.ListenAddr and
// then calls Serve to handle the irc.Conn sessions.
// Accepted connections are configured to enable TCP keep-alives.
//
// If srv.ListenAddr is blank, ":6667" is used.
//
// ListenAndServe always returns a non-nil error.
func (srv *Server) ListenAndServe() error {
	srv.warmup()
	logger := srv.logger.WithField("category", "listener")
	if srv.listenAddr == nil {
		addr, addrErr := net.ResolveTCPAddr("tcp", "localhost:6697")
		if addrErr != nil {
			return errors.Join(addrErr, errors.New("error attempting to use fallback default address"))
		}
		logger.Infof("no address/port port specified, defaulting to [%v]", addr)
		srv.listenAddr = addr
	}

	listener, listenErr := net.ListenTCP("tcp", srv.listenAddr)
	if listenErr != nil {
		return errors.Join(listenErr, errors.New("error attempting to create TCP listener"))
	}

	servErr := srv.serve(&tcpKeepAliveListener{*listener})

	logger.Debug("waiting for connections to terminate")
	srv.connectionGroup.Wait()
	return servErr
}

// ListenAndServeTLS listens on the TCP network address srv.Addr and
// then calls Serve to handle the irc.Conn sessions on a TLS connection.
// Accepted connections are configured to enable TCP keep-alive.
//
// Filenames containing a certificate and matching private key for the
// server must be provided if neither the Server's TLSConfig.Certificates
// nor TLSConfig.GetCertificate are populated. If the certificate is
// signed by a certificate authority, the certFile should be the
// concatenation of the server's certificate, any intermediates, and
// the CA's certificate.
//
// If the address of the Server is not configured, "localhost:6697" is used.
//
// ListenAndServeTLS always returns a non-nil error.
func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {
	srv.warmup()
	logger := srv.logger.WithField("category", "listener")
	if srv.listenAddr == nil {
		addr, addrErr := net.ResolveTCPAddr("tcp", "localhost:6697")
		if addrErr != nil {
			return errors.Join(addrErr, errors.New("error attempting to use fallback default address"))
		}
		logger.Infof("no address/port port specified, defaulting to %v", addr)
	}

	config := srv.cloneTLSConfig()

	if len(config.Certificates) == 0 && certFile != "" && keyFile != "" {
		cert, certErr := tls.LoadX509KeyPair(certFile, keyFile)
		if certErr != nil {
			return errors.Join(certErr, errors.New("error attempting to load TLS key pair"))
		}
		config.Certificates = append(config.Certificates, cert)
	}

	listener, err := net.ListenTCP("tcp", srv.listenAddr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(&tcpKeepAliveListener{*listener}, config)
	servErr := srv.serve(tlsListener)

	logger.Debug("waiting for connections to terminate")
	srv.connectionGroup.Wait()
	return servErr
}

// Serve starts an IRC server which listens for connections on the given
// net.Listener, accepts them when they arrive, then assigns them to a new
// instance of irc.Conn
func (srv *Server) Serve(listen net.Listener) error {
	servErr := srv.serve(listen)
	srv.logger.WithField("category", "listener").Debug("waiting for connections to terminate")
	srv.connectionGroup.Wait()
	return servErr
}

const initialAcceptRetryDelay = 5 * time.Millisecond
const maxAcceptRetryDelay = 1 * time.Second

// onceCloseListener wraps a net.Listener, protecting it from
// multiple Close calls.
type onceCloseListener struct {
	net.Listener
	once     sync.Once
	closeErr error
}

func (oc *onceCloseListener) close() {
	oc.closeErr = oc.Listener.Close()
}

func (oc *onceCloseListener) Close() error {
	oc.once.Do(oc.close)
	return oc.closeErr
}

func (srv *Server) serve(listen net.Listener) error {
	logger := srv.logger.WithFields(logrus.Fields{"category": "listener"})

	listen = &onceCloseListener{Listener: listen}
	defer func() {
		closeErr := listen.Close()
		if closeErr != nil {
			logger.Error(fmt.Errorf("error encountered while closing listener: %w", closeErr))
		}
	}()

	if !srv.trackListener(&listen, true) {
		return ErrServerClosed
	}
	defer srv.trackListener(&listen, false)

	logger.Printf("starting IRC server listener at local address [%s]", listen.Addr())

	var retryDelay time.Duration // how long to sleep on accept failure
	for {
		logger.Debug("listening for connection...")
		sock, acceptErr := listen.Accept()
		if srv.shuttingDown() {
			logger.Debug("server shutting down, listener closed")
		} else {
			logger.Debug("accepting connection...")
		}

		if acceptErr != nil {
			if srv.shuttingDown() {
				return ErrServerClosed
			}

			var netErr net.Error
			if errors.As(acceptErr, &netErr) && netErr.Temporary() {
				if retryDelay == 0 {
					retryDelay = initialAcceptRetryDelay
				} else {
					retryDelay *= 2
				}

				if retryDelay > maxAcceptRetryDelay {
					retryDelay = maxAcceptRetryDelay
				}

				logger.Error(fmt.Errorf("error accepting connection, retrying in %v: %w", retryDelay, netErr))
				time.Sleep(retryDelay)
				continue
			}

			logger.Debug(fmt.Errorf("listener encountered unretryable error: %w", acceptErr))
			return acceptErr
		}

		logger.Debug("accepted connection")

		retryDelay = 0
		conn := NewConn(context.Background(), srv, sock, srv.logger)
		srv.connectionGroup.Go(func() { serve(conn) })
	}
}

// Returns a shallow clone of the specified tls config. If cfg is nil, a new default tls.Config is returned.
func (srv *Server) cloneTLSConfig() *tls.Config {
	if srv.tlsConfig == nil {
		return &tls.Config{}
	}

	return srv.tlsConfig.Clone()
}

// debugServerConnections controls whether all server connections are wrapped
// with a verbose logging wrapper.
// const debugServerConnections = false

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	net.TCPListener
}

func (listen *tcpKeepAliveListener) Accept() (net.Conn, error) {
	conn, err := listen.AcceptTCP()
	if err != nil {
		return nil, err
	}
	_ = conn.SetKeepAlive(true)
	_ = conn.SetKeepAlivePeriod(keepAliveTimeout)
	return conn, nil
}

func (srv *Server) ValidateName(name string) (error, uint16) {
	// TODO: restricted/reserved nicknames
	switch {
	case name == "":
		return ErrNoNickGiven, ReplyNoNicknameGiven
	case len(name) > MaxNickLength:
		fallthrough
	case strings.HasPrefix(name, "#"):
		fallthrough
	case strings.HasPrefix(name, ":"):
		fallthrough
	case strings.Contains(name, SPACE):
		return ErrErroneousNickname, ReplyErroneusNickname
	case srv.Nicks.Exists(name):
		return ErrNickInUse, ReplyNicknameInUse
	}
	return nil, ReplyNone
}
