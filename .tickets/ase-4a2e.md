---
id: ase-4a2e
status: closed
deps: [ase-2bae]
links: []
created: 2026-02-06T04:13:47Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-0vin
---
# Add backup CLI commands

Implement 'ayo backup' CLI commands for manual backup/restore operations.

## Design

## Commands
ayo backup                   # Create timestamped backup
ayo backup --name <name>     # Create named backup
ayo backup list              # List backups
ayo backup show <name>       # Show backup details
ayo backup restore <name>    # Restore from backup
ayo backup export <path>     # Export to portable archive
ayo backup import <path>     # Import from archive
ayo backup prune             # Clean old auto-backups

## Backup Contents
- sandbox/ (homes, shared, workspaces)
- config/ (~/.config/ayo/)
- data/ (~/.local/share/ayo/ minus sandbox)
- manifest.json (metadata)

## Storage
~/.local/share/ayo/backups/
├── manual/
│   └── 2025-02-05-pre-upgrade.tar.zst
└── auto/
    └── latest.tar.zst

## Rolling Auto-Backup
- One 'latest' backup maintained
- Updated periodically (on session end or interval)
- Stored in backups/auto/latest.tar.zst

## Export Format
Portable tar.zst with manifest:
- Includes version info
- Checksums for integrity
- Can be imported on any machine

## Restore
1. Stop daemon if running
2. Backup current state (safety)
3. Extract backup
4. Start daemon

## Acceptance Criteria

- All backup commands work
- Backups include all relevant data
- Export/import portable
- Restore is safe and complete

