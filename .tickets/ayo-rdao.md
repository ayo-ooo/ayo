---
id: ayo-rdao
status: closed
deps: [ayo-ydub]
links: []
created: 2026-02-23T22:15:03Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [removal, cli]
---
# Remove cmd/ayo/chat.go

~~Delete the `chat` command for web-based chat interface.~~

## Resolution

**Not applicable.** The ticket was based on outdated information.

`cmd/ayo/chat.go` contains `runInteractiveChat()` which implements the terminal-based
interactive TUI chat - NOT the web-based chat. The web-based chat was in
`internal/server/chat.go` which was already removed in ayo-ydub.

There is no separate `ayo chat` subcommand - `runInteractiveChat()` is called from
the main root command when running in interactive mode (no prompt argument).

This file must be kept as it's core CLI functionality.
