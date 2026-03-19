# First Agent

A complete walkthrough of agent structure and configuration.

## Project Files

An Ayo agent project can include these files:

| File | Required | Purpose |
|------|----------|---------|
| `config.toml` | Yes | Agent metadata and model settings |
| `system.md` | Yes | System prompt for the LLM |
| `input.jsonschema` | No | Define input types and CLI flags |
| `output.jsonschema` | No | Define output structure |
| `prompt.tmpl` | No | Dynamic prompt template |
| `skills/` | No | Reusable skill modules |
| `hooks/` | No | Lifecycle event handlers |

## config.toml

The main configuration file:

```toml
[agent]
name = "my-agent"           # Required: agent name
version = "1.0.0"           # Required: semantic version
description = "Description" # Required: short description

[model]
# Model requirements
requires_structured_output = false  # Needs JSON mode
requires_tools = false              # Needs function calling
requires_vision = false             # Needs image understanding
suggested = [                       # Suggested models
  "anthropic/claude-3.5-sonnet",
  "openai/gpt-4o"
]
default = "anthropic/claude-3.5-sonnet"

[defaults]
# Default generation parameters
temperature = 0.7
max_tokens = 2048
```

### Model Requirements

| Field | Effect |
|-------|--------|
| `requires_structured_output` | Agent needs JSON output mode |
| `requires_tools` | Agent uses function calling |
| `requires_vision` | Agent processes images |

When `output.jsonschema` exists, `requires_structured_output` is implied.

## system.md

The system prompt defines the agent's behavior:

```markdown
# Agent Name

Brief description of the agent's purpose.

## Capabilities

- What the agent can do
- How it should respond

## Guidelines

- Tone and style preferences
- Response format expectations

## Available Skills

If you have skills, list them here so the LLM knows about them.
```

## input.jsonschema

Defines inputs using JSON Schema with CLI extensions:

```json
{
  "type": "object",
  "properties": {
    "text": {
      "type": "string",
      "description": "Input text",
      "x-cli-position": 1
    },
    "format": {
      "type": "string",
      "description": "Output format",
      "enum": ["json", "text", "markdown"],
      "default": "text"
    },
    "verbose": {
      "type": "boolean",
      "description": "Enable verbose output",
      "default": false
    }
  },
  "required": ["text"]
}
```

### CLI Extensions

| Extension | Purpose |
|-----------|---------|
| `x-cli-position` | Make a positional argument (1-indexed) |
| `x-cli-flag` | Custom flag name |
| `x-cli-short` | Short flag (e.g., `-f`) |
| `x-cli-file` | Load file content into field |

## output.jsonschema

Defines the response structure:

```json
{
  "type": "object",
  "properties": {
    "result": {
      "type": "string"
    },
    "metadata": {
      "type": "object",
      "properties": {
        "tokens_used": { "type": "integer" },
        "processing_time_ms": { "type": "integer" }
      }
    }
  }
}
```

When present, the LLM is instructed to respond with JSON matching this schema.

## prompt.tmpl

Optional template for dynamic prompts:

```gotemplate
Process the following: {{.text}}

{{if .verbose}}Provide detailed analysis.{{end}}

{{if .context_file}}Context:
{{file .context_file}}
{{end}}

Output format: {{.format}}
```

### Template Functions

| Function | Description |
|----------|-------------|
| `{{.field}}` | Access input field |
| `{{file "path"}}` | Load file contents |
| `{{env "VAR"}}` | Get environment variable |
| `{{upper .text}}` | Uppercase |
| `{{lower .text}}` | Lowercase |
| `{{title .text}}` | Title case |
| `{{trim .text}}` | Trim whitespace |
| `{{json .data}}` | JSON encode |

## Build and Run

```bash
# Build the agent
ayo build .

# Run with help
./my-agent --help

# Run with inputs
./my-agent "input text" --format json --verbose
```

## Next Steps

- [Input Schema](../reference/input-schema.md) - Detailed schema reference
- [Output Schema](../reference/output-schema.md) - Output type definitions
- [Skills](../reference/skills.md) - Add modular capabilities
- [Hooks](../reference/hooks.md) - Handle lifecycle events
