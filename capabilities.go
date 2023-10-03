/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package dircd

// TODO: bitmasks :)

// Capabilities contains the information for CAP negotiation.
type Capabilities struct {
	AccountNotify   bool // Notifies clients when other clients in comon channels authenticate or deauthenticate (eg: NickServ, SASL).
	AccountTag      bool // Attach a tag containing the user's account to every message they send.
	AwayNotify      bool // Notifies clients when other clients in common channels go away or come back.
	Batch           bool // Allow server to bundle common messages together.
	CapNotify       bool // Notify when capabilties become available or are no longer available.
	ChgHost         bool // Enable CHGHOST message, which lets servers notify clients when another cleint's username and/or hostname changes.
	EchoMessage     bool // Notifies clients when their PRIVMSG and NOTICEs are correctly received by the server.
	ExtendedJoin    bool // Extends the JOIN message to include the account name of the joining client.
	InviteNotify    bool // Notifies clients when other clients are invited to common channels.
	LabeledResponse bool // Allows clients to correlate requests with server responses.
	MessageTags     bool // Allows Clients and servers to use tags more broadly.
	Metadata        bool // Lets clients store metadata about themselves with the server, for other clients to request and retrieve later.
	Monitor         bool // Lets users request notifications for qhen clients become online/offline.
	MultiPrefix     bool // Makes the server send all prefixes in NAMES and WHO output, in order of rank from highest to lowest.
	Multiline       bool // Allows clients and servers to use send messages that can exceed the usual byte length limit and that can contain line breaks.
	SASL            bool // Indicates support for SASL authentication.
	ServerTime      bool // Lets clients show the actual time messages were received by the server.
	Setname         bool // Lets clients change their realname after connecting to the server.
	TLS             bool // Indicates support for the STARTTLS command, which lets clients upgrade their connection to use TLS encryption.
	UserhostInNames bool // Extends the NAMEREPLY message to contain the full nickmask (nick!user@host) of every user, rather than just the nickname.
}

// Capabilities bitmask flags for CAP negotiation.
const (
	AccountNotify   int = 1 << iota // Notifies clients when other clients in comon channels authenticate or deauthenticate (eg: NickServ, SASL).
	AccountTag                      // Attach a tag containing the user's account to every message they send.
	AwayNotify                      // Notifies clients when other clients in common channels go away or come back.
	Batch                           // Allow server to bundle common messages together.
	CapNotify                       // Notify when capabilties become available or are no longer available.
	ChgHost                         // Enable CHGHOST message, which lets servers notify clients when another cleint's username and/or hostname changes.
	EchoMessage                     // Notifies clients when their PRIVMSG and NOTICEs are correctly received by the server.
	ExtendedJoin                    // Extends the JOIN message to include the account name of the joining client.
	InviteNotify                    // Notifies clients when other clients are invited to common channels.
	LabeledResponse                 // Allows clients to correlate requests with server responses.
	MessageTags                     // Allows Clients and servers to use tags more broadly.
	Metadata                        // Lets clients store metadata about themselves with the server, for other clients to request and retrieve later.
	Monitor                         // Lets users request notifications for qhen clients become online/offline.
	MultiPrefix                     // Makes the server send all prefixes in NAMES and WHO output, in order of rank from highest to lowest.
	Multiline                       // Allows clients and servers to use send messages that can exceed the usual byte length limit and that can contain line breaks.
	SASL                            // Indicates support for SASL authentication.
	ServerTime                      // Lets clients show the actual time messages were received by the server.
	Setname                         // Lets clients change their realname after connecting to the server.
	TLS                             // Indicates support for the STARTTLS command, which lets clients upgrade their connection to use TLS encryption.
	UserhostInNames                 // Extends the NAMEREPLY message to contain the full nickmask (nick!user@host) of every user, rather than just the nickname.
)

// SASL Types
const (
	SaslPlain uint8 = iota
	SaslLogin
	SaslExternal
	SaslGSSAPI
	SaslCramMD5
	SaslDigestMD5
	SaslScramSHA1
)
