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

// Error is a workaround to allow for immutable error strings
// which satisfy the error interface.
type Error string

func (err Error) Error() string {
	return string(err)
}

func (err Error) String() string {
	return string(err)
}

// Immutable error strings
const (
	ErrNotEnoughData  Error = "Did not receive enough data from the client"
	ErrDataTooLong    Error = "Received data from the client is too long"
	ErrCRLF           Error = "No CRLF"
	ErrWhitespace     Error = "All Whitepace"
	ErrPrefixed       Error = "Prefixed message from client"
	ErrInvalidCapCmd  Error = "Invalid CAP command"
	ErrMissingParams  Error = "Missing parameters"
	ErrTooManyParams  Error = "Too many parameters"
	ErrUserInUse      Error = "This username is currently in use"
	ErrUserRestricted Error = "This username is restricted"
	ErrUserAreadySet  Error = "You have already registered"
	ErrNickInUse      Error = "This nickname is currently in use"
	ErrNickRestricted Error = "This nickname is restricted"
	ErrNickAlreadySet Error = "You already have that nickname"
	ErrNotImplemented Error = "That command is not yet implemented"
	ErrNotRegistered  Error = "You must register first"
	ErrNoNickGiven    Error = "No nickname given"
	ErrNoSuchNick     Error = "Nick not found"
	ErrNoSuchChan     Error = "Channel not found"
	ErrInsuffPerms    Error = "Insufficient permissions"
	ErrUnknownMode    Error = "Unknown mode"
	ErrModeAlreadySet Error = "Mode already set"
	ErrModeNotSet     Error = "Mode is not set"
)
