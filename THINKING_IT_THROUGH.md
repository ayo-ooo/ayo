# Thinking It Through: Sandbox-Based Agent Teams with Per-Sandbox @ayo Orchestrators

This document analyzes a proposed architectural change to how ayo manages sandboxes and agent teams. The goal is to help you understand the implications before committing to this direction.

---

## Your Idea (As I Understand It)

You want to:

1. **Create different sets of agents for different sandboxes** — each sandbox would have its own "team" of agents that work together
2. **Make @ayo exist independently in each sandbox** — rather than being a global orchestrator, each sandbox gets its own @ayo instance with its own rules and related agents
3. **Use sandboxes as team boundaries** — a sandbox becomes a container for a cohesive team of agents that collaborate
4. **Still have @ayo orchestrate across all sandboxes** — @ayo remains the universal orchestrator, but now it orchestrates *teams* rather than individual agents

The mental model seems to be:

```
┌─────────────────────────────────────────────────────────────────┐
│                        HOST / DAEMON                             │
│                                                                  │
│   @ayo (global) — orchestrates teams, not individual agents     │
│                                                                  │
│   ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│   │   Sandbox A     │  │   Sandbox B     │  │   Sandbox C     │ │
│   │   "frontend"    │  │   "backend"     │  │   "devops"      │ │
│   │                 │  │                 │  │                 │ │
│   │  @ayo (local)   │  │  @ayo (local)   │  │  @ayo (local)   │ │
│   │  @designer      │  │  @api-dev       │  │  @terraform     │ │
│   │  @react-dev     │  │  @db-admin      │  │  @docker        │ │
│   │  @css-wizard    │  │  @tester        │  │  @k8s           │ │
│   │                 │  │                 │  │                 │ │
│   │  Shared state   │  │  Shared state   │  │  Shared state   │ │
│   │  Team tickets   │  │  Team tickets   │  │  Team tickets   │ │
│   └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

---

## Where We Stand Now

### Current Sandbox Architecture

**Sandboxes are transient execution environments, not team boundaries.**

| Aspect | Current Behavior |
|--------|------------------|
| Creation | Pool of warm sandboxes, acquired per-agent-invocation |
| Lifecycle | Agent runs → releases sandbox → sandbox returns to pool or is destroyed |
| Agent association | Temporary — agent uses sandbox for one task, then releases it |
| @ayo | Single global instance, runs on host (or in any available sandbox) |
| Team concept | None explicit — only "collaboration groups" for sharing |

**Key mechanisms:**

1. **Sandbox Pool** (`internal/sandbox/pool.go`)
   - Maintains 1-4 warm sandboxes
   - Agents acquire/release sandboxes dynamically
   - `AcquireOptions.Group` allows agents to share a sandbox temporarily
   - No persistent team assignment

2. **Collaboration Groups** (existing but limited)
   - Agents can request same sandbox via `Group` parameter
   - Multiple agents tracked in `sandbox.Agents []string`
   - But: groups are ad-hoc, not configured entities

3. **@ayo Agent** (`internal/builtin/agents/@ayo/`)
   - Single definition in builtin agents
   - Global scope, global memory
   - Delegates to other agents via `agent_call` tool
   - Sees all agents except `unrestricted` ones

### What Works Well

1. **Isolation** — Sandboxes provide strong filesystem/process isolation
2. **Resource management** — Pool handles container lifecycle efficiently
3. **Flexibility** — Any agent can use any sandbox from the pool
4. **Simplicity** — One @ayo, one source of truth for orchestration

### What Falls Short

1. **No team boundaries** — Agents are individuals, not team members
2. **No persistent sandbox identity** — Sandboxes are interchangeable
3. **Global @ayo can't specialize** — Same orchestrator for all contexts
4. **No team-local state** — No way to have team-specific memory/tickets
5. **Collaboration is manual** — Agents must coordinate via tickets/files, no structural support

---

## Implications of Your Proposal

### Architectural Changes Required

#### 1. Sandbox Identity and Persistence

**Current:** Sandboxes are anonymous, pooled, transient.
**Proposed:** Sandboxes are named, persistent, team-scoped.

```go
// Current
type PoolEntry struct {
    Sandbox providers.Sandbox
    Agents  []string  // temporary list
}

// Proposed
type TeamSandbox struct {
    ID          string
    Name        string              // "frontend", "backend"
    Agents      []AgentConfig       // permanent team roster
    LocalAyo    *AgentConfig        // team-specific @ayo
    Tickets     *tickets.Service    // team-scoped tickets
    Memory      *memory.Store       // team-scoped memory
    Persistent  bool                // survives daemon restart
}
```

**Impact:** Major rework of sandbox lifecycle. Pooling becomes team management.

#### 2. @ayo Becomes Multi-Instance

**Current:** One @ayo process/definition globally.
**Proposed:** One global @ayo + one local @ayo per sandbox.

Questions to resolve:
- How does global @ayo communicate with local @ayos?
- Do local @ayos have different system prompts?
- Can local @ayo override global @ayo decisions?
- What happens when prompts conflict?

```
Global @ayo: "Deploy the feature"
    └─> Routes to "backend" team
        └─> Local @ayo (backend): "I'll coordinate @api-dev and @tester"
            └─> @api-dev: works on API
            └─> @tester: writes tests
```

**Impact:** Requires defining @ayo hierarchy, communication protocol, authority boundaries.

#### 3. Agent Registration Becomes Team-Scoped

**Current:** Agents registered globally in `~/.config/ayo/agents/`.
**Proposed:** Agents can be global OR team-scoped.

```
~/.config/ayo/
├── agents/              # Global agents
│   └── @researcher/
└── teams/
    ├── frontend/
    │   ├── config.json  # Team config including local @ayo rules
    │   └── agents/
    │       ├── @react-dev/
    │       └── @designer/
    └── backend/
        ├── config.json
        └── agents/
            ├── @api-dev/
            └── @db-admin/
```

**Impact:** New directory structure, agent resolution order, team config format.

#### 4. Tickets Become Team-Scoped (or Cross-Team)

**Current:** Tickets in session directories, global namespace.
**Proposed:** Team-local tickets + cross-team tickets for handoffs.

```
~/.local/share/ayo/
├── sessions/{id}/.tickets/     # Session tickets (current)
└── teams/
    ├── frontend/.tickets/      # Frontend team tickets
    ├── backend/.tickets/       # Backend team tickets
    └── global/.tickets/        # Cross-team coordination
```

**Impact:** Ticket routing, visibility rules, dependency resolution across teams.

#### 5. Memory Becomes Team-Scoped

**Current:** Memory scoped to agent or global.
**Proposed:** Add team scope.

| Scope | Visibility |
|-------|------------|
| Agent | Only that agent |
| Team | All agents in team's sandbox |
| Global | All agents everywhere |

**Impact:** Memory retrieval needs team context, embedding search adds team filter.

---

## Things to Consider Before Proceeding

### 1. What Problem Are You Solving?

Be specific about the pain point:

- **"Agents step on each other"** → Teams provide clear boundaries
- **"@ayo doesn't understand context"** → Local @ayo can specialize
- **"I want isolated projects"** → Teams = projects?
- **"I need different rules for different work"** → Local @ayo has team-specific prompts

If the problem is simpler, a simpler solution might work:
- **Collaboration groups** (existing) for ad-hoc sharing
- **Project-level .ayo.json** for context-specific delegation
- **Tickets with assignees** for work distribution

### 2. One @ayo or Many?

This is the core philosophical question.

**Single @ayo (current):**
- Simple mental model: "ask @ayo anything"
- One source of truth for delegation
- One personality, one set of rules
- Easier to maintain/debug

**Multiple @ayos:**
- Specialized for different domains
- Team autonomy: backend team's @ayo knows backend patterns
- Complexity: which @ayo handles a request?
- Potential conflicts: two @ayos disagree

**Hybrid (your proposal):**
- Global @ayo routes to teams
- Local @ayos handle team-internal work
- Requires clear authority hierarchy
- Global @ayo becomes "manager of managers"

### 3. How Do Teams Communicate?

If sandboxes become team boundaries, cross-team work needs explicit mechanisms:

| Mechanism | Pros | Cons |
|-----------|------|------|
| Cross-team tickets | Explicit handoffs, audit trail | Latency, overhead |
| Shared files in /shared | Simple | No structure, race conditions |
| Message passing | Real-time | Complexity (Matrix v2?) |
| Global @ayo mediates | Clear authority | Bottleneck |

### 4. Sandbox Lifecycle Changes

**Current:** Sandboxes are cheap, disposable, pooled.
**Proposed:** Sandboxes are persistent team environments.

Implications:
- **Startup time:** First team invocation creates sandbox (slow)
- **Resource usage:** N teams = N running containers
- **State management:** Team sandbox has persistent state
- **Failure modes:** If team sandbox dies, whole team is impacted

### 5. Agent Discovery and Routing

**Current:** All agents visible globally (except unrestricted).
**Proposed:** Some agents team-local, some global.

When user says `ayo @react-dev "do something"`:
1. Which sandbox does @react-dev live in?
2. Does @react-dev exist only in "frontend" team?
3. Can you call a team agent from outside its team?
4. Does calling @react-dev automatically route to frontend sandbox?

### 6. Backwards Compatibility

Current users expect:
- `ayo` talks to global @ayo
- All agents accessible by name
- Sandboxes are invisible infrastructure

Your proposal makes sandboxes visible ("teams"). Migration path?

---

## Alternative Approaches

Before implementing the full proposal, consider lighter-weight options:

### Option A: Project-Scoped Agents (Minimal Change)

Use existing `.ayo.json` more heavily:

```json
// .ayo.json in project directory
{
  "team": "frontend",
  "agents": ["@react-dev", "@designer"],
  "delegates": {
    "coding": "@react-dev",
    "design": "@designer"
  },
  "collaboration_group": "frontend-team"
}
```

- Agents in same `collaboration_group` share sandbox
- No structural changes needed
- Teams are implicit via config

### Option B: Named Sandboxes (Medium Change)

Add named, persistent sandboxes without changing @ayo:

```bash
ayo sandbox create frontend --agents @react-dev,@designer
ayo sandbox create backend --agents @api-dev,@tester

# Agents auto-route to their sandbox
ayo @react-dev "build component"  # → runs in "frontend" sandbox
```

- Teams are sandbox names
- Agents bound to sandboxes at creation
- Single @ayo still orchestrates globally

### Option C: Team Profiles (Larger Change)

Introduce first-class "Team" concept:

```yaml
# ~/.config/ayo/teams/frontend.yaml
name: frontend
description: Frontend development team
sandbox:
  persistent: true
  resources: {cpus: 4, memory_mb: 4096}
agents:
  - @react-dev
  - @designer
local_ayo:
  system_prompt: "You are the frontend team lead..."
  delegates: {coding: @react-dev, design: @designer}
tickets_dir: .tickets/frontend/
```

- Explicit team entity
- Local @ayo with custom prompt
- Team-scoped resources
- Closest to your full proposal

---

## Implementation Complexity Estimate

| Component | Effort | Risk |
|-----------|--------|------|
| Team config format | Low | Low |
| Named/persistent sandboxes | Medium | Medium |
| Agent-to-team binding | Medium | Medium |
| Team-scoped tickets | Low | Low (already have per-session) |
| Team-scoped memory | Medium | Medium |
| Local @ayo instances | High | High |
| Global↔Local @ayo communication | High | High |
| Cross-team coordination | High | High |
| Migration/backwards compat | Medium | Medium |

**Total: Major architectural change.** Probably 2-4 weeks of focused work.

---

## Questions for You to Answer

Before proceeding, clarify:

1. **What's the primary use case?** Describe a concrete scenario where current architecture fails and teams would help.

2. **Do local @ayos need different personalities?** Or just different delegate configurations?

3. **Are teams static or dynamic?** Created once and persistent? Or ad-hoc per project?

4. **Cross-team work — how common?** If rare, simple handoffs work. If common, need real protocol.

5. **One user or multiple?** Does each user get their own teams? Or shared team infrastructure?

6. **Sandboxes as isolation or as identity?** Current: isolation. Proposed: identity. Which matters more?

---

## My Recommendation

**Start with Option B (Named Sandboxes)**, then evolve:

1. **Phase 1:** Named, persistent sandboxes with agent bindings
   - `ayo sandbox create frontend --agents @react-dev,@designer`
   - Agents auto-route to their sandbox
   - Global @ayo unchanged

2. **Phase 2:** Team config files
   - Move sandbox definitions to config
   - Add team-scoped tickets (directory per team)

3. **Phase 3:** Local @ayo (if needed)
   - Only if global @ayo proves insufficient
   - Start with just different delegate configs, not full prompt override

This incremental approach lets you:
- Validate the concept with real use
- Avoid over-engineering
- Keep existing functionality working
- Learn what's actually needed

---

## Summary

Your idea of sandbox-based teams with per-sandbox @ayo orchestrators is architecturally significant. It transforms sandboxes from invisible infrastructure into explicit organizational units.

**What's good about the idea:**
- Clear team boundaries
- Specialized orchestration per domain
- Team-local state (tickets, memory)
- Scales to complex multi-team projects

**What's risky:**
- Multiple @ayos = complexity and potential conflicts
- Persistent sandboxes = resource overhead
- Cross-team coordination needs new protocols
- Major architectural change

**Before committing:**
- Clarify the specific problem being solved
- Consider lighter-weight alternatives first
- Plan incremental delivery
- Accept that this is a multi-week effort

The ticket system you just built is a great foundation for team coordination — it can work at team scope naturally. The question is whether you also need per-team @ayo instances, or whether global @ayo + team-scoped tickets + collaboration groups is sufficient.
