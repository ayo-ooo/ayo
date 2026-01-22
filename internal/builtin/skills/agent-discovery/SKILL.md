---
name: agent-discovery
description: Discover and inspect other agents managed by ayo. Required when using the agent_call tool to delegate tasks to specialized agents.
license: MIT
compatibility: Requires ayo CLI
metadata:
  author: ayo
  version: "1.0"
---

# Agent Discovery Skill

Use this skill to discover available agents before delegating tasks via `agent_call`.

## When to Use

Activate this skill when:
- You need to find which agents are available for delegation
- You want to understand what an agent does before calling it
- You're planning to use `agent_call` to delegate a task
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

## Delegation Decision Process

Before using `agent_call`:

1. **List available agents** to see what's available
2. **Show agent details** for candidates that might help
3. **Match capabilities** - choose the agent whose skills best fit the task
4. **Compose the prompt** - be specific about what you need

## Example Workflow

```bash
# See what agents exist
ayo agents list

# Inspect a promising agent
ayo agents show @ayo

# Then use agent_call with confidence
# agent_call(agent="@ayo", prompt="...")
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

Agent handles start with `@` (e.g., `@ayo`, `@code-reviewer`).
