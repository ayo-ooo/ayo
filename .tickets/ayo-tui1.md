---
id: ayo-tui1
status: closed
deps: []
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-memx
tags: [tui, interactive]
---
# Design simplified interactive mode

Design a simpler interactive chat mode that replaces the current slow, complex TUI.

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

## Proposed Architecture

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

## Deliverables

1. Design document in `docs/internal/tui-design.md`
2. Mock screenshots or ASCII diagrams
3. Component list with responsibilities
4. State machine for interactive mode
5. Event flow diagram
