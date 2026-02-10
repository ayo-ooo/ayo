# Ayo Manual Testing Guide

This guide walks through testing all ayo CLI commands in a logical order, starting with system prerequisites and building up to complex multi-agent scenarios.

## Prerequisites

Before testing, ensure:
- macOS 26+ with Apple Silicon (for sandbox features)
- Go 1.21+ installed
- Built binary: `go build -o ayo ./cmd/ayo/...`

---

## 1. System Health & Setup

Start here to verify the system is ready.

### 1.1 Doctor Check

```bash
# Basic health check
ayo doctor

# Verbose output for debugging
ayo doctor -v
```

**Expected:** All checks pass (green). Note any warnings about missing optional dependencies.

### 1.2 Initial Setup

```bash
# Run setup wizard
ayo setup

# Force re-setup (overwrites modifications)
ayo setup --force
```

**Expected:** Creates `~/.config/ayo/` directory with agents, skills, and provider configs.

---

## 2. Service Management

The sandbox service must be running for agent execution.

### 2.1 Start Service

```bash
# Start background service
ayo sandbox service start

# Or run in foreground for debugging
ayo sandbox service start -f
```

**Expected:** Service starts, creates `/tmp/ayo/daemon.sock`.

### 2.2 Check Status

```bash
ayo sandbox service status
```

**Expected:** Shows "running" with PID and uptime.

### 2.3 Stop Service

```bash
ayo sandbox service stop
```

**Expected:** Service stops cleanly.

### 2.4 Restart for Remaining Tests

```bash
ayo sandbox service start
```

---

## 3. Agent Management

### 3.1 List Agents

```bash
# List all agents
ayo agents list

# Filter by trust level
ayo agents list --trust full

# Filter by type
ayo agents list --type builtin
ayo agents list --type user
```

**Expected:** Shows built-in agents (@ayo, @code, etc.) with handles, models, and trust levels.

### 3.2 Show Agent Details

```bash
ayo agents show @ayo
ayo agents show @code
```

**Expected:** Displays full agent configuration including system prompt, tools, and skills.

### 3.3 Create Custom Agent

```bash
# Minimal agent
ayo agents create @test-agent -d "Test agent for manual testing"

# Agent with specific model
ayo agents create @test-model -m claude-sonnet-4-20250514 -d "Sonnet agent"

# Agent with custom system prompt
ayo agents create @test-prompt -s "You are a helpful assistant that speaks like a pirate."

# Agent with specific tools
ayo agents create @test-tools -t "bash,edit,view" -d "Limited tools agent"
```

**Expected:** Agent files created in `~/.config/ayo/agents/`.

### 3.4 Agent Capabilities

```bash
# Show capabilities for one agent
ayo agents capabilities @ayo

# List all agents with capabilities
ayo agents capabilities --all

# Search capabilities
ayo agents capabilities --search "code review"

# Refresh capabilities
ayo agents capabilities refresh @ayo
ayo agents capabilities refresh --all
```

**Expected:** Shows what each agent can do based on skills and tools.

### 3.5 Update Built-in Agents

```bash
# Update to latest versions
ayo agents update

# Force overwrite
ayo agents update --force
```

**Expected:** Built-in agents updated from embedded defaults.

### 3.6 Remove Agent

```bash
# With confirmation
ayo agents rm @test-agent

# Skip confirmation
ayo agents rm @test-model --force

# Dry run
ayo agents rm @test-tools --dry-run
```

**Expected:** Agent files removed (or would be removed for dry-run).

---

## 4. Skills Management

### 4.1 List Skills

```bash
ayo skills list
```

**Expected:** Shows available skills with descriptions and sources.

### 4.2 Show Skill Details

```bash
# Pick a skill from the list
ayo skills show <skill-name>
```

**Expected:** Shows skill configuration, tools, and content.

### 4.3 Create Skill

```bash
# Create in user skills directory
ayo skills create test-skill

# Create in current directory
ayo skills create local-skill --local

# Create for specific agent
ayo skills create agent-skill -a @ayo
```

**Expected:** Skill directory created with template files.

### 4.4 Validate Skill

```bash
ayo skills validate ~/.config/ayo/skills/test-skill
```

**Expected:** Reports any validation errors or confirms skill is valid.

### 4.5 Update Built-in Skills

```bash
ayo skills update
ayo skills update --force
```

**Expected:** Built-in skills updated from embedded defaults.

---

## 5. Share System

The share system provides instant file access to sandboxes via symlinks.

### 5.1 Share a Directory

```bash
# Share current directory
ayo share .

# Share with custom name
ayo share ~/Code/myproject --as project

# Share home directory
ayo share ~

# Session-only share (removed when session ends)
ayo share /tmp/testdata --session
```

**Expected:** Symlink created in `~/.local/share/ayo/sandbox/workspace/`.

### 5.2 List Shares

```bash
# List all shares
ayo share list

# JSON output
ayo share list --json
```

**Expected:** Shows share name, host path, workspace path, and whether permanent (●) or session (○).

### 5.3 Verify Share in Sandbox

```bash
# Get sandbox ID
ayo sandbox list

# Check share is visible
ayo sandbox exec <id> ls -la /workspace/
```

**Expected:** Shared directories appear in `/workspace/`.

### 5.4 Remove Share

```bash
# Remove by name
ayo share rm project

# Remove by path
ayo share rm ~/Code/myproject

# Remove all shares
ayo share rm --all
```

**Expected:** Symlink removed, share no longer appears in list.

---

## 6. Mount System (Legacy)

> **Note:** The mount system is being replaced by shares. Use shares for new workflows.

### 6.1 Add Mount

```bash
# Read-write access
ayo mount add ~/Code/project

# Read-only access
ayo mount add ~/Documents --ro
```

**Expected:** Mount grant recorded in `~/.local/share/ayo/mounts.json`.

### 6.2 List Mounts

```bash
ayo mount list
ayo mount list --json
```

**Expected:** Shows granted paths with access levels.

### 6.3 Remove Mount

```bash
ayo mount rm ~/Code/project
ayo mount rm --all
```

**Expected:** Mount grants removed.

---

## 7. Sandbox Operations

### 7.1 List Sandboxes

```bash
ayo sandbox list
ayo sandbox ls
```

**Expected:** Shows active sandboxes with ID, status, agents, and uptime.

### 7.2 Show Sandbox Details

```bash
ayo sandbox show --id <sandbox-id>
```

**Expected:** Detailed sandbox info including mounts, agents, and resource usage.

### 7.3 Execute Commands

```bash
# Run as root
ayo sandbox exec <id> whoami

# Run as specific user
ayo sandbox exec <id> -u agent-ayo whoami

# Run in specific directory
ayo sandbox exec <id> -w /workspace pwd
```

**Expected:** Command output from inside container.

### 7.4 Shell Access

```bash
# Non-interactive shell (for scripting)
ayo sandbox shell --id <id>

# As specific agent
ayo sandbox shell --id <id> --as @ayo

# Interactive login shell (human use)
ayo sandbox login --id <id>
```

**Expected:** Shell access to container.

### 7.5 File Transfer

```bash
# Create test file
echo "test content" > /tmp/testfile.txt

# Push to sandbox
ayo sandbox push <id> /tmp/testfile.txt /tmp/testfile.txt

# Verify
ayo sandbox exec <id> cat /tmp/testfile.txt

# Modify in sandbox
ayo sandbox exec <id> sh -c 'echo "modified" >> /tmp/testfile.txt'

# Pull back
ayo sandbox pull <id> /tmp/testfile.txt /tmp/pulled.txt
cat /tmp/pulled.txt
```

**Expected:** File transfers work bidirectionally.

### 7.6 Working Copy Sync

```bash
# Share a directory first
ayo share ~/Code/testproject --as testproject

# Make changes inside sandbox
ayo sandbox exec <id> sh -c 'echo "new file" > /workspace/testproject/newfile.txt'

# Show differences
ayo sandbox diff <id> /workspace/testproject ~/Code/testproject

# Sync changes back (dry run first)
ayo sandbox sync <id> /workspace/testproject ~/Code/testproject --dry-run

# Actually sync
ayo sandbox sync <id> /workspace/testproject ~/Code/testproject
```

**Expected:** Changes from sandbox synced to host.

### 7.7 Resource Stats

```bash
ayo sandbox stats --id <id>
```

**Expected:** Shows CPU, memory, and disk usage.

### 7.8 Logs

```bash
# Recent logs
ayo sandbox logs --id <id>

# Follow logs
ayo sandbox logs --id <id> -f

# Last N lines
ayo sandbox logs --id <id> -n 50
```

**Expected:** Container logs displayed.

### 7.9 Multi-Agent Sandbox

```bash
# Add another agent to existing sandbox
ayo sandbox join <id> @code

# List agents in sandbox
ayo sandbox users --id <id>
```

**Expected:** Multiple agents share the sandbox.

### 7.10 Lifecycle Control

```bash
# Stop sandbox
ayo sandbox stop --id <id>

# Start it again
ayo sandbox start --id <id>

# Force stop
ayo sandbox stop --id <id> --force
```

**Expected:** Sandbox starts and stops cleanly.

### 7.11 Cleanup

```bash
# Remove stopped sandboxes
ayo sandbox prune

# Force without confirmation
ayo sandbox prune --force

# Also remove agent home directories
ayo sandbox prune --homes
```

**Expected:** Stopped sandboxes removed.

---

## 8. Triggers

### 8.1 List Triggers

```bash
ayo triggers list
```

**Expected:** Shows registered triggers (may be empty initially).

### 8.2 Create Schedule Trigger

```bash
# Run every hour
ayo triggers schedule @ayo "0 * * * *" -p "Check system status"

# Run daily at 9am
ayo triggers schedule @ayo "0 9 * * *" -p "Good morning report"
```

**Expected:** Trigger created with ID.

### 8.3 Create Watch Trigger

```bash
# Watch directory for changes
ayo triggers watch ~/Code/project @code -p "Review changes"

# Watch with specific patterns
ayo triggers watch ~/Code/project @code "*.go" "*.md" -p "Files changed"

# Watch recursively
ayo triggers watch ~/Code/project @code -r -p "Something changed"

# Watch specific events
ayo triggers watch ~/Code/project @code --events create,modify -p "New or modified files"
```

**Expected:** Watch trigger created.

### 8.4 Show Trigger Details

```bash
ayo triggers show <trigger-id>
```

**Expected:** Full trigger configuration.

### 8.5 Test Trigger

```bash
ayo triggers test <trigger-id>
```

**Expected:** Trigger fires immediately (useful for debugging).

### 8.6 Enable/Disable

```bash
ayo triggers disable <trigger-id>
ayo triggers enable <trigger-id>
```

**Expected:** Trigger state changes.

### 8.7 Remove Trigger

```bash
ayo triggers rm <trigger-id>
ayo triggers rm <trigger-id> --force
```

**Expected:** Trigger removed.

---

## 9. Session Management

### 9.1 Run Agent (Creates Session)

```bash
# Interactive chat
ayo @ayo

# Single prompt
ayo @ayo "What time is it?"

# With attachments
ayo @ayo "Review this file" -a README.md

# With debug output
ayo @ayo "Hello" --debug
```

**Expected:** Agent responds, session created.

### 9.2 List Sessions

```bash
# All sessions
ayo sessions list

# Filter by agent
ayo sessions list -a @ayo

# Limit results
ayo sessions list -n 5
```

**Expected:** Shows session history with IDs, agents, and timestamps.

### 9.3 Show Session

```bash
# Show specific session
ayo sessions show <session-id>

# Show most recent
ayo sessions show --latest
```

**Expected:** Full session details including messages.

### 9.4 Continue Session

```bash
# Continue specific session
ayo sessions continue <session-id>

# Continue most recent
ayo -c
ayo sessions continue --latest
```

**Expected:** Resumes previous conversation.

### 9.5 Delete Session

```bash
ayo sessions delete <session-id>
ayo sessions delete --latest --force
```

**Expected:** Session removed from history.

### 9.6 Session Maintenance

```bash
# Migrate old format (if needed)
ayo sessions migrate

# Rebuild index
ayo sessions reindex
ayo sessions reindex --full
```

**Expected:** Maintenance operations complete.

---

## 10. Memory System

### 10.1 Store Memory

```bash
# Basic memory
ayo memory store "The project uses PostgreSQL for the database"

# With category
ayo memory store "API keys are in .env" -c secrets

# Scoped to agent
ayo memory store "User prefers dark mode" -a @ayo

# Scoped to path
ayo memory store "This project uses pnpm" -p ~/Code/myproject
```

**Expected:** Memory stored with auto-generated ID.

### 10.2 List Memories

```bash
# All memories
ayo memory list

# Filter by agent
ayo memory list -a @ayo

# Filter by category
ayo memory list -c preferences

# Limit results
ayo memory list -n 10
```

**Expected:** Shows memories with IDs and content previews.

### 10.3 Search Memories

```bash
# Semantic search
ayo memory search "database configuration"

# With threshold
ayo memory search "API" -t 0.5

# Limit results
ayo memory search "preferences" -n 5
```

**Expected:** Relevant memories ranked by similarity.

### 10.4 Show Memory

```bash
ayo memory show <memory-id>
```

**Expected:** Full memory details.

### 10.5 Memory Topics

```bash
ayo memory topics
```

**Expected:** Shows topic clusters in memory.

### 10.6 Link Memories

```bash
ayo memory link <id1> <id2>
```

**Expected:** Creates connection between memories.

### 10.7 Merge Similar

```bash
# Dry run
ayo memory merge --dry-run

# Actually merge
ayo memory merge --threshold 0.9
```

**Expected:** Consolidates duplicate/similar memories.

### 10.8 Forget Memory

```bash
ayo memory forget <memory-id>
ayo memory forget <memory-id> --force
```

**Expected:** Memory soft-deleted.

### 10.9 Memory Stats

```bash
ayo memory stats
ayo memory stats --json
```

**Expected:** Shows memory counts and distribution.

### 10.10 Clear Memories

```bash
# Clear for specific agent
ayo memory clear -a @ayo --force

# Clear all (dangerous!)
ayo memory clear --force
```

**Expected:** Memories removed.

### 10.11 Maintenance

```bash
ayo memory reindex
ayo memory migrate  # If upgrading from old format
```

**Expected:** Index rebuilt.

---

## 11. Flows

### 11.1 List Flows

```bash
ayo flows list
ayo flows list --source user
ayo flows list --json
```

**Expected:** Shows available flows with sources.

### 11.2 Show Flow

```bash
ayo flows show <flow-name>
ayo flows show <flow-name> --script
```

**Expected:** Flow configuration and steps.

### 11.3 Create Flow

```bash
# Create in user flows
ayo flows new my-flow

# Create in project
ayo flows new project-flow --project

# With schemas
ayo flows new typed-flow --with-schemas
```

**Expected:** Flow YAML file created.

### 11.4 Validate Flow

```bash
ayo flows validate ~/.config/ayo/flows/my-flow.yaml
```

**Expected:** Reports validation errors or confirms valid.

### 11.5 Run Flow

```bash
# Run with inline input
ayo flows run my-flow '{"key": "value"}'

# Run with input file
ayo flows run my-flow -i input.json

# Validate only
ayo flows run my-flow --validate

# With timeout
ayo flows run my-flow -t 600
```

**Expected:** Flow executes, shows output.

### 11.6 Flow History

```bash
# List runs
ayo flows history
ayo flows history --flow my-flow
ayo flows history --status success

# Show specific run
ayo flows history show <run-id>
```

**Expected:** Shows execution history.

### 11.7 Replay Flow

```bash
ayo flows replay <run-id>
```

**Expected:** Re-runs flow with same inputs.

### 11.8 Flow Stats

```bash
ayo flows stats
ayo flows stats my-flow
```

**Expected:** Shows success rates, timing, etc.

### 11.9 Remove Flow

```bash
ayo flows rm my-flow
ayo flows rm my-flow --force
```

**Expected:** Flow removed.

---

## 12. Agent Chaining

### 12.1 List Chainable Agents

```bash
ayo chain ls
ayo chain ls --json
```

**Expected:** Shows agents with input/output schemas.

### 12.2 Inspect Schemas

```bash
ayo chain inspect @code
```

**Expected:** Shows input and output schemas.

### 12.3 Check Compatibility

```bash
# What can receive output from @code?
ayo chain from @code

# What can send to @code?
ayo chain to @code
```

**Expected:** Shows compatible agents.

### 12.4 Validate Input

```bash
ayo chain validate @code '{"files": ["main.go"]}'
```

**Expected:** Reports if input matches schema.

### 12.5 Generate Example

```bash
ayo chain example @code
```

**Expected:** Generates valid example input JSON.

---

## 13. Plugins

### 13.1 List Plugins

```bash
ayo plugins list
```

**Expected:** Shows installed plugins (may be empty).

### 13.2 Install Plugin

```bash
# From git
ayo plugins install https://github.com/user/ayo-plugin

# From local directory
ayo plugins install --local ~/Code/my-plugin

# Force reinstall
ayo plugins install https://github.com/user/ayo-plugin --force
```

**Expected:** Plugin installed to `~/.config/ayo/plugins/`.

### 13.3 Show Plugin

```bash
ayo plugins show <plugin-name>
```

**Expected:** Plugin details and capabilities.

### 13.4 Update Plugins

```bash
# Update all
ayo plugins update

# Update specific
ayo plugins update <plugin-name>

# Dry run
ayo plugins update --dry-run
```

**Expected:** Plugins updated.

### 13.5 Remove Plugin

```bash
ayo plugins remove <plugin-name>
ayo plugins rm <plugin-name> -y
```

**Expected:** Plugin uninstalled.

---

## 14. Backup & Sync

### 14.1 Create Backup

```bash
ayo backup create
ayo backup create --name "before-upgrade"
```

**Expected:** Backup created with timestamp.

### 14.2 List Backups

```bash
ayo backup list
ayo backup list --json
```

**Expected:** Shows available backups.

### 14.3 Show Backup

```bash
ayo backup show <backup-name>
```

**Expected:** Backup details and contents.

### 14.4 Restore Backup

```bash
ayo backup restore <backup-name>
ayo backup restore <backup-name> --no-safety
```

**Expected:** State restored from backup.

### 14.5 Export/Import

```bash
# Export to archive
ayo backup export <backup-name> ~/backup.tar.gz

# Import from archive
ayo backup import ~/backup.tar.gz
```

**Expected:** Backup transferred as archive file.

### 14.6 Prune Old Backups

```bash
ayo backup prune
ayo backup prune --keep 5
```

**Expected:** Old auto-backups removed.

### 14.7 Sync Init

```bash
ayo sync init
ayo sync init --branch
```

**Expected:** Git repo initialized in sandbox directory.

### 14.8 Sync Status

```bash
ayo sync status
```

**Expected:** Shows changes since last sync.

### 14.9 Remote Sync

```bash
# Add remote
ayo sync remote add git@github.com:user/ayo-sync.git

# Show remote
ayo sync remote show

# Push changes
ayo sync push -m "Sync from macbook"

# Pull changes
ayo sync pull
```

**Expected:** Sandbox state synced via git.

---

## 15. Matrix Communication

### 15.1 Check Status

```bash
ayo matrix status
```

**Expected:** Shows Matrix connection state.

### 15.2 List Rooms

```bash
ayo matrix rooms
ayo matrix rooms --session <session-id>
```

**Expected:** Shows Matrix rooms.

### 15.3 Create Room

```bash
ayo matrix create -n "Project Discussion"
ayo matrix create -n "Code Review" --invite @code,@ayo
```

**Expected:** Room created with ID.

### 15.4 Send Message

```bash
ayo matrix send <room-id> "Hello agents!"
ayo matrix send <room-id> -f message.md --markdown
ayo matrix send <room-id> "From @code" --as @code
```

**Expected:** Message sent to room.

### 15.5 Read Messages

```bash
ayo matrix read <room-id>
ayo matrix read <room-id> -n 50
ayo matrix read <room-id> -f  # Follow/stream
```

**Expected:** Shows room messages.

### 15.6 Room Members

```bash
ayo matrix who <room-id>
ayo matrix invite <room-id> @code
```

**Expected:** Shows/modifies room membership.

---

## 16. HTTP Server

### 16.1 Start Server

```bash
# Default (localhost:random port)
ayo serve

# Specific port
ayo serve -p 8080

# Bind to all interfaces
ayo serve --host 0.0.0.0 -p 8080

# With HTTPS tunnel
ayo serve -t
```

**Expected:** Server starts, shows URL.

### 16.2 Test API (in another terminal)

```bash
# Health check
curl http://localhost:8080/health

# Run agent
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"agent": "@ayo", "message": "Hello"}'
```

**Expected:** API responds correctly.

---

## 17. Global Flags

Test these flags work across commands:

```bash
# JSON output
ayo agents list --json
ayo sessions list --json
ayo sandbox list --json

# Quiet mode
ayo agents list -q
ayo @ayo "test" -q

# Config file
ayo --config ~/.config/ayo/custom.json agents list

# Debug mode
ayo @ayo "test" --debug
```

**Expected:** Flags modify output appropriately.

---

## 18. Error Handling

### 18.1 Invalid Agent

```bash
ayo @nonexistent
```

**Expected:** Clear error message about agent not found.

### 18.2 Service Not Running

```bash
ayo sandbox service stop
ayo sandbox list
```

**Expected:** Error indicating service needs to be started.

### 18.3 Invalid Share Path

```bash
ayo share /nonexistent/path
```

**Expected:** Error about path not existing.

### 18.4 Invalid Trigger Schedule

```bash
ayo triggers schedule @ayo "invalid cron" -p "test"
```

**Expected:** Error about invalid cron expression.

---

## 19. Cleanup

After testing, clean up test artifacts:

```bash
# Remove test agents
ayo agents rm @test-agent --force 2>/dev/null
ayo agents rm @test-model --force 2>/dev/null
ayo agents rm @test-prompt --force 2>/dev/null
ayo agents rm @test-tools --force 2>/dev/null

# Remove test shares
ayo share rm --all

# Remove test triggers
ayo triggers list  # Note IDs
ayo triggers rm <id> --force

# Prune sandboxes
ayo sandbox prune --force

# Stop service
ayo sandbox service stop
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check for existing process
pgrep -f 'ayo service'

# Remove stale socket
rm -f /tmp/ayo/daemon.sock
rm -f /tmp/ayo/daemon.pid

# Try foreground mode
ayo sandbox service start -f
```

### Sandbox Commands Fail

```bash
# Check service status
ayo sandbox service status

# Check sandbox exists
ayo sandbox list

# Check container directly (macOS)
container list
```

### Share Not Visible in Sandbox

```bash
# Verify symlink exists
ls -la ~/.local/share/ayo/sandbox/workspace/

# Check sandbox mounts
ayo sandbox show --id <id>

# Restart sandbox
ayo sandbox stop --id <id>
ayo sandbox start --id <id>
```

### Memory Search Returns Nothing

```bash
# Check embedding service
ayo doctor -v

# Reindex memories
ayo memory reindex
```
