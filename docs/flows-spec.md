# YAML Flow Specification

This document describes the YAML format for defining multi-step flows in ayo.

## Overview

YAML flows enable orchestrating multiple steps (shell commands and agent invocations) in a declarative format. They support:

- Sequential and parallel execution
- Template substitution between steps
- Conditional execution
- Automatic triggers (cron and filesystem watch)
- Input/output schema validation

## File Format

Flow files are YAML files stored in `~/.config/ayo/flows/` or `./.config/ayo/flows/` (project-local).

### Basic Structure

```yaml
version: 1
name: my-flow
description: What this flow does

steps:
  - id: step1
    type: shell
    run: echo "Hello"
```

## Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | integer | Yes | Spec version (currently `1`) |
| `name` | string | Yes | Flow identifier (unique within directory) |
| `description` | string | No | Human-readable description |
| `created_by` | string | No | Creator (`"@ayo"` or `"user"`) |
| `created_at` | timestamp | No | Creation timestamp (ISO 8601) |
| `input` | schema | No | JSON Schema for input parameters |
| `output` | schema | No | JSON Schema for flow output |
| `steps` | array | Yes | List of execution steps |
| `triggers` | array | No | Automatic execution triggers |

## Steps

Each step has an `id` and `type`, plus type-specific fields.

### Shell Steps

Execute a shell command:

```yaml
- id: list-files
  type: shell
  run: |
    find ~/notes -name "*.md" -mtime -1
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique step identifier |
| `type` | string | Yes | Must be `"shell"` |
| `run` | string | Yes | Shell command to execute |
| `env` | map | No | Additional environment variables |
| `timeout` | duration | No | Max execution time (default: 5m) |
| `when` | template | No | Condition for execution |
| `depends_on` | array | No | Steps that must complete first |
| `continue_on_error` | bool | No | Continue flow on failure |

### Agent Steps

Invoke an agent:

```yaml
- id: summarize
  type: agent
  agent: "@summarizer"
  prompt: "Summarize the following notes"
  input: "{{ steps.list-files.stdout }}"
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique step identifier |
| `type` | string | Yes | Must be `"agent"` |
| `agent` | string | Yes | Agent handle (e.g., `"@summarizer"`) |
| `prompt` | string | Yes | Prompt to send to agent |
| `context` | string | No | Additional context (prepended to prompt) |
| `input` | string | No | Data to include with prompt |
| `timeout` | duration | No | Max execution time (default: 10m) |
| `when` | template | No | Condition for execution |
| `depends_on` | array | No | Steps that must complete first |
| `continue_on_error` | bool | No | Continue flow on failure |

## Template Syntax

Steps can reference parameters and previous step outputs using `{{ }}` syntax:

### Available Variables

| Variable | Description |
|----------|-------------|
| `{{ params.NAME }}` | Input parameter value |
| `{{ steps.ID.stdout }}` | Shell step stdout |
| `{{ steps.ID.stderr }}` | Shell step stderr |
| `{{ steps.ID.exit_code }}` | Shell step exit code |
| `{{ steps.ID.output }}` | Agent step response |
| `{{ env.NAME }}` | Environment variable |

### Fallback Syntax

Use `//` for fallback values:

```yaml
input: "{{ steps.translate.output // steps.summarize.output }}"
```

This uses `translate.output` if available, otherwise `summarize.output`.

### Functions

| Function | Example | Description |
|----------|---------|-------------|
| `trim` | `{{ steps.x.stdout \| trim }}` | Trim whitespace |
| `lines` | `{{ steps.x.stdout \| lines }}` | Split into lines |
| `first` | `{{ steps.x.stdout \| lines \| first }}` | First element |
| `last` | `{{ steps.x.stdout \| lines \| last }}` | Last element |
| `join` | `{{ items \| join "," }}` | Join array with delimiter |

## Conditional Execution

Use `when` to conditionally execute a step:

```yaml
- id: translate
  type: agent
  agent: "@translator"
  prompt: "Translate to {{ params.language }}"
  input: "{{ steps.summarize.output }}"
  when: "{{ params.language != 'english' }}"
```

**Operators:** `==`, `!=`, `<`, `>`, `<=`, `>=`, `&&`, `||`, `!`

## Dependencies

Use `depends_on` to control execution order:

```yaml
steps:
  - id: fetch-a
    type: shell
    run: curl https://api.a.com/data

  - id: fetch-b
    type: shell
    run: curl https://api.b.com/data

  - id: combine
    type: shell
    run: |
      echo "A: {{ steps.fetch-a.stdout }}"
      echo "B: {{ steps.fetch-b.stdout }}"
    depends_on: [fetch-a, fetch-b]
```

Steps without dependencies can run in parallel.

## Input/Output Schemas

Define schemas for validation:

```yaml
input:
  type: object
  properties:
    language:
      type: string
      default: english
    format:
      type: string
      enum: [text, html, markdown]

output:
  type: object
  properties:
    result:
      type: string
```

## Triggers

### Cron Triggers

Run on a schedule:

```yaml
triggers:
  - id: daily-morning
    type: cron
    schedule: "0 9 * * *"  # 9 AM daily
    params:
      language: spanish
```

**Cron Format:** `minute hour day month weekday` (5 fields) or with seconds (6 fields).

### Watch Triggers

Run on filesystem changes:

```yaml
triggers:
  - id: notes-changed
    type: watch
    path: ~/notes
    patterns: ["*.md"]
    recursive: true
    events: [create, modify]
```

**Fields:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `path` | string | - | Directory to watch |
| `patterns` | array | `["*"]` | Glob patterns |
| `recursive` | bool | `false` | Watch subdirectories |
| `events` | array | `[create, modify]` | Events to trigger on |

### Trial Period

New triggers can have a trial period:

```yaml
triggers:
  - id: experimental
    type: cron
    schedule: "0 * * * *"
    runs_before_permanent: 5  # Becomes permanent after 5 successful runs
```

## Complete Example

```yaml
version: 1
name: daily-digest
description: Summarize notes, translate if needed, format as email
created_by: "@ayo"
created_at: 2026-02-08T20:30:00Z

input:
  type: object
  properties:
    language:
      type: string
      default: english

steps:
  - id: gather
    type: shell
    run: |
      find ~/notes -name '*.md' -mtime -1 -exec cat {} \;

  - id: summarize
    type: agent
    agent: "@summarizer"
    prompt: "Extract key points from these notes"
    input: "{{ steps.gather.stdout }}"

  - id: translate
    type: agent
    agent: "@translator"
    prompt: "Translate to {{ params.language }}"
    input: "{{ steps.summarize.output }}"
    when: "{{ params.language != 'english' }}"

  - id: format
    type: agent
    agent: "@formatter"
    prompt: "Format as a professional email"
    input: "{{ steps.translate.output // steps.summarize.output }}"

triggers:
  - id: morning
    type: cron
    schedule: "0 9 * * *"
    params:
      language: spanish

  - id: note-change
    type: watch
    path: ~/notes
    patterns: ["*.md"]
    runs_before_permanent: 10
```

## CLI Commands

```bash
# List flows
ayo flows list

# Show flow details
ayo flows show daily-digest

# Run a flow
ayo flows run daily-digest
ayo flows run daily-digest --param language=spanish
ayo flows run daily-digest -i input.json

# View execution history
ayo flows history
ayo flows history --flow daily-digest
```
