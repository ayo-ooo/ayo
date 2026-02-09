# Ayo Tutorial

Welcome to ayo! This guide walks you through everything you need to know to work effectively with AI agents in sandboxed environments.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic Agent Usage](#basic-agent-usage)
3. [Sandbox Execution](#sandbox-execution)
4. [File System Access](#file-system-access)
5. [Creating Custom Agents](#creating-custom-agents)
6. [Flows and Automation](#flows-and-automation)
7. [Inter-Agent Communication](#inter-agent-communication)
8. [Triggers and Scheduling](#triggers-and-scheduling)
9. [Advanced Topics](#advanced-topics)

---

## Getting Started

### Installation

```bash
# Install from source
go install github.com/alexcabrera/ayo/cmd/ayo@latest

# Or build from repository
git clone https://github.com/alexcabrera/ayo
cd ayo
go build -o ayo ./cmd/ayo/...
```

### Initial Setup

```bash
# Run setup to install built-in agents and skills
ayo setup

# Verify installation
ayo doctor
```

### Configure API Keys

Set at least one provider API key:

```bash
export OPENAI_API_KEY="sk-..."
# or
export ANTHROPIC_API_KEY="sk-ant-..."
# or
export GOOGLE_API_KEY="..."
```

---

## Basic Agent Usage

### Interactive Chat

Start an interactive chat session:

```bash
# Chat with the default agent (@ayo)
ayo

# Chat with a specific agent
ayo @ayo
```

### Single Prompts

For one-off tasks:

```bash
# Simple question
ayo "What is the capital of France?"

# Task with file attachment
ayo -a main.go "Review this code for bugs"

# Multiple attachments
ayo -a file1.txt -a file2.txt "Compare these files"
```

### Sessions

Resume previous conversations:

```bash
# List recent sessions
ayo sessions list

# Continue the most recent session
ayo sessions continue --latest

# Continue a specific session
ayo sessions continue <session-id>
```

---

## Sandbox Execution

Ayo runs agent commands inside isolated containers for security and reproducibility.

### Understanding Sandboxes

- **Isolation**: Agents can't access your host filesystem directly
- **Safety**: Malicious or buggy commands can't damage your system
- **Reproducibility**: Same environment every time

### Starting the Service

The sandbox service must be running for agent execution:

```bash
# Start the background service
ayo sandbox service start

# Check status
ayo sandbox service status

# Stop when done
ayo sandbox service stop
```

### Interacting with Sandboxes

```bash
# List active sandboxes
ayo sandbox list

# Execute a command in the sandbox
ayo sandbox exec ls -la

# Open an interactive shell
ayo sandbox login

# Check what OS the sandbox runs
ayo sandbox exec cat /etc/os-release
```

### Sandbox Lifecycle

When you run an agent command, ayo:

1. Allocates a sandbox from the warm pool
2. Creates an agent user inside the container
3. Executes commands as that user
4. Releases the sandbox when the session ends

---

## File System Access

### The Mount System

Since sandboxes are isolated, you must explicitly grant filesystem access:

```bash
# Grant read-write access to current directory
ayo mount add .

# Grant read-only access
ayo mount add ~/Documents --ro

# List all grants
ayo mount list

# Revoke access
ayo mount rm /path/to/directory
```

### How It Works

1. **Grant access** on the host with `ayo mount add`
2. **Path appears** inside sandbox at the same location
3. **Agent can read/write** based on grant mode

### Working Copy Model

For safety, ayo uses a working copy model:

1. Files are copied into the sandbox when accessed
2. Agent modifies the copy (not your originals)
3. You review and sync changes back:

```bash
# See what changed
ayo sandbox diff /sandbox/path /host/path

# Sync changes back to host
ayo sandbox sync /sandbox/path /host/path
```

---

## Creating Custom Agents

### Quick Creation

Ask @ayo to help create agents:

```bash
ayo "Help me create an agent for code review"
```

### Manual Creation

```bash
# Create with CLI
ayo agents create @reviewer \
  -m gpt-4 \
  -d "Reviews code for best practices" \
  -f ~/prompts/reviewer.md

# List your agents
ayo agents list

# Show agent details
ayo agents show @reviewer
```

### Agent Configuration

Agents live in `~/.config/ayo/agents/@handle/`:

```
~/.config/ayo/agents/@reviewer/
├── config.json     # Model, tools, skills
└── system.md       # System prompt
```

**config.json example:**

```json
{
  "model": "gpt-4",
  "description": "Code review specialist",
  "tools": ["bash", "agent_call"],
  "skills": ["coding", "debugging"]
}
```

---

## Flows and Automation

Flows are multi-step workflows that orchestrate agents and shell commands.

### Flow Types

1. **Shell Flows** (`.sh`): Bash scripts with JSON I/O
2. **YAML Flows** (`.yaml`): Declarative workflows with dependencies

### Creating a Shell Flow

```bash
# Create a new flow
ayo flows new my-flow

# Edit the generated file
code ~/.config/ayo/flows/my-flow.sh
```

**Example shell flow:**

```bash
#!/usr/bin/env bash
# ayo:flow
# name: git-summary
# description: Summarize recent commits

set -euo pipefail

INPUT="${1:-$(cat)}"
DAYS=$(echo "$INPUT" | jq -r '.days // 7')

# Get git log
GIT_LOG=$(git log --oneline --since="${DAYS} days ago")

# Have agent summarize
echo "{\"commits\": \"$GIT_LOG\"}" | ayo @ayo "
  Summarize these commits as JSON with:
  - summary: string
  - highlights: array
"
```

### Creating a YAML Flow

```yaml
# ~/.config/ayo/flows/daily-report.yaml
version: 1
name: daily-report
description: Generate daily status report

steps:
  - id: git-changes
    type: shell
    run: git log --oneline --since="1 day ago"

  - id: test-status
    type: shell
    run: go test ./... -json 2>&1 | tail -10

  - id: generate-report
    type: agent
    agent: "@ayo"
    prompt: |
      Create a daily report from:
      
      Commits: {{ steps.git-changes.stdout }}
      Tests: {{ steps.test-status.stdout }}
    depends_on: [git-changes, test-status]
```

### Running Flows

```bash
# Run with inline input
ayo flows run my-flow '{"days": 7}'

# Run with file input
ayo flows run my-flow -i input.json

# Run YAML flow with parameters
ayo flows run daily-report --param days=3

# View run history
ayo flows history
```

---

## Inter-Agent Communication

Agents can communicate via Matrix, a secure messaging protocol.

### Matrix Basics

```bash
# Check Matrix status
ayo matrix status

# List available rooms
ayo matrix rooms

# Create a room
ayo matrix create project-chat

# Send a message
ayo matrix send project-chat "Build completed!"

# Read messages
ayo matrix read project-chat 10
```

### Multi-Agent Collaboration

1. Create a shared room
2. Invite agents to the room
3. Agents can send/receive messages

```bash
# Create team room
ayo matrix create code-review-team

# Invite agents
ayo matrix invite code-review-team @reviewer
ayo matrix invite code-review-team @fixer
```

### Matrix in Flows

Agents inside sandboxes can use Matrix:

```yaml
steps:
  - id: notify
    type: shell
    run: ayo matrix send alerts "Build started"

  - id: build
    type: shell
    run: go build ./...

  - id: complete
    type: shell
    run: ayo matrix send alerts "Build completed: {{ steps.build.exit_code }}"
    depends_on: [build]
```

---

## Triggers and Scheduling

Automate agent execution with triggers.

### Cron Triggers

Run on a schedule:

```bash
# Every hour
ayo trigger schedule @backup "every hour"

# Daily at 9am
ayo trigger schedule @reports "every day at 9am" --prompt "Generate daily report"

# Using cron syntax
ayo trigger schedule @cleanup "0 0 3 * * *"  # 3am daily
```

### Watch Triggers

Run when files change:

```bash
# Watch a directory
ayo trigger watch ./src @build

# Watch specific patterns
ayo trigger watch ./src @build "*.go" "*.mod"

# Recursive with event filter
ayo trigger watch ./docs @docs "*.md" --recursive --events modify,create
```

### Managing Triggers

```bash
# List all triggers
ayo trigger list

# Show trigger details
ayo trigger show <id>

# Test a trigger manually
ayo trigger test <id>

# Disable/enable
ayo trigger disable <id>
ayo trigger enable <id>

# Remove
ayo trigger rm <id>
```

### Triggers in YAML Flows

Define triggers directly in flow files:

```yaml
version: 1
name: auto-build
description: Build on code changes

triggers:
  - type: watch
    path: ./src
    patterns: ["*.go"]
  
  - type: cron
    schedule: "0 6 * * *"  # 6am daily

steps:
  - id: build
    type: shell
    run: go build ./...
```

---

## Advanced Topics

### Agent Skills

Skills are reusable instruction sets:

```bash
# List available skills
ayo skills list

# Create a skill
ayo skills create my-skill

# Add to agent
ayo agents create @helper --skills my-skill,debugging
```

### Memory System

Agents remember facts across sessions:

```bash
# Store a memory
ayo memory store "User prefers TypeScript over JavaScript"

# Search memories
ayo memory search "programming preferences"

# List all memories
ayo memory list

# View statistics
ayo memory stats
```

### Agent Chaining

Pipe output between agents:

```bash
# Chain agents
ayo @analyzer '{"code":"..."}' | ayo @formatter

# Validate chain compatibility
ayo chain validate @agent '{"field": "value"}'

# See chainable agents
ayo chain ls
```

### Backup and Sync

```bash
# Create backup
ayo backup create --name "before-experiment"

# List backups
ayo backup list

# Restore
ayo backup restore "before-experiment"

# Export for transfer
ayo backup export "before-experiment" ~/backup.tar.gz
```

### Debugging

```bash
# Enable debug output
ayo --debug @ayo "Run a command"

# Check system health
ayo doctor -v

# Debug scripts
./debug/sandbox-status.sh --verbose
./debug/daemon-status.sh --logs
./debug/mount-check.sh
```

---

## Next Steps

- **[TUTORIAL-daily-digest.md](TUTORIAL-daily-digest.md)** - Build an automated daily report flow
- **[TUTORIAL-code-review.md](TUTORIAL-code-review.md)** - Create a code review pipeline
- **[TUTORIAL-file-watcher.md](TUTORIAL-file-watcher.md)** - Set up file watch triggers

For complete command reference, see [docs/cli-reference.md](docs/cli-reference.md).
