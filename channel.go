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

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/btnmasher/dircd/shared/concurrentmap"
)

// Channel represents an IRC channel
type Channel struct {
	mu sync.RWMutex

	name  string
	topic string

	modes uint64

	owner      *User
	savedOwner string // Owner username

	// Active Lists
	Nicks   UserMap
	Ops     UserMap
	HalfOps UserMap
	Voiced  UserMap

	// Persisted Lists
	// map[hostPattern]setter
	OpList     concurrentmap.ConcurrentMap[string, string]
	HalfOpList concurrentmap.ConcurrentMap[string, string]
	VoiceList  concurrentmap.ConcurrentMap[string, string]
	BanList    concurrentmap.ConcurrentMap[string, string]
	InviteList concurrentmap.ConcurrentMap[string, string]
}

type ChanMap concurrentmap.ConcurrentMap[string, *Channel]

// NewChannel initializes a Channel with the given name and owner.
func NewChannel(cname string, creator *User) *Channel {
	channel := &Channel{
		name:       cname,
		owner:      creator,
		Nicks:      concurrentmap.New[string, *User](),
		Ops:        concurrentmap.New[string, *User](),
		HalfOps:    concurrentmap.New[string, *User](),
		Voiced:     concurrentmap.New[string, *User](),
		OpList:     concurrentmap.New[string, string](),
		HalfOpList: concurrentmap.New[string, string](),
		VoiceList:  concurrentmap.New[string, string](),
		BanList:    concurrentmap.New[string, string](),
		InviteList: concurrentmap.New[string, string](),
	}

	return channel
}

// Name returns the name of the channel in a currency safe manner
func (channel *Channel) Name() string {
	channel.mu.RLock()
	defer channel.mu.RUnlock()

	return channel.name
}

// SetName sets the name of the channel in a currency safe manner
func (channel *Channel) SetName(new string) {
	channel.mu.Lock()
	defer channel.mu.Unlock()

	channel.name = new
}

// Topic returns the topic of the channel in a currency safe manner
func (channel *Channel) Topic() string {
	channel.mu.RLock()
	defer channel.mu.RUnlock()

	return channel.topic
}

// SetTopic sets the topic of the channel in a currency safe manner
func (channel *Channel) SetTopic(new string) {
	channel.mu.Lock()
	defer channel.mu.Unlock()

	channel.topic = new
}

// Owner returns the owner of the channel in a currency safe manner
func (channel *Channel) Owner() *User {
	channel.mu.RLock()
	defer channel.mu.RUnlock()

	return channel.owner
}

// SetOwner sets the owner of the channel in a currency safe manner
func (channel *Channel) SetOwner(new *User) {
	channel.mu.Lock()
	defer channel.mu.Unlock()

	channel.owner = new
	channel.savedOwner = new.Name()
}

// TODO: channel modes

// Send takes a message, then iterates the list of Users joined to the channel stored
// in the Nicks map, and sends the message to each of the User's underlying connection.
func (channel *Channel) Send(msg *Message, exclude string) {
	channel.mu.RLock()
	defer channel.mu.RUnlock()

	// TODO: Check if sender is allowed to send

	buf := msg.RenderBuffer()

	_ = channel.Nicks.ForEach(func(nick string, user *User) error {
		if nick != exclude {
			user.conn.Write(buf)
		}
		return nil
	})
}

// Join adds the user to the channel and alerts all channel members of the event.
func (channel *Channel) Join(user *User, msg *Message) bool {
	channel.mu.RLock()
	defer channel.mu.RUnlock()

	// TODO: Check if sender is allowed to send

	channel.Nicks.Set(user.Nick(), user)
	channel.Send(msg, "")

	return true
}

// RemoveUser removes the user from the channel and alerts all channel members of the event.
func (channel *Channel) RemoveUser(nick string, msg *Message) error {
	if !channel.Nicks.Exists(nick) {
		return fmt.Errorf("nick [%s] not present in channel [%s]", nick, channel.Name())
	}
	channel.Send(msg, nick)
	channel.Nicks.Delete(nick)
	channel.Ops.Delete(nick)
	channel.HalfOps.Delete(nick)
	channel.Voiced.Delete(nick)
	return nil
}

func (channel *Channel) ChangeNick(oldNick, newNick string, msg *Message) error {
	if !channel.Nicks.ChangeKey(oldNick, newNick) {
		return errors.New("nickname not present in channel")
	}
	channel.Ops.ChangeKey(oldNick, newNick)
	channel.HalfOps.ChangeKey(oldNick, newNick)
	channel.Voiced.ChangeKey(oldNick, newNick)
	channel.Send(msg, newNick)

	return nil
}

// GetNicks returns an array of the current nicknames of the users in the channel.
func (channel *Channel) GetNicks() []string {
	var buffer bytes.Buffer
	nicks := make([]string, 0, channel.Nicks.Length())

	channel.Nicks.ForEach(func(nick string, user *User) error {
		switch {
		case channel.owner.Nick() == nick:
			buffer.WriteRune('~')
		case channel.Ops.Exists(nick):
			buffer.WriteRune('@')
		case channel.HalfOps.Exists(nick):
			buffer.WriteRune('%')
		case channel.Voiced.Exists(nick):
			buffer.WriteRune('+')
		}

		buffer.WriteString(nick)

		nicks = append(nicks, buffer.String())
		buffer.Reset()
		return nil
	})

	return nicks
}
