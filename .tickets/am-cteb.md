---
id: am-cteb
status: open
deps: []
links: []
created: 2026-02-02T02:56:24Z
type: task
priority: 2
assignee: Alex Cabrera
---
# Remove duplicate BashParams struct definitions

BashParams is defined in two places:
- run/fantasy_tools.go:26
- ui/shared/toolformat.go:18

The shared package version has an extra RunInBackground field. Consolidate to a single definition in a shared location.

