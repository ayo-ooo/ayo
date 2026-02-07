---
id: ase-2bae
status: closed
deps: []
links: []
created: 2026-02-06T04:13:19Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-0vin
---
# Initialize git repository for sandbox state

Set up git repository structure in the sandbox state directory for tracking changes and enabling sync.

## Design

## Repository Location
~/.local/share/ayo/sandbox/ becomes a git repo

## Tracked Directories
- homes/         # Agent home directories
- shared/        # Shared files
- irc-logs/      # IRC server logs

## .gitignore
# Large/binary files
*.log
*.tmp
**/node_modules/
**/.cache/

## Initial Setup
On first daemon start (or ayo sync init):
1. git init in sandbox directory
2. Create .gitignore
3. Initial commit: 'Initial sandbox state'

## Branch Structure
- main: Canonical state
- machines/{hostname}: Per-machine branch

## Implementation
internal/sync/git.go:
- Init() - initialize repo
- IsInitialized() - check if repo exists
- Commit(message) - commit current state
- GetBranch() - current branch name
- CreateMachineBranch() - create machines/{hostname}

## Acceptance Criteria

- Git repo initialized in sandbox dir
- Correct directories tracked
- .gitignore excludes noise
- Branch structure ready

