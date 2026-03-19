# API Reference

Auto-documented types for generated agents.

## Standard Flags

All generated agents include these flags:

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` | - | Show help |
| `-o, --output` | string | Write output to file |
| `--provider` | string | AI provider to use |
| `--model` | string | Model to use |

## Input Types

Generated from `input.jsonschema`:

```go
type Input struct {
    // Generated fields based on schema
    Field1 string `json:"field1"`
    Field2 int    `json:"field2"`
    Field3 bool   `json:"field3"`
}
```

## Output Types

Generated from `output.jsonschema`:

```go
type Output struct {
    // Generated fields based on schema
    Result  string   `json:"result"`
    Items   []Item   `json:"items"`
    Score   int      `json:"score"`
}
```

## Hook Types

Available hook types:

```go
type HookType string

const (
    HookAgentStart  HookType = "agent-start"
    HookAgentFinish HookType = "agent-finish"
    HookAgentError  HookType = "agent-error"
    HookStepStart   HookType = "step-start"
    HookStepFinish  HookType = "step-finish"
    HookTextStart   HookType = "text-start"
    HookTextDelta   HookType = "text-delta"
    HookTextEnd     HookType = "text-end"
    HookToolCall    HookType = "tool-call"
    HookToolResult  HookType = "tool-result"
)
```

## Hook Runner

```go
type HookRunner struct {
    embeddedHooks map[HookType][]byte
    userHooks     map[HookType]string
    tempDir       string
}

func NewHookRunner(embedded, user map[HookType]string, tempDir string) *HookRunner
func (r *HookRunner) Run(ctx context.Context, hookType HookType, data any) error
```

## Configuration Types

```go
type Config struct {
    Agent    AgentConfig    `toml:"agent"`
    Model    ModelConfig    `toml:"model"`
    Defaults DefaultsConfig `toml:"defaults"`
}

type AgentConfig struct {
    Name        string `toml:"name"`
    Version     string `toml:"version"`
    Description string `toml:"description"`
}

type ModelConfig struct {
    RequiresStructuredOutput bool     `toml:"requires_structured_output"`
    RequiresTools            bool     `toml:"requires_tools"`
    RequiresVision           bool     `toml:"requires_vision"`
    Suggested                []string `toml:"suggested"`
    Default                  string   `toml:"default"`
}

type DefaultsConfig struct {
    Temperature float64 `toml:"temperature"`
    MaxTokens   int     `toml:"max_tokens"`
}
```

## See Also

- [Generated Code](../reference/generated-code.md) - Understanding generated files
- [Configuration](../reference/config.md) - Configuration options
