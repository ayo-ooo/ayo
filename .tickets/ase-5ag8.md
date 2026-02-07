---
id: ase-5ag8
status: closed
deps: [ase-2bae]
links: []
created: 2026-02-06T04:13:28Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-0vin
---
# Implement sync push/pull operations

Implement git push and pull operations for syncing sandbox state between machines.

## Design

## Push Operation
1. Stage all changes: git add -A
2. Commit with message: 'Session {id}: {title}' or auto-generated
3. Push to remote: git push origin machines/{hostname}

## Pull Operation
1. Fetch from remote: git fetch origin
2. Check for divergence between local and main
3. If behind, merge or rebase
4. Handle conflicts (prompt or auto-resolve)

## Conflict Resolution
For agent homes: prefer local (agent was working)
For shared: prompt user or use latest timestamp
For IRC logs: merge (append-only)

## Auto-Sync Points
- Session end: push
- Daemon start: pull
- Configurable interval: background sync

## Implementation
internal/sync/sync.go:
- Push(message) error
- Pull() error
- Fetch() error
- Status() SyncStatus
- HasRemote() bool

## SyncStatus
type SyncStatus struct {
    LocalBranch    string
    RemoteConfigured bool
    Ahead          int
    Behind         int
    LastSync       time.Time
}

## Acceptance Criteria

- Push commits and pushes to remote
- Pull fetches and merges
- Conflicts handled gracefully
- Status shows sync state

