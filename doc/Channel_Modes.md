# Channel Modes

## Permission Levels
| Short Name | Long Name       | Description                                                                      |
|:----------:|:---------------:|:--------------------------------------------------------------------------------:|
| Admin     | Network Admin    | Highest level of Administrator.
| Net Op    | Network Operator | High level operator, beholden to the network admin.
| Help Op   | Helper Operator  | Low level operator, typically for staffers.
| Owner     | Channel Owner    | The owner of a particular channel. (+O)
| Chan Op   | Channel Operator | Chanel administrator (+o)
| Half Op   | Channel Half Op  | A lower level channel administrator, less control over the channel. (+h)
| User      | Standard User    | Base level permission of a user in a channel.


- May Set: Lowest permission level required to set mode. Commands with set permission of "User" are only able to be set on self if not Helper Operator or higher.

## Channel Modes
| C | Channel Mode | Paramter | May Set | Description                                                                 |
|:-:|:-------------|:--------:|:-------:|:----------------------------------------------------------------------------|
| a | Anonymous    |          | Chan Op | Sets the channel to anonymous mode. No channel names list, messages are all sent from <anonymous!anonymous@anonymous>
| A | Admin Only   |          | Admin   | Sets the channel to Admin-only mode. Only Admins can see the channel in list, join|the channel, o| talk in t|e channel|
| b | Ban          | Hostmask | Half Op | Bans the hostmask from the channel.
| B | Banned Chan  |          | Help Op | Bans the given channel from the network. Similar to Reserved but users will receive a different error message when attempting to join.
| c | Censored     |          | Half Op | Sets the channel to censored mode using the server's word blacklist.
| C |              |          |         |
| d |              |          |         |
| D |              |          |         |
| e | External     |          | Chan Op | Sets the channel to allow external messages.
| E | Event Mode   |          | Chan Op | Sets the channel to Event mode, only showing messages and nicknames of op/halfops. Also hides join/part/quit/nickchange notifications.
| f |              |          |         |
| F | Flood Immune |          | Net Op  | Sets the channel to be ignored by the server's flood protecton.
| g |              |          |         |
| G |              |          |         |
| h | Half Op      | Nickname | Chan Op | Sets the given user to Half Op status for the channel.
| H | HelpOp Only  |          | Help Op | Sets the channel to HelpOp-only mode. Only users with HelpOp permission and above can see the channel in list, join the channel, or talk in the channel.
| i | Invte Only   |          | Chan Op | Sets the channel to invie-only mode for users below network staff or channel owner.
| I | No Invites   |          | Chan Op | Disables the abiltiy for users below channel operator from inviting to the channel.
| j |              |          |         |
| J |              |          |         |
| k |              |          |         |
| K |              |          |         |
| l |              |          |         |
| L | Linked       | Channel  | Owner   | Sets the channel to Linked mode. This links the channel to the specified channel. Both channels must be owned by the same user. Channel modes (eg: +mn) will apply to messages sent from the linked channels as if they were a normal user.
| m | Moderated    |          | Half Op | Sets the channel to Moderated mode, only users with +v can speak.
| M | Moved        | Channel  | Owner   | Sets the channel to Moved mode, redirecting joins to the specified channel. The specified channel cannot have +M set, and must be owned by the same user. Other permission and protection flags still apply (eg: Admin/NetOp/HelpOp only,  Protected, Reg Only).
| n | Normal Text  |          | Half Op | Sets the channel to Normal Text mode, stripping color codes, formatting codes and non alpha-numeric or standard keyboard symbol caracters from channel messages.
| N | NetOp Only   |          |         | Sets the channel to NetOp-only mode. Only users with NetOp permission and above can see the channel in list, join the channel, or talk in the channel.
| o | Op           | Nickname | Chan Op | Sets the given user to Op status for the channel.
| O | Owner        | Nickname | Owner   | Sets the given user to the Channel Owner. Will break Link and Move modes.
| p | Private      |          | Chan Op | Sets the channel to private mode. Will not show up in channel list or on WHOIS requests unless the requesting user shares the channel with the target user.
| P | Protected    | Password | Chan Op | Sets the channel to be password proteted with the given password. If none specified the flag is ignored by the server.
| q | Quiet        |          | Chan Op | Sets the channel to quiet mode, joins/parts/quits/nickchanges will not be announced.
| Q |              |          |         |
| r | Reg Only     |          | Chan Op | Sets the channel to only allow Registered nicknames (Usermode +r) to join/speak.
| R | Reserved     |          | Help Op | Sets a channel to reserved mode, making the channel unusable, hidden, and owned by no one.
| s | Secure Only  |          | Chan Op | Sets the channel to only allow users connecting with secure connections (+s).
| S |              |          |         |
| t | Topic Lock   |          | Chan Op | Locks the topic of the channel to only be able to be changed by the channel owner.
| T | Throttled    | Msg/Sec  | Chan Op | Sets the channel to limit messages per second (Minimum/Default 1).
| u |              |          |         |
| U |              |          |         |
| v | Voice        | Nickname | Half Op | Sets the user to have voice in the channel, allowing them to speak when +m is set.
| V | Verified     |          | Net Op  | Sets the channel to Verified mode, marking the channel as an Official chanel that has been sanctioned by the network staff.
| w |              |          |         |
| W |              |          |         |
| x |              |          |         |
| X |              |          |         |
| y |              |          |         |
| Y |              |          |         |
| z |              |          |         |
| Z |              |          |         |