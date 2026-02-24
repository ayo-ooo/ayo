---
id: ayo-dupl
status: closed
deps: []
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-6h19
tags: [cleanup, tech-debt]
---
# Consolidate duplicate interfaces and types

**Status: N/A** - Upon review, the "duplicates" are intentionally different:

1. **AgentInvoker**: The two remaining interfaces (`squads.AgentInvoker` and `daemon.AgentInvoker`) have different method signatures - one uses `InvokeParams/InvokeResult` structs, the other uses individual string params. They serve different architectural purposes and cannot be consolidated without breaking the design.

2. **Todo types**: `run.Todo` uses `TodoStatus` type while `shared.Todo` uses plain `string`. The UI type intentionally avoids importing `run` to prevent import cycles.

3. **yaml_executor.go** (listed as 3rd location) was already deleted in ayo-1ryh.
