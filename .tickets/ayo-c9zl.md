---
id: ayo-c9zl
status: closed
deps: []
links: []
created: 2026-02-23T22:15:11Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-6h19
tags: [removal, irc, cleanup]
---
# Remove IRC integration code

Remove all IRC-related code from the codebase. This was an abandoned experiment for agent-to-agent communication.

## Known IRC Files

| File | Type | Description |
|------|------|-------------|
| `debug/irc-status.sh` | Script | 154-line debug script for ngircd |
| `debug/sandbox-status.sh` | Script | References IRC server status |
| `internal/integration/sandbox_alpine_test.go` | Test | `TestAlpineSandbox_InterAgentMessage` function |
| `internal/sandbox/images/alpine.md` | Doc | References `/var/log/irc/` |
| `internal/builtin/skills/sandbox/SKILL.md` | Doc | Mentions ngircd for agent communication |

## Deep Search Required

The IRC integration may have touched other files. Perform comprehensive search:

```bash
# Search for IRC references
grep -r -i "irc" --include="*.go" --include="*.sh" --include="*.md" .
grep -r "ngircd" .
grep -r "6667" .  # IRC port
grep -r "InterAgent" .

# Search for related Matrix code (also removed)
grep -r "matrix" --include="*.go" .
grep -r "conduit" .
grep -r "matrix_broker" .
grep -r "matrix_rpc" .
grep -r "matrix_chat" .
```

## Related Dead Code

Also search for and remove any:
- Matrix chat integration remnants (`matrix_broker.go`, `matrix_rpc.go`, `conduit.go`)
- Broker/message passing code that was IRC-dependent
- Test fixtures or mocks for IRC
- Any `nc` (netcat) commands used for IRC communication

## Implementation Steps

1. Run comprehensive grep searches above
2. Document all found references
3. Remove IRC-specific files
4. Remove IRC references from other files
5. Remove IRC-related test cases
6. Update documentation that references IRC
7. Remove Matrix-related remnants if found
8. Run `go build ./...` to verify no broken imports
9. Run `go test ./...` to verify tests pass

## Acceptance Criteria

- [ ] No IRC code remains
- [ ] No IRC config options
- [ ] No Matrix broker remnants
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] No references to ngircd, port 6667, or InterAgent
