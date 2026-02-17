# ayo

The root command provides the primary chat interface for interacting with AI agents.

## Synopsis

```
ayo [@agent] [prompt] [flags]
```

## Description

Run AI agents that can execute tasks, use tools, and chain together via Unix pipes. Without arguments, starts an interactive chat session with the default agent (`@ayo`).

## Arguments

| Argument | Description |
|----------|-------------|
| `@agent` | Agent to use (e.g., `@reviewer`, `@ayo`). Optional, defaults to `@ayo` |
| `prompt` | The prompt to send. If omitted, starts interactive mode |

## Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--attachment` | `-a` | strings | File attachments (can be repeated) |
| `--continue` | `-c` | bool | Continue the most recent session |
| `--session` | `-s` | string | Continue a specific session by ID |
| `--model` | `-m` | string | Model to use (overrides config) |
| `--output` | `-o` | string | Target directory for work products |
| `--debug` | | bool | Show debug output including raw tool payloads |

## Examples

### Interactive Chat

```bash
$ ayo
You: Hello!
@ayo: Hello! How can I help you today?
You: 
```

### Single Prompt

```bash
$ ayo "What is the capital of France?"
Paris is the capital of France.
```

### With Specific Agent

```bash
$ ayo @reviewer "Review this code for security issues"
```

### With File Attachment

```bash
$ ayo -a main.go "Explain what this code does"
```

```bash
$ ayo -a report.pdf -a data.csv "Summarize the findings"
```

### Continue Previous Session

```bash
# Continue most recent session
$ ayo -c "Follow up on that last point"

# Continue specific session
$ ayo -s ses_abc123 "What else did we discuss?"
```

### Piped Input

```bash
$ cat error.log | ayo "What went wrong?"

$ git diff | ayo @reviewer "Review these changes"
```

### Chain Multiple Agents

```bash
$ echo "Write a haiku about coding" | ayo @writer | ayo @critic "Rate this writing"
```

## JSON Output

```bash
$ ayo --json "Tell me a joke"
```

```json
{
  "session_id": "ses_x7k9m2p1",
  "response": "Why do programmers prefer dark mode? Because light attracts bugs!",
  "model": "claude-sonnet-4-20250514",
  "tokens": {
    "input": 42,
    "output": 18
  }
}
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error during execution |
| 2 | Invalid arguments |

## See Also

- [ayo agents](cli-agents.md) - Manage agents
- [ayo session](cli-session.md) - Session management
