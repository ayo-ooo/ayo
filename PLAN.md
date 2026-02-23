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
- **Multi-agent coordination**: Squads with SQUAD.md constitutions for team collaboration
- **Flexible triggers**: Time-based and event-based ambient agents
- **Experimentation-friendly**: Same agent can behave differently in different sandboxes

---

## Current State Analysis

### What Works (Keep & Polish)

| Component | Status | Notes |
|-----------|--------|-------|
| **Sandbox providers** | Good | Apple Container (macOS 26+), systemd-nspawn (Linux) |
| **Agent definition** | Good | Directory-based agents with system.md, config.json |
| **@ayo default agent** | Good | But needs clearer sandbox home directory semantics |
| **Squads** | Good foundation | SQUAD.md constitution, ticket coordination |
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

### What Needs Work (Build/Refine)

| Component | Work Needed |
|-----------|-------------|
| **@ayo home directory** | Standardize /home/ayo in sandbox |
| **Host mount semantics** | Mount user's home to /mnt/{username} read-only |
| **File request workflow** | Agent requests → user approves → file written |
| **Safe write zone** | Designated area for unrestricted agent writes |
| **Squad lead orchestration** | Clear lead agent semantics, work distribution |
| **Ambient triggers** | Event/time-based proactive agent execution |
| **I/O schemas for squads** | Enforce input/output contracts |

---

## Architecture Simplification

### Current Architecture (Too Complex)

```
┌─────────────────────────────────────────────────────────────────────┐
│                              HOST                                    │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐  │
│  │   ayo CLI       │    │   REST Server   │    │   Web UI        │  │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘  │
│           │                     │                      │            │
│           ▼                     ▼                      ▼            │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                         DAEMON                                   ││
│  │  • Session management    • Webhook server    • Trigger engine   ││
│  │  • Flow executor         • Matrix broker     • IRC bridge       ││
│  └─────────────────────────────────────────────────────────────────┘│
│                              │                                      │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                       SANDBOX(ES)                               ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

### Target Architecture (Simplified)

```
┌─────────────────────────────────────────────────────────────────────┐
│                              HOST                                    │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                         AYO CLI                                  ││
│  │  • Direct agent invocation    • Squad dispatch                  ││
│  │  • Trigger management         • Shell access to sandboxes       ││
│  └─────────────────────────────────────────────────────────────────┘│
│                              │                                      │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                         DAEMON                                   ││
│  │  • Session lifecycle     • Sandbox pool                         ││
│  │  • Trigger engine        • Ticket watcher                       ││
│  └─────────────────────────────────────────────────────────────────┘│
│                              │                                      │
│           ┌──────────────────┼──────────────────┐                   │
│           ▼                  ▼                  ▼                   │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐             │
│  │ @ayo sandbox│    │Squad sandbox│    │Agent sandbox│             │
│  │ /home/ayo   │    │ /workspace  │    │ /home/agent │             │
│  └─────────────┘    └─────────────┘    └─────────────┘             │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Sandbox & File System Model

### @ayo Default Agent Sandbox

The default `@ayo` agent should feel like a regular user account:

```
SANDBOX FILESYSTEM:
/
├── home/
│   └── ayo/                    # Agent's home directory (persistent)
│       ├── .config/            # Agent configuration
│       ├── .local/             # Local data
│       └── workspace/          # Current working directory
├── mnt/
│   └── {host_username}/        # Host home directory (READ-ONLY)
│       ├── Documents/
│       ├── Projects/
│       └── ...
├── workspace/                  # Shared workspace for collaborations
└── output/                     # WRITE ZONE - unrestricted agent writes
    └── {session_id}/           # Per-session output directory
```

### File Access Model

| Zone | Access | Purpose |
|------|--------|---------|
| `/home/ayo` | Read-write | Agent's persistent storage |
| `/mnt/{user}` | Read-only | Access to host files |
| `/workspace` | Read-write | Shared collaboration space |
| `/output/{session}` | Read-write | **Safe write zone** - freely writable to host |

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

### Safe Write Zone

Files written to `/output/{session}/` are automatically synced to host:
- Host path: `~/.local/share/ayo/output/{session}/`
- No approval needed
- Agent can generate reports, code, artifacts freely
- User retrieves when ready

---

## Squad Coordination Model

### Squad Structure

```
~/.local/share/ayo/sandboxes/squads/{squad-name}/
├── SQUAD.md               # Constitution (mission, roles, rules)
├── input.jsonschema       # Optional: input contract
├── output.jsonschema      # Optional: output contract
├── .tickets/              # Coordination tickets
├── workspace/             # Shared code workspace
└── agent-homes/           # Per-agent home directories
    ├── @frontend/
    ├── @backend/
    └── @qa/
```

### SQUAD.md Constitution

```markdown
---
lead: "@architect"
input_accepts: "@planner"
planners:
  near_term: ayo-todos
  long_term: ayo-tickets
---
# Squad: ecommerce-auth

## Mission
Implement secure authentication for the e-commerce platform.

## Agents

### @architect (Lead)
- Decomposes tasks into tickets
- Reviews agent output
- Makes architectural decisions

### @backend
- Implements API endpoints
- Writes tests

### @frontend  
- Implements UI components
- Integrates with backend

## Coordination
1. All work flows through tickets
2. @architect creates and assigns tickets
3. Agents close tickets when done
4. @architect reviews and approves
```

### Squad Lead Semantics

The squad lead (`@architect` in example above):
- Receives all unrouted dispatches to the squad
- Has `ticket_create`, `ticket_assign`, `delegate` tools
- Does NOT have direct file editing (must delegate)
- Synthesizes final output from agent work

---

## Planner System

### Two Planning Horizons

| Type | Scope | Default Plugin | Purpose |
|------|-------|----------------|---------|
| **Near-term** | Session | `ayo-todos` | Sequential task tracking |
| **Long-term** | Persistent | `ayo-tickets` | Multi-agent coordination |

### Near-Term: ayo-todos

Simple todo list for individual agent session:

```
# Agent's internal todo tracking
[ ] Parse user request
[x] Find relevant files
[ ] Make changes
[ ] Run tests
```

- SQLite-backed, session-scoped
- Tools: `todo_add`, `todo_complete`, `todo_list`
- Visible in TUI sidebar

### Long-Term: ayo-tickets

Markdown ticket system for persistent work:

```markdown
---
id: auth-0a3b
status: in_progress
assignee: @backend
deps: [auth-0x01]
priority: 1
---
# Implement JWT token validation

Validate JWT tokens on all protected endpoints.

## Acceptance Criteria
- [ ] Middleware checks Authorization header
- [ ] Invalid tokens return 401
- [ ] Expired tokens return 401 with refresh hint
```

- File-based, git-friendly
- Tools: `ticket_create`, `ticket_start`, `ticket_close`, `ticket_assign`
- Daemon watches for changes, triggers agent actions

---

## Ambient & Proactive Agents

### Trigger Types

| Type | Configuration | Use Case |
|------|---------------|----------|
| **Cron** | `schedule: "0 9 * * MON"` | Daily standup, weekly reports |
| **File watch** | `watch: ~/Projects` | Auto-review on file changes |
| **Ticket** | (automatic) | Agent wakes when assigned ticket |

### Trigger Configuration

```yaml
# ~/.config/ayo/triggers/morning-standup.yaml
name: morning-standup
type: cron
schedule: "0 9 * * MON-FRI"
agent: "@standup"
prompt: |
  Review tickets from yesterday, summarize progress,
  identify blockers, and create today's priorities.
output: /output/standup/{date}.md
```

### Proactive Agent Patterns

**Pattern 1: Watcher Agent**
```yaml
name: code-reviewer
type: watch
watch:
  path: ~/Projects/myapp
  patterns: ["*.go", "*.py"]
  events: [modify]
agent: "@reviewer"
prompt: "Review the changed files for issues"
```

**Pattern 2: Scheduled Reporter**
```yaml
name: weekly-summary
type: cron
schedule: "0 17 * * FRI"
agent: "@analyst"
prompt: "Summarize this week's commits and PRs"
```

**Pattern 3: Ticket-Driven Worker**
```yaml
# No explicit trigger needed - daemon watches .tickets/
# When ticket assigned to @backend, agent wakes up
```

### Notification & Approval

Proactive agents may generate output that needs user review:

```
┌─────────────────────────────────────────────────────────────────────┐
│ ⏰ Scheduled: @standup completed morning standup                     │
│                                                                      │
│ Output: ~/.local/share/ayo/output/standup/2026-02-23.md             │
│                                                                      │
│ [V]iew  [O]pen in editor  [D]ismiss                                 │
└─────────────────────────────────────────────────────────────────────┘
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

## Documentation Simplification

### Current Docs (12,000+ lines across 20+ files)

Remove or consolidate:
- `docs/flows-spec.md` - Remove (YAML flows going away)
- `docs/plugins.md` - Simplify (most complexity removed)
- `docs/TUTORIAL.md` - Rewrite as concise getting-started
- `docs/reference/` - Auto-generate from CLI help

### Target Docs (~3,000 lines)

| Doc | Purpose | Lines |
|-----|---------|-------|
| `README.md` | Quick intro, install, hello world | ~200 |
| `docs/getting-started.md` | First 30 minutes with ayo | ~300 |
| `docs/agents.md` | Creating and configuring agents | ~400 |
| `docs/squads.md` | Multi-agent collaboration | ~500 |
| `docs/triggers.md` | Ambient/proactive agents | ~400 |
| `docs/planners.md` | Todos and tickets | ~300 |
| `docs/sandbox.md` | Understanding isolation | ~400 |
| `docs/cli-reference.md` | Command reference | ~500 |

---

## Implementation Phases

### Phase 1: Foundation (Simplification)

**Goal**: Remove complexity, establish clear mental model

1. Remove server, web UI, webhook code
2. Remove YAML flow executor
3. Simplify daemon to core functions
4. Standardize @ayo sandbox home directory
5. Implement host mount at /mnt/{username}

### Phase 2: File System Model

**Goal**: Clear, safe file access patterns

1. Implement file request workflow
2. Add /output safe write zone
3. Add approval UI to terminal
4. Implement "always allow for session" option

### Phase 3: Squad Polish

**Goal**: Squads as first-class coordination primitive

1. Clarify squad lead semantics
2. Implement squad dispatch routing
3. Add I/O schema enforcement
4. Polish ticket tools for agents

### Phase 4: Triggers & Ambient Agents

**Goal**: Proactive agents that act without prompts

1. Polish cron trigger configuration
2. Add file watch triggers
3. Implement notification system
4. Add trigger management CLI

### Phase 5: Documentation & Polish

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
2. "What's a squad?" → A team of agents working together
3. "What's a trigger?" → What makes an agent act without prompting
4. "What's a planner?" → How agents track their work

---

## Appendix: Ambient Agent Research

### Industry Patterns

Based on research of ambient/proactive agent systems:

**Key Design Patterns:**
1. **Trigger → Context → Action → Notification**
2. **Human-in-the-loop for destructive actions**
3. **Output artifacts over direct system changes**
4. **Audit trail for all autonomous actions**

**Trigger Mechanisms:**
- Cron/schedule (most common)
- File system events (powerful for dev workflows)
- Git hooks (CI/CD integration)
- API webhooks (external system integration)
- Manual "start watching" commands

**Lifecycle Management:**
- Background daemon manages agent lifecycle
- Agents are ephemeral - start, work, exit
- Persistent state in files/database
- Notification aggregation to avoid spam

**Approval Patterns:**
- Approve individual actions
- Approve classes of actions ("always allow read")
- Session-scoped approvals
- Time-bounded approvals

---

## Ticket Tracking

All implementation work is tracked in `.tickets/`. Use `tk list` to see current state.

### Epic Structure

| Epic | Phase | Dependencies | Status |
|------|-------|--------------|--------|
| `ayo-6h19` | Phase 1: Foundation | - | Open |
| `ayo-whmn` | Phase 2: File System | Phase 1 | Open |
| `ayo-xfu3` | Phase 3: Squad Polish | Phase 2 | Open |
| `ayo-sqad` | Phase 4: Triggers | Phase 3 | Open |
| `ayo-i2qo` | Phase 5: Documentation | Phase 4 | Open |

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
