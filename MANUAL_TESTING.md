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
| `ayo squad create <name>` | Create squad (generates SQUAD.md) |
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

### 3. Squad Creation with SQUAD.md

```bash
# Create squad with agents
ayo squad create alpha \
  -d "Feature team" \
  -a @architect,@backend,@qa \
  -o /tmp/alpha-output

# Verify SQUAD.md was created
cat ~/.local/share/ayo/sandboxes/squads/alpha/SQUAD.md
# Shows template with mission, context, roles for each agent

# Edit the constitution to define team purpose
cat > ~/.local/share/ayo/sandboxes/squads/alpha/SQUAD.md << 'EOF'
# Squad: alpha

## Mission

Build a user authentication system with login, registration, and password reset.

## Context

- Using Express.js with TypeScript
- PostgreSQL database with Prisma ORM
- Must follow OWASP security guidelines

## Agents

### @architect
**Role**: System design
**Responsibilities**:
- Define API schema
- Make technology decisions
- Review implementation

### @backend
**Role**: Implementation
**Responsibilities**:
- Implement API endpoints
- Write database migrations
- Handle authentication logic

### @qa
**Role**: Quality assurance
**Responsibilities**:
- Review code changes
- Write test cases
- Verify security requirements

## Coordination

1. @architect designs schema first
2. @backend implements based on design
3. @qa reviews after implementation
4. Use ticket dependencies to enforce order

## Guidelines

- All endpoints must have tests
- Follow conventional commit messages
- Security changes need @qa sign-off
EOF

# Verify constitution
ayo squad show alpha
```

### 4. Squad with Tickets and Constitution

```bash
# Create a squad
ayo squad create auth-team -a @backend,@qa

# Edit SQUAD.md
$EDITOR ~/.local/share/ayo/sandboxes/squads/auth-team/SQUAD.md

# Create tickets in the squad
ayo squad ticket auth-team create "Implement login endpoint" -a @backend
ayo squad ticket auth-team create "Test login endpoint" -a @qa --depends-on <login-id>

# Verify constitution is in context
# When agents start, they receive SQUAD.md content in their system prompt
ayo squad start auth-team
ayo squad agent auth-team start @backend
# @backend now knows: mission, its role, coordination rules
```

### 5. File Sharing

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

### 6. Memory System

```bash
# Store facts
ayo memory store "This project uses PostgreSQL"
ayo memory store "API runs on port 8080"

# Search
ayo memory search "database"

# List all
ayo memory list
```

### 7. Triggers

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

# Remove test squads
rm -rf ~/.local/share/ayo/sandboxes/squads/alpha
rm -rf ~/.local/share/ayo/sandboxes/squads/auth-team
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

## Level 5: Squad with Constitution

```bash
# Create squad with SQUAD.md constitution
ayo squad create feature -a @dev,@qa -o /tmp/output

# Squad structure:
# ~/.local/share/ayo/sandboxes/squads/feature/
# ├── SQUAD.md        ← Team constitution (mission, roles, coordination)
# ├── .tickets/       ← Coordination tickets
# ├── .context/       ← Persistent state
# ├── workspace/      ← Shared code
# └── agent-homes/    ← Per-agent directories

# Edit constitution to define team
$EDITOR ~/.local/share/ayo/sandboxes/squads/feature/SQUAD.md

# When agents start, they receive SQUAD.md in their system prompt
# Each agent knows:
# - The team's mission
# - Their specific role and responsibilities
# - How to coordinate with other agents
# - Team-specific guidelines
```

**Key Concept:** SQUAD.md is the team's constitution. All agents in a squad receive it in their context, ensuring shared understanding of mission, roles, and coordination rules.

## Level 6: Full Orchestration

```bash
# @ayo receives complex task
ayo @ayo "Build user auth with tests and docs"

# @ayo internally:
# 1. Creates squad with SQUAD.md defining auth team
# 2. Creates tickets with dependencies
# 3. Assigns to specialist agents
# 4. Each agent starts with constitution context
# 5. Monitors progress via tickets
# 6. Syncs output when done
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
├───────┤    ├───────┤    ├───────┤
│SQUAD.md│   │SQUAD.md│   │SQUAD.md│
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
- **Constitution**: SQUAD.md defines mission, roles, coordination for each squad
- **Isolation**: Squads can't see each other
- **Coordination**: Tickets within squad, not cross-squad
- **Context**: All agents in squad share constitution
- **Visibility**: @ayo sees everything
- **Persistence**: State survives restarts
- **Automation**: Triggers for events/schedules

---

## SQUAD.md Reference

The `SQUAD.md` file is the team constitution. Location:
```
~/.local/share/ayo/sandboxes/squads/{name}/SQUAD.md
```

### Template Structure

```markdown
# Squad: {name}

## Mission
{What the team is trying to accomplish}

## Context
{Technical constraints, dependencies, deadlines}

## Agents
### @agent-handle
**Role**: {brief role}
**Responsibilities**:
- {responsibility 1}
- {responsibility 2}

## Coordination
{How agents hand off work, dependency rules}

## Guidelines
{Team-specific rules, coding style, review process}
```

### How It Works

1. **Creation**: `ayo squad create` generates a template SQUAD.md
2. **Editing**: User edits SQUAD.md to define team purpose
3. **Injection**: When agents start, SQUAD.md is injected into system prompt
4. **Shared Context**: All agents in squad see same constitution
5. **Updates**: Edit SQUAD.md, restart agents to pick up changes
