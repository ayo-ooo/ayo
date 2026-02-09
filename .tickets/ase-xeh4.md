---
id: ase-xeh4
status: closed
deps: []
links: []
created: 2026-02-09T03:26:24Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-zlew
---
# Update CLI reference documentation

## Background

The CLI reference documentation (docs/cli-reference.md) needs to be updated to reflect all new commands from the agent orchestration system.

## Why This Matters

Documentation that doesn't match the CLI:
- Frustrates users trying to learn
- Causes support requests
- Makes the tool appear unprofessional

## New Commands to Document

### Matrix/Chat Commands

```bash
ayo chat rooms                    # List rooms
ayo chat create                   # Create room
ayo chat send                     # Send message
ayo chat read                     # Read messages
ayo chat history                  # Full history
ayo chat who                      # Room members
ayo chat invite                   # Invite to room
ayo chat join                     # Join room
ayo chat leave                    # Leave room
```

### Flow Commands

```bash
ayo flows list                    # List flows
ayo flows show <name>             # Show flow details
ayo flows run <name>              # Execute flow
ayo flows create                  # Create new flow
ayo flows edit <name>             # Edit flow
ayo flows delete <name>           # Delete flow
ayo flows history                 # Execution history
```

### Updated Agent Commands

```bash
ayo agents list                   # Updated with trust/created-by columns
ayo agents show <name>            # Updated with metadata
ayo agents create                 # New - create agent (internal)
ayo agents refine <name>          # New - refine agent
ayo agents capabilities <name>    # New - show/search capabilities
```

### Updated Trigger Commands

```bash
ayo trigger list                  # Renamed from triggers
ayo trigger schedule              # New - cron triggers
ayo trigger watch                 # New - file triggers
ayo trigger show <id>             # Show details
ayo trigger rm <id>               # Remove trigger
ayo trigger enable <id>           # Enable trigger
ayo trigger disable <id>          # Disable trigger
ayo trigger test <id>             # Test trigger
```

### Database Commands

```bash
ayo db status                     # Migration status
ayo db migrate                    # Run migrations
ayo db history                    # Migration history
```

## Documentation Format

Each command should have:
1. Synopsis (usage line)
2. Description (what it does)
3. Arguments (positional)
4. Flags (with types and defaults)
5. Examples (2-3 realistic examples)
6. See Also (related commands)

## Example Entry

```markdown
### ayo flows run

Run a flow with optional parameters.

**Synopsis**
\`\`\`
ayo flows run <name> [flags]
\`\`\`

**Arguments**
- `name` - Flow name or path to flow file

**Flags**
- `--param, -p` - Set flow parameter (can be repeated)
- `--input, -i` - Read input from file
- `--async` - Run in background, return immediately
- `--json` - Output in JSON format

**Examples**
\`\`\`bash
# Run flow with defaults
ayo flows run daily-digest

# Run with parameters
ayo flows run daily-digest -p lang=spanish -p format=html

# Run with input file
ayo flows run process-data -i data.json
\`\`\`

**See Also**
- ayo flows list
- ayo flows show
\`\`\`
```

## Files to Update

1. `docs/cli-reference.md` - Main reference document
2. `README.md` - Update quick start examples if needed

## Acceptance Criteria

- [ ] All chat commands documented
- [ ] All flows commands documented
- [ ] Updated agents commands documented
- [ ] Updated trigger commands documented
- [ ] Database commands documented
- [ ] Each command has synopsis, description, args, flags, examples
- [ ] Examples are realistic and tested
- [ ] See Also cross-references correct
- [ ] TOC updated with new sections

