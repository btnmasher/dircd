/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

import (
	"bytes"
	"sync"

	"github.com/btnmasher/dircd/shared/concurrentmap"
)

// User holds all the state in the context of a connected user.
type User struct {
	mu sync.RWMutex

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

type UserMap concurrentmap.ConcurrentMap[string, *User]

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
// It will return the Vanity hostname insteead of the regular hostname if
// VanityEnabled is set to true, and the VanityHost is set in the User object.
//
// <nick>!<username>@<hostname|vanityhost>
func (user *User) Hostmask() string {
	user.mu.RLock()
	defer user.mu.RUnlock()
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
// It will not return the Vanity hostname even if VanityEnabled is set to true.
//
// <nick>!<username>@<hostname>
func (user *User) RealHostmask() string {
	user.mu.RLock()
	defer user.mu.RUnlock()
	var buffer bytes.Buffer

	buffer.WriteString(user.nick)
	buffer.WriteString("!")
	buffer.WriteString(user.name)
	buffer.WriteString("@")
	buffer.WriteString(user.host)

	return buffer.String()
}

// Nick returns the nick field of the user in a concurrency-safe manner
func (user *User) Nick() string {
	user.mu.RLock()
	defer user.mu.RUnlock()
	return user.nick
}

// SetNick sets the nick field of the user in a concurrency-safe manner
func (user *User) SetNick(new string) {
	user.mu.Lock()
	defer user.mu.Unlock()
	user.nick = new
}

// Name returns the username field of the user in a concurrency-safe manner
func (user *User) Name() string {
	user.mu.RLock()
	defer user.mu.RUnlock()
	return user.name
}

// SetName sets the username field of the user in a concurrency-safe manner
func (user *User) SetName(new string) {
	user.mu.Lock()
	defer user.mu.Unlock()
	user.name = new
}

// Realname returns the realname field of the user in a concurrency-safe manner
func (user *User) Realname() string {
	user.mu.RLock()
	defer user.mu.RUnlock()
	return user.real
}

// SetRealname sets the realname field of the user in a concurrency-safe manner
func (user *User) SetRealname(new string) {
	user.mu.Lock()
	defer user.mu.Unlock()
	user.real = new
}

// SetHostname sets the hostname field of the user in a concurrency-safe manner
func (user *User) SetHostname(new string) {
	user.mu.Lock()
	defer user.mu.Unlock()
	user.host = new
}

// VanityHost returns the vanityhost field of the user in a concurrency-safe manner
func (user *User) VanityHost() string {
	user.mu.RLock()
	defer user.mu.RUnlock()
	return user.vanityHost
}

// SetVanityHost sets the vanityhost field of the user in a
// concurrency-safe manner.
func (user *User) SetVanityHost(new string) {
	user.mu.Lock()
	defer user.mu.Unlock()
	user.vanityHost = new
}

// Permission returns the permission field of the user in a
// concurrency-safe manner.
func (user *User) Permission() uint8 {
	user.mu.RLock()
	defer user.mu.RUnlock()
	return user.perm
}

// SetPermission the permission field of the user in a
// concurrency-safe manner.
func (user *User) SetPermission(new uint8) {
	user.mu.Lock()
	defer user.mu.Unlock()
	user.perm = new
}

// Mode returns the mode field of the user in a
// concurrency-safe manner.
func (user *User) Mode() uint64 {
	user.mu.RLock()
	defer user.mu.RUnlock()
	return user.mode
}

// AddMode appends the specified mode flag to the user in a
// concurrency-safe manner.
func (user *User) AddMode(umode uint64) {
	user.mu.Lock()
	defer user.mu.Unlock()
	user.mode |= umode
}

// DelMode removes the specified mode flag from the user in a
// concurrency-safe manner.
func (user *User) DelMode(umode uint64) {
	user.mu.Lock()
	defer user.mu.Unlock()
	user.mode &^= umode
}

// ModeIsSet checks if a given user mode is currently
// set in a concurrency-safe manner.
func (user *User) ModeIsSet(umode uint64) bool {
	user.mu.Lock()
	defer user.mu.Unlock()
	return user.mode&umode == umode
}

// VanityEnabled returns the vanityenabled field of the user in a
// concurrency-safe manner.
func (user *User) VanityEnabled() bool {
	user.mu.RLock()
	defer user.mu.RUnlock()
	return user.vanityEnabled
}

// SetVanityEnabled the vanityenabled field of the user in a
// concurrency-safe manner.
func (user *User) SetVanityEnabled(new bool) {
	user.mu.Lock()
	defer user.mu.Unlock()
	user.vanityEnabled = new
}

// HigherPerms checks if the given target User has a higher
// permission level than the Given user being checked.
func (user *User) HigherPerms(target uint8) bool {
	user.mu.RLock()
	defer user.mu.RUnlock()
	return user.perm > target
}
