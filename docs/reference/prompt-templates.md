# Prompt Templates

Create dynamic prompts with Go templates.

## Overview

The `prompt.tmpl` file lets you construct dynamic user prompts using Go's text/template syntax with custom functions.

## Basic Template

```
Process the following: {{.text}}

Format: {{.format}}
```

Access input fields with `{{.field_name}}`.

## Template Functions

### file

Load file contents:

```
{{file .config_path}}
```

The file path is provided via the input schema with `x-cli-file: true`.

### env

Access environment variables:

```
{{if .env "DEBUG"}}Debug mode enabled{{end}}

API Version: {{.env "API_VERSION"}}
```

### json

JSON encode data:

```
Configuration:
{{json .config}}
```

### String Functions

| Function | Description | Example |
|----------|-------------|---------|
| `upper` | Uppercase | `{{upper .text}}` |
| `lower` | Lowercase | `{{lower .text}}` |
| `title` | Title case | `{{title .text}}` |
| `trim` | Trim whitespace | `{{trim .text}}` |

## Conditionals

### if

```
{{if .verbose}}
Provide detailed analysis with step-by-step explanations.
{{end}}
```

### if-else

```
{{if .format}}
Output in {{.format}} format.
{{else}}
Output in plain text.
{{end}}
```

### with

```
{{with .context}}
Additional context:
{{.}}
{{end}}
```

## Loops

### range

Iterate over arrays:

```
Focus areas:
{{range .focus_areas}}
- {{.}}
{{end}}
```

Iterate with index:

```
{{range $i, $item := .items}}
{{$i}}: {{$item}}
{{end}}
```

## Complete Example

From the research example:

**input.jsonschema**:
```json
{
  "type": "object",
  "properties": {
    "topic": {
      "type": "string",
      "description": "Research topic",
      "x-cli-position": 1
    },
    "depth": {
      "type": "string",
      "description": "Research depth",
      "enum": ["quick", "standard", "comprehensive"],
      "default": "standard"
    },
    "context_file": {
      "type": "string",
      "description": "Additional context file",
      "x-cli-file": true
    }
  },
  "required": ["topic"]
}
```

**prompt.tmpl**:
```gotemplate
Research topic: {{.topic}}

Depth: {{.depth}}

{{if .context_file}}Additional context:
{{file .context_file}}
{{end}}

{{if .env "RESEARCH_MODE"}}Mode: {{.env "RESEARCH_MODE"}}
{{end}}

Please provide:
1. Key findings
2. Supporting evidence
3. Recommendations
```

## Template Variables

All input fields are available as top-level variables:

| Input Field | Template Access |
|-------------|-----------------|
| `text` | `{{.text}}` |
| `max_count` | `{{.max_count}}` |
| `include_debug` | `{{.include_debug}}` |

## Escaping

To include literal template syntax:

```
{{"{{"}}variable{{"}}"}}
```

Produces: `{{variable}}`

## Best Practices

1. **Keep prompts focused**: One template per concern
2. **Use conditionals sparingly**: Complex logic belongs in code
3. **Document expected fields**: Comment what each variable should contain
4. **Test with edge cases**: Empty values, missing files, etc.

## Without Template

If no `prompt.tmpl` exists, the CLI input is passed directly as the user message.

## Next Steps

- [Input Schema](input-schema.md) - Define template variables
- [Examples](../examples/research.md) - See the research example
