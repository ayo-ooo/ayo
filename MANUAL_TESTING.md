# Manual Testing Guide

Complete testing guide for the ayo agent orchestration system.

## Prerequisites

- macOS 26+ with Apple Silicon
- Go 1.21+
- Build: `go build -o ayo ./cmd/ayo/...`

---

## Quick Start

```bash
# 1. Build and setup
go build -o ayo ./cmd/ayo/...
./ayo setup

# 2. Start daemon
./ayo sandbox service start

# 3. Basic test
./ayo "Hello, what can you do?"
```

---

## Core Commands

### System

| Command | Description |
|---------|-------------|
| `ayo doctor` | Health check |
| `ayo setup` | Initial configuration |
| `ayo sandbox service start` | Start daemon |
| `ayo sandbox service stop` | Stop daemon |
| `ayo sandbox service status` | Check daemon status |

### Agents

| Command | Description |
|---------|-------------|
| `ayo agents list` | List all agents |
| `ayo agents show @ayo` | Show agent details |
| `ayo agents create @test -d "Test"` | Create agent |
| `ayo agents rm @test` | Remove agent |

### Chat

| Command | Description |
|---------|-------------|
| `ayo` | Interactive chat with @ayo |
| `ayo @agent` | Interactive chat with specific agent |
| `ayo "prompt"` | Single prompt (must be quoted) |
| `ayo -c` | Continue last session |
| `ayo -a file.txt "analyze"` | Attach file |

### Tickets

| Command | Description |
|---------|-------------|
| `ayo ticket create "title"` | Create ticket |
| `ayo ticket list` | List tickets |
| `ayo ticket show <id>` | Show ticket |
| `ayo ticket start <id>` | Mark in-progress |
| `ayo ticket close <id>` | Mark closed |
| `ayo ticket assign <id> @agent` | Assign to agent |
| `ayo ticket ready` | List ready tickets |
| `ayo ticket blocked` | List blocked tickets |

### Squads

| Command | Description |
|---------|-------------|
| `ayo squad create <name>` | Create squad |
| `ayo squad list` | List squads |
| `ayo squad show <name>` | Show squad details |
| `ayo squad add-agent <squad> @agent` | Add agent |
| `ayo squad remove-agent <squad> @agent` | Remove agent |
| `ayo squad start <name>` | Start sandbox |
| `ayo squad stop <name>` | Stop sandbox |
| `ayo squad destroy <name>` | Destroy squad |

### Sandbox

| Command | Description |
|---------|-------------|
| `ayo sandbox list` | List sandboxes |
| `ayo sandbox exec --id <id> <cmd>` | Execute command |
| `ayo share <path>` | Share directory |
| `ayo share list` | List shares |

### Memory

| Command | Description |
|---------|-------------|
| `ayo memory store "fact"` | Store memory |
| `ayo memory search "query"` | Search memories |
| `ayo memory list` | List memories |

### Triggers

| Command | Description |
|---------|-------------|
| `ayo triggers list` | List triggers |
| `ayo triggers schedule @agent "cron" -p "prompt"` | Schedule trigger |
| `ayo triggers watch <path> @agent -p "prompt"` | Watch trigger |
| `ayo triggers rm <id>` | Remove trigger |

---

## Test Scenarios

### 1. Basic Agent Interaction

```bash
# Start daemon
ayo sandbox service start

# Interactive chat
ayo @ayo
> What time is it?
> /exit

# Single prompt
ayo "What is 2+2?"

# With file attachment
echo "Hello World" > /tmp/test.txt
ayo "What does this file contain?" -a /tmp/test.txt
```

### 2. Ticket Workflow

```bash
# Create tickets with dependencies
ayo ticket create "Design API schema" --assignee @architect --priority high
# Note the ID, e.g., proj-a1b2

ayo ticket create "Implement endpoints" --assignee @backend --depends-on proj-a1b2

# Check what's ready
ayo ticket ready --assignee @architect
# Shows: "Design API schema"

ayo ticket ready --assignee @backend
# Shows: nothing (blocked)

# Complete first ticket
ayo ticket start proj-a1b2
ayo ticket close proj-a1b2

# Now backend can work
ayo ticket ready --assignee @backend
# Shows: "Implement endpoints"
```

### 3. Squad Collaboration

```bash
# Create squad with output directory
ayo squad create alpha \
  -d "Feature team" \
  -a @architect,@backend \
  -o /tmp/alpha-output

# Check squad
ayo squad show alpha

# Add another agent
ayo squad add-agent alpha @qa

# List squads
ayo squad list

# Cleanup
ayo squad stop alpha
ayo squad destroy alpha --delete-data
```

### 4. File Sharing

```bash
# Share a directory
mkdir -p /tmp/testproject
echo "# Test" > /tmp/testproject/README.md
ayo share /tmp/testproject --as project

# Verify share
ayo share list

# Use in sandbox
ayo sandbox list
# Get sandbox ID
ayo sandbox exec --id <id> cat /workspace/project/README.md

# Remove share
ayo share rm project
```

### 5. Memory System

```bash
# Store facts
ayo memory store "This project uses PostgreSQL"
ayo memory store "API runs on port 8080"

# Search
ayo memory search "database"

# List all
ayo memory list
```

### 6. Triggers

```bash
# Create watch trigger
mkdir -p /tmp/watched
ayo triggers watch /tmp/watched @ayo -p "Files changed"

# List triggers
ayo triggers list

# Test by creating file
touch /tmp/watched/test.txt

# Cleanup
ayo triggers rm <id>
```

---

## Error Cases

### Unknown Command Detection

```bash
# Single lowercase word = error
ayo foobar
# Error: unknown command "foobar"

# Quoted = works as prompt
ayo "foobar"
# Agent responds

# Uppercase = works as prompt  
ayo Hello
# Agent responds
```

### Service Not Running

```bash
ayo sandbox service stop
ayo sandbox list
# Error: cannot connect to daemon
```

### Invalid Agent

```bash
ayo @nonexistent
# Error: agent not found
```

---

## Cleanup

```bash
# Stop daemon
ayo sandbox service stop

# Remove test data
rm -rf /tmp/testproject /tmp/watched /tmp/alpha-output /tmp/test.txt
```

---

## Troubleshooting

### Daemon Won't Start

```bash
rm -f ~/.local/share/ayo/daemon.sock
rm -f ~/.local/share/ayo/daemon.pid
ayo sandbox service start -f  # Foreground for debugging
```

### "Method Not Found" Errors

Restart daemon to pick up code changes:
```bash
ayo sandbox service stop
ayo sandbox service start
```

### Container Issues

```bash
# Check container runtime
container list

# Check ayo doctor
ayo doctor -v
```

---

# Appendix: Architecture Demonstration

Progressive examples showing system capabilities.

## Level 1: Single Agent

```bash
ayo "What is the capital of France?"
```

**Flow:** User → CLI → Agent → LLM → Response

## Level 2: Agent with Files

```bash
ayo share ~/Code/project --as proj
ayo "Review /workspace/proj/main.go"
```

**Flow:** User → Agent → Sandbox → Filesystem → LLM

## Level 3: Persistent Memory

```bash
ayo memory store "Project uses React 18"
ayo "How should I add components?"
# Agent recalls stored context
```

**Flow:** User → Agent → Memory Service → Vector Search → LLM

## Level 4: Ticket Coordination

```bash
# Create dependent tickets
ayo ticket create "Design schema" --assignee @architect
ayo ticket create "Implement API" --assignee @backend --depends-on <id>

# Agents check ready queue
ayo ticket ready --assignee @backend
# Empty until schema done
```

**Flow:** Tickets → Dependency Graph → Agent Work Queue

## Level 5: Squad Isolation

```bash
# Create isolated team sandbox
ayo squad create feature -a @dev,@qa -o /tmp/output

# Agents share workspace but are isolated from other squads
# Coordinate via tickets in squad's .tickets/ directory
```

**Flow:** 
```
Squad Sandbox
├── .tickets/     (coordination)
├── .context/     (state)
├── /workspace/   (shared files)
└── /agent-homes/ (per-agent)
```

## Level 6: Full Orchestration

```bash
# @ayo receives complex task
ayo @ayo "Build user auth with tests and docs"

# @ayo internally:
# 1. Creates squad
# 2. Creates tickets with dependencies
# 3. Assigns to specialist agents
# 4. Monitors progress
# 5. Syncs output when done
```

**Architecture:**
```
┌─────────────────────────────────────┐
│            @ayo Orchestrator        │
│         (sees all squads)           │
└─────────────────┬───────────────────┘
                  │
    ┌─────────────┼─────────────┐
    ▼             ▼             ▼
┌───────┐    ┌───────┐    ┌───────┐
│Squad A│    │Squad B│    │Squad C│
│tickets│    │tickets│    │tickets│
│agents │    │agents │    │agents │
└───────┘    └───────┘    └───────┘
    │             │             │
    └─────────────┼─────────────┘
                  ▼
┌─────────────────────────────────────┐
│           Daemon Services           │
│  Sandbox Pool │ Triggers │ Memory  │
└─────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────┐
│      Apple Container Runtime        │
└─────────────────────────────────────┘
```

**Key Properties:**
- **Isolation**: Squads can't see each other
- **Coordination**: Tickets, not messages
- **Visibility**: @ayo sees everything
- **Persistence**: State survives restarts
- **Automation**: Triggers for events/schedules
