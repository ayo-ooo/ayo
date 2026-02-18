---
id: am-okzf
status: closed
deps: []
links: []
created: 2026-02-18T03:12:59Z
type: epic
priority: 2
assignee: Alex Cabrera
---
# Flow-Squad Integration

Enable flows to target specific squads. Implement AgentInvoker with sandbox awareness. Flows are external triggers (CI/CD, webhooks) that orchestrate across squads.

## Acceptance Criteria

- AgentInvoker implemented with sandbox context
- Flow steps can specify squad target
- Cross-squad data flow via flow runtime serialization
- Flows validate I/O against squad schemas
- Squad sandboxes started on-demand by flow runtime

