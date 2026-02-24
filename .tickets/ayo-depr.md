---
id: ayo-depr
status: open
deps: []
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-6h19
tags: [cleanup, tech-debt]
---
# Remove deprecated functions

Remove all deprecated functions and update callers.

## Functions to Remove

| File | Function | Replacement |
|------|----------|-------------|
| `internal/run/fantasy_tools.go:139` | `NewFantasyToolSetWithOptions` | Use newer API |
| `internal/ui/styles.go:262` | `TruncateWithEllipsis` | Use lipgloss builtin |
| `internal/ui/shared/toolformat.go:176` | `FormatDuration` | Use standard time formatting |
| `internal/ui/shared/toolformat.go:182` | `TruncateText` | Use lipgloss builtin |

## Steps

1. Find all callers of each deprecated function
2. Update callers to use replacement
3. Remove deprecated function
4. Run tests

## Testing

- All tests pass after removal
- No compiler errors
