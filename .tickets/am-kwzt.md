---
id: am-kwzt
status: open
deps: []
links: []
created: 2026-02-02T02:56:34Z
type: task
priority: 3
assignee: Alex Cabrera
---
# Remove unused pubsub.AgentEvent type

pubsub.AgentEvent is defined but never used anywhere in the codebase. Other event types (MessageEvent, ToolEvent, etc.) are used, but AgentEvent has no references.

