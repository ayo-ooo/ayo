# Configuration

The `config.toml` file defines agent metadata and model settings.

## Structure

```toml
[agent]
name = "agent-name"
version = "1.0.0"
description = "Agent description"

[model]
requires_structured_output = false
requires_tools = false
requires_vision = false
suggested = ["anthropic/claude-3.5-sonnet"]
default = "anthropic/claude-3.5-sonnet"

[defaults]
temperature = 0.7
max_tokens = 2048
```

## [agent] Section

Required section defining agent identity.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Agent name (used for binary name) |
| `version` | string | Yes | Semantic version |
| `description` | string | Yes | Short description for help text |

### Example

```toml
[agent]
name = "code-review"
version = "2.1.0"
description = "AI-powered code review assistant"
```

## [model] Section

Optional section for model requirements and preferences.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `requires_structured_output` | bool | false | Agent needs JSON mode |
| `requires_tools` | bool | false | Agent uses function calling |
| `requires_vision` | bool | false | Agent processes images |
| `suggested` | []string | [] | Recommended models |
| `default` | string | "" | Default model to use |

### Model Requirements

**requires_structured_output**: Set to `true` when you need guaranteed JSON output. Automatically enabled when `output.jsonschema` exists.

**requires_tools**: Set to `true` when the agent uses function calling capabilities.

**requires_vision**: Set to `true` when the agent needs to process images.

### Example

```toml
[model]
requires_structured_output = true
suggested = [
  "anthropic/claude-3.5-sonnet",
  "openai/gpt-4o",
  "google/gemini-1.5-pro"
]
default = "anthropic/claude-3.5-sonnet"
```

## [defaults] Section

Optional section for default generation parameters.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `temperature` | float | varies | Sampling temperature (0.0-2.0) |
| `max_tokens` | int | varies | Maximum tokens in response |

### Example

```toml
[defaults]
temperature = 0.3  # Lower for more deterministic output
max_tokens = 4096  # Higher for longer responses
```

## Complete Example

```toml
[agent]
name = "data-pipeline"
version = "1.0.0"
description = "ETL pipeline orchestrator with AI-powered transformations"

[model]
requires_structured_output = true
requires_tools = false
requires_vision = false
suggested = ["anthropic/claude-3.5-sonnet", "openai/gpt-4o"]
default = "anthropic/claude-3.5-sonnet"

[defaults]
temperature = 0.3
max_tokens = 4096
```

## Provider Format

Model names use the format `provider/model`:

| Provider | Example |
|----------|---------|
| Anthropic | `anthropic/claude-3.5-sonnet` |
| OpenAI | `openai/gpt-4o` |
| Google | `google/gemini-1.5-pro` |
| OpenRouter | `openrouter/anthropic/claude-3.5-sonnet` |

## Next Steps

- [Input Schema](input-schema.md) - Define agent inputs
- [Output Schema](output-schema.md) - Define output structure
