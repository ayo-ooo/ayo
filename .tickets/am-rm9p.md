---
id: am-rm9p
status: done
deps: []
links: []
created: 2026-02-02T02:57:06Z
type: task
priority: 3
assignee: Alex Cabrera
---
# Avoid double config loading in fantasy_tools.go

fantasy_tools.go loads config twice:
- Line 149: config.Load() for category resolution
- Line 480: config.Load() again for tool alias resolution

Pass config as parameter or cache the result to avoid redundant loading.

