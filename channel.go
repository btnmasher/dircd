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
	"sync"

	"github.com/btnmasher/util"
)

// Channel represents an IRC channel
type Channel struct {
	sync.RWMutex

	name  string
	topic string

	modes uint64

	owner      *User
	savedOwner string // Owner username

	// Active Lists
	Nicks   *UserMap
	Ops     *UserMap
	HalfOps *UserMap
	Voiced  *UserMap

	// Persisted Lists
	// map[hostpattern]setter
	OpList     *util.ConcurrentMapString
	HalfOpList *util.ConcurrentMapString
	VoiceList  *util.ConcurrentMapString
	BanList    *util.ConcurrentMapString
	InviteList *util.ConcurrentMapString
}

// NewChannel initializes a Channel with the given name and owner.
func NewChannel(cname string, creator *User) *Channel {
	channel := &Channel{
		name:       cname,
		owner:      creator,
		Nicks:      NewUserMap(),
		Ops:        NewUserMap(),
		HalfOps:    NewUserMap(),
		Voiced:     NewUserMap(),
		OpList:     util.NewConcurrentMapString(),
		HalfOpList: util.NewConcurrentMapString(),
		VoiceList:  util.NewConcurrentMapString(),
		BanList:    util.NewConcurrentMapString(),
		InviteList: util.NewConcurrentMapString(),
	}

	return channel
}

// Name returns the name of the channel in a currency safe manner.
func (channel *Channel) Name() string {
	channel.RLock()
	defer channel.RUnlock()

	return channel.name
}

// SetName sets the name of the channel in a currency safe manner.
func (channel *Channel) SetName(new string) {
	channel.Lock()
	defer channel.Unlock()

	channel.name = new
}

// Topic returns the topic of the channel in a currency safe manner.
func (channel *Channel) Topic() string {
	channel.RLock()
	defer channel.RUnlock()

	return channel.topic
}

// SetTopic sets the topic of the channel in a currency safe manner.
func (channel *Channel) SetTopic(new string) {
	channel.Lock()
	defer channel.Unlock()

	channel.topic = new
}

// Owner returns the owner of the channel in a currency safe manner.
func (channel *Channel) Owner() *User {
	channel.RLock()
	defer channel.RUnlock()

	return channel.owner
}

// SetOwner sets the owner of the channel in a currency safe manner.
func (channel *Channel) SetOwner(new *User) {
	channel.Lock()
	defer channel.Unlock()

	channel.owner = new
	channel.savedOwner = new.Name()
}

// TODO: channel modes

// Send takes a message, then iterates the list of Users joined
// to the channel stored in the Nicks map, and sends the message
// to each of the User's underlying connection.
func (channel *Channel) Send(msg *Message, exclude string) {
	channel.RLock()
	defer channel.RUnlock()

	// TODO: Check if sender is allowed to send

	buf := msg.RenderBuffer()

	channel.Nicks.ForEach(func(user *User) {
		if user.Nick() != exclude {
			user.conn.Write(buf)
		}
	})
}

// Join adds the user to the channel and alerts all channel
// members of the event.
func (channel *Channel) Join(user *User, msg *Message) bool {
	channel.RLock()
	defer channel.RUnlock()

	// TODO: Check if sender is allowed to send

	channel.Nicks.Add(user.Nick(), user)
	channel.Send(msg, "")

	return true
}

// Part removes the user from the channel and alerts all channel
// members of the event.
func (channel *Channel) Part(user *User, msg *Message) {
	channel.Send(msg, "")
	channel.Nicks.Del(user.Nick())
}

// GetNicks returns an array of the current nicknames of the users
// in the chanel.
func (channel *Channel) GetNicks() []string {
	channel.RLock()
	defer channel.RUnlock()

	var buffer bytes.Buffer
	nicks := make([]string, channel.Nicks.Length())
	i := 0

	channel.Nicks.ForEach(func(user *User) {
		nick := user.Nick()

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

		nicks[i] = buffer.String()
		buffer.Reset()
		i++

	})

	return nicks
}
