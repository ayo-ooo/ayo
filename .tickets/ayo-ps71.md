---
id: ayo-ps71
status: closed
deps: []
links: []
created: 2026-02-06T22:20:08Z
type: bug
priority: 2
assignee: Alex Cabrera
tags: [sandbox, cli]
---
# Missing 'ayo sandbox start' command

## Issue
The sandbox CLI has stop, exec, show, list, prune, logs, shell commands but no 'start' command.

## Expected Behavior
- `ayo sandbox start <id>` should restart a stopped sandbox
- AGENT_MANUAL_TEST.md Section 3.7 expects this command

## Current Behavior
- No start command available
- Stopped sandboxes cannot be restarted via CLI
- User must prune and wait for daemon to recreate

## Impact
Cannot restart stopped sandboxes without pruning them

## Suggested Fix
Add `ayo sandbox start <id>` command that:
1. Takes sandbox ID as argument
2. Calls provider.Start(id)
3. Returns confirmation or error


## Notes

**2026-02-06T22:32:26Z**

FIXED: Added newSandboxStartCmd() function to cmd/ayo/sandbox.go. Tested - sandbox start ayo-sandbox-176f6c89 works correctly.
