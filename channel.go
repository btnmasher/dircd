/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/btnmasher/dircd/shared/safemap"
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
	OpList     safemap.SafeMap[string, string]
	HalfOpList safemap.SafeMap[string, string]
	VoiceList  safemap.SafeMap[string, string]
	BanList    safemap.SafeMap[string, string]
	InviteList safemap.SafeMap[string, string]
}

type ChanMap safemap.SafeMap[string, *Channel]

// NewChannel initializes a Channel with the given name and owner.
func NewChannel(cname string, creator *User) *Channel {
	channel := &Channel{
		name:       cname,
		owner:      creator,
		Nicks:      safemap.NewMutexMap[string, *User](),
		Ops:        safemap.NewMutexMap[string, *User](),
		HalfOps:    safemap.NewMutexMap[string, *User](),
		Voiced:     safemap.NewMutexMap[string, *User](),
		OpList:     safemap.NewMutexMap[string, string](),
		HalfOpList: safemap.NewMutexMap[string, string](),
		VoiceList:  safemap.NewMutexMap[string, string](),
		BanList:    safemap.NewMutexMap[string, string](),
		InviteList: safemap.NewMutexMap[string, string](),
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
