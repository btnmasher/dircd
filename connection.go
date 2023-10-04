/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/conc/panics"

	"github.com/btnmasher/dircd/shared/concurrentmap"

	"github.com/btnmasher/random"
)

// Conn represents the server side of an IRC connection.
type Conn struct {
	mu            sync.RWMutex
	logger        *logrus.Entry
	ctx           context.Context
	cancel        context.CancelCauseFunc
	curState      atomic.Uint64
	registered    atomic.Bool
	shuttingDown  atomic.Bool
	timeoutForced atomic.Bool

	// server is the server on which the connection arrived.
	// Immutable; never nil.
	server *Server

	hostname string

	// rwc is the underlying network connection.
	// This is never wrapped by other types and is the value given out
	// to CloseNotifier callers. It is usually of type *net.TCPConn or
	// *tls.Conn.
	sock net.Conn

	// remAddr is sock.RemoteAddr().String(). It is not populated synchronously
	// inside the Listener's Accept goroutine, as some implementations block.
	// It is populated immediately inside the (*Conn).serve goroutine.
	remAddr string

	user     *User
	channels ChanMap

	capabilities  int
	capRequested  bool
	capNegotiated bool

	incoming *bufio.Scanner
	outgoing *bufio.Writer

	writeQueue chan *bytes.Buffer

	heartbeat *time.Timer

	lastPingSent string
	lastPingRecv string
}

// pingTimeout sets the PING/PONG timeout duration on the client IRC connections.
const pingTimeout = 30 * time.Second

// writeQueueLength sets the length of each connection's write queue channel.
const writeQueueLength = 10

// NewConn initializes a new instance of Conn
func NewConn(ctx context.Context, srv *Server, sck net.Conn, logger *logrus.Entry) *Conn {
	connCtx, cancel := context.WithCancelCause(ctx)
	conn := &Conn{
		ctx:        connCtx,
		cancel:     cancel,
		logger:     logger.WithField("component", "connection"),
		server:     srv,
		sock:       sck,
		hostname:   srv.Hostname(),
		heartbeat:  time.NewTimer(pingTimeout),
		channels:   concurrentmap.New[string, *Channel](),
		incoming:   bufio.NewScanner(sck),
		outgoing:   bufio.NewWriter(sck),
		writeQueue: make(chan *bytes.Buffer, writeQueueLength),
	}
	conn.user = &User{
		conn: conn,
	}
	// TODO: implement test hooks/debug like stdlib?
	// if debugServerConnections {
	// 	c.sock = newLoggingConn("server", c.sock)
	// }
	return conn
}

type ConnState int

const (
	StateNew ConnState = 1 << iota
	StateHandshake
	StateConnected
	StateClosed
)

func (conn *Conn) setState(state ConnState) {
	srv := conn.server
	switch state {
	case StateNew:
		srv.trackConn(conn, true)
	case StateClosed:
		srv.trackConn(conn, false)
	}
	if state > 0xff || state < 0 {
		panic("internal error")
	}
	packedState := uint64(time.Now().Unix()<<8) | uint64(state)
	conn.curState.Store(packedState)
}

func (conn *Conn) getState() (state ConnState, unixSec int64) {
	packedState := conn.curState.Load()
	return ConnState(packedState & 0xff), int64(packedState >> 8)
}

func serve(conn *Conn) {
	logger := conn.logger.WithField("category", "serve")
	defer conn.cleanup()
	conn.start()
	logger.Info("client connection established")

	recovered := panics.Try(func() {
		if tlsConn, ok := conn.sock.(*tls.Conn); ok {
			conn.setDeadlines()

			conn.setState(StateHandshake)
			if tlsErr := tlsConn.Handshake(); tlsErr != nil {
				logger.Error(fmt.Errorf("error occurred during TLS handshake: %w", tlsErr))
				return
			}
		}
		conn.setState(StateConnected)

		logger.Debug("starting reade/write routines")
		go conn.writeLoop() // Runs until context is cancelled or socket error occurs
		conn.readLoop()     // Blocks until error
		cause := context.Cause(conn.ctx)
		if cause != nil {
			logger.Debugf("connection context cancelled with cause: %s", cause.Error())
		}
	})

	logger.Info("client connection terminated")

	panicErr := recovered.AsError()
	if panicErr != nil {
		logger.Errorf(panicLogTemplate, recovered.String())
		conn.cancel(fmt.Errorf("panic occurred during serve: %w", panicErr))
		conn.doQuit("Socket Error.")
	}
}

func (conn *Conn) start() {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	//This can block until the address is acquired, so just wait.
	conn.remAddr = conn.sock.RemoteAddr().String()

	conn.setState(StateNew)
	conn.logger.Debugf("new connection from remote address: [%s]", conn.remAddr)
	conn.logger = conn.logger.WithField("address", conn.remAddr)
}

func (conn *Conn) shutdown() {
	conn.doQuit("Server shutting down.")
}

func (conn *Conn) isForceTimedOut() bool {
	return conn.timeoutForced.Load()
}

func (conn *Conn) isShuttingDown() bool {
	return conn.shuttingDown.Load()
}

func (conn *Conn) isRegistered() bool {
	return conn.registered.Load()
}

func (conn *Conn) isConnected() bool {
	state, _ := conn.getState()
	return state&StateConnected != 0
}

func (conn *Conn) isClosed() bool {
	state, _ := conn.getState()
	return state&StateClosed != 0
}

func (conn *Conn) readLoop() {
	logger := conn.logger.WithField("category", "reader")
	defer logger.Debug("reader terminated")
	for {
		select {
		case <-conn.ctx.Done():
			logger.Debug("connection context cancelled, closing reader")
			conn.forceTimeout()
			return

		default:
			conn.setReadDeadline()

			if !conn.incoming.Scan() { // Will block here until there is a read or a timeout.
				// scan failed, either an error or a timeout
				func() {
					if scanErr := conn.incoming.Err(); scanErr != nil {
						var netErr net.Error
						if errors.As(scanErr, &netErr) && netErr.Timeout() {
							if !conn.isForceTimedOut() {
								logger.Error(fmt.Errorf("connection timed out: %w", netErr))
								conn.doQuit("Connection timeout.")
							}
							if !conn.isShuttingDown() && !conn.isClosed() {
								conn.doQuit("Connection terminated.")
							}
						} else {
							defer conn.cancel(fmt.Errorf("connection line reader terminated: %w", conn.incoming.Err()))
						}
					}

					logger.Debug("line reader completed, closing connection")
					conn.setState(StateClosed)
				}()
				return
			}

			data := conn.incoming.Text()
			logger.Infof("received: [%s]", data)

			msg, parseErr := Parse(data)
			if parseErr != nil {
				msgPool.Recycle(msg)
				logger.Warn(fmt.Errorf("error parsing message from client: %w", parseErr))
				continue
			}

			conn.heartbeat.Reset(pingTimeout)

			conn.server.Router.RouteCommand(conn, msg)
		}
	}
}

func (conn *Conn) writeLoop() {
	logger := conn.logger.WithField("category", "writer")
	defer logger.Debug("writer terminated")
	for {
		select {
		case <-conn.ctx.Done():
			conn.setState(StateClosed)
			logger.Debug("connection context cancelled, closing writer")
			conn.forceTimeout()
			return

		case buf := <-conn.writeQueue:
			conn.write(buf)

		case <-conn.heartbeat.C:
			conn.doHeartbeat()
		}
	}
}

func (conn *Conn) Write(buffer *bytes.Buffer) {
	if buffer.Len() > MaxMsgLength {
		conn.logger.Error("error rendering message to buffer, message too long")
		return
	}

	if !conn.isConnected() {
		conn.logger.Error("attempted write on unready connection")
		return
	}

	conn.writeQueue <- buffer // Hand message context over to the write loop goroutine here.
}

func (conn *Conn) write(buffer *bytes.Buffer) {
	defer bufPool.Recycle(buffer)
	logger := conn.logger.WithField("category", "writer")

	recovered := panics.Try(func() {
		conn.setWriteDeadline()

		if _, writeErr := conn.outgoing.Write(buffer.Bytes()); writeErr != nil {
			conn.setState(StateClosed)
			logger.Error(fmt.Errorf("error writing to socket %w", writeErr))
			conn.doQuit("Socket Error.")
			return
		}

		if flushErr := conn.outgoing.Flush(); flushErr != nil {
			conn.setState(StateClosed)
			logger.Error(fmt.Errorf("error writing to socket: %w", flushErr))
			conn.doQuit("Socket Error.")
			return
		}

		logger.Infof("sent: [%s]", strings.TrimSpace(buffer.String()))
	})

	panicErr := recovered.AsError()
	if panicErr != nil {
		logger.Errorf(panicLogTemplate, recovered.String())
		conn.setState(StateClosed)
		conn.cancel(fmt.Errorf("panic occurred during write: %w", panicErr))
		conn.doQuit("Socket Error.")
	}
}

func (conn *Conn) doHeartbeat() {
	logger := conn.logger.WithField("category", "heartbeat")
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.lastPingRecv != conn.lastPingSent {
		conn.heartbeat.Stop()
		logger.Debugf("PING timeout: last sent: %s, last received: %s", conn.lastPingSent, conn.lastPingRecv)
		conn.cancel(errors.New("heartbeat timeout"))
		conn.doQuit("Connection timeout.")
		return
	}

	str := random.String(10)
	msg := msgPool.New()
	msg.Command = CmdPing
	msg.Trailing = str
	conn.lastPingSent = str
	conn.heartbeat.Reset(pingTimeout)
	conn.Write(msg.RenderBuffer())
}

func (conn *Conn) doChatMessage(msg *Message) {
	if !enoughParams(msg, 1) || len(msg.Trailing) == 0 {
		conn.ReplyNeedMoreParams(msg.Command)
		return
	}

	// TODO: Send Message permission check

	targetUser, userExists := conn.server.Nicks.Get(strings.ToLower(msg.Params[0]))
	targetChannel, chanExists := conn.server.Channels.Get(strings.ToLower(msg.Params[0]))

	if !userExists && !chanExists {
		conn.logger.WithField("category", "chat message").Debug("did not find target")
		conn.ReplyNoSuchNick(msg.Params[0])
		return
	}

	msg.Params = msg.Params[0:1] // Strip erroneous parameters.
	msg.Source = conn.user.Hostmask()

	if targetUser != nil {
		targetUser.conn.Write(msg.RenderBuffer())
	} else {
		targetChannel.Send(msg, conn.user.Nick())
	}
}

func (conn *Conn) doKill(reason, source string) {
	logger := conn.logger.WithField("operation", "quit")
	if source == "" {
		source = conn.server.Hostname()
	}

	if len(reason) == 0 {
		reason = "Server killed."
	}

	if !conn.isClosed() {
		reply := conn.newMessage()
		reply.Command = CmdError
		reply.Trailing = fmt.Sprintf("Closing link: %s [Killed: [%s [%s]]]", conn.server.Hostname(), source, reason)
		conn.Write(reply.RenderBuffer())
	}

	if conn.channels.Length() > 0 && conn.isRegistered() {
		logger.Debug("quitting user from joined channels")
		msg := msgPool.New()
		msg.Source = conn.user.Hostmask()
		msg.Command = CmdQuit
		msg.Trailing = fmt.Sprintf("Killed: [%s [%s]]", source, reason)
		nick := conn.user.Nick()
		chanErr := conn.channels.ForEach(func(name string, channel *Channel) error {
			return channel.RemoveUser(nick, msg)
		})
		conn.channels.Clear()
		logger.Debug("user channels cleared")
		if chanErr != nil {
			errs := chanErr.(interface{ Unwrap() []error }).Unwrap()
			for i := range errs {
				logger.Error(fmt.Errorf("encountered error quitting user from channel: %w", errs[i]))
			}
		}
	}

	conn.cancel(fmt.Errorf("kill called with reason: %s", reason))
}

func (conn *Conn) doQuit(reason string) {
	if len(reason) == 0 {
		reason = "Client issued QUIT command."
	}

	logger := conn.logger.WithField("operation", "quit")
	logger.Debugf("quit called with reason: %s", reason)

	if !conn.isClosed() {
		reply := conn.newMessage()
		reply.Command = CmdError
		reply.Trailing = fmt.Sprintf("Closing link: %s [Quit: %s]", conn.user.Hostmask(), reason)
		conn.write(reply.RenderBuffer())
		conn.shuttingDown.Store(true)
	}

	if conn.channels.Length() > 0 && conn.isRegistered() {
		msg := msgPool.New()
		msg.Source = conn.user.Hostmask()
		msg.Command = CmdQuit
		msg.Trailing = reason

		nick := conn.user.Nick()
		chanErr := conn.channels.ForEach(func(name string, channel *Channel) error {
			return channel.RemoveUser(nick, msg)
		})
		conn.channels.Clear()
		if chanErr != nil {
			errs := chanErr.(interface{ Unwrap() []error }).Unwrap()
			for i := range errs {
				logger.Error(fmt.Errorf("encountered error quitting user from channel: %w", errs[i]))
			}
		}
	}

	conn.cancel(fmt.Errorf("quit called with reason: %s", reason))
}

func (conn *Conn) registerUser() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.registered.Store(true)
	name := conn.user.Name()
	nick := conn.user.Nick()
	conn.server.Users.Set(name, conn.user)
	conn.server.Nicks.Set(nick, conn.user)
	conn.logger.Debugf("registered user: %s - %s", name, nick)
}

func (conn *Conn) cleanup() {
	defer func() {
		closeErr := conn.sock.Close()
		if closeErr != nil {
			conn.logger.Error(fmt.Errorf("error occurred closing socket: %w", closeErr))
		}
	}()
	conn.setState(StateClosed)
	conn.logger.Debug("cleaning up connection state from server")
	name := conn.user.Name()
	nick := conn.user.Nick()
	conn.server.Users.Delete(name)
	conn.server.Nicks.Delete(nick)
	conn.logger.Debugf("cleaned up user: %s - %s", name, nick)
}

// writeTimeout sets the write timeout duration on the client IRC connections.
const writeTimeout = 5 * time.Second

func (conn *Conn) setWriteDeadline() {
	_ = conn.sock.SetWriteDeadline(time.Now().Add(writeTimeout))
}

func (conn *Conn) setReadDeadline() {
	_ = conn.sock.SetReadDeadline(time.Now().Add(keepAliveTimeout))
}

func (conn *Conn) forceTimeout() {
	conn.timeoutForced.Store(true)
	_ = conn.sock.SetReadDeadline(time.Now().Add(time.Microsecond))
}

func (conn *Conn) setDeadlines() {
	conn.setReadDeadline()
	conn.setWriteDeadline()
}

func (conn *Conn) newMessage() *Message {
	msg := msgPool.New()

	msg.Source = conn.hostname

	return msg
}
