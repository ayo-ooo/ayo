# ayo-create: Build ayo agents from scratch

Use this skill when the user wants to create a new AI agent, CLI tool powered by an LLM, or automate a task using ayo.

## Overview

Ayo compiles agent definitions (plain files) into standalone, zero-dependency CLI binaries. You create the definition files, then `ayo runthat` compiles them into a distributable executable.

## Workflow

1. `ayo fresh <name>` — scaffold a new agent project
2. Edit the generated files to define agent behavior
3. `ayo checkit .` — validate the project
4. `ayo runthat . --register` — compile and register in the ayo registry

## Project Structure

```
my-agent/
  config.toml           # Required: agent metadata & model config
  system.md             # Required: system prompt
  input.jsonschema      # Optional: makes it a "tool agent" with forms
  output.jsonschema     # Optional: structured JSON output
  prompt.tmpl           # Optional: Go template for dynamic prompts
  skills/               # Optional: embedded capability docs
    skill-name/
      SKILL.md
  hooks/                # Optional: lifecycle event scripts
    agent-start
    agent-finish
    agent-error
```

## config.toml Reference

```toml
[agent]
name = "my-agent"              # Agent name (becomes binary name)
version = "1.0.0"              # Semantic version
description = "What it does"   # Short description
interactive = true             # Enable TUI forms (default: true)
input_order = ["field1", "field2"]  # Custom form field ordering

[model]
requires_structured_output = false  # Needs JSON mode
requires_tools = false              # Needs tool calling
requires_vision = false             # Needs image input
suggested = ["claude-sonnet-4-6", "gpt-4o"]  # Recommended models
default = "claude-sonnet-4-6"       # Default model

[defaults]
temperature = 0.7    # 0.0-1.0 (lower = more deterministic)
max_tokens = 4096    # Max response tokens
```

## system.md Best Practices

The system prompt defines agent personality and behavior. Structure it clearly:

```markdown
# Agent Name

Brief role description.

## Behavior

- What the agent does
- How it responds
- What it avoids

## Output Format

How to structure responses.
```

Tips:
- Be specific about the agent's role and boundaries
- Include examples of expected output
- Use markdown headers to organize sections
- Reference skills if the agent has them

## input.jsonschema

Presence of this file makes the agent a "tool agent" with auto-generated TUI forms and CLI flags. Without it, the agent is "conversational" with --chat and --session support.

### Supported Types and Their TUI Mapping

```json
{
  "type": "object",
  "properties": {
    "text": {
      "type": "string",
      "description": "Text input"
    },
    "language": {
      "type": "string",
      "enum": ["go", "python", "typescript"],
      "description": "Select from options"
    },
    "count": {
      "type": "integer",
      "description": "Numeric input",
      "default": 10
    },
    "score": {
      "type": "number",
      "description": "Float input"
    },
    "verbose": {
      "type": "boolean",
      "description": "Toggle",
      "default": false
    }
  },
  "required": ["text"]
}
```

- `string` -> text input field / CLI flag
- `string` with `enum` -> select dropdown
- `integer` / `number` -> numeric input
- `boolean` -> confirm toggle
- `required` array -> validated as non-empty

### Special Properties

- `"flag": "custom-name"` — custom CLI flag name
- `"file": true` — read file contents into this field

## output.jsonschema

Defines structured JSON output. The LLM is instructed to respond in this format:

```json
{
  "type": "object",
  "properties": {
    "summary": { "type": "string" },
    "score": { "type": "number" },
    "issues": {
      "type": "array",
      "items": { "type": "string" }
    }
  },
  "required": ["summary", "score"]
}
```

## prompt.tmpl

Go template for dynamic prompts. Available functions:

- `{{.input}}` — access input data
- `{{file "path"}}` — read file contents
- `{{env "VAR"}}` — environment variable
- `{{upper .text}}`, `{{lower .text}}`, `{{title .text}}` — string transforms
- `{{json .data}}` — JSON encode
- `{{trim .text}}` — trim whitespace

Example:
```
Analyze the following {{.input.Language}} code:

{{file .input.Path}}

Focus on: {{.input.Focus}}
```

## Skills

Skills are embedded documentation that extend the agent's capabilities. Create a directory under `skills/` with a `SKILL.md`:

```
skills/
  analyze/
    SKILL.md        # Describes when and how to use this skill
    scripts/        # Optional supporting files
```

## Hooks

Lifecycle event handlers (shell scripts). Receive JSON payload on stdin:

- `agent-start` — before agent runs
- `agent-finish` — after completion
- `agent-error` — on error
- `step-start` / `step-finish` — per-step events
- `text-start` / `text-delta` / `text-end` — streaming events
- `tool-call` / `tool-result` — tool use events

## Agent Types

### Tool Agent (has input.jsonschema)
- One-shot execution
- Auto-generated TUI forms for missing fields
- CLI flags from schema properties
- Structured I/O

### Conversational Agent (no input.jsonschema)
- Interactive chat via `--chat`
- Session persistence via `--session <id>`
- Free-form text input
- Streaming responses

## Generated Binary Features

Every compiled agent includes:
- Multi-provider LLM support (Anthropic, OpenAI, OpenRouter, Groq, Gemini, etc.)
- First-run model selection TUI
- Config persistence at `~/.config/agents/<name>.toml`
- Sandboxed POSIX shell tool
- JSON input via args, flags, or stdin pipe
- `--output` / `-o` flag for file output
- `--non-interactive` flag for scripted use
- `--ayo-describe` flag for self-description (used by registry)

## Examples

### Code Reviewer (tool agent)

config.toml:
```toml
[agent]
name = "code-reviewer"
version = "1.0.0"
description = "Reviews code for bugs, security issues, and style"

[model]
default = "claude-sonnet-4-6"

[defaults]
temperature = 0.3
max_tokens = 4096
```

system.md:
```markdown
# Code Reviewer

You are a senior code reviewer. Review code for bugs, security vulnerabilities, and style issues.

## Output Format
Group findings by severity: critical > warning > suggestion. Be specific and constructive.
```

### Data Formatter (tool agent with schemas)

config.toml:
```toml
[agent]
name = "data-formatter"
version = "1.0.0"
description = "Converts data between formats"
input_order = ["input_format", "output_format", "data"]

[model]
default = "claude-sonnet-4-6"

[defaults]
temperature = 0.1
max_tokens = 8192
```

input.jsonschema:
```json
{
  "type": "object",
  "properties": {
    "data": { "type": "string", "description": "Data to convert", "file": true },
    "input_format": { "type": "string", "enum": ["json", "yaml", "csv", "xml"] },
    "output_format": { "type": "string", "enum": ["json", "yaml", "csv", "xml"] }
  },
  "required": ["data", "input_format", "output_format"]
}
```

### Research Assistant (conversational agent)

config.toml:
```toml
[agent]
name = "researcher"
version = "1.0.0"
description = "Research assistant with shell access"

[model]
default = "claude-sonnet-4-6"

[defaults]
temperature = 0.5
max_tokens = 8192
```

system.md:
```markdown
# Research Assistant

You help users research topics using available tools. Use the shell to search files, read documents, and gather information. Synthesize findings into clear summaries.
```

### Translation Service (tool agent with output schema)

config.toml:
```toml
[agent]
name = "translator"
version = "1.0.0"
description = "Translates text between languages"

[model]
requires_structured_output = true
default = "claude-sonnet-4-6"

[defaults]
temperature = 0.3
max_tokens = 4096
```

input.jsonschema:
```json
{
  "type": "object",
  "properties": {
    "text": { "type": "string", "description": "Text to translate" },
    "target_language": { "type": "string", "enum": ["en", "es", "fr", "de", "ja", "zh"] }
  },
  "required": ["text", "target_language"]
}
```

output.jsonschema:
```json
{
  "type": "object",
  "properties": {
    "translated_text": { "type": "string" },
    "source_language": { "type": "string" },
    "confidence": { "type": "number" }
  },
  "required": ["translated_text", "source_language"]
}
```
