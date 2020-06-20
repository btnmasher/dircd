dircd
=========

A derpy IRCd written in [Go](https://golang.org/).

Why?
----

Because I wanted to.

Status
------

**Heavy Development.**

Can connect, can PRIVMSG, join/talk on channels. Connections are -fairly- resilient. I modeled the connection handling code after Golang stdlib http server, and just trimmed out the HTTP protocol handling and shoved in IRC protocol handling. Trying to model it after the ease of use of Golang stdlib HTTP server and have a mostly clean API. I wanted to do it right cause I had plans to make a bigger project out of it.

Currently iterating on modes/permissions before I start fleshing out more of the features that heavily rely on them.

Need to do better with ISUPPORT, but for the most part the static configuration it currently has is RFC compliant, even if some of the stuff it says is a lie.

Some stuff to do (not exhaustive):

- [ ] Modes
  - [ ] User Modes
  - [ ] Channel Modes
  - [ ] Setting
  - [ ] Parameter Lists
  - [ ] Mode effect logic
- [ ] Nickname rule enforcement
- [ ] CAP negotiation
  - [ ] message-tag
  - [ ] SASL
  - [ ] more when I decide what I want to include
- [ ] Persistance
- [ ] Message Filtering
- [ ] Authentication
- [ ] Multiple Server architecture design?

Contribute?
-----------

Send pull requests, submit issues, whatever.

Licsense
--------

[3-Clause BSD](https://opensource.org/licenses/BSD-3-Clause) (At the top of each source file).
