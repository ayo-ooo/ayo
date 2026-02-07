---
id: ase-0vin
status: closed
deps: [ase-95o4]
links: []
created: 2026-02-06T04:09:06Z
type: epic
priority: 1
assignee: Alex Cabrera
parent: ase-95o4
---
# Git Sync System

Implement git-based synchronization for sandbox state, enabling backup and multi-machine support.

## Design

## What Gets Synced
- /home/{agent}/ (agent homes)
- /shared/ (shared files)
- /var/log/irc/ (IRC logs)
- Config files from host

## Sync Model
- Sandbox state is a git repo
- Each machine is a branch
- Sync on session end (push)
- Sync on ayo start (pull)
- Manual sync commands available

## CLI Commands
- ayo sync init
- ayo sync remote <url>
- ayo sync push
- ayo sync pull
- ayo sync status

## Backup Commands
- ayo backup (manual snapshot)
- ayo backup list
- ayo backup restore <name>
- ayo backup export/import

## Acceptance Criteria

- Sandbox state tracked in git
- Automatic sync on session boundaries
- Multi-machine branches supported
- Manual backup/restore works
- Rolling auto-backup maintained

