---
id: ase-hjhk
status: closed
deps: []
links: []
created: 2026-02-06T04:12:37Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-76ox
---
# Implement mounts.json for persistent grants

Create the mounts.json file and service for managing persistent host filesystem access grants.

## Design

## mounts.json Location
~/.local/share/ayo/mounts.json

## Format
{
  'version': 1,
  'permissions': [
    {
      'path': '/Users/alex/Code/myproject',
      'mode': 'readwrite',
      'granted_at': '2025-02-05T10:00:00Z',
      'granted_by': 'user'
    }
  ]
}

## Mount Service
internal/mounts/mounts.go:
- Load() - load mounts.json
- Save() - save mounts.json
- Grant(path, mode) - add permission
- Revoke(path) - remove permission
- List() - list all grants
- IsGranted(path, mode) - check if path is accessible

## Path Resolution
Store absolute paths.
Grant to '/Users/alex/Code/project' covers all subdirectories.
Check parent directories when resolving.

## Modes
- readonly: can read files
- readwrite: can read and write files

## Integration Points
- CLI: ayo mount commands use this service
- Run: Check grants before mounting paths
- Config: .ayo.json mounts merged with this

## Acceptance Criteria

- mounts.json created and managed
- Grant/revoke operations work
- Path resolution handles subdirectories
- Service loadable by run package

