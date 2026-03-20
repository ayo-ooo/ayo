# Input Schema

Define agent inputs using JSON Schema. When `input.jsonschema` exists, the generated CLI accepts JSON payloads with optional flag overrides.

## Overview

The `input.jsonschema` file defines:

- Input field names and types
- Required vs optional fields
- Default values
- Validation constraints

**Key concept**: Pass a JSON payload matching your schema. Use flags to override specific fields without editing the JSON.

## Usage Patterns

### Pattern 1: Full JSON Payload

```bash
./code-review '{"file": "main.go", "language": "go", "strict": true}'
```

### Pattern 2: JSON + Flag Overrides

```bash
./code-review '{"file": "main.go"}' --strict --language python
```

Flags always override values in the JSON payload.

### Pattern 3: Stdin

```bash
cat data.json | ./code-review -
echo '{"file": "main.go"}' | ./code-review -
```

Use `-` to read JSON from stdin.

### Pattern 4: Flags Only

For simple schemas, you can skip the JSON entirely:

```bash
./translate --text "hello" --to spanish
```

## Basic Structure

```json
{
  "type": "object",
  "properties": {
    "field_name": {
      "type": "string",
      "description": "Field description"
    }
  },
  "required": ["field_name"]
}
```

## Supported Types

| JSON Schema Type | Go Type | Flag Type |
|-----------------|---------|-----------|
| `string` | `string` | StringVar |
| `integer` | `int` | IntVar |
| `number` | `float64` | Float64Var |
| `boolean` | `bool` | BoolVar |

**Note**: Flags are only generated for primitive types (string, integer, number, boolean). Nested objects and arrays must be provided via JSON payload.

## Flag Generation

Flags are auto-generated from property names:

| Property Name | Generated Flag |
|--------------|----------------|
| `text` | `--text` |
| `source_language` | `--source-language` |
| `maxIssues` | `--max-issues` |

### Custom Flag Names

Override the flag name with the `flag` property:

```json
{
  "source_language": {
    "type": "string",
    "flag": "from"
  }
}
```

Usage: `./agent --from english`

## Required Fields

List required fields in the `required` array:

```json
{
  "type": "object",
  "properties": {
    "input": { "type": "string" },
    "format": { "type": "string" }
  },
  "required": ["input"]
}
```

Required fields without defaults must be provided either in JSON or via flag.

## Default Values

Specify defaults with the `default` keyword:

```json
{
  "format": {
    "type": "string",
    "enum": ["json", "text", "markdown"],
    "default": "text"
  },
  "verbose": {
    "type": "boolean",
    "default": false
  },
  "count": {
    "type": "integer",
    "default": 10
  }
}
```

## Enum Constraints

Limit values to a specific set:

```json
{
  "language": {
    "type": "string",
    "description": "Programming language",
    "enum": ["go", "python", "javascript", "typescript", "rust"]
  },
  "severity": {
    "type": "string",
    "description": "Minimum severity level",
    "enum": ["info", "warning", "error", "critical"],
    "default": "warning"
  }
}
```

## File Loading

For file paths that should load contents (not just the path), use the `file` property:

```json
{
  "source": {
    "type": "string",
    "description": "Path to source file",
    "file": true
  }
}
```

```bash
# File contents are loaded automatically
./agent '{"source": "main.go"}'

# Also works with flags
./agent --source main.go
```

The file contents (not the path) are passed to the LLM.

## Schema Complexity Levels

### Level 1: Simple (Primitives Only)

```json
{
  "properties": {
    "text": { "type": "string" },
    "count": { "type": "integer", "default": 10 }
  }
}
```

All fields get flags. Use JSON or flags interchangeably.

### Level 2: Nested Objects

```json
{
  "properties": {
    "config": {
      "type": "object",
      "properties": {
        "timeout": { "type": "integer" },
        "retries": { "type": "integer" }
      }
    }
  }
}
```

No flags for nested objects. Must use JSON payload:

```bash
./agent '{"config": {"timeout": 30, "retries": 3}}'
```

### Level 3: Arrays

```json
{
  "properties": {
    "files": {
      "type": "array",
      "items": { "type": "string" }
    }
  }
}
```

No flags for arrays. Must use JSON payload:

```bash
./agent '{"files": ["a.go", "b.go", "c.go"]}'
```

## Complete Example

Translate agent with minimal schema:

```json
{
  "type": "object",
  "properties": {
    "text": { "type": "string", "description": "Text to translate" },
    "from": { "type": "string", "description": "Source language", "default": "auto" },
    "to": { "type": "string", "description": "Target language" },
    "formal": { "type": "boolean", "description": "Use formal tone", "default": false }
  },
  "required": ["to"]
}
```

Generates CLI:

```
Usage:
  translate [json-input] [flags]

Flags:
      --text string      Text to translate
      --from string      Source language (default "auto")
      --to string        Target language (required)
      --formal           Use formal tone (default false)
```

Usage examples:

```bash
# Full JSON
translate '{"text": "hello", "to": "spanish"}'

# JSON with flag override
translate '{"text": "hello"}' --to spanish --formal

# Stdin
echo '{"text":"hello","to":"spanish"}' | translate -

# Flags only
translate --text "hello" --to spanish
```

## Input Resolution Order

1. Parse JSON payload (if provided)
2. Apply flag overrides
3. Apply defaults from schema
4. Validate against schema

Flags always take precedence over JSON values.

## Generated Code

The schema generates:

1. **Go struct** in `types.go`:
```go
type Input struct {
    Text   string `json:"text"`
    From   string `json:"from"`
    To     string `json:"to"`
    Formal bool   `json:"formal"`
}
```

2. **CLI parsing** in `cli.go` that handles JSON input and flag overrides.

## Migration from x-cli-* Extensions

Previously, ayo used `x-cli-*` extensions for CLI configuration:

| Old Property | New Approach |
|-------------|--------------|
| `x-cli-position: 1` | Use JSON payload |
| `x-cli-flag: "name"` | Use `flag: "name"` |
| `x-cli-short: "s"` | Not supported (removed) |
| `x-cli-file: true` | Use `file: true` |

The old extensions still work during migration but will be removed in a future version.

**Before (verbose):**
```json
{
  "text": {
    "type": "string",
    "x-cli-position": 1
  },
  "from": {
    "type": "string",
    "x-cli-flag": "source-language",
    "x-cli-short": "s"
  }
}
```

**After (minimal):**
```json
{
  "text": { "type": "string" },
  "from": { "type": "string", "flag": "source-language" }
}
```

## Next Steps

- [Output Schema](output-schema.md) - Define output structure
- [Examples](../examples/README.md) - See complete examples
