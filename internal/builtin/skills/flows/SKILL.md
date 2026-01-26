---
name: flows
description: Creating, running, and managing flows - composable agent pipelines. Use when the user wants to create workflows, pipelines, or automate multi-step agent tasks.
compatibility: Requires bash
metadata:
  author: ayo
  version: "1.0"
---

# Flows Skill

Flows are composable agent pipelines - shell scripts with structured frontmatter that orchestrate agent calls. They are the unit of work that external systems invoke.

## When to Use

Activate this skill when:
- User wants to create a pipeline or workflow
- User wants to automate multi-step tasks with agents
- User asks about flows or flow management
- User wants to run a flow
- User wants to check flow history

## Flow File Format

Flows are shell scripts with a special frontmatter format:

```bash
#!/usr/bin/env bash
# ayo:flow
# name: my-flow
# description: What this flow does
# version: 1.0.0
# author: username

set -euo pipefail

INPUT="${1:-$(cat)}"

# Process input through agents
echo "$INPUT" | ayo @ayo "Process this and return JSON"
```

### Required Frontmatter

- `# ayo:flow` - Marker that identifies this as a flow
- `# name:` - Flow name (lowercase, hyphens allowed)
- `# description:` - What the flow does

### Optional Frontmatter

- `# version:` - Semantic version
- `# author:` - Author name

## Flow Directories

Flows are discovered from (in priority order):
1. **Project**: `.ayo/flows/` (in current directory or parents)
2. **User**: `~/.config/ayo/flows/`
3. **Built-in**: `~/.local/share/ayo/flows/`

## CLI Commands

### List Flows

```bash
# List all flows
ayo flows list

# Filter by source
ayo flows list --source=project
ayo flows list --source=user

# JSON output
ayo flows list --json
```

### Show Flow Details

```bash
# Show flow details
ayo flows show my-flow

# Show full script
ayo flows show my-flow --script

# JSON output
ayo flows show my-flow --json
```

### Run a Flow

```bash
# Run with inline input
ayo flows run my-flow '{"key": "value"}'

# Run with input from stdin
echo '{"key": "value"}' | ayo flows run my-flow

# Run with input from file
ayo flows run my-flow -i input.json

# Custom timeout (seconds)
ayo flows run my-flow -t 600 '{"key": "value"}'

# Validate input without running
ayo flows run my-flow --validate '{"key": "value"}'

# Skip history recording
ayo flows run my-flow --no-history '{"key": "value"}'
```

### Create a Flow

```bash
# Create in user flows directory
ayo flows new my-flow

# Create in project directory
ayo flows new my-flow --project

# Create with input/output schemas
ayo flows new my-flow --with-schemas

# Overwrite existing
ayo flows new my-flow --force
```

### Validate a Flow

```bash
# Validate flow file or directory
ayo flows validate /path/to/flow.sh
ayo flows validate /path/to/flow-dir/
```

### Flow History

```bash
# List recent runs
ayo flows history

# Filter by flow name
ayo flows history --flow=my-flow

# Filter by status
ayo flows history --status=failed

# Limit results
ayo flows history --limit=20

# Show specific run details
ayo flows history show <run-id>

# JSON output
ayo flows history --json
```

### Replay a Flow

```bash
# Replay a previous run with original input
ayo flows replay <run-id>

# Custom timeout
ayo flows replay <run-id> -t 600

# Skip history recording
ayo flows replay <run-id> --no-history
```

## Flow Patterns

### Simple Agent Pipeline

```bash
#!/usr/bin/env bash
# ayo:flow
# name: summarize
# description: Summarize input text

set -euo pipefail

INPUT="${1:-$(cat)}"
echo "$INPUT" | ayo @ayo "Summarize this text concisely" 2>/dev/null
```

### Multi-Step Pipeline

```bash
#!/usr/bin/env bash
# ayo:flow
# name: analyze-and-summarize
# description: Analyze content and summarize findings

set -euo pipefail

INPUT="${1:-$(cat)}"
TOPIC=$(echo "$INPUT" | jq -r '.topic')

# Step 1: Analyze
ANALYSIS=$(ayo @ayo "Analyze the following topic in detail: $TOPIC" 2>/dev/null)

# Step 2: Summarize
echo "$ANALYSIS" | ayo @ayo "Create a concise summary in JSON format" 2>/dev/null
```

### With Error Handling

```bash
#!/usr/bin/env bash
# ayo:flow
# name: safe-process
# description: Process with error handling

set -euo pipefail

INPUT="${1:-$(cat)}"

# Validate input
if ! echo "$INPUT" | jq -e '.required_field' > /dev/null 2>&1; then
    echo '{"error": "Missing required_field"}' >&2
    exit 2
fi

# Process
if result=$(echo "$INPUT" | ayo @ayo "Process this" 2>/dev/null); then
    echo "$result"
else
    echo '{"error": "Processing failed"}' >&2
    exit 1
fi
```

## Structured I/O

Flows can define JSON schemas for type-safe input/output:

### Package Structure

```
my-flow/
  flow.sh          # The flow script
  input.jsonschema # Input validation schema
  output.jsonschema # Output validation schema
```

### Example Schemas

**input.jsonschema**:
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "topic": { "type": "string" },
    "max_length": { "type": "integer" }
  },
  "required": ["topic"]
}
```

**output.jsonschema**:
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "summary": { "type": "string" },
    "sources": { "type": "array", "items": { "type": "string" } }
  },
  "required": ["summary"]
}
```

## Environment Variables

During execution, flows have access to:

| Variable | Description |
|----------|-------------|
| `AYO_FLOW_NAME` | Name of the flow |
| `AYO_FLOW_RUN_ID` | Unique run ID (ULID) |
| `AYO_FLOW_DIR` | Directory containing the flow |
| `AYO_FLOW_INPUT_FILE` | Path to input file (for large inputs) |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Input validation failed |
| 124 | Timeout |

## Best Practices

1. **Use `set -euo pipefail`** - Fail fast on errors
2. **Log to stderr** - Keep stdout clean for JSON output
3. **Validate input early** - Check required fields before processing
4. **Return valid JSON** - Flows should output structured data
5. **Use meaningful names** - Flow names should describe what they do
6. **Add descriptions** - Help users understand when to use the flow
