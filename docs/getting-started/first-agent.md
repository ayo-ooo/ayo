# First Agent

A complete walkthrough of agent structure and configuration.

## Project Files

An Ayo agent project can include these files:

|| File | Required | Purpose |
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

|| Field | Effect |
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

Defines inputs using JSON Schema:

```json
{
  "type": "object",
  "properties": {
    "text": {
      "type": "string",
      "description": "Input text"
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

### CLI Properties

|| Property | Purpose |
|----------|---------|
| `flag` | Custom flag name (default: kebab-case of property name) |
| `file` | Set to `true` to load file content into field |

### Input Patterns

Generated CLIs accept JSON input as the primary mechanism:

```bash
# Inline JSON
./my-agent '{"text": "hello"}'

# From file
./my-agent input.json

# From stdin
echo '{"text": "hello"}' | ./my-agent -

# Flag overrides
./my-agent --text "hello" --format json
```

Only primitive types (string, integer, number, boolean) get flag overrides.

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

|| Function | Description |
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
ayo runthat .

# Run with help
./my-agent --help

# Run with JSON input
./my-agent '{"text": "input text", "format": "json"}'

# Run with flag overrides
./my-agent --text "input text" --format json --verbose
```

## Next Steps

- [Input Schema](../reference/input-schema.md) - Detailed schema reference
- [Output Schema](../reference/output-schema.md) - Output type definitions
- [Skills](../reference/skills.md) - Add modular capabilities
- [Hooks](../reference/hooks.md) - Handle lifecycle events
