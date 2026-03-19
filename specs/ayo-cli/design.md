# Ayo CLI Design

## Overview

Ayo compiles AI agent definitions into standalone, dependency-free CLI executables. Agents are defined by a directory convention (`config.toml`, `system.md`, `prompt.tmpl`, `input.jsonschema`, `output.jsonschema`, `skills/`, `hooks/`) and compiled to native binaries backed by the Fantasy multi-provider abstraction layer.

## Detailed Requirements

### Agent Definition Structure

```
<agent-name>/
  config.toml         # Required: metadata, model requirements, defaults
  system.md           # Required: system message governing agent behavior
  prompt.tmpl         # Optional: Go template for rendering prompts
  input.jsonschema    # Optional: defines CLI interface and input structure
  output.jsonschema   # Optional: defines structured output schema
  skills/             # Optional: Agent Skills compatible packages
  hooks/              # Optional: lifecycle hook executables
```

### ayo CLI Commands

| Command | Description |
|---------|-------------|
| `ayo fresh <name>` | Create new agent project from template |
| `ayo build [path]` | Compile project into standalone executable |
| `ayo checkit [path]` | Validate project integrity |

### Generated Executable Behavior

- **Self-contained**: No runtime dependencies (no Go runtime required)
- **First-run model selection**: Bubbletea TUI if no config exists
- **Config location**: `~/.config/agents/<agent-name>.toml`
- **Non-interactive first**: All functionality via flags

### Input/Output Matrix

| input.jsonschema | prompt.tmpl | Input Behavior | Prompt Behavior |
|------------------|-------------|----------------|-----------------|
| ✓ | ✓ | CLI parses args/flags | Template renders structured data |
| ✓ | ✗ | JSON input via stdin/arg | JSON sent directly as prompt |
| ✗ | ✓ | Single freeform text arg | Template uses `{{ .input }}` |
| ✗ | ✗ | Single freeform text arg | Raw text sent as prompt |

---

## Architecture Overview

```mermaid
graph TB
    subgraph "ayo CLI"
        fresh[ayo fresh]
        build[ayo build]
        checkit[ayo checkit]
    end

    subgraph "Agent Definition"
        config[config.toml]
        system[system.md]
        prompt[prompt.tmpl]
        input[input.jsonschema]
        output[output.jsonschema]
        skills[skills/]
        hooks[hooks/]
    end

    subgraph "Build Process"
        validate[Validate Structure]
        generate[Generate Go Code]
        embed[Embed Assets]
        compile[Go Build]
    end

    subgraph "Generated Binary"
        cli[CLI Parser]
        tui[Model Selection TUI]
        renderer[Template Renderer]
        agent[Agent Runtime]
        hookrunner[Hook Runner]
    end

    subgraph "Runtime Dependencies"
        fantasy[Fantasy Library]
        catwalk[Catwalk Registry]
        bubbletea[Bubbletea TUI]
    end

    fresh --> Agent Definition
    checkit --> validate
    build --> validate
    validate --> generate
    config --> generate
    input --> generate
    output --> generate
    generate --> embed
    system --> embed
    prompt --> embed
    skills --> embed
    hooks --> embed
    embed --> compile
    compile --> Generated Binary

    cli --> tui
    tui --> catwalk
    cli --> renderer
    prompt --> renderer
    renderer --> agent
    system --> agent
    fantasy --> agent
    hookrunner --> hooks
    agent --> hookrunner
```

---

## Components and Interfaces

### 1. Project Parser

**Responsibility**: Parse and validate agent definition directory.

```go
type Project struct {
    Path    string
    Config  AgentConfig
    System  string
    Prompt  *string
    Input   *Schema
    Output  *Schema
    Skills  []Skill
    Hooks   map[HookType]string
}

type AgentConfig struct {
    Name        string
    Version     string
    Description string
    Model       ModelRequirements
    Defaults    AgentDefaults
}

type ModelRequirements struct {
    RequiresStructuredOutput bool
    RequiresTools            bool
    RequiresVision           bool
    Suggested                []string
    Default                  string
}

func ParseProject(path string) (*Project, error)
func ValidateProject(p *Project) []error
```

### 2. Code Generator

**Responsibility**: Generate Go source code from agent definition.

```go
type CodeGenerator interface {
    GenerateCLI(p *Project) ([]byte, error)      // Cobra command definitions
    GenerateTypes(p *Project) ([]byte, error)    // Go structs from schemas
    GenerateHooks(p *Project) ([]byte, error)    // Hook runner code
    GenerateAgent(p *Project) ([]byte, error)    // Agent setup code
    GenerateMain(p *Project) ([]byte, error)     // Main entry point
}
```

### 3. Schema Parser

**Responsibility**: Parse JSON Schema and generate CLI flags/types.

```go
type SchemaParser interface {
    Parse(schema []byte) (*ParsedSchema, error)
    GenerateFlags(schema *ParsedSchema) []FlagDef
    GenerateStruct(schema *ParsedSchema) (string, error)
}

type FlagDef struct {
    Name         string
    ShortName    string
    Type         string
    DefaultValue interface{}
    Description  string
    Position     int  // 0 for flags, >0 for positional
    IsFile       bool
}
```

### 4. Template Renderer

**Responsibility**: Render prompt templates with input data.

```go
type PromptRenderer interface {
    Render(template string, data map[string]any) (string, error)
}
```

Uses Go's `text/template` with custom functions.

### 5. Hook Runner

**Responsibility**: Execute lifecycle hooks at appropriate times.

```go
type HookType string

const (
    HookAgentStart       HookType = "agent-start"
    HookAgentFinish      HookType = "agent-finish"
    HookAgentError       HookType = "agent-error"
    HookStepStart        HookType = "step-start"
    HookStepFinish       HookType = "step-finish"
    HookTextStart        HookType = "text-start"
    HookTextDelta        HookType = "text-delta"
    HookTextEnd          HookType = "text-end"
    HookReasoningStart   HookType = "reasoning-start"
    HookReasoningDelta   HookType = "reasoning-delta"
    HookReasoningEnd     HookType = "reasoning-end"
    HookToolInputStart   HookType = "tool-input-start"
    HookToolInputDelta   HookType = "tool-input-delta"
    HookToolInputEnd     HookType = "tool-input-end"
    HookToolCall         HookType = "tool-call"
    HookToolResult       HookType = "tool-result"
    HookSource           HookType = "source"
    HookStreamFinish     HookType = "stream-finish"
    HookWarnings         HookType = "warnings"
)

type HookRunner interface {
    Run(ctx context.Context, hookType HookType, payload any) error
}

type EmbeddedHookRunner struct {
    embeddedHooks map[HookType][]byte  // from embed.FS
    userHooks     map[HookType]string  // from config
}
```

**Execution order:**
1. Run embedded hook (if exists and executable)
2. Run user hook (if configured)
3. Both must complete before flow continues

### 6. Model Selector

**Responsibility**: First-run model selection via TUI or flags.

```go
type ModelSelector interface {
    Select(ctx context.Context, reqs ModelRequirements) (*ModelSelection, error)
}

type ModelSelection struct {
    Provider string
    Model    string
    APIKey   string
}

type TUIModelSelector struct {
    providers []catwalk.Provider
}

type FlagModelSelector struct {
    Provider string
    Model    string
}
```

**Environment variable detection:**

| Variable | Provider |
|----------|----------|
| `ANTHROPIC_API_KEY` | Anthropic |
| `OPENAI_API_KEY` | OpenAI |
| `GEMINI_API_KEY` | Google Gemini |
| `GROQ_API_KEY` | Groq |
| `OPENROUTER_API_KEY` | OpenRouter |
| `VERCEL_API_KEY` | Vercel AI Gateway |
| `CEREBRAS_API_KEY` | Cerebras |
| `HF_TOKEN` | Hugging Face |
| `AZURE_OPENAI_API_KEY` | Azure OpenAI |
| `AWS_ACCESS_KEY_ID` + `AWS_SECRET_ACCESS_KEY` | Amazon Bedrock |
| `VERTEXAI_PROJECT` | Google VertexAI |

### 7. Config Manager

**Responsibility**: Manage user configuration files.

```go
type ConfigManager interface {
    Load(name string) (*UserConfig, error)
    Save(name string, cfg *UserConfig) error
    Exists(name string) bool
}

type UserConfig struct {
    Provider string
    Model    string
    APIKey   string
    Hooks    map[HookType]string
}
```

Config path: `~/.config/agents/<agent-name>.toml`

---

## Data Models

### config.toml

```toml
[agent]
name = "code-reviewer"
version = "1.0.0"
description = "Reviews code for quality, bugs, and improvements"

[model]
requires_structured_output = false
requires_tools = true
requires_vision = false
suggested = ["claude-sonnet-4-6", "gpt-4o", "gemini-2.5-pro"]
default = "claude-sonnet-4-6"

[defaults]
temperature = 0.7
max_tokens = 4096
```

### input.jsonschema (with CLI extensions)

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "source": {
      "type": "string",
      "description": "Source code or file path to analyze",
      "x-cli-position": 1,
      "x-cli-file": true
    },
    "language": {
      "type": "string",
      "description": "Programming language",
      "x-cli-flag": "--language",
      "x-cli-short": "-l",
      "default": "auto",
      "enum": ["auto", "go", "python", "javascript", "rust"]
    },
    "strict": {
      "type": "boolean",
      "description": "Enable strict mode",
      "x-cli-flag": "--strict",
      "default": false
    }
  },
  "required": ["source"]
}
```

### output.jsonschema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "score": {
      "type": "integer",
      "minimum": 0,
      "maximum": 100,
      "description": "Overall code quality score"
    },
    "issues": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "severity": {"type": "string", "enum": ["error", "warning", "info"]},
          "message": {"type": "string"},
          "line": {"type": "integer"}
        },
        "required": ["severity", "message"]
      }
    },
    "suggestions": {
      "type": "array",
      "items": {"type": "string"}
    }
  },
  "required": ["score", "issues"]
}
```

### Hook Payload (JSON on stdin)

```json
{
  "event": "tool-call",
  "timestamp": "2025-03-18T12:00:00Z",
  "data": {
    "tool_call_id": "call_123",
    "tool_name": "read_file",
    "input": "{\"path\": \"/src/main.go\"}"
  }
}
```

---

## Error Handling

| Error Type | Handling |
|------------|----------|
| Invalid project structure | `ayo checkit` reports all issues, exits non-zero |
| Missing required files | Build fails with clear error message |
| Schema parse errors | Build fails with line/column info |
| Template parse errors | Build fails with template error location |
| Hook execution failure | Log warning, continue execution |
| Model selection cancelled | Exit 0 (user cancelled) |
| No providers available | TUI shows message, suggests setting API keys |
| API key invalid | Runtime error from provider, clear message |

---

## Acceptance Criteria

### AC1: Create new agent project
```
Given I run `ayo fresh my-agent`
When the command completes
Then a directory `my-agent/` exists
And it contains `config.toml` and `system.md`
And config.toml has the agent name set to "my-agent"
```

### AC2: Build generates executable
```
Given a valid agent project at `./my-agent`
When I run `ayo build ./my-agent`
Then an executable `my-agent` is created
And running `./my-agent --help` shows usage
```

### AC3: First-run model selection
```
Given a newly built agent binary
And no config exists at ~/.config/agents/my-agent.toml
When I run `./my-agent "test prompt"` without --provider/--model
Then a TUI appears for model selection
And after selection, config is saved
And the agent runs with selected model
```

### AC4: Non-interactive model selection
```
Given a newly built agent binary
When I run `./my-agent --provider anthropic --model claude-sonnet-4-6 "test"`
Then no TUI appears
And config is saved with specified values
And the agent runs with specified model
```

### AC5: Input schema generates CLI
```
Given an agent with input.jsonschema defining:
  - positional "source" (position 1)
  - flag "--language" with short "-l"
When I run `./my-agent src/main.go -l go`
Then the CLI parses these correctly
And passes {"source": "src/main.go", "language": "go"} to the template
```

### AC6: Output schema produces structured output
```
Given an agent with output.jsonschema
When I run `./my-agent "analyze this"`
Then output is valid JSON conforming to the schema
And output goes to stdout
```

### AC7: Hooks execute in order
```
Given an agent with embedded hooks
And user config with additional hooks
When the agent runs
Then embedded hooks run first for each event
Then user hooks run second
And execution blocks until both complete
```

### AC8: Skills are embedded and discoverable
```
Given an agent with skills/ containing skill-a/SKILL.md
When I build the agent
Then the skill is embedded in the binary
And at runtime, system message includes skill paths
```

### AC9: Template rendering
```
Given an agent with prompt.tmpl containing {{.name}}
And input.jsonschema defining "name" property
When I run `./my-agent --name Alice`
Then the template renders with "Alice" substituted
And rendered text is sent as the prompt
```

### AC10: Binary is self-contained
```
Given a built agent binary
When I copy it to a machine without Go installed
And run it
Then it executes successfully without errors about missing dependencies
```

---

## Testing Strategy

### Unit Tests
- Schema parser: JSON Schema → CLI flags
- Template renderer: template + data → prompt
- Config manager: load/save user config
- Hook runner: execute and capture output

### Integration Tests
- `ayo fresh` → validate project structure
- `ayo checkit` → valid/invalid projects
- `ayo build` → generates working binary
- Generated binary → CLI parsing, model selection

### End-to-End Tests
- Full agent execution with mock provider
- Hook execution verification
- Skills discovery and loading
- Output schema validation

---

## Appendices

### A. Technology Choices

| Choice | Rationale |
|--------|-----------|
| Go | Native compilation, single binary output, Fantasy/Catwalk written in Go |
| Fantasy | Multi-provider abstraction, structured outputs, streaming, tools |
| Catwalk | Provider/model registry with feature metadata |
| Bubbletea | Declarative TUI framework, well-maintained |
| Cobra | Industry standard CLI library |
| Fang | Styled help/errors for Cobra |
| embed.FS | Native Go embedding, no external dependencies |

### B. JSON Schema Extensions

| Extension | Purpose |
|-----------|---------|
| `x-cli-position` | Positional argument order (1, 2, 3...) |
| `x-cli-flag` | Long flag name (e.g., `--language`) |
| `x-cli-short` | Short flag name (e.g., `-l`) |
| `x-cli-file` | Treat string as file path (validate exists) |

### C. Hook Event Types (Full List)

| Hook | Fantasy Callback | Payload Fields |
|------|------------------|----------------|
| `agent-start` | `OnAgentStart` | `{}` |
| `agent-finish` | `OnAgentFinish` | `{result}` |
| `agent-error` | `OnError` | `{error}` |
| `step-start` | `OnStepStart` | `{step_number}` |
| `step-finish` | `OnStepFinish` | `{step_result}` |
| `text-start` | `OnTextStart` | `{id}` |
| `text-delta` | `OnTextDelta` | `{id, delta}` |
| `text-end` | `OnTextEnd` | `{id}` |
| `reasoning-start` | `OnReasoningStart` | `{id, reasoning}` |
| `reasoning-delta` | `OnReasoningDelta` | `{id, delta}` |
| `reasoning-end` | `OnReasoningEnd` | `{id, reasoning}` |
| `tool-input-start` | `OnToolInputStart` | `{id, tool_name}` |
| `tool-input-delta` | `OnToolInputDelta` | `{id, delta}` |
| `tool-input-end` | `OnToolInputEnd` | `{id}` |
| `tool-call` | `OnToolCall` | `{tool_call}` |
| `tool-result` | `OnToolResult` | `{result}` |
| `source` | `OnSource` | `{source}` |
| `stream-finish` | `OnStreamFinish` | `{usage, finish_reason}` |
| `warnings` | `OnWarnings` | `{warnings}` |
