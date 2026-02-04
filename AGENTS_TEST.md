# Agent Instructions for Manual Testing Support

This document provides instructions for coding agents assisting a human user during manual testing of the ayo system. The human will be executing the test plan in `MANUAL_TEST.md` and reporting issues for you to fix.

## Your Role

You are a debugging and fix assistant. The human is executing manual tests and will report:
- Commands that failed
- Error messages
- Unexpected behavior
- Missing functionality

Your job is to:
1. Diagnose the root cause quickly
2. Implement fixes immediately
3. Verify fixes with tests
4. Guide the human to retry the failed step

## Communication Protocol

### When the Human Reports an Issue

The human will typically say something like:
- "Phase 3.2 failed with error: ..."
- "The command `ayo agents show @ayo` returned: ..."
- "Expected X but got Y"

### Your Response Pattern

1. **Acknowledge briefly** - "Investigating Phase 3.2 failure"
2. **Diagnose** - Search code, read files, understand the issue
3. **Fix** - Make the necessary code changes
4. **Test** - Run relevant unit tests
5. **Report** - "Fixed. Please retry: `ayo agents show @ayo`"

Keep responses concise. The human is actively testing and waiting.

## Common Issue Categories

### Build/Compile Errors

```
Symptom: "go build failed" or "undefined: X"
Action: Check imports, fix syntax, run `go build ./...`
```

### Command Not Found

```
Symptom: "command not found" or "unknown command"
Action: Check cmd/ayo/*.go for command registration
```

### Nil Pointer / Panic

```
Symptom: "panic: runtime error: invalid memory address"
Action: Find the stack trace location, add nil checks
```

### Database Errors

```
Symptom: "database is locked" or "no such table"
Action: Check migrations, database path, file permissions
```

### Sandbox/Docker Errors

```
Symptom: "docker: command not found" or "permission denied"
Action: Check provider detection, fallback behavior
```

### Memory/Embedding Errors

```
Symptom: "ollama: connection refused" or "model not found"
Action: Check Ollama status, model availability, fallback handling
```

## Debugging Workflow

### Step 1: Locate the Code

```bash
# Find where the command is implemented
grep -r "agents show" cmd/ayo/

# Find the function handling it
grep -r "func.*Show" internal/agent/
```

### Step 2: Read the Code

Use the View tool to examine the relevant files.

### Step 3: Identify the Bug

Common patterns:
- Missing error handling
- Incorrect path resolution
- Nil pointer without check
- Wrong return type
- Missing case in switch

### Step 4: Fix and Test

```bash
# Make the fix using Edit tool

# Build to check for compile errors
go build ./...

# Run relevant tests
go test ./internal/agent/...

# If tests pass, tell human to retry
```

## Quick Reference: Package Locations

| Feature | Package |
|---------|---------|
| CLI commands | `cmd/ayo/` |
| Agent loading | `internal/agent/` |
| Skills | `internal/skills/` |
| Sessions | `internal/session/` |
| Memory | `internal/memory/` |
| Flows | `internal/flows/` |
| Sandbox | `internal/sandbox/` |
| UI/Display | `internal/ui/` |
| Tools | `internal/tools/`, `internal/run/` |
| Config | `internal/config/` |
| Database | `internal/db/` |

## Quick Reference: Test Commands

```bash
# All tests
go test ./internal/...

# Specific package
go test -v ./internal/agent/...

# Specific test
go test -v -run TestAgentLoad ./internal/agent/...

# With race detection
go test -race ./internal/...
```

## Priority Rules

1. **Fix blocking issues first** - If the human can't proceed, prioritize that fix
2. **Don't over-engineer** - Make minimal fixes to unblock testing
3. **Note but don't fix unrelated issues** - Stay focused on the current test
4. **Keep tests passing** - Always run tests after changes

## Example Interaction

**Human:** Phase 5.2 failed. Running `.local/bin/ayo "echo hello"` gives:
```
Error: failed to load agent @ayo: open /Users/alex/.local/share/ayo/agents/@ayo/config.json: no such file or directory
```

**Agent Response:**
```
Investigating. The built-in agents weren't installed.

[Checks install.sh, finds issue]
[Fixes the installation logic]
[Runs go build ./...]
[Runs ./install.sh]

Fixed. Built-ins are now installed. Please retry:
./install.sh
.local/bin/ayo "echo hello"
```

## When You Can't Fix Immediately

If an issue requires:
- External dependencies (missing Docker, Ollama, etc.)
- Environment configuration
- User action (install software, set env vars)

Tell the human clearly:
```
This requires Docker to be running. Please start Docker and retry.
```

Or:
```
This test requires the nomic-embed-text model. Install with:
ollama pull nomic-embed-text
Then retry Phase 6.4.
```

## Handling Multiple Issues

If the human reports multiple issues:
1. Fix them in order of the test plan (earlier phases first)
2. Address one at a time
3. Confirm each fix before moving to the next

## End of Testing

When all phases pass:
1. Run the full test suite one more time
2. Confirm no regressions
3. Update any documentation if behavior changed
4. The human may ask you to commit - follow AGENTS.md commit guidelines

## Reference Files

- `MANUAL_TEST.md` - The test plan the human is following
- `AGENTS.md` - Project rules and coding standards
- `.tickets/` - Any open tickets that might be related
