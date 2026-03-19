# Template Rendering API

Template rendering allows generated agents to dynamically construct prompts using input data.

## Function Signatures

### ParseTemplate

```go
func ParseTemplate(data string) (*template.Template, error)
```

Parses a template string and returns a compiled template. The template uses Go's `text/template` syntax with custom functions.

### Render

```go
func Render(tmpl *template.Template, data map[string]any) (string, error)
```

Executes a parsed template with the provided data map. Returns the rendered string or an error.

### RenderString

```go
func RenderString(templateStr string, data map[string]any) (string, error)
```

Convenience function that combines `ParseTemplate` and `Render`. Use when parsing once is not needed.

## Available Template Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `json` | `(any) (string, error)` | Converts value to JSON string |
| `file` | `(string) (string, error)` | Reads file contents at given path |
| `env` | `(string) string` | Returns environment variable value |
| `upper` | `(string) string` | Converts to uppercase |
| `lower` | `(string) string` | Converts to lowercase |
| `title` | `(string) string` | Converts to title case |
| `trim` | `(string) string` | Removes leading/trailing whitespace |

## Input Data Structure

The `data` parameter is a `map[string]any` containing:

- **Input fields**: Key-value pairs from the agent's configured input schema
- **System data**: Additional context provided by the runtime

Example:
```go
data := map[string]any{
    "name":    "example",
    "count":   42,
    "enabled": true,
}
```

## Template Syntax

Templates use standard Go text/template syntax:

```
Hello, {{.name}}!
You have {{.count}} items.
{{if .enabled}}Feature is enabled.{{end}}
```

### Using Functions

```
Environment: {{env "HOME"}}
Upper name: {{upper .name}}
JSON data: {{json .}}
File contents: {{file "/path/to/data.txt"}}
```

## Error Handling

| Error | Cause | Resolution |
|-------|-------|------------|
| `parsing template: ...` | Invalid template syntax | Fix template syntax |
| `rendering template: ...` | Execution failure | Check data types and function arguments |
| `reading file ...` | File not found (file function) | Verify file path exists |
| `json: ...` | JSON encoding failure | Ensure value is serializable |

## Integration in Generated Code

The `renderPrompt` function in generated agents should:

1. Accept the typed `Input` struct (if schema defined) or raw string
2. Convert input to `map[string]any`
3. Call `RenderString(promptTemplate, data)`
4. Return the rendered string

Example generated signature:
```go
func renderPrompt(input Input) string {
    data := map[string]any{
        "field1": input.Field1,
        "field2": input.Field2,
    }
    result, err := template.RenderString(promptTemplate, data)
    if err != nil {
        return promptTemplate // fallback to raw template
    }
    return result
}
```
