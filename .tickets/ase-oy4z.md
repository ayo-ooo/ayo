---
id: ase-oy4z
status: closed
deps: [ase-2bae]
links: []
created: 2026-02-06T04:13:37Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-0vin
---
# Add sync CLI commands

Implement 'ayo sync' CLI commands for manual sync operations.

## Design

## Commands
ayo sync init              # Initialize sync (git init)
ayo sync remote <url>      # Set sync remote
ayo sync push              # Push to remote
ayo sync pull              # Pull from remote
ayo sync status            # Show sync state

## Implementation
cmd/ayo/sync.go:
- newSyncCmd() - parent command
- newSyncInitCmd()
- newSyncRemoteCmd()
- newSyncPushCmd()
- newSyncPullCmd()
- newSyncStatusCmd()

## sync init
- Initialize git repo if not exists
- Create machine branch
- Prompt for remote URL (optional)

## sync remote
- Set git remote 'origin'
- Validate URL format
- Test connection (optional)

## sync status
Output:
  Sync: enabled
  Remote: git@github.com:user/ayo-sync.git
  Branch: machines/macbook-pro
  Status: 2 commits ahead, 1 behind
  Last sync: 2 hours ago

## Flags
--json: JSON output
--quiet: Minimal output
--force: Force push/pull

## Acceptance Criteria

- All sync commands work
- Remote configuration persisted
- Status shows useful info
- JSON output available

