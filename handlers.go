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

// All command handler functions do not return an error. Instead, it
// must process all error conditions relating to the command and reply
// to the user in the correct way specified by RFC2812. They can also
// stop execution of the handler chain in which they may be included
// by calling ctx.Handled() or ctx.AbortError(error) to have an error
// logged about why the command was aborted.

// HandleQuit processes a QUIT command.
//
// The connection will be scheduled for immediate deadline, and the
// server will broadcast the QUIT message to all channels the user is
// joined to.
//
//	Command: QUIT
//	Parameters: :<reason>
func HandleQuit(ctx *MessageContext) {
	ctx.Conn.doQuit(ctx.Msg.Trailing)
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
func HandleNick(ctx *MessageContext) {
	ctx.Handled()
	if !enoughParams(ctx.Msg, 1) {
		// Some dumb ass clients (mIRC) don't follow the spec and add a colon before the name
		ctx.Msg.Params = strings.Fields(ctx.Msg.Trailing)
		if !enoughParams(ctx.Msg, 1) {
			ctx.Conn.ReplyNoNicknameGiven()
			return
		}
	}

	reply := ctx.Conn.newMessage()
	defer msgPool.Recycle(reply)

	if ctx.Conn.user.Nick() == ctx.Msg.Params[0] {
		reply.Trailing = ErrNickAlreadySet.String()
		reply.Code = ReplyNicknameInUse
		ctx.Conn.Write(reply.RenderBuffer())
		return
	}

	if validationErr, code := ctx.Conn.server.ValidateName(ctx.Msg.Params[0]); validationErr != nil {
		reply.Trailing = validationErr.Error()
		reply.Code = code
		ctx.Conn.Write(reply.RenderBuffer())
		return
	}

	reply.Source = ctx.Conn.user.Hostmask()
	oldNick := ctx.Conn.user.Nick()
	newNick := ctx.Msg.Params[0]
	ctx.Conn.user.SetNick(newNick)

	if !ctx.Conn.isRegistered() {
		return
	}

	ctx.Conn.server.Nicks.ChangeKey(oldNick, newNick)
	reply.Code = ReplyNone
	reply.Command = CmdNick
	reply.Params = ctx.Msg.Params[0:1]
	reply.Trailing = ""

	ctx.Conn.Write(reply.RenderBuffer())

	if ctx.Conn.channels.Length() > 0 {
		changeErr := ctx.Conn.channels.ForEach(func(name string, channel *Channel) error {
			return channel.ChangeNick(oldNick, newNick, reply)
		})
		if changeErr != nil {
			ctx.Conn.logger.WithField("handler", "NICK").
				Error(fmt.Errorf("error encountered attempting to change nick for channel: %w", changeErr))
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
func HandleUser(ctx *MessageContext) {
	ctx.Handled()
	if !enoughParams(ctx.Msg, 3) {
		ctx.Conn.ReplyNeedMoreParams(ctx.Msg.Command)
		return
	}

	if len(ctx.Conn.user.Nick()) == 0 {
		ctx.Conn.ReplyNoNicknameGiven()
		return
	}

	reply := ctx.Conn.newMessage()
	defer msgPool.Recycle(reply)

	reply.Params = []string{ctx.Conn.user.Nick()}
	reply.Code = ReplyAlreadyRegistered

	if ctx.Conn.isRegistered() {
		reply.Trailing = ErrUserAreadySet.String()
		ctx.Conn.Write(reply.RenderBuffer())
		return
	}

	if ctx.Conn.server.Users.Exists(ctx.Msg.Params[0]) {
		reply.Trailing = ErrUserInUse.String()
		ctx.Conn.Write(reply.RenderBuffer())
		return
	}

	// TODO: Username restriction check

	// TODO: Username formatting checks
	// This ties into configurations such as:
	// - username length
	// - realname length
	// - reserved names

	ctx.Conn.user.SetName(ctx.Msg.Params[0])
	ctx.Conn.user.SetRealname(ctx.Msg.Trailing)
	ctx.Conn.user.SetHostname(ctx.Conn.remAddr)
	ctx.Conn.registerUser()

	if !ctx.Conn.capRequested || ctx.Conn.capNegotiated {
		ctx.Conn.ReplyWelcome()
		ctx.Conn.ReplyISupport()
	}
}

// HandleCap processes the CAP command and sub commands for
// negotiating capabilities per the IRCv3.2 spec.
//
//	Command: CAP
//	Parameters: <subcommand> [param] :[capability] [capability]
func HandleCap(ctx *MessageContext) {
	ctx.Handled()
	if !enoughParams(ctx.Msg, 2) || !ctx.Conn.capNegotiated { // TODO: see what the spec says about CAP after end
		ctx.Conn.ReplyInvalidCapCommand(ctx.Msg.Command)
		return
	}

	switch ctx.Msg.Params[0] {
	case "LS":
		fallthrough
	case "LIST":
		// ctx.Conn.ListCapabilities() // TODO: List capabilities
	case "REQ":
		if !enoughParams(ctx.Msg, 3) {
			ctx.Conn.ReplyNeedMoreParams(ctx.Msg.Command)
		}
		// ctx.Conn.HandleCapRequest(msg.Params[2]) // TODO: Capability request handler
	case "END":
		ctx.Conn.capNegotiated = true
		if ctx.Conn.isRegistered() {
			ctx.Conn.ReplyWelcome()
			ctx.Conn.ReplyISupport()
		}
	default:
		ctx.Conn.ReplyInvalidCapCommand(ctx.Msg.Command)
		return
	}
}

// HandlePrivmsg processes a PRIVMSG command.
//
// First, it checks if the specified nickname or channel exists; then
// checks if the sender is disallowed from sending the message by the
// sender's usermode. If all the requirements are met, it sends the
// message to the intended recipient.
//
//	Command: PRIVMSG
//	Parameters: <target> :<text>
func HandlePrivmsg(ctx *MessageContext) {
	ctx.Handled()
	ctx.Conn.doChatMessage(ctx.Msg)
}

// HandleNotice processes a NOTICE command.
//
// First, it checks if the specified nickname or channel exists; then
// checks if the sender is disallowed from sending the message by the
// sender's usermode. If all the requirements are met, it sends
// the message to the intended recipient.
//
//	Command: NOTICE
//	Parameters: <target> :<text>
func HandleNotice(ctx *MessageContext) {
	ctx.Handled()
	ctx.Conn.doChatMessage(ctx.Msg)
}

// HandleJoin processes a JOIN command.
//
// The server will first check if the channel exists, if not,
// create a new channel. Then, the user will be added to the
// channel members if the user has sufficient permissions;
// which are implied if the channel must first be created.
//
//	Command: JOIN
//	Parameters: <channel>
func HandleJoin(ctx *MessageContext) {
	ctx.Handled()
	if !enoughParams(ctx.Msg, 1) {
		ctx.Conn.ReplyNeedMoreParams(ctx.Msg.Command)
		return
	}

	ctx.Msg.Source = ctx.Conn.user.Hostmask()
	ctx.Msg.Params = ctx.Msg.Params[0:1]

	channel, exists := ctx.Conn.server.Channels.Get(strings.ToLower(ctx.Msg.Params[0]))

	if !exists {
		channel = NewChannel(ctx.Msg.Params[0], ctx.Conn.user)
		ctx.Conn.server.Channels.Set(strings.ToLower(ctx.Msg.Params[0]), channel)
	}

	if !channel.Join(ctx.Conn.user, ctx.Msg) {
		// TODO: channel join error
	} else {
		ctx.Conn.channels.Set(channel.Name(), channel)
		ctx.Conn.ReplyChannelNames(channel)
	}
}

// HandleUserhost processes a USERHOST command originated from the client.
//
// The server will respond with the matching hostname of the requested nicks.
// Limit 5
//
//	Command: USERHOST
//	Parameters: <nickname1> [nickname2] [nickname3] [nickname4] [nickname5]
func HandleUserhost(ctx *MessageContext) {
	ctx.Handled()
	var hosts []string
	var buffer bytes.Buffer

	for _, nick := range ctx.Msg.Params {
		host, exists := ctx.Conn.server.Nicks.Get(strings.ToLower(nick))
		if !exists {
			ctx.Conn.ReplyNoSuchNick(nick)
			return
		}

		// TODO: Visibility permissions
		buffer.WriteString(nick)
		buffer.WriteString("=+")
		buffer.WriteString(host.Hostmask())
		hosts = append(hosts, buffer.String())
		buffer.Reset()

	}

	ctx.Msg.Source = ctx.Conn.hostname
	ctx.Msg.Command = ""
	ctx.Msg.Code = ReplyUserHost
	ctx.Msg.Params = []string{ctx.Conn.user.Nick()}
	ctx.Msg.Trailing = strings.Join(hosts, " ")

	ctx.Conn.Write(ctx.Msg.RenderBuffer())
}

// HandlePing processes a PING command originated from the client.
//
// The server will respond with the matching ping token.
//
//	Command: PING
//	Parameters: :<token>
func HandlePing(ctx *MessageContext) {
	ctx.Handled()
	ctx.Msg.Source = ctx.Conn.hostname
	ctx.Msg.Command = CmdPong
	ctx.Conn.Write(ctx.Msg.RenderBuffer())
}

// HandlePong processes a PONG command in reply to a server sent PING command.
//
// Command: PONG
// Parameters: :<token>
func HandlePong(ctx *MessageContext) {
	ctx.Handled()

	ctx.Conn.mu.Lock()
	defer ctx.Conn.mu.Unlock()

	if len(ctx.Msg.Trailing) == 0 {
		ctx.Conn.ReplyNeedMoreParams(ctx.Msg.Command)
		return
	}

	ctx.Conn.lastPingRecv = ctx.Msg.Trailing
}

func MustBeRegistered(ctx *MessageContext) {
	if !ctx.Conn.isRegistered() {
		ctx.Handled()
		ctx.Conn.ReplyNotRegistered()
	}
}
