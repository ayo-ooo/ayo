---
id: am-yf41
status: closed
deps: []
links: []
created: 2026-02-20T02:50:24Z
type: task
priority: 1
assignee: Alex Cabrera
tags: [squads, planners]
---
# Initialize squad-specific planners during squad dispatch

When dispatching to a squad, planners should be initialized based on SQUAD.md frontmatter configuration (planners.near_term, planners.long_term). The planner tools should then be available to the agent running in that squad context.

## Current State

**Already implemented:**
- `PlannerPlugin` interface (internal/planners/interface.go:9-42)
- `SandboxPlannerManager.GetPlanners()` (internal/planners/manager.go:65)
- `PlannersConfig` with `NearTerm`/`LongTerm` fields (internal/planners/types.go:49-75)
- SQUAD.md frontmatter parsing for planners (internal/squads/context.go:157-185)
- `InitSquadPlanners()` (internal/sandbox/squad.go:209-223)
- `GetPlannerTools()` helper (internal/run/fantasy_tools.go:266-279)
- `PlannerTools` field in `ToolSetOptions` (internal/run/fantasy_tools.go:125)

**Dependencies closed:**
- am-gt91: closed
- am-y55g: closed (added PlannerTools for @ayo)

## The Gap

Squad planner tools are initialized but **not passed to the agent toolset**.

In `squad_invoker.go:89-92`:
```go
if squadPlanners != nil {
    runnerOpts.PlannerManager = i.plannerManager
}
// MISSING: Need to pass squadPlanners to runner
```

In `run.go:950-957`, planner tools only added for @ayo:
```go
if isAyoAgent(ag.Handle) && r.ayoPlanners != nil {
    plannerTools = GetPlannerTools(...)
}
// MISSING: Squad agents don't get planner tools
```

## Implementation Plan

### 1. Add `SquadPlanners` field to `RunnerOptions`

```go
// internal/run/run.go, around line 94
type RunnerOptions struct {
    // ... existing fields ...
    SquadPlanners *planners.SandboxPlanners // Planners for squad context
}
```

### 2. Store squad planners in Runner

```go
// internal/run/run.go, around line 49
type Runner struct {
    // ... existing fields ...
    squadPlanners *planners.SandboxPlanners // Planners for squad context
}
```

### 3. Pass squad planners from invoker

```go
// internal/daemon/squad_invoker.go:83-92
runnerOpts := run.RunnerOptions{
    Services:        i.services,
    SandboxProvider: i.sandboxProvider,
    SquadName:       params.SquadName,
    SquadPlanners:   squadPlanners,  // ADD THIS
}
```

### 4. Use squad planners in tool creation

```go
// internal/run/run.go, around line 950-967
// In createFantasyAgent or similar:
var plannerTools []fantasy.AgentTool
if r.squadPlanners != nil {
    plannerTools = GetPlannerTools(r.squadPlanners)
} else if isAyoAgent(ag.Handle) && r.ayoPlanners != nil {
    plannerTools = GetPlannerTools(r.ayoPlanners)
}
```

## Key Code Locations

| File | Lines | Purpose |
|------|-------|---------|
| internal/daemon/squad_invoker.go | 63-92 | Squad invocation with planners |
| internal/run/run.go | 82-95 | RunnerOptions struct |
| internal/run/run.go | 950-967 | Planner tools injection for @ayo |
| internal/run/fantasy_tools.go | 266-279 | GetPlannerTools() helper |
| internal/planners/manager.go | 65-100 | GetPlanners() implementation |

## Acceptance Criteria

- Squad SQUAD.md frontmatter planners config is respected
- Agent in squad has access to configured planner tools (ticket_create, ticket_list, etc.)
- Planner state is stored in squad-specific directories (~/.local/share/ayo/sandboxes/squads/{name}/)
