---
id: ase-zak6
status: closed
deps: [ase-et8m, ase-frc8, ase-gbox]
links: []
created: 2026-02-10T01:37:18Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-8d04
---
# Update AGENTS.md with share system documentation

Update the agent memory file with comprehensive share system documentation.

## File to Modify
- AGENTS.md

## Dependencies
- ase-et8m (share add works)
- ase-frc8 (share list works)
- ase-gbox (share rm works)

## Content to Add

### Add new 'Share Commands' section after Mount Commands:

```markdown
## Share Commands

Shares provide instant access to host directories without sandbox restart.
Shares appear at /workspace/{name} inside the sandbox.

```bash
# Share a directory
ayo share ~/Code/project           # → /workspace/project
ayo share . --as myproject          # → /workspace/myproject
ayo share /tmp/data --session       # Session-only (removed on end)

# List shares
ayo share list
ayo share list --json

# Remove shares
ayo share rm project                # By name
ayo share rm ~/Code/project         # By path
ayo share rm --all                  # Remove all

# Migrate from deprecated mount system
ayo share migrate --dry-run
ayo share migrate
```

### Difference from Mount (deprecated)
- Shares take effect immediately (no sandbox restart)
- Shares use symlinks (mounts used bind mounts requiring daemon restart)
- Use `ayo share migrate` to convert existing mounts
```

### Update 'Key Files' table:
Add:
| internal/share/share.go | Share service for /workspace/ symlinks |
| ~/.local/share/ayo/shares.json | Persistent share configuration |

### Update 'Directory Purposes' in Architecture section:
Add /workspace/ to the container paths table:
| /workspace/ | User-shared host directories | User (ayo share) |

### Update Mount Commands section:
Add deprecation notice:
```markdown
## Mount Commands (DEPRECATED)

> ⚠️ The mount system is deprecated. Use `ayo share` instead.
> See Share Commands section above.
> Migrate existing mounts with `ayo share migrate`.

[existing content...]
```

### Update debugging workflows if relevant:
Add share-related debugging if needed

## Acceptance Criteria

- [ ] Share Commands section added
- [ ] All share subcommands documented
- [ ] /workspace/ path documented
- [ ] Key Files table updated
- [ ] Mount section has deprecation notice
- [ ] Migration command mentioned
- [ ] Difference between share/mount explained

