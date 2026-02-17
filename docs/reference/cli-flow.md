# ayo flow

Manage composable workflows that orchestrate multi-step agent pipelines.

## Synopsis

```
ayo flow <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List all flows |
| `run` | Execute a flow |
| `new` | Create a new flow |
| `show` | Show flow details |
| `validate` | Validate a flow |

---

## ayo flow list

List all available flows.

### Synopsis

```
ayo flow list [flags]
```

### Example

```bash
$ ayo flow list
NAME            TYPE    DESCRIPTION
daily-summary   yaml    Generate daily project summary
code-review     yaml    Multi-stage code review
deploy          shell   Deploy to production
```

---

## ayo flow run

Execute a flow with optional input.

### Synopsis

```
ayo flow run <name> [input] [flags]
```

### Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--input` | `-i` | string | Input JSON |
| `--file` | `-f` | string | Input from file |
| `--dry-run` | | bool | Show what would run |

### Examples

```bash
# Run with inline JSON
$ ayo flow run daily-summary '{"project": "./myapp"}'

# Run with input from file
$ ayo flow run code-review -f review-request.json

# Dry run
$ ayo flow run deploy --dry-run
Would execute:
  1. @builder: Build application
  2. @tester: Run test suite
  3. @deployer: Deploy to staging
```

### JSON Output

```json
{
  "flow": "daily-summary",
  "status": "completed",
  "duration": "45s",
  "steps": [
    {
      "name": "gather",
      "agent": "@analyzer",
      "status": "completed",
      "duration": "12s"
    },
    {
      "name": "summarize",
      "agent": "@writer",
      "status": "completed",
      "duration": "33s"
    }
  ],
  "output": {
    "summary": "Today's progress: ..."
  }
}
```

---

## ayo flow new

Create a new flow.

### Synopsis

```
ayo flow new <name> [flags]
```

### Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--type` | `-t` | string | Type: yaml or shell |
| `--template` | | string | Use a template |

### Example

```bash
$ ayo flow new weekly-report -t yaml
Created flow: /Users/user/.config/ayo/flows/weekly-report.yaml
Edit the flow to add steps.
```

---

## ayo flow show

Show flow details and structure.

### Synopsis

```
ayo flow show <name>
```

### Example

```bash
$ ayo flow show code-review
Name:        code-review
Type:        yaml
Path:        /Users/user/.config/ayo/flows/code-review.yaml
Description: Multi-stage code review pipeline

Steps:
  1. analyze (parallel)
     └── @security: Security audit
     └── @performance: Performance review
  2. synthesize
     └── @reviewer: Combine findings
  3. report
     └── @writer: Generate report
```

---

## ayo flow validate

Validate a flow's configuration.

### Synopsis

```
ayo flow validate <name>
```

### Example

```bash
$ ayo flow validate code-review
✓ YAML syntax valid
✓ All agents exist
✓ Step dependencies valid
✓ No circular dependencies
```

---

## Flow Types

### YAML Flows

Declarative workflow definition:

```yaml
name: code-review
description: Multi-stage code review

input:
  properties:
    files:
      type: array
      description: Files to review

steps:
  - name: security
    agent: "@security"
    prompt: "Audit these files for security issues: {{.input.files}}"
    
  - name: performance
    agent: "@performance"
    prompt: "Review for performance issues"
    
  - name: report
    agent: "@writer"
    depends_on: [security, performance]
    prompt: "Synthesize findings into a report"
    
output:
  report: "{{.steps.report.output}}"
```

### Shell Flows

Bash scripts with JSON I/O:

```bash
#!/usr/bin/env bash
# ayo:flow
# name: deploy
# description: Deploy application

set -e

# Read input
INPUT=$(cat)
ENV=$(echo "$INPUT" | jq -r '.environment')

# Build
echo "Building for $ENV..." >&2
ayo @builder "Build the application for $ENV"

# Test
echo "Running tests..." >&2
ayo @tester "Run test suite"

# Deploy
echo "Deploying..." >&2
ayo @deployer "Deploy to $ENV"

# Output
echo '{"status": "deployed", "environment": "'"$ENV"'"}'
```

## See Also

- [Flows Guide](../flows.md) - Conceptual overview
- [Flows Specification](../flows-spec.md) - YAML schema reference
