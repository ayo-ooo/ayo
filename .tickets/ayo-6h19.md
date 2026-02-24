---
id: ayo-6h19
status: open
deps: []
links: []
created: 2026-02-23T22:14:48Z
type: epic
priority: 0
assignee: Alex Cabrera
tags: [gtm, phase1]
---
# Phase 1: Foundation - Simplification

Remove complexity and establish clear mental model. This is the prerequisite for all other phases.

## Goals

- Remove ~5,000+ lines of unused/confusing code
- Simplify daemon to core functions only
- Implement ayod (in-sandbox daemon) for elegant sandbox management
- Establish shared sandbox model with agents as real Unix users
- Set up host home mount for file access

## Code to Remove

| Path | Reason | Est. Lines |
|------|--------|------------|
| `internal/server/` | REST API not needed for CLI | ~2500 |
| `web/` | Web interface incomplete, distracts | ~1000 |
| `cmd/ayo/serve.go` | Server command | ~100 |
| `cmd/ayo/chat.go` | Web chat | ~50 |
| `internal/flows/yaml_executor.go` | Complex YAML flows | ~500 |
| `internal/flows/yaml_validate.go` | YAML validation | ~200 |
| `internal/daemon/webhook_server.go` | Premature integration | ~500 |
| `internal/server/tunnel/` | Cloudflare tunnel | ~200 |
| `internal/server/qrcode.go` | Mobile QR | ~100 |

## Code to Simplify

- `internal/daemon/server.go` - Remove webhook/serve endpoints
- `cmd/ayo/flows.go` - Keep only inspect/graph, remove run/execute
- `internal/flows/` - Keep discover, parse, DAG inspection only

## Sandbox Architecture

Each agent runs as a **real Unix user** inside the shared sandbox:

```
@AYO SANDBOX (default for all agents):
/home/
├── ayo/         # Unix user: ayo (@ayo)
├── crush/       # Unix user: crush (@crush)
├── reviewer/    # Unix user: reviewer
└── {agent}/     # Created on first use via ayod

/mnt/{user}/     # Host home (read-only)
/workspace/      # Shared workspace (group writable)
/output/         # Safe write zone
```

ayod runs as PID 1 and manages users, execution, and host communication.

## Child Tickets

### Sandbox Bootstrap (do first)
- `ayo-kkxg`: **Implement ayod in-sandbox daemon** (critical path)
- `ayo-ao4q`: Implement shared sandbox with per-agent Unix users
- `ayo-1xg8`: Standardize @ayo sandbox home directory
- `ayo-c8px`: Add host home mount at /mnt/{username}

### Removal
- `ayo-8nn8`: Remove cmd/ayo/serve.go
- `ayo-rdao`: Remove cmd/ayo/chat.go
- `ayo-ydub`: Remove internal/server/ package
- `ayo-tha0`: Remove web/ directory
- `ayo-1ryh`: Remove YAML flow executor
- `ayo-ieiy`: Remove webhook server from daemon
- `ayo-c9zl`: Remove IRC integration code
- `ayo-fwye`: Simplify cmd/ayo/flows.go
- `ayo-qbsu`: Clean up daemon server.go
- `ayo-enaj`: Clean up dead code

### Cleanup
- `ayo-xhox`: Update go.mod dependencies

### Technical Debt
- `ayo-clns`: Clean slate preparation (run before starting)
- `ayo-debt`: Apply gopls modernize suggestions
- `ayo-depr`: Remove deprecated functions
- `ayo-dupl`: Consolidate duplicate interfaces
- `ayo-splt`: Split large files

### Externalized Prompts
- `ayo-xprm`: **Externalize all prompts** - zero hardcoded prompt strings in codebase
