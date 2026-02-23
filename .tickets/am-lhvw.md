---
id: am-lhvw
status: closed
deps: []
links: []
created: 2026-02-20T02:51:03Z
type: feature
priority: 2
assignee: Alex Cabrera
tags: [squads, delegation]
---
# Implement agent-to-agent delegation within squad context

Squad lead (@ayo) should be able to delegate work to other agents within the squad. Currently there's no mechanism for one agent to invoke another agent in the same squad sandbox.

## Resolution

Implemented the `delegate` tool for agent-to-agent delegation within squads.

### Files Created

| File | Purpose |
|------|---------|
| internal/tools/delegate/delegate.go | Delegate tool implementation |

### Files Modified

| File | Changes |
|------|---------|
| internal/squads/context.go | Added `Agents` field to frontmatter, `GetAgents()` method, `parseAgentSections()` helper |
| internal/run/fantasy_tools.go | Added squad delegation fields to `ToolSetOptions`, delegate tool case in switch, auto-add delegate tool in squad context |
| internal/run/run.go | Added squad delegation fields to `Runner` struct and `RunnerOptions` |
| internal/daemon/squad_invoker.go | Pass SquadInvoker (self-reference), SquadConstitution, and SquadAgents to runner |

### Implementation Details

1. **Delegate Tool** (`internal/tools/delegate/delegate.go`)
   - Accepts `agent` (handle) and `prompt` parameters
   - Validates agent is in squad (using constitution's agent list)
   - Uses `squads.AgentInvoker.Invoke()` for execution
   - Returns agent's response or error

2. **Squad Agents Parsing** (`internal/squads/context.go`)
   - Agents can be defined in SQUAD.md frontmatter: `agents: [@backend, @frontend]`
   - Or parsed from `### @agent` sections in the markdown body
   - `Constitution.GetAgents()` returns normalized handles

3. **Auto-injection in Squad Context**
   - Delegate tool is automatically added when `SquadInvoker` is configured
   - No need to add "delegate" to agent's AllowedTools
   - Works like planner tools - always available in squad context

### Usage

Squad agents can delegate tasks:
```
delegate(agent: "@reviewer", prompt: "Review the code changes in src/main.go")
```

## Acceptance Criteria

- ✅ Squad lead can use delegate tool to invoke @reviewer
- ✅ Delegated agent runs in same squad sandbox
- ✅ Delegated agent receives SQUAD.md in system prompt
- ✅ Results are returned to lead agent's context
