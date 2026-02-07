---
id: ase-14z2
status: closed
deps: [ase-si5i]
links: []
created: 2026-02-06T04:14:15Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-x0vq
---
# Add ayo sandbox login command

Implement 'ayo sandbox login' for interactive shell access to the sandbox, with user impersonation support.

## Design

## Command
ayo sandbox login              # Login as root
ayo sandbox login --as @ayo    # Login as @ayo user
ayo sandbox login --as @crush  # Login as @crush user

## Implementation
1. Check sandbox is running
2. Ensure user exists (if --as specified)
3. Exec interactive shell

## Shell Selection
Prefer in order:
1. /bin/bash (if installed)
2. /bin/ash (Alpine default)
3. /bin/sh

## PTY Handling
Use golang.org/x/term for PTY allocation.
Pass through terminal size.
Handle SIGWINCH for resize.

## IMPORTANT: Agent Exclusion
This command MUST be excluded from agent access:
1. Not listed in 'ayo' skill SKILL.md
2. Hidden from --help when in agent context (if possible)
3. Add note in command help: 'Note: This command is for human use only.'

## Environment
Set sensible environment:
- TERM from host
- HOME=/home/{user} or /root
- USER={user} or root
- PATH includes /usr/local/bin

## Exit
Exit code from shell passed through.
Clean disconnect on Ctrl+D or exit.

## Acceptance Criteria

- Login as root works
- Login as agent user works
- PTY works correctly
- Excluded from agent tooling

