# PRD: Squad-Based Agent Coordination

## Overview

This document specifies the architecture for squad-based agent coordination in ayo. Squads are persistent or ephemeral sandboxes containing teams of agents that work together on tasks. A single `@ayo` agent orchestrates all squads from its own dedicated sandbox.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                              HOST                                    │
│                                                                      │
│   ayo CLI / Daemon (infrastructure only - no agent execution)       │
│                                                                      │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │                    @ayo's Sandbox                            │   │
│   │                                                              │   │
│   │   /home/ayo/              <- @ayo's home, state, todos       │   │
│   │   /squads/frontend/       <- mount to frontend squad         │   │
│   │   /squads/backend/        <- mount to backend squad          │   │
│   │   /squads/ephemeral-xyz/  <- mount to ephemeral squad        │   │
│   │   /output/                <- work product staging area       │   │
│   │                                                              │   │
│   │   @ayo executes here, creates tickets in squad mounts        │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                                    │                                 │
│              ┌─────────────────────┼─────────────────────┐          │
│              ▼                     ▼                     ▼          │
│   ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐   │
│   │  Squad:frontend │   │  Squad:backend  │   │ Squad:ephemeral │   │
│   │                 │   │                 │   │                 │   │
│   │  @react-dev     │   │  @api-dev       │   │  @task-agent    │   │
│   │  @designer      │   │  @db-admin      │   │                 │   │
│   │                 │   │                 │   │                 │   │
│   │  .tickets/      │   │  .tickets/      │   │  .tickets/      │   │
│   │  .context/      │   │  .context/      │   │  .context/      │   │
│   │  /workspace/    │   │  /workspace/    │   │  /workspace/    │   │
│   └─────────────────┘   └─────────────────┘   └─────────────────┘   │
│                                    │                                 │
│                                    ▼                                 │
│                         User's Working Directory                     │
│                         (work product synced here)                   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Core Concepts

### @ayo Agent

The single orchestrating agent for the entire system.

| Aspect | Specification |
|--------|---------------|
| Identity | Single logical agent, one system prompt, one personality |
| Execution | Runs inside its own dedicated sandbox (never on host) |
| Sandbox | Created during `ayo setup`, always exists, Alpine-based with git |
| Mounts | Has read/write mounts to all squad sandboxes via `/squads/{name}/` |
| State | Manages todos for coordination, has persistent home at `/home/ayo/` |
| Role | Creates tickets in squads, monitors progress, assembles results |

### Squads

Named sandboxes containing teams of agents.

| Aspect | Specification |
|--------|---------------|
| Definition | JSON config in `~/.config/ayo/squads/{name}.json` |
| Persistence | Can be persistent (survive restarts) or ephemeral (task-scoped) |
| Agents | Roster of agents assigned to the squad |
| Context | `.context/` directory preserves state across sessions |
| Tickets | `.tickets/` directory for task coordination |
| Workspace | `/workspace/` for work product creation |

### Agents in Squads

| Aspect | Specification |
|--------|---------------|
| Multi-squad | An agent can be in multiple squads |
| Isolation | Each squad instance has separate context (no cross-squad state) |
| Sessions | Each agent gets its own LLM session when working |
| Visibility | Agents see other agents' tickets in same squad (coordination) |
| Execution | Agents execute inside their squad's sandbox |

---

## Configuration

### Squad Configuration

Location: `~/.config/ayo/squads/{name}.json`

```json
{
  "name": "frontend",
  "description": "Frontend development team",
  "persistent": true,
  "agents": ["@react-dev", "@designer", "@css-wizard"],
  "resources": {
    "cpus": 2,
    "memory_mb": 2048
  },
  "image": "docker.io/library/alpine:3.21"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Unique squad identifier |
| `description` | string | no | Human-readable description |
| `persistent` | bool | no | If true, survives daemon restarts (default: true) |
| `agents` | string[] | yes | Agent handles assigned to this squad |
| `resources.cpus` | int | no | CPU limit (default: 2) |
| `resources.memory_mb` | int | no | Memory limit in MB (default: 2048) |
| `image` | string | no | Container image (default: alpine:3.21) |

### @ayo Sandbox Configuration

Location: `~/.config/ayo/ayo-sandbox.json`

```json
{
  "resources": {
    "cpus": 2,
    "memory_mb": 2048
  },
  "image": "docker.io/library/alpine:3.21",
  "packages": ["git"]
}
```

The @ayo sandbox is auto-created during `ayo setup` and always exists. @ayo can install additional packages as needed during operation.

---

## Directory Structure

### Host Filesystem

```
~/.config/ayo/
├── ayo.json                    # Main config
├── ayo-sandbox.json            # @ayo sandbox config
├── agents/                     # Agent definitions
│   ├── @react-dev/
│   ├── @designer/
│   └── @api-dev/
└── squads/                     # Squad definitions
    ├── frontend.json
    ├── backend.json
    └── devops.json

~/.local/share/ayo/
├── ayo.db                      # Sessions, memory
├── sandboxes/
│   ├── ayo/                    # @ayo sandbox data
│   │   └── home/               # /home/ayo contents
│   ├── frontend/               # Frontend squad data
│   │   ├── home/               # Agent homes
│   │   ├── context/            # Preserved context
│   │   ├── tickets/            # Squad tickets
│   │   └── workspace/          # Work products
│   └── backend/
│       └── ...
└── output/                     # Default work product destination
```

### Inside @ayo's Sandbox

```
/home/ayo/                      # @ayo's persistent home
/squads/
├── frontend/                   # Mount: squad's .tickets/, .context/, workspace/
│   ├── .tickets/
│   ├── .context/
│   └── workspace/
├── backend/
│   └── ...
└── ephemeral-abc123/
    └── ...
/output/                        # Mount: work product staging
```

### Inside Squad Sandboxes

```
/home/
├── react-dev/                  # @react-dev's home
├── designer/                   # @designer's home
└── ...
/.tickets/                      # Squad's ticket directory
/.context/                      # Preserved context across sessions
/workspace/                     # Work product directory
/shared/                        # Shared scratch space for agents
```

---

## Workflow

### User Request Flow

```
1. User runs: ayo "build me a web app with React frontend and Go API"

2. Daemon ensures @ayo sandbox is running

3. @ayo receives prompt, executes in its sandbox:
   - Creates todos for coordination:
     - "Create frontend squad tickets"
     - "Create backend squad tickets"  
     - "Monitor progress"
     - "Assemble final result"

4. @ayo creates/uses squads:
   - Uses existing "frontend" squad (persistent)
   - Creates ephemeral "backend-abc123" squad for this task
   - Writes tickets to /squads/{name}/.tickets/

5. @ayo signals "tickets ready" via RPC to daemon

6. Daemon notifies squad agents:
   - Spawns agent sessions in each squad
   - Each agent reads .tickets/, finds assigned work
   - Agents work independently, can see each other's tickets

7. Agents complete work:
   - Update ticket status (in_progress → closed)
   - Write work products to /workspace/
   - Daemon notifies @ayo of ticket completions

8. @ayo monitors progress:
   - Sees real-time ticket updates via /squads/{name}/.tickets/
   - TUI shows progress (using charm tooling)
   - Handles failures (retry, reassign, escalate)

9. All tickets complete:
   - @ayo collects work products from squad workspaces
   - Assembles into /output/
   - Daemon syncs /output/ to user's working directory (or --output path)
   - Ephemeral squads are destroyed

10. @ayo responds to user with summary
```

### Ticket Lifecycle in Squads

```
@ayo creates ticket          Agent picks up ticket
in /squads/X/.tickets/   →   reads from /.tickets/
        │                           │
        ▼                           ▼
   [pending]                  [in_progress]
        │                           │
        │                     Agent adds notes
        │                     Agent works
        │                           │
        │                           ▼
        │                     [closed] + work product
        │                           │
        ▼                           ▼
   @ayo sees closure         Daemon notifies @ayo
   via mount                 via RPC
```

---

## Daemon Responsibilities

### Sandbox Management

| Operation | Description |
|-----------|-------------|
| `EnsureAyoSandbox()` | Create @ayo sandbox if not exists, mount all squads |
| `CreateSquad(config)` | Create squad sandbox, register with @ayo mounts |
| `DestroySquad(name)` | Tear down squad, remove @ayo mount, cleanup |
| `ListSquads()` | Return all squad configs and status |
| `AddAgentToSquad(agent, squad)` | Update roster, ensure agent home in sandbox |
| `RemoveAgentFromSquad(agent, squad)` | Update roster, optionally preserve home |

### Ticket Coordination

| Operation | Description |
|-----------|-------------|
| `SignalTicketsReady(squad)` | @ayo calls this after creating tickets |
| `NotifyAgents(squad)` | Spawn agent sessions for assigned tickets |
| `OnTicketClosed(squad, ticket)` | Notify @ayo of completion |
| `OnAllTicketsClosed(squad)` | Signal squad work complete |

### Work Product Sync

| Operation | Description |
|-----------|-------------|
| `SyncWorkProduct(squad, target)` | Copy /workspace/ contents to target |
| `CleanupEphemeralSquad(squad)` | After sync, destroy ephemeral squad |

---

## RPC Protocol Additions

### Squad Management

```
squads.create       {name, config}           → {squad_id}
squads.destroy      {name, force}            → {}
squads.list         {}                       → {squads: [...]}
squads.get          {name}                   → {squad}
squads.add_agent    {squad, agent}           → {}
squads.remove_agent {squad, agent}           → {}
```

### Ticket Coordination

```
squads.tickets_ready    {squad}              → {}
squads.notify_agents    {squad}              → {sessions: [...]}
squads.wait_completion  {squad, timeout}     → {status, results}
```

### Work Product

```
squads.sync_output      {squad, target}      → {files: [...]}
squads.cleanup          {squad}              → {}
```

---

## CLI Commands

### Squad Management

```bash
# List squads
ayo squad list

# Create persistent squad
ayo squad create frontend --agents @react-dev,@designer

# Create ephemeral squad (destroyed after use)
ayo squad create temp-task --ephemeral --agents @worker

# Show squad details
ayo squad show frontend

# Add/remove agents
ayo squad add-agent frontend @new-agent
ayo squad remove-agent frontend @old-agent

# Destroy squad
ayo squad destroy temp-task
```

### Squad Tickets (from @ayo's perspective)

```bash
# @ayo creates tickets in a squad (run inside @ayo sandbox)
ayo ticket create "Build login component" -s frontend -a @react-dev
ayo ticket create "Design login page" -s frontend -a @designer

# Signal tickets ready
ayo squad start frontend

# Check progress
ayo squad status frontend

# Get work product
ayo squad output frontend --target ./build/
```

---

## Context Preservation

Squad sandboxes (except @ayo's) preserve context across sessions via `/.context/`:

| File | Purpose |
|------|---------|
| `/.context/session.json` | Last session metadata |
| `/.context/memory.json` | Agent memories from previous runs |
| `/.context/notes/` | Persistent notes between sessions |

When a squad is reactivated:
1. Daemon loads `/.context/` from persistent storage
2. Agents can read previous session state
3. Enables "pick up where you left off" for interrupted work

Ephemeral squads also have context - if an ephemeral squad is recreated with the same name pattern, it can optionally load context from a previous ephemeral squad.

---

## Progress Monitoring

### Real-time UI (using Charm tooling)

When @ayo is working, the CLI shows:

```
┌─────────────────────────────────────────────────────────────┐
│ @ayo: Building web app with React frontend and Go API       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Squads:                                                     │
│                                                             │
│ ▶ frontend (3 tickets)                                      │
│   ├─ ✓ Design login page (@designer) [closed]               │
│   ├─ ◐ Build login component (@react-dev) [in_progress]     │
│   │   └─ "Implementing form validation..."                  │
│   └─ ○ Style login page (@css-wizard) [pending]             │
│                                                             │
│ ▶ backend-abc123 (2 tickets)                                │
│   ├─ ◐ Create auth endpoints (@api-dev) [in_progress]       │
│   │   └─ "JWT middleware complete, working on routes"       │
│   └─ ○ Set up database schema (@db-admin) [blocked]         │
│       └─ "Waiting for auth endpoint spec"                   │
│                                                             │
│ @ayo todos:                                                 │
│   ✓ Create frontend squad tickets                           │
│   ✓ Create backend squad tickets                            │
│   ◐ Monitor progress                                        │
│   ○ Assemble final result                                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

The daemon streams ticket updates to @ayo's sandbox, which @ayo reads and the CLI renders.

---

## Failure Handling

### Agent Failure

```
1. Agent marks ticket as "blocked" with reason
2. Daemon notifies @ayo
3. @ayo evaluates:
   - Can another agent in the squad handle it? → Reassign
   - Is it a dependency issue? → Wait or reorder
   - Is it a real blocker? → Retry with different approach
4. If @ayo can't resolve after N attempts → Escalate to user
```

### Squad Failure

```
1. Sandbox crashes or becomes unresponsive
2. Daemon detects via health check
3. Daemon notifies @ayo
4. @ayo options:
   - Recreate squad, reload context, retry tickets
   - Fail gracefully, report to user
```

### @ayo Failure

```
1. @ayo sandbox crashes
2. Daemon detects, recreates @ayo sandbox
3. @ayo's /home/ayo/ is persistent, state survives
4. @ayo resumes from last known state
```

---

## Implementation Phases

### Phase 1: @ayo Sandbox (Foundation)

- Create @ayo sandbox during `ayo setup`
- @ayo executes inside sandbox, not on host
- Mount `/output/` for work product staging
- Basic work product sync to user's CWD

### Phase 2: Squad Infrastructure

- Squad config format (`~/.config/ayo/squads/*.json`)
- Squad sandbox lifecycle (create, destroy, list)
- Mount squads into @ayo's `/squads/` directory
- CLI: `ayo squad create/destroy/list/show`

### Phase 3: Ticket Integration

- Tickets write to squad's `/.tickets/` directory
- Daemon watches squad ticket directories
- `squads.tickets_ready` RPC for @ayo to signal
- Agent notification and session spawning

### Phase 4: Coordination Protocol

- `squads.notify_agents` spawns agent sessions
- Agents work on tickets, update status
- Daemon notifies @ayo of completions
- @ayo monitors via mounted ticket directories

### Phase 5: Work Product & Sync

- Agents write to squad's `/workspace/`
- @ayo collects from `/squads/{name}/workspace/`
- Sync to user's CWD or `--output` path
- Ephemeral squad cleanup after sync

### Phase 6: Context Preservation

- `/.context/` directory in squads
- Save/restore across sessions
- Ephemeral squad context linking

### Phase 7: Progress UI

- Real-time ticket streaming to CLI
- Charm-based TUI for progress display
- Integrated with existing ayo UI components

---

## Success Criteria

| Criteria | Measurement |
|----------|-------------|
| @ayo never executes on host | All @ayo bash commands run in sandbox |
| Squads isolate teams | Agents in different squads can't see each other's files |
| Work products delivered | Files sync to user's directory correctly |
| Context preserved | Interrupted work can be resumed |
| Failures handled | @ayo retries, then escalates gracefully |
| Real-time progress | User sees ticket updates as they happen |

---

## Open Items (None)

All questions have been resolved. This PRD is ready for implementation.
