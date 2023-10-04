/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

import "strings"

// Parse takes IRC-formatted text into a message object.
// Will return an error if the message doesn't fit the protocol.
func Parse(line string) (*Message, error) {
	if len(line) < 4 {
		return nil, ErrMessageTooShort
	}

	if len(line) > MaxTagsLength+MaxMsgLength {
		return nil, ErrMessageTooLong
	}

	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return nil, ErrWhitespace
	}

	if line[0] == ':' { // Clients shouldn't be sending prefixed messages, so we're going to just error
		return nil, ErrPrefixed
	}

	msg := msgPool.New()

	if line[0] == '@' {
		parts := strings.SplitN(line[1:], SPACE, 2)
		if len(parts) < 2 {
			return nil, ErrInvalidMessage
		}
		rawTags := parts[0]
		line = parts[1] // the remainder of the message for later processing

		// Parse tags
		tags := strings.Split(rawTags, SEMICOLON)
		msg.Tags = make(map[string]string)
		for _, tag := range tags {
			kv := strings.SplitN(tag, EQUAL, 2)
			if len(kv) == 2 {
				msg.Tags[kv[0]] = kv[1]
			} else {
				msg.Tags[kv[0]] = EMPTY
			}
		}
	}

	if len(line) > MaxMsgLength {
		msgPool.Recycle(msg)
		return nil, ErrMessageTooLong
	}

	if line[0] == ':' { // Clients shouldn't be sending prefixed messages, so we're going to just error
		msgPool.Recycle(msg)
		return nil, ErrPrefixed
	}

	split := strings.SplitN(line, COLON, 2)
	args := strings.Fields(split[0])

	msg.Command = strings.ToUpper(args[0])

	if len(split) > 1 {
		msg.Trailing = split[1]
	}

	if len(args) > 1 {
		msg.Params = args[1:]
		msg.Trailing = args[len(args)-1]

		if len(msg.Params) > MaxMsgParams {
			msgPool.Recycle(msg)
			return nil, ErrTooManyParams
		}
	}

	return msg, nil
}
