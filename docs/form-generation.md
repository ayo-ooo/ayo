# Form Generation Specification

This document describes how `input.jsonschema` is transformed into interactive forms using the `huh` library.

## Type Mapping

JSON Schema types are mapped to `huh` form fields as follows:

| JSON Schema Type | Huh Field | Description |
|------------------|-----------|-------------|
| `string` | `huh.Input` | Single-line text input |
| `string` + `enum` | `huh.Select` | Dropdown selection |
| `string` + `format: multiline` | `huh.Text` | Multi-line text area |
| `boolean` | `huh.Confirm` | Yes/No confirmation |
| `integer` | `huh.Input` | Numeric input (validated) |
| `number` | `huh.Input` | Numeric input (validated) |
| `array` + `enum` | `huh.MultiSelect` | Multiple choice selection |

## Field Properties

Each field is configured using schema properties:

| Schema Property | Form Usage |
|-----------------|------------|
| `title` | Field label |
| `description` | Help text (shown via `?` toggle) |
| `default` | Pre-populated value |
| `enum` | Options for Select/MultiSelect |
| `required` | Validation: field must have value |

## Validation Rules

Schema constraints are mapped to inline validation:

| Schema Constraint | Validation |
|-------------------|------------|
| `required` | Field must not be empty |
| `minLength` | Minimum character count |
| `maxLength` | Maximum character count |
| `minimum` | Minimum numeric value |
| `maximum` | Maximum numeric value |
| `pattern` | Regex pattern matching |

Validation errors are shown inline immediately when the user moves away from a field.

## Help Toggle

Each field supports a help toggle:

- Press `?` to show/hide the field's description
- Description text comes from the schema's `description` property
- Help text is styled with a dimmed appearance

## Pre-population

Form fields are pre-populated in this order (later overrides earlier):

1. **Schema default** - From `input.jsonschema` `"default"` property
2. **CLI flags** - Values provided via command-line flags

When a form is shown, any values already provided via CLI flags appear pre-filled in their respective fields.

## Example Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["prompt"],
  "properties": {
    "prompt": {
      "type": "string",
      "title": "Prompt",
      "description": "What would you like the agent to do?",
      "minLength": 10
    },
    "scope": {
      "type": "string",
      "title": "Scope",
      "description": "Where should the agent operate?",
      "enum": ["file", "directory", "project"],
      "default": "project"
    },
    "dry_run": {
      "type": "boolean",
      "title": "Dry Run",
      "description": "Preview changes without applying",
      "default": false
    },
    "count": {
      "type": "integer",
      "title": "Count",
      "description": "Number of results to return",
      "minimum": 1,
      "maximum": 100,
      "default": 10
    }
  }
}
```

## Generated Form Behavior

For the example schema above:

1. **Prompt** (required, string)
   - Text input field
   - Validates minimum 10 characters
   - Cannot be empty (required)

2. **Scope** (string, enum)
   - Dropdown with options: file, directory, project
   - Pre-selected to "project" (default)

3. **Dry Run** (boolean)
   - Yes/No confirmation
   - Defaults to No (false)

4. **Count** (integer)
   - Numeric input
   - Validates between 1 and 100
   - Pre-filled with 10 (default)

## Nested Objects

For v1, nested object types in the schema are **not supported** in forms. Agents with nested input schemas should:
- Use CLI flags for nested values, or
- Provide JSON input directly

This may be addressed in a future version.

## Arrays

Arrays with `enum` values generate `huh.MultiSelect` fields allowing multiple selections. Arrays without `enum` are not supported in v1.
