---
id: asmft-2gq7
status: done
deps: [asmft-doam]
links: []
created: 2026-02-13T14:37:16Z
type: task
priority: 1
assignee: Alex Cabrera
parent: asmft-rge8
tags: [phase7, ui]
---
# Add ticket status streaming

Create internal/squads/progress.go with StreamProgress(squad, callback) function. Watches ticket changes and calls callback with updates.

