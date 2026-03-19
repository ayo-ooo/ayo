# Best Practices

Production-ready patterns for Ayo agents.

## Agent Design

### Single Responsibility

Each agent should do one thing well:

```
# Good
- summarize: Summarize text
- translate: Translate between languages
- code-review: Review code

# Avoid
- do-everything: Summarize, translate, and review code
```

### Clear Boundaries

Define what the agent does and doesn't do:

```markdown
# Code Review Agent

You review code for quality issues.

## What You Do
- Identify bugs and potential errors
- Suggest improvements
- Check style consistency

## What You Don't Do
- Execute code
- Modify files directly
- Run tests
```

## Schema Design

### Meaningful Names

Use descriptive property names:

```json
{
  "source_code": { "type": "string" },
  "target_language": { "type": "string" }
}
```

Not:

```json
{
  "s": { "type": "string" },
  "t": { "type": "string" }
}
```

### Appropriate Types

Choose the right type for each field:

| Use | For |
|-----|-----|
| `string` | Text, identifiers, paths |
| `integer` | Counts, positions, IDs |
| `number` | Measurements, ratios |
| `boolean` | Flags, options |
| `enum` | Known values, categories |

### Sensible Defaults

Provide defaults for optional fields:

```json
{
  "format": {
    "type": "string",
    "enum": ["json", "yaml", "text"],
    "default": "json"
  },
  "verbose": {
    "type": "boolean",
    "default": false
  }
}
```

### Required vs Optional

Mark fields required only when truly necessary:

```json
{
  "required": ["input"],
  "properties": {
    "input": { "type": "string" },
    "format": { "type": "string", "default": "text" },
    "verbose": { "type": "boolean", "default": false }
  }
}
```

## System Prompts

### Structure

Use consistent formatting:

```markdown
# Agent Name

One-line description.

## Purpose

Detailed explanation.

## Capabilities

- What you can do
- How you work

## Guidelines

- Tone and style
- Response format

## Limitations

- What you don't do
- Edge cases
```

### Be Specific

Vague prompts produce vague results:

```
# Bad
You are a helpful assistant.

# Good
You are a code review assistant that identifies bugs, style issues,
and potential improvements in source code.
```

### Include Examples

Show expected behavior:

```markdown
## Example

Input: function add(a, b) { return a + b }
Output:
{
  "issues": [],
  "score": 95,
  "summary": "Clean, simple function with no issues."
}
```

## Prompt Templates

### Keep It Simple

Avoid complex template logic:

```
# Good
Process: {{.text}}
{{if .verbose}}Provide detailed analysis.{{end}}

# Avoid
{{if and (gt .count 10) (eq .format "json") (ne .mode "fast")}}
Complex conditional logic here
{{end}}
```

### Handle Missing Data

Check for existence:

```
{{if .context}}
Context: {{.context}}
{{end}}

{{with .metadata}}
Metadata: {{json .}}
{{end}}
```

## Skills

### Single Purpose

Each skill should do one thing:

```
# Good
- analyze: Analyze input data
- transform: Apply transformations
- validate: Check results

# Avoid
- process: Analyze, transform, and validate
```

### Clear Documentation

Document when to use each skill:

```markdown
## Usage

Invoke this skill when you need to:
1. Parse CSV, JSON, or XML data
2. Extract structured records
3. Handle encoding issues

Do not use for:
- Binary data
- Streaming input
```

## Hooks

### Fail Gracefully

Hooks shouldn't crash the agent:

```bash
#!/bin/bash
PAYLOAD=$(cat) || exit 0  # Exit silently on error

# Check for required fields
TIMESTAMP=$(echo "$PAYLOAD" | jq -r '.timestamp // empty')
[ -z "$TIMESTAMP" ] && exit 0

echo "[$TIMESTAMP] Event logged" >> /var/log/agent.log
```

### Quick Execution

Hooks should be fast:

```bash
# Good: Simple, fast operation
echo "Agent started" >> /tmp/log

# Avoid: Long-running operations
curl -X POST "https://slow-api.example.com/hook" --max-time 30
```

### Non-Blocking

Background long operations:

```bash
#!/bin/bash
# Send notification in background
(
  curl -X POST "https://hooks.example.com/notify" \
    -d "$(cat)" &
) &

# Exit immediately
exit 0
```

## Configuration

### Model Selection

Choose appropriate models:

| Task Type | Recommended |
|-----------|-------------|
| Simple text | Any model |
| Structured output | claude-3.5-sonnet, gpt-4o |
| Complex reasoning | claude-3.5-sonnet |
| Fast responses | gpt-4o-mini |

### Temperature Settings

| Value | Use Case |
|-------|----------|
| 0.0-0.3 | Deterministic, precise output |
| 0.4-0.7 | Balanced creativity |
| 0.8-1.0 | Creative, varied responses |

### Token Limits

Set appropriate limits:

```toml
[defaults]
max_tokens = 2048  # Short responses
max_tokens = 4096  # Medium responses
max_tokens = 8192  # Long documents
```

## Error Handling

### Validate Input

Check inputs early:

```markdown
Before processing, verify:
- Required fields are present
- File paths exist (for x-cli-file)
- Enums match allowed values
```

### Graceful Degradation

Handle missing optional data:

```markdown
If context is not provided, proceed with general analysis.
If verbose mode is off, provide concise summaries.
```

## Testing

### Test Commands

```bash
# Help
./agent --help

# Minimal input
./agent "test"

# With options
./agent "test" --format json --verbose

# Output to file
./agent "test" -o output.json

# Edge cases
./agent ""
./agent "$(cat large_file.txt)"
```

### Verify Output

Check structured output:

```bash
./agent "test" -o output.json
jq '.' output.json  # Validate JSON
```

## Distribution

### Binary Size

Optimize for size:

```bash
# Build with optimizations
go build -ldflags="-s -w" -o agent main.go

# Or use UPX
upx --best agent
```

### Documentation

Include usage docs:

```markdown
# Agent Name

Brief description.

## Installation

Download the binary for your platform.

## Usage

./agent [input] [flags]

## Examples

./agent "example input"
```

## Security

### API Keys

Never log or expose API keys:

```bash
# Good
export ANTHROPIC_API_KEY=xxx

# Avoid
./agent --api-key xxx  # Visible in process list
```

### Input Validation

Be cautious with user input:

```markdown
Treat all input as untrusted.
Validate file paths before processing.
Don't execute user-provided code.
```

### Hook Security

Hooks run with user permissions:

```bash
# Validate hook inputs
PAYLOAD=$(cat)
SAFE_INPUT=$(echo "$PAYLOAD" | jq -r '.input | @sh')
```
