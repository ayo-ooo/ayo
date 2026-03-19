# Quickstart

Build your first AI agent in 5 minutes.

## Prerequisites

- Ayo CLI installed ([Installation](installation.md))
- API key configured

## Create the Echo Agent

The simplest agent takes a string input and returns a response.

### 1. Create Project Directory

```bash
mkdir echo-agent && cd echo-agent
```

### 2. Create config.toml

```toml
[agent]
name = "echo"
version = "1.0.0"
description = "Echo agent example"

[model]
suggested = ["anthropic/claude-3.5-sonnet", "openai/gpt-4o"]
```

### 3. Create system.md

```markdown
# Echo Agent

You are a helpful assistant that echoes back messages with enthusiasm.

When given input, respond with a friendly, enthusiastic version of the message.
```

### 4. Create input.jsonschema

```json
{
  "type": "object",
  "properties": {
    "message": {
      "type": "string",
      "description": "Message to echo",
      "x-cli-position": 1
    }
  },
  "required": ["message"]
}
```

### 5. Build

```bash
ayo build .
```

This generates an `echo` executable in the current directory.

### 6. Run

```bash
./echo "Hello, World!"
```

Output:
```
🎉 HELLO, WORLD! 🎉
```

## What Happened

1. Ayo read your configuration, system prompt, and input schema
2. Generated a standalone Go binary with:
   - CLI flag parsing for the `message` input
   - Type-safe input validation
   - LLM integration configured
3. The binary calls the LLM with your system prompt and user input
4. Returns the LLM's response

## Next Steps

- [First Agent](first-agent.md) - Learn about all project files
- [Input Schema](../reference/input-schema.md) - Define complex inputs
- [Output Schema](../reference/output-schema.md) - Structure responses
