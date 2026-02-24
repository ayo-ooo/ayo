---
id: ayo-evnt
status: closed
deps: []
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-memx
tags: [tui, interactive]
---
# Task: Interactive Mode Event Rendering

## Summary

Design and implement a clean, CLI-first event rendering system for interactive mode that handles all events from agent interactions without hanging or blocking.

## Context

Interactive mode needs to render:
- Agent text responses (streaming)
- Tool calls and their results
- Planning/thinking indicators
- Delegate invocations
- Error states
- Multi-agent coordination events

Current TUI is complex and slow. New approach: streaming-first with inline status.

## Design Principles

1. **CLI-first**: Works in any terminal, no special capabilities required
2. **Non-blocking**: Never hang waiting for events
3. **Streaming**: Text appears as it's generated
4. **Minimal state**: Don't track complex UI state
5. **Charm libraries**: Use lipgloss, glamour sparingly

## Event Types

```go
type EventType string

const (
    EventText           EventType = "text"           // Streaming text
    EventToolStart      EventType = "tool_start"     // Tool invocation starting
    EventToolProgress   EventType = "tool_progress"  // Tool progress update
    EventToolComplete   EventType = "tool_complete"  // Tool finished
    EventToolError      EventType = "tool_error"     // Tool failed
    EventThinking       EventType = "thinking"       // Agent thinking indicator
    EventDelegate       EventType = "delegate"       // Delegation to another agent
    EventPlanUpdate     EventType = "plan_update"    // Todo/ticket change
    EventMemory         EventType = "memory"         // Memory operation
    EventError          EventType = "error"          // Error occurred
    EventComplete       EventType = "complete"       // Response complete
)

type Event struct {
    Type      EventType
    Timestamp time.Time
    Data      any
}
```

## Rendering Strategy

### Text Streaming
```
@ayo: I'll look at the authentication code and fix the bug.
```

Just print directly. No buffering.

### Tool Calls (Inline)
```
@ayo: Looking at the code...

  ▸ grep: "auth" internal/auth/*.go
    └─ 23 matches in 4 files

  ▸ view: internal/auth/handler.go:45-80
    └─ Read 35 lines

  ▸ edit: internal/auth/handler.go
    └─ Fixed nil pointer check

The bug was a missing nil check on line 52.
```

### Planning Updates
```
  ⊡ Analyzing code structure
  ⊠ Finding the bug
  ◉ Implementing fix ← in progress
  ○ Running tests
```

### Delegate Invocations
```
  → Delegating to @crush for implementation...
  
  @crush: I'll implement the authentication handler.
  
  ← @crush completed
```

### Errors
```
  ✗ bash: Command failed (exit 1)
    └─ Error: file not found
```

## Implementation

### Renderer Interface

```go
type EventRenderer interface {
    // Render handles a single event
    Render(event Event) error
    
    // Flush ensures all buffered output is written
    Flush() error
    
    // Reset clears any state for new conversation turn
    Reset()
}
```

### Simple Renderer (Default)

```go
type SimpleRenderer struct {
    out     io.Writer
    spinner *spinner  // Optional, nil if not interactive
    glamour *glamour.TermRenderer  // Reused, not recreated
}

func (r *SimpleRenderer) Render(event Event) error {
    switch event.Type {
    case EventText:
        // Direct write, no processing
        fmt.Fprint(r.out, event.Data.(string))
        
    case EventToolStart:
        tool := event.Data.(*ToolStartData)
        fmt.Fprintf(r.out, "\n  ▸ %s: %s\n", tool.Name, tool.Summary)
        
    case EventToolComplete:
        result := event.Data.(*ToolCompleteData)
        fmt.Fprintf(r.out, "    └─ %s\n", result.Summary)
        
    // ... etc
    }
    return nil
}
```

### Channel-Based Event Flow

```go
func RunInteractive(ctx context.Context, agent Agent) error {
    events := make(chan Event, 100)  // Buffered to prevent blocking
    renderer := NewSimpleRenderer(os.Stdout)
    
    // Event consumer (non-blocking)
    go func() {
        for event := range events {
            renderer.Render(event)
        }
    }()
    
    // Run agent with event callback
    return agent.Run(ctx, func(e Event) {
        select {
        case events <- e:
        default:
            // Drop event if buffer full (shouldn't happen)
            log.Warn("event buffer full, dropping event")
        }
    })
}
```

## Avoiding Hangs

1. **Buffered channels**: Events go to buffered channel, never block agent
2. **Timeouts**: All I/O operations have timeouts
3. **No complex state**: Simple linear rendering, no viewport management
4. **Graceful degradation**: If terminal doesn't support features, fall back

## Glamour Usage

**Critical**: Initialize glamour ONCE at startup, reuse for all markdown rendering.

```go
var glamourRenderer *glamour.TermRenderer

func init() {
    var err error
    glamourRenderer, err = glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(80),
    )
    if err != nil {
        // Fall back to no rendering
        glamourRenderer = nil
    }
}

func RenderMarkdown(md string) string {
    if glamourRenderer == nil {
        return md
    }
    out, err := glamourRenderer.Render(md)
    if err != nil {
        return md
    }
    return out
}
```

## Files to Create

- `internal/ui/interactive/events.go` - Event types
- `internal/ui/interactive/renderer.go` - Renderer interface + simple impl
- `internal/ui/interactive/session.go` - Session management
- `internal/ui/interactive/input.go` - Input handling

## Files to Remove/Replace

- `internal/ui/chat/model.go` - Complex TUI model
- `internal/ui/chat/event_aggregator.go` - Over-engineered
- `internal/ui/chat/tool_call_tree.go` - Too complex

## Implementation Steps

1. [ ] Define Event types
2. [ ] Create EventRenderer interface
3. [ ] Implement SimpleRenderer
4. [ ] Create singleton glamour renderer
5. [ ] Implement channel-based event flow
6. [ ] Create session input handling
7. [ ] Add tool call formatting
8. [ ] Add plan/todo rendering
9. [ ] Add delegate rendering
10. [ ] Add error rendering
11. [ ] Test with real agent interactions
12. [ ] Remove old TUI code

## Dependencies

- Depends on: None
- Blocks: `ayo-tui1`, `ayo-tui2`, `ayo-tui3`, `ayo-tui4`

## Acceptance Criteria

- [ ] Interactive mode never hangs
- [ ] All event types render cleanly
- [ ] Works in basic terminals (no special requirements)
- [ ] Glamour initialized once, reused
- [ ] Tool calls show inline with results
- [ ] Planning updates visible
- [ ] Delegate invocations clear
- [ ] Ctrl+C always works

---

*Created: 2026-02-23*
