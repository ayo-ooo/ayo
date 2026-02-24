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
- **Persistent memory**: Agents learn and adapt over time through semantic memory
- **Flexible triggers**: Time-based and event-based ambient agents
- **Experimentation-friendly**: Configurable guardrails, tools, and behaviors per agent

### The Vision: Ambient Agents

Imagine agents that don't wait to be asked. They watch your world—your inbox, your calendar, your codebase—and act when the moment is right:

- **Email triage**: An agent monitors your inbox, drafts replies to routine messages, flags urgent items, and summarizes your unread mail every morning
- **Calendar awareness**: An agent reviews tomorrow's meetings, prepares briefing docs, and reminds you of follow-ups from last week's calls
- **Project pulse**: An agent watches your GitHub repos, summarizes PR activity, alerts you to breaking CI, and tracks issues nearing their deadlines
- **Feed curation**: An agent monitors RSS feeds, Hacker News, and industry blogs, surfacing only what matches your interests
- **Chat presence**: Message your agent through Telegram or WhatsApp like you'd text a friend—ask questions, delegate tasks, get updates

These aren't scheduled cron jobs or dumb automations. They're **persistent, learning agents** that understand context, remember preferences, and improve over time. Ayo's trigger system, memory architecture, and chat integrations make this vision achievable today.

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
| **Memory system** | Good | Semantic memory with deduplication, scoping |
| **Capability search** | Good | Semantic agent/squad discovery |

### What's Confusing (Simplify/Remove)

| Component | Issue | Decision |
|-----------|-------|----------|
| **REST API server** | Overkill for CLI tool, confuses purpose | **REMOVE** |
| **Flows/Chains** | Shell scripts with extra steps | **SIMPLIFY** - keep DAG visualization only |
| **Web interface** | Incomplete, distracts from CLI focus | **REMOVE** |
| **YAML executor** | Overly complex flow definition | **REMOVE** |
| **Webhook server** | Premature integration point | **DEFER** |
| **Tunnel/QR code** | Mobile connectivity features | **REMOVE** |
| **IRC integration** | Abandoned experiment | **REMOVE** |
| **SQUAD.md frontmatter** | Inconsistent with agent config | **MIGRATE** to ayo.json |
| **Interactive TUI** | Janky, slow, over-engineered | **REWRITE** |

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
| **Memory as first-class** | Promote memory to core system primitive |
| **@ayo smart routing** | Intelligent agent vs squad selection |
| **Guardrails architecture** | Layered, configurable safety model |
| **Simpler interactive mode** | Streaming TUI without complexity |

---

## Memory System: First-Class Citizen

Memory is what makes ayo **grow and adapt** with the user over time. It's a key differentiator that must be treated as a core primitive, not an afterthought.

### Current Memory Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     MEMORY SYSTEM                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐     ┌─────────────────┐                   │
│  │  Memory Service │────▶│   SQLite Store  │ (vectors, search) │
│  └────────┬────────┘     └─────────────────┘                   │
│           │                                                     │
│           │              ┌─────────────────┐                   │
│           └─────────────▶│  Zettelkasten   │ (human-readable)  │
│                          └─────────────────┘                   │
│                                                                 │
│  Memory Categories:                                             │
│  • Preference - User preferences (tools, styles)               │
│  • Fact - Facts about user/project                              │
│  • Correction - User corrections to agent behavior              │
│  • Pattern - Observed behavioral patterns                       │
│                                                                 │
│  Memory Scopes:                                                 │
│  • global - All agents, all directories                        │
│  • agent - Specific agent only                                  │
│  • path - Specific project/directory                            │
│  • hybrid - Combines all (default)                              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Memory Design Principles

1. **Automatic formation**: Small LLM extracts memorable content without explicit commands
2. **Semantic deduplication**: Prevents redundant memories via embedding similarity
3. **Supersession chains**: Old memories linked to new ones, preserving history
4. **Non-blocking**: Memory storage is async, never blocks conversation
5. **Human-readable**: Zettelkasten markdown files alongside SQLite
6. **Local-only**: All data stored locally, no cloud sync (privacy)

### Memory Integration Points

| Integration | Description |
|-------------|-------------|
| **Session start** | Relevant memories injected into system prompt |
| **Tool calls** | Memory can inform tool behavior (e.g., preferred commands) |
| **Squad context** | Memories can be shared across squad agents |
| **Triggers** | Ambient agents can access/update memories |
| **Agent discovery** | Memory can influence agent selection |

### Memory Tickets (New)

- **`ayo-mem1`**: Expose memory as CLI commands (`ayo memory list/add/search`)
- **`ayo-mem2`**: Add memory tools for agents (`memory_store`, `memory_search`)
- **`ayo-mem3`**: Implement squad-scoped memories
- **`ayo-mem4`**: Add memory export/import for backup
- **`ayo-mem5`**: Document memory system in user guide
- **`ayo-zett`**: Zettelkasten tools and embedding-note linking

---

## Externalized Prompts: No Hardcoded Strings

**Critical principle**: Zero hardcoded prompts in the codebase. All prompts are sourced at runtime from `~/.local/share/ayo/prompts/`.

### Why This Matters

1. **Inspectable**: Users can see exactly what prompts are being used
2. **Customizable**: Edit any prompt without modifying source code
3. **Plugin-friendly**: Plugins can override prompts
4. **Harness experimentation**: Swap entire prompt sets for testing
5. **Transparency**: No hidden behaviors

### Prompt Directory Structure

```
~/.local/share/ayo/prompts/
├── system/
│   ├── base.md              # Base system prompt for all agents
│   ├── tool-usage.md        # How to use tools
│   ├── memory-usage.md      # How to use memory
│   └── planning.md          # Planning/reasoning guidance
├── guardrails/
│   ├── default.md           # Default safety guardrails
│   └── sandbox-aware.md     # Sandbox-specific rules
├── sandwich/
│   ├── prefix.md            # Conversation prefix
│   └── suffix.md            # Conversation suffix
├── agents/
│   └── @ayo/
│       └── system.md        # @ayo specific additions
├── errors/
│   ├── tool-failed.md       # Tool failure message template
│   └── permission-denied.md # Permission error template
└── templates/
    └── delegation.md        # Delegation instructions
```

### Runtime Loading

```go
// All prompt access goes through loader
prompts := prompts.NewLoader()  // Initialized once
systemPrompt := prompts.MustLoad("system/base.md")
guardrails := prompts.Load("guardrails/default.md")  // May fail gracefully
```

### CLI Commands

```bash
ayo prompt list                      # List all prompts
ayo prompt show system/base.md       # View a prompt
ayo prompt edit guardrails/default.md  # Edit in $EDITOR
ayo prompt reset system/base.md      # Reset to default
ayo prompt validate                  # Check for missing required prompts
```

### Plugin Prompt Overrides

Plugins can provide alternative prompts:

```json
{
  "name": "my-harness",
  "prompts": {
    "system/base.md": "prompts/my-base.md",
    "guardrails/default.md": "prompts/my-guardrails.md"
  }
}
```

Resolution order: User prompts → Plugin prompts → Defaults

### Ticket

- **`ayo-xprm`**: Externalized prompts system

---

## Interactive Mode: Simplified TUI

The current TUI is slow and complex. We need a simpler approach that handles streaming, tools, and multi-agent coordination elegantly.

### Current Problems

| Issue | Cause |
|-------|-------|
| **Slow rendering** | Full re-render on every tick (80ms), no incremental updates |
| **Complex architecture** | Triple abstraction for tool rendering, registry + interface + wrappers |
| **Animation overhead** | ID-scoped ticks to prevent race conditions indicate fragility |
| **Message duplication** | Complex pending/rendered tracking sets |

### New Interactive Mode Design

**Principle**: Streaming text with inline status, not complex viewport management.

```
┌─────────────────────────────────────────────────────────────────┐
│ @ayo                                                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ You: Fix the bug in auth.go                                     │
│                                                                 │
│ @ayo: I'll look at the authentication code.                    │
│                                                                 │
│   ▸ bash: grep -n "auth" internal/auth/*.go                    │
│     └─ 23 matches in 4 files                                   │
│                                                                 │
│   ▸ view: internal/auth/handler.go:45-80                       │
│     └─ Read 35 lines                                           │
│                                                                 │
│   ▸ edit: internal/auth/handler.go                             │
│     └─ Fixed nil pointer check                                 │
│                                                                 │
│ The bug was a missing nil check on line 52. Fixed.             │
│                                                                 │
│ ─────────────────────────────────────────────────────────────── │
│ > _                                                             │
└─────────────────────────────────────────────────────────────────┘
```

### Key Simplifications

1. **Stream-first**: Text streams directly, no buffering entire responses
2. **Inline tools**: Tool calls appear inline with chevron prefix, collapse when done
3. **No viewport**: Just terminal scrollback, user scrolls naturally
4. **Simple focus**: Only the input prompt is interactive
5. **No sidebar**: Todos/tickets shown inline or via commands

### Components to Keep

- `glamour` for markdown rendering
- `lipgloss` for styling
- `bubbles/textarea` for input

### Components to Remove/Simplify

- Complex `Model` with viewport management
- `EventAggregator` (use direct channel read)
- `ToolCallTree` (use simple list)
- Sidebar panels
- Tick-based animations

### New Files

- `internal/ui/interactive/interactive.go` - Simple streaming chat
- `internal/ui/interactive/toolcall.go` - Inline tool rendering
- `internal/ui/interactive/input.go` - Input handling

---

## Guardrails Architecture: Layered Safety

Guardrails are implemented through infrastructure and externalized prompts—no hardcoded prompt strings in the codebase.

### Current Guardrails

| Layer | Component | Type |
|-------|-----------|------|
| Prompt | LegacyGuardrails (6 rules) | **Hardcoded - REMOVE** |
| Prompt | Sandwich pattern (prefix/suffix) | Configurable files |
| Input | AdversarialPatterns (17 regexes) | **Hardcoded - EXTERNALIZE** |
| Agent | Trust levels | Configurable |
| Runtime | Sandbox isolation | Infrastructure |

### Proposed Layered Model

```
┌─────────────────────────────────────────────────────────────────┐
│                     GUARDRAILS LAYERS                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  L1: INFRASTRUCTURE (Sandbox)                                   │
│  ├─ Filesystem isolation (read-only host mount)                │
│  ├─ Network controls (per-agent)                                │
│  ├─ Process isolation (Unix users)                              │
│  └─ Resource limits (optional)                                  │
│                                                                 │
│  L2: PROTOCOL (Host Daemon)                                     │
│  ├─ file_request approval flow                                  │
│  ├─ Audit logging                                               │
│  └─ Adversarial input detection (pattern file, not hardcoded)   │
│                                                                 │
│  L3: PROMPT (Externalized Files)                                │
│  ├─ All prompts in ~/.local/share/ayo/prompts/                  │
│  ├─ Sandwich prefix/suffix files                                │
│  ├─ Per-agent system.md                                         │
│  └─ Squad constitution (SQUAD.md)                               │
│                                                                 │
│  L4: BEHAVIORAL (Optional)                                      │
│  ├─ Output filters (PII, secrets)                               │
│  ├─ Rate limiting                                               │
│  └─ Human-in-the-loop for specific actions                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Guardrail Configuration (Simplified)

No abstract "levels" - just explicit boolean flags:

```json
// ayo.json for agent
{
  "agent": {
    "sandbox": {
      "network": true,
      "filesystem": "readonly"  // "readonly", "workspace", "full"
    },
    "prompts": {
      "prefix": "sandwich/prefix.md",   // Relative to prompts dir
      "suffix": "sandwich/suffix.md"
    },
    "permissions": {
      "auto_approve": false,
      "allowed_paths": ["~/Projects/*"],
      "denied_paths": ["~/.ssh", "~/.aws"]
    }
  }
}
```

### What Moves Where

| Current Guardrail | New Layer | Rationale |
|-------------------|-----------|-----------|
| "Stay in scope" | L1 (Sandbox) | Filesystem isolation handles this |
| "Confirm destructive actions" | L2 (Protocol) | file_request flow |
| Network restrictions | L1 (Sandbox) | Container network controls |
| "No credential exposure" | L1+L3 | Sandbox env + prompt reminder |
| Adversarial detection | L2 (Protocol) | Pre-LLM input filtering |
| "No malicious code" | L3 (Prompt) | Behavioral guidance |
| Trust levels | L2 (Protocol) | Enforcement before execution |

### Experimentation Support

For users experimenting with agent harnesses:
- `guardrails.enabled: false` disables L3/L4 entirely
- L1/L2 always active (infrastructure safety)
- Custom sandwich files for prompt experimentation
- Per-agent override of any setting

---

## @ayo Smart Routing: Agent vs Squad Selection

@ayo needs to intelligently decide whether to handle a task itself, delegate to an agent, or dispatch to a squad.

### Current State

The codebase has:
- `Dispatcher` - Semantic routing via embeddings
- `find_agent` tool - Agent discovery for @ayo
- `UnifiedSearcher` - Indexes agents/squads with embeddings
- Static delegation config - Task type → agent mapping

### Decision Flow

```
User: "Build the auth feature"
         │
         ▼
┌─────────────────────────────────────────────────────────────────┐
│ @AYO ROUTING DECISION                                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Is this trivial? (<100 chars, greeting, clarification)     │
│     └─▶ Handle directly                                        │
│                                                                 │
│  2. Does user explicitly target? (@agent or #squad)            │
│     └─▶ Route to target                                        │
│                                                                 │
│  3. Is there a matching squad? (semantic search)               │
│     └─▶ Dispatch to squad (multi-agent coordination)          │
│                                                                 │
│  4. Is there a specialist agent? (semantic search)             │
│     └─▶ Invoke agent directly                                  │
│                                                                 │
│  5. Can @ayo handle it?                                        │
│     └─▶ Handle directly with available tools                   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Key Principle

> If a task needs multiple agents collaborating, use a squad.
> If a single agent can do it, invoke that agent directly.
> Squads are for coordination, not for single-agent tasks.

### Routing Signals

| Signal | Weight | Notes |
|--------|--------|-------|
| Explicit targeting | Highest | User said `@agent` or `#squad` |
| Squad mission match | High | Task matches squad SQUAD.md mission |
| Agent description match | Medium | Task matches agent description |
| Task complexity | Medium | Multi-file, multi-concern → squad |
| Memory hints | Low | Previous similar tasks routed somewhere |

### Tickets (New)

- **`ayo-rout`**: Implement @ayo routing decision logic
- Document routing in agent guide

---

## Sandbox Architecture: Agent Coexistence Model

### The Key Question

> Should each agent get its own sandbox, or do agents coexist in a shared sandbox?

### Decision: Shared Default Sandbox with Optional Isolation

**Default behavior**: All agents invoked directly (`ayo @agent "prompt"`) execute in the **@ayo sandbox**. This is the "workbench" where you interact with agents.

**Squads**: Get their own isolated sandbox where multiple agents collaborate.

**Explicit isolation**: Agents can request their own sandbox via config.

```
┌─────────────────────────────────────────────────────────────────┐
│                         SANDBOX LANDSCAPE                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────────────────────────────────┐              │
│  │           @AYO SANDBOX (default)             │              │
│  │  /home/ayo/          - @ayo's home           │              │
│  │  /home/crush/        - @crush's home         │              │
│  │  /home/reviewer/     - @reviewer's home      │              │
│  │  /mnt/{user}/        - Host home (read-only) │              │
│  │  /workspace/         - Shared workspace      │              │
│  │  /output/            - Safe write zone       │              │
│  │                                              │              │
│  │  When you run: ayo @crush "write code"       │              │
│  │  → @crush executes HERE, in @ayo sandbox     │              │
│  └──────────────────────────────────────────────┘              │
│                                                                 │
│  ┌──────────────────────────────────────────────┐              │
│  │           #dev-team SQUAD SANDBOX            │              │
│  │  /home/frontend/     - @frontend's home      │              │
│  │  /home/backend/      - @backend's home       │              │
│  │  /workspace/         - Shared code           │              │
│  │  /.tickets/          - Coordination          │              │
│  │                                              │              │
│  │  When you run: ayo #dev-team "build feature" │              │
│  │  → Squad lead orchestrates HERE              │              │
│  └──────────────────────────────────────────────┘              │
│                                                                 │
│  ┌──────────────────────────────────────────────┐              │
│  │     @isolated-agent SANDBOX (if configured)  │              │
│  │  sandbox: { isolated: true } in ayo.json     │              │
│  └──────────────────────────────────────────────┘              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
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

To simplify sandbox management, we install a lightweight **ayod** (ayo daemon) inside each sandbox. This replaces ad-hoc `container exec` calls with a clean, extensible service.

### Bootstrap Sequence (No `sleep infinity`)

The key insight is that we can use a **staged image build** instead of runtime injection:

```
┌─────────────────────────────────────────────────────────────────┐
│ AYOD BOOTSTRAP (Image Build Time)                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. ayo builds derived image from base (alpine:3.21)           │
│     ┌─────────────────────────────────────────────────┐        │
│     │ FROM alpine:3.21                                │        │
│     │ COPY ayod /usr/local/bin/ayod                   │        │
│     │ ENTRYPOINT ["/usr/local/bin/ayod"]              │        │
│     └─────────────────────────────────────────────────┘        │
│                                                                 │
│  2. Derived image stored in ~/.local/share/ayo/images/         │
│                                                                 │
│  3. Container created from derived image                        │
│     → ayod starts as PID 1 immediately                         │
│     → No intermediate "sleep infinity" state                    │
│                                                                 │
│  4. Host daemon connects to /run/ayod.sock (mounted)           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Alternative: Runtime Socket Connection

If image building isn't feasible (e.g., Apple Container doesn't support custom images):

```
1. Start container with base image + mounted socket path
2. Host injects ayod via `container copy` (before any exec)
3. Host sends "start ayod" command via initial exec
4. ayod daemonizes, opens socket
5. All subsequent operations go through ayod socket

This is a single "bootstrap exec" vs the current per-operation exec calls.
```

### What ayod Does

| Function | RPC Method | Description |
|----------|------------|-------------|
| User management | `UserAdd` | Create agent user with home directory |
| Command execution | `Exec` | Run command as specified user, stream output |
| File operations | `ReadFile`, `WriteFile` | Direct file access (no shell) |
| File request proxy | `FileRequest` | Proxy to host for approval |
| Health check | `Health` | Report sandbox status |
| Process list | `Processes` | List running processes per user |

### ayod Core Implementation

```go
// cmd/ayod/main.go
package main

import (
    "net"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // Create socket
    listener, _ := net.Listen("unix", "/run/ayod.sock")
    os.Chmod("/run/ayod.sock", 0666) // Allow all sandbox users
    
    // Initialize user manager
    users := NewUserManager()
    
    // Handle signals for clean shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
    
    // Start RPC server
    server := NewServer(users)
    go server.Serve(listener)
    
    // Stay alive as PID 1
    <-sigCh
    server.Shutdown()
}
```

### Benefits

1. **Single entry point**: All sandbox operations go through ayod
2. **No shell overhead**: Direct exec vs `sh -c` wrapper
3. **Streaming output**: Real-time stdout/stderr without buffering
4. **Extensible**: Easy to add new capabilities
5. **Consistent**: Same behavior across Apple Container and systemd-nspawn
6. **Debuggable**: `ayo sandbox shell` connects to ayod directly

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
      "network": true,
      "filesystem": "readonly"
    },
    "permissions": {
      "auto_approve": false,
      "allowed_paths": ["~/Projects/*"],
      "denied_paths": ["~/.ssh", "~/.aws"]
    },
    "prompts": {
      "prefix": "sandwich/prefix.md",
      "suffix": "sandwich/suffix.md"
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
    "memory": {
      "shared": true,
      "scope": "squad"
    }
  }
}
```

### SQUAD.md is the Orchestrator's System Prompt

SQUAD.md is the **system prompt** for the squad's invisible orchestration agent. It defines mission, roles, and coordination rules that the orchestrator uses when dispatching work to squad agents.

**Configuration** moves to `ayo.json`. **Instructions** stay in `SQUAD.md`:

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
Agents should check in with @architect before major decisions.
Never commit directly to main - all changes go through PR review.
```

This is human-readable AND human-editable - users can tune squad behavior by editing SQUAD.md.

### Migration Path

No migration needed - destroy existing configs and start fresh with ayo.json.

---

## Flow Visualization (Optional Enhancement)

While we're removing the YAML executor, flow visualization could still be valuable for understanding agent execution graphs.

### Potential Approach

Use a cell-based terminal visualization library to render DAGs:

```
┌─────────────────────────────────────────────────────────────────┐
│ Flow: build-and-deploy                                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────┐     ┌─────────┐     ┌─────────┐                   │
│  │  lint   │────▶│  test   │────▶│  build  │                   │
│  └─────────┘     └─────────┘     └────┬────┘                   │
│       │                               │                         │
│       │         ┌─────────┐           │                         │
│       └────────▶│ typecheck│──────────┤                         │
│                 └─────────┘           │                         │
│                                       ▼                         │
│                                ┌─────────┐                      │
│                                │ deploy  │                      │
│                                └─────────┘                      │
│                                                                 │
│  Legend: ✓ completed  ◉ running  ○ pending  ✗ failed           │
└─────────────────────────────────────────────────────────────────┘
```

### Implementation Notes

- Could use `lipgloss` for box drawing
- Would need DAG layout algorithm (already have topological sort)
- Nice-to-have, not required for GTM (create ticket post-GTM if desired)

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
| IRC-related code | Abandoned | ~200 |

**Estimated removal: ~5,000+ lines**

### Files to Simplify

| Path | Change |
|------|--------|
| `internal/flows/` | Keep only discover, parse, DAG inspection |
| `internal/daemon/server.go` | Remove webhook, serve endpoints |
| `cmd/ayo/flows.go` | Remove run/execute, keep inspect |
| `internal/ui/chat/` | Replace with simpler interactive mode |

### Dead Code to Remove

| Item | Location |
|------|----------|
| Deprecated `NewFantasyToolSetWithOptions` | `internal/run/fantasy_tools.go` |
| Deprecated `TruncateWithEllipsis` | `internal/ui/styles.go` |
| Deprecated `FormatDuration`, `TruncateText` | `internal/ui/shared/toolformat.go` |
| Empty `handleTicketClosed` | `internal/daemon/server.go` |
| Duplicate `AgentInvoker` interface | 3 locations |
| Custom `trimWhitespace` | `internal/flows/yaml_executor.go` |
| Duplicate `Todo` types | `internal/run/todo.go`, `internal/ui/todo.go` |

---

## Technical Debt Cleanup

### Modernization (gopls hints)

| Pattern | Count | Fix |
|---------|-------|-----|
| `slicescontains` | 7 | Use `slices.Contains` |
| `omitzero` | 5 | Fix or remove `omitempty` on nested structs |
| `mapsloop` | 1 | Replace loop with `maps.Copy` |
| `stringscutprefix` | 1 | Use `CutPrefix` |
| `stringsseq` | 1 | Use `SplitSeq` |
| `fmtappendf` | 1 | Use `fmt.Appendf` |

### Interface Consolidation

Extract shared interfaces to `internal/interfaces/`:
- `AgentInvoker` (currently in 3 packages)
- `ToolRenderer` (currently split between ui/shared and ui/chat)

### Large File Splitting

| File | Lines | Split Into |
|------|-------|------------|
| `internal/agent/agent.go` | ~1100 | `agent.go`, `config.go`, `loading.go`, `memory.go` |
| `internal/daemon/server.go` | ~1100 | `server.go`, `rpc_handlers.go`, `squad_handlers.go` |

---

## Implementation Phases

### Phase 1: Foundation (Simplification)

**Goal**: Remove complexity, establish clear mental model

1. Remove server, web UI, webhook code
2. Remove YAML flow executor
3. Remove IRC-related dead code
4. Simplify daemon to core functions
5. Implement ayod (in-sandbox daemon)
6. Implement shared sandbox with per-agent homes
7. Implement host mount at /mnt/{username}
8. Clean up deprecated functions and dead code
9. **Externalize all prompts** (`ayo-xprm`) - zero hardcoded prompt strings

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
4. Update CLI commands for new schema

### Phase 4: Advanced Scheduler

**Goal**: Powerful, persistent scheduling with gocron v2

1. Replace robfig/cron with gocron v2
2. Implement job persistence in SQLite
3. Add one-time and duration jobs
4. Implement job monitoring
5. Add trigger CLI improvements

### Phase 5: Squad & Routing Polish

**Goal**: Squads and @ayo routing as first-class primitives

1. Clarify squad lead semantics
2. Implement squad dispatch routing
3. Add I/O schema enforcement
4. Polish ticket tools for agents
5. Implement @ayo smart routing

### Phase 6: Memory & Interactive Mode

**Goal**: Memory as first-class citizen, simpler TUI

1. Add memory CLI commands (`ayo-mem1`)
2. Add memory tools for agents (`ayo-mem2`)
3. Implement Zettelkasten note linking (`ayo-zett`)
4. Implement squad-scoped memories (`ayo-mem3`)
5. Rewrite interactive mode with event rendering (`ayo-evnt`)
6. Optimize glamour initialization (`ayo-glam`)
7. Document memory system (`ayo-mem5`)

### Phase 7: CLI Polish

**Goal**: Improve CLI experience

1. Polish CLI help text
2. Add `ayo doctor` improvements
3. Minor UX improvements

Each phase includes an **E2E verification ticket** that must pass before the phase is considered complete.

### Phase 8: Plugin Ecosystem

**Goal**: Extensible plugin system for integrations and custom components

1. Complete external planner loading (Go plugin .so)
2. Add squad plugin support (packageable squad definitions)
3. Implement trigger plugin architecture (pluggable trigger types)
4. Add sandbox config plugins (alternative harnesses)
5. Improve plugin registry (discovery, dependencies, validation)

### Phase 9: Documentation

**Goal**: Comprehensive, accurate documentation written after all features are implemented

Documentation is written **after** all code phases (1-8) are complete to ensure accuracy.

**Documentation Structure**:
```
README.md                    # Project overview, quick start
docs/
├── getting-started.md       # Installation, first agent, 5-min tutorial
├── concepts.md              # Core concepts (agents, squads, sandboxes, memory)
├── tutorials/
│   ├── first-agent.md       # Create your first agent
│   ├── squads.md            # Multi-agent coordination
│   ├── triggers.md          # Event-driven agents
│   ├── memory.md            # Memory system deep dive
│   └── plugins.md           # Creating plugins
├── guides/
│   ├── agents.md            # Agent configuration guide
│   ├── squads.md            # Squad configuration guide
│   ├── triggers.md          # Trigger configuration guide
│   ├── tools.md             # Built-in tools reference
│   ├── sandbox.md           # Sandbox architecture
│   └── security.md          # Security model and guardrails
├── reference/
│   ├── cli.md               # CLI command reference
│   ├── ayo-json.md          # ayo.json schema reference
│   ├── prompts.md           # Externalized prompts reference
│   ├── rpc.md               # Daemon RPC reference
│   └── plugins.md           # Plugin manifest reference
└── advanced/
    ├── architecture.md      # System architecture deep dive
    ├── extending.md         # Extending ayo
    └── troubleshooting.md   # Common issues and debugging
```

**Documentation Tickets**:
- `ayo-doc1`: Code analysis - document existing behavior
- `ayo-doc2`: Write README.md and getting-started.md
- `ayo-doc3`: Write concepts.md (core concepts)
- `ayo-doc4`: Write tutorials (5 guides)
- `ayo-doc5`: Write configuration guides (6 guides)
- `ayo-doc6`: Write reference documentation (5 docs)
- `ayo-doc7`: Write advanced documentation (3 docs)
- `ayo-doc8`: Accuracy verification - test all examples
- `ayo-docv`: Phase 9 E2E verification

---

## Plugin System Architecture

The plugin system enables ayo to be extended with new capabilities without modifying core code. Plugins can provide multiple component types, making them powerful integration packages.

### Current Plugin Capabilities

| Component | Description | Example |
|-----------|-------------|---------|
| `agents` | Agent definitions | @crush, @reviewer |
| `skills` | Instruction files | code-review.md |
| `tools` | External commands | searxng search |
| `delegates` | Task routing | "coding" → @crush |
| `default_tools` | Tool aliases | "search" → "searxng" |
| `providers` | System providers | memory, sandbox |
| `planners` | Planning tools | ayo-todos, ayo-tickets |

### New Plugin Capabilities (Phase 8)

| Component | Description | Ticket |
|-----------|-------------|--------|
| `squads` | Reusable squad definitions | `ayo-plsq` |
| `triggers` | Custom trigger types | `ayo-pltg` |
| `sandbox_configs` | Alternative container setups | `ayo-plsb` |

### Multi-Component Plugin Example

A single plugin can provide agents, tools, triggers, and skills together:

```json
{
  "name": "ayo-plugins-imap",
  "version": "1.0.0",
  "components": {
    "triggers": {
      "imap": {
        "path": "triggers/imap",
        "description": "IMAP email trigger"
      }
    },
    "tools": {
      "email-send": { "path": "tools/email-send" },
      "email-search": { "path": "tools/email-search" }
    },
    "agents": {
      "@email-handler": { "path": "agents/email-handler" }
    },
    "skills": {
      "email-triage": { "path": "skills/email-triage.md" }
    }
  }
}
```

### Trigger Plugin Architecture

Triggers are now pluggable via a standardized interface:

```go
type TriggerPlugin interface {
    Name() string
    Type() TriggerType  // "poll", "push", "watch"
    Init(ctx context.Context, config map[string]any) error
    Start(ctx context.Context, callback EventCallback) error
    Stop() error
}
```

External triggers communicate via JSON over stdin/stdout for cross-platform compatibility.

### Delegate System (Kept)

The delegate system remains valuable for task routing:
- Maps task types to specialist agents
- Enables @ayo to intelligently dispatch
- Configured in plugin manifests
- Example: `"coding" → @crush`, `"research" → @researcher`

### Planned Plugin Repositories

**Open Standards Focus**: Plugin ecosystem prioritizes open protocols over proprietary services.

| Plugin | Priority | Protocol/Standard | Components | Ticket |
|--------|----------|-------------------|------------|--------|
| `ayo-plugins-imap` | High | IMAP/SMTP (RFC 3501) | Trigger + Tools + Agent | `ayo-pimap` |
| `ayo-plugins-caldav` | Medium | CalDAV (RFC 4791) | Trigger + Tools + Agent | `ayo-pcal` |
| `ayo-plugins-rss` | Medium | RSS 2.0 / Atom (RFC 4287) | Trigger + Tools + Agent | `ayo-prss` |
| `ayo-plugins-xmpp` | Low | XMPP (RFC 6120) | Trigger + Tools + Agent | `ayo-pxmpp` |
| `ayo-plugins-webhook` | High | HTTP Webhooks | Trigger + Tools | `ayo-pwhk` |

**Chat Integrations** (via trigger plugins):

| Plugin | Priority | Library | Components | Ticket |
|--------|----------|---------|------------|--------|
| `ayo-plugins-telegram` | High | go-telegram-bot-api | Trigger + Tools + Agent | `ayo-ptgram` |
| `ayo-plugins-whatsapp` | High | whatsmeow | Trigger + Tools + Agent | `ayo-pwhats` |
| `ayo-plugins-matrix` | Medium | mautrix-go | Trigger + Tools + Agent | `ayo-pmatrix` |

Chat plugins enable **conversational ambient agents** - users message their agents through familiar chat apps like they'd message a friend. Your agent monitors your inbox, watches your calendar, tracks your projects, and proactively reaches out when it has something useful to say. This transforms ayo from a CLI tool into a persistent digital assistant that lives in your messaging apps.

### MCP Ecosystem Inspiration

Analysis of 500+ MCP servers revealed these popular integration categories:

1. **Communication** (IMAP, Slack, Discord) - High demand for messaging triggers
2. **Productivity** (Calendar, Tasks) - Meeting prep, reminders
3. **Development** (GitHub, Linear) - Code review, issue triage
4. **Data** (Databases, APIs) - Data-driven automation

Key insight: Users want **event-driven agents** that respond to real-world triggers, not just CLI invocation.

---

## Success Criteria

### For GTM Readiness

- [ ] New user can install and run first agent in < 5 minutes
- [ ] Documentation explains all concepts clearly
- [ ] `ayo doctor` catches all setup issues
- [ ] Squads work reliably for multi-agent tasks
- [ ] Triggers enable ambient agent use cases
- [ ] Memory system is discoverable and documented
- [ ] No dead code or unused features
- [ ] Test coverage > 70%

### Mental Model Test

A user should be able to answer:
1. "What is ayo?" → CLI for managing AI agents in sandboxes
2. "Where do agents run?" → In the @ayo sandbox, or isolated squad sandboxes
3. "What's a squad?" → A team of agents with their own sandbox
4. "What's a trigger?" → What makes an agent act without prompting
5. "What's --no-jodas?" → Auto-approve mode for power users
6. "How do agents learn?" → Through the memory system that persists across sessions

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
| `ayo-memx` | Phase 6: Memory & Interactive | Phase 5 | Open |
| `ayo-i2qo` | Phase 7: CLI Polish | Phase 6 | Open |
| `ayo-plug` | Phase 8: Plugin Ecosystem | Phase 5 | Open |
| `ayo-docs` | Phase 9: Documentation | Phase 7, 8 | Open |

### Plugin Repository Epics (Open Standards)

| Epic | Protocol/Standard | Priority | Status |
|------|-------------------|----------|--------|
| `ayo-pimap` | IMAP/SMTP (RFC 3501) | High | Open |
| `ayo-pwhk` | HTTP Webhooks | High | Open |
| `ayo-pcal` | CalDAV (RFC 4791) | Medium | Open |
| `ayo-prss` | RSS/Atom (RFC 4287) | Medium | Open |
| `ayo-pxmpp` | XMPP (RFC 6120) | Low | Open |

### Chat Plugin Epics

| Epic | Library | Priority | Status |
|------|---------|----------|--------|
| `ayo-ptgram` | go-telegram-bot-api | High | Open |
| `ayo-pwhats` | whatsmeow | High | Open |
| `ayo-pmatrix` | mautrix-go | Medium | Open |

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

## Clean Slate Preparation

Before beginning implementation, clean up any existing state:

```bash
# Stop any running daemons
ayo daemon stop

# Kill any running sandboxes
ayo sandbox list | xargs -I{} ayo sandbox destroy {}

# Remove all local state
rm -rf ~/.local/share/ayo
rm -rf ~/.config/ayo

# Remove any launchd/systemd services
launchctl unload ~/Library/LaunchAgents/land.charm.ayod.plist 2>/dev/null
systemctl --user disable ayo-daemon 2>/dev/null

# Fresh start
ayo doctor
```

---

*Last updated: 2026-02-23*
