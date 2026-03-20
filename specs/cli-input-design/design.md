# CLI Input Definition Design

## Overview

Unified approach: **JSON payload as primary input, flags as convenience layer.**

## Core Concept

When `input.jsonschema` exists, the agent accepts:

1. **JSON payload** (positional or stdin) matching the schema
2. **Flags** for convenient overrides of top-level primitive fields

No short flags - keeps things simple and predictable.

## Usage Patterns

### Pattern 1: Full JSON Payload
```bash
./code-review '{"file": "main.go", "language": "go", "strict": true}'
```

### Pattern 2: JSON + Flag Overrides
```bash
./code-review '{"file": "main.go"}' --strict --language python
```

### Pattern 3: Stdin JSON
```bash
cat data.json | ./code-review -
echo '{"file": "main.go"}' | ./code-review -
```

### Pattern 4: Flags Only (Simple Schemas)
For schemas with only primitive top-level fields:
```bash
./translate '{"text": "hello world"}' --to spanish
```

## Input.jsonschema Changes

### Current (verbose)
```json
{
  "properties": {
    "text": { "type": "string", "x-cli-position": 1 },
    "from": { "type": "string", "x-cli-flag": "source-language", "x-cli-short": "s" },
    "to": { "type": "string", "x-cli-short": "t" }
  }
}
```

### Proposed (minimal)
```json
{
  "properties": {
    "text": { "type": "string" },
    "from": { "type": "string", "default": "auto" },
    "to": { "type": "string" }
  },
  "required": ["to"]
}
```

CLI is auto-generated:
- All top-level primitive properties get flags (kebab-case)
- No positional args except optional JSON payload

### Customization (when needed)
```json
{
  "properties": {
    "source_language": {
      "type": "string",
      "flag": "from"
    },
    "target_language": {
      "type": "string",
      "flag": "to",
      "required": true
    }
  }
}
```

## Flag Name Resolution

1. Property `source_language` → flag `--source-language`
2. If `flag` specified → use that instead

## Input Resolution Order

```
1. Parse JSON payload (if provided)
2. Apply flag overrides
3. Apply defaults from schema
4. Validate against schema
```

Flags always override JSON values.

## Schema Complexity Levels

### Level 1: Simple (primitives only)
```json
{
  "properties": {
    "text": { "type": "string" },
    "count": { "type": "integer", "default": 10 }
  }
}
```
→ Use flags for everything, or pass partial JSON

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
→ JSON payload for nested structure (no flags generated for objects)

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
→ JSON payload: `{"files": ["a.go", "b.go"]}` (no flags generated for arrays)

## File Loading

For file paths that should load contents:

```json
{
  "properties": {
    "source": { "type": "string", "file": true }
  }
}
```

```bash
# JSON with file path - contents loaded automatically
./agent '{"source": "main.go"}'

# Flag also loads file
./agent --source main.go
```

## Migration Path

1. **Phase 1**: Support new approach alongside `x-cli-*`
2. **Phase 2**: Deprecate `x-cli-*` extensions with warnings
3. **Phase 3**: Remove `x-cli-*` support

Old extensions still work during migration:
```json
{
  "text": { "x-cli-position": 1 }  // Still works, warns
}
```

## Generated CLI Examples

### translate (input.jsonschema)
```json
{
  "properties": {
    "text": { "type": "string" },
    "from": { "type": "string", "default": "auto" },
    "to": { "type": "string" }
  },
  "required": ["to"]
}
```

Generates:
```
Usage:
  translate [json-input] [flags]

Flags:
      --from string    (default "auto")
      --to string      (required)
      --text string

Examples:
  translate '{"text": "hello", "to": "spanish"}'
  translate --to spanish '{"text": "hello"}'
  echo '{"text":"hello","to":"spanish"}' | translate -
```

### code-review (input.jsonschema with nested)
```json
{
  "properties": {
    "file": { "type": "string", "file": true },
    "options": {
      "type": "object",
      "properties": {
        "strict": { "type": "boolean" },
        "max_issues": { "type": "integer" }
      }
    }
  }
}
```

Generates:
```
Usage:
  code-review [json-input] [flags]

Flags:
      --file string    (file contents loaded)

Examples:
  code-review '{"file": "main.go", "options": {"strict": true}}'
  code-review --file main.go '{"options": {"strict": true}}'
```

## Implementation

### Schema Changes

```go
type Property struct {
    Type        string `json:"type"`
    Description string `json:"description"`
    Default     any     `json:"default"`
    Enum        []string `json:"enum"`

    // CLI hints (optional)
    Flag     string `json:"flag"`   // Custom flag name (default: kebab-case property name)
    File     bool   `json:"file"`   // Load file contents

    // Deprecated but supported during migration
    CLIPosition int    `json:"x-cli-position"`
    CLIFlag     string `json:"x-cli-flag"`
    CLIShort    string `json:"x-cli-short"`
    CLIFile     bool   `json:"x-cli-file"`
}
```

### CLI Generation

```go
func GenerateCLI(schema *ParsedSchema) {
    // 1. Single optional positional arg for JSON payload
    // 2. Generate flag for every top-level primitive property
    // 3. Flag name: flag > kebab(property_name)
    // 4. No flags for nested objects or arrays

    for name, prop := range schema.Properties {
        if !isPrimitive(prop.Type) {
            continue // No flags for nested objects/arrays
        }

        flagName := prop.Flag
        if flagName == "" {
            flagName = toKebab(name)
        }

        // Generate flag...
    }
}
```

### Input Parsing

```go
func parseInput(args []string, flags map[string]any) (map[string]any, error) {
    input := make(map[string]any)

    // 1. Parse JSON payload if present
    if len(args) > 0 && (args[0] == "-" || isValidJSON(args[0])) {
        jsonData, err := parseJSONInput(args[0])
        if err != nil {
            return nil, err
        }
        input = jsonData
    }

    // 2. Apply flag overrides
    for key, value := range flags {
        if value != nil && value != zeroValue {
            input[key] = value
        }
    }

    // 3. Validate against schema
    if err := validateAgainstSchema(input, schema); err != nil {
        return nil, err
    }

    return input, nil
}
```

## Benefits

1. **No more position numbering** - Just JSON payload
2. **Nested schemas work naturally** - JSON handles structure
3. **Flags still convenient** - Override without editing JSON
4. **Pipeline friendly** - Stdin JSON support
5. **Backwards compatible** - Old x-cli-* still work during migration
6. **Cleaner schemas** - Minimal CLI hints needed
7. **No short flag complexity** - One way to specify flags
