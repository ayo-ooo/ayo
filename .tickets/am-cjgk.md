---
id: am-cjgk
status: closed
deps: []
links: []
created: 2026-02-18T03:19:50Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-1wsq
---
# Deprecate 'chains' terminology

Remove or redirect any references to 'chains' terminology.

## Context
- 'Chains' concept merged into flows (schema validation between steps)
- Need to update docs and code comments

## Tasks
1. Search for 'chain' references in docs
2. Replace with 'flow' or 'schema validation' as appropriate
3. Update any CLI help text
4. Add deprecation note if referenced externally

## Files to Search
- docs/*.md
- cmd/ayo/*.go
- internal/flows/*.go

## Acceptance
- No standalone 'chain' concept in docs
- Redirects to flow documentation
- Code comments updated

