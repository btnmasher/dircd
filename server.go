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
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/btnmasher/util"
	"github.com/sirupsen/logrus"
)

// KeepAliveTimeout sets the connection timeout duration on the client IRC connections.
const KeepAliveTimeout time.Duration = 2 * time.Minute

// WriteTimeout sets the write timeout duration on the client IRC connections.
const WriteTimeout time.Duration = 5 * time.Second

// PingTimeout sets the PING/PONG timeout duration on the client IRC connections.
const PingTimeout time.Duration = 30 * time.Second

// MessagePoolMax sets the message pool buffer length
const MessagePoolMax = 1000

// BufferPoolMax sets the bytes.Buffer pool length
const BufferPoolMax = 1000

// WriteQueueLength sets the length of each connections write queue channel.
const WriteQueueLength = 10

// msgpool holds a reference to the global Message object pool.
var msgpool = NewMessagePool(MessagePoolMax)

// bufpool holds a reference to the global bytes.Buffer object pool.
var bufpool = util.NewBufferPool(BufferPoolMax)

var log *logrus.Logger

// Server holds the state of an IRC server instance.
type Server struct {
	sync.RWMutex

	// Configuration related stuff
	listenAddr string
	hostname   string
	motd       string
	welcome    string
	support    *util.ConcurrentMapString

	// Active State
	Users     *UserMap
	Nicks     *UserMap
	Conns     *ConnMap
	Channels  *ChanMap
	TLSConfig *tls.Config

	listener net.Listener
}

// Warmup initializes the irc library for use.
func Warmup(logger *logrus.Logger) {
	log = logger
	log.Info("irc: Registering message handlers")
	registerHandlers()

	log.Info("irc: Warming up message pool")
	msgpool.Warmup(MessagePoolMax)

}

// NewServer initializes and returns a new instance of a Server.
func NewServer() *Server {
	server := &Server{
		Users:    NewUserMap(),
		Nicks:    NewUserMap(),
		Conns:    NewConnMap(),
		Channels: NewChanMap(),
		support:  util.NewConcurrentMapString(),
	}
	server.setISupport()
	return server
}

// Network returns the configured network name of the server in a
// concurrency safe manner.
func (server *Server) Network() string {
	val, err := server.support.Get("network")
	if err != nil {
		return server.Hostname()
	}
	return val
}

// SetNetwork sets the configured network name of the server in a
// concurrency safe manner.
func (server *Server) SetNetwork(new string) {
	if server.support.Set("network", new) != nil {
		log.Error("irc: Could not set server parameter: network")
	}
}

// Address returns the configured address of the server in a
// concurrency safe manner.
func (server *Server) Address() string {
	server.RLock()
	defer server.RUnlock()

	if len(server.listenAddr) < 1 {
		if server.listener != nil {
			return server.listener.Addr().String()
		}
		return ""
	}
	return server.listenAddr
}

// SetAddress sets the configured address of the server in a
// concurrency safe manner.
func (server *Server) SetAddress(addr string) {
	server.Lock()
	defer server.Unlock()

	server.listenAddr = addr
}

// Hostname returns the configured hostname of the server in a
// concurrency safe manner.
func (server *Server) Hostname() string {
	server.RLock()
	defer server.RUnlock()

	if len(server.hostname) < 1 {
		return server.listener.Addr().String()
	}
	return server.hostname
}

// SetHostname sets the configured hostname of the server in a
// concurrency safe manner.
func (server *Server) SetHostname(host string) {
	server.Lock()
	defer server.Unlock()

	server.hostname = host
}

// MOTD returns the configured MOTD of the server in a
// concurrency safe manner.
func (server *Server) MOTD() string {
	server.RLock()
	defer server.RUnlock()

	if len(server.motd) < 1 {
		return "Server has no MOTD message set."
	}
	return server.motd
}

// SetMOTD sets the configured MOTD of the server in a
// concurrency safe manner.
func (server *Server) SetMOTD(motd string) {
	server.Lock()
	defer server.Unlock()

	server.listenAddr = motd
}

// Welcome returns the configured welcome message of the server in a
// concurrency safe manner.
func (server *Server) Welcome() string {
	server.RLock()
	defer server.RUnlock()

	if len(server.welcome) < 1 {
		return "Server has no welcome message set."
	}
	return server.welcome
}

// SetWelcome sets the configured welcome message of the server in a
// concurrency safe manner.
func (server *Server) SetWelcome(msg string) {
	server.Lock()
	defer server.Unlock()

	server.welcome = msg
}

// ISupport returns a slice of formatted ISupport key=value pairs.
func (server *Server) ISupport() []string {
	support := make([]string, server.support.Length())
	index := 0
	var buffer bytes.Buffer

	server.support.ForEach(func(config, setting string) {
		buffer.WriteString(strings.ToUpper(config))

		if len(setting) > 0 {
			buffer.WriteString("=")
			buffer.WriteString(setting)
		}

		support[index] = buffer.String()
		buffer.Reset()
		index++
	})

	return support
}

func (server *Server) setISupport() {
	server.support.Add("chanmodes", "bhoOv,p,LMT,AacEeFHIimNnPqRrstV")
	server.support.Add("prefix", "(Oohv)~@%+")
	server.support.Add("maxpara", fmt.Sprint(MaxMsgParams))
	server.support.Add("modes", fmt.Sprint(MaxModeChange))
	server.support.Add("chanlimit", fmt.Sprintf("#!:%v", MaxJoinedChans))
	server.support.Add("nicklen", fmt.Sprint(MaxNickLength))
	server.support.Add("maxlist", fmt.Sprintf("bhov:%v,O:1", MaxListItems))
	server.support.Add("casemapping", "ascii")
	server.support.Add("topiclen", fmt.Sprint(MaxTopicLength))
	server.support.Add("kicklen", fmt.Sprint(MaxKickLength))
	server.support.Add("chanlen", fmt.Sprint(MaxChanLength))
	server.support.Add("awaylen", fmt.Sprint(MaxAwayLength))
}

// ListenAndServe listens on the TCP network address srv.ListenAddr and
// then calls Serve to handle the irc.Conn sessions.
// Accepted connections are configured to enable TCP keep-alives.
//
// If srv.ListenAddr is blank, ":6667" is used.
//
// ListenAndServe always returns a non-nil error.
func (server *Server) ListenAndServe() error {
	addr := server.Address()
	if addr == "" {
		addr = ":6667"
	}

	listen, err := net.Listen("tcp4", addr)

	if err != nil {
		return err
	}

	return server.Serve(tcpKeepAliveListener{listen.(*net.TCPListener)})
}

// ListenAndServeTLS listens on the TCP network address srv.Addr and
// then calls Serve to handle the irc.Conn sessions on a TLS connection.
// Accepted connections are configured to enable TCP keep-alives.
//
// Filenames containing a certificate and matching private key for the
// server must be provided if neither the Server's TLSConfig.Certificates
// nor TLSConfig.GetCertificate are populated. If the certificate is
// signed by a certificate authority, the certFile should be the
// concatenation of the server's certificate, any intermediates, and
// the CA's certificate.
//
// If srv.ListenAddr is blank, ":6697" is used.
//
// ListenAndServeTLS always returns a non-nil error.
func (server *Server) ListenAndServeTLS(certFile, keyFile string) error {
	addr := server.Address()
	if addr == "" {
		addr = ":6697"
	}

	config := cloneTLSConfig(server.TLSConfig)

	configHasCert := len(config.Certificates) > 0 || config.GetCertificate != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
	}

	listen, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(tcpKeepAliveListener{listen.(*net.TCPListener)}, config)
	return server.Serve(tlsListener)
}

// Serve starts an IRC server which listens for connections on the given
// net.Listener, accepts them when they arrive, then assigns them to a new
// instance of irc.Conn
func (server *Server) Serve(listen net.Listener) error {
	defer listen.Close()

	log.Printf("irc: Starting IRC server listener at local address [%s]", listen.Addr())

	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		log.Debug("irc: Listening for connection...")
		sock, err := listen.Accept()
		log.Debug("irc: Accepting connection...")

		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}

				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				log.Errorf("irc: Error accepting connection: %v; retrying in %vms", err, tempDelay.Nanoseconds()/int64(time.Millisecond))
				time.Sleep(tempDelay)
				continue
			}

			return err
		}

		log.Debug("irc: Accepted connection.")

		tempDelay = 0
		conn := NewConn(server, sock)
		go serve(conn)
	}
}

// cloneTLSConfig returns a shallow clone of the exported
// fields of cfg, ignoring the unexported sync.Once, which
// contains a mutex and must not be copied.
//
// The cfg must not be in active use by tls.Server, or else
// there can still be a race with tls.Server updating SessionTicketKey
// and our copying it, and also a race with the server setting
// SessionTicketsDisabled=false on failure to set the random
// ticket key.
//
// If cfg is nil, a new zero tls.Config is returned.
func cloneTLSConfig(cfg *tls.Config) *tls.Config {
	if cfg == nil {
		return &tls.Config{}
	}
	return &tls.Config{
		Rand:                     cfg.Rand,
		Time:                     cfg.Time,
		Certificates:             cfg.Certificates,
		NameToCertificate:        cfg.NameToCertificate,
		GetCertificate:           cfg.GetCertificate,
		RootCAs:                  cfg.RootCAs,
		NextProtos:               cfg.NextProtos,
		ServerName:               cfg.ServerName,
		ClientAuth:               cfg.ClientAuth,
		ClientCAs:                cfg.ClientCAs,
		InsecureSkipVerify:       cfg.InsecureSkipVerify,
		CipherSuites:             cfg.CipherSuites,
		PreferServerCipherSuites: cfg.PreferServerCipherSuites,
		SessionTicketsDisabled:   cfg.SessionTicketsDisabled,
		SessionTicketKey:         cfg.SessionTicketKey,
		ClientSessionCache:       cfg.ClientSessionCache,
		MinVersion:               cfg.MinVersion,
		MaxVersion:               cfg.MaxVersion,
		CurvePreferences:         cfg.CurvePreferences,
	}
}

// debugServerConnections controls whether all server connections are wrapped
// with a verbose logging wrapper.
// const debugServerConnections = false

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (listen tcpKeepAliveListener) Accept() (net.Conn, error) {
	conn, err := listen.AcceptTCP()
	if err != nil {
		return nil, err
	}
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(KeepAliveTimeout)
	return conn, nil
}
