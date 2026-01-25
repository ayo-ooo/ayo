# Delegation

Delegation allows agents to route specific task types to specialized agents. For example, `@ayo` can delegate coding tasks to `@crush` when configured.

## Overview

When delegation is configured:

1. User sends a request to `@ayo`
2. `@ayo` recognizes it as a coding task
3. `@ayo` delegates to `@crush` via `agent_call` tool
4. `@crush` handles the task
5. Result returns to user

## Task Types

| Type | Description |
|------|-------------|
| `coding` | Source code creation/modification |
| `research` | Web research and information gathering |
| `debug` | Debugging and troubleshooting |
| `test` | Test creation and execution |
| `docs` | Documentation generation |

## Configuration Priority

Delegation is resolved from three sources (highest priority first):

1. **Directory config** - `.ayo.json` in project or parent
2. **Agent config** - `delegates` in user agent's `config.json`
3. **Global config** - `~/.config/ayo/ayo.json`

## Setting Up Delegation

### Project-Level (Recommended)

Create `.ayo.json` in your project root:

```json
{
  "delegates": {
    "coding": "@crush",
    "research": "@research"
  }
}
```

Ayo searches from current directory up to find this file.

### Global Configuration

Add to `~/.config/ayo/ayo.json`:

```json
{
  "delegates": {
    "coding": "@crush",
    "research": "@research"
  }
}
```

### Agent-Level

For user-defined agents, add to `config.json`:

```json
{
  "delegates": {
    "coding": "@crush"
  }
}
```

**Note:** Built-in agents do not support the `delegates` field. Use directory or global config for built-in agents.

## Directory Config Options

The `.ayo.json` file supports additional settings:

```json
{
  "agent": "@ayo",
  "model": "gpt-4.1",
  "delegates": {
    "coding": "@crush",
    "research": "@research"
  }
}
```

| Field | Description |
|-------|-------------|
| `agent` | Default agent for this directory |
| `model` | Override default model |
| `delegates` | Task type to agent mappings |

## Plugin-Provided Delegates

Plugins can declare delegates in their `manifest.json`:

```json
{
  "name": "crush",
  "delegates": {
    "coding": "@crush"
  }
}
```

When installed, you're prompted to set these as global defaults:

```
? Set @crush as default for coding tasks? [Y/n]
```

## How Delegation Works

### With agent_call Tool

The agent must have `agent_call` in its allowed tools:

```json
{
  "allowed_tools": ["bash", "agent_call"]
}
```

When delegating, the agent uses the `agent_call` tool:

```json
{
  "agent": "@crush",
  "prompt": "Refactor the authentication module to use JWT tokens"
}
```

### Session Tracking

Delegated work creates linked sessions:

```bash
ayo sessions list --source crush-via-ayo
```

This shows sessions where `@crush` was called through `@ayo`.

## Example Workflow

### 1. Install Plugins

```bash
ayo plugins install https://github.com/alexcabrera/ayo-plugins-crush
ayo plugins install https://github.com/alexcabrera/ayo-plugins-research
```

### 2. Configure Delegation

```bash
cat > .ayo.json << 'EOF'
{
  "delegates": {
    "coding": "@crush",
    "research": "@research"
  }
}
EOF
```

### 3. Use Naturally

```bash
ayo "Refactor the authentication module"
# → @ayo delegates to @crush

ayo "Research best practices for JWT tokens"
# → @ayo delegates to @research
```

## Viewing Effective Delegates

Check what delegates are active:

```bash
ayo doctor -v
```

Or inspect an agent:

```bash
ayo agents show @ayo
```

## Disabling Delegation

### Per-Request

Use the specific agent directly:

```bash
ayo @ayo "write this code yourself"  # Won't delegate
```

### Per-Project

Set empty delegates in `.ayo.json`:

```json
{
  "delegates": {}
}
```

### Globally

Remove delegates from `~/.config/ayo/ayo.json`.

## Delegation vs Tool Aliases

| Feature | Delegation | Tool Aliases |
|---------|-----------|--------------|
| Purpose | Route task types | Swap tool implementations |
| Level | Full task handling | Single tool call |
| Example | `coding` → `@crush` | `search` → `searxng` |
| Config | `delegates` | `default_tools` |

Use delegation for complex tasks that benefit from specialized agents.
Use tool aliases for swappable implementations of specific tools.
