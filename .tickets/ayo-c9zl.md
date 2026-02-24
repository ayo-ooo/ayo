---
id: ayo-c9zl
status: open
deps: []
links: []
created: 2026-02-23T22:15:11Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-6h19
tags: [removal, irc]
---
# Remove IRC integration code

Remove abandoned IRC bridge/integration code.

## Context

IRC integration was an early experiment that was never completed. Remove any traces.

## Tasks

1. Search for IRC references:
   ```bash
   grep -ri "irc" --include="*.go" .
   grep -ri "irc" --include="*.sh" debug/
   ```
2. Delete any IRC-related files
3. Remove debug/irc-status.sh if present
4. Remove any IRC config options

## Verification Steps

1. No "irc" references in codebase (case-insensitive search)
2. Build passes
3. Tests pass

## Acceptance Criteria

- [ ] No IRC code remains
- [ ] No IRC config options
- [ ] Build passes
