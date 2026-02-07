---
id: ase-n40v
status: closed
deps: [ase-fb0m]
links: []
created: 2026-02-06T04:10:13Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-ka3q
---
# Create sandbox directory structure

Create the standard directory structure inside the sandbox for agent homes, workspaces, shared files, and IRC logs.

## Design

## Directory Structure
/home/{agent}/           # Agent home directories (created lazily)
/shared/                 # Permanent cross-agent storage
/workspaces/            # Session-scoped workspaces (created per session)
/var/log/irc/           # IRC server logs
/mnt/host/              # Mount point for host filesystem

## Implementation
1. Add directory creation to sandbox startup
2. Set appropriate permissions:
   - /shared: 1777 (sticky, world-writable)
   - /workspaces: 755, owned by root
   - /var/log/irc: 755, owned by ngircd user
   - /mnt/host: 755, owned by root

## Host-Side Storage
These directories map to ~/.local/share/ayo/sandbox/:
- homes/ -> bind mount to /home/ (or overlay)
- shared/ -> bind mount to /shared/
- workspaces/ -> bind mount to /workspaces/

This ensures persistence across container restarts.

## Acceptance Criteria

- All directories exist on sandbox start
- Permissions set correctly
- Persistence via bind mounts to host

