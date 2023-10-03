/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

// Limiter Constants
const (
	// Messages
	MaxMsgLength  int = 512
	MaxMsgParams      = 15
	MaxTagsLength int = 4096

	// Channels
	MaxChanLength  = 16
	MaxKickLength  = 400
	MaxTopicLength = 400
	MaxListItems   = 256
	MaxModeChange  = 6

	// Users
	MaxNickLength  = 16
	MaxUserLength  = 16
	MaxVHostLength = 64
	MaxJoinedChans = 32
	MaxAwayLength  = 100
)
