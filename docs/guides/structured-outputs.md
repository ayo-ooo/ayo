# Structured Outputs

Working with JSON output schemas.

## Overview

Output schemas define the structure of agent responses, enabling:

- Type-safe responses
- Consistent output format
- Easy integration with other tools

## Basic Output

Simple output schema:

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

## Nested Structures

Complex output with nested objects:

```json
{
  "type": "object",
  "properties": {
    "analysis": {
      "type": "object",
      "properties": {
        "score": { "type": "integer" },
        "confidence": { "type": "number" }
      }
    },
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "category": { "type": "string" },
          "description": { "type": "string" },
          "severity": { "type": "string" }
        }
      }
    }
  }
}
```

## Design Patterns

### Status + Data Pattern

```json
{
  "properties": {
    "success": { "type": "boolean" },
    "data": { "type": "object" },
    "error": { "type": "string" }
  }
}
```

### List with Metadata Pattern

```json
{
  "properties": {
    "items": {
      "type": "array",
      "items": { "type": "object" }
    },
    "total": { "type": "integer" },
    "page": { "type": "integer" }
  }
}
```

### Hierarchical Pattern

```json
{
  "properties": {
    "sections": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "title": { "type": "string" },
          "content": { "type": "string" },
          "subsections": {
            "type": "array",
            "items": { "type": "object" }
          }
        }
      }
    }
  }
}
```

## Prompting for Structure

Guide the LLM with your system prompt:

```markdown
## Output Format

Respond with JSON matching this structure:
{
  "result": "Your analysis result",
  "confidence": 0.0-1.0,
  "sources": ["List of sources used"]
}

Ensure:
- All required fields are present
- Numbers are within expected ranges
- Arrays are properly formatted
```

## Handling Output

### In Shell Scripts

```bash
./agent "input" -o output.json
result=$(jq -r '.result' output.json)
score=$(jq -r '.score' output.json)
```

### In Go

```go
type Output struct {
    Result  string  `json:"result"`
    Score   int     `json:"score"`
    Details []Item  `json:"details"`
}

var output Output
json.Unmarshal(response, &output)
```

### In Python

```python
import json

with open('output.json') as f:
    data = json.load(f)

result = data['result']
score = data['score']
```

## Validation

### Schema Constraints

Use JSON Schema validation features:

```json
{
  "properties": {
    "score": {
      "type": "integer",
      "minimum": 0,
      "maximum": 100
    },
    "email": {
      "type": "string",
      "format": "email"
    },
    "priority": {
      "type": "string",
      "enum": ["low", "medium", "high"]
    }
  }
}
```

### Prompt Validation

Ask the LLM to validate:

```markdown
Before responding, verify:
- All required fields are present
- Values are within expected ranges
- Arrays are not empty when required
```

## Error Handling

### Parsing Errors

Handle malformed output:

```bash
output=$(./agent "input" 2>&1)
if echo "$output" | jq -e . >/dev/null 2>&1; then
    # Valid JSON
    result=$(echo "$output" | jq -r '.result')
else
    echo "Error: Invalid JSON output"
fi
```

### Partial Output

Handle incomplete responses:

```json
{
  "properties": {
    "result": { "type": "string" },
    "partial": { "type": "boolean" },
    "error": { "type": "string" }
  }
}
```

## Best Practices

1. **Keep schemas focused**: Don't over-specify
2. **Use descriptions**: Help the LLM understand fields
3. **Validate in prompt**: Remind about constraints
4. **Handle edge cases**: Plan for empty or partial results
5. **Test thoroughly**: Verify with various inputs

## Next Steps

- [Input Schema](../reference/input-schema.md) - Define inputs
- [Output Schema](../reference/output-schema.md) - Full reference
- [Examples](../examples/README.md) - See structured output examples
