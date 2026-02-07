---
id: ase-wl6m
status: closed
deps: [ase-95o4]
links: []
created: 2026-02-06T04:14:56Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-x0vq
---
# Add --json and --quiet flags to all commands

Ensure all CLI commands support --json and --quiet flags for agentic use.

## Design

## Flags
--json: Output in JSON format
--quiet/-q: Minimal output, suppress informational messages

## Implementation Pattern
1. Add flags to command definitions
2. Check flags before output
3. Use structured output helpers

## Structured Output Helper
internal/cli/output.go:
type Output struct {
    JSON  bool
    Quiet bool
}

func (o *Output) Print(data any, text string)
func (o *Output) PrintError(err error)
func (o *Output) PrintSuccess(text string)

## JSON Output Requirements
- Consistent structure across commands
- Include 'success' boolean
- Include relevant data
- Error includes 'error' field

## Quiet Mode
- Only output essential information
- No progress indicators
- No success confirmations
- Still output errors

## Exit Codes
Document and implement consistent exit codes:
- 0: Success
- 1: General error
- 2: Invalid input
- 3: Not found
- etc.

## Commands to Update
All existing commands:
- agents, skills, sessions, memory
- sandbox, flows, plugins, chain
- setup, doctor, status

New commands:
- mount, sync, backup, triggers, messages

## Acceptance Criteria

- All commands have --json flag
- All commands have --quiet flag
- Output is consistent
- Exit codes documented

