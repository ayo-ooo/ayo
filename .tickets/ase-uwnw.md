---
id: ase-uwnw
status: closed
deps: []
links: []
created: 2026-02-10T01:31:57Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ase-8d04
---
# Add WorkspaceDir() function to sync package

Add a WorkspaceDir() function to internal/sync/git.go that returns the path to the workspace directory, and ensure this directory is created during Init().

## Context
The share system needs a workspace directory that will be mounted into containers at /workspace/. This directory will contain symlinks to user-shared host paths.

## Location
File: internal/sync/git.go

## Implementation Details

1. Add WorkspaceDir() function after SharedDir():
```go
// WorkspaceDir returns the directory for user-shared files within sandbox.
// This directory contains symlinks to host paths that users have shared.
func WorkspaceDir() string {
    return filepath.Join(SandboxDir(), "workspace")
}
```

2. Update Init() function to create workspace directory alongside homes/shared:
- Add WorkspaceDir() to the dirs slice on line ~55-58
- The existing loop creates directories with 0755 permissions and .gitkeep files

## Existing Pattern (for reference)
Look at HomesDir() and SharedDir() implementations (lines 27-35) and how they're used in Init() (lines 55-69).

## Testing
- Verify WorkspaceDir() returns {SandboxDir}/workspace
- Verify Init() creates the directory with 0755 permissions
- Verify .gitkeep file is created in the directory
- Test both dev mode (uses .local/share/ayo/) and production mode (~/.local/share/ayo/)

## Acceptance Criteria

- [ ] WorkspaceDir() function exists and returns correct path
- [ ] Init() creates workspace directory
- [ ] Directory has 0755 permissions
- [ ] .gitkeep file created in workspace directory
- [ ] Works in both dev mode and production mode
- [ ] Unit test added for WorkspaceDir()

