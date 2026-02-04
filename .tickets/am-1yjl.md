---
id: am-1yjl
status: done
deps: []
links: []
created: 2026-02-02T02:56:57Z
type: task
priority: 2
assignee: Alex Cabrera
---
# Extract duplicated buffer/result types in run package

Duplicated types in run package:
- fantasyBashResult in fantasy_tools.go:113-128
- externalToolResult in external_tools.go:416-440

Both have same structure and String() method. Extract to internal/run/util.go or internal/run/result.go.

