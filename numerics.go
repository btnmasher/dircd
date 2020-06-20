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

// RFC 2812/1459 numerics
const (
	ReplyNone                uint16 = 000
	ReplyWelcome                    = 001
	ReplyYourHost                   = 002
	ReplyCreated                    = 003
	ReplyMyInfo                     = 004
	ReplyISupport                   = 005
	ReplyBounce                     = 010
	ReplyNickForceChanged           = 043
	ReplyTraceLink                  = 200
	ReplyTraceConnecting            = 201
	ReplyTraceHandshake             = 202
	ReplyTraceUnknown               = 203
	ReplyTraceOperator              = 204
	ReplyTraceUser                  = 205
	ReplyTraceServer                = 206
	ReplyTraceService               = 207
	ReplyTraceNewType               = 208
	ReplyTraceClass                 = 209
	ReplyStats                      = 210
	ReplyStatsLinkInfo              = 211
	ReplyStatsCommands              = 212
	ReplyStatsCLine                 = 213
	ReplyStatsNLine                 = 214
	ReplyStatsILine                 = 215
	ReplyStatsKLine                 = 216
	ReplyStatsQLine                 = 217
	ReplyStatsYLine                 = 218
	ReplyEndOfStats                 = 219
	ReplyUserModeIs                 = 221
	ReplyServiceInfo                = 231
	ReplyEndOfServices              = 232
	ReplyServerList                 = 234
	ReplyEndOfServerList            = 235
	ReplyStatsUptime                = 242
	ReplyStatsNetOp                 = 243
	ReplyStatsHelpOp                = 244
	ReplyStatsPing                  = 246
	ReplyUsersOnlineGlobal          = 251
	ReplyOpersOnline                = 252
	ReplyUnknownConnections         = 253
	ReplyChannelCount               = 254
	ReplyUsersOnlineLocal           = 255
	ReplyAdminInfoStart             = 256
	ReplyAdminInfo1                 = 257
	ReplyAdminInfo2                 = 258
	ReplyAdminEmail                 = 259
	ReplyTraceLog                   = 261
	ReplyEndOfTrace                 = 262
	ReplyTryAgain                   = 263
	ReplyAway                       = 301
	ReplyUserHost                   = 302
	ReplyIsOn                       = 303
	ReplyUnAway                     = 305
	ReplyNowAway                    = 306
	ReplyWhoisUser                  = 311
	ReplyWhoisServer                = 312
	ReplyWhoisOperator              = 313
	ReplyWhoWasUser                 = 314
	ReplyEndOfWho                   = 315
	ReplyWhoisChanOp                = 316
	ReplyWhoisIdle                  = 317
	ReplyEndOfWhois                 = 318
	ReplyWhoisChannels              = 319
	ReplyListStart                  = 321
	ReplyList                       = 322
	ReplyEndOfList                  = 323
	ReplyChannelModeIs              = 324
	ReplyNoTopic                    = 331
	ReplyChanTopic                  = 332
	ReplyInviting                   = 341
	ReplyInvited                    = 345
	ReplyInviteList                 = 346
	ReplyEndOfInviteList            = 347
	ReplyExceptList                 = 348
	ReplyEndOfExceptList            = 349
	ReplyVersion                    = 351
	ReplyWho                        = 352
	ReplyNames                      = 353
	ReplyLinks                      = 384
	ReplyEndOfLinks                 = 365
	ReplyEndOfNames                 = 366
	ReplyBanList                    = 367
	ReplyEndOfBanList               = 368
	ReplyEndOfWhoWas                = 369
	ReplyInfo                       = 371
	ReplyMOTD                       = 372
	ReplyEndOfInfo                  = 374
	ReplyMOTDStart                  = 375
	ReplyEndOFMOTD                  = 376
	ReplyYoureOper                  = 381
	ReplyRehashing                  = 382
	ReplyYoureService               = 383
	ReplyTime                       = 391
	ReplyUsersStart                 = 392
	ReplyUsers                      = 393
	ReplyEndOfUsers                 = 394
	ReplyNoUsers                    = 395
	ReplyNoSuchNick                 = 401
	ReplyNoSuchServer               = 402
	ReplyNoSuchChannel              = 403
	ReplyCannotSendToChan           = 404
	ReplyTooManyChannels            = 405
	ReplyWasNoSuchNick              = 406
	ReplyTooManyTargets             = 407
	ReplyNoSuchService              = 408
	ReplyNoOrigin                   = 409
	ReplyInvalidCapCmd              = 410
	ReplyNoRecipient                = 411
	ReplyNoTextToSend               = 412
	ReplyNoTopLevel                 = 413
	ReplyWildTopLevel               = 414
	ReplyBadMask                    = 415
	ReplyTooManyMatches             = 416
	ReplyUnknownCommand             = 421
	ReplyNoMOTD                     = 422
	ReplyNoAdminInfo                = 423
	ReplyFileError                  = 424
	ReplyNoNicknameGiven            = 431
	ReplyErroneusNickname           = 432
	ReplyNicknameInUse              = 433
	ReplyNickCollision              = 436
	ReplyResourceUnavailable        = 437
	ReplyUserNotInChannel           = 441
	ReplyNotOnChannel               = 442
	ReplyUserOnChannel              = 443
	ReplyNoLogin                    = 447
	ReplySummonDisabled             = 446
	ReplyUsersDisabled              = 446
	ReplyNotRegistered              = 451
	ReplyNeedMoreParams             = 461
	ReplyAlreadyRegistered          = 462
	ReplyNoPermForHost              = 463
	ReplyPasswordMistmatch          = 464
	ReplyYoureBanned                = 465
	ReplyYouWillBeBanned            = 466
	ReplyChanPassAlreadySet         = 467
	ReplyChannelIsFull              = 471
	ReplyUnknownMode                = 472
	ReplyInviteOnlyChan             = 473
	ReplyBannedFromChan             = 474
	ReplyBadChannelPass             = 475
	ReplyBadChannelName             = 476
	ReplyNoChanModes                = 477
	ReplyBanListFUll                = 478
	ReplyNoPrivileges               = 481
	ReplyChanOpPrivsNeeded          = 482
	ReplyCantKillServer             = 483
	ReplyRestricted                 = 484
	ReplyChanOwnerRequired          = 485
	ReplyNoOperHost                 = 491
	ReplyNoServiceHost              = 492
	ReplyUnknownUserMode            = 501
	ReplyUsersDontMatch             = 502
	ReplyLoggedIn                   = 900
	ReplyLoggedOut                  = 901
	ReplySASLSuccess                = 903
	ReplySASLFail                   = 904
	ReplySASLTooLong                = 905
	ReplySASLAborted                = 906
	ReplySASLAlready                = 907
)
