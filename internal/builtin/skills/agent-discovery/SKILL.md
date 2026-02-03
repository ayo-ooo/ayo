---
name: agent-discovery
description: Discover and inspect other agents managed by ayo. Use this before delegating tasks to specialized agents via the CLI.
license: MIT
compatibility: Requires ayo CLI
metadata:
  author: ayo
  version: "2.0"
---

# Agent Discovery Skill

Use this skill to discover available agents before delegating tasks.

## When to Use

Activate this skill when:
- You need to find which agents are available for delegation
- You want to understand what an agent does before calling it
- You're planning to delegate a task to another agent
- The user asks about available agents

## Discovery Commands

### List All Agents

```bash
ayo agents list
```

Shows all available agents grouped by source (user-defined vs built-in).

### Show Agent Details

```bash
ayo agents show @agent-handle
```

Displays full configuration including:
- Description
- Model
- Allowed tools
- Skills
- System prompt
- Input/output schemas (for chainable agents)

## Delegation via CLI

To delegate to another agent, use the ayo CLI via bash:

```bash
# Run a prompt with another agent
ayo @crush "implement feature X"

# Continue a previous session with an agent
ayo @crush -s SESSION_ID "follow up on that"
```

## Delegation Decision Process

Before delegating:

1. **List available agents** to see what's available
2. **Show agent details** for candidates that might help
3. **Match capabilities** - choose the agent whose skills best fit the task
4. **Compose the prompt** - be specific about what you need

## Example Workflow

```bash
# See what agents exist
ayo agents list

# Inspect a promising agent
ayo agents show @crush

# Then delegate via CLI
ayo @crush "create a REST API in Go"
```

## Best Practices

- **Discover first**: Always check what agents exist before delegating
- **Read descriptions**: Agent descriptions indicate their specialization
- **Check tools**: Ensure the target agent has tools needed for the task
- **Be specific**: When calling an agent, provide clear, focused prompts
- **Handle responses**: Process the agent's response appropriately

## Output Interpretation

When `ayo agents list` returns:
- **User-defined**: Agents you or users created
- **Built-in**: Agents shipped with ayo

Agent handles start with `@` (e.g., `@ayo`, `@crush`).
