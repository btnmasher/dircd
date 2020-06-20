/*
   Copyright (c) 2020, btnmasher
   All rights reserved.

   Redistribution and use in source and binary forms, with or without modification, are permitted provided that
   the following conditions are met:

   1. Redistributions of source code must retain the above copyright notice, this list of conditions and the
      following disclaimer.

   2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and
      the following disclaimer in the documentation and/or other materials provided with the distribution.

   3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or
      promote products derived from this software without specific prior written permission.

   THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED
   WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
   PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
   ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
   TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
   HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
   NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
   POSSIBILITY OF SUCH DAMAGE.
*/

package dircd

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/btnmasher/random"
)

// Conn represents the server side of an IRC connection.
type Conn struct {
	sync.RWMutex

	// server is the server on which the connection arrived.
	// Immutable; never nil.
	server *Server

	// rwc is the underlying network connection.
	// This is never wrapped by other types and is the value given out
	// to CloseNotifier callers. It is usually of type *net.TCPConn or
	// *tls.Conn.
	sock net.Conn

	// remAddr is sock.RemoteAddr().String(). It is not populated synchronously
	// inside the Listener's Accept goroutine, as some implementations block.
	// It is populated immediately inside the (*Conn).serve goroutine.
	remAddr string

	user          *User
	channels      *ChanMap
	capabilities  *Capabilities
	capRequested  bool
	capNegotiated bool

	incoming *bufio.Scanner
	outgoing *bufio.Writer

	writeQueue chan *bytes.Buffer

	heartbeat *time.Timer

	lastPingSent string
	lastPingRecv string

	kill chan bool

	timeoutForced bool
	registered    bool
}

// NewConn initializes a new instance of Conn
func NewConn(srv *Server, sck net.Conn) *Conn {
	conn := &Conn{
		server:     srv,
		sock:       sck,
		heartbeat:  time.NewTimer(PingTimeout),
		channels:   NewChanMap(),
		incoming:   bufio.NewScanner(sck),
		outgoing:   bufio.NewWriter(sck),
		writeQueue: make(chan *bytes.Buffer, WriteQueueLength),
		kill:       make(chan bool, 5),
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

func serve(conn *Conn) {
	defer conn.cleanup()
	conn.start()

	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Errorf("irc: Panic serving %v: %v\n%s", conn.remAddr, err, buf)
			conn.doQuit("Server Error.")
		}

		conn.sock.Close()
	}()

	if tlsConn, ok := conn.sock.(*tls.Conn); ok {
		conn.setDeadlines()

		if err := tlsConn.Handshake(); err != nil {
			log.Errorf("irc: TLS handshake error from [%s]: %s", conn.remAddr, err)
			return
		}
	}

	go conn.writeLoop() // Runs until conn.kill channel is signaled
	conn.readLoop()     // Blocks until error
	log.Debugf("irc: readLoop() exited for [%s]", conn.remAddr)
}

func (conn *Conn) start() {
	conn.Lock()
	defer conn.Unlock()

	//This can block until the address is acquired, so just wait.
	conn.remAddr = conn.sock.RemoteAddr().String()

	log.Debugf("irc: Got new connection remote address: [%s]", conn.remAddr)

	//Add self to server connections map now that we have the address to index by.
	conn.server.Conns.Add(conn.remAddr, conn)
}

func (conn *Conn) readLoop() {
	for {
		conn.setReadDeadline()

		if !conn.incoming.Scan() { // Will block here until there is a read or a timeout.
			defer func() { conn.kill <- true }()

			if err := conn.incoming.Err(); err != nil {
				if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
					if !conn.timeoutForced {
						log.Infof("irc: Connection timed out for [%s]", conn.remAddr)
						conn.doQuit("Connection timeout.")
					}
				} else {
					log.Error(err)
				}
			}

			log.Debugf("irc: Closing socket for [%s]", conn.remAddr)

			if err := conn.sock.Close(); err != nil {
				log.Errorf("irc: Socket error when trying to close socket from [%s]: %s", conn.remAddr, err)
			}

			return
		}

		data := conn.incoming.Text()
		log.Infof("irc: [%s]->[SERVER]: %s", conn.remAddr, data)
		msg, err := Parse(data)
		//log.Debugf("[%s]->[SERVER]: %s", conn.remAddr, pretty.Sprint(msg))

		if err != nil {
			log.Errorf("irc: Error parsing message from client [%s]: %s", conn.remAddr, err)
			return
		}

		conn.heartbeat.Reset(PingTimeout)

		RouteCommand(conn, msg)
	}
}

func (conn *Conn) writeLoop() {
	for {
		select {
		case <-conn.kill:
			log.Debug("irc: conn.kill signal received in writeLoop(), closing goroutine.")
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
		log.Errorf("irc: Error rendering message to buffer for [%s]: Message too long.", conn.remAddr)
		return
	}

	conn.writeQueue <- buffer // Hand message context over to the writeloop goroutine here.
}

func (conn *Conn) write(buffer *bytes.Buffer) {
	defer func() {
		bufpool.Recycle(buffer)
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Errorf("irc: Panic in write socket operation for [%s]: %v\n%s", conn.remAddr, err, buf)

			conn.doQuit("Socket Error.")
		}
	}()

	conn.setWriteDeadline()

	if _, err := conn.outgoing.Write(buffer.Bytes()); err != nil {
		log.Errorf("irc: Error writing to socket for [%s]: %s", conn.remAddr, err)
		conn.doQuit("Socket Error.")
		return
	}

	if err := conn.outgoing.Flush(); err != nil {
		log.Errorf("irc: Error writing to socket [%s]: %s", conn.remAddr, err)
		conn.doQuit("Socket Error.")
		return
	}

	log.Infof("irc: [SERVER]->[%s]: %s", conn.remAddr, strings.TrimSpace(buffer.String()))
}

func (conn *Conn) doHeartbeat() {
	conn.Lock()
	defer conn.Unlock()

	if conn.lastPingRecv != conn.lastPingSent {
		conn.heartbeat.Stop()
		log.Debugf("irc: PING timeout for [%s]: last sent: %s, last received: %s", conn.remAddr, conn.lastPingSent, conn.lastPingRecv)
		conn.doQuit("Connection timeout.")
		return
	}

	str := random.String(10)
	msg := msgpool.New()
	msg.Command = CmdPing
	msg.Text = str
	conn.lastPingSent = str
	conn.heartbeat.Reset(PingTimeout)
	conn.Write(msg.RenderBuffer())
}

func (conn *Conn) doQuit(reason string) {
	if conn.channels.Length() > 0 {
		msg := msgpool.New()
		msg.Sender = conn.user.Hostmask()
		msg.Command = CmdQuit
		msg.Text = reason

		if len(reason) < 1 {
			msg.Text = "Client issued QUIT command."
		}

		conn.channels.ForEach(func(channel *Channel) {
			channel.Nicks.Del(conn.user.Nick())
			channel.Send(msg, "")
		})
	}

	conn.kill <- true
}

func (conn *Conn) regiserUser() {
	conn.Lock()
	defer conn.Unlock()
	conn.registered = true
	conn.server.Users.Add(strings.ToLower(conn.user.Name()), conn.user)
	conn.server.Nicks.Add(strings.ToLower(conn.user.Nick()), conn.user)
}

func (conn *Conn) cleanup() {
	conn.server.Users.Del(strings.ToLower(conn.user.Name()))
	conn.server.Nicks.Del(strings.ToLower(conn.user.Nick()))
	conn.server.Conns.Del(conn.remAddr)
}

func (conn *Conn) setWriteDeadline() {
	if WriteTimeout != 0 {
		conn.sock.SetWriteDeadline(time.Now().Add(WriteTimeout))
	}
}

func (conn *Conn) setReadDeadline() {
	if KeepAliveTimeout != 0 {
		conn.sock.SetReadDeadline(time.Now().Add(KeepAliveTimeout))
	}
}

func (conn *Conn) forceTimeout() {
	conn.Lock()
	defer conn.Unlock()
	conn.timeoutForced = true
	conn.sock.SetReadDeadline(time.Now().Add(time.Microsecond))
}

func (conn *Conn) setDeadlines() {
	conn.setReadDeadline()
	conn.setWriteDeadline()
}

func (conn *Conn) newMessage() *Message {
	msg := msgpool.New()

	msg.Sender = conn.server.Hostname()

	return msg
}
