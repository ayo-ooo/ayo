# In Progress: Sandbox CLI and Agent Identity System

## Completed Work

### Sandbox CLI (ayo-15ou) - CLOSED
Implemented full `ayo sandbox` command tree:
- `ayo sandbox list` - Lists active sandboxes from container runtime
- `ayo sandbox show <id>` - Shows sandbox details, mounts, user, status
- `ayo sandbox exec <id> <cmd>` - Executes commands with `--user`, `--workdir` flags
- `ayo sandbox shell <id>` - Interactive shell loop (workaround for Apple Container TTY limitation)
- `ayo sandbox logs <id>` - Uses container CLI directly
- `ayo sandbox stop <id>` - Stops sandbox with `--force`, `--timeout` flags
- `ayo sandbox prune` - Removes stopped sandboxes with `--force`, `--all` flags

### Agent Identity System (ayo-41ms) - CLOSED
Agents can now run as dedicated users inside sandboxes instead of root:

**Config example:**
```json
{
  "sandbox": {
    "enabled": true,
    "user": "ayo"
  }
}
```

**Changes made:**
| File | Changes |
|------|---------|
| `internal/agent/agent.go` | Added `User` field to `SandboxConfig` |
| `internal/providers/providers.go` | Added `User`, `SetupCommands` to `SandboxCreateOptions`; `User` to `Sandbox` |
| `internal/sandbox/apple.go` | User creation via `adduser -D` at container startup |
| `internal/sandbox/bash.go` | Executor passes `User` to all exec calls |
| `internal/run/run.go` | Passes agent's sandbox user to executor |
| `cmd/ayo/sandbox.go` | Displays user in `sandbox show` output |

## Remaining Open Tickets

| Ticket | Description | Depends On |
|--------|-------------|------------|
| ayo-0u7e | Persistent Agent Home Directories | ayo-41ms (done) |
| ayo-1rw2 | File Transfer: push/pull | ayo-15ou (done) |
| ayo-3lv4 | Multi-Agent Sandbox Collaboration | ayo-41ms (done) |
| ayo-z7yy | Working Copy Model with sync | ayo-1rw2 |
| ayo-q84j | File Request Tool | ayo-1rw2 |
| ayo-je5i | Publish Tool | ayo-q84j |
| ayo-nb7o | Integration tests for sandbox CLI | ayo-dwgv (done) |
| ayo-gj3k | Update AGENTS.md documentation | ayo-nb7o |
| ayo-z91f | Sandbox Stats and Resource Monitoring | ayo-12vf (done) |

## Next Priority

1. **ayo-0u7e** - Persistent Agent Home Directories (unblocked)
2. **ayo-1rw2** - File Transfer push/pull (unblocked)
3. **ayo-nb7o** - Integration tests (unblocked)
