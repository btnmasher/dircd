/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
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
	ErrServerClosed      Error = "Server has closed"
	ErrMessageTooShort   Error = "Did not receive enough data from the client"
	ErrMessageTooLong    Error = "Received data from the client is too long"
	ErrCRLF              Error = "No CRLF"
	ErrWhitespace        Error = "All Whitepace"
	ErrInvalidMessage    Error = "Invalid message format"
	ErrPrefixed          Error = "Prefixed message from client"
	ErrInvalidCapCmd     Error = "Invalid CAP command"
	ErrMissingParams     Error = "Missing parameters"
	ErrTooManyParams     Error = "Too many parameters"
	ErrUserInUse         Error = "This username is currently in use"
	ErrUserRestricted    Error = "This username is restricted"
	ErrUserAreadySet     Error = "You have already registered"
	ErrNickInUse         Error = "This nickname is currently in use"
	ErrNickRestricted    Error = "This nickname is restricted"
	ErrErroneousNickname Error = "This nickname is invalid"
	ErrNickAlreadySet    Error = "You already have that nickname"
	ErrNotImplemented    Error = "That command is not yet implemented"
	ErrNotRegistered     Error = "You must register first"
	ErrNoNickGiven       Error = "No nickname given"
	ErrNoSuchNick        Error = "Nick not found"
	ErrNoSuchChan        Error = "Channel not found"
	ErrInsuffPerms       Error = "Insufficient permissions"
	ErrUnknownMode       Error = "Unknown mode"
	ErrModeAlreadySet    Error = "Mode already set"
	ErrModeNotSet        Error = "Mode is not set"
)

const panicLogTemplate = "panic recovered:\n------ PANIC -----\n%s\n------------------"
