---
id: ayo-glam
status: open
deps: []
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-memx
tags: [tui, performance]
---
# Task: Glamour Initialization Optimization

## Summary

Optimize glamour (Charm's markdown renderer) usage by initializing once and reusing, rather than creating new instances repeatedly.

## Context

Glamour is used throughout the codebase for markdown rendering. The tendency to initialize it repeatedly causes performance issues:

```go
// BAD: Creates new renderer each call
func RenderMarkdown(md string) string {
    r, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())
    out, _ := r.Render(md)
    return out
}
```

This is expensive because:
- Style detection runs each time
- Word wrap calculation runs each time
- Memory allocations for each renderer

## Solution

Create a singleton glamour renderer initialized once at startup:

```go
package render

import (
    "sync"
    "github.com/charmbracelet/glamour"
)

var (
    renderer     *glamour.TermRenderer
    rendererOnce sync.Once
    rendererErr  error
)

// Markdown renders markdown to styled terminal output.
// Thread-safe, uses singleton renderer.
func Markdown(md string) string {
    rendererOnce.Do(func() {
        renderer, rendererErr = glamour.NewTermRenderer(
            glamour.WithAutoStyle(),
            glamour.WithWordWrap(100),
        )
    })
    
    if rendererErr != nil || renderer == nil {
        return md  // Graceful fallback
    }
    
    out, err := renderer.Render(md)
    if err != nil {
        return md
    }
    return out
}

// MarkdownWithWidth renders with custom width (creates new renderer)
// Use sparingly for special cases only
func MarkdownWithWidth(md string, width int) string {
    r, err := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(width),
    )
    if err != nil {
        return md
    }
    out, _ := r.Render(md)
    return out
}
```

## Implementation Steps

1. [ ] Search codebase for all glamour usages
2. [ ] Create `internal/render/markdown.go` with singleton
3. [ ] Replace all `glamour.NewTermRenderer` calls with `render.Markdown()`
4. [ ] Add benchmark tests comparing old vs new approach
5. [ ] Document the pattern in code comments

## Files to Search

```bash
grep -r "glamour.NewTermRenderer" internal/
grep -r "glamour.Render" internal/
```

## Expected Locations

- `internal/ui/chat/` - Chat rendering
- `internal/ui/shared/` - Shared UI components
- `internal/run/` - Agent output rendering
- `cmd/ayo/` - CLI output

## Acceptance Criteria

- [ ] Single glamour renderer instance across entire process
- [ ] Thread-safe access to renderer
- [ ] Graceful fallback if renderer fails
- [ ] No performance regression in rendering
- [ ] Benchmark shows improvement

## Benchmark Test

```go
func BenchmarkMarkdownOld(b *testing.B) {
    md := "# Hello\n\nThis is **bold** and _italic_."
    for i := 0; i < b.N; i++ {
        r, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())
        r.Render(md)
    }
}

func BenchmarkMarkdownNew(b *testing.B) {
    md := "# Hello\n\nThis is **bold** and _italic_."
    for i := 0; i < b.N; i++ {
        render.Markdown(md)
    }
}
```

Expected improvement: 10-100x faster after first call.

---

*Created: 2026-02-23*
