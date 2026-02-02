---
id: am-awxr
status: open
deps: []
links: []
created: 2026-02-02T02:56:49Z
type: task
priority: 2
assignee: Alex Cabrera
---
# Consolidate duplicate SpinnerType definitions

SpinnerType enum and spinner frame constants are duplicated:
- internal/ui/spinner.go (lines 24-37)
- internal/ui/shared/spinner.go (lines 4-17)

Consolidate to a single definition.

