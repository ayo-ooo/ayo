---
id: am-tmxv
status: open
deps: []
links: []
created: 2026-02-02T02:56:19Z
type: task
priority: 2
assignee: Alex Cabrera
---
# Extract common execution logic in flows package

flows/execute.go has ~80% duplicate code between Run() and RunStreaming():
- Input resolution
- Validation logic
- History recording
- Timeout setup
- Command preparation
- Error handling

Extract into a runCore() helper to reduce duplication.

