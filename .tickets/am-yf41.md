---
id: am-yf41
status: open
deps: [am-gt91, am-y55g]
links: []
created: 2026-02-20T02:50:24Z
type: task
priority: 1
assignee: Alex Cabrera
tags: [squads, planners]
---
# Initialize squad-specific planners during squad dispatch

When dispatching to a squad, planners should be initialized based on SQUAD.md frontmatter configuration (planners.near_term, planners.long_term). The planner tools should then be available to the agent running in that squad context.

## Design

In squad dispatch, after loading constitution, extract planner configuration from frontmatter. Call SandboxPlannerManager.GetPlanners() with the squad's sandbox directory and planner config. Pass planner tools to the agent tool set.

## Acceptance Criteria

- Squad SQUAD.md frontmatter planners config is respected
- Agent in squad has access to configured planner tools
- Planner state is stored in squad-specific directories

