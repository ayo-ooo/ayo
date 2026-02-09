---
id: ase-ns3k
status: open
deps: []
links: []
created: 2026-02-09T03:26:45Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-zlew
---
# Standardize error message format across CLI

## Background

With many new CLI commands, we need consistent error message formatting for a professional user experience.

## Why This Matters

Inconsistent errors:
- Confuse users about what went wrong
- Make debugging harder
- Appear unprofessional
- Don't help users fix the problem

## Error Message Guidelines

### Format

```
Error: <brief description>

<details if helpful>

Suggestion: <how to fix>
```

### Examples

**Good:**
```
Error: Agent '@nonexistent' not found

Available agents:
  @ayo, @researcher, @writer

Suggestion: Use 'ayo agents list' to see all agents
```

**Bad:**
```
agent not found: nonexistent
```

**Good:**
```
Error: Invalid cron expression 'every day'

Expected format: '* * * * *' (minute hour day month weekday)

Suggestion: Use 'ayo trigger schedule @agent "0 9 * * *"' for 9 AM daily
           Or try natural language: 'ayo trigger schedule @agent "every day at 9am"'
```

**Bad:**
```
parse error: invalid cron
```

### Categories

1. **User errors** (recoverable):
   - Missing arguments
   - Invalid input
   - Not found errors
   - Permission errors
   
2. **System errors** (not user's fault):
   - Daemon not running
   - Network failures
   - Database errors

### Color Coding

- `Error:` in red
- Suggestions in cyan
- Available options in yellow

### Error Codes

Return appropriate exit codes:
- 0: Success
- 1: General error
- 2: Usage error (wrong arguments)
- 3: Not found
- 4: Permission denied
- 5: Timeout

## Implementation

### Error Helper Package

```go
// internal/cli/errors.go
type CLIError struct {
    Brief      string
    Details    string
    Suggestion string
    ExitCode   int
}

func (e *CLIError) Error() string {
    return e.Brief
}

func (e *CLIError) Print(w io.Writer) {
    fmt.Fprintf(w, "Error: %s\n", e.Brief)
    if e.Details != "" {
        fmt.Fprintf(w, "\n%s\n", e.Details)
    }
    if e.Suggestion != "" {
        fmt.Fprintf(w, "\nSuggestion: %s\n", e.Suggestion)
    }
}

// Common errors
func ErrAgentNotFound(name string, available []string) *CLIError {
    return &CLIError{
        Brief: fmt.Sprintf("Agent '%s' not found", name),
        Details: fmt.Sprintf("Available agents:\n  %s", strings.Join(available, ", ")),
        Suggestion: "Use 'ayo agents list' to see all agents",
        ExitCode: 3,
    }
}

func ErrDaemonNotRunning() *CLIError {
    return &CLIError{
        Brief: "Daemon is not running",
        Suggestion: "Start with 'ayo service start'",
        ExitCode: 5,
    }
}
```

### Apply to All Commands

Audit each command file:
- cmd/ayo/agents.go
- cmd/ayo/flows.go
- cmd/ayo/chat.go
- cmd/ayo/trigger.go
- etc.

Replace ad-hoc error handling with CLIError usage.

## Files to Create/Modify

1. `internal/cli/errors.go` - Error helper package
2. All cmd/ayo/*.go files - Use new error format

## Acceptance Criteria

- [ ] CLIError type with Brief, Details, Suggestion
- [ ] Common error constructors (NotFound, InvalidInput, etc.)
- [ ] Color coding for terminal output
- [ ] Exit codes consistent
- [ ] All CLI commands use CLIError
- [ ] Suggestions help users recover
- [ ] Errors tested in unit tests

