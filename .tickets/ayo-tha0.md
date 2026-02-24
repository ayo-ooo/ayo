---
id: ayo-tha0
status: open
deps: []
links: []
created: 2026-02-23T22:15:03Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [removal, web]
---
# Remove web/ directory

Delete the web interface frontend code.

## Context

The `web/` directory contains HTML, JavaScript, and WASM for a browser-based UI. This is being removed to focus on CLI-first experience.

## Files to Delete

```
web/
├── index.html
├── app.js
├── styles.css
├── wasm/
└── ... (~1000 lines total)
```

## Verification Steps

1. Delete `web/` directory
2. Search for references: `grep -r "web/" --include="*.go" .`
3. Remove any embed directives that reference web/
4. Run `go build ./...` - should pass

## Acceptance Criteria

- [ ] `web/` directory deleted
- [ ] No Go embed references to web/
- [ ] Build passes
