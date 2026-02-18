<!-- Scope: Flow system - shell flows, YAML flows, execution. For flow YAML schema, see flows-spec.md. -->

# Flows Guide

Flows are composable agent pipelines that orchestrate multi-step workflows. There are two flow types:

- **Shell Flows** (`.sh`): Bash scripts with JSON I/O frontmatter
- **YAML Flows** (`.yaml`): Declarative multi-step workflows with dependencies, parallel execution, and templates

## Quick Start

### Create Your First Shell Flow

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

## YAML Flows

YAML flows provide a declarative way to define multi-step workflows with:

- **Step dependencies** - control execution order
- **Parallel execution** - independent steps run concurrently
- **Template variables** - reference outputs from previous steps
- **Conditional execution** - run steps based on conditions
- **Built-in triggers** - cron schedules and file watchers

### Create a YAML Flow

```yaml
# ~/.config/ayo/flows/daily-report.yaml
version: 1
name: daily-report
description: Generate a daily project status report

steps:
  - id: git-status
    type: shell
    run: git log --oneline --since="1 day ago"

  - id: test-results
    type: shell
    run: go test ./... -json 2>&1 | tail -20

  - id: generate-report
    type: agent
    agent: "@ayo"
    prompt: |
      Generate a daily report from this data:
      
      Recent commits:
      {{ steps.git-status.stdout }}
      
      Test results:
      {{ steps.test-results.stdout }}
      
      Return JSON with: headline, summary, action_items
    depends_on: [git-status, test-results]
```

### YAML Flow Structure

```yaml
version: 1                          # Schema version (required)
name: flow-name                     # Flow identifier (required)
description: What the flow does     # Human description

params:                             # Optional input parameters
  param-name:
    type: string
    default: "value"
    required: true

env:                                # Environment variables
  MY_VAR: "value"
  FROM_PARAM: "{{ params.param-name }}"

triggers:                           # Optional auto-triggers
  - type: cron
    schedule: "0 9 * * *"           # Daily at 9am
  - type: watch
    path: ./src
    patterns: ["*.go"]

steps:                              # Execution steps (required)
  - id: step-id                     # Unique step identifier
    type: shell | agent             # Step type
    # ... type-specific fields
```

### Step Types

#### Shell Steps

Execute shell commands:

```yaml
- id: build
  type: shell
  run: go build -o app ./cmd/...
  workdir: /path/to/project         # Optional working directory
  env:                              # Step-specific environment
    GOOS: linux
```

#### Agent Steps

Invoke an AI agent:

```yaml
- id: analyze
  type: agent
  agent: "@ayo"
  prompt: |
    Analyze this code and suggest improvements.
    {{ steps.read-code.stdout }}
```

#### Squad Steps

Dispatch work to a squad for collaborative execution:

```yaml
- id: implement
  type: squad
  squad: "#dev-team"
  prompt: |
    Implement the feature described in:
    {{ steps.plan.stdout }}
  input: "{{ steps.plan.output }}"    # Validated against squad's input.jsonschema
  timeout: 30m                        # Optional timeout
  startup: auto                       # auto | manual | required
```

**Squad step fields:**

| Field | Description | Required |
|-------|-------------|----------|
| `squad` | Target squad (e.g., `#dev-team`) | Yes |
| `prompt` | Instructions for the squad | No |
| `input` | Structured input data (validated against `input.jsonschema`) | No |
| `timeout` | Maximum execution time | No |
| `startup` | Squad startup behavior (see below) | No |

**Startup modes:**

| Mode | Behavior |
|------|----------|
| `auto` | Start squad if not running, reuse if running (default) |
| `manual` | Assume squad is already running, fail if not |
| `required` | Always start a fresh squad, stop when done |

**Example with schema validation:**

```yaml
steps:
  - id: plan
    type: agent
    agent: "@planner"
    prompt: "Create a development plan for: {{ params.feature }}"
    output_schema: planning-output.jsonschema  # Validate planner output

  - id: implement
    type: squad
    squad: "#dev-team"
    input: "{{ steps.plan.output }}"
    # Squad's input.jsonschema validates the incoming data
    # Squad's output.jsonschema defines the structure returned
    depends_on: [plan]

  - id: review
    type: agent
    agent: "@reviewer"
    input: "{{ steps.implement.output }}"
    depends_on: [implement]
```

### Dependencies and Parallel Execution

Steps run in parallel unless they have dependencies:

```yaml
steps:
  # These run in parallel
  - id: step-a
    type: shell
    run: echo "A"

  - id: step-b
    type: shell
    run: echo "B"

  # This waits for both A and B
  - id: step-c
    type: shell
    run: echo "C after A and B"
    depends_on: [step-a, step-b]
```

### Template Variables

Reference data from previous steps and parameters:

| Template | Description |
|----------|-------------|
| `{{ steps.ID.stdout }}` | Standard output from step |
| `{{ steps.ID.stderr }}` | Standard error from step |
| `{{ steps.ID.exit_code }}` | Exit code (0 = success) |
| `{{ params.NAME }}` | Input parameter value |
| `{{ env.VAR }}` | Environment variable |

### Conditional Execution

Run steps only when conditions are met:

```yaml
- id: deploy
  type: shell
  run: ./deploy.sh
  when: "{{ steps.test.exit_code == 0 }}"
  depends_on: [test]
```

### Triggers

Define how flows are automatically triggered:

#### Cron Triggers

```yaml
triggers:
  - type: cron
    schedule: "*/30 * * * *"        # Every 30 minutes
    enabled: true
```

#### Watch Triggers

```yaml
triggers:
  - type: watch
    path: ./src
    patterns: ["*.go", "*.mod"]
    events: [create, modify]
    debounce: 5s
```

### Running YAML Flows

```bash
# Run by name
ayo flows run daily-report

# With parameters
ayo flows run my-flow --param key=value

# Validate before running
ayo flows validate daily-report.yaml
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

---

## Flows vs Squad Dispatch

Flows and squads can both orchestrate multi-agent work. Here's when to use each:

### Use Flows When

| Scenario | Why Flows Work Better |
|----------|----------------------|
| **Known, repeatable steps** | Flow steps are explicit and versioned |
| **Pipeline processing** | Output from one step feeds the next |
| **Mixed step types** | Combine shell, agent, and squad steps |
| **Scheduled execution** | Built-in cron and watch triggers |
| **Cross-squad orchestration** | Flows can invoke multiple squads |
| **Auditable pipelines** | Flow runs are logged and replayable |

### Use Squad Dispatch When

| Scenario | Why Squad Dispatch Works Better |
|----------|--------------------------------|
| **Unknown steps** | Squad lead determines approach |
| **Parallel collaboration** | Multiple agents work simultaneously |
| **Iterative work** | Agents adapt as they learn more |
| **Persistent workspace** | Files persist across sessions |
| **Human-like delegation** | "Build this feature" without specifying how |

### Decision Tree

```
                Is the workflow...
                       │
         ┌─────────────┴─────────────┐
         │                           │
   Known steps?              Unknown steps?
   (do X, then Y)           (achieve goal G)
         │                           │
         ▼                           ▼
      ┌──────┐                ┌────────────┐
      │ FLOW │                │   SQUAD    │
      │      │                │  DISPATCH  │
      └──────┘                └────────────┘
```

### Combining Both

Flows can orchestrate squads, giving you the best of both:

```yaml
# Pipeline that uses squads for implementation
version: 1
name: feature-pipeline
steps:
  - id: requirements
    type: agent
    agent: "@pm"
    prompt: "Break down this feature request: {{ params.request }}"

  - id: design
    type: squad
    squad: "#architecture"
    input: "{{ steps.requirements.output }}"
    depends_on: [requirements]

  - id: implement
    type: squad
    squad: "#dev-team"
    input: "{{ steps.design.output }}"
    depends_on: [design]

  - id: test
    type: squad
    squad: "#qa-team"
    input: |
      {
        "implementation": {{ steps.implement.output | tojson }},
        "requirements": {{ steps.requirements.output | tojson }}
      }
    depends_on: [implement]

  - id: deploy
    type: shell
    run: ./deploy.sh
    when: "{{ steps.test.output.passed == true }}"
    depends_on: [test]
```

This flow:
1. Uses an agent to break down requirements
2. Dispatches to `#architecture` squad for design
3. Dispatches to `#dev-team` squad for implementation
4. Dispatches to `#qa-team` squad for testing
5. Runs deployment if tests pass

Each squad handles its work autonomously while the flow orchestrates the overall pipeline.

---

## See Also

- [Architecture Overview](architecture.md) - Mental model and decision tree
- [Squads](squads.md) - Team sandboxes and SQUAD.md
- [I/O Schemas](io-schemas.md) - Schema validation for flows and squads
- [Flow Specification](flows-spec.md) - YAML schema reference
