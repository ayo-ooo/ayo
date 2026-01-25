# Agent Chaining

Agents with structured I/O schemas can be composed via Unix pipes. The output of one agent becomes the input to the next.

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

## Chain Commands

### List Chainable Agents

```bash
ayo chain ls
```

Shows agents with input or output schemas.

### Inspect Schemas

```bash
ayo chain inspect @my-agent
```

Shows input and output schemas with descriptions.

```bash
ayo chain inspect @my-agent --json
```

### Find Compatible Agents

```bash
# What can receive this agent's output?
ayo chain from @code-reviewer

# What can feed into this agent?
ayo chain to @issue-reporter
```

### Validate Input

```bash
# Validate JSON against input schema
ayo chain validate @my-agent '{"files": ["main.go"]}'

# Or via stdin
echo '{"files": ["main.go"]}' | ayo chain validate @my-agent
```

### Generate Example

```bash
ayo chain example @my-agent
```

Generates example JSON matching the input schema.

## Schema Compatibility

When piping agents, schemas are checked for compatibility:

1. **Exact match**: Output schema identical to input schema
2. **Structural match**: Output has all required fields of input (superset OK)
3. **Freeform**: Target agent has no input schema (accepts anything)

If incompatible, validation fails with a clear error.

## Creating Chainable Agents

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
  -m gpt-4.1 \
  -s "Analyze code and return issues as JSON." \
  --input-schema input.jsonschema \
  --output-schema output.jsonschema
```

### Step 3: Test

```bash
# Verify schemas
ayo chain inspect @analyzer

# Validate input
ayo chain validate @analyzer '{"code": "print(x)", "language": "python"}'

# Run agent
ayo @analyzer '{"code": "print(x)", "language": "python"}'
```

## Chaining Example

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

## Chain Context

When agents are chained, context is passed via environment variable:

```
AYO_CHAIN_CONTEXT={"depth":1,"source":"@code-reviewer","source_description":"Code review agent"}
```

This allows downstream agents to understand the pipeline context.

Freeform agents (without input schema) receive a preamble describing the chain context.

## Pipeline Behavior

| Condition | UI Output | JSON Output |
|-----------|-----------|-------------|
| stdout is terminal | Full UI | Rendered |
| stdout is pipe | stderr only | Raw JSON to stdout |
| stdin is pipe | Read JSON input | N/A |

The full UI (spinners, reasoning, tool calls) is always visible on stderr.

## Tips

- Use `--json` flag on chain commands for machine-readable output
- Test schemas with `ayo chain validate` before running pipelines
- Use `ayo chain example` to generate test input
- Chain discovery shows which agents can connect
