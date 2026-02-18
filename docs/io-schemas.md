# I/O Schemas

> **Note**: This document focuses on agent I/O schema definitions. For orchestrating agents in pipelines, see [Flows](flows.md).

Agents with structured I/O schemas can be composed via Unix pipes or through flow steps. The output of one agent becomes the input to the next.

## Overview

```bash
ayo @code-reviewer '{"files":["main.go"]}' | ayo @issue-reporter
```

When piping:
- UI (spinners, reasoning, tool calls) goes to stderr
- Raw JSON output goes to stdout for downstream consumption

## Structured I/O Schemas

Agents can define optional JSON schemas:

| File | Purpose |
|------|---------|
| `input.jsonschema` | Validates input; agent only accepts JSON matching this schema |
| `output.jsonschema` | Structures output; final response is formatted as this schema |

### Agent Structure

```
@my-agent/
├── config.json
├── system.md
├── input.jsonschema    # Optional
└── output.jsonschema   # Optional
```

### Example Input Schema

```json
{
  "type": "object",
  "properties": {
    "files": {
      "type": "array",
      "items": { "type": "string" },
      "description": "List of files to analyze"
    },
    "options": {
      "type": "object",
      "properties": {
        "verbose": { "type": "boolean" }
      }
    }
  },
  "required": ["files"]
}
```

### Example Output Schema

```json
{
  "type": "object",
  "properties": {
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "file": { "type": "string" },
          "line": { "type": "integer" },
          "severity": { "type": "string" },
          "message": { "type": "string" }
        }
      }
    },
    "summary": { "type": "string" }
  }
}
```

## Schema Commands

### List Agents with Schemas

```bash
ayo flow schema ls
```

Shows agents with input or output schemas.

### Inspect Schemas

```bash
ayo flow schema inspect @my-agent
```

Shows input and output schemas with descriptions.

```bash
ayo flow schema inspect @my-agent --json
```

### Find Compatible Agents

```bash
# What can receive this agent's output?
ayo flow schema from @code-reviewer

# What can feed into this agent?
ayo flow schema to @issue-reporter
```

### Validate Input

```bash
# Validate JSON against input schema
ayo flow schema validate @my-agent '{"files": ["main.go"]}'

# Or via stdin
echo '{"files": ["main.go"]}' | ayo flow schema validate @my-agent
```

### Generate Example

```bash
ayo flow schema example @my-agent
```

Generates example JSON matching the input schema.

## Schema Compatibility

When piping agents or connecting flow steps, schemas are checked for compatibility:

1. **Exact match**: Output schema identical to input schema
2. **Structural match**: Output has all required fields of input (superset OK)
3. **Freeform**: Target agent has no input schema (accepts anything)

If incompatible, validation fails with a clear error.

## Creating Agents with Schemas

### Step 1: Create Schemas

```bash
# Input schema
cat > input.jsonschema << 'EOF'
{
  "type": "object",
  "properties": {
    "code": { "type": "string", "description": "Code to analyze" },
    "language": { "type": "string", "description": "Programming language" }
  },
  "required": ["code"]
}
EOF

# Output schema
cat > output.jsonschema << 'EOF'
{
  "type": "object",
  "properties": {
    "issues": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "line": { "type": "integer" },
          "message": { "type": "string" }
        }
      }
    }
  }
}
EOF
```

### Step 2: Create Agent

```bash
ayo agents create @analyzer \
  -m gpt-5.2 \
  -s "Analyze code and return issues as JSON." \
  --input-schema input.jsonschema \
  --output-schema output.jsonschema
```

### Step 3: Test

```bash
# Verify schemas
ayo flow schema inspect @analyzer

# Validate input
ayo flow schema validate @analyzer '{"code": "print(x)", "language": "python"}'

# Run agent
ayo @analyzer '{"code": "print(x)", "language": "python"}'
```

## Piping Example

### Two-Agent Pipeline

```bash
# Reviewer finds issues
ayo @code-reviewer '{"files": ["main.go"]}' \
  | ayo @issue-reporter
```

### Multi-Step Pipeline

```bash
# Analyze -> Prioritize -> Report
ayo @analyzer '{"code": "..."}' \
  | ayo @prioritizer \
  | ayo @reporter
```

## Pipeline Context

When agents are piped, context is passed via environment variable:

```
AYO_PIPELINE_CONTEXT={"depth":1,"source":"@code-reviewer","source_description":"Code review agent"}
```

This allows downstream agents to understand the pipeline context.

Freeform agents (without input schema) receive a preamble describing the pipeline context.

## Pipeline Behavior

| Condition | UI Output | JSON Output |
|-----------|-----------|-------------|
| stdout is terminal | Full UI | Rendered |
| stdout is pipe | stderr only | Raw JSON to stdout |
| stdin is pipe | Read JSON input | N/A |

The full UI (spinners, reasoning, tool calls) is always visible on stderr.

## Tips

- Use `--json` flag on schema commands for machine-readable output
- Test schemas with `ayo flow schema validate` before running pipelines
- Use `ayo flow schema example` to generate test input
- Schema discovery shows which agents can connect

## Deprecated Commands

The following commands are deprecated and will be removed in a future release:

| Deprecated | Replacement |
|------------|-------------|
| `ayo chain ls` | `ayo flow schema ls` |
| `ayo chain inspect` | `ayo flow schema inspect` |
| `ayo chain from` | `ayo flow schema from` |
| `ayo chain to` | `ayo flow schema to` |
| `ayo chain validate` | `ayo flow schema validate` |
| `ayo chain example` | `ayo flow schema example` |
