---
id: ayo-fwye
status: open
deps: [ayo-1ryh]
links: []
created: 2026-02-23T22:15:11Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-6h19
tags: [simplify, flows, cli]
---
# Simplify cmd/ayo/flows.go

Remove flow execution commands. Keep inspection-only commands.

## Context

With YAML executor removed (ayo-1ryh), the `run` and `execute` commands have no backend. We keep inspection commands for understanding flow dependencies.

## Commands to Remove

- `ayo flow run <name>` - Execute a flow
- `ayo flow execute <name>` - Alias for run

## Commands to Keep

- `ayo flow list` - List available flows
- `ayo flow show <name>` - Show flow definition
- `ayo flow inspect <name>` - Show parsed structure
- `ayo flow graph <name>` - Output DAG visualization

## Verification Steps

1. Remove run/execute subcommand registration
2. Remove associated handler functions
3. Run `go build ./cmd/ayo/...`
4. Verify `ayo flow list` works
5. Verify `ayo flow run` returns "unknown command"

## Acceptance Criteria

- [ ] Run/execute commands removed
- [ ] List/show/inspect/graph commands work
- [ ] Build passes
- [ ] Help text updated
