---
id: ayo-ac3v
status: closed
deps: []
links: []
created: 2026-02-12T19:47:02Z
type: chore
priority: 2
assignee: Alex Cabrera
parent: ayo-7j3v
tags: [cleanup, refactor]
---
# Consolidate duplicate truncate functions

8+ implementations of truncate/truncateText exist across codebase. Create single shared utility:
- internal/run/run.go:936 truncateForTitle
- internal/sandbox/bash.go:224 truncateOutput  
- internal/memory/zettelkasten/observer.go:324 truncate
- internal/ui/memory.go:135 truncateMemory
- internal/ui/json_render.go:338 truncate
- internal/ui/chat/messages/template_renderer.go:148 truncateFunc
- internal/ui/chat/messages/renderer.go:302 truncateText
- internal/ui/chat/chat.go:632 truncateOutput

Create internal/util/truncate.go with single implementation, update all call sites.

## Acceptance Criteria

Single truncate utility in internal/util, all duplicates removed, tests pass

