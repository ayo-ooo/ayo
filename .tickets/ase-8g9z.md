---
id: ase-8g9z
status: open
deps: []
links: []
created: 2026-02-07T03:25:16Z
type: feature
priority: 2
assignee: Alex Cabrera
parent: ase-6khq
---
# Add `trigger watch` subcommand

Add `watch` subcommand for creating filesystem triggers.

Current:
  ayo triggers add --type watch --agent @build --path ./src --patterns "*.go"

Proposed:
  ayo trigger watch ./src @build
  ayo trigger watch ./src @build "*.go"
  ayo trigger watch ./src @build "*.go" "*.mod" --recursive

Usage: trigger watch <path> <agent> [patterns...] [flags]

Flags:
  --recursive    Watch subdirectories
  --events       Events to trigger on (create, modify, delete)
  --prompt       Prompt to pass to agent

## Acceptance Criteria

- `ayo trigger watch ./src @build` creates trigger
- `ayo trigger watch ./src @build "*.go"` filters patterns
- --recursive flag works
- Shows created trigger ID on success

