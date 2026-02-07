---
id: ase-hbyj
status: closed
deps: [ase-95o4]
links: []
created: 2026-02-06T04:14:02Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-x0vq
---
# Add --latest flag for session commands

Add --latest/-l flag to session-related commands to automatically select the most recent session.

## Design

## Affected Commands
- ayo sessions continue --latest
- ayo sessions show --latest
- ayo sessions archive --latest
- ayo sessions delete --latest

## Also Add to Root
- ayo -c/--continue 'prompt' → continue most recent session
- ayo -s/--session <id> 'prompt' → continue specific session

## Implementation
1. Add --latest/-l flag to relevant commands
2. Query for most recent session in current directory
3. If in directory, prefer sessions from that directory
4. If no directory context, use global most recent

## Session Resolution
GetLatestSession(ctx, directory string) (*Session, error)
- If directory provided, filter by directory first
- Sort by last_active_at descending
- Return first result

## Error Cases
- No sessions exist: 'No sessions found'
- No sessions in directory: 'No sessions found in this directory. Use --all for global.'

## Root Command Integration
ayo --continue 'follow up message'
  → Find latest session
  → Resume with new message
  → Single turn, exit after response

## Acceptance Criteria

- --latest works on all session commands
- -c/--continue works on root
- Directory preference applied
- Clear error messages

