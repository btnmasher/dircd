/*
   Copyright (c) 2023, btnmasher
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

import "strings"

// Parse takes IRC-formatted text into a message object.
// Will return an error if the message doesn't fit the protocol.
func Parse(data string) (*Message, error) {
	if len(data) < 4 {
		return nil, ErrNotEnoughData
	}

	if len(data) > MaxMsgLength {
		return nil, ErrDataTooLong
	}

	// if data[len(data)-2:] != CRLF {
	// 	// return nil, fmt.Errorf("No CRLF")
	// 	return nil, ErrCRLF
	// }

	data = strings.TrimSpace(data)
	if len(data) == 0 {
		return nil, ErrWhitespace
	}

	if data[0] == ':' { // Clients shouldn't be sending prefixed messages so we're going to just error
		// split := strings.SplitN(data, " ", 2)
		// if len(split) < 2 {
		return nil, ErrPrefixed
		// }
		// data = split[1]
	}

	msg := msgPool.New()

	split := strings.SplitN(data, ":", 2)
	args := strings.Fields(split[0])

	msg.Command = strings.ToUpper(args[0])

	msg.Params = args[1:]

	if len(msg.Params) > MaxMsgParams {
		return nil, ErrTooManyParams
	}

	if len(split) > 1 {
		msg.Text = split[1]
	}

	return msg, nil
}
