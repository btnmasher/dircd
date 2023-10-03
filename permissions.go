/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

// User permission levels on the server.
const (
	UPermBan uint8 = iota
	UPermNone
	UPermUser
	UPermHelpOp
	UPermNetOp
	UPermAdmin
	UPermServer
)
