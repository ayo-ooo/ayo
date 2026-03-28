# ayo-invoke: Run ayo agents as tools

Use this skill when the user wants to run an ayo agent, delegate a task to an agent, or use an agent as a tool in a workflow.

## Overview

Ayo agents are standalone CLI binaries. They accept input via JSON arguments, CLI flags, or stdin pipes. Use `ayo run <agent>` to invoke by name, or call the binary directly.

## Invocation Workflow

### 1. Discover the agent's capabilities

```bash
ayo describe <agent> --json
```

Parse the response to understand:
- `type`: "tool" (one-shot, has input schema) or "conversational" (chat-based)
- `input_schema`: what fields are required/optional and their types
- `output_schema`: if present, output will be structured JSON

### 2. Construct the invocation

**Tool agents** (have input_schema):

```bash
# JSON argument
ayo run <agent> '{"field": "value", "other": 123}'

# CLI flags
ayo run <agent> --field "value" --other 123

# Stdin pipe
echo '{"field": "value"}' | ayo run <agent>

# File input (if schema has "file": true)
ayo run <agent> --field @path/to/file.txt

# Mixed: JSON base + flag overrides (flags win)
ayo run <agent> '{"field": "a"}' --field "b"
```

**Conversational agents** (no input_schema):

```bash
# Direct prompt
ayo run <agent> "Your question or task here"

# Stdin pipe
echo "Your question" | ayo run <agent>
```

### 3. Parse the output

**If the agent has an output_schema**: output is JSON on stdout. Parse it directly.

```bash
result=$(ayo run translator '{"text": "hello", "target_language": "es"}')
echo "$result" | jq '.translated_text'
```

**If no output_schema**: output is plain text on stdout.

```bash
result=$(ayo run summarize "Summarize this document...")
echo "$result"
```

### 4. Save output to a file

```bash
ayo run <agent> '{"input": "data"}' -o output.json
```

## Important Flags

Always use these when invoking programmatically:

- `--non-interactive` — prevents TUI forms from launching (critical for non-TTY contexts)
- `-o <path>` — write output to file
- `--provider <name>` — override LLM provider
- `--model <name>` — override model

## Invocation Patterns

### Direct binary execution (fastest)

If you know the binary path from `ayo describe --json`:

```bash
/path/to/agent --non-interactive '{"field": "value"}'
```

### Via ayo run (name-based lookup)

```bash
ayo run agent-name '{"field": "value"}'
```

### Capturing structured output

```bash
output=$(ayo run formatter '{"data": "...", "format": "json"}' --non-interactive 2>/dev/null)
echo "$output" | jq '.result'
```

### Error handling

Check exit codes:
- `0` — success
- `1` — agent error (check stderr)
- `127` — binary not found

```bash
if ! result=$(ayo run agent '{"input": "test"}' --non-interactive 2>/tmp/agent-err); then
    echo "Agent failed: $(cat /tmp/agent-err)"
fi
```

## Example: Using an Agent as a Tool

```bash
# 1. Check what's available
ayo list --json --type tool

# 2. Find the right agent
ayo describe code-reviewer --json | jq '.description'

# 3. Check required fields
ayo describe code-reviewer --json | jq '.input_schema.required'

# 4. Invoke it
result=$(ayo run code-reviewer '{"path": "./src/main.go", "language": "go"}' --non-interactive)

# 5. Use the result
echo "$result"
```

## Notes

- Tool agents with required fields will error if invoked without them and `--non-interactive` is set
- Conversational agents support `--session <id>` to resume conversations
- All agents have a sandboxed shell tool — they can read files and run safe commands
- First run of an agent triggers model selection; set `--provider` and `--model` to skip this
- Agent configs are cached at `~/.config/agents/<name>.toml` after first setup
