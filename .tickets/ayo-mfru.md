---
id: ayo-mfru
status: closed
deps: []
links: []
created: 2026-02-12T19:46:39Z
type: chore
priority: 3
assignee: Alex Cabrera
parent: ayo-7j3v
tags: [cleanup, cmd]
---
# Remove unused parameters in cmd/ayo package

Remove unused function parameters flagged by gopls:
- agents_capabilities.go:214 cfg in listAllCapabilities()
- agents_capabilities.go:295 cfg in searchCapabilities()
- agents_lifecycle.go:61 cfgPath in archiveAgentCmd()
- agents_lifecycle.go:107 cfgPath in unarchiveAgentCmd()
- chat.go:14 debug in runInteractiveChat()
- flows.go:729 cfgPath
- plugins.go:140,255 cfgPath

## Acceptance Criteria

gopls unusedparams warnings for cmd/ayo reduced to zero

