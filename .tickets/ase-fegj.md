---
id: ase-fegj
status: closed
deps: []
links: []
created: 2026-02-07T03:22:13Z
type: feature
priority: 2
assignee: Alex Cabrera
parent: ase-7l1g
---
# Add top-level `ayo watch` command

Create a new top-level `watch` command for file system triggers.

Current:
  ayo triggers add --type watch --agent @build --path ./src --patterns "*.go"

Proposed:
  ayo watch ./src @build --patterns "*.go"
  ayo watch ./src @build "*.go"           # patterns as positional arg

Implementation:
- New cmd/ayo/watch.go file
- Positional args: <path> <agent> [patterns...]
- Flags: --recursive, --events, --prompt
- Calls same daemon.TriggerRegister under the hood
- `triggers add --type watch` becomes hidden alias

## Acceptance Criteria

- `ayo watch ./src @build "*.go"` creates watch trigger
- `ayo watch . @review` watches all files in current dir
- `ayo watch --help` shows clear usage
- Output shows created trigger ID

