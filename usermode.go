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

// UModeReqs is a map of usermodes with required setter/target permissions levels defined.
var UModeReqs = map[uint64]UModeReq{
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
	setter.Lock()
	target.Lock()
	defer setter.Unlock()
	defer target.Unlock()

	reqs, exists := UModeReqs[umode]
	if !exists {
		return ErrUnknownMode
	}

	// Check if setter has required permission to set the specified mode,
	// target has required permission to receive the mode, and if the
	// setter has a higher permission than the target or if the target
	// is also the setter.
	if setter.perm >= reqs.Setter &&
		target.perm >= reqs.Target &&
		(setter.perm > target.perm ||
			strings.ToLower(setter.nick) == strings.ToLower(target.nick)) {
		if target.mode&umode == umode { // Check if mode flag already set
			return ErrModeAlreadySet
		}

		target.mode |= umode // Set the mode

	} else {
		return ErrInsuffPerms
	}

	return nil
}

// UnsetUserMode is  used to unset a mode on a target user.
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
	setter.Lock()
	target.Lock()
	defer setter.Unlock()
	defer target.Unlock()

	reqs, exists := UModeReqs[umode]
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
