# Extending Ayo

Guide to creating custom providers, planners, and plugins for ayo.

## Overview

Ayo is designed to be extensible at multiple levels:

| Extension Point | Purpose | Complexity |
|-----------------|---------|------------|
| Agents | Custom AI personas | Low |
| Skills | Knowledge injection | Low |
| Tools | New capabilities | Medium |
| Planners | Work coordination | Medium |
| Providers | Infrastructure | High |
| Sandbox Providers | Execution environments | High |

## Creating Agents

The simplest extension—just markdown and JSON:

```bash
mkdir -p ~/.config/ayo/agents/my-agent
```

**agent.md** (system prompt):
```markdown
You are a specialized agent for...

## Capabilities
- ...

## Behavior
- ...
```

**ayo.json** (configuration):
```json
{
  "provider": "anthropic",
  "model": "your-model",
  "tools": ["bash", "view", "edit", "write"],
  "skills": ["./skills/domain.md"],
  "memory": {
    "enabled": true,
    "scope": "agent"
  }
}
```

Test your agent:
```bash
ayo run my-agent "Hello, introduce yourself"
```

## Creating Skills

Skills inject domain knowledge into agents:

```bash
mkdir -p ~/.config/ayo/skills/my-skill
```

**skill.md**:
```markdown
# Domain Knowledge

## Key Concepts
- ...

## Common Patterns
- ...

## Best Practices
- ...
```

Reference in agent's `ayo.json`:
```json
{
  "skills": ["my-skill"]
}
```

## Creating Tools

Tools extend agent capabilities with external commands or scripts.

### Tool Definition

Create `~/.config/ayo/tools/my-tool/tool.json`:

```json
{
  "name": "my-tool",
  "description": "Does something useful for the agent",
  "parameters": {
    "type": "object",
    "properties": {
      "input": {
        "type": "string",
        "description": "The input to process"
      },
      "format": {
        "type": "string",
        "enum": ["json", "text", "csv"],
        "description": "Output format"
      }
    },
    "required": ["input"]
  },
  "command": ["./run.sh", "{{input}}", "--format", "{{format}}"],
  "working_dir": "{{plugin_dir}}",
  "timeout": 30
}
```

### Tool Implementation

Create `run.sh`:
```bash
#!/bin/bash
input="$1"
format="${3:-text}"

# Your tool logic here
echo "Processed: $input (format: $format)"
```

Make executable:
```bash
chmod +x ~/.config/ayo/tools/my-tool/run.sh
```

### Template Variables

| Variable | Description |
|----------|-------------|
| `{{param}}` | Parameter value from LLM |
| `{{sandbox_cwd}}` | Current sandbox working directory |
| `{{sandbox_home}}` | Sandbox home directory |
| `{{plugin_dir}}` | Tool's installation directory |

## Creating Sandbox Providers

Implement the `SandboxProvider` interface for new container runtimes.

### Interface

```go
// internal/providers/providers.go
type SandboxProvider interface {
    Provider

    // Create a new sandbox
    Create(ctx context.Context, opts SandboxCreateOptions) (Sandbox, error)

    // Get existing sandbox by ID
    Get(ctx context.Context, id string) (Sandbox, error)

    // List all sandboxes
    List(ctx context.Context) ([]Sandbox, error)

    // Start a stopped sandbox
    Start(ctx context.Context, id string) error

    // Stop a running sandbox
    Stop(ctx context.Context, id string, opts SandboxStopOptions) error

    // Delete a sandbox
    Delete(ctx context.Context, id string, force bool) error

    // Execute command in sandbox
    Exec(ctx context.Context, id string, opts ExecOptions) (ExecResult, error)

    // Get sandbox status
    Status(ctx context.Context, id string) (SandboxStatus, error)

    // Get resource statistics
    Stats(ctx context.Context, id string) (SandboxStats, error)

    // Ensure agent user exists in sandbox
    EnsureAgentUser(ctx context.Context, id string, agentHandle string, dotfilesPath string) error
}
```

### Implementation Example

```go
package myprovider

import (
    "context"
    "github.com/alexcabrera/ayo/internal/providers"
)

type MySandboxProvider struct {
    config Config
}

func New(config Config) *MySandboxProvider {
    return &MySandboxProvider{config: config}
}

func (p *MySandboxProvider) Name() string {
    return "my-sandbox"
}

func (p *MySandboxProvider) Type() providers.ProviderType {
    return providers.ProviderTypeSandbox
}

func (p *MySandboxProvider) Init(ctx context.Context, config map[string]any) error {
    // Initialize provider
    return nil
}

func (p *MySandboxProvider) Close() error {
    // Cleanup resources
    return nil
}

func (p *MySandboxProvider) Create(ctx context.Context, opts providers.SandboxCreateOptions) (providers.Sandbox, error) {
    // Create container
    // ...
    return &MySandbox{id: "..."}, nil
}

func (p *MySandboxProvider) Exec(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
    // Execute command in container
    // ...
    return providers.ExecResult{
        ExitCode: 0,
        Stdout:   "output",
        Stderr:   "",
    }, nil
}

// Implement remaining methods...
```

### Registration

Register via plugin manifest or programmatically:

```json
{
  "providers": [
    {
      "name": "my-sandbox",
      "type": "sandbox",
      "entry_point": "providers/sandbox.so"
    }
  ]
}
```

## Creating Embedding Providers

For custom embedding models (local or API-based).

### Interface

```go
// internal/providers/providers.go
type EmbeddingProvider interface {
    Provider

    // Embed single text
    Embed(ctx context.Context, text string) ([]float32, error)

    // Embed multiple texts (batch)
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

    // Vector dimensions
    Dimensions() int

    // Model name
    Model() string
}
```

### Implementation Example

```go
package local

import (
    "context"
    "github.com/alexcabrera/ayo/internal/providers"
)

type LocalEmbedding struct {
    modelPath string
    dims      int
}

func New(modelPath string) *LocalEmbedding {
    return &LocalEmbedding{
        modelPath: modelPath,
        dims:      384, // depends on model
    }
}

func (e *LocalEmbedding) Name() string {
    return "local-embedding"
}

func (e *LocalEmbedding) Type() providers.ProviderType {
    return providers.ProviderTypeEmbedding
}

func (e *LocalEmbedding) Embed(ctx context.Context, text string) ([]float32, error) {
    // Load model, generate embedding
    // ...
    return embedding, nil
}

func (e *LocalEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
    results := make([][]float32, len(texts))
    for i, text := range texts {
        emb, err := e.Embed(ctx, text)
        if err != nil {
            return nil, err
        }
        results[i] = emb
    }
    return results, nil
}

func (e *LocalEmbedding) Dimensions() int {
    return e.dims
}

func (e *LocalEmbedding) Model() string {
    return "all-MiniLM-L6-v2"
}
```

## Creating Memory Providers

For custom memory storage backends.

### Interface

```go
// internal/providers/providers.go
type MemoryProvider interface {
    Provider

    Create(ctx context.Context, m Memory) (Memory, error)
    Get(ctx context.Context, id string) (Memory, error)
    Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)
    List(ctx context.Context, opts ListOptions) ([]Memory, error)
    Update(ctx context.Context, m Memory) error
    Forget(ctx context.Context, id string) error
    Supersede(ctx context.Context, oldID string, newMemory Memory, reason string) (Memory, error)
    Topics(ctx context.Context) ([]string, error)
    Link(ctx context.Context, id1, id2 string) error
    Unlink(ctx context.Context, id1, id2 string) error
    Reindex(ctx context.Context) error
}
```

### Implementation Considerations

- **Vector search**: Use embedding provider for similarity search
- **Persistence**: Store memories durably (SQLite, file system, etc.)
- **Indexing**: Support full-text and semantic search
- **Relationships**: Implement knowledge graph linking

## Creating Planners

Planners coordinate work through tools injected into agent conversations.

### Interface

```go
// internal/planners/interface.go
type PlannerPlugin interface {
    // Plugin name
    Name() string

    // NearTerm (session-scoped) or LongTerm (persistent)
    Type() PlannerType

    // Initialize with context
    Init(ctx context.Context) error

    // Cleanup
    Close() error

    // Tools this planner provides
    Tools() []fantasy.AgentTool

    // Instructions to inject into system prompt
    Instructions() string

    // Directory for state storage
    StateDir() string
}
```

### Near-term vs Long-term

| Aspect | Near-term | Long-term |
|--------|-----------|-----------|
| Scope | Single session | Persistent |
| Storage | In-memory | Database/files |
| Example | Todo list | Ticket system |
| Reset | On session end | Never (manual) |

### Implementation Example

```go
package myplanner

import (
    "context"
    "github.com/alexcabrera/ayo/internal/fantasy"
    "github.com/alexcabrera/ayo/internal/planners"
)

type MyPlanner struct {
    stateDir string
    tasks    []Task
}

func New(stateDir string) *MyPlanner {
    return &MyPlanner{stateDir: stateDir}
}

func (p *MyPlanner) Name() string {
    return "my-planner"
}

func (p *MyPlanner) Type() planners.PlannerType {
    return planners.NearTerm
}

func (p *MyPlanner) Init(ctx context.Context) error {
    // Load state from stateDir if needed
    return nil
}

func (p *MyPlanner) Close() error {
    // Save state, cleanup
    return nil
}

func (p *MyPlanner) Tools() []fantasy.AgentTool {
    return []fantasy.AgentTool{
        {
            Name:        "add_task",
            Description: "Add a task to the plan",
            Parameters: map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "task": map[string]any{
                        "type":        "string",
                        "description": "Task description",
                    },
                },
                "required": []string{"task"},
            },
            Handler: p.addTask,
        },
        {
            Name:        "complete_task",
            Description: "Mark a task as complete",
            Parameters: map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "task_id": map[string]any{
                        "type":        "string",
                        "description": "Task ID",
                    },
                },
                "required": []string{"task_id"},
            },
            Handler: p.completeTask,
        },
    }
}

func (p *MyPlanner) Instructions() string {
    return `
## Task Management

You have access to task management tools:
- Use add_task to track work items
- Use complete_task when finished

Keep tasks focused and achievable.
`
}

func (p *MyPlanner) StateDir() string {
    return p.stateDir
}

func (p *MyPlanner) addTask(params map[string]any) (any, error) {
    task := params["task"].(string)
    // Add to tasks list
    return map[string]string{"status": "added", "task": task}, nil
}

func (p *MyPlanner) completeTask(params map[string]any) (any, error) {
    taskID := params["task_id"].(string)
    // Mark complete
    return map[string]string{"status": "completed", "task_id": taskID}, nil
}
```

### Building as Plugin

```bash
go build -buildmode=plugin -o my-planner.so planner.go
```

Register in manifest:
```json
{
  "planners": [
    {
      "name": "my-planner",
      "type": "near",
      "entry_point": "planners/my-planner.so"
    }
  ]
}
```

## Creating Trigger Plugins

Triggers invoke agents on events.

### Trigger Types

| Type | Description |
|------|-------------|
| `poll` | Periodic polling (cron, interval) |
| `push` | External events (webhooks) |
| `watch` | File system changes |

### Configuration Schema

```json
{
  "triggers": [
    {
      "name": "my-trigger",
      "category": "poll",
      "entry_point": "triggers/my-trigger.so",
      "config_schema": {
        "type": "object",
        "properties": {
          "endpoint": {"type": "string"},
          "interval": {"type": "string"}
        },
        "required": ["endpoint"]
      }
    }
  ]
}
```

## Contributing to Core

### Development Setup

```bash
# Clone repository
git clone https://github.com/alexcabrera/ayo.git
cd ayo

# Build
go build ./cmd/ayo/...

# Test
go test ./...

# Lint
golangci-lint run
```

### Code Organization

| Directory | Purpose |
|-----------|---------|
| `cmd/ayo/` | CLI commands |
| `internal/agent/` | Agent loading, config |
| `internal/providers/` | Provider interfaces and implementations |
| `internal/sandbox/` | Sandbox provider implementations |
| `internal/memory/` | Memory system |
| `internal/daemon/` | Background service |
| `internal/planners/` | Planner system |
| `internal/tools/` | Built-in tools |
| `internal/squads/` | Squad management |

### Coding Standards

- Use `golangci-lint` before committing
- Follow existing patterns in similar code
- Add tests for new functionality
- Document exported functions
- Use `context.Context` for cancellation

### Pull Request Process

1. Fork the repository
2. Create feature branch: `git checkout -b feature/my-feature`
3. Make changes with tests
4. Run `go test ./...` and `golangci-lint run`
5. Commit with clear message
6. Push and create PR
7. Respond to review feedback

### Testing

```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/memory/...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...
```

## Plugin Distribution

### Git Repository

Name repositories `ayo-plugins-<name>`:

```bash
# Create repository
mkdir ayo-plugins-mytools
cd ayo-plugins-mytools
git init

# Add manifest and components
# ...

# Push
git remote add origin https://github.com/user/ayo-plugins-mytools.git
git push -u origin main

# Tag release
git tag v1.0.0
git push --tags
```

Users install via:
```bash
ayo plugin install https://github.com/user/ayo-plugins-mytools
```

### Local Development

Install from local path during development:
```bash
ayo plugin install ~/dev/my-plugin
```

## See Also

- [Architecture](architecture.md) - System internals
- [Plugin Reference](../reference/plugins.md) - Manifest schema
- [Troubleshooting](troubleshooting.md) - Debugging guide
