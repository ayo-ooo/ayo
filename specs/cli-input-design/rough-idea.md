# Redesign CLI Input Definition

## Rough Idea

The current `input.jsonschema` with `x-cli-*` extensions feels bolted-on and verbose. We want to either:

1. **Infer CLI arguments/parameters automatically** from a simpler definition
2. **Find a better format than JSON Schema** for specifying CLI inputs

## Current Problems

From `examples/translate/input.jsonschema`:
```json
{
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
      "x-cli-short": "s",
      "default": "auto"
    },
    "to": {
      "type": "string",
      "description": "Target language",
      "x-cli-short": "t"
    }
  },
  "required": ["to"]
}
```

Issues:
1. **Position numbering** - Why number positions 1, 2, 3? Order could be inferred
2. **Redundant flag names** - `x-cli-flag` duplicates info when property name could work
3. **Verbose JSON Schema** - Lots of boilerplate for simple inputs
4. **Mixed concerns** - JSON Schema is for validation, CLI is for UX

## Goals

- Simpler, more intuitive input definition
- Less boilerplate for common cases
- Still support complex validation when needed
- Maintain backwards compatibility or clean migration path
