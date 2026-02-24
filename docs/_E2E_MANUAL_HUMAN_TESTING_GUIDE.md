# E2E Manual Human Testing Guide

Complete end-to-end testing guide for validating the ayo agent orchestration system. This guide is designed as a sequential walkthrough that builds on itself - if any step fails, return to a clean state and restart.

**Version:** GTM 1.0  
**Last Updated:** 2026-02-24

---

## Overview

### Purpose

This guide validates complete ayo system functionality through manual testing. Each section depends on previous sections succeeding.

### Approach

1. **Sequential**: Execute sections in order (0 → 11)
2. **Cumulative**: Each section builds on previous work
3. **Verifiable**: Clear pass/fail criteria at each step
4. **Recoverable**: Clean state script to restart from scratch

### If a Test Fails

1. Note which section and step failed
2. Run the clean state script (Section 0)
3. Run `ayo setup`
4. Restart from Section 1
5. Debug the failing section

---

## Section 0: Clean State & Prerequisites

### System Requirements

| Requirement | Minimum | Command to Check |
|-------------|---------|------------------|
| macOS | 26+ (Tahoe) | `sw_vers -productVersion` |
| CPU | Apple Silicon | `uname -m` (expect: arm64) |
| Go | 1.24+ | `go version` |
| RAM | 8GB | `sysctl -n hw.memsize` |
| Disk | 500MB free | `df -h ~/.local/share/` |

### Verify Prerequisites

```bash
#!/bin/bash
# verify-prerequisites.sh

echo "=== Ayo E2E Testing Prerequisites Check ==="
echo

echo "macOS Version:"
sw_vers -productVersion
echo

echo "CPU Architecture:"
uname -m
echo

echo "Go Version:"
go version
echo

echo "Container Runtime:"
if container list >/dev/null 2>&1; then
    echo "✓ Container runtime available"
else
    echo "✗ Container runtime NOT available"
    echo "  Install: macOS 26+ includes Apple Container runtime"
fi
echo

echo "Disk Space (~/.local/share/):"
df -h ~/.local/share/ 2>/dev/null || echo "Directory doesn't exist yet (OK)"
echo

echo "RAM:"
sysctl -n hw.memsize | awk '{printf "%.0f GB\n", $1/1024/1024/1024}'
echo

echo "=== Prerequisites Check Complete ==="
```

**Pass criteria**: All requirements met, container runtime available.

### Clean State Script

Run this before starting E2E tests or after a failure:

```bash
#!/bin/bash
# clean-state.sh

echo "=== Cleaning Ayo State ==="

# Stop daemon if running
echo "Stopping daemon..."
pkill -f ayod 2>/dev/null || true
rm -f ~/.local/share/ayo/daemon.sock
rm -f ~/.local/share/ayo/daemon.pid

# Remove all data directories
echo "Removing data directories..."
rm -rf ~/.local/share/ayo/sandboxes/
rm -rf ~/.local/share/ayo/sessions/
rm -rf ~/.local/share/ayo/memory/
rm -rf ~/.local/share/ayo/prompts/
rm -rf ~/.local/share/ayo/agents/
rm -rf ~/.local/share/ayo/triggers/

# Optionally remove config (uncomment if needed)
# echo "Removing configuration..."
# rm -rf ~/.config/ayo/

echo
echo "✓ Clean state achieved"
echo "Ready for fresh E2E testing"
echo
echo "Next: Run './ayo setup' to initialize"
```

### Provider Configuration

Before testing, ensure you have at least one LLM provider configured:

**Example `~/.config/ayo/ayo.json`:**

```json
{
  "providers": {
    "anthropic": {
      "api_key": "sk-ant-..."
    },
    "ollama": {
      "base_url": "http://localhost:11434"
    }
  },
  "defaults": {
    "provider": "anthropic",
    "model": "claude-sonnet-4-20250514",
    "embedding_provider": "ollama",
    "embedding_model": "nomic-embed-text"
  }
}
```

**Required:**
- At least one chat model provider with valid API key
- Embedding model for memory system (ollama recommended)

**Verify provider:**
```bash
# For Anthropic
echo "test" | curl -s https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{"model":"claude-sonnet-4-20250514","max_tokens":10,"messages":[{"role":"user","content":"hi"}]}' | head -c 100

# For Ollama
curl -s http://localhost:11434/api/tags | head -c 200
```

### Section 0 Checklist

- [ ] macOS 26+ confirmed
- [ ] Apple Silicon confirmed
- [ ] Go 1.24+ installed
- [ ] Container runtime works (`container list`)
- [ ] Clean state script runs without errors
- [ ] Provider API key available and valid
- [ ] Embedding model available

**Pass**: All boxes checked → Continue to Section 1

---

## Section 1: Build, Installation & Setup

### Build from Source

```bash
cd /path/to/ayo

# Build
go build -o ayo ./cmd/ayo/...

# Verify build
./ayo version
```

**Expected output:** Version information (e.g., `ayo version 1.0.0`)

```bash
# Verify help works
./ayo --help
```

**Expected**: Shows available commands and flags.

### Initial Setup

```bash
./ayo setup
```

**Expected outputs:**
- `✓ Created configuration directory`
- `✓ Created data directory`
- `✓ Installed default agents`
- `✓ Installed default prompts`

### Verify Configuration

```bash
# Check config exists
ls -la ~/.config/ayo/
cat ~/.config/ayo/ayo.json

# Check data directories
ls -la ~/.local/share/ayo/

# Check default agents installed
ls ~/.local/share/ayo/agents/
```

**Expected**: Config file exists, data directories created, agents present.

### Start Daemon

```bash
# Start daemon
./ayo sandbox service start

# Verify daemon running
./ayo sandbox service status
```

**Expected:** Status shows `Running` with PID.

```bash
# Verify socket exists
ls -l ~/.local/share/ayo/daemon.sock

# Test daemon connection
./ayo sandbox list
```

**Expected:** Socket file exists, sandbox list returns (empty is OK).

### Doctor Check

```bash
./ayo doctor
```

**Expected checks (all should pass):**
- ✓ Configuration valid
- ✓ Daemon running
- ✓ Container runtime available
- ✓ Provider configured
- ✓ Default agents installed

```bash
# Verbose for detailed info
./ayo doctor -v
```

### Section 1 Checklist

- [ ] Build succeeds without errors
- [ ] `ayo version` shows version info
- [ ] `ayo setup` creates config and data directories
- [ ] Daemon starts successfully
- [ ] `ayo doctor` shows all green

**Pass**: All boxes checked → Continue to Section 2

---

## Section 2: Agent Management

### List Default Agents

```bash
./ayo agents list
```

**Expected:** Shows `@ayo` (meta-agent) and other defaults.

### View Agent Details

```bash
./ayo agents show @ayo
```

**Expected output includes:**
- Agent handle: `@ayo`
- Description
- System prompt path or content
- Provider/model configuration

### Create Custom Agent

```bash
./ayo agents create @tester \
  --description "E2E test agent" \
  --prompt "You are a testing assistant. Help verify system functionality. Be concise." \
  --model claude-sonnet-4-20250514
```

**Verify creation:**
```bash
./ayo agents list
# Expected: @tester appears in list

./ayo agents show @tester
# Expected: Shows description, prompt, model
```

### Test Agent

```bash
./ayo @tester "Hello, what is your purpose?"
```

**Expected:** Agent responds mentioning testing/verification.

### Agent Configuration

```bash
# View agent config file
cat ~/.local/share/ayo/agents/tester.json
```

**Expected:** Valid JSON with agent configuration.

### Agent Inheritance (if supported)

```bash
./ayo agents create @tester-verbose \
  --description "Verbose tester" \
  --inherits @tester

./ayo agents show @tester-verbose
```

**Expected:** Shows inheritance from @tester.

### Remove Agent

```bash
./ayo agents rm @tester-verbose

./ayo agents list
```

**Expected:** @tester-verbose no longer appears.

### Section 2 Checklist

- [ ] Default agents present after setup
- [ ] Can view agent details
- [ ] Custom agent creation works
- [ ] Agent responds to prompts
- [ ] Agent configuration viewable
- [ ] Agent deletion works

**Pass**: All boxes checked → Continue to Section 3

---

## Section 3: Sessions & Chat

### Interactive Chat

```bash
./ayo
```

**Expected:** TUI opens, cursor ready for input.

**Test interaction:**
```
> What is 2+2?
```
**Expected:** Agent responds with "4" (or equivalent).

**Exit:**
```
> /exit
```
Or press `Ctrl+D`.

### Single Prompt (Quoted)

```bash
./ayo "What is the capital of France?"
```

**Expected:** "Paris" (returns to shell after response).

### Prompt Detection Rules

```bash
# Single lowercase word = command (error expected)
./ayo foobar
```
**Expected:** `Error: unknown command "foobar"`

```bash
# Quoted = always prompt
./ayo "foobar"
```
**Expected:** Agent responds (treating as prompt).

```bash
# Uppercase or multi-word = prompt
./ayo Hello
./ayo Hello world
```
**Expected:** Agent responds.

### Session Continuation

```bash
# First interaction
./ayo "Remember: the secret code is ALPHA-7"

# Continue same session
./ayo -c "What was the secret code?"
```

**Expected:** Agent recalls "ALPHA-7".

### File Attachments

```bash
# Create test file
echo "Hello from E2E test" > /tmp/e2e-test.txt

# Attach file to prompt
./ayo "What does this file contain?" -a /tmp/e2e-test.txt
```

**Expected:** Agent describes file contents ("Hello from E2E test").

### Session Listing

```bash
./ayo session list
```

**Expected:** Shows recent sessions with IDs and timestamps.

### Session Management

```bash
# Get a session ID from the list
./ayo session list

# View specific session (use actual ID)
./ayo session show <session-id>

# Delete session
./ayo session rm <session-id>

# Verify deletion
./ayo session list
```

### Multi-Turn Context

```bash
./ayo @tester
> My name is Alice
> What is my name?
```
**Expected:** Agent responds "Alice".

```
> /exit
```

### Section 3 Checklist

- [ ] Interactive chat works (TUI opens/closes)
- [ ] Single quoted prompts work
- [ ] Prompt detection rules work correctly
- [ ] Session continuation works (-c flag)
- [ ] File attachments work
- [ ] Session listing works
- [ ] Session deletion works
- [ ] Multi-turn context maintained

**Pass**: All boxes checked → Continue to Section 4

---

## Section 4: Memory System

### Prerequisites Check

```bash
./ayo doctor | grep -i embedding
```

**Expected:** Embedding provider configured.

### Store Facts

```bash
./ayo memory store "This E2E test uses PostgreSQL 15"
./ayo memory store "The API runs on port 8080"
./ayo memory store "Project lead is Alice"
./ayo memory store "Deployment happens on Fridays"
```

**Expected:** Each command confirms storage (ID returned or success message).

### List Memories

```bash
./ayo memory list
```

**Expected:** Shows all 4 stored memories with IDs and timestamps.

### Search Memories

```bash
# Semantic search
./ayo memory search "database"
```
**Expected:** Returns PostgreSQL fact.

```bash
./ayo memory search "who is the lead"
```
**Expected:** Returns Alice fact.

```bash
./ayo memory search "when do we deploy"
```
**Expected:** Returns Friday fact.

### Memory in Conversations

```bash
./ayo "What database does this project use?"
```
**Expected:** Agent mentions PostgreSQL (retrieved from memory).

```bash
./ayo "What port is the API on?"
```
**Expected:** Agent mentions 8080.

### Memory Deletion

```bash
# Get memory ID from list
./ayo memory list

# Delete specific memory
./ayo memory rm <memory-id>

# Verify deletion
./ayo memory list
```

**Expected:** Memory no longer appears.

### Section 4 Checklist

- [ ] Memory storage works (returns ID/confirmation)
- [ ] Memory listing shows stored facts
- [ ] Semantic search returns relevant results
- [ ] Memories appear in conversation context
- [ ] Memory deletion works

**Pass**: All boxes checked → Continue to Section 5

---

## Section 5: Tickets & Planning

### Create Tickets

```bash
# Create simple ticket
./ayo ticket create "Implement login endpoint"
```

**Expected:** Ticket ID returned (e.g., `ayo-a1b2`).

```bash
# Create with metadata
./ayo ticket create "Design API schema" \
  --assignee @ayo \
  --priority high \
  --tags api,design
```

**Save this ID for later steps (e.g., `SCHEMA_ID=ayo-xxxx`).**

### List Tickets

```bash
./ayo ticket list
```

**Expected:** Shows both tickets with status, assignee, priority.

### View Ticket Details

```bash
./ayo ticket show <SCHEMA_ID>
```

**Expected:** Full ticket details including title, status, assignee, dependencies.

### Ticket Dependencies

```bash
# Create dependent ticket (replace <SCHEMA_ID> with actual ID)
./ayo ticket create "Implement API endpoints" \
  --depends-on <SCHEMA_ID> \
  --assignee @tester

# Verify dependency
./ayo ticket show <new-id>
```

**Expected:** Shows dependency on schema ticket.

### Ready/Blocked Queries

```bash
# Check what's ready for @ayo
./ayo ticket ready --assignee @ayo
```
**Expected:** Shows "Design API schema" (no blockers).

```bash
# Check what's ready for @tester
./ayo ticket ready --assignee @tester
```
**Expected:** Empty (blocked by schema design).

```bash
# Check blocked tickets
./ayo ticket blocked
```
**Expected:** Shows "Implement API endpoints" blocked by schema.

### Ticket Workflow

```bash
# Start work on schema ticket
./ayo ticket start <SCHEMA_ID>

# Verify status
./ayo ticket show <SCHEMA_ID>
```
**Expected:** Status = `in_progress`

```bash
# Close ticket
./ayo ticket close <SCHEMA_ID>

# Verify status
./ayo ticket show <SCHEMA_ID>
```
**Expected:** Status = `closed`

```bash
# Check if dependent is now ready
./ayo ticket ready --assignee @tester
```
**Expected:** Shows "Implement API endpoints" (no longer blocked).

### Ticket Assignment

```bash
# Reassign ticket
./ayo ticket assign <dependent-id> @ayo

# Verify assignment
./ayo ticket show <dependent-id>
```

### Section 5 Checklist

- [ ] Ticket creation works (returns ID)
- [ ] Ticket listing shows all tickets
- [ ] Ticket details viewable
- [ ] Dependencies correctly block/unblock
- [ ] Ready queue reflects dependencies
- [ ] Status transitions work (start, close)
- [ ] Assignment works

**Pass**: All boxes checked → Continue to Section 6

---

## Section 6: Squads & Coordination

### Create Squad

```bash
./ayo squad create e2e-squad \
  --description "E2E testing squad"
```

**Expected:** Squad created successfully.

### Verify Squad Structure

```bash
ls -la ~/.local/share/ayo/sandboxes/squads/e2e-squad/
```

**Expected structure:**
```
├── SQUAD.md          ← Constitution
├── .tickets/         ← Squad-scoped tickets
├── .context/         ← Persistent state
├── workspace/        ← Shared code
└── agent-homes/      ← Per-agent directories
```

### Edit SQUAD.md Constitution

```bash
# View default SQUAD.md
cat ~/.local/share/ayo/sandboxes/squads/e2e-squad/SQUAD.md

# Edit constitution
cat > ~/.local/share/ayo/sandboxes/squads/e2e-squad/SQUAD.md << 'EOF'
---
lead: "@ayo"
planners:
  long_term: "ayo-tickets"
agents:
  - "@ayo"
  - "@tester"
---
# Squad: e2e-squad

## Mission

Test the complete ayo system including squad coordination, ticket management, and workspace operations.

## Context

- This is an E2E test squad
- All operations should be logged for verification
- Use workspace for any file outputs

## Agents

### @ayo
**Role**: Squad lead and coordinator
**Responsibilities**:
- Create tickets for work items
- Delegate to @tester
- Verify completion

### @tester
**Role**: Testing agent
**Responsibilities**:
- Execute test tasks
- Report results
- Create test files in workspace

## Coordination

1. @ayo creates tickets for test tasks
2. @tester picks up ready tickets
3. @tester completes work and closes tickets
4. @ayo verifies results

## Guidelines

- All output files go in workspace/
- Use descriptive ticket titles
- Mark tickets with [TEST] prefix
EOF
```

**Verify:**
```bash
cat ~/.local/share/ayo/sandboxes/squads/e2e-squad/SQUAD.md
```

### Add/Remove Agents

```bash
# Add agent
./ayo squad add-agent e2e-squad @tester

# Verify
./ayo squad show e2e-squad
```
**Expected:** Shows @tester in agent list.

```bash
# Remove and re-add (optional test)
./ayo squad remove-agent e2e-squad @tester
./ayo squad add-agent e2e-squad @tester
```

### Start/Stop Squad

```bash
# Start squad sandbox
./ayo squad start e2e-squad

# Verify running
./ayo squad status e2e-squad
```
**Expected:** Status = running

```bash
# Stop and restart (optional test)
./ayo squad stop e2e-squad
./ayo squad start e2e-squad
```

### Squad Ticket Management

```bash
# Create ticket in squad
./ayo squad ticket e2e-squad create "[TEST] Create hello.txt" \
  --assignee @tester

# List squad tickets
./ayo squad ticket e2e-squad list

# Show ticket details
./ayo squad ticket e2e-squad show <id>
```

### Dispatch to Squad

```bash
# Dispatch task to squad (uses #squad-name syntax)
./ayo "#e2e-squad" "Create a file called hello.txt containing 'Hello from E2E test'"
```

**Expected:** Squad receives task, potentially creates ticket, executes work.

### Verify Workspace Output

```bash
# Check workspace
ls ~/.local/share/ayo/sandboxes/squads/e2e-squad/workspace/

# Verify file contents (if created)
cat ~/.local/share/ayo/sandboxes/squads/e2e-squad/workspace/hello.txt
```
**Expected:** File exists with "Hello from E2E test" content.

### Squad Ticket Verification

```bash
./ayo squad ticket e2e-squad list
```

**Expected:** Shows tickets from squad operations.

### Section 6 Checklist

- [ ] Squad creation works
- [ ] Squad directory structure correct
- [ ] SQUAD.md constitution editable
- [ ] Agent add/remove works
- [ ] Squad start/stop works
- [ ] Squad tickets work
- [ ] Dispatch to squad works (#squad-name)
- [ ] Workspace output verified

**Pass**: All boxes checked → Continue to Section 7

---

## Section 7: Triggers & Scheduling

### Create Schedule Trigger

```bash
# Create schedule trigger (every minute for testing)
./ayo triggers schedule @tester "*/1 * * * *" \
  --prompt "Log a test message with timestamp"
```

**Expected:** Trigger ID returned.

### Create Watch Trigger

```bash
# Create watch directory
mkdir -p /tmp/e2e-watch

# Create watch trigger
./ayo triggers watch /tmp/e2e-watch @tester \
  --prompt "A file was changed in the watched directory"
```

**Expected:** Trigger ID returned.

### List Triggers

```bash
./ayo triggers list
```

**Expected:** Shows both triggers (schedule and watch).

### Trigger Details

```bash
./ayo triggers show <trigger-id>
```

**Expected:** Shows type, agent, prompt, schedule/path.

### Test Watch Trigger

```bash
# Create a file in watched directory
touch /tmp/e2e-watch/test-file.txt

# Wait a moment, then check sessions
sleep 2
./ayo session list
```

**Expected:** New session from trigger execution (may take a moment).

### Trigger Disable/Enable

```bash
# Get trigger ID
./ayo triggers list

# Disable trigger
./ayo triggers disable <trigger-id>

# Verify disabled
./ayo triggers list
```
**Expected:** Shows disabled status.

```bash
# Re-enable trigger
./ayo triggers enable <trigger-id>

# Verify enabled
./ayo triggers list
```

### Remove Triggers

```bash
# Remove both triggers
./ayo triggers rm <schedule-trigger-id>
./ayo triggers rm <watch-trigger-id>

# Verify removal
./ayo triggers list
```

**Expected:** Triggers no longer present.

```bash
# Cleanup watch directory
rm -rf /tmp/e2e-watch
```

### Section 7 Checklist

- [ ] Schedule trigger creation works
- [ ] Watch trigger creation works
- [ ] Trigger listing shows all triggers
- [ ] Trigger disable/enable works
- [ ] Watch trigger fires on file change
- [ ] Trigger removal works
- [ ] Cleanup completed

**Pass**: All boxes checked → Continue to Section 8

---

## Section 8: Plugins

### List Plugins

```bash
./ayo plugin list
```

**Expected:** Shows installed plugins with name, version, status.

### Plugin Info

```bash
./ayo plugin info ayo-tickets
```

**Expected:** Full description, configuration options, etc.

### Built-in Plugins

```bash
./ayo plugin list --builtin
```

**Expected built-in plugins:**
- `ayo-tickets` (long-term planner)
- `ayo-todos` (near-term planner)
- Others depending on system

### Plugin Enable/Disable

```bash
# Disable a plugin
./ayo plugin disable ayo-todos

# Verify disabled
./ayo plugin list
```
**Expected:** ayo-todos shows disabled.

```bash
# Re-enable
./ayo plugin enable ayo-todos

# Verify enabled
./ayo plugin list
```

### Plugin Configuration

```bash
# View plugin config (if supported)
./ayo plugin config ayo-tickets
```

### Plugin in Agent Context

```bash
# View how plugins are assigned to agents
./ayo agents show @ayo
```

**Expected:** Shows plugin configuration for agent.

### Section 8 Checklist

- [ ] Plugin listing works
- [ ] Plugin info shows details
- [ ] Plugin enable/disable works
- [ ] Plugin configuration accessible
- [ ] Built-in plugins listed

**Pass**: All boxes checked → Continue to Section 9

---

## Section 9: Full Orchestration Scenarios

### Scenario 1: Complex Multi-Step Task

```bash
./ayo @ayo "Create a simple calculator module in Python with add, subtract, multiply, divide functions. Put it in a file called calculator.py"
```

**Verification:**
```bash
# Check if file was created in workspace
ls ~/.local/share/ayo/workspace/
cat ~/.local/share/ayo/workspace/calculator.py
```

### Scenario 2: Memory-Driven Task

```bash
# Store context (if not already stored)
./ayo memory store "Always use type hints in Python code"
./ayo memory store "Follow PEP 8 style guidelines"

# Request code (should use context)
./ayo "Write a function to validate email addresses"
```

**Expected:** Response includes type hints and follows PEP 8 (influenced by memories).

### Scenario 3: Squad Feature Build

```bash
# Create a feature squad
./ayo squad create feature-test -a @ayo

# Start squad
./ayo squad start feature-test

# Dispatch task
./ayo "#feature-test" "Create a README.md file explaining what this squad does"

# Verify output
cat ~/.local/share/ayo/sandboxes/squads/feature-test/workspace/README.md

# Cleanup
./ayo squad stop feature-test
./ayo squad destroy feature-test --delete-data
```

### Scenario 4: Ticket-Driven Workflow

```bash
# Create dependent tickets
FIRST=$(./ayo ticket create "Define data model" --assignee @ayo --json | jq -r '.id')
./ayo ticket create "Implement data model" --assignee @ayo --depends-on $FIRST

# Check ready queue
./ayo ticket ready
```
**Expected:** Only "Define data model" is ready.

```bash
# Complete first ticket
./ayo ticket start $FIRST
./ayo ticket close $FIRST

# Check ready queue again
./ayo ticket ready
```
**Expected:** "Implement data model" is now ready.

### Section 9 Checklist

- [ ] Complex tasks complete successfully
- [ ] Memory context influences responses
- [ ] Squad feature build works end-to-end
- [ ] Ticket dependencies enforced correctly

**Pass**: All boxes checked → Continue to Section 10

---

## Section 10: Error Handling & Edge Cases

### Invalid Command

```bash
./ayo foobar
```
**Expected:** `Error: unknown command "foobar"`

### Daemon Not Running

```bash
# Stop daemon
./ayo sandbox service stop

# Try operation
./ayo sandbox list
```
**Expected:** Error about daemon not running.

```bash
# Recovery
./ayo sandbox service start
```

### Unknown Agent

```bash
./ayo @nonexistent "hello"
```
**Expected:** `Error: agent "@nonexistent" not found`

### Invalid Squad

```bash
./ayo "#no-such-squad" "test"
```
**Expected:** `Error: squad "no-such-squad" not found`

### Invalid Ticket

```bash
./ayo ticket show xyz-9999
```
**Expected:** `Error: ticket not found`

### Permission Errors (Sandbox)

```bash
# Try to access protected path (in sandbox)
./ayo sandbox list
# Get sandbox ID
./ayo sandbox exec <id> "cat /etc/shadow"
```
**Expected:** Permission denied or error.

### Section 10 Checklist

- [ ] Invalid commands show helpful errors
- [ ] Missing daemon clearly reported
- [ ] Unknown agent error works
- [ ] Invalid squad error works
- [ ] Invalid ticket error works
- [ ] Recovery from errors possible

**Pass**: All boxes checked → Continue to Section 11

---

## Section 11: Cleanup & Final Verification

### Remove Test Squads

```bash
./ayo squad list

# Destroy test squads
./ayo squad stop e2e-squad 2>/dev/null
./ayo squad destroy e2e-squad --delete-data 2>/dev/null

./ayo squad list
```
**Expected:** Test squads removed.

### Remove Test Agents

```bash
./ayo agents list

# Remove test agent
./ayo agents rm @tester

./ayo agents list
```
**Expected:** Only default agents remain.

### Clear Test Tickets

```bash
./ayo ticket list

# Close remaining test tickets
# (or remove with ./ayo ticket rm <id>)
```

### Clear Test Memories

```bash
./ayo memory list

# Remove test memories or clear all
./ayo memory clear --confirm
```

### Remove Temporary Files

```bash
rm -rf /tmp/e2e-*
rm -f /tmp/test.txt
```

### Final Health Check

```bash
./ayo doctor -v
```

**Expected:** All checks pass.

### Final Verification Checklist

| Item | Status |
|------|--------|
| All test squads destroyed | ☐ |
| All test agents removed | ☐ |
| All test tickets cleared | ☐ |
| All test memories cleared | ☐ |
| Temp files cleaned | ☐ |
| Daemon running | ☐ |
| Doctor passes | ☐ |

---

## E2E Testing Complete

### If All Sections Passed

**Congratulations!** The ayo system has been verified working end-to-end.

The following capabilities have been validated:
- ✓ Build and installation
- ✓ Agent management
- ✓ Sessions and chat
- ✓ Memory system
- ✓ Tickets and planning
- ✓ Squads and coordination
- ✓ Triggers and scheduling
- ✓ Plugins
- ✓ Full orchestration
- ✓ Error handling

### If Any Section Failed

1. Note which section and step failed
2. Run `clean-state.sh` from Section 0
3. Run `./ayo setup`
4. Restart from Section 1
5. Debug the failing section
6. Report issues with:
   - Section number
   - Step that failed
   - Expected vs actual behavior
   - Error messages

---

## Appendix: Quick Reference

### Common Commands

| Action | Command |
|--------|---------|
| Start daemon | `ayo sandbox service start` |
| Stop daemon | `ayo sandbox service stop` |
| Health check | `ayo doctor` |
| List agents | `ayo agents list` |
| Chat with agent | `ayo @agent` or `ayo "prompt"` |
| List squads | `ayo squad list` |
| Dispatch to squad | `ayo "#squad-name" "task"` |
| List tickets | `ayo ticket list` |
| List triggers | `ayo triggers list` |
| Search memory | `ayo memory search "query"` |

### Key Paths

| Path | Contents |
|------|----------|
| `~/.config/ayo/` | Configuration |
| `~/.local/share/ayo/` | All data |
| `~/.local/share/ayo/agents/` | Agent definitions |
| `~/.local/share/ayo/sandboxes/` | Squad sandboxes |
| `~/.local/share/ayo/memory/` | Vector memory store |
| `~/.local/share/ayo/sessions/` | Chat sessions |

### Troubleshooting

| Issue | Solution |
|-------|----------|
| Daemon won't start | `rm -f ~/.local/share/ayo/daemon.sock && ayo sandbox service start` |
| "Method not found" | Restart daemon to pick up code changes |
| Squad dispatch hangs | Check LLM provider: `ayo "test"` |
| Memory search empty | Verify embedding provider configured |
| Triggers not firing | Check `ayo triggers list` for status |
