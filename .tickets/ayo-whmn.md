---
id: ayo-whmn
status: open
deps: [ayo-6h19]
links: []
created: 2026-02-23T22:14:55Z
type: epic
priority: 0
assignee: Alex Cabrera
tags: [gtm, phase2]
---
# Phase 2: File System Model

Implement clear, safe file access patterns. This phase establishes how agents interact with host files safely.

## Goals

- Agent can read host files from `/mnt/{user}` (read-only)
- Agent requests permission before modifying host files
- User gets clear approval UI in terminal
- `--no-jodas` mode for power users who trust agents
- All file modifications are audit-logged for safety

## File System Layout

```
SANDBOX:
/home/{agent}/          # Per-agent home (read-write)
/mnt/{host_username}/   # Host home (READ-ONLY)
/workspace/             # Shared workspace (read-write)
/output/{session}/      # Safe write zone → auto-syncs to host
```

## File Request Workflow

```
1. Agent detects need to modify /mnt/user/Projects/app/main.go
2. Agent calls file_request tool:
   file_request({
     "action": "update",
     "path": "/mnt/user/Projects/app/main.go",
     "content": "...",
     "reason": "Fixed authentication bug"
   })
3. User sees prompt:
   ┌─────────────────────────────────────────────┐
   │ @ayo wants to update:                       │
   │   ~/Projects/app/main.go                    │
   │ Reason: Fixed authentication bug            │
   │                                             │
   │ [Y]es  [N]o  [D]iff  [A]lways for session   │
   └─────────────────────────────────────────────┘
4. User approves → file written to host
```

## --no-jodas Mode

Three ways to enable auto-approval:

1. **CLI flag**: `ayo --no-jodas "refactor codebase"`
2. **Global config**: `~/.config/ayo/config.json` → `permissions.no_jodas: true`
3. **Per-agent**: `ayo.json` → `permissions.auto_approve: true`

Safety: Even with --no-jodas, all modifications logged to audit.log

## Child Tickets

### Core
- `ayo-dicu`: Implement file_request tool
- `ayo-c5mt`: Implement file request approval UI (using charmbracelet/huh)
- `ayo-66df`: Implement /output safe write zone
- `ayo-0e81`: Add publish tool for moving files to host

### Permissions
- `ayo-evik`: Implement --no-jodas CLI flag
- `ayo-bw7o`: Add no_jodas to global config
- `ayo-xyby`: Add auto_approve to agent ayo.json
- `ayo-5kns`: Add session-scoped approval caching

### Audit
- `ayo-vclt`: Implement file modification audit logging
