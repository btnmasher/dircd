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
	"strings"
)

// Handlers is a map of functions where the handlers are stored.
var Handlers = make(map[string]MessageHandler)

// MessageHandler defines the function signature of a handler used to
// process IRC messages.
type MessageHandler func(*Conn, *Message)

// All of command handler functions do not return an error. Instead it
// must process all error conditions relating to the command and reply
// to the user in the correct way specified by RFC2812.

// HandleQuit processes a QUIT command.
//
// The connection will be scheduled for immediate deadline, and the
// server will broadcast the QUIT message to all channels the user is
// joined to.
//
//    Command: QUIT
//    Parameters: :<reason>
func HandleQuit(conn *Conn, msg *Message) {
	defer msgpool.Recycle(msg)
	conn.doQuit(msg.Text)
}

// HandleNick processes a NICK command.
//
// First, it checks if the current nickname is in use by the user issuing
// the command; by another user on the server; or disallowed by the server
// configuration. Then it checks the validity of the nickname formatting
// before finally, if all of the requirements are met, sets the User object
// Nick field to the specified name in the command parameters.
//
//    Command: NICK
//    Parameters: <nickname>
func HandleNick(conn *Conn, msg *Message) {
	defer msgpool.Recycle(msg)
	ok := true

	if !enoughParams(msg, 1) {
		conn.ReplyNoNicknameGiven()
		return
	}

	reply := conn.newMessage()
	defer msgpool.Recycle(reply)

	reply.Code = ReplyNicknameInUse

	if conn.user.Nick() == msg.Params[0] {
		reply.Text = ErrNickAlreadySet.String()
		ok = false
	}

	if ok && conn.server.Nicks.Exists(msg.Params[0]) {
		reply.Text = ErrNickInUse.String()
		ok = false
	}

	// TODO: Nick restriction check

	// TODO: Nick formatting checks
	// This ties into configurations such as:
	// - nick length
	// - reserved nicks

	if ok { // Nick formatting check stub
		conn.user.SetNick(msg.Params[0])
		reply.Code = ReplyNone
		reply.Command = CmdNick
		reply.Text = ""
		// TODO: Send nick change to all channels user is joined to.
	}

	reply.Params = []string{conn.user.Nick()}

	conn.Write(reply.RenderBuffer())
}

// HandleUser processes a USER command.
//
// First, it checks if the specieifed username is in use by the user issuing
// the command; by another user on the server; or disallowed by the server
// configuration. Then it checks the validity of the username formatting
// before finally, if all of the requirements are met, sets the User object
// Name field to the specified name in the command parameters.
//
//    Command: USER
//    Parameters: <username> <modemask> -0(unused)- :[realname]
func HandleUser(conn *Conn, msg *Message) {
	defer msgpool.Recycle(msg)

	if !enoughParams(msg, 3) {
		conn.ReplyNeedMoreParams(msg.Command)
		return
	}

	if len(conn.user.Nick()) < 1 {
		conn.ReplyNoNicknameGiven()
		return
	}

	reply := conn.newMessage()
	defer msgpool.Recycle(reply)

	reply.Params = []string{conn.user.Nick()}
	reply.Code = ReplyAlreadyRegistered

	if len(conn.user.Name()) > 0 {
		reply.Text = ErrUserAreadySet.String()
		conn.Write(reply.RenderBuffer())
		return
	}

	if conn.server.Users.Exists(msg.Params[0]) {
		reply.Text = ErrUserInUse.String()
		conn.Write(reply.RenderBuffer())
		return
	}

	// TODO: Username restriction check

	// TODO: Username formatting checks
	// This ties into configurations such as:
	// - username length
	// - realname length
	// - reserved names

	conn.user.SetName(msg.Params[0])
	conn.user.SetRealname(msg.Text)
	conn.user.SetHostname(conn.remAddr)
	conn.regiserUser()

	if !conn.capRequested || conn.capNegotiated {
		conn.ReplyWelcome()
		conn.ReplyISupport()
	}
}

// HandleCap processes the CAP command and sub commands for
// negotiating capabilties per the IRCv3.2 spec.
//
//    Command: CAP
//    Parameters: <subcommand> [param] :[capabiliy] [capability]
func HandleCap(conn *Conn, msg *Message) {
	defer msgpool.Recycle(msg)

	if !enoughParams(msg, 2) {
		conn.ReplyInvalidCapCommand(msg.Command)
		return
	}

	switch msg.Params[1] {
	case "LS":
		fallthrough
	case "LIST":
		// conn.ListCapabilities() // TODO: List capabilities
	case "REQ":
		if !enoughParams(msg, 3) {
			conn.ReplyNeedMoreParams(msg.Command)
		}
		// conn.HandleCapRequest(msg.Params[2]) // TODO: Capability request handler
	case "END":
		conn.capNegotiated = true
		if conn.registered {
			conn.ReplyWelcome()
			conn.ReplyISupport()
		}
	default:
		conn.ReplyInvalidCapCommand(msg.Command)
		return
	}
}

// HandlePrivmsg processes a PRIVMSG command.
//
// First, it checks if the specified nickname or channel exists; then
// checks if the sender is disallowed from sending the message by the
// sender's usermode. If all of the requirements are met, it sends
// the message to the intended recpient.
//
//    Command: PRIVMSG
//    Parameters: <target> :<text>
func HandlePrivmsg(conn *Conn, msg *Message) {
	doChatMessage(conn, msg)
}

// HandleNotice processes a NOTICE command.
//
// First, it checks if the specified nickname or channel exists; then
// checks if the sender is disallowed from sending the message by the
// sender's usermode. If all of the requirements are met, it sends
// the message to the intended recpient.
//
//    Command: NOTICE
//    Parameters: <target> :<text>
func HandleNotice(conn *Conn, msg *Message) {
	doChatMessage(conn, msg)
}

func doChatMessage(conn *Conn, msg *Message) {
	defer msgpool.Recycle(msg)

	if !enoughParams(msg, 1) || len(msg.Text) < 1 {
		conn.ReplyNeedMoreParams(msg.Command)
		return
	}

	// TODO: Send Message permission check

	targetuser, uerr := conn.server.Nicks.Get(strings.ToLower(msg.Params[0]))
	targetchan, cerr := conn.server.Channels.Get(strings.ToLower(msg.Params[0]))

	if uerr != nil && cerr != nil {
		log.Debug("irc: Chat Message: did not find target")
		conn.ReplyNoSuchNick(msg.Params[0])
		return
	}

	msg.Params = msg.Params[0:1] // Strip erroneous parameters.
	msg.Sender = conn.user.Hostmask()

	if targetuser != nil {
		targetuser.conn.Write(msg.RenderBuffer())
	} else {
		targetchan.Send(msg, conn.user.Nick())
	}
}

// HandleJoin processes a JOIN command.
//
// The server will first check if the channel exists, if not,
// create a new channel. Then, the user will be added to the
// channel members if the the user has sufficient permissions;
// which are implied if the channel must first be created.
//
//    Command: JOIN
//    Prameters: <channel>
func HandleJoin(conn *Conn, msg *Message) {
	defer msgpool.Recycle(msg)

	if !enoughParams(msg, 1) {
		conn.ReplyNeedMoreParams(msg.Command)
		return
	}

	msg.Sender = conn.user.Hostmask()
	msg.Params = msg.Params[0:1]

	channel, err := conn.server.Channels.Get(strings.ToLower(msg.Params[0]))

	if err != nil {
		channel = NewChannel(msg.Params[0], conn.user)
		conn.server.Channels.Add(strings.ToLower(msg.Params[0]), channel)
	}

	if !channel.Join(conn.user, msg) {
		// TODO: channel join error
	} else {
		conn.channels.Add(channel.Name(), channel)
		conn.ReplyChannelNames(channel)
	}
}

// HandleUserhost processes a USERHOST command originated from the client.
//
// The server will respond with the matching hostname of the requested nicks.
// Limit 5
//
//    Command: USERHOST
//    Parameters: <nickname1> [nickname2] [nickname3] [nickname4] [nickname5]
func HandleUserhost(conn *Conn, msg *Message) {
	defer msgpool.Recycle(msg)

	hosts := []string{}

	var buffer bytes.Buffer

	for _, nick := range msg.Params {
		host, err := conn.server.Nicks.Get(strings.ToLower(nick))
		if err != nil {
			// TODO: Nick not fouind
			conn.ReplyNoSuchNick(nick)
			return
		}

		// TODO: Visibility permissions
		buffer.WriteString(nick)
		buffer.WriteString("=+")
		buffer.WriteString(host.Hostmask())
		hosts = append(hosts, buffer.String())
		buffer.Reset()

	}

	msg.Sender = conn.server.Hostname()
	msg.Command = ""
	msg.Code = ReplyUserHost
	msg.Params = []string{conn.user.Nick()}
	msg.Text = strings.Join(hosts, " ")

	conn.Write(msg.RenderBuffer())
}

// HandlePing processes a PING command originated from the client.
//
// The server will respond with the matching ping token.
//
//    Command: PING
//    Parameters: :<token>
func HandlePing(conn *Conn, msg *Message) {
	defer msgpool.Recycle(msg)

	msg.Sender = conn.server.Hostname()

	msg.Command = CmdPong

	conn.Write(msg.RenderBuffer())
}

// HandlePong processes a PONG command in reply to a server sent PING command.
//
// Command: PONG
// Parameters: :<token>
func HandlePong(conn *Conn, msg *Message) {
	defer msgpool.Recycle(msg)

	if len(msg.Text) < 1 {
		conn.ReplyNeedMoreParams(msg.Command)
		return
	}

	conn.Lock()
	defer conn.Unlock()
	conn.lastPingRecv = msg.Text
}

// RouteCommand accepts an IRC message and routes it to a function
// in which is designed to process the command.
func RouteCommand(conn *Conn, msg *Message) {
	handler, exists := Handlers[msg.Command]

	if !exists {
		conn.ReplyNotImplemented(msg.Command)
		msgpool.Recycle(msg)
		return
	}

	if !conn.registered {
		if msg.Command != CmdPing &&
			msg.Command != CmdPong &&
			msg.Command != CmdCap &&
			msg.Command != CmdPass &&
			msg.Command != CmdNick &&
			msg.Command != CmdUser &&
			msg.Command != CmdQuit {

			conn.ReplyNotRegistered()
			return
		}
	}

	handler(conn, msg)
}

func enoughParams(msg *Message, expected int) bool {
	return !(len(msg.Params) < expected)
}

func registerHandlers() {
	Handlers[CmdQuit] = HandleQuit
	Handlers[CmdNick] = HandleNick
	Handlers[CmdUser] = HandleUser
	Handlers[CmdPing] = HandlePing
	Handlers[CmdPong] = HandlePong
	Handlers[CmdJoin] = HandleJoin
	Handlers[CmdPrivMsg] = HandlePrivmsg
	Handlers[CmdNotice] = HandleNotice
	Handlers[CmdUserhost] = HandleUserhost
}
