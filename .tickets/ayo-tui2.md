---
id: ayo-tui2
status: open
deps: [ayo-tui1]
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-memx
tags: [tui, interactive]
---
# Implement streaming text renderer

Implement the core streaming renderer for interactive mode.

## Requirements

1. **Direct streaming**: Write tokens to terminal as they arrive
2. **Markdown rendering**: Use glamour for final markdown formatting
3. **Incremental display**: Don't re-render completed text
4. **Line buffering**: Buffer partial lines until newline received

## Implementation

### StreamRenderer

```go
// internal/ui/interactive/stream.go
type StreamRenderer struct {
    out       io.Writer
    glamour   *glamour.TermRenderer
    buffer    strings.Builder
    lastLine  string  // For incremental updates
}

func (r *StreamRenderer) WriteToken(token string) error {
    // Append to buffer
    r.buffer.WriteString(token)
    
    // If contains newline, render complete lines
    content := r.buffer.String()
    if idx := strings.LastIndex(content, "\n"); idx >= 0 {
        complete := content[:idx+1]
        r.buffer.Reset()
        r.buffer.WriteString(content[idx+1:])
        
        // Render and write
        rendered, _ := r.glamour.Render(complete)
        r.out.Write([]byte(rendered))
    }
    return nil
}

func (r *StreamRenderer) Flush() error {
    // Render any remaining content
    if r.buffer.Len() > 0 {
        rendered, _ := r.glamour.Render(r.buffer.String())
        r.out.Write([]byte(rendered))
        r.buffer.Reset()
    }
    return nil
}
```

### Integration with Runner

```go
// The runner sends stream events
runner.OnToken(func(token string) {
    renderer.WriteToken(token)
})

runner.OnComplete(func() {
    renderer.Flush()
})
```

## Testing

- Test streaming renders incrementally
- Test markdown formatting preserved
- Test code blocks render correctly
- Test partial line buffering works
