/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

// Command constants.
const (
	// RFC 1459
	CmdPrivMsg  string = "PRIVMSG"
	CmdNotice          = "NOTICE"
	CmdUserhost        = "USERHOST"
	CmdPass            = "PASS"
	CmdPing            = "PING"
	CmdPong            = "PONG"
	CmdTopic           = "TOPIC"
	CmdJoin            = "JOIN"
	CmdPart            = "PART"
	CmdKick            = "KICK"
	CmdQuit            = "QUIT"
	CmdNick            = "NICK"
	CmdUser            = "USER"
	CmdMode            = "MODE"
	CmdWallops         = "WALLOPS"
	CmdInvite          = "INVITE"
	CmdKill            = "KILL"

	// CTCP
	CmdCTCPPing       = "CTCP PING"
	CmdCTCPVersion    = "CTCP VERSION"
	CmdCTCPSource     = "CTCP SOURCE"
	CmdCTCPTime       = "CTCP TIME"
	CmdCTCPUserInfo   = "CTCP USERINFO"
	CmdCTCPClientInfo = "CTCP CLIENTINFO"
	CmdCTCPError      = "CTCP ERRMSG"
	CmdCTCPFinger     = "CTCP FINGER"
	CmdCTCPAction     = "CTCP ACTION"

	// IRCv3 Base
	CmdCap      = "CAP"
	CmdCapLs    = "CAP LS"
	CmdCapList  = "CAP LIST"
	CmdCapReq   = "CAP REQ"
	CmdCapAck   = "CAP ACK"
	CmdCapNak   = "CAP NAK"
	CmdCapEnd   = "CAP END"
	CmdAuth     = "AUTHENTICATE"
	CmdMetadata = "METADATA"
	CmdError    = "ERROR"

	// IRCv3 account-notify
	CmdAccount = "ACCOUNT"

	// IRCv3 away-notify
	CmdAway = "AWAY"
)
