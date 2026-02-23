---
id: am-q1p3
status: closed
deps: [am-yf41]
links: []
created: 2026-02-20T02:51:14Z
type: task
priority: 2
assignee: Alex Cabrera
tags: [squads, planners, tickets]
---
# Create tickets via planner tools during squad work coordination

When squad lead coordinates work between agents, it should use the long-term planner (ayo-tickets) to create and track tickets for each piece of work. This enables work visibility and progress tracking.

## Resolution

All components are implemented:

1. **Ticket tools** - Available via ayo-tickets planner plugin (internal/planners/builtin/tickets/tools.go)
2. **CLI commands** - All commands exist in cmd/ayo/squad.go:
   - `ayo squad ticket NAME create` (line 429)
   - `ayo squad ticket NAME list` (line 484)
   - `ayo squad ticket NAME start` (line ~520)
   - `ayo squad ticket NAME close` (line ~560)
3. **Squad planner initialization** - Completed in am-yf41
4. **Documentation** - Added "Work Coordination with Tickets" section to docs/squads.md

### Key Code Locations

| Component | File | Lines |
|-----------|------|-------|
| squadTicketCmd() | cmd/ayo/squad.go | 407-427 |
| squadTicketCreateCmd() | cmd/ayo/squad.go | 429-482 |
| squadTicketListCmd() | cmd/ayo/squad.go | 484-540 |
| TicketsPlugin.Tools() | internal/planners/builtin/tickets/tools.go | 144+ |

## Acceptance Criteria

- ✅ Squad lead creates tickets using ticket tool
- ✅ Tickets appear in squad's .tickets/ directory
- ✅ 'ayo squad ticket list' shows squad work items
- ✅ Documentation updated with ticket workflow
