/*
   Copyright (c) 2023, btnmasher
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

// RFC 2812/1459 numerics
const (
	ReplyNone                uint16 = 000
	ReplyWelcome             uint16 = 001
	ReplyYourHost            uint16 = 002
	ReplyCreated             uint16 = 003
	ReplyMyInfo              uint16 = 004
	ReplyISupport            uint16 = 005
	ReplyBounce              uint16 = 010
	ReplyNickForceChanged    uint16 = 043
	ReplyTraceLink           uint16 = 200
	ReplyTraceConnecting     uint16 = 201
	ReplyTraceHandshake      uint16 = 202
	ReplyTraceUnknown        uint16 = 203
	ReplyTraceOperator       uint16 = 204
	ReplyTraceUser           uint16 = 205
	ReplyTraceServer         uint16 = 206
	ReplyTraceService        uint16 = 207
	ReplyTraceNewType        uint16 = 208
	ReplyTraceClass          uint16 = 209
	ReplyStats               uint16 = 210
	ReplyStatsLinkInfo       uint16 = 211
	ReplyStatsCommands       uint16 = 212
	ReplyStatsCLine          uint16 = 213
	ReplyStatsNLine          uint16 = 214
	ReplyStatsILine          uint16 = 215
	ReplyStatsKLine          uint16 = 216
	ReplyStatsQLine          uint16 = 217
	ReplyStatsYLine          uint16 = 218
	ReplyEndOfStats          uint16 = 219
	ReplyUserModeIs          uint16 = 221
	ReplyServiceInfo         uint16 = 231
	ReplyEndOfServices       uint16 = 232
	ReplyServerList          uint16 = 234
	ReplyEndOfServerList     uint16 = 235
	ReplyStatsUptime         uint16 = 242
	ReplyStatsNetOp          uint16 = 243
	ReplyStatsHelpOp         uint16 = 244
	ReplyStatsPing           uint16 = 246
	ReplyUsersOnlineGlobal   uint16 = 251
	ReplyOpersOnline         uint16 = 252
	ReplyUnknownConnections  uint16 = 253
	ReplyChannelCount        uint16 = 254
	ReplyUsersOnlineLocal    uint16 = 255
	ReplyAdminInfoStart      uint16 = 256
	ReplyAdminInfo1          uint16 = 257
	ReplyAdminInfo2          uint16 = 258
	ReplyAdminEmail          uint16 = 259
	ReplyTraceLog            uint16 = 261
	ReplyEndOfTrace          uint16 = 262
	ReplyTryAgain            uint16 = 263
	ReplyAway                uint16 = 301
	ReplyUserHost            uint16 = 302
	ReplyIsOn                uint16 = 303
	ReplyUnAway              uint16 = 305
	ReplyNowAway             uint16 = 306
	ReplyWhoisUser           uint16 = 311
	ReplyWhoisServer         uint16 = 312
	ReplyWhoisOperator       uint16 = 313
	ReplyWhoWasUser          uint16 = 314
	ReplyEndOfWho            uint16 = 315
	ReplyWhoisChanOp         uint16 = 316
	ReplyWhoisIdle           uint16 = 317
	ReplyEndOfWhois          uint16 = 318
	ReplyWhoisChannels       uint16 = 319
	ReplyListStart           uint16 = 321
	ReplyList                uint16 = 322
	ReplyEndOfList           uint16 = 323
	ReplyChannelModeIs       uint16 = 324
	ReplyNoTopic             uint16 = 331
	ReplyChanTopic           uint16 = 332
	ReplyInviting            uint16 = 341
	ReplyInvited             uint16 = 345
	ReplyInviteList          uint16 = 346
	ReplyEndOfInviteList     uint16 = 347
	ReplyExceptList          uint16 = 348
	ReplyEndOfExceptList     uint16 = 349
	ReplyVersion             uint16 = 351
	ReplyWho                 uint16 = 352
	ReplyNames               uint16 = 353
	ReplyLinks               uint16 = 384
	ReplyEndOfLinks          uint16 = 365
	ReplyEndOfNames          uint16 = 366
	ReplyBanList             uint16 = 367
	ReplyEndOfBanList        uint16 = 368
	ReplyEndOfWhoWas         uint16 = 369
	ReplyInfo                uint16 = 371
	ReplyMOTD                uint16 = 372
	ReplyEndOfInfo           uint16 = 374
	ReplyMOTDStart           uint16 = 375
	ReplyEndOFMOTD           uint16 = 376
	ReplyYoureOper           uint16 = 381
	ReplyRehashing           uint16 = 382
	ReplyYoureService        uint16 = 383
	ReplyTime                uint16 = 391
	ReplyUsersStart          uint16 = 392
	ReplyUsers               uint16 = 393
	ReplyEndOfUsers          uint16 = 394
	ReplyNoUsers             uint16 = 395
	ReplyNoSuchNick          uint16 = 401
	ReplyNoSuchServer        uint16 = 402
	ReplyNoSuchChannel       uint16 = 403
	ReplyCannotSendToChan    uint16 = 404
	ReplyTooManyChannels     uint16 = 405
	ReplyWasNoSuchNick       uint16 = 406
	ReplyTooManyTargets      uint16 = 407
	ReplyNoSuchService       uint16 = 408
	ReplyNoOrigin            uint16 = 409
	ReplyInvalidCapCmd       uint16 = 410
	ReplyNoRecipient         uint16 = 411
	ReplyNoTextToSend        uint16 = 412
	ReplyNoTopLevel          uint16 = 413
	ReplyWildTopLevel        uint16 = 414
	ReplyBadMask             uint16 = 415
	ReplyTooManyMatches      uint16 = 416
	ReplyUnknownCommand      uint16 = 421
	ReplyNoMOTD              uint16 = 422
	ReplyNoAdminInfo         uint16 = 423
	ReplyFileError           uint16 = 424
	ReplyNoNicknameGiven     uint16 = 431
	ReplyErroneusNickname    uint16 = 432
	ReplyNicknameInUse       uint16 = 433
	ReplyNickCollision       uint16 = 436
	ReplyResourceUnavailable uint16 = 437
	ReplyUserNotInChannel    uint16 = 441
	ReplyNotOnChannel        uint16 = 442
	ReplyUserOnChannel       uint16 = 443
	ReplyNoLogin             uint16 = 447
	ReplySummonDisabled      uint16 = 446
	ReplyUsersDisabled       uint16 = 446
	ReplyNotRegistered       uint16 = 451
	ReplyNeedMoreParams      uint16 = 461
	ReplyAlreadyRegistered   uint16 = 462
	ReplyNoPermForHost       uint16 = 463
	ReplyPasswordMistmatch   uint16 = 464
	ReplyYoureBanned         uint16 = 465
	ReplyYouWillBeBanned     uint16 = 466
	ReplyChanPassAlreadySet  uint16 = 467
	ReplyChannelIsFull       uint16 = 471
	ReplyUnknownMode         uint16 = 472
	ReplyInviteOnlyChan      uint16 = 473
	ReplyBannedFromChan      uint16 = 474
	ReplyBadChannelPass      uint16 = 475
	ReplyBadChannelName      uint16 = 476
	ReplyNoChanModes         uint16 = 477
	ReplyBanListFUll         uint16 = 478
	ReplyNoPrivileges        uint16 = 481
	ReplyChanOpPrivsNeeded   uint16 = 482
	ReplyCantKillServer      uint16 = 483
	ReplyRestricted          uint16 = 484
	ReplyChanOwnerRequired   uint16 = 485
	ReplyNoOperHost          uint16 = 491
	ReplyNoServiceHost       uint16 = 492
	ReplyUnknownUserMode     uint16 = 501
	ReplyUsersDontMatch      uint16 = 502
	ReplyLoggedIn            uint16 = 900
	ReplyLoggedOut           uint16 = 901
	ReplySASLSuccess         uint16 = 903
	ReplySASLFail            uint16 = 904
	ReplySASLTooLong         uint16 = 905
	ReplySASLAborted         uint16 = 906
	ReplySASLAlready         uint16 = 907
)
