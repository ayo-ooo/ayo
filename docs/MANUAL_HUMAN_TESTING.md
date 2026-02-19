# Manual Human Testing Guide

> **📍 RESUME POINT (2026-02-18):** Resume testing from **Section 2.1 (ayo doctor)**.
> Previous issues found in Phase 3 have been fixed:
> - ✅ `ayo agents list` now shows description column with better formatting
> - ✅ First @ayo invocation is now fast (sandbox pre-warmed on daemon start)
> - ✅ @ayo now understands it's in a sandbox with limited filesystem access
> - ✅ New `request_access` tool allows @ayo to request host file mounts from user

A comprehensive, step-by-step guide for testing the entire ayo system from a clean install. This document proves that all components work together as designed.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Phase 1: Clean Installation](#phase-1-clean-installation)
3. [Phase 2: System Health Verification](#phase-2-system-health-verification)
4. [Phase 3: Basic Agent Operations](#phase-3-basic-agent-operations)
5. [Phase 4: Skills System](#phase-4-skills-system)
6. [Phase 5: Memory System](#phase-5-memory-system)
7. [Phase 6: Sessions and Continuity](#phase-6-sessions-and-continuity)
8. [Phase 7: Sandbox Isolation](#phase-7-sandbox-isolation)
9. [Phase 8: Planner System](#phase-8-planner-system)
10. [Phase 9: Ticket Coordination](#phase-9-ticket-coordination)
11. [Phase 10: Entity Index and Semantic Search](#phase-10-entity-index-and-semantic-search)
12. [Phase 11: Squad Creation and Management](#phase-11-squad-creation-and-management)
13. [Phase 12: Squad I/O Schemas](#phase-12-squad-io-schemas)
14. [Phase 13: Squad Dispatch and Routing](#phase-13-squad-dispatch-and-routing)
15. [Phase 14: Flows](#phase-14-flows)
16. [Phase 15: Flow-Squad Integration](#phase-15-flow-squad-integration)
17. [Phase 16: Triggers](#phase-16-triggers)
18. [Phase 17: Multi-Agent Coordination](#phase-17-multi-agent-coordination)
19. [Phase 18: End-to-End Workflow](#phase-18-end-to-end-workflow)
20. [Debugging Reference](#debugging-reference)

---

## Prerequisites

Before starting, ensure you have:

- **macOS 26+** with Apple Silicon (for Apple Container sandbox)
- **Go 1.22+** installed
- **Git** installed
- **Ollama** installed (or willingness to install during setup)
- A valid LLM API key configured (Anthropic, OpenAI, or Vertex AI)

### Pre-test Environment Check

```bash
# Check Go version
go version
# Expected: go version go1.22.x darwin/arm64 (or higher)

# Check Git
git --version
# Expected: git version 2.x.x

# Check Ollama (optional but recommended)
ollama --version
# Expected: ollama version x.x.x

# Verify you're in the ayo repository
cd /path/to/ayo
git status
# Expected: On branch chain-flows-squads-alignment
```

---

## Phase 1: Clean Installation

This phase completely removes any existing ayo installation and starts fresh.

### 1.1 Run Clean Install

```bash
# From the ayo repository root
./install.sh --clean

# If prompted, confirm removal of existing data
# Watch for:
# - "Building ayo..."
# - "Running ayo setup..."
# - "Pulling Ollama models..."
# - "Installation complete"
```

**Expected Output:**
```
🗑️  Cleaning existing installation...
🔨 Building ayo...
✅ Build successful
📦 Running ayo setup...
  Installing built-in agents...
  Installing built-in skills...
✅ Setup complete
🤖 Checking Ollama...
  Pulling ministral-3:3b...
  Pulling nomic-embed-text...
✅ Models ready
🎉 Installation complete!
```

### 1.2 Verify Installation

```bash
# Check ayo is accessible
ayo --version
# Expected: ayo version x.x.x (commit: abc1234)

# Verify PATH (if using dev install)
which ayo
# Expected: /path/to/ayo/.local/bin/ayo (dev) or ~/go/bin/ayo (prod)
```

### 1.3 Verify Directory Structure

```bash
# Check config directory was created
ls -la ~/.config/ayo/
# Expected: ayo.json, possibly squads/ directory

# Check data directory was created
ls -la ~/.local/share/ayo/
# Expected: agents/, skills/, db/, sandboxes/, logs/

# Check built-in agents exist
ls ~/.local/share/ayo/agents/
# Expected: @ayo/ and potentially other built-in agents
```

---

## Phase 2: System Health Verification

### 2.1 Run Doctor

```bash
ayo doctor
# Expected: All checks pass
```

**Expected Output:**
```
Checking ayo installation...
  ✓ Configuration file exists
  ✓ Data directory exists
  ✓ Database accessible
  ✓ Built-in agents installed
  ✓ Built-in skills installed

Checking dependencies...
  ✓ Ollama running
  ✓ Embedding model available (nomic-embed-text)
  ✓ Local inference model available (ministral-3:3b)

Checking sandbox provider...
  ✓ Apple Container available
  ✓ Sandbox service responsive

All checks passed!
```

### 2.2 Verbose Doctor Check

```bash
ayo doctor -v
# Shows detailed information including model list
```

### 2.3 Start Sandbox Service

```bash
# Start the daemon/sandbox service
ayo sandbox service start

# Verify it's running
ayo sandbox service status
# Expected: Service running, PID: xxxxx

# Alternative: use debug script
./debug/daemon-status.sh
```

**Expected Output from daemon-status.sh:**
```
═══════════════════════════════════════════════════════════════════════════════
  DAEMON STATUS
═══════════════════════════════════════════════════════════════════════════════

───────────────────────────────────────────────────────────────────────────────
  Daemon Process
───────────────────────────────────────────────────────────────────────────────
  Running:      true
  PID:          12345
  ...

───────────────────────────────────────────────────────────────────────────────
  Socket Status
───────────────────────────────────────────────────────────────────────────────
  Path:         /var/folders/.../ayo/daemon.sock
  Exists:       true
```

### 2.4 Collect Full Diagnostics (for reference)

```bash
# Generate a full diagnostic report
./debug/collect-all.sh --output /tmp/ayo-baseline-diagnostics.txt

# Review it
cat /tmp/ayo-baseline-diagnostics.txt
```

---

## Phase 3: Basic Agent Operations

### 3.1 List Agents

```bash
ayo agents list
# Expected: List showing @ayo and any other built-in agents
```

**Expected Output:**
```
HANDLE     TRUST       TYPE      DESCRIPTION
@ayo       sandboxed   builtin   The default ayo agent
```

### 3.2 Show Agent Details

```bash
ayo agents show @ayo
# Expected: Full agent configuration
```

### 3.3 Basic Conversation

```bash
# Simple prompt (should work without any special setup)
ayo "What is 2 + 2?"
# Expected: "4" or similar simple response

# With debug output
AYO_DEBUG=1 ayo "What is 2 + 2?"
# Expected: Shows raw API calls and tool usage
```

### 3.4 Verify Sandbox Execution

```bash
# Ask agent to run a command (proves sandbox works)
ayo "List the files in your home directory"
# Expected: Agent uses bash tool, shows files from sandbox

# Verify with debug
AYO_DEBUG=1 ayo "Run 'pwd' and tell me the result"
# Expected: Shows bash tool call, result from sandbox container
```

### 3.5 Inspect Sandbox After Use

```bash
# List active sandboxes
ayo sandbox list
# Expected: Shows sandbox created for @ayo

# Show sandbox details
ayo sandbox show
# Expected: Shows working directory, mounts, status
```

---

## Phase 4: Skills System

### 4.1 List Available Skills

```bash
ayo skill list
# Expected: Shows built-in skills
```

**Expected Output:**
```
NAME               SOURCE    DESCRIPTION
coding             builtin   Software development best practices
memory             builtin   How to use the memory system
memory-usage       builtin   Guidelines for memory formation
sandbox            builtin   Working in sandboxed environments
ayo/orchestration  builtin   Agent orchestration and delegation
flows              builtin   Flow creation and management
...
```

### 4.2 Show Skill Details

```bash
ayo skill show coding
# Expected: Full skill content displayed
```

### 4.3 Use Agent with Specific Skill

```bash
# Attach a skill using + syntax
ayo @ayo +coding "Write a simple hello world in Python"
# Expected: Agent responds with Python code, following coding skill guidelines
```

### 4.4 Verify Skill Injection

```bash
# Debug mode shows skill injection
AYO_DEBUG=1 ayo @ayo +coding "Write hello world" 2>&1 | grep -i skill
# Expected: Shows skills being loaded into system prompt
```

---

## Phase 5: Memory System

### 5.1 Check Memory Stats

```bash
ayo memory stats
# Expected: Shows memory statistics (may be empty initially)
```

### 5.2 Store a Memory Manually

```bash
ayo memory store --agent @ayo --category preference "User prefers TypeScript over JavaScript"
# Expected: Memory stored, shows ID
```

### 5.3 List Memories

```bash
ayo memory list
# Expected: Shows the memory we just stored
```

### 5.4 Search Memories

```bash
ayo memory search "programming language preference"
# Expected: Finds the TypeScript preference memory
```

### 5.5 Test Memory Auto-Formation

```bash
# Have a conversation that should trigger memory formation
ayo "I always want you to use tabs instead of spaces for indentation"
# Expected: Agent acknowledges preference

# Check if memory was formed
ayo memory list --limit 5
# Expected: New memory about indentation preference
```

### 5.6 Test Memory Retrieval

```bash
# Start new session, ask about preferences
ayo "What do you know about my coding preferences?"
# Expected: Agent recalls TypeScript preference and tabs preference
```

---

## Phase 6: Sessions and Continuity

### 6.1 List Sessions

```bash
ayo session list
# Expected: Shows recent sessions from our tests
```

### 6.2 View Session Details

```bash
# Get a session ID from the list
ayo session show <session-id>
# Expected: Full conversation history
```

### 6.3 Continue a Session

```bash
# Continue the last session
ayo -c "What was I asking about before?"
# Expected: Agent has context from previous messages

# Or continue specific session
ayo -s <session-id> "Continue from where we left off"
```

### 6.4 Verify Session Persistence

```bash
# Check sessions are stored in database
ls -la ~/.local/share/ayo/db/
# Expected: SQLite database file(s)
```

---

## Phase 7: Sandbox Isolation

### 7.1 Verify File Isolation

```bash
# Create a test file on host
echo "SECRET_DATA" > /tmp/host-secret.txt

# Ask agent to read it (should fail - sandbox isolation)
ayo "Read the file /tmp/host-secret.txt and tell me its contents"
# Expected: Agent cannot access host filesystem directly
```

### 7.2 Share a Directory

```bash
# Share a directory with the sandbox
mkdir -p /tmp/ayo-test-share
echo "SHARED_DATA" > /tmp/ayo-test-share/test.txt

ayo share add /tmp/ayo-test-share --as testshare
# Expected: Share added successfully

# List shares
ayo share list
# Expected: Shows testshare pointing to /tmp/ayo-test-share
```

### 7.3 Access Shared Directory

```bash
# Now agent can access the shared directory
ayo "Read the file at /mnt/testshare/test.txt"
# Expected: Agent returns "SHARED_DATA"
```

### 7.4 Exec Directly in Sandbox

```bash
# Get sandbox ID
SANDBOX_ID=$(ayo sandbox list --json | jq -r '.[0].id')

# Execute command directly
ayo sandbox exec $SANDBOX_ID "whoami"
# Expected: Shows sandbox user (not your host user)

ayo sandbox exec $SANDBOX_ID "cat /etc/os-release"
# Expected: Shows container OS info
```

### 7.5 Interactive Shell (Manual Check)

```bash
# Open interactive shell in sandbox
ayo sandbox shell
# Expected: Drops into shell inside sandbox container
# Type 'exit' to leave
```

### 7.6 Clean Up Share

```bash
ayo share rm testshare
# Expected: Share removed

ayo share list
# Expected: testshare no longer listed
```

---

## Phase 8: Planner System

### 8.1 List Available Planners

```bash
ayo planner list
# Expected: Shows ayo-todos and ayo-tickets
```

**Expected Output:**
```
NAME          TYPE        DESCRIPTION
ayo-todos     near-term   Session-scoped task tracking
ayo-tickets   long-term   Persistent ticket-based coordination
```

### 8.2 Show Planner Details

```bash
ayo planner show ayo-todos
# Expected: Shows planner details including tools provided

ayo planner show ayo-tickets
# Expected: Shows ticket planner details
```

### 8.3 Test Near-Term Planner (Todos)

```bash
# Have agent create todos
ayo "I need to: 1) write tests, 2) update docs, 3) refactor code. Create a todo list for these."
# Expected: Agent uses todos tool to create task list
```

### 8.4 Verify Planner Tool Injection

```bash
# Debug shows planner tools being injected
AYO_DEBUG=1 ayo "Show me my current todo list" 2>&1 | grep -i "todos\|planner"
# Expected: Shows todos tool available and being used
```

---

## Phase 9: Ticket Coordination

### 9.1 Create a Ticket

```bash
ayo ticket create "Implement user authentication" --assignee @ayo
# Expected: Ticket created, shows ID
```

### 9.2 List Tickets

```bash
ayo ticket list
# Expected: Shows the authentication ticket
```

**Expected Output:**
```
ID          STATUS    TYPE    PRIORITY  ASSIGNEE  TITLE
abc-1234    open      task    2         @ayo      Implement user authentication
```

### 9.3 Show Ticket Details

```bash
ayo ticket show <ticket-id>
# Expected: Full ticket information
```

### 9.4 Update Ticket Status

```bash
# Start working on ticket
ayo ticket start <ticket-id>
# Expected: Status changes to in_progress

# Add a note
ayo ticket note <ticket-id> "Started working on JWT implementation"
# Expected: Note added

# Close the ticket
ayo ticket close <ticket-id>
# Expected: Status changes to closed

# Verify
ayo ticket show <ticket-id>
# Expected: Shows status: closed
```

### 9.5 Test Ticket Dependencies

```bash
# Create parent and child tickets
ayo ticket create "Build auth system" --assignee @ayo
# Note the parent ID

ayo ticket create "Create login endpoint" --assignee @ayo --deps <parent-id>
# Expected: Child ticket created with dependency
```

### 9.6 View Ready Tickets

```bash
ayo ticket ready
# Expected: Shows tickets with all dependencies satisfied
```

---

## Phase 10: Entity Index and Semantic Search

### 10.1 Check Index Status

```bash
ayo index status
# Expected: Shows indexed agents and squads count
```

**Expected Output:**
```
Entity Index Status
-------------------
Agents indexed: 1
Squads indexed: 0
Last rebuild: 2026-02-18T10:30:00Z
```

### 10.2 Rebuild Index

```bash
ayo index rebuild
# Expected: Index rebuilt, shows counts
```

### 10.3 Search the Index

```bash
# Search for agents by capability
ayo index search "code review and debugging"
# Expected: Shows ranked results with similarity scores
```

**Expected Output:**
```
RANK  SCORE   TYPE   HANDLE  DESCRIPTION
1     0.85    agent  @ayo    The default ayo agent
```

### 10.4 Create Agent with Description (for better search)

```bash
# Create a specialized agent
ayo agents create @reviewer \
  -m "claude-sonnet-4-20250514" \
  -d "Expert code reviewer specializing in security and performance" \
  -s "You are an expert code reviewer. Focus on security vulnerabilities, performance issues, and best practices."

# Expected: Agent created

# Rebuild index to include new agent
ayo index rebuild

# Search should now find both
ayo index search "security code review"
# Expected: @reviewer ranks higher than @ayo for this query
```

---

## Phase 11: Squad Creation and Management

### 11.1 Create a Basic Squad

```bash
ayo squad create dev-team --description "Development team for testing"
# Expected: Squad created
```

### 11.2 List Squads

```bash
ayo squad list
# Expected: Shows dev-team
```

**Expected Output:**
```
NAME       STATUS   AGENTS  EPHEMERAL  DESCRIPTION
dev-team   stopped  0       false      Development team for testing
```

### 11.3 Show Squad Details

```bash
ayo squad show dev-team
# Expected: Full squad information including paths
```

### 11.4 Inspect Squad Directory Structure

```bash
ls -la ~/.local/share/ayo/sandboxes/squads/dev-team/
# Expected: SQUAD.md, .tickets/, .context/, workspace/
```

### 11.5 View Default SQUAD.md

```bash
cat ~/.local/share/ayo/sandboxes/squads/dev-team/SQUAD.md
# Expected: Template SQUAD.md content
```

### 11.6 Edit SQUAD.md (Create Proper Constitution)

```bash
cat > ~/.local/share/ayo/sandboxes/squads/dev-team/SQUAD.md << 'EOF'
---
lead: "@ayo"
input_accepts: "@ayo"
planners:
  near_term: ayo-todos
  long_term: ayo-tickets
---
# Squad: dev-team

## Mission

Build and test software components with high quality and thorough testing.

## Context

- Using TypeScript and Node.js
- Focus on clean, maintainable code
- All code must have tests

## Agents

### @ayo
**Role**: Squad lead and coordinator
**Responsibilities**:
- Coordinate work between agents
- Review completed work
- Manage tickets

### @reviewer
**Role**: Code review specialist
**Responsibilities**:
- Review code for security and performance
- Suggest improvements
- Verify test coverage

## Coordination

1. @ayo receives tasks and creates tickets
2. Work is assigned to appropriate agents
3. @reviewer reviews completed work
4. @ayo closes tickets after review approval

## Guidelines

- Use conventional commit messages
- Write tests before implementation
- Document public APIs
EOF
```

### 11.7 Add Agents to Squad

```bash
ayo squad add-agent dev-team @ayo
# Expected: Agent added

ayo squad add-agent dev-team @reviewer
# Expected: Agent added

# Verify
ayo squad show dev-team
# Expected: Shows 2 agents
```

### 11.8 Start the Squad

```bash
ayo squad start dev-team
# Expected: Squad started, sandbox created
```

### 11.9 Verify Squad is Running

```bash
ayo squad list
# Expected: dev-team status: running

# Check sandbox is active
ayo sandbox list
# Expected: Shows sandbox for dev-team
```

### 11.10 Rebuild Index to Include Squad

```bash
ayo index rebuild

ayo index status
# Expected: Squads indexed: 1

# Search should find the squad
ayo index search "software development testing"
# Expected: #dev-team appears in results
```

---

## Phase 12: Squad I/O Schemas

### 12.1 Create Input Schema

```bash
cat > ~/.local/share/ayo/sandboxes/squads/dev-team/input.jsonschema << 'EOF'
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "task_type": {
      "type": "string",
      "enum": ["feature", "bugfix", "refactor", "test"],
      "description": "Type of task to perform"
    },
    "description": {
      "type": "string",
      "description": "Detailed description of what to do"
    },
    "priority": {
      "type": "string",
      "enum": ["low", "medium", "high"],
      "default": "medium"
    }
  },
  "required": ["task_type", "description"]
}
EOF
```

### 12.2 Create Output Schema

```bash
cat > ~/.local/share/ayo/sandboxes/squads/dev-team/output.jsonschema << 'EOF'
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "status": {
      "type": "string",
      "enum": ["success", "partial", "failed"]
    },
    "summary": {
      "type": "string",
      "description": "Summary of work completed"
    },
    "files_changed": {
      "type": "array",
      "items": { "type": "string" }
    },
    "notes": {
      "type": "string"
    }
  },
  "required": ["status", "summary"]
}
EOF
```

### 12.3 View Squad Schemas

```bash
ayo squad schema show dev-team
# Expected: Shows both input and output schemas
```

### 12.4 Validate Input Against Schema

```bash
# Create test input
echo '{"task_type": "feature", "description": "Add user login"}' > /tmp/test-input.json

ayo squad schema validate dev-team --input /tmp/test-input.json
# Expected: Validation passed

# Test invalid input
echo '{"description": "Missing task_type"}' > /tmp/bad-input.json

ayo squad schema validate dev-team --input /tmp/bad-input.json
# Expected: Validation failed - missing required field
```

---

## Phase 13: Squad Dispatch and Routing

### 13.1 Basic Squad Dispatch

```bash
# Dispatch to squad using # syntax
ayo "#dev-team" "Create a simple hello world function"
# Expected: Squad lead receives and processes the request
```

### 13.2 Dispatch with Debug

```bash
AYO_DEBUG=1 ayo "#dev-team" "What is your mission?"
# Expected: Shows routing to squad lead, SQUAD.md injection
```

### 13.3 Direct Agent in Squad

```bash
# Send to specific agent within squad context
ayo @reviewer "#dev-team" "Review the code in /workspace"
# Expected: @reviewer receives task with squad context
```

### 13.4 Create Squad Ticket via Dispatch

```bash
ayo "#dev-team" "Create a ticket for implementing user registration"
# Expected: Ticket created in squad's .tickets/ directory

# Verify ticket was created
ls ~/.local/share/ayo/sandboxes/squads/dev-team/.tickets/
# Expected: Shows ticket file

# Or use squad ticket command
ayo squad ticket list dev-team
# Expected: Shows the registration ticket
```

### 13.5 Monitor Squad Activity

```bash
# Watch squad logs (in separate terminal)
tail -f ~/.local/state/ayo/daemon.log | grep -i "dev-team"
```

---

## Phase 14: Flows

### 14.1 List Available Flows

```bash
ayo flow list
# Expected: May be empty or show built-in flows
```

### 14.2 Create a Simple Flow

```bash
mkdir -p ~/.config/ayo/flows

cat > ~/.config/ayo/flows/test-flow.yaml << 'EOF'
version: 1
name: test-flow
description: A simple test flow

steps:
  - id: greet
    type: shell
    run: echo "Hello from the flow!"

  - id: get-date
    type: shell
    run: date

  - id: summarize
    type: agent
    agent: "@ayo"
    prompt: "Given this date: {{ steps.get-date.stdout }}, tell me what day of the week it is."
EOF
```

### 14.3 Validate the Flow

```bash
ayo flow validate test-flow
# Expected: Flow is valid
```

### 14.4 Run the Flow

```bash
ayo flow run test-flow
# Expected: Executes all steps, shows output
```

**Expected Output:**
```
Running flow: test-flow

Step [greet] ✓
  stdout: Hello from the flow!

Step [get-date] ✓
  stdout: Tue Feb 18 10:30:00 PST 2026

Step [summarize] ✓
  output: Based on the date February 18, 2026, that is a Tuesday.

Flow completed successfully!
```

### 14.5 View Flow History

```bash
ayo flow history test-flow
# Expected: Shows the run we just did
```

### 14.6 Create Flow with Dependencies

```bash
cat > ~/.config/ayo/flows/dep-flow.yaml << 'EOF'
version: 1
name: dep-flow
description: Flow with step dependencies

steps:
  - id: step-a
    type: shell
    run: echo "A"

  - id: step-b
    type: shell
    run: echo "B"
    depends_on: [step-a]

  - id: step-c
    type: shell
    run: echo "C - received A={{ steps.step-a.stdout }}, B={{ steps.step-b.stdout }}"
    depends_on: [step-a, step-b]
EOF

ayo flow run dep-flow
# Expected: Steps run in correct order based on dependencies
```

---

## Phase 15: Flow-Squad Integration

### 15.1 Create Flow That Uses Squad

```bash
cat > ~/.config/ayo/flows/squad-flow.yaml << 'EOF'
version: 1
name: squad-flow
description: Flow that dispatches to a squad

steps:
  - id: prepare
    type: shell
    run: echo "Preparing work for squad..."

  - id: squad-work
    type: agent
    agent: "@ayo"
    squad: "#dev-team"
    prompt: "Check the squad workspace and report what files exist"
    depends_on: [prepare]

  - id: report
    type: agent
    agent: "@ayo"
    prompt: "Summarize this squad report: {{ steps.squad-work.output }}"
    depends_on: [squad-work]
EOF
```

### 15.2 Run Squad Flow

```bash
ayo flow run squad-flow
# Expected: Flow runs agent in squad context
```

### 15.3 Verify Squad Context Was Used

```bash
# Debug mode shows squad injection
AYO_DEBUG=1 ayo flow run squad-flow 2>&1 | grep -i "squad"
# Expected: Shows squad context being loaded
```

---

## Phase 16: Triggers

### 16.1 List Triggers

```bash
ayo trigger list
# Expected: Empty or shows existing triggers
```

### 16.2 Create a Schedule Trigger

```bash
# Create a trigger that runs every minute (for testing)
ayo trigger schedule @ayo "*/1 * * * *" --prompt "Log that you are alive"
# Expected: Trigger created, shows ID
```

### 16.3 List Triggers Again

```bash
ayo trigger list
# Expected: Shows the schedule trigger
```

### 16.4 Test Trigger Manually

```bash
# Get trigger ID from list
TRIGGER_ID=$(ayo trigger list --json | jq -r '.[0].id')

ayo trigger test $TRIGGER_ID
# Expected: Trigger fires, agent runs
```

### 16.5 Create File Watch Trigger

```bash
mkdir -p /tmp/ayo-watch-test

ayo trigger watch /tmp/ayo-watch-test @ayo "*.txt" --prompt "A file changed: {path}"
# Expected: Watch trigger created

# Test it
echo "test content" > /tmp/ayo-watch-test/trigger-test.txt
# Expected: Trigger fires (check daemon logs)

tail -20 ~/.local/state/ayo/daemon.log | grep -i trigger
```

### 16.6 Disable and Clean Up Triggers

```bash
# Disable triggers
ayo trigger list --json | jq -r '.[].id' | while read id; do
  ayo trigger disable $id
done

# Remove triggers
ayo trigger list --json | jq -r '.[].id' | while read id; do
  ayo trigger rm $id
done

ayo trigger list
# Expected: Empty
```

---

## Phase 17: Multi-Agent Coordination

### 17.1 Create Second Specialized Agent

```bash
ayo agents create @coder \
  -m "claude-sonnet-4-20250514" \
  -d "Expert programmer who writes clean, tested code" \
  -s "You are an expert programmer. Write clean, well-documented code with tests. Follow best practices."

# Expected: Agent created
```

### 17.2 Add to Squad

```bash
ayo squad add-agent dev-team @coder
# Expected: Agent added to squad

ayo squad show dev-team
# Expected: Shows 3 agents: @ayo, @reviewer, @coder
```

### 17.3 Update SQUAD.md with New Agent

```bash
# Add @coder to SQUAD.md
cat >> ~/.local/share/ayo/sandboxes/squads/dev-team/SQUAD.md << 'EOF'

### @coder
**Role**: Implementation specialist
**Responsibilities**:
- Write production code
- Create unit tests
- Document functions
EOF
```

### 17.4 Test Agent Delegation

```bash
# Ask squad lead to delegate work
ayo "#dev-team" "Create a utility function to validate email addresses. Have @coder write the code and @reviewer review it."
# Expected: Squad lead coordinates between agents
```

### 17.5 Verify Ticket-Based Coordination

```bash
# Check tickets created during coordination
ayo squad ticket list dev-team
# Expected: Shows tickets for the email validator task
```

---

## Phase 18: End-to-End Workflow

This phase combines everything into a realistic workflow.

### 18.1 Scenario Setup

We'll simulate a complete feature development cycle:
1. User requests a feature via squad dispatch
2. Squad lead creates tickets
3. Agents work on tickets
4. Work is reviewed
5. Output is synced

### 18.2 Dispatch Feature Request

```bash
# Send feature request to squad
ayo "#dev-team" "Build a password strength checker utility that:
1. Takes a password string as input
2. Returns a score from 0-100
3. Checks for length, complexity, and common patterns
4. Includes unit tests"
```

### 18.3 Monitor Ticket Creation

```bash
# In another terminal, watch tickets
watch -n 2 'ayo squad ticket list dev-team'

# Or check manually
ayo squad ticket list dev-team
# Expected: See tickets being created
```

### 18.4 Inspect Squad Workspace

```bash
# Check what files are being created
ls -la ~/.local/share/ayo/sandboxes/squads/dev-team/workspace/
# Expected: Code files being created

# Or exec into sandbox
ayo sandbox exec "$(ayo sandbox list --json | jq -r '.[0].id')" "ls -la /workspace"
```

### 18.5 Watch Daemon Logs

```bash
# Real-time log monitoring
tail -f ~/.local/state/ayo/daemon.log
# Expected: See agent activities, ticket updates
```

### 18.6 Check Memory Formation

```bash
# Agents should be forming memories about the work
ayo memory list --limit 10
# Expected: Memories about the password checker implementation
```

### 18.7 Verify Final Output

```bash
# After work completes, check tickets
ayo squad ticket list dev-team --status closed
# Expected: All tickets should be closed

# Check workspace for deliverables
cat ~/.local/share/ayo/sandboxes/squads/dev-team/workspace/password-checker.ts
# Expected: The implemented code

# Check for tests
cat ~/.local/share/ayo/sandboxes/squads/dev-team/workspace/password-checker.test.ts
# Expected: Unit tests
```

### 18.8 Semantic Search Verification

```bash
# Now that we have richer content, test semantic search
ayo index rebuild

ayo index search "password validation security"
# Expected: Relevant results based on squad mission and agent capabilities
```

---

## Debugging Reference

### Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `AYO_DEBUG=1` | Enable debug output | `AYO_DEBUG=1 ayo "test"` |
| `AYO_FLOW_NAME` | Set during flow execution | Automatic |
| `AYO_FLOW_RUN_ID` | Current flow run ID | Automatic |

### Key Log Files

| File | Purpose |
|------|---------|
| `~/.local/state/ayo/daemon.log` | Daemon activity log |
| `~/.local/share/ayo/logs/` | General log directory |

### Debug Scripts

```bash
# Full system info
./debug/system-info.sh

# Daemon status with logs
./debug/daemon-status.sh --logs

# Sandbox container status
./debug/sandbox-status.sh --verbose

# Comprehensive report
./debug/collect-all.sh --output report.txt
```

### Common Issues

#### Daemon Won't Start

```bash
# Remove stale socket
rm -f ~/.local/share/ayo/daemon.sock
rm -f /tmp/ayo/daemon.sock
rm -f "${XDG_RUNTIME_DIR}/ayo/daemon.sock"

# Restart
ayo sandbox service start
```

#### Sandbox Not Working

```bash
# Check sandbox provider
ayo doctor -v

# Verify Apple Container is running
ayo sandbox service status
```

#### Memory/Embedding Issues

```bash
# Check Ollama is running
ollama ps

# Verify embedding model
ollama pull nomic-embed-text

# Reindex
ayo index rebuild --force
ayo memory reindex
```

#### Squad Issues

```bash
# Check squad status
ayo squad show dev-team

# Restart squad
ayo squad stop dev-team
ayo squad start dev-team

# Check SQUAD.md is valid
cat ~/.local/share/ayo/sandboxes/squads/dev-team/SQUAD.md
```

### Inspection Commands

```bash
# Database inspection
sqlite3 ~/.local/share/ayo/db/ayo.db ".tables"
sqlite3 ~/.local/share/ayo/db/ayo.db "SELECT * FROM memories LIMIT 5;"

# Config inspection
cat ~/.config/ayo/ayo.json | jq .

# Session inspection
ayo session list --json | jq .

# Ticket file inspection
cat ~/.local/share/ayo/sandboxes/squads/dev-team/.tickets/*.md
```

---

## Success Criteria

After completing all phases, verify:

- [ ] Clean install completes without errors
- [ ] `ayo doctor` passes all checks
- [ ] Basic agent conversations work
- [ ] Skills are listed and can be attached
- [ ] Memories are stored and retrieved
- [ ] Sessions persist across invocations
- [ ] Sandbox isolation is enforced
- [ ] Shares work for explicit mounts
- [ ] Planners provide tools to agents
- [ ] Tickets can be created and managed
- [ ] Entity index returns semantic search results
- [ ] Squads can be created with SQUAD.md
- [ ] Squad I/O schemas validate data
- [ ] Squad dispatch routes to correct agent
- [ ] Flows execute with dependencies
- [ ] Flow-squad integration works
- [ ] Triggers can be created and fired
- [ ] Multi-agent coordination via tickets works
- [ ] End-to-end workflow completes successfully

---

## Clean Up After Testing

```bash
# Stop squad
ayo squad stop dev-team

# Destroy squad (optional)
ayo squad destroy dev-team

# Stop daemon
ayo sandbox service stop

# Remove test flows
rm -f ~/.config/ayo/flows/test-flow.yaml
rm -f ~/.config/ayo/flows/dep-flow.yaml
rm -f ~/.config/ayo/flows/squad-flow.yaml

# Remove test agents (optional)
ayo agents rm @reviewer
ayo agents rm @coder

# Clean up test files
rm -rf /tmp/ayo-test-share
rm -rf /tmp/ayo-watch-test
rm -f /tmp/test-input.json
rm -f /tmp/bad-input.json
```

---

## Notes for Future Testing

1. **Baseline diagnostics**: Save `/tmp/ayo-baseline-diagnostics.txt` before making changes
2. **Compare before/after**: Use `diff` on diagnostic reports to spot regressions
3. **Automate when ready**: Many of these manual tests can become automated integration tests
4. **Document failures**: If something fails, capture full debug output before fixing

This guide should be updated as new features are added to ayo.
