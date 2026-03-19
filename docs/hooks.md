# Hook System

Hooks allow custom scripts to run at specific points during agent execution. They receive structured JSON payloads via stdin and can perform actions like logging, notifications, or data transformation.

## Hook Types

| Hook Type | When Triggered | Data Provided |
|-----------|---------------|---------------|
| `agent-start` | Before agent execution begins | Input data, configuration |
| `agent-finish` | After successful completion | Output data, duration |
| `agent-error` | On execution failure | Error message, stack trace |
| `text-start` | Before text generation | Prompt preview |
| `text-delta` | During streaming | Text chunk |
| `text-end` | After text generation complete | Full response |
| `tool-call` | Before tool invocation | Tool name, arguments |
| `tool-result` | After tool returns | Tool output |

## Execution Order

Hooks execute in a specific order:

1. **Embedded hooks** - Scripts bundled with the agent (in `hooks/` directory)
2. **User hooks** - Scripts configured by the user at runtime

```go
// Run executes hooks for the given hook type
func (r *HookRunner) Run(ctx context.Context, hookType HookType, data any) error {
    // 1. Run embedded hook if exists
    if embedded, ok := r.embeddedHooks[hookType]; ok {
        r.executeEmbedded(ctx, hookType, embedded, payload)
    }

    // 2. Run user hook if configured
    if userPath, ok := r.userHooks[hookType]; ok {
        r.executeUser(ctx, userPath, payload)
    }

    return nil
}
```

## Payload Format

All hooks receive JSON via stdin with this base structure:

```json
{
    "event": "agent-start",
    "timestamp": "2024-01-15T10:30:00Z",
    "data": { ... }
}
```

### agent-start

```json
{
    "event": "agent-start",
    "timestamp": "2024-01-15T10:30:00Z",
    "data": {
        "agent": "my-agent",
        "provider": "anthropic",
        "model": "claude-3-5-sonnet",
        "input": { ... }
    }
}
```

### agent-finish

```json
{
    "event": "agent-finish",
    "timestamp": "2024-01-15T10:30:05Z",
    "data": {
        "agent": "my-agent",
        "duration_ms": 5000,
        "output": { ... }
    }
}
```

### agent-error

```json
{
    "event": "agent-error",
    "timestamp": "2024-01-15T10:30:03Z",
    "data": {
        "agent": "my-agent",
        "error": "API rate limit exceeded",
        "duration_ms": 3000
    }
}
```

### text-delta

```json
{
    "event": "text-delta",
    "timestamp": "2024-01-15T10:30:01Z",
    "data": {
        "delta": "Hello, "
    }
}
```

### tool-call

```json
{
    "event": "tool-call",
    "timestamp": "2024-01-15T10:30:02Z",
    "data": {
        "tool": "search",
        "arguments": {
            "query": "example search"
        }
    }
}
```

### tool-result

```json
{
    "event": "tool-result",
    "timestamp": "2024-01-15T10:30:02Z",
    "data": {
        "tool": "search",
        "result": "Found 10 results"
    }
}
```

## Blocking Behavior

Hooks are **non-blocking** by default:

- Hook failures are logged but don't stop agent execution
- Each hook runs independently
- Context cancellation is respected for long-running hooks

```go
// Hook failures are logged but don't block
if err := r.executeEmbedded(ctx, hookType, embedded, payload); err != nil {
    log.Printf("embedded hook %s failed: %v", hookType, err)
    // Execution continues
}
```

### Timeout Handling

Hooks inherit the context from the agent execution:

- If agent times out, hooks are cancelled
- Use `context.Context` in hook implementations for cleanup

## Hook Configuration

### Embedded Hooks

Place executable scripts in the agent's `hooks/` directory:

```
my-agent/
├── agent.rive
├── hooks/
│   ├── agent-start      # Runs on agent-start
│   ├── agent-finish     # Runs on agent-finish
│   └── text-delta       # Runs on text-delta
└── ...
```

The file name must match the hook type.

### User Hooks

Configure in agent config file (`~/.config/agents/{agent-name}.toml`):

```toml
provider = "anthropic"
model = "claude-3-5-sonnet"

[hooks]
agent-start = "/path/to/my/start-hook.sh"
agent-finish = "/path/to/my/finish-hook.sh"
```

## Implementation Details

### HookRunner Interface

```go
type HookRunner struct {
    embeddedHooks map[HookType][]byte
    userHooks     map[HookType]string
    tempDir       string
}

func NewHookRunner(embedded, user map[HookType]string, tempDir string) *HookRunner
func (r *HookRunner) Run(ctx context.Context, hookType HookType, data any) error
```

### Embedded Hook Execution

Embedded hooks are extracted to a temp file and executed:

```go
func (r *HookRunner) executeEmbedded(ctx context.Context, hookType HookType, content []byte, payload []byte) error {
    tmpFile := filepath.Join(r.tempDir, fmt.Sprintf("hook-%s", hookType))
    os.WriteFile(tmpFile, content, 0755)
    defer os.Remove(tmpFile)
    return r.executeHook(ctx, tmpFile, payload)
}
```

### User Hook Execution

User hooks are executed directly from their configured path:

```go
func (r *HookRunner) executeUser(ctx context.Context, hookPath string, payload []byte) error {
    return r.executeHook(ctx, hookPath, payload)
}
```

## Example Hook Script

```bash
#!/bin/bash
# hooks/agent-finish - Log completion to file

read -r payload
echo "$(date -Iseconds): $(echo "$payload" | jq -r '.data.agent') completed in $(echo "$payload" | jq -r '.data.duration_ms')ms" >> /var/log/agents.log
```

## Current Status

Hook system code is generated but not yet wired into agent execution. The `GenerateHooks` function creates the hook runner, but it needs to be:

1. Initialized in `runAgent` via `initHooks()`
2. Called at appropriate points during execution
3. Wired into streaming callbacks for delta events
