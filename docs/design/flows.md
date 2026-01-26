# Flows Design Specification

> **Status**: Implemented  
> **Last Updated**: January 2025

## Overview

Flows are shell scripts that compose agents into pipelines. They are the **unit of work** that external orchestrators invoke.

### Key Insight

Ayo is the execution engine, not the orchestrator. Flows have no triggers—they are invoked by external systems:

- Django background tasks
- Cron / systemd timers
- GitHub Actions
- Webhooks routed by your web framework
- n8n / Zapier / Pipedream
- Manual CLI invocation

---

## File Format

### Single File Approach

A flow is a shell script with structured frontmatter:

```bash
#!/usr/bin/env bash
# ayo:flow
# name: code-review
# description: Review code and create GitHub issues
# input: input.jsonschema
# output: output.jsonschema

set -euo pipefail

INPUT="${1:-$(cat)}"

# Stage 1: Review code
REVIEW=$(echo "$INPUT" | ayo @code-reviewer)

# Stage 2: Create issues if findings exist
if echo "$REVIEW" | jq -e '.findings | length > 0' > /dev/null; then
  echo "$REVIEW" | ayo @issue-reporter
else
  echo '{"status": "clean", "findings": []}'
fi
```

### Why Shell Scripts?

1. **LLM-friendly**: Every LLM understands bash
2. **Immediately executable**: No custom runtime needed
3. **Full power**: Conditionals, loops, error handling, external tools
4. **Portable**: Works anywhere bash works
5. **Debuggable**: Standard shell debugging techniques apply

### Frontmatter Spec

The `# ayo:flow` marker identifies the file as a flow. Subsequent `# key: value` lines are parsed as metadata.

| Key | Required | Description |
|-----|----------|-------------|
| `name` | Yes | Flow identifier (must match filename without extension) |
| `description` | Yes | Human-readable description |
| `input` | No | Path to input JSON Schema (relative to flow file) |
| `output` | No | Path to output JSON Schema (relative to flow file) |
| `version` | No | Semantic version |
| `author` | No | Author name or handle |

---

## Directory Structure

### Simple Flows

For flows without schemas:

```
~/.config/ayo/flows/
├── code-review.sh
├── daily-standup.sh
└── research-report.sh
```

### Flows with Schemas

For flows with structured I/O:

```
~/.config/ayo/flows/
├── code-review/
│   ├── flow.sh
│   ├── input.jsonschema
│   └── output.jsonschema
├── daily-standup.sh
└── research-report/
    ├── flow.sh
    └── output.jsonschema
```

### Discovery Rules

1. Files ending in `.sh` directly in `flows/` are simple flows
2. Directories in `flows/` containing `flow.sh` are flow packages
3. Flow name is the filename (without `.sh`) or directory name

---

## I/O Contract

For orchestrators to reliably invoke flows:

### Input

| Method | Example |
|--------|---------|
| Argument | `ayo flows run code-review '{"repo": "."}'` |
| Stdin | `echo '{"repo": "."}' \| ayo flows run code-review` |
| File | `ayo flows run code-review --input data.json` |

Inside the flow:
```bash
INPUT="${1:-$(cat)}"
```

### Output

- **Stdout**: JSON output only (for orchestrator to capture)
- **Stderr**: Logs, spinners, progress, errors (for debugging)

```bash
# Good - JSON to stdout
echo '{"status": "success", "issues": []}'

# Bad - don't mix logs with output
echo "Processing..." # This breaks JSON parsing!
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Input validation failed |
| 3 | Agent execution failed |
| 124 | Timeout |

### Environment Variables

Flows receive context via environment:

| Variable | Description |
|----------|-------------|
| `AYO_FLOW_NAME` | Name of the executing flow |
| `AYO_FLOW_RUN_ID` | Unique run identifier |
| `AYO_FLOW_INPUT_FILE` | Path to input file (if provided) |

---

## CLI Commands

### Discovery

```bash
# List available flows
ayo flows list

# Example output:
# NAME            DESCRIPTION                              INPUT   OUTPUT
# code-review     Review code and create GitHub issues     yes     yes
# daily-standup   Generate daily standup from git log      no      yes
# research        Research a topic and summarize           yes     no
```

```bash
# Show flow details
ayo flows show code-review

# Example output:
# Name:        code-review
# Description: Review code and create GitHub issues
# Path:        ~/.config/ayo/flows/code-review/flow.sh
# 
# Input Schema:
#   repo: string (required) - Repository path
#   files: array of strings - Specific files to review
#
# Output Schema:
#   status: string - "clean" or "issues_found"
#   findings: array of objects
#     - file: string
#     - line: number
#     - severity: string
#     - message: string
```

### Execution

```bash
# Run with argument
ayo flows run code-review '{"repo": ".", "files": ["main.go"]}'

# Run with stdin
echo '{"repo": "."}' | ayo flows run code-review

# Run with input file
ayo flows run code-review --input request.json

# Validate input without running
ayo flows run code-review --validate '{"repo": "."}'

# Set timeout
ayo flows run code-review --timeout 300 '{"repo": "."}'
```

### Authoring

```bash
# Create new flow (interactive)
ayo flows new my-flow

# Create with schemas
ayo flows new my-flow --with-schemas

# Validate flow file
ayo flows validate ./my-flow.sh
ayo flows validate ./my-flow/
```

---

## Schema Validation

### Input Validation

If `input.jsonschema` exists:
1. Parse input JSON
2. Validate against schema
3. Exit 2 if validation fails
4. Proceed if valid

```bash
$ ayo flows run code-review '{"invalid": "input"}'
Error: Input validation failed
  - missing required field: repo
  
Exit code: 2
```

### Output Validation

If `output.jsonschema` exists:
1. Capture stdout from flow
2. Parse as JSON
3. Validate against schema
4. If invalid, log warning but still return output

Output validation is advisory—the flow output is returned even if it doesn't match the schema.

### Schema Discovery for Chaining

Flows participate in the chaining system:

```bash
# Find flows that can receive code-reviewer output
ayo chain from @code-reviewer

# Output:
# AGENT/FLOW       COMPATIBILITY
# @issue-reporter  exact
# code-review      structural (flow)
```

---

## Error Handling

### In Flow Scripts

```bash
#!/usr/bin/env bash
# ayo:flow
# name: safe-flow
# description: Flow with error handling

set -euo pipefail

# Trap errors
trap 'echo "{\"error\": \"Flow failed at line $LINENO\"}" >&2; exit 1' ERR

INPUT="${1:-$(cat)}"

# Validate input manually if needed
if ! echo "$INPUT" | jq -e '.required_field' > /dev/null 2>&1; then
  echo '{"error": "missing required_field"}' >&2
  exit 2
fi

# Main logic...
```

### From CLI

```bash
$ ayo flows run broken-flow '{}' 2>errors.log
$ echo $?
1
$ cat errors.log
Error in stage 1: Agent @code-reviewer failed
  Tool 'bash' returned non-zero exit code
  Command: git status
  Error: not a git repository
```

---

## Examples

### Simple Pipeline

```bash
#!/usr/bin/env bash
# ayo:flow
# name: summarize-repo
# description: Summarize a repository's purpose

set -euo pipefail

REPO="${1:-$(cat)}"

# Single agent, simple output
ayo @ayo "Summarize the purpose of the repository at $REPO. Be concise."
```

### Conditional Pipeline

```bash
#!/usr/bin/env bash
# ayo:flow
# name: smart-review
# description: Review code, create issues only if problems found

set -euo pipefail

INPUT="${1:-$(cat)}"

# Stage 1: Review
REVIEW=$(echo "$INPUT" | ayo @code-reviewer)

# Stage 2: Conditional
FINDING_COUNT=$(echo "$REVIEW" | jq '.findings | length')

if [ "$FINDING_COUNT" -gt 0 ]; then
  # Create issues
  ISSUES=$(echo "$REVIEW" | ayo @issue-reporter)
  echo "$ISSUES"
else
  echo '{"status": "clean", "message": "No issues found"}'
fi
```

### Fan-out Pattern

```bash
#!/usr/bin/env bash
# ayo:flow
# name: multi-review
# description: Review code with multiple specialized agents

set -euo pipefail

INPUT="${1:-$(cat)}"

# Run multiple reviewers in parallel
SECURITY=$(echo "$INPUT" | ayo @security-reviewer &)
PERF=$(echo "$INPUT" | ayo @perf-reviewer &)
STYLE=$(echo "$INPUT" | ayo @style-reviewer &)

wait

# Combine results
jq -n \
  --argjson security "$SECURITY" \
  --argjson perf "$PERF" \
  --argjson style "$STYLE" \
  '{security: $security, performance: $perf, style: $style}'
```

### With External Tools

```bash
#!/usr/bin/env bash
# ayo:flow
# name: pr-review
# description: Review a GitHub PR

set -euo pipefail

PR_URL="${1:-$(cat)}"

# Fetch PR diff using gh CLI
DIFF=$(gh pr diff "$PR_URL")

# Review with agent
REVIEW=$(echo "$DIFF" | ayo @code-reviewer "Review this PR diff:")

# Post comment back to PR
gh pr comment "$PR_URL" --body "$REVIEW"

echo '{"status": "commented", "pr": "'"$PR_URL"'"}'
```

---

## Orchestrator Integration

### Django Example

```python
# tasks.py (using django-q2)
from django_q.tasks import async_task
import subprocess
import json

def run_flow(flow_name: str, input_data: dict) -> dict:
    """Execute an ayo flow and return the result."""
    result = subprocess.run(
        ["ayo", "flows", "run", flow_name, json.dumps(input_data)],
        capture_output=True,
        text=True,
        timeout=300,
    )
    
    if result.returncode != 0:
        raise FlowExecutionError(result.stderr)
    
    return json.loads(result.stdout)

# Async execution
async_task(run_flow, "code-review", {"repo": "/path/to/repo"})
```

### GitHub Actions Example

```yaml
# .github/workflows/review.yml
name: Code Review
on:
  pull_request:
    types: [opened, synchronize]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install ayo
        run: go install github.com/alexcabrera/ayo/cmd/ayo@latest
      
      - name: Run review flow
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          ayo flows run code-review '${{ toJson(github.event) }}' > review.json
      
      - name: Post review comment
        run: |
          gh pr comment ${{ github.event.pull_request.number }} \
            --body "$(jq -r '.summary' review.json)"
```

### Cron Example

```bash
# crontab -e
# Run daily standup generation at 9am
0 9 * * 1-5 /usr/local/bin/ayo flows run daily-standup | \
  /usr/local/bin/slack-post --channel engineering
```

---

## Implementation Plan

### Phase 1: Core Infrastructure

1. Flow discovery and loading
2. Frontmatter parsing
3. `ayo flows list` command
4. `ayo flows show <name>` command

### Phase 2: Execution

1. `ayo flows run <name> [input]` command
2. I/O handling (stdin, args, files)
3. Environment variable injection
4. Exit code propagation

### Phase 3: Validation

1. Input schema validation
2. Output schema validation (advisory)
3. `--validate` flag
4. `ayo flows validate <path>` command

### Phase 4: Authoring

1. `ayo flows new <name>` command
2. Interactive scaffolding
3. Schema generation helpers

### Phase 5: Integration

1. Chain command integration
2. Flows skill for @ayo
3. Documentation and examples

---

## Resolved Decisions

These questions were resolved during implementation:

| Question | Decision | Rationale |
|----------|----------|----------|
| **Run history** | Store in ayo SQLite database (`flow_runs` table) | Enables debugging, replay, analytics without requiring orchestrator persistence |
| **Streaming output** | Stream stderr, buffer stdout | Human-in-the-loop interactions require real-time feedback; stdout reserved for JSON |
| **Flow-to-flow calls** | Naturally supported | Flows are shell scripts; `ayo flows run` works inside flows |
| **Built-in examples** | Documentation only | Examples provided in skill documentation, not embedded in binary |
| **History retention** | 30 days OR 1000 runs (configurable) | Balance between useful history and storage management |
| **Run ID format** | ULID | Sortable, unique, URL-safe identifiers |
| **Exit codes** | 0=success, 1=error, 2=validation, 124=timeout | Match common Unix conventions |
| **History CLI** | `ayo flows history` + `ayo flows replay` | Simple commands for viewing and replaying runs |
| **Auto-prune** | On each new run | Automatic cleanup without manual intervention |

### Implementation Details

**Database schema** (`internal/db/migrations/002_flows.sql`):
- `flow_runs` table with run ID, flow identification, status, I/O, timing, and relationships
- Indexes on flow_name, status, started_at for efficient queries
- Foreign keys to parent_run_id and session_id for traceability

**Configuration** (`~/.config/ayo/ayo.json`):
```json
{
  "flows": {
    "history_retention_days": 30,
    "history_max_runs": 1000
  }
}
```

**Environment variables** available during execution:
- `AYO_FLOW_NAME` - Flow name
- `AYO_FLOW_RUN_ID` - Unique run ID
- `AYO_FLOW_DIR` - Flow directory
- `AYO_FLOW_INPUT_FILE` - Temp file for large inputs (>1KB)

See [flows-implementation.md](flows-implementation.md) for the full implementation plan.

---

## Appendix: Frontmatter Grammar

```
flow_file     = shebang frontmatter script
shebang       = "#!/usr/bin/env bash" newline
frontmatter   = marker (metadata)*
marker        = "# ayo:flow" newline
metadata      = "# " key ":" value newline
key           = identifier
value         = text_until_newline
script        = <any bash script>
```

Example parsed structure:
```json
{
  "name": "code-review",
  "description": "Review code and create GitHub issues",
  "input": "input.jsonschema",
  "output": "output.jsonschema"
}
```
