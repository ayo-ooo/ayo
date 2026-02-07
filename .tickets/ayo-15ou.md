---
id: ayo-15ou
status: closed
deps: []
links: []
created: 2026-02-05T18:51:32Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, cli]
---
# Add cmd/ayo/sandbox.go with subcommand structure

Create sandbox.go with cobra subcommand structure: list, show, shell, exec, logs, stop, prune. Wire into root command.

## Acceptance Criteria

- sandbox.go exists with all subcommand stubs
- ayo sandbox --help shows all commands
- Commands return 'not implemented' placeholder

