# Ayo Go-To-Market Refinement Plan

> **Status**: Active planning document for GTM branch
> **Goal**: Transform ayo into a coherent, broadly useful system for managing AI agents

---

## Executive Summary

Ayo is a CLI framework for creating, managing, and orchestrating AI agents that operate in isolated sandbox environments. After months of development, we have accumulated many good ideas in a system that almost works together but lacks cohesion. This plan outlines the work to bring ayo to a go-to-market ready state through **simplification**, **cohesion**, and **polish**.

### Core Value Proposition

> **Ayo lets you manage the AI agents that work for you, while providing flexible hooks for experimenting with new forms of agent harnesses.**

Key differentiators:
- **Sandboxed execution**: Agents run in isolated containers, not on your host
- **Multi-agent coordination**: Squads with shared sandboxes for team collaboration
- **Flexible triggers**: Time-based and event-based ambient agents
- **Experimentation-friendly**: Same agent can behave differently in different sandboxes

---

## Current State Analysis

### What Works (Keep & Polish)

| Component | Status | Notes |
|-----------|--------|-------|
| **Sandbox providers** | Good | Apple Container (macOS 26+), systemd-nspawn (Linux) |
| **Agent definition** | Good | Directory-based agents with system.md, config.json |
| **@ayo default agent** | Good | But needs clearer sandbox semantics |
| **Squads** | Good foundation | Constitution in SQUAD.md, ticket coordination |
| **Planners** | Good foundation | Plugin architecture with todos/tickets |
| **Trigger engine** | Good foundation | Cron + file watch in daemon |
| **Share system** | Good foundation | Host directory mounting |

### What's Confusing (Simplify/Remove)

| Component | Issue | Decision |
|-----------|-------|----------|
| **REST API server** | Overkill for CLI tool, confuses purpose | **REMOVE** |
| **Flows/Chains** | Shell scripts with extra steps | **SIMPLIFY** - keep DAG inspection only |
| **Web interface** | Incomplete, distracts from CLI focus | **REMOVE** |
| **YAML executor** | Overly complex flow definition | **REMOVE** |
| **Webhook server** | Premature integration point | **DEFER** |
| **Tunnel/QR code** | Mobile connectivity features | **REMOVE** |
| **IRC integration** | Abandoned experiment | **REMOVE** |
| **SQUAD.md frontmatter** | Inconsistent with agent config | **MIGRATE** to ayo.json |

### What Needs Work (Build/Refine)

| Component | Work Needed |
|-----------|-------------|
| **Sandbox coexistence model** | Define how agents share vs isolate |
| **Host mount semantics** | Mount user's home to /mnt/{username} read-only |
| **File request workflow** | Agent requests → user approves → file written |
| **--no-jodas mode** | Bypass permission prompts for power users |
| **Unified ayo.json schema** | Single config format for agents AND squads |
| **Advanced scheduler** | Replace robfig/cron with gocron v2 |
| **Ambient triggers** | Event/time-based proactive agent execution |

---

## Sandbox Architecture: Agent Coexistence Model

### The Key Question

> Should each agent get its own sandbox, or do agents coexist in a shared sandbox?

### Decision: Shared Default Sandbox with Optional Isolation

**Default behavior**: All agents invoked directly (`ayo @agent "prompt"`) execute in the **@ayo sandbox**. This is the "workbench" where you interact with agents.

**Squads**: Get their own isolated sandbox where multiple agents collaborate.

**Explicit isolation**: Agents can request their own sandbox via config.

```
┌─────────────────────────────────────────────────────────────────────┐
│                         SANDBOX LANDSCAPE                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────────────────────────────────────┐                   │
│  │           @AYO SANDBOX (default)             │                   │
│  │  /home/ayo/          - @ayo's home           │                   │
│  │  /home/crush/        - @crush's home         │                   │
│  │  /home/reviewer/     - @reviewer's home      │                   │
│  │  /mnt/{user}/        - Host home (read-only) │                   │
│  │  /workspace/         - Shared workspace      │                   │
│  │  /output/            - Safe write zone       │                   │
│  │                                              │                   │
│  │  When you run: ayo @crush "write code"       │                   │
│  │  → @crush executes HERE, in @ayo sandbox     │                   │
│  └──────────────────────────────────────────────┘                   │
│                                                                      │
│  ┌──────────────────────────────────────────────┐                   │
│  │           #dev-team SQUAD SANDBOX            │                   │
│  │  /home/frontend/     - @frontend's home      │                   │
│  │  /home/backend/      - @backend's home       │                   │
│  │  /workspace/         - Shared code           │                   │
│  │  /.tickets/          - Coordination          │                   │
│  │                                              │                   │
│  │  When you run: ayo #dev-team "build feature" │                   │
│  │  → Squad lead orchestrates HERE              │                   │
│  └──────────────────────────────────────────────┘                   │
│                                                                      │
│  ┌──────────────────────────────────────────────┐                   │
│  │     @isolated-agent SANDBOX (if configured)  │                   │
│  │  sandbox: { isolated: true } in ayo.json     │                   │
│  └──────────────────────────────────────────────┘                   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Why Shared by Default?

1. **Simpler mental model**: One sandbox to understand and explore
2. **File sharing**: Agents can hand off files to each other naturally
3. **Resource efficiency**: One container vs many
4. **Easier debugging**: `ayo sandbox shell` drops you into familiar environment
5. **Natural orchestration**: @ayo can invoke other agents and they see the same files

### When to Use Isolated Sandboxes

1. **Squads**: Always isolated - they need their own workspace and tickets
2. **Untrusted agents**: Agents you don't fully trust get their own box
3. **Resource-intensive agents**: Agents that need special resources
4. **Conflicting dependencies**: Agents that need different environments

### Agents as Real Unix Users

Each agent runs as a **real Unix user** inside the sandbox, not a shared user with `$HOME` tricks:

```
/home/
├── ayo/              # @ayo user (orchestrator)
├── crush/            # @crush user (coding agent)
├── reviewer/         # @reviewer user
└── {agent-name}/     # Created on first use
```

**Why real users?**
- Clear ownership and permissions (`ls -la` shows who created files)
- Standard Unix semantics (no confusion about identity)
- Agents can `su` to each other if needed for handoff
- Process isolation via standard Unix mechanisms

---

## Sandbox Bootstrap: ayod

To simplify sandbox management, we install a lightweight **ayod** (ayo daemon) inside each sandbox. This replaces ad-hoc setup with a clean, extensible service.

### What ayod Does

| Function | Description |
|----------|-------------|
| **User management** | Creates agent users on-demand (`ayod useradd @agent`) |
| **Environment setup** | Sets up agent home directories with correct permissions |
| **File request proxy** | Handles `file_request` tool calls, proxies to host daemon |
| **Output sync** | Syncs `/output/` contents to host |
| **Health check** | Reports sandbox status to host daemon |

### Bootstrap Flow

```
1. Sandbox created with base image (alpine/debian)
2. Host injects ayod binary + config into /usr/local/bin/ayod
3. ayod starts as PID 1 (replaces "sleep infinity")
4. ayod listens on /run/ayod.sock for commands
5. Host daemon connects via mounted socket

On agent invocation:
1. Host: "Run @crush with this command"
2. ayod: Ensures @crush user exists
3. ayod: Executes command as @crush user
4. ayod: Returns output to host
```

### ayod Binary

Small, statically-compiled Go binary (~5MB) included in ayo distribution:

```go
// cmd/ayod/main.go
package main

func main() {
    // Listen on /run/ayod.sock
    // Handle: useradd, exec, sync, health
}
```

### Benefits

1. **Single entry point**: All sandbox operations go through ayod
2. **Extensible**: Easy to add new capabilities (package install, network config)
3. **Debuggable**: `ayo sandbox shell` can talk to ayod directly
4. **Consistent**: Same behavior across Apple Container and systemd-nspawn
5. **Clean separation**: Host daemon doesn't need provider-specific exec code

---

## File Access & Permission Model

### File System Layout

```
SANDBOX FILESYSTEM:
/
├── home/
│   └── {agent}/              # Per-agent home directories
│       ├── .config/          # Agent-specific config
│       └── .local/           # Agent-specific data
├── mnt/
│   └── {host_username}/      # Host home directory (READ-ONLY)
│       ├── Documents/
│       ├── Projects/
│       └── ...
├── workspace/                # Shared workspace (read-write)
└── output/                   # WRITE ZONE - syncs to host
    └── {session_id}/         # Per-session output
```

### File Access Model

| Zone | Access | Purpose |
|------|--------|---------|
| `/home/{agent}` | Read-write | Agent's persistent storage |
| `/mnt/{user}` | **Read-only** | Access to host files |
| `/workspace` | Read-write | Shared collaboration space |
| `/output/{session}` | Read-write | **Safe write zone** - auto-syncs to host |

### File Request Workflow

When an agent needs to modify files on the host:

```
1. Agent detects need to modify /mnt/user/Projects/myapp/main.go
2. Agent calls file_request tool:
   file_request({
     "action": "update",
     "path": "/mnt/user/Projects/myapp/main.go",
     "content": "...",
     "reason": "Fixed bug in authentication"
   })
3. User sees prompt in terminal:
   ┌─────────────────────────────────────────────┐
   │ @ayo wants to update:                       │
   │   ~/Projects/myapp/main.go                  │
   │ Reason: Fixed bug in authentication         │
   │                                             │
   │ [Y]es  [N]o  [D]iff  [A]lways for session   │
   └─────────────────────────────────────────────┘
4. User approves → file is written to host
```

### --no-jodas Mode

For power users who trust their agents, `--no-jodas` mode auto-approves all file requests:

```bash
# Enable for a session
ayo --no-jodas "refactor my entire codebase"

# Enable globally in config
# ~/.config/ayo/config.json
{
  "permissions": {
    "no_jodas": true
  }
}

# Enable per-agent
# ~/.config/ayo/agents/@trusted/ayo.json
{
  "permissions": {
    "auto_approve": true
  }
}
```

**Safety considerations**:
- `--no-jodas` still respects the `/mnt/{user}` boundary (only host home, not system)
- All file modifications are logged to `~/.local/share/ayo/audit.log`
- Can be combined with `--dry-run` to see what would happen

---

## Unified Configuration: ayo.json

### Current Problem

We have inconsistent configuration:
- Agents use `config.json` with one schema
- Squads use `SQUAD.md` frontmatter with different fields
- Global config in `~/.config/ayo/config.json` has yet another schema

### Solution: Unified ayo.json Schema

Both agents and squads use `ayo.json` with namespaced sections:

```json
{
  "$schema": "https://ayo.dev/schemas/ayo.json",
  "version": "1",
  
  "agent": {
    "description": "A helpful coding assistant",
    "model": "claude-sonnet-4-5-20250929",
    "tools": ["bash", "memory", "file_request"],
    "skills": ["coding", "debugging"],
    "memory": {
      "enabled": true,
      "scope": "global"
    },
    "sandbox": {
      "isolated": false,
      "network": true
    },
    "permissions": {
      "auto_approve": false
    }
  }
}
```

For squads:

```json
{
  "$schema": "https://ayo.dev/schemas/ayo.json",
  "version": "1",
  
  "squad": {
    "description": "Development team for auth features",
    "lead": "@architect",
    "input_accepts": "@planner",
    "agents": ["@frontend", "@backend", "@qa"],
    "planners": {
      "near_term": "ayo-todos",
      "long_term": "ayo-tickets"
    },
    "sandbox": {
      "image": "alpine:3.21",
      "network": true,
      "mounts": ["~/Projects/myapp:/workspace"]
    },
    "io": {
      "input_schema": "input.jsonschema",
      "output_schema": "output.jsonschema"
    }
  }
}
```

### SQUAD.md Becomes Pure Documentation

SQUAD.md remains as the human-readable constitution but NO LONGER contains configuration:

```markdown
# Squad: auth-team

## Mission
Implement secure authentication for the e-commerce platform.

## Agents

### @architect (Lead)
Decomposes tasks, reviews output, makes architectural decisions.

### @backend
Implements API endpoints, writes tests.

### @frontend
Implements UI components, integrates with backend.

## Coordination
All work flows through tickets. @architect creates and assigns.
```

### Migration Path

1. Parse existing SQUAD.md frontmatter
2. Generate `ayo.json` with squad section
3. Strip frontmatter from SQUAD.md
4. Deprecation warning for old format

---

## Advanced Scheduler: gocron v2

### Current Problem

We use `robfig/cron` which only supports cron expressions. Users need:
- One-time scheduled jobs
- "In 30 minutes" style scheduling
- Weekly/monthly jobs with friendly syntax
- Job persistence across daemon restarts
- Job monitoring and history

### Solution: Migrate to go-co-op/gocron v2

[gocron v2](https://github.com/go-co-op/gocron) provides:

| Feature | robfig/cron | gocron v2 |
|---------|-------------|-----------|
| Cron expressions | ✓ | ✓ |
| Duration-based | ✗ | ✓ (`10*time.Second`) |
| One-time jobs | ✗ | ✓ (`OneTimeJob`) |
| Daily/Weekly/Monthly | ✗ | ✓ (fluent API) |
| Random intervals | ✗ | ✓ |
| Singleton mode | ✗ | ✓ |
| Concurrency limits | ✗ | ✓ |
| Event listeners | ✗ | ✓ |
| Distributed locking | ✗ | ✓ |
| Monitoring interface | ✗ | ✓ |

### New Trigger Configuration

```yaml
# ~/.config/ayo/triggers/morning-standup.yaml
name: morning-standup
type: daily
schedule:
  times: ["09:00"]
  days: [monday, tuesday, wednesday, thursday, friday]
agent: "@standup"
prompt: "Generate daily standup report"
output: /output/standup/{date}.md

# One-time job
name: deploy-reminder
type: once
schedule:
  at: "2026-02-24T14:00:00Z"
agent: "@notifier"
prompt: "Remind about deployment"

# Duration-based
name: health-check
type: interval
schedule:
  every: 5m
agent: "@monitor"
prompt: "Check system health"
singleton: true  # Don't overlap
```

### Persistence

Jobs are stored in SQLite and restored on daemon restart:

```sql
CREATE TABLE scheduled_jobs (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL,  -- 'cron', 'daily', 'weekly', 'once', 'interval'
  schedule TEXT NOT NULL,  -- JSON
  agent TEXT NOT NULL,
  prompt TEXT,
  last_run_at TIMESTAMP,
  next_run_at TIMESTAMP,
  enabled BOOLEAN DEFAULT true
);
```

---

## Code Removal Plan

### Files/Packages to Remove

| Path | Reason | Lines |
|------|--------|-------|
| `internal/server/` | REST API - not needed for CLI | ~2500 |
| `web/` | Web interface | ~1000 |
| `cmd/ayo/serve.go` | Server command | ~100 |
| `cmd/ayo/chat.go` | Web chat | ~50 |
| `internal/flows/yaml_executor.go` | Complex YAML flows | ~500 |
| `internal/flows/yaml_validate.go` | YAML validation | ~200 |
| `internal/daemon/webhook_server.go` | Premature | ~500 |
| `internal/server/tunnel/` | Cloudflare tunnel | ~200 |
| `internal/server/qrcode.go` | Mobile QR | ~100 |

**Estimated removal: ~5,000+ lines**

### Files to Simplify

| Path | Change |
|------|--------|
| `internal/flows/` | Keep only discover, parse, DAG inspection |
| `internal/daemon/server.go` | Remove webhook, serve endpoints |
| `cmd/ayo/flows.go` | Remove run/execute, keep inspect/graph |

---

## Implementation Phases

### Phase 1: Foundation (Simplification)

**Goal**: Remove complexity, establish clear mental model

1. Remove server, web UI, webhook code
2. Remove YAML flow executor
3. Simplify daemon to core functions
4. Implement shared sandbox with per-agent homes
5. Implement host mount at /mnt/{username}

### Phase 2: File System & Permissions

**Goal**: Clear, safe file access patterns

1. Implement file_request tool
2. Add approval UI to terminal
3. Implement --no-jodas mode
4. Add /output safe write zone
5. Implement audit logging

### Phase 3: Unified Configuration

**Goal**: Single ayo.json schema for agents and squads

1. Design unified ayo.json schema
2. Implement ayo.json loader for agents
3. Implement ayo.json loader for squads
4. Migrate SQUAD.md frontmatter to ayo.json
5. Update CLI commands for new schema

### Phase 4: Advanced Scheduler

**Goal**: Powerful, persistent scheduling with gocron v2

1. Replace robfig/cron with gocron v2
2. Implement job persistence in SQLite
3. Add one-time and duration jobs
4. Implement job monitoring
5. Add trigger CLI improvements

### Phase 5: Squad Polish

**Goal**: Squads as first-class coordination primitive

1. Clarify squad lead semantics
2. Implement squad dispatch routing
3. Add I/O schema enforcement
4. Polish ticket tools for agents

### Phase 6: Documentation & Polish

**Goal**: Make ayo approachable

1. Rewrite all documentation
2. Add examples and recipes
3. Polish CLI help text
4. Add `ayo doctor` improvements

---

## Success Criteria

### For GTM Readiness

- [ ] New user can install and run first agent in < 5 minutes
- [ ] Documentation explains all concepts clearly
- [ ] `ayo doctor` catches all setup issues
- [ ] Squads work reliably for multi-agent tasks
- [ ] Triggers enable ambient agent use cases
- [ ] No dead code or unused features
- [ ] Test coverage > 70%

### Mental Model Test

A user should be able to answer:
1. "What is ayo?" → CLI for managing AI agents in sandboxes
2. "Where do agents run?" → In the @ayo sandbox, or isolated squad sandboxes
3. "What's a squad?" → A team of agents with their own sandbox
4. "What's a trigger?" → What makes an agent act without prompting
5. "What's --no-jodas?" → Auto-approve mode for power users

---

## Ticket Tracking

All implementation work is tracked in `.tickets/`. Use `tk list` to see current state.

### Epic Structure

| Epic | Phase | Dependencies | Status |
|------|-------|--------------|--------|
| `ayo-6h19` | Phase 1: Foundation | - | Open |
| `ayo-whmn` | Phase 2: File System | Phase 1 | Open |
| `ayo-pv3a` | Phase 3: Unified Config | Phase 2 | Open |
| `ayo-sqad` | Phase 4: Advanced Scheduler | Phase 3 | Open |
| `ayo-xfu3` | Phase 5: Squad Polish | Phase 4 | Open |
| `ayo-i2qo` | Phase 6: Documentation | Phase 5 | Open |

### Quick Commands

```bash
# See all tickets
tk list

# See dependency tree
tk dep tree ayo-i2qo

# Start work on a ticket
tk start <id>

# Close completed ticket
tk close <id>

# See what's ready to work on
tk ready
```

---

*Last updated: 2026-02-23*
