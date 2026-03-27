# Agent Configuration

This document describes the configuration options available in an agent's `config.toml` file.

## Basic Structure

```toml
[agent]
name = "my-agent"
version = "1.0.0"
description = "A description of what this agent does"

[model]
suggested = ["anthropic/claude-3.5-sonnet", "openai/gpt-4o"]

[defaults]
temperature = 0.7
```

## Interactive Mode Configuration

### `interactive`

Controls whether the agent can present an interactive form when required inputs are missing.

```toml
[agent]
interactive = true  # default
```

**Type:** `boolean` (optional)  
**Default:** `true`

When `true` (default):
- If required inputs are missing and a TTY is available, an interactive form is displayed
- Users can fill in missing values through a polished terminal UI

When `false`:
- The agent will fail immediately if required inputs are missing
- No interactive form will be shown, even in a terminal
- Useful for agents designed for automation/CI environments

### `input_order`

Specifies the order of input fields in the interactive form.

```toml
[agent]
input_order = ["prompt", "scope", "dry_run"]
```

**Type:** `array of strings` (optional)  
**Default:** Schema property order

When specified:
- Fields appear in the form in the listed order
- Only properties from `input.jsonschema` should be listed

When not specified:
- Fields appear in the order defined in `input.jsonschema`

## Precedence

Values are resolved in this order (highest priority first):

1. **CLI flags** - Values passed via `--flag value`
2. **Form input** - Values entered in interactive form
3. **Schema defaults** - Values specified in `input.jsonschema` `"default"` field

## Validation Rules

The `ayo checkit` command validates these rules:

| Rule | Severity | Description |
|------|----------|-------------|
| `input_order` entries must exist in schema | Error | Each name in `input_order` must be a property in `input.jsonschema` |
| Schema properties not in `input_order` | Warning | Properties missing from `input_order` will appear last in the form |
| `interactive` must be boolean | Error | If provided, must be `true` or `false` |

## Examples

### Agent with Interactive Form

```toml
[agent]
name = "code-reviewer"
version = "1.0.0"
description = "Reviews code changes and provides feedback"
interactive = true
input_order = ["files", "style_guide", "strictness"]

[model]
suggested = ["anthropic/claude-3.5-sonnet"]
```

### Agent for Automation (No Interactive Mode)

```toml
[agent]
name = "ci-linter"
version = "1.0.0"
description = "Lints code in CI pipelines"
interactive = false

[model]
suggested = ["anthropic/claude-3-haiku"]
```

### Minimal Configuration

```toml
[agent]
name = "simple-agent"
version = "1.0.0"
description = "A simple agent with defaults"

[model]
suggested = ["anthropic/claude-3.5-sonnet"]
```

This agent will:
- Allow interactive mode (default `true`)
- Use schema property order for the form (no `input_order` specified)
