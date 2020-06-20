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
	"encoding/json"
	"fmt"
	"strings"
)

// Message is an object that represents the components of an IRC message.
type Message struct {
	Text    string   // The portion of the message after the prefix and command.
	Sender  string   // The sender parameter of the message.
	Params  []string // The person of the message after prefix and command in array form.
	Command string   // The IRC string command of the message.
	Code    uint16   // The IRC numeric code of the message.
}

// Message represents an IRC protocol message.
// See RFC1459 section 2.3.1.
//
//    <message>  = [':' <prefix> <SPACE> ] <command> <params> <crlf>
//    <prefix>   = <servername> | <nick> [ '!' <user> ] [ '@' <host> ]
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
	SPACE  string = " "
	CRLF          = "\r\n"
	COLON         = ":"
	EMPTY         = ""
	PADNUM        = "%03d"
)

// String returns the IRC-formatted string version of a message object.
// This is here to satisfy a Stringer interface
func (msg *Message) String() string {
	return msg.Render()
}

// RenderBuffer returns the IRC-formatted byte buffer version of a message object.
// Will return an error if there is something wrong with the meesage.
func (msg *Message) RenderBuffer() *bytes.Buffer {
	buffer := bufpool.New()

	if msg.Sender != EMPTY {
		buffer.WriteString(COLON)
		buffer.WriteString(msg.Sender)
		buffer.WriteString(SPACE)
	}

	if msg.Code > 0 {
		buffer.WriteString(fmt.Sprintf(PADNUM, msg.Code))
	} else if msg.Command != EMPTY {
		buffer.WriteString(msg.Command)
	}

	if len(msg.Params) > 0 {
		if len(msg.Params) > 14 {
			msg.Params = msg.Params[0:15]
		}

		buffer.WriteString(SPACE)
		buffer.WriteString(strings.Join(msg.Params, SPACE))
	}

	if msg.Text != EMPTY {
		buffer.WriteString(SPACE)
		buffer.WriteString(COLON)
		buffer.WriteString(msg.Text)
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
	bytes, _ := json.Marshal(msg) // Ignoring the error because it literally can't happen.
	return string(bytes)
}

func (msg *Message) scrub() {
	msg.Code = 0
	msg.Command = ""
	msg.Sender = ""
	msg.Params = nil
	msg.Text = ""
}

// MessagePool holds the Message objects in a channel as a queue.
type MessagePool struct {
	Messages chan *Message
}

// NewMessagePool creates a new pool of Messages.
func NewMessagePool(max int) *MessagePool {
	return &MessagePool{
		Messages: make(chan *Message, max),
	}
}

// Warmup fills the MessagePool with the specified number of objects
// up to one below the maximum capacity of the internal channel
func (pool *MessagePool) Warmup(num int) {
	for i := 0; i < num; i++ {
		select {
		case pool.Messages <- &Message{}:
			/* nop */
		default:
			return
		}
	}
}

// New takes a Message from the pool.
func (pool *MessagePool) New() (msg *Message) {
	select {
	case msg = <-pool.Messages:
	default:
		msg = &Message{}
	}
	return
}

// Recycle returns a Message to the pool.
func (pool *MessagePool) Recycle(msg *Message) {
	msg.scrub()
	select {
	case pool.Messages <- msg:
	default:
		// let it go, let it go...
	}
}
