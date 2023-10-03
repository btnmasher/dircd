/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

import (
	"bytes"
	"fmt"
	"strings"
)

// Handlers is a map of functions where the handlers are stored.
var Handlers = make(map[string]MessageHandler)
var UnregAllowedCommands = make(map[string]struct{})

// MessageHandler defines the function signature of a handler used to
// process IRC messages.
type MessageHandler func(*Conn, *Message)

// All command handler functions do not return an error. Instead, it
// must process all error conditions relating to the command and reply
// to the user in the correct way specified by RFC2812.

// HandleQuit processes a QUIT command.
//
// The connection will be scheduled for immediate deadline, and the
// server will broadcast the QUIT message to all channels the user is
// joined to.
//
//	Command: QUIT
//	Parameters: :<reason>
func HandleQuit(conn *Conn, msg *Message) {
	defer msgPool.Recycle(msg)
	conn.doQuit(msg.Trailing)
}

// HandleNick processes a NICK command.
//
// First, it checks if the current nickname is in use by the user issuing
// the command; by another user on the server; or disallowed by the server
// configuration. Then it checks the validity of the nickname formatting
// before finally, if all the requirements are met, sets the User object
// Nick field to the specified name in the command parameters.
//
//	Command: NICK
//	Parameters: <nickname>
func HandleNick(conn *Conn, msg *Message) {
	defer msgPool.Recycle(msg)

	logger := conn.logger.WithField("handler", "NICK")

	if !enoughParams(msg, 1) {
		msg.Params = strings.Fields(msg.Trailing) // Some dumb ass clients don't follow the spec and add the nickname to the text field (mIRC)
		if !enoughParams(msg, 1) {
			conn.ReplyNoNicknameGiven()
			return
		}
		logger.Warn("received non-standard command format, adapting")
	}

	reply := conn.newMessage()
	defer msgPool.Recycle(reply)

	if conn.user.Nick() == msg.Params[0] {
		reply.Trailing = ErrNickAlreadySet.String()
		reply.Code = ReplyNicknameInUse
		conn.Write(reply.RenderBuffer())
		return
	}

	if validationErr, code := conn.server.ValidateName(msg.Params[0]); validationErr != nil {
		reply.Trailing = validationErr.Error()
		reply.Code = code
		conn.Write(reply.RenderBuffer())
		return
	}

	reply.Source = conn.user.Hostmask()
	oldNick := conn.user.Nick()
	newNick := msg.Params[0]
	conn.user.SetNick(newNick)

	if !conn.isRegistered() {
		return
	}

	reply.Code = ReplyNone
	reply.Command = CmdNick
	reply.Params = msg.Params[0:1]
	reply.Trailing = ""

	conn.Write(reply.RenderBuffer())

	if conn.channels.Length() > 0 {
		changeErr := conn.channels.ForEach(func(name string, channel *Channel) error {
			return channel.ChangeNick(oldNick, newNick, reply)
		})
		if changeErr != nil {
			logger.Error(fmt.Errorf("error encountered attempting to change nick for channel: %w", changeErr))
			// TODO: this is real bad cause then state is gonna be all fuckered, need some reconciliation
		}
	}
}

// HandleUser processes a USER command.
//
// First, it checks if the specified username is in use by the user issuing
// the command; by another user on the server; or disallowed by the server
// configuration. Then it checks the validity of the username formatting
// before finally, if all the requirements are met, sets the User object
// Name field to the specified name in the command parameters.
//
//	Command: USER
//	Parameters: <username> <modemask> -0(unused)- :[realname]
func HandleUser(conn *Conn, msg *Message) {
	defer msgPool.Recycle(msg)

	if !enoughParams(msg, 3) {
		conn.ReplyNeedMoreParams(msg.Command)
		return
	}

	if len(conn.user.Nick()) == 0 {
		conn.ReplyNoNicknameGiven()
		return
	}

	reply := conn.newMessage()
	defer msgPool.Recycle(reply)

	reply.Params = []string{conn.user.Nick()}
	reply.Code = ReplyAlreadyRegistered

	if len(conn.user.Name()) > 0 {
		reply.Trailing = ErrUserAreadySet.String()
		conn.Write(reply.RenderBuffer())
		return
	}

	if conn.server.Users.Exists(msg.Params[0]) {
		reply.Trailing = ErrUserInUse.String()
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
	conn.user.SetRealname(msg.Trailing)
	conn.user.SetHostname(conn.remAddr)
	conn.registerUser()

	if !conn.capRequested || conn.capNegotiated {
		conn.ReplyWelcome()
		conn.ReplyISupport()
	}
}

// HandleCap processes the CAP command and sub commands for
// negotiating capabilties per the IRCv3.2 spec.
//
//	Command: CAP
//	Parameters: <subcommand> [param] :[capabiliy] [capability]
func HandleCap(conn *Conn, msg *Message) {
	defer msgPool.Recycle(msg)

	if !enoughParams(msg, 2) {
		conn.ReplyInvalidCapCommand(msg.Command)
		return
	}

	switch msg.Params[0] {
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
		if conn.isRegistered() {
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
//	Command: PRIVMSG
//	Parameters: <target> :<text>
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
//	Command: NOTICE
//	Parameters: <target> :<text>
func HandleNotice(conn *Conn, msg *Message) {
	doChatMessage(conn, msg)
}

func doChatMessage(conn *Conn, msg *Message) {
	defer msgPool.Recycle(msg)

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

// HandleJoin processes a JOIN command.
//
// The server will first check if the channel exists, if not,
// create a new channel. Then, the user will be added to the
// channel members if the the user has sufficient permissions;
// which are implied if the channel must first be created.
//
//	Command: JOIN
//	Prameters: <channel>
func HandleJoin(conn *Conn, msg *Message) {
	defer msgPool.Recycle(msg)

	if !enoughParams(msg, 1) {
		conn.ReplyNeedMoreParams(msg.Command)
		return
	}

	msg.Source = conn.user.Hostmask()
	msg.Params = msg.Params[0:1]

	channel, exists := conn.server.Channels.Get(strings.ToLower(msg.Params[0]))

	if !exists {
		channel = NewChannel(msg.Params[0], conn.user)
		conn.server.Channels.Set(strings.ToLower(msg.Params[0]), channel)
	}

	if !channel.Join(conn.user, msg) {
		// TODO: channel join error
	} else {
		conn.channels.Set(channel.Name(), channel)
		conn.ReplyChannelNames(channel)
	}
}

// HandleUserhost processes a USERHOST command originated from the client.
//
// The server will respond with the matching hostname of the requested nicks.
// Limit 5
//
//	Command: USERHOST
//	Parameters: <nickname1> [nickname2] [nickname3] [nickname4] [nickname5]
func HandleUserhost(conn *Conn, msg *Message) {
	defer msgPool.Recycle(msg)

	var hosts []string
	var buffer bytes.Buffer

	for _, nick := range msg.Params {
		host, exists := conn.server.Nicks.Get(strings.ToLower(nick))
		if !exists {
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

	msg.Source = conn.hostname
	msg.Command = ""
	msg.Code = ReplyUserHost
	msg.Params = []string{conn.user.Nick()}
	msg.Trailing = strings.Join(hosts, " ")

	conn.Write(msg.RenderBuffer())
}

// HandlePing processes a PING command originated from the client.
//
// The server will respond with the matching ping token.
//
//	Command: PING
//	Parameters: :<token>
func HandlePing(conn *Conn, msg *Message) {
	defer msgPool.Recycle(msg)

	msg.Source = conn.hostname
	msg.Command = CmdPong
	conn.Write(msg.RenderBuffer())
}

// HandlePong processes a PONG command in reply to a server sent PING command.
//
// Command: PONG
// Parameters: :<token>
func HandlePong(conn *Conn, msg *Message) {
	defer msgPool.Recycle(msg)

	if len(msg.Trailing) == 0 {
		conn.ReplyNeedMoreParams(msg.Command)
		return
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.lastPingRecv = msg.Trailing
}

// RouteCommand accepts an IRC message and routes it to a function
// in which is designed to process the command.
func RouteCommand(conn *Conn, msg *Message) {
	handler, exists := Handlers[msg.Command]
	if !exists {
		conn.ReplyNotImplemented(msg.Command)
		msgPool.Recycle(msg)
		return
	}

	if !conn.isRegistered() {
		if _, allowed := UnregAllowedCommands[msg.Command]; !allowed {
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
	Handlers[CmdCap] = HandleCap
	Handlers[CmdPrivMsg] = HandlePrivmsg
	Handlers[CmdNotice] = HandleNotice
	Handlers[CmdUserhost] = HandleUserhost

	UnregAllowedCommands[CmdPing] = struct{}{}
	UnregAllowedCommands[CmdPong] = struct{}{}
	UnregAllowedCommands[CmdCap] = struct{}{}
	UnregAllowedCommands[CmdPass] = struct{}{}
	UnregAllowedCommands[CmdNick] = struct{}{}
	UnregAllowedCommands[CmdUser] = struct{}{}
	UnregAllowedCommands[CmdQuit] = struct{}{}
}
