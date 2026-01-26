# Flows Guide

Flows are composable agent pipelines - shell scripts with structured frontmatter that orchestrate agent calls. They are the unit of work that external systems invoke.

## Quick Start

### Create Your First Flow

```bash
# Create a simple flow
ayo flows new my-first-flow

# Edit the generated file
cat ~/.config/ayo/flows/my-first-flow.sh
```

The generated template:

```bash
#!/usr/bin/env bash
# ayo:flow
# name: my-first-flow
# description: TODO: Describe what this flow does

set -euo pipefail

INPUT="${1:-$(cat)}"

# TODO: Implement your flow
echo "$INPUT" | ayo @ayo "Process this input and return JSON"
```

### Run the Flow

```bash
# Run with inline input
ayo flows run my-first-flow '{"message": "Hello, world!"}'

# Run with stdin
echo '{"message": "Hello, world!"}' | ayo flows run my-first-flow

# Run with input file
ayo flows run my-first-flow -i input.json
```

---

## Flow Patterns

### Pattern 1: Sequential Pipeline

Chain agents in sequence, passing output from one to the next.

```bash
#!/usr/bin/env bash
# ayo:flow
# name: ticket-pipeline
# description: Process support ticket through classification and response

set -euo pipefail

INPUT="${1:-$(cat)}"

# Stage 1: Classify the ticket
CLASSIFICATION=$(echo "$INPUT" | ayo @ayo "
  Classify this support ticket. Return JSON with:
  - category: string (billing, technical, general)
  - priority: string (low, medium, high, urgent)
  - sentiment: string (positive, neutral, negative)
" 2>/dev/null)

# Stage 2: Generate response based on classification
echo "$CLASSIFICATION" | jq --argjson input "$INPUT" '. + {original: $input}' | \
  ayo @ayo "
    Based on this classified ticket, draft a helpful response.
    Return JSON with:
    - response: string (the draft response)
    - suggested_actions: array of strings
  " 2>/dev/null
```

**Usage:**
```bash
ayo flows run ticket-pipeline '{"subject": "Cannot login", "body": "I forgot my password"}'
```

---

### Pattern 2: Conditional Logic

Execute different paths based on agent output.

```bash
#!/usr/bin/env bash
# ayo:flow
# name: code-review
# description: Review code and create issues only if problems found

set -euo pipefail

INPUT="${1:-$(cat)}"
REPO=$(echo "$INPUT" | jq -r '.repo // "."')

# Stage 1: Review the code
REVIEW=$(ayo @ayo "
  Review the code in $REPO for bugs, security issues, and improvements.
  Return JSON with:
  - status: string ('clean' or 'issues_found')
  - findings: array of {file, line, severity, message}
" 2>/dev/null)

# Stage 2: Conditional - only create issues if problems found
if echo "$REVIEW" | jq -e '.status == "issues_found"' > /dev/null 2>&1; then
  echo "$REVIEW" | ayo @ayo "
    Create GitHub issues for these findings. Group related issues.
    Return JSON with:
    - issues_created: number
    - issue_urls: array of strings
  " 2>/dev/null
else
  echo '{"status": "clean", "message": "No issues found"}'
fi
```

**Usage:**
```bash
ayo flows run code-review '{"repo": ".", "files": ["main.go", "utils.go"]}'
```

---

### Pattern 3: Error Handling

Robust error handling with validation and fallbacks.

```bash
#!/usr/bin/env bash
# ayo:flow
# name: safe-research
# description: Research with error handling and validation

set -euo pipefail

INPUT="${1:-$(cat)}"

# Validate input
TOPIC=$(echo "$INPUT" | jq -r '.topic // empty')
if [[ -z "$TOPIC" ]]; then
  echo '{"error": "Missing required field: topic"}' >&2
  exit 2
fi

# Attempt research with timeout
if ! RESEARCH=$(timeout 120 ayo @ayo "
  Research this topic: $TOPIC
  Return JSON with:
  - summary: string
  - sources: array of strings
  - confidence: number (0-1)
" 2>/dev/null); then
  echo '{"error": "Research timed out", "topic": "'"$TOPIC"'"}' >&2
  exit 124
fi

# Validate output
if ! echo "$RESEARCH" | jq -e '.summary' > /dev/null 2>&1; then
  echo '{"error": "Invalid research output", "raw": '"$RESEARCH"'}' >&2
  exit 1
fi

echo "$RESEARCH"
```

---

### Pattern 4: External Integration

Combine agents with external tools and APIs.

```bash
#!/usr/bin/env bash
# ayo:flow
# name: git-summary
# description: Summarize recent git commits

set -euo pipefail

INPUT="${1:-$(cat)}"
DAYS=$(echo "$INPUT" | jq -r '.days // 7')
REPO=$(echo "$INPUT" | jq -r '.repo // "."')

# Get git log (external tool)
cd "$REPO"
GIT_LOG=$(git log --oneline --since="${DAYS} days ago" 2>/dev/null || echo "")

if [[ -z "$GIT_LOG" ]]; then
  echo '{"summary": "No commits in the last '"$DAYS"' days", "commits": []}'
  exit 0
fi

# Have agent summarize
echo "{\"commits\": \"$GIT_LOG\"}" | ayo @ayo "
  Summarize these git commits. Return JSON with:
  - summary: string (2-3 sentences)
  - categories: array of {name, count, commits}
  - highlights: array of notable changes
" 2>/dev/null
```

---

### Pattern 5: Nested Flows

Call other flows from within a flow.

```bash
#!/usr/bin/env bash
# ayo:flow
# name: daily-report
# description: Generate daily report combining multiple sub-flows

set -euo pipefail

INPUT="${1:-$(cat)}"
DATE=$(date +%Y-%m-%d)

# Run sub-flows
GIT_SUMMARY=$(ayo flows run git-summary '{"days": 1}' 2>/dev/null)
CODE_QUALITY=$(ayo flows run code-review '{"repo": "."}' 2>/dev/null)

# Combine results
echo "{
  \"date\": \"$DATE\",
  \"git\": $GIT_SUMMARY,
  \"quality\": $CODE_QUALITY
}" | ayo @ayo "
  Create a daily development report from this data.
  Return JSON with:
  - headline: string
  - sections: array of {title, content}
  - action_items: array of strings
" 2>/dev/null
```

---

## Structured I/O with Schemas

For type-safe flows, create a flow package with schemas.

### Create with Schemas

```bash
ayo flows new my-typed-flow --with-schemas
```

This creates:
```
~/.config/ayo/flows/my-typed-flow/
├── flow.sh
├── input.jsonschema
└── output.jsonschema
```

### Example Schemas

**input.jsonschema:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "topic": {
      "type": "string",
      "description": "Topic to research"
    },
    "depth": {
      "type": "string",
      "enum": ["brief", "detailed", "comprehensive"],
      "default": "detailed"
    }
  },
  "required": ["topic"]
}
```

**output.jsonschema:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "summary": {
      "type": "string",
      "description": "Research summary"
    },
    "sources": {
      "type": "array",
      "items": { "type": "string" }
    },
    "confidence": {
      "type": "number",
      "minimum": 0,
      "maximum": 1
    }
  },
  "required": ["summary"]
}
```

### Validation

Input is validated before execution:
```bash
# This will fail with exit code 2
ayo flows run my-typed-flow '{"wrong_field": "value"}'
# Error: Missing required field: topic
```

---

## Best Practices

### 1. Use `set -euo pipefail`

Always start with strict mode:
```bash
set -euo pipefail  # Exit on error, undefined vars, pipe failures
```

### 2. Log to stderr, Output to stdout

Keep stdout clean for JSON output:
```bash
echo "Starting process..." >&2  # Log
echo '{"result": "done"}'       # Output
```

### 3. Validate Input Early

Check required fields before processing:
```bash
REQUIRED=$(echo "$INPUT" | jq -r '.required_field // empty')
if [[ -z "$REQUIRED" ]]; then
  echo '{"error": "Missing required_field"}' >&2
  exit 2
fi
```

### 4. Use Timeouts

Protect against hanging operations:
```bash
if ! RESULT=$(timeout 60 ayo @ayo "..."); then
  echo '{"error": "Operation timed out"}' >&2
  exit 124
fi
```

### 5. Return Valid JSON

Always return structured output:
```bash
# Good
echo '{"status": "success", "data": []}'

# Bad
echo "Success!"
```

---

## Debugging Flows

### Check Flow History

```bash
# See recent runs
ayo flows history

# Filter by flow
ayo flows history --flow=my-flow

# Filter by status
ayo flows history --status=failed

# Show run details
ayo flows history show <run-id>
```

### Replay a Failed Run

```bash
# Replay with original input
ayo flows replay <run-id>
```

### Verbose Execution

Redirect stderr to see logs:
```bash
ayo flows run my-flow '{"input": "data"}' 2>&1 | tee debug.log
```

---

## Integration with External Systems

### Cron

```cron
# Run daily report at 9am
0 9 * * * /usr/local/bin/ayo flows run daily-report '{}' >> /var/log/daily-report.log 2>&1
```

### GitHub Actions

```yaml
- name: Run code review flow
  run: |
    echo '{"repo": ".", "pr": "${{ github.event.pull_request.number }}"}' | \
      ayo flows run code-review
```

### Webhooks

```python
# Flask example
@app.route('/webhook', methods=['POST'])
def handle_webhook():
    import subprocess
    result = subprocess.run(
        ['ayo', 'flows', 'run', 'webhook-handler'],
        input=request.get_json(),
        capture_output=True,
        text=True
    )
    return result.stdout
```

---

## Exit Codes

| Code | Meaning | When |
|------|---------|------|
| 0 | Success | Flow completed successfully |
| 1 | Error | General execution error |
| 2 | Validation Failed | Input didn't match schema |
| 124 | Timeout | Execution exceeded time limit |

---

## Environment Variables

During execution, these variables are available:

| Variable | Description |
|----------|-------------|
| `AYO_FLOW_NAME` | Name of the current flow |
| `AYO_FLOW_RUN_ID` | Unique run identifier (ULID) |
| `AYO_FLOW_DIR` | Directory containing the flow |
| `AYO_FLOW_INPUT_FILE` | Temp file with input (for large inputs) |
