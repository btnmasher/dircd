/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Message is an object that represents the components of an IRC message.
type Message struct {
	Tags map[string]string `json:"tags"`

	Source   string   `json:"sender"`   // The sender parameter of the message.
	Command  string   `json:"command"`  // The IRC string command of the message.
	Code     uint16   `json:"code"`     // The IRC numeric code of the message (substituted as the command when replying).
	Params   []string `json:"params"`   // The person of the message after prefix and command in array form.
	Trailing string   `json:"trailing"` // The final parameter of a message, may contain spaces (eg: text of a PRIVMSG)
}

// Message represents an IRC protocol message.
//
//    <message>  = ['@' <tags> SPACE] [':' <source> <SPACE> ] <command> <params> <crlf>
//
//    <tags>          ::= <tag> [';' <tag>]*
//    <tag>           ::= <key> ['=' <escaped value>]
//    <key>           ::= [ <client_prefix> ] [ <vendor> '/' ] <sequence of letters, digits, hyphens (`-`)>
//    <client_prefix> ::= '+'
//    <escaped value> ::= <sequence of any characters except NUL, CR, LF, semicolon (`;`) and SPACE>
//    <vendor>        ::= <host>
//
//    <source>   = <servername> | <nick> [ '!' <user> ] [ '@' <host> ]
//
//    <command>  = <letter> { <letter> } | <number> <number> <number>
//    <SPACE>    = ' ' { ' ' }
//    <params>   = <SPACE> [ ':' <trailing> | <middle> <params> ]
//
//    <middle>   = <Any *non-empty* sequence of octets not including SPACE
//                   or NUL or CR or LF, the first of which may not be ':'>
//    <trailing> = <Any, possibly *empty*, sequence of octets not including
//                   NUL or CR or LF>
//
//    <crlf>     = CR LF

// String constants for constructing the message
const (
	SPACE        string = " "
	EMPTY               = ""
	PADNUM              = "%03d"
	AT                  = "@"
	EQUAL               = "="
	COLON               = ":"
	SEMICOLON           = ";"
	ESCSEMICOLON        = "\\:"
	BACKSLAH            = "\\"
	ESCBACKSLASH        = "\\\\"
	CRLF                = "\r\n"
	ESCCRLF             = "\\r\\n"
	CR                  = "\r"
	ESCCR               = "\\r"
	LF                  = "\n"
	ESCLF               = "\\n"
	ESCSPACE            = "\\s"
)

func NewMessage() *Message {
	return &Message{}
}

// String returns the IRC-formatted string version of a message object.
// This is here to satisfy a Stringer interface
func (msg *Message) String() string {
	return msg.Render()
}

// RenderBuffer returns the IRC-formatted byte buffer version of a message object.
func (msg *Message) RenderBuffer() *bytes.Buffer {
	buffer := bufPool.New()

	if len(msg.Tags) > 0 {
		buffer.WriteString(AT)
		for key, value := range msg.Tags {
			buffer.WriteString(escapeTagString(key))
			buffer.WriteString(EQUAL)
			buffer.WriteString(escapeTagString(value))
			buffer.WriteString(SEMICOLON)
		}
		buffer.Truncate(buffer.Len() - 1) // remove trailing ";"
		buffer.WriteString(SPACE)
	}

	if msg.Source != EMPTY {
		buffer.WriteString(COLON)
		buffer.WriteString(msg.Source)
		buffer.WriteString(SPACE)
	}

	if msg.Code > 0 {
		buffer.WriteString(fmt.Sprintf(PADNUM, msg.Code))
	} else if msg.Command != EMPTY {
		buffer.WriteString(msg.Command)
	}

	if len(msg.Params) > 0 {
		if len(msg.Params) > MaxMsgParams {
			msg.Params = msg.Params[0:MaxMsgParams]
		}

		buffer.WriteString(SPACE)
		buffer.WriteString(strings.Join(msg.Params, SPACE))
	}

	if msg.Trailing != EMPTY {
		buffer.WriteString(SPACE)
		buffer.WriteString(COLON)
		buffer.WriteString(msg.Trailing)
	}

	buffer.WriteString(CRLF)

	return buffer
}

// Render returns the IRC-formatted string version of a message object.
// Will return an error if there is something wrong with the meesage.
func (msg *Message) Render() string {
	return msg.RenderBuffer().String()
}

// Debug prints a message object to a string with verbose information about the object fields.
func (msg *Message) Debug() string {
	data, _ := json.Marshal(msg)
	return string(data)
}

func (msg *Message) Reset() {
	clear(msg.Tags)
	msg.Source = ""
	msg.Command = ""
	msg.Code = 0
	msg.Params = nil
	msg.Trailing = ""
}

func escapeTagString(str string) string {
	// Escape tag value
	escapedValue := strings.ReplaceAll(str, SEMICOLON, ESCSEMICOLON)
	escapedValue = strings.ReplaceAll(escapedValue, SPACE, ESCSPACE)
	escapedValue = strings.ReplaceAll(escapedValue, BACKSLAH, ESCBACKSLASH)
	escapedValue = strings.ReplaceAll(escapedValue, CR, ESCCR)
	escapedValue = strings.ReplaceAll(escapedValue, LF, ESCLF)
	return escapedValue
}
