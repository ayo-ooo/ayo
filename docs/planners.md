# Planners

Planners are plugins that provide work coordination tools to agents. They handle task tracking and management within sandboxes, enabling agents to organize their work systematically.

## Overview

Ayo uses a two-tier planning system:

| Type | Scope | Purpose | Default Plugin |
|------|-------|---------|----------------|
| **Near-term** | Session | Immediate task tracking, todos | `ayo-todos` |
| **Long-term** | Persistent | Project coordination, tickets | `ayo-tickets` |

Each sandbox (@ayo or squad) gets its own pair of planner instances with isolated state.

## Near-Term vs Long-Term

### Near-Term Planners

Near-term planners handle immediate work within a session:
- Todo lists
- Step-by-step task breakdowns
- Session-scoped work queues

**Characteristics:**
- State may be ephemeral or persist across sessions
- Focus on "what to do next"
- Lightweight, minimal overhead

### Long-Term Planners

Long-term planners handle persistent project coordination:
- Ticket systems
- Issue trackers
- Kanban boards

**Characteristics:**
- State always persists across sessions
- Focus on "what needs to be done overall"
- Rich metadata, dependencies, status tracking

## Configuration

### Global Defaults

Set default planners in `~/.config/ayo/config.toml`:

```toml
[planners]
near_term = "ayo-todos"
long_term = "ayo-tickets"
```

### Per-Squad Overrides

Squads can override planners in their `SQUAD.md` frontmatter:

```yaml
---
planners:
  near_term: custom-todos
  long_term: custom-tickets
---
# Squad: Frontend Team
...
```

### Resolution Order

1. Squad `SQUAD.md` frontmatter (if specified)
2. Global `config.toml` settings
3. Built-in defaults (`ayo-todos`, `ayo-tickets`)

## Built-in Planners

### ayo-todos

The default near-term planner providing a simple todo list:

**Tools:**
- `todos` - Create, update, and manage todo items

**Features:**
- Session-scoped task tracking
- Status tracking (pending, in_progress, completed)
- Progress display in UI

### ayo-tickets

The default long-term planner providing file-based ticket management:

**Tools:**
- `tk create` - Create new tickets
- `tk list` - List tickets with filters
- `tk show` - View ticket details
- `tk start/close` - Update ticket status

**Features:**
- Markdown-based tickets in `.tickets/` directory
- Dependencies between tickets
- Priority and assignment tracking
- Git-friendly workflow

## State Management

Planners store state in dedicated directories within each sandbox:

```
~/.local/share/ayo/sandboxes/
├── ayo/                          # @ayo sandbox
│   ├── .planner.near/            # Near-term planner state
│   └── .planner.long/            # Long-term planner state
└── squads/{name}/                # Squad sandbox
    ├── .planner.near/
    └── .planner.long/
```

State directories are:
- Created automatically when planners initialize
- Mounted into sandbox containers
- Persisted across container restarts

## Creating Custom Planners

### Plugin Interface

Implement the `PlannerPlugin` interface:

```go
package planners

import (
    "context"
    "charm.land/fantasy"
)

type PlannerPlugin interface {
    // Name returns the unique identifier for this planner.
    Name() string

    // Type returns whether this is near-term or long-term.
    Type() PlannerType

    // Init initializes the planner. StateDir is guaranteed to exist.
    Init(ctx context.Context) error

    // Close releases resources when sandbox shuts down.
    Close() error

    // Tools returns the fantasy tools this planner provides.
    Tools() []fantasy.AgentTool

    // Instructions returns text to inject into system prompts.
    Instructions() string

    // StateDir returns the planner's state directory.
    StateDir() string
}
```

### Factory Function

Create a factory to instantiate your planner:

```go
func NewMyPlanner(ctx PlannerContext) (PlannerPlugin, error) {
    return &MyPlanner{
        name:     "my-planner",
        stateDir: ctx.StateDir,
        config:   ctx.Config,
    }, nil
}
```

### Registration

Register your planner with the global registry:

```go
func init() {
    planners.Register("my-planner", NewMyPlanner)
}
```

### PlannerContext

The factory receives context about the sandbox:

```go
type PlannerContext struct {
    // SandboxName is "ayo" or the squad name
    SandboxName string

    // SandboxDir is the sandbox root directory
    SandboxDir string

    // StateDir is where to store planner state
    StateDir string

    // Config contains planner-specific configuration
    Config map[string]any
}
```

### Example: Custom Todo Planner

```go
package mytodos

import (
    "context"
    "charm.land/fantasy"
    "github.com/alexcabrera/ayo/internal/planners"
)

type MyTodos struct {
    name     string
    stateDir string
}

func New(ctx planners.PlannerContext) (planners.PlannerPlugin, error) {
    return &MyTodos{
        name:     "my-todos",
        stateDir: ctx.StateDir,
    }, nil
}

func (m *MyTodos) Name() string              { return m.name }
func (m *MyTodos) Type() planners.PlannerType { return planners.NearTerm }
func (m *MyTodos) StateDir() string          { return m.stateDir }

func (m *MyTodos) Init(ctx context.Context) error {
    // Load state from m.stateDir if it exists
    return nil
}

func (m *MyTodos) Close() error {
    // Save state to m.stateDir
    return nil
}

func (m *MyTodos) Tools() []fantasy.AgentTool {
    return []fantasy.AgentTool{
        // Define your tools here
    }
}

func (m *MyTodos) Instructions() string {
    return `Use the my-todos tool to track your work.`
}

func init() {
    planners.Register("my-todos", New)
}
```

## Lifecycle

1. **Sandbox Start**: `SandboxPlannerManager.GetPlanners()` called
2. **Planner Creation**: Factory creates planner instance
3. **Initialization**: `Init()` called with state directory ready
4. **Agent Sessions**: Tools and instructions injected into agents
5. **Sandbox Stop**: `Close()` called to release resources

## Best Practices

### Tool Design
- Keep tools focused and single-purpose
- Use clear, action-oriented names
- Handle errors gracefully

### Instructions
- Be concise - agents have limited context
- Focus on when and why to use tools
- Include examples of common patterns

### State Management
- Use atomic file operations
- Handle concurrent access safely
- Clean up stale state periodically

### Testing
- Test with mock sandboxes
- Verify state persistence
- Check concurrent access scenarios

## CLI Commands

```bash
# List available planners
ayo planner list

# Show planner details
ayo planner show ayo-todos

# View current sandbox planners
ayo planner status
```

## Comparison to Other Systems

| Feature | Planners | Flows | Tickets |
|---------|----------|-------|---------|
| Scope | Per-sandbox | Global | Per-directory |
| Control | Agent-driven | User-defined | Agent-driven |
| Persistence | Configurable | None | Always |
| Tools | Dynamic | None | Static |

**Use Planners when:**
- Agents need work coordination tools
- Different sandboxes need different coordination
- You want pluggable coordination systems

**Use Flows when:**
- Work is sequential and user-defined
- No agent decision-making needed
- Simple input → output pipelines

**Use Tickets directly when:**
- Working outside sandbox context
- Manual project management
- Shared across multiple agents
