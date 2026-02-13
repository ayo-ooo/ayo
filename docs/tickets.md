# Ticket-Based Coordination

Tickets provide a lightweight, file-based system for coordinating work between agents. Unlike real-time messaging, tickets create a persistent, auditable record of tasks and their progress.

## Overview

Tickets are markdown files with YAML frontmatter stored in session directories:

```
~/.local/share/ayo/sessions/{session-id}/.tickets/
├── proj-a1b2.md    # "Implement auth module"
├── proj-c3d4.md    # "Write unit tests"
└── proj-e5f6.md    # "Code review"
```

Each ticket contains:
- **Metadata**: Status, assignee, priority, dependencies
- **Description**: Task details in markdown
- **Notes**: Progress updates and comments

## Two-Tier Task Management

Ayo uses a two-tier system for task management:

| Layer | Tool | Scope | Lifetime | Purpose |
|-------|------|-------|----------|---------|
| **Near-term** | `todo` | Single session | Ephemeral | Steps to complete current work |
| **Medium/Long-term** | `ticket` | Across sessions | Persistent | Project-level work items |

### How They Work Together

```
Ticket: "Implement authentication module" (proj-a1b2)
  │
  └── Agent picks up ticket, starts session
      │
      └── Todo list (internal to this session):
          - [x] Read existing auth code
          - [x] Design JWT structure
          - [ ] Implement login endpoint
          - [ ] Write tests
```

**Todos** are the agent's internal working memory—what it's doing *right now* to make progress on a ticket. When the session ends, todos are discarded.

**Tickets** persist across sessions. When an agent finishes or hands off work:
1. It adds notes to the ticket documenting progress
2. It closes or updates the ticket status
3. The todo list disappears with the session
4. A new agent (or same agent, new session) can pick up the ticket and create fresh todos

### Example Workflow

```bash
# Coordinator creates a ticket (persists in files)
ayo ticket create "Implement auth module" -a @backend -s project

# @backend agent starts working
# Internally, agent creates todos:
#   - Read existing code
#   - Design token structure
#   - Implement endpoints

# Agent adds progress notes to ticket (persists)
ayo ticket note auth-impl "Completed JWT design, starting endpoints"

# Session ends (agent stops, context limit, etc.)
# Todos are gone, but ticket and notes remain

# Later, same or different agent resumes
ayo ticket show auth-impl -s project
# Sees: "Completed JWT design, starting endpoints"

# Agent creates new todos for remaining work
# Continues from where previous session left off
```

### When to Use Each

| Situation | Use Todos | Use Tickets |
|-----------|-----------|-------------|
| Breaking down current task into steps | ✓ | |
| Tracking what you're doing right now | ✓ | |
| Work that spans multiple sessions | | ✓ |
| Coordinating between agents | | ✓ |
| Creating audit trail | | ✓ |
| Dependencies between work items | | ✓ |
| Handing off work to another agent | | ✓ |

## When to Use Tickets

| Scenario | Use Tickets | Use Other |
|----------|-------------|-----------|
| Multi-step project | ✓ | |
| Task with dependencies | ✓ | |
| Work needs audit trail | ✓ | |
| Simple delegation | | `agent_call` |
| Data pipeline | | Chaining |
| Real-time collaboration | | Shared files |

## Quick Start

### Create a Ticket

```bash
# Simple task
ayo ticket create "Implement user authentication" -s my-session

# With assignment and priority
ayo ticket create "Fix login bug" -a @debugger -p 1 -s my-session

# With dependencies
ayo ticket create "Deploy to staging" --deps auth-impl,tests -s my-session
```

### Work on Tickets

```bash
# See what's ready to work on
ayo ticket ready -a @coder -s my-session

# Start working
ayo ticket start proj-a1b2 -s my-session

# Add progress notes
ayo ticket note proj-a1b2 "Completed login endpoint" -s my-session

# Mark complete
ayo ticket close proj-a1b2 -s my-session
```

### View Tickets

```bash
# List all tickets
ayo ticket list -s my-session

# Filter by status
ayo ticket list --status in_progress -s my-session

# Show ticket details
ayo ticket show proj-a1b2 -s my-session
```

## Ticket Lifecycle

```
                    ┌─────────────┐
                    │   pending   │
                    └──────┬──────┘
                           │ start
                           ▼
    ┌─────────────────────────────────────────┐
    │              in_progress                │
    └───────┬─────────────────────┬───────────┘
            │ close               │ block
            ▼                     ▼
    ┌─────────────┐       ┌─────────────┐
    │   closed    │       │   blocked   │
    └─────────────┘       └──────┬──────┘
            ▲                    │ reopen
            │                    ▼
            │             ┌─────────────┐
            └─────────────│   pending   │
                          └─────────────┘
```

### Status Meanings

| Status | Meaning |
|--------|---------|
| `pending` | Not yet started, waiting for assignment or dependencies |
| `in_progress` | Actively being worked on |
| `blocked` | Cannot proceed (dependency, external blocker) |
| `closed` | Completed |

## Dependencies

Tickets can depend on other tickets. A ticket with unresolved dependencies won't appear in the "ready" list.

```bash
# Add a dependency
ayo ticket dep add deploy-001 auth-impl -s my-session

# Remove a dependency
ayo ticket dep remove deploy-001 auth-impl -s my-session

# See what's blocked
ayo ticket blocked -s my-session
```

The system prevents dependency cycles—you can't create circular dependencies.

## Ticket Types

| Type | Use Case |
|------|----------|
| `task` | General work item (default) |
| `bug` | Defect to fix |
| `feature` | New functionality |
| `subtask` | Child of another ticket |

```bash
# Create a bug ticket
ayo ticket create "Login fails with special chars" --type bug -s my-session

# Create subtasks
ayo ticket create "Implement endpoint" --parent auth-impl -s my-session
ayo ticket create "Write tests" --parent auth-impl -s my-session
```

## Priority

Priority is a number where lower = more urgent:

| Priority | Meaning |
|----------|---------|
| 0 | Critical/Blocker |
| 1 | High |
| 2 | Medium (default) |
| 3 | Low |

```bash
ayo ticket create "Security vulnerability" -p 0 -s my-session
```

## Agent Workflow

When an agent is working in a session with tickets, it receives instructions for using the ticket system. Here's the typical workflow:

### 1. Find Work

```bash
# List assigned tickets that are ready (deps resolved)
ayo ticket ready -a @coder

# Or see all assigned tickets
ayo ticket list -a @coder
```

### 2. Start Working

```bash
# Claim the ticket
ayo ticket start proj-a1b2

# This sets status to in_progress
```

### 3. Track Progress

```bash
# Add notes as you work
ayo ticket note proj-a1b2 "Implemented login endpoint"
ayo ticket note proj-a1b2 "Added password validation"
```

### 4. Handle Blockers

```bash
# If blocked, mark it and explain
ayo ticket block proj-a1b2
ayo ticket note proj-a1b2 "Blocked: need API spec from @architect"
```

### 5. Complete

```bash
# Mark done
ayo ticket close proj-a1b2
```

### 6. Create Follow-up Work

```bash
# If new work discovered, create tickets
ayo ticket create "Refactor token handling" --deps proj-a1b2
```

## File Format

Tickets are markdown with YAML frontmatter:

```markdown
---
id: proj-a1b2
status: in_progress
type: task
priority: 2
assignee: "@coder"
created: 2024-01-15T10:30:00Z
started: 2024-01-15T11:00:00Z
deps: []
links: []
tags: [auth, backend]
---

# Implement user authentication

Add JWT-based authentication to the API.

## Requirements

- Login endpoint with email/password
- Token refresh endpoint
- Logout endpoint

## Notes

**2024-01-15 11:05** - Started implementation, using existing user model

**2024-01-15 14:30** - Login endpoint complete, moving to refresh
```

## External References

Link tickets to external systems:

```bash
# Link to GitHub issue
ayo ticket create "Fix bug" --ref "github:org/repo#123" -s my-session

# Link to Jira
ayo ticket create "Feature" --ref "jira:PROJ-456" -s my-session
```

## CLI Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `ayo ticket list` | List tickets |
| `ayo ticket create` | Create a ticket |
| `ayo ticket show` | Show ticket details |
| `ayo ticket delete` | Delete a ticket |

### Status Commands

| Command | Description |
|---------|-------------|
| `ayo ticket start` | Set status to in_progress |
| `ayo ticket close` | Set status to closed |
| `ayo ticket reopen` | Reopen a closed ticket |
| `ayo ticket block` | Set status to blocked |

### Workflow Commands

| Command | Description |
|---------|-------------|
| `ayo ticket ready` | List tickets ready to work on |
| `ayo ticket blocked` | List blocked tickets |
| `ayo ticket assign` | Assign ticket to agent |
| `ayo ticket note` | Add a note to ticket |

### Dependency Commands

| Command | Description |
|---------|-------------|
| `ayo ticket dep add` | Add dependency |
| `ayo ticket dep remove` | Remove dependency |

### Common Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--session` | `-s` | Session ID (required) |
| `--assignee` | `-a` | Filter by or set assignee |
| `--status` | | Filter by status |
| `--priority` | `-p` | Set priority (0-3) |
| `--type` | `-t` | Set ticket type |
| `--deps` | | Comma-separated dependency IDs |
| `--parent` | | Parent ticket ID |
| `--ref` | | External reference |
| `--json` | | JSON output |

## Tickets in Squads

When working within a squad, tickets are stored in the squad's `.tickets/` directory:

```
~/.local/share/ayo/sandboxes/squads/{name}/.tickets/
├── alpha-a1b2.md    # "Implement auth"
├── alpha-c3d4.md    # "Write tests"
└── alpha-e5f6.md    # "Code review"
```

### Squad Ticket Commands

```bash
# Create ticket in squad
ayo squad ticket my-squad create "Implement login" -a @backend

# List squad tickets
ayo squad ticket my-squad list

# Show ticket details
ayo squad ticket my-squad show alpha-a1b2

# Mark complete
ayo squad ticket my-squad close alpha-a1b2
```

### Tickets + SQUAD.md

Tickets define *what* needs to be done. The squad's `SQUAD.md` constitution defines *how* agents should work together:

```markdown
# SQUAD.md coordination section example

## Coordination

1. @backend creates API endpoint
2. @backend creates ticket for @frontend when ready
3. @frontend implements UI after @backend ticket closes
4. @qa reviews after each component completes

Use ticket dependencies:
- frontend-login depends on backend-login
- qa-review depends on both
```

The constitution provides shared context that helps agents interpret tickets consistently.

## Comparison to Other Coordination

| Feature | Tickets | Flows | Delegation |
|---------|---------|-------|------------|
| **Persistence** | ✓ Files on disk | None | None |
| **Parallelism** | ✓ Multiple agents | Sequential | Single call |
| **Dependencies** | ✓ Built-in | Chained output | Direct call |
| **Audit trail** | ✓ Git-friendly | None | None |
| **Isolation** | Squad-scoped | Process-scoped | Process-scoped |

**Use Tickets when:**
- Work spans multiple agents
- Need dependency ordering
- Want audit trail
- Working within a squad

**Use Flows when:**
- Work is sequential pipeline
- Output of one → input of next
- No parallel work

**Use Delegation when:**
- Quick synchronous subtask
- No persistence needed
- Single agent helps another
