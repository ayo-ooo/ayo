# Building Agents

Best practices for designing and building Ayo agents.

## Design Process

### 1. Define the Purpose

Start with a clear problem statement:

```
I need an agent that [action] for [audience] to achieve [outcome].
```

Example:
```
I need an agent that reviews code for developers to improve code quality.
```

### 2. Identify Inputs

What does the agent need to know?

| Input | Type | Required | Notes |
|-------|------|----------|-------|
| code | string (file) | yes | Source code to review |
| language | enum | no | Detect automatically |
| severity | enum | no | Filter issues |

### 3. Define Outputs

What should the agent produce?

| Output | Type | Description |
|--------|------|-------------|
| issues | array | List of problems found |
| score | integer | Quality score 0-100 |
| summary | string | Brief overview |

### 4. Design the Prompt

Structure your system prompt:

```markdown
# Code Review Agent

You are a code review assistant.

## Purpose
Identify issues and suggest improvements.

## Capabilities
- Detect bugs and errors
- Suggest improvements
- Check style consistency

## Output Format
Return JSON matching the output schema.
```

## Building Steps

### 1. Create Project

```bash
mkdir my-agent && cd my-agent
```

### 2. Create Configuration

```toml
# config.toml
[agent]
name = "my-agent"
version = "1.0.0"
description = "Agent description"

[model]
suggested = ["anthropic/claude-3.5-sonnet"]

[defaults]
temperature = 0.7
```

### 3. Create System Prompt

```markdown
# system.md
You are a helpful assistant that...
```

### 4. Define Input Schema

```json
{
  "type": "object",
  "properties": {
    "input": {
      "type": "string",
      "description": "Primary input",
      "x-cli-position": 1
    }
  },
  "required": ["input"]
}
```

### 5. Define Output Schema

```json
{
  "type": "object",
  "properties": {
    "result": { "type": "string" }
  }
}
```

### 6. Build and Test

```bash
ayo build .
./my-agent "test input"
```

## Iteration

### Testing

1. Test with various inputs
2. Check edge cases
3. Validate output format

### Refinement

1. Adjust prompts based on results
2. Add validation constraints
3. Improve error handling

## Common Patterns

### Text Processing

```json
{
  "properties": {
    "text": { "type": "string", "x-cli-position": 1 },
    "format": { "type": "string", "default": "text" }
  }
}
```

### File Processing

```json
{
  "properties": {
    "file": { "type": "string", "x-cli-position": 1, "x-cli-file": true }
  }
}
```

### Multi-Step Processing

Use skills for complex workflows:

```
skills/
├── analyze/
│   └── SKILL.md
├── process/
│   └── SKILL.md
└── validate/
    └── SKILL.md
```

## Next Steps

- [Structured Outputs](structured-outputs.md) - Working with JSON schemas
- [File Processing](file-processing.md) - Handling file inputs
- [Best Practices](best-practices.md) - Production patterns
