# Simplified Interactive Mode Design

This document describes the design for a simplified interactive chat mode that replaces the current slow, complex TUI.

## Current Problems

| Issue | Root Cause |
|-------|------------|
| Slow rendering | Full viewport re-render on every 80ms tick |
| Complex code | Triple abstraction for tool rendering |
| Animation overhead | ID-scoped ticks to prevent race conditions |
| Memory overhead | Complex pending/rendered tracking sets |

## Design Principles

1. **Stream-first**: Text streams directly to terminal, no buffering
2. **No viewport**: Use terminal scrollback, user scrolls naturally
3. **Inline tools**: Tool calls appear inline, not in separate panels
4. **Simple focus**: Only the input prompt is interactive
5. **Minimal state**: Just current message + input

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│ Components                                                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐                                           │
│  │  Interactive    │ - Main bubbletea model                    │
│  │  Model          │ - Handles input, delegates rendering      │
│  └────────┬────────┘                                           │
│           │                                                     │
│  ┌────────┴────────┐                                           │
│  │                 │                                            │
│  ▼                 ▼                                            │
│  ┌──────────┐  ┌──────────┐                                    │
│  │ Streamer │  │  Input   │                                    │
│  │          │  │  Prompt  │                                    │
│  └──────────┘  └──────────┘                                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Components

| Component | File | Responsibility |
|-----------|------|----------------|
| Interactive Model | `internal/ui/interactive/model.go` | Main bubbletea model, coordinates input and output |
| Streamer | `internal/ui/interactive/streamer.go` | Renders agent text and tool calls to terminal |
| Input Prompt | `internal/ui/interactive/input.go` | Handles user input with line editing |
| Event Renderer | `internal/ui/interactive/renderer.go` | Formats events for display |

## Visual Design

```
┌─────────────────────────────────────────────────────────────────┐
│ @ayo                                                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ You: Fix the bug in auth.go                                     │
│                                                                 │
│ @ayo: I'll look at the authentication code.                    │
│                                                                 │
│   ▸ bash: grep -n "auth" internal/auth/*.go                    │
│     └─ 23 matches in 4 files                                   │
│                                                                 │
│   ▸ view: internal/auth/handler.go:45-80                       │
│     └─ Read 35 lines                                           │
│                                                                 │
│   ▸ edit: internal/auth/handler.go                             │
│     └─ Fixed nil pointer check                                 │
│                                                                 │
│ The bug was a missing nil check on line 52. Fixed.             │
│                                                                 │
│ ─────────────────────────────────────────────────────────────── │
│ > _                                                             │
└─────────────────────────────────────────────────────────────────┘
```

## Key Differences from Current

| Current | New |
|---------|-----|
| Viewport with scrolling | Direct terminal output |
| Complex tool tree | Simple inline list |
| Sidebar panels | No sidebars |
| 80ms tick animations | No tick-based animation |
| EventAggregator | Direct channel read |
| Multiple focus modes | Single input focus |

## State Machine

```
                    ┌─────────────────────────────────────────────┐
                    │                                             │
                    ▼                                             │
┌─────────────┐  input  ┌─────────────┐  done  ┌─────────────┐   │
│   IDLE      │ ──────► │  STREAMING  │ ─────► │    IDLE     │   │
│ (awaiting   │         │  (receiving │        │  (awaiting  │   │
│   input)    │         │   response) │        │    input)   │   │
└─────────────┘         └─────────────┘        └─────────────┘   │
      ▲                        │                                  │
      │                        │ ctrl+c                           │
      │                        ▼                                  │
      │                 ┌─────────────┐                           │
      └──────────────── │  CANCELLED  │ ──────────────────────────┘
                        │             │
                        └─────────────┘
```

### States

| State | Description |
|-------|-------------|
| IDLE | Waiting for user input. Input prompt is active. |
| STREAMING | Receiving and rendering agent response. Input disabled. |
| CANCELLED | User pressed Ctrl+C to cancel. Returns to IDLE. |

## Event Flow

```
┌──────────────┐
│    User      │
│   Input      │
└──────┬───────┘
       │
       ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Bubbletea  │────►│    Agent     │────►│   Provider   │
│    Model     │     │   Execute    │     │   (LLM)      │
└──────────────┘     └──────────────┘     └──────────────┘
       ▲                    │
       │                    ▼
       │             ┌──────────────┐
       │             │    Event     │
       │             │   Channel    │
       │             └──────┬───────┘
       │                    │
       │                    ▼
       │             ┌──────────────┐
       └─────────────│   Streamer   │
                     │   (render)   │
                     └──────────────┘
```

### Event Types

| Event | Description | Rendering |
|-------|-------------|-----------|
| `text` | Streaming text from agent | Print directly |
| `tool_start` | Tool invocation starting | `  ▸ tool: args` |
| `tool_complete` | Tool finished | `    └─ summary` |
| `tool_error` | Tool failed | `    └─ ✗ error` |
| `thinking` | Agent thinking | Spinner (optional) |
| `complete` | Response finished | Enable input |

## Implementation Files

| File | Purpose |
|------|---------|
| `internal/ui/interactive/model.go` | Bubbletea model |
| `internal/ui/interactive/input.go` | Input prompt component |
| `internal/ui/interactive/streamer.go` | Event streaming to terminal |
| `internal/ui/interactive/renderer.go` | Event formatting |
| `internal/ui/interactive/events.go` | Event type definitions |

## Files to Remove

After implementation, remove the old TUI code:

- `internal/ui/chat/model.go` - Complex TUI model
- `internal/ui/chat/event_aggregator.go` - Over-engineered event handling
- `internal/ui/chat/tool_call_tree.go` - Complex tool rendering

## Performance Goals

| Metric | Current | Target |
|--------|---------|--------|
| Render latency | 80ms tick | <1ms (immediate) |
| Memory usage | High (tracking sets) | Low (stream through) |
| Startup time | ~500ms | <100ms |

## Accessibility

- Works in any terminal (no special requirements)
- Standard keyboard shortcuts (Ctrl+C, Ctrl+D)
- Screen reader friendly (plain text output)
- No animations required for function

---

*Design document for ayo-tui1*
