---
id: ase-0hi0
status: open
deps: [ase-euxv]
links: []
created: 2026-02-09T03:07:46Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-k48b
---
# Implement ayo flows CLI commands

Add CLI commands for flow management: list, show, run, create, edit, delete.

## Commands

```bash
# List flows
ayo flows list                      # List all flows
ayo flows list --json               # JSON output

# Show flow details
ayo flows show daily-digest         # Show flow spec
ayo flows show daily-digest --json  # JSON output

# Run flow
ayo flows run daily-digest                    # Run with defaults
ayo flows run daily-digest --param lang=es    # With parameters
ayo flows run daily-digest -i input.json      # Input from file
echo '{...}' | ayo flows run daily-digest     # Input from stdin

# Create flow (interactive - @ayo generates)
ayo flows create                    # Interactive creation
ayo flows create --from-session X   # Create from recent session

# Edit flow
ayo flows edit daily-digest         # Open in $EDITOR

# Delete flow
ayo flows delete daily-digest       # Delete with confirmation
ayo flows delete daily-digest -f    # Force delete
```

## Implementation

1. Add flows.go command file with subcommands
2. `list`: Scan ~/.config/ayo/flows/ for YAML files
3. `show`: Parse and display flow
4. `run`: 
   - Parse flow
   - Validate input
   - Connect to daemon
   - Execute via daemon RPC (daemon has Matrix connection)
   - Stream output
5. `create`:
   - Interactive: Ask for description, @ayo generates flow
   - From session: Analyze recent session, extract pattern
6. `edit`: Open file in editor
7. `delete`: Remove file with confirmation

## Daemon RPC

Add FlowRun RPC method:
```go
type FlowRunParams struct {
    FlowName string
    Params   map[string]any
}

type FlowRunResult struct {
    RunID   string
    Status  string
    Output  any
    Error   string
}
```

## Files to create/modify

- cmd/ayo/flows.go (rewrite or extend existing)
- internal/daemon/protocol.go (add FlowRun RPC)
- internal/daemon/server.go (add handler)

## Acceptance Criteria

- All commands work as documented
- JSON output is valid
- Run streams output in real-time
- Create generates valid flow via @ayo
- Edit opens correct file
- Delete requires confirmation unless -f

