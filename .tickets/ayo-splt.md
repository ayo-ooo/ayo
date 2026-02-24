---
id: ayo-splt
status: open
deps: []
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 3
assignee: Alex Cabrera
parent: ayo-6h19
tags: [cleanup, tech-debt]
---
# Split large files

Split oversized files into smaller, focused modules.

## Files to Split

### internal/agent/agent.go (~1100 lines)

Split into:
| New File | Content |
|----------|---------|
| `agent.go` | Core Agent struct, main methods |
| `config.go` | Config structs, loading |
| `loading.go` | Agent discovery, directory scanning |
| `memory.go` | Memory-related config and helpers |
| `guardrails.go` | Guardrails injection |

### internal/daemon/server.go (~1100 lines)

Split into:
| New File | Content |
|----------|---------|
| `server.go` | Core Server struct, lifecycle |
| `rpc_handlers.go` | General RPC handlers |
| `squad_handlers.go` | Squad-specific handlers |
| `trigger_handlers.go` | Trigger-related handlers |
| `session_handlers.go` | Session management |

## Guidelines

- Keep related code together
- Minimize cross-file dependencies
- Maintain clear package boundaries
- Update imports as needed

## Testing

- All tests pass after split
- No behavior changes
