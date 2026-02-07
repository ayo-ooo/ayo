---
id: ase-2msm
status: closed
deps: [ase-alok]
links: []
created: 2026-02-06T04:10:23Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-ka3q
---
# Add IRC helper scripts to sandbox

Add shell helper scripts for agents to easily send and receive IRC messages without needing to know raw IRC protocol.

## Design

## Helper Scripts
Install in /usr/local/bin/:

### msg - Send IRC message
#!/bin/sh
# Usage: msg <target> <message>
# target: #channel or @agent
target="$1"; shift
echo "PRIVMSG $target :$*" | nc -q0 localhost 6667

### irc-log - Read IRC logs
#!/bin/sh
# Usage: irc-log [channel] [lines]
channel="${1:-general}"
lines="${2:-20}"
tail -n "$lines" "/var/log/irc/$channel.log"

### irc-join - Join a channel
#!/bin/sh
echo "JOIN $1" | nc -q0 localhost 6667

## Agent .bashrc
Add to skeleton:
export IRC_NICK=$USER
alias msg='msg'

## Skill Documentation
Update sandbox skill to document IRC usage patterns.

## Acceptance Criteria

- msg command works for sending messages
- irc-log shows recent channel messages
- Scripts installed in base image
- Documented in sandbox skill

