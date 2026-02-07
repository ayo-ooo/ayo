---
id: ase-ka3q
status: closed
deps: [ase-95o4]
links: []
created: 2026-02-06T04:08:35Z
type: epic
priority: 1
assignee: Alex Cabrera
parent: ase-95o4
---
# Alpine Base Image with IRC

Create the Alpine-based sandbox image with ngircd IRC server, agent user management, and core infrastructure.

## Design

## Components
1. Alpine rootfs setup (replacing busybox)
2. ngircd installation and configuration
3. Agent user creation system
4. Message routing via filesystem conventions
5. Startup scripts for sandbox services

## Directory Structure
/home/{agent}/           # Agent home directories
/shared/                 # Cross-agent permanent storage
/workspaces/{session}/   # Session-scoped workspaces
/var/log/irc/           # IRC logs for git sync
/mnt/host/              # Host filesystem mounts

## Acceptance Criteria

- Alpine rootfs replaces busybox
- ngircd starts on sandbox boot
- Agents auto-join IRC on session start
- IRC logs persisted to /var/log/irc/
- Agent users created lazily on first use

