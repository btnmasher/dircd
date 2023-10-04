/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

import (
	"github.com/btnmasher/dircd/shared/sliceutils"
	"github.com/btnmasher/dircd/shared/stringutils"
)

// ReplyWelcome returns the configured welcome message to
// the user. This is sent when a client first connects
// and registers successfully.
func (conn *Conn) ReplyWelcome() {
	msg := conn.newMessage()
	defer msgPool.Recycle(msg)

	msg.Code = ReplyWelcome
	msg.Params = []string{conn.user.Nick()}
	msg.Trailing = conn.server.Welcome()

	conn.Write(msg.RenderBuffer())
}

// ReplyInvalidCapCommand returns an error message to the user
// in the event that a CAP command issued by the user is not
// a valid subcommand per the IRCv3 CAP specifications.
func (conn *Conn) ReplyInvalidCapCommand(cmd string) {
	msg := conn.newMessage()
	defer msgPool.Recycle(msg)

	nick := conn.user.Nick()

	if len(nick) == 0 {
		nick = "*"
	}

	params := []string{nick}

	if cmd != "" {
		params = append(params, cmd)
	}

	msg.Code = ReplyInvalidCapCmd
	msg.Params = params
	msg.Trailing = ErrInvalidCapCmd.Error()

	conn.Write(msg.RenderBuffer())
}

// ReplyNeedMoreParams returns an error message to the user
// in the event that a command issued by the user that does
// not satisfy the minimum number of parameters expected of
// the particular command.
func (conn *Conn) ReplyNeedMoreParams(cmd string) {
	msg := conn.newMessage()
	defer msgPool.Recycle(msg)

	nick := conn.user.Nick()

	if len(nick) == 0 {
		nick = "*"
	}

	params := []string{nick}

	if cmd != "" {
		params = append(params, cmd)
	}

	msg.Code = ReplyNeedMoreParams
	msg.Params = params
	msg.Trailing = ErrMissingParams.Error()

	conn.Write(msg.RenderBuffer())
}

// ReplyNoNicknameGiven returns an error message to the user
// in the event that a command issued by the user that does
// not satisfy the requirement of specifying a nickname.
func (conn *Conn) ReplyNoNicknameGiven() {
	msg := conn.newMessage()
	defer msgPool.Recycle(msg)

	nick := conn.user.Nick()

	if len(nick) == 0 {
		nick = "*"
	}

	msg.Params = []string{nick}
	msg.Code = ReplyNoNicknameGiven
	msg.Trailing = ErrNoNickGiven.Error()

	conn.Write(msg.RenderBuffer())
}

// ReplyNoSuchNick returns an error message to the user
// in the event that a command issued by the user with
// a target nickname cannot find the target or is unable
// to know of the targets existence due to permissions.
func (conn *Conn) ReplyNoSuchNick(nick string) {
	msg := conn.newMessage()
	defer msgPool.Recycle(msg)

	msg.Params = []string{conn.user.Nick(), nick}
	msg.Code = ReplyNoSuchNick
	msg.Trailing = ErrNoSuchNick.Error()

	conn.Write(msg.RenderBuffer())
}

// ReplyNoSuchChan returns an error message to the user
// in the event that a command issued by the user with
// a target channel cannot find the target or is unable
// to know of the targets existence due to permissions.
func (conn *Conn) ReplyNoSuchChan(channel string) {
	msg := conn.newMessage()
	defer msgPool.Recycle(msg)

	msg.Params = []string{conn.user.Nick(), channel}
	msg.Code = ReplyNoSuchChannel
	msg.Trailing = ErrNoSuchChan.Error()

	conn.Write(msg.RenderBuffer())
}

// ReplyNotImplemented returns an error message to the user
// in the event the given command is not apart of the handlers
// found in RouteCommand()
func (conn *Conn) ReplyNotImplemented(cmd string) {
	msg := conn.newMessage()
	defer msgPool.Recycle(msg)

	msg.Code = ReplyUnknownCommand
	msg.Params = []string{conn.user.Nick(), cmd}
	msg.Trailing = ErrNotImplemented.Error()
	conn.Write(msg.RenderBuffer())
}

// ReplyNotRegistered returns an error message to the user
// in the event the given command is not a part of the handlers
// found in RouteCommand()
func (conn *Conn) ReplyNotRegistered() {
	msg := conn.newMessage()
	defer msgPool.Recycle(msg)

	nick := conn.user.Nick()

	if len(nick) == 0 {
		nick = "*"
	}

	msg.Code = ReplyNotRegistered
	msg.Params = []string{nick}
	msg.Trailing = ErrNotRegistered.Error()

	conn.Write(msg.RenderBuffer())
}

// ReplyChannelTopic returns the topic reply to the user for
// the given channel.
func (conn *Conn) ReplyChannelTopic(channel *Channel) {
	msg := conn.newMessage()
	defer msgPool.Recycle(msg)

	msg.Code = ReplyChanTopic
	msg.Params = []string{conn.user.Nick(), channel.Name()}
	msg.Trailing = channel.Topic()
	conn.Write(msg.RenderBuffer())
}

// ReplyChannelNames returns the topic reply to the user for
// the given channel.
func (conn *Conn) ReplyChannelNames(channel *Channel) {
	nickList := channel.GetNicks()
	userNick := conn.user.Nick()
	channelName := channel.Name()
	params := []string{userNick, "=", channelName}

	temp := conn.newMessage()
	temp.Code = ReplyNames
	temp.Params = params

	joined := stringutils.ChunkJoinStrings(MaxMsgLength-len(temp.String()), SPACE, nickList...)
	msgPool.Recycle(temp)

	messages := make([]*Message, 0, len(joined)+1)

	for _, line := range joined {
		msg := conn.newMessage()
		msg.Code = ReplyNames
		msg.Params = params
		msg.Trailing = line
		messages = append(messages, msg)
	}

	end := conn.newMessage()
	end.Code = ReplyEndOfNames
	end.Params = []string{userNick, channelName}
	end.Trailing = "End of NAMES list."
	messages = append(messages, end)

	defer func() {
		for i := range messages {
			msgPool.Recycle(messages[i])
		}
	}()

	for i := range messages {
		conn.Write(messages[i].RenderBuffer())
	}
}

// ReplyISupport returns the ISupport information about the server
func (conn *Conn) ReplyISupport() {

	support := conn.server.ISupport()
	params := []string{conn.user.Nick()}

	temp := conn.newMessage()
	defer msgPool.Recycle(temp)
	temp.Code = ReplyISupport
	temp.Params = params

	var paramLines []string

	// Spec only allows for 15 max parameters, with two parameters being taken up by the client host and message text
	if len(support) <= MaxMsgParams-2 {
		paramLines = stringutils.ChunkJoinStrings(MaxMsgLength-len(temp.String()), SPACE, support...)
	} else {
		// We have more parameters than can fit onto a single message, chunk them into groups of 13
		supportChunks := sliceutils.ChunkBy(support, MaxMsgParams-2)
		for i := range supportChunks {
			joinedPartial := stringutils.ChunkJoinStrings(MaxMsgLength-len(temp.String()), SPACE, supportChunks[i]...)
			paramLines = append(paramLines, joinedPartial...)
		}
	}

	messages := make([]*Message, 0, len(paramLines))

	for i := range paramLines {
		msg := conn.newMessage()

		msg.Code = ReplyISupport
		msg.Params = append(params, paramLines[i])

		messages = append(messages, msg)
	}

	defer func() {
		for i := range messages {
			msgPool.Recycle(messages[i])
		}
	}()

	for _, m := range messages {
		conn.Write(m.RenderBuffer())
	}
}
