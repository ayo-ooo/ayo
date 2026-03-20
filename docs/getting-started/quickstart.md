# Quickstart

Build your first AI agent in 5 minutes.

## Prerequisites

- Ayo CLI installed ([Installation](installation.md))
- API key configured

## Create the Echo Agent

The simplest agent takes a string input and returns a response.

### 1. Generate Project

```bash
ayo fresh echo-agent
cd echo-agent
```

This creates:
- `config.toml` - Agent configuration
- `system.md` - System prompt
- `.gitignore` - Common ignore patterns

### 2. Customize system.md

Edit `system.md` to define your agent's behavior:

```markdown
# Echo Agent

You are a helpful assistant that echoes back messages with enthusiasm.

When given input, respond with a friendly, enthusiastic version of the message.
```

### 3. Create input.jsonschema

Create `input.jsonschema` to define the input:

```json
{
  "type": "object",
  "properties": {
    "message": {
      "type": "string",
      "description": "Message to echo"
    }
  },
  "required": ["message"]
}
```

### 4. Build

```bash
ayo runthat .
```

This generates an `echo-agent` executable in the current directory.

### 5. Run

```bash
./echo-agent '{"message": "Hello, World!"}'
```

Output:
```
🎉 HELLO, WORLD! 🎉
```

## JSON Input

Generated CLIs accept JSON input as the first positional argument:

```bash
# From a file
./echo-agent input.json

# From stdin
echo '{"message": "Hello"}' | ./echo-agent -

# Inline JSON
./echo-agent '{"message": "Hello"}'
```

## Flag Overrides

Flags let you override specific fields from the command line:

```bash
# Override just the message field
./echo-agent --message "Hello, World!"
```

Only primitive types (strings, numbers, integers, booleans) get flag overrides.

## What Happened

1. `ayo fresh` scaffolded a new agent project with default files
2. Ayo read your configuration, system prompt, and input schema
3. Generated a standalone Go binary with:
   - JSON input parsing (positional or stdin)
   - Flag overrides for primitive fields
   - Type-safe input validation
   - LLM integration configured
4. The binary calls the LLM with your system prompt and user input
5. Returns the LLM's response

## Next Steps

- [First Agent](first-agent.md) - Learn about all project files
- [Input Schema](../reference/input-schema.md) - Define complex inputs
- [Output Schema](../reference/output-schema.md) - Structure responses
