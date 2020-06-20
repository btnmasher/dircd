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
)

// User holds all of the state in the context of a connected user.
type User struct {
	sync.RWMutex

	nick          string
	name          string
	host          string
	real          string
	vanityHost    string
	vanityEnabled bool
	perm          uint8
	mode          uint64

	conn *Conn
}

// // NewUser returns a new instance of a user object with the given parameters
// func NewUser(nickname, username, realname, hostname string) *User {
// 	user := User{
// 		nick: nickname,
// 		name: username,
// 		real: realname,
// 		host: hostname,
// 		perm: UPermUser,
// 	}
// 	return &user
// }

// Hostmask returns the string form of the full IRC hostmask.
// It will return the Vanity hostname insteead of the regular
// hostname if VanityEnabled is set to true, and the VanityHost
// is set in the User object.
//
// <nick>!<username>@<hostname|vanityhost>
func (user *User) Hostmask() string {
	user.RLock()
	defer user.RUnlock()
	var buffer bytes.Buffer

	buffer.WriteString(user.nick)
	buffer.WriteString("!")
	buffer.WriteString(user.name)
	buffer.WriteString("@")

	if user.vanityEnabled && len(user.vanityHost) > 0 {
		buffer.WriteString(user.vanityHost)
	} else {
		buffer.WriteString(user.host)
	}

	return buffer.String()
}

// RealHostmask returns the string form of the full IRC hostmask.
// It will not return the Vanity hostname even if VanityEnabled
// is set to true.
//
// <nick>!<username>@<hostname>
func (user *User) RealHostmask() string {
	user.RLock()
	defer user.RUnlock()
	var buffer bytes.Buffer

	buffer.WriteString(user.nick)
	buffer.WriteString("!")
	buffer.WriteString(user.name)
	buffer.WriteString("@")
	buffer.WriteString(user.host)

	return buffer.String()
}

// Nick returns the nick field of the user in a
// concurrency-safe manner.
func (user *User) Nick() string {
	user.RLock()
	defer user.RUnlock()
	return user.nick
}

// SetNick sets the nick field of the user in a
// concurrency-safe manner.
func (user *User) SetNick(new string) {
	user.Lock()
	defer user.Unlock()
	user.nick = new
}

// Name returns the username field of the user in a
// concurrency-safe manner.
func (user *User) Name() string {
	user.RLock()
	defer user.RUnlock()
	return user.name
}

// SetName sets the username field of the user in a
// concurrency-safe manner.
func (user *User) SetName(new string) {
	user.Lock()
	defer user.Unlock()
	user.name = new
}

// Realname returns the realname field of the user in a
// concurrency-safe manner.
func (user *User) Realname() string {
	user.RLock()
	defer user.RUnlock()
	return user.real
}

// SetRealname sets the realname field of the user in a
// concurrency-safe manner.
func (user *User) SetRealname(new string) {
	user.Lock()
	defer user.Unlock()
	user.real = new
}

// SetHostname sets the hostname field of the user in a
// concurrency-safe manner.
func (user *User) SetHostname(new string) {
	user.Lock()
	defer user.Unlock()
	user.host = new
}

// VanityHost returns the vanityhost field of the user in a
// concurrency-safe manner.
func (user *User) VanityHost() string {
	user.RLock()
	defer user.RUnlock()
	return user.vanityHost
}

// SetVanityHost sets the vanityhost field of the user in a
// concurrency-safe manner.
func (user *User) SetVanityHost(new string) {
	user.Lock()
	defer user.Unlock()
	user.vanityHost = new
}

// Permission returns the permission field of the user in a
// concurrency-safe manner.
func (user *User) Permission() uint8 {
	user.RLock()
	defer user.RUnlock()
	return user.perm
}

// SetPermission the permission field of the user in a
// concurrency-safe manner.
func (user *User) SetPermission(new uint8) {
	user.Lock()
	defer user.Unlock()
	user.perm = new
}

// Mode returns the mode field of the user in a
// concurrency-safe manner.
func (user *User) Mode() uint64 {
	user.RLock()
	defer user.RUnlock()
	return user.mode
}

// AddMode appends the specified mode flag to the user in a
// concurrency-safe manner.
func (user *User) AddMode(umode uint64) {
	user.Lock()
	defer user.Unlock()
	user.mode |= umode
}

// DelMode removes the specified mode flag from the user in a
// concurrency-safe manner.
func (user *User) DelMode(umode uint64) {
	user.Lock()
	defer user.Unlock()
	user.mode &^= umode
}

// ModeIsSet checks if a given user mode is currently
// set in a concurrency-safe manner.
func (user *User) ModeIsSet(umode uint64) bool {
	user.Lock()
	defer user.Unlock()
	return (user.mode&umode == umode)
}

// VanityEnabled returns the vanityenabled field of the user in a
// concurrency-safe manner.
func (user *User) VanityEnabled() bool {
	user.RLock()
	defer user.RUnlock()
	return user.vanityEnabled
}

// SetVanityEnabled the vanityenabled field of the user in a
// concurrency-safe manner.
func (user *User) SetVanityEnabled(new bool) {
	user.Lock()
	defer user.Unlock()
	user.vanityEnabled = new
}

// HigherPerms checks if the given target User has a higher
// permission level than the Given user being checked.
func (user *User) HigherPerms(target uint8) bool {
	user.RLock()
	defer user.RUnlock()
	return user.perm > target
}
