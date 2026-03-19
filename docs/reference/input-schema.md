# Input Schema

Define agent inputs using JSON Schema with CLI extensions.

## Overview

The `input.jsonschema` file defines:

- Input field names and types
- Required vs optional fields
- Default values
- CLI flag configuration
- Validation constraints

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

| JSON Schema Type | Go Type | CLI Flag Type |
|-----------------|---------|---------------|
| `string` | `string` | `StringVar` |
| `integer` | `int` | `IntVar` |
| `number` | `float64` | `Float64Var` |
| `boolean` | `bool` | `BoolVar` |

## CLI Extensions

### x-cli-position

Make a field a positional argument:

```json
{
  "text": {
    "type": "string",
    "description": "Text to process",
    "x-cli-position": 1
  }
}
```

Usage: `./agent "input text"`

Position numbers determine argument order (1-indexed).

### x-cli-flag

Customize the flag name:

```json
{
  "from": {
    "type": "string",
    "description": "Source language",
    "x-cli-flag": "source-language"
  }
}
```

Usage: `./agent --source-language english`

### x-cli-short

Add a short flag:

```json
{
  "format": {
    "type": "string",
    "description": "Output format",
    "x-cli-short": "-f"
  }
}
```

Usage: `./agent -f json`

### x-cli-file

Load file contents into a field:

```json
{
  "config_file": {
    "type": "string",
    "description": "Path to config file",
    "x-cli-file": true
  }
}
```

When specified, the file contents are loaded and passed to the LLM, not just the path.

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

Required fields without defaults must be provided.

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

## Complete Example

From the translate example:

```json
{
  "type": "object",
  "properties": {
    "text": {
      "type": "string",
      "description": "Text to translate",
      "x-cli-position": 1
    },
    "from": {
      "type": "string",
      "description": "Source language",
      "x-cli-flag": "source-language",
      "x-cli-short": "-s",
      "default": "auto"
    },
    "to": {
      "type": "string",
      "description": "Target language",
      "x-cli-short": "-t"
    },
    "formal": {
      "type": "boolean",
      "description": "Use formal tone",
      "default": false
    }
  },
  "required": ["to"]
}
```

Generates CLI:

```
Usage:
  translate [flags]

Flags:
      --formal                Use formal tone (default false)
  -s, --source-language       Source language (default "auto")
  -t, --to string             Target language
```

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

2. **Flag parsing** in `cli.go`:
```go
func parseFlags() (*Input, error) {
    var input Input
    flag.StringVar(&input.Text, "text", "", "Text to translate")
    flag.StringVar(&input.From, "source-language", "auto", "Source language")
    flag.StringVar(&input.To, "to", "", "Target language")
    flag.BoolVar(&input.Formal, "formal", false, "Use formal tone")
    // ...
}
```

## Limitations

- Arrays and nested objects are parsed but not fully supported for CLI flags
- Complex types should use `x-cli-file` to load JSON/JSON5 content

## Next Steps

- [Output Schema](output-schema.md) - Define output structure
- [CLI Flags](cli-flags.md) - Advanced flag configuration
- [Examples](../examples/README.md) - See complete examples
