---
id: am-m4uw
status: open
deps: []
links: []
created: 2026-02-02T02:56:52Z
type: task
priority: 3
assignee: Alex Cabrera
---
# Consolidate duplicate markdown renderer caching

GetMarkdownRenderer functions exist in multiple locations:
- internal/ui/shared/markdown.go
- internal/ui/chat/messages/markdown.go

Consolidate markdown renderer caching logic to a single location.

