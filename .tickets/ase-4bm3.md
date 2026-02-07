---
id: ase-4bm3
status: closed
deps: [ase-alok]
links: []
created: 2026-02-06T04:14:26Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-x0vq
---
# Add ayo messages command for IRC logs

Implement 'ayo messages' command for viewing IRC logs from the sandbox.

## Design

## Commands
ayo messages                   # Recent from #general
ayo messages -f/--follow       # Live tail mode
ayo messages -c <channel>      # Specific channel
ayo messages -s/--search <q>   # Search messages
ayo messages --from @ayo       # Filter by sender
ayo messages --to @crush       # Filter by recipient

## Implementation
cmd/ayo/messages.go:
- Read from sandbox's /var/log/irc/
- Parse IRC log format
- Apply filters
- Format output

## Log Location
~/.local/share/ayo/sandbox/irc-logs/ on host
(bind mounted to /var/log/irc/ in sandbox)

## Log Format
[2025-02-05 10:30:45] <ayo> Message text here
[2025-02-05 10:30:50] <crush> Response text

## Output Format
Default: colorized terminal output
--json: JSON array of messages

## Follow Mode
Use fsnotify to watch for changes.
Print new messages as they appear.
Ctrl+C to exit.

## Search
Simple substring match.
Future: regex support.

## Channels
Read from channel-specific log files:
- general.log
- session-{id}.log
- private messages in separate files

## Acceptance Criteria

- Basic message viewing works
- Follow mode works
- Search and filters work
- JSON output available

