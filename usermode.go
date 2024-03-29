/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

import "strings"

// Usermode Bitmasks
const (
	UModeAway uint64 = 1 << iota
	UModeAdmin
	UModeBot
	UModeBanned
	UModeCensored
	UModeConnInfo
	UModeDeaf
	UModeDebug
	UModeFloodInfo
	UModeFloodImmune
	UModeGodmode
	UModeHiddenHost
	UModeHidden
	UModeInvisible
	UModeImmune
	UModeKeyMaster
	UModeMuted
	UModeHelpOp
	UModeNetOp
	UModeProtected
	UModeRegistered
	UModeSecured
	UModeThrottled
	UModeGlobalVoice
	UModeWhoisInfo
	UModeWatch
)

// UModeReq is used to define the required setter/target permission levels.
type UModeReq struct {
	Setter uint8
	Target uint8
}

// uModeReqs is a map of usermodes with required setter/target permissions levels defined.
var uModeReqs = map[uint64]UModeReq{
	UModeAway:        {UPermUser, UPermUser},
	UModeAdmin:       {UPermServer, UPermUser},
	UModeBot:         {UPermNetOp, UPermUser},
	UModeBanned:      {UPermNetOp, UPermNone},
	UModeCensored:    {UPermHelpOp, UPermUser},
	UModeConnInfo:    {UPermAdmin, UPermNetOp},
	UModeDeaf:        {UPermNetOp, UPermUser},
	UModeDebug:       {UPermAdmin, UPermNetOp},
	UModeFloodInfo:   {UPermNetOp, UPermHelpOp},
	UModeFloodImmune: {UPermNetOp, UPermUser},
	UModeGodmode:     {UPermServer, UPermAdmin},
	UModeHiddenHost:  {UPermHelpOp, UPermUser},
	UModeHidden:      {UPermNetOp, UPermHelpOp},
	UModeInvisible:   {UPermNetOp, UPermHelpOp},
	UModeImmune:      {UPermNetOp, UPermUser},
	UModeKeyMaster:   {UPermNetOp, UPermHelpOp},
	UModeMuted:       {UPermHelpOp, UPermUser},
	UModeHelpOp:      {UPermNetOp, UPermUser},
	UModeNetOp:       {UPermAdmin, UPermUser},
	UModeProtected:   {UPermNetOp, UPermUser},
	UModeRegistered:  {UPermServer, UPermUser},
	UModeSecured:     {UPermServer, UPermUser},
	UModeThrottled:   {UPermHelpOp, UPermUser},
	UModeWhoisInfo:   {UPermUser, UPermUser},
	UModeWatch:       {UPermNetOp, UPermHelpOp},
}

// SetUserMode is  used to set a mode on a target user.
//
// This function will lock both setter and target user mutexes.
//
// First it determines if a user mode is valid. If this is not the case,
// this function will return ErrUnknownMode
//
// Then it will then determine if the permission level of the setting user is higher
// than the target user, as well as if the target user's permission level is
// defined as being allowed to receive the specified usermode. If both are true
// then the mode will be set. Otherwise, this function will return ErrInsuffPerms
//
// If the mode is already present on the user, then this function will return
// ErrModeAlreadySet
func SetUserMode(umode uint64, setter, target *User) error {
	setter.mu.RLock()
	target.mu.Lock()
	defer setter.mu.RUnlock()
	defer target.mu.Unlock()

	reqs, exists := uModeReqs[umode]
	if !exists {
		return ErrUnknownMode
	}

	// Check if setter has required permission to set the specified mode,
	// target has required permission to receive the mode, and if the
	// setter has a higher permission than the target or if the target
	// is also the setter.
	canSet := setter.perm >= reqs.Setter
	canReceive := target.perm >= reqs.Target
	higherPerm := setter.perm > target.perm
	sameUser := strings.ToLower(setter.nick) == strings.ToLower(target.nick)

	if canSet && canReceive && (higherPerm || sameUser) {
		if target.mode&umode == umode { // Check if mode flag already set
			return ErrModeAlreadySet
		}

		target.mode |= umode // Set the mode

	} else {
		return ErrInsuffPerms
	}

	return nil
}

// UnsetUserMode is used to unset a mode on a target user.
//
// This function will lock both setter and target user mutexes.
//
// First it determines if a user mode is valid. If this is not the case,
// this function will return ErrUnknownMode
//
// Then it will then determine if the permission level of the setting user is higher
// than the target user. If this is true, then the mode will be set. Otherwise,
// this function will return ErrInsuffPerms
//
// If the mode is not already present on the user, then this function will return
// ErrModeNotSet
func UnsetUserMode(umode uint64, setter, target *User) error {
	setter.mu.RLock()
	target.mu.Lock()
	defer setter.mu.RUnlock()
	defer target.mu.Unlock()
	// TODO: figure out these exported mutexes

	reqs, exists := uModeReqs[umode]
	if !exists {
		return ErrUnknownMode
	}

	// Check if setter has required permission to set the specified mode,
	// target has required permission to receive the mode, and if the
	// setter has a higher permission than the target or if the target
	// is also the setter.
	if setter.perm >= reqs.Setter &&
		(setter.perm > target.perm ||
			strings.ToLower(setter.nick) == strings.ToLower(target.nick)) {
		if target.mode&umode != umode { // Check if mode flag already unset
			return ErrModeNotSet
		}

		target.mode &^= umode // Unset the mode

	} else {
		return ErrInsuffPerms
	}

	return nil
}
