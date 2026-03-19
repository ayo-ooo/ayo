# Output Schema

Define the structure of agent responses.

## Overview

The `output.jsonschema` file defines the expected structure of the LLM's response. When present:

1. The LLM is instructed to respond with JSON
2. The response is validated against the schema
3. Output can be written to a file with `-o`

## Basic Structure

```json
{
  "type": "object",
  "properties": {
    "result": {
      "type": "string"
    }
  }
}
```

## Supported Types

| JSON Schema Type | Go Type | Description |
|-----------------|---------|-------------|
| `string` | `string` | Text values |
| `integer` | `int` | Whole numbers |
| `number` | `float64` | Decimal numbers |
| `boolean` | `bool` | True/false |
| `array` | `[]T` | Lists of items |
| `object` | `struct{}` | Nested objects |

## Simple Output

Single field output:

```json
{
  "type": "object",
  "properties": {
    "summary": {
      "type": "string",
      "description": "Summary of the input"
    }
  }
}
```

## Nested Objects

Complex output structure:

```json
{
  "type": "object",
  "properties": {
    "summary": {
      "type": "string"
    },
    "key_points": {
      "type": "array",
      "items": {
        "type": "string"
      }
    },
    "metadata": {
      "type": "object",
      "properties": {
        "word_count": { "type": "integer" },
        "reading_time_minutes": { "type": "number" }
      }
    }
  }
}
```

## Arrays

Arrays of simple types:

```json
{
  "tags": {
    "type": "array",
    "items": {
      "type": "string"
    }
  }
}
```

Arrays of objects:

```json
{
  "issues": {
    "type": "array",
    "items": {
      "type": "object",
      "properties": {
        "line": { "type": "integer" },
        "severity": { "type": "string" },
        "message": { "type": "string" }
      }
    }
  }
}
```

## Complete Example

From the code-review example:

```json
{
  "type": "object",
  "properties": {
    "issues": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "line": { "type": "integer" },
          "severity": { "type": "string" },
          "message": { "type": "string" },
          "suggestion": { "type": "string" }
        }
      }
    },
    "summary": {
      "type": "string"
    },
    "score": {
      "type": "integer",
      "description": "Code quality score 0-100"
    }
  }
}
```

## Generated Types

The schema generates Go types in `types.go`:

```go
type Issue struct {
    Line       int    `json:"line"`
    Severity   string `json:"severity"`
    Message    string `json:"message"`
    Suggestion string `json:"suggestion"`
}

type Output struct {
    Issues  []Issue `json:"issues"`
    Summary string  `json:"summary"`
    Score   int     `json:"score"`
}
```

## Output Format

When an output schema exists:

1. The agent requests structured JSON output from the LLM
2. The response is parsed and validated
3. Output is printed as formatted JSON

Example output:

```json
{
  "issues": [
    {
      "line": 42,
      "severity": "warning",
      "message": "Unused variable 'temp'",
      "suggestion": "Remove or use the variable"
    }
  ],
  "summary": "Code is generally clean with minor improvements possible",
  "score": 85
}
```

## Writing to File

Use the `-o` flag to write output to a file:

```bash
./code-review main.go -o review.json
```

## Model Requirements

When `output.jsonschema` exists, `requires_structured_output` is automatically implied. The agent will use models with JSON mode support.

## Best Practices

1. **Keep schemas focused**: Define only what you need
2. **Use descriptions**: Help the LLM understand expected values
3. **Validate in prompt**: Remind the LLM about constraints
4. **Handle optional fields**: Not all properties may be present

## Next Steps

- [Input Schema](input-schema.md) - Define inputs
- [Examples](../examples/README.md) - See complete examples
