---
id: ase-2fnp
status: closed
deps: [ase-hjhk]
links: []
created: 2026-02-06T04:13:06Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-76ox
---
# Support mounts in .ayo.json project config

Add mounts section to .ayo.json for project-level mount declarations.

## Design

## Format
// .ayo.json or ayo.json in project root
{
  'mounts': {
    '.': 'readwrite',        // Current directory
    '../shared-lib': 'readonly',
    '~/Documents/notes': 'readonly'
  }
}

## Path Resolution
- Relative paths: relative to config file location
- ~/ paths: expanded to user home
- Absolute paths: used as-is

## Implementation
1. Add Mounts field to config.DirConfig
2. Parse in config loading
3. Pass to Runner during setup
4. Merge with other mount sources

## Merge with Other Sources
Project mounts are middle priority:
1. CLI --mount (highest)
2. .ayo.json mounts (middle)
3. mounts.json persistent (lowest)

## Validation
On ayo start in directory:
- Load .ayo.json
- Validate mount paths exist
- Warn if path not granted via mounts.json

## Security
Project config can only RESTRICT access, not grant new access.
Paths in .ayo.json must also be in mounts.json (or --mount).
This prevents malicious .ayo.json from granting access.

## Acceptance Criteria

- .ayo.json mounts parsed
- Paths resolved correctly
- Merged with other mount sources
- Security: can't grant new access

