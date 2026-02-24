---
id: ayo-tui4
status: open
deps: [ayo-tui2, ayo-tui3]
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-memx
tags: [tui, cleanup]
---
# Remove old TUI code

After new interactive mode is working, remove the old complex TUI.

## Files to Remove

| File | Lines | Reason |
|------|-------|--------|
| `internal/ui/chat/chat.go` | ~1242 | Main complex model |
| `internal/ui/chat/aggregator.go` | ~150 | EventAggregator not needed |
| `internal/ui/chat/messages/tree.go` | ~200 | Complex tool tree |
| `internal/ui/chat/messages/toolcall.go` | ~150 | Tool call component |
| `internal/ui/chat/messages/renderer.go` | ~200 | Render logic |
| `internal/ui/chat/messages/registry.go` | ~100 | Tool registry |
| `internal/ui/chat/panels/sidebar.go` | ~150 | Sidebar panels |
| `internal/ui/chat/statusbar.go` | ~100 | Status bar |

**Estimated removal: ~2,300 lines**

## Files to Simplify

| File | Change |
|------|--------|
| `internal/ui/shared/` | Remove TUI-specific abstractions |
| `internal/run/stream_writer.go` | Simplify to work with new renderer |
| `internal/run/channel_writer.go` | May not be needed |

## Migration Steps

1. Ensure new interactive mode passes all tests
2. Update any code that imports old TUI
3. Remove old files
4. Update imports
5. Run tests
6. Remove unused dependencies from go.mod

## Testing

- Verify new interactive mode works for all use cases
- Test non-interactive (piped) mode still works
- Test JSON output mode still works
