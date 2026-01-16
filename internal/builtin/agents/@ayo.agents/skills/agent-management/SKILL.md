---
name: agent-management
description: >
  Specialized knowledge for creating, configuring, and managing ayo agents.
  Use when helping users design agent system prompts, configure agent options,
  or troubleshoot agent behavior.
metadata:
  author: ayo
  version: "1.0"
---

# Agent Management Skill

This skill provides specialized knowledge for creating and managing ayo agents effectively.

## Creating Effective Agents

### System Prompt Design

A good system prompt should:

1. **Define the role clearly**: Start with "You are..." to establish identity
2. **Set boundaries**: Explain what the agent should and shouldn't do
3. **Provide context**: Include relevant background information
4. **Give examples**: Show expected input/output patterns
5. **Define output format**: Specify how responses should be structured

### System Prompt Template

```markdown
# {Role Title}

You are {specific role description}. {Additional context about the role}.

## Your Responsibilities

- {Primary task}
- {Secondary task}
- {Additional tasks}

## Guidelines

- {Important behavioral guideline}
- {Quality standard}
- {Constraint or limitation}

## Output Format

{Describe how responses should be structured}

## Examples

### Example 1: {Scenario}

Input: {example input}

Output: {example output}
```

### Agent Configuration Patterns

**Minimal agent** (just bash):
```json
{
  "description": "Simple automation agent",
  "allowed_tools": ["bash"]
}
```

**Agent with delegation** (can call other agents):
```json
{
  "description": "Orchestrator agent",
  "allowed_tools": ["bash", "agent_call"]
}
```

**Skill-focused agent**:
```json
{
  "description": "Debugging specialist",
  "allowed_tools": ["bash"],
  "skills": ["debugging"],
  "ignore_builtin_skills": true
}
```

**Chainable agent** (structured I/O):
```json
{
  "description": "Code analyzer",
  "allowed_tools": ["bash"]
}
```
Plus `input.jsonschema` and/or `output.jsonschema` files.

## Common Agent Patterns

### Task-Specific Agent

For agents that do one thing well:

```bash
ayo agents create @formatter -n \
  -m gpt-4.1-mini \
  -d "Code formatting and style fixes" \
  -s "You format and fix code style issues. Output only the corrected code." \
  -t bash
```

### Research Agent

For agents that gather information:

```bash
ayo agents create @researcher -n \
  -m gpt-4.1 \
  -d "Research and summarize topics" \
  -s "You research topics thoroughly and provide structured summaries with sources." \
  -t bash
```

### Review Agent

For agents that analyze and critique:

```bash
ayo agents create @reviewer -n \
  -m gpt-4.1 \
  -d "Code review and feedback" \
  -s "You review code for bugs, security issues, and best practices. Be constructive." \
  -t bash
```

### Orchestrator Agent

For agents that coordinate other agents:

```bash
ayo agents create @orchestrator -n \
  -m gpt-4.1 \
  -d "Coordinates multiple agents for complex tasks" \
  -s "You break down complex tasks and delegate to specialized agents using agent_call." \
  -t bash,agent_call
```

## Agent-Specific Skills

Agents can have their own skills in the `skills/` subdirectory:

```
@my-agent/
├── config.json
├── system.md
└── skills/
    └── my-special-skill/
        └── SKILL.md
```

These skills are only available to this specific agent and take highest priority in skill discovery.

### When to Use Agent-Specific Skills

- Knowledge only relevant to one agent
- Custom workflows for a specific purpose
- Overriding built-in skill behavior for one agent

## Troubleshooting Agents

### Agent gives generic responses

**Cause**: System prompt is too vague
**Fix**: Add specific instructions, examples, and constraints

### Agent doesn't use tools

**Cause**: Tools not in `allowed_tools`
**Fix**: Add required tools to config.json

### Agent ignores skills

**Cause**: Skills not configured or excluded
**Fix**: 
1. Check `ayo agents show @agent` for loaded skills
2. Verify skill name in `skills` array
3. Check skill isn't in `exclude_skills`

### Agent behaves unsafely

**Cause**: `no_system_wrapper` is enabled
**Fix**: Remove `no_system_wrapper` or set to false

## Updating and Maintaining Agents

### Editing an Agent

User agents are stored in `~/.config/ayo/agents/@{name}/`. Edit files directly:

```bash
# Edit system prompt
$EDITOR ~/.config/ayo/agents/@my-agent/system.md

# Edit configuration
$EDITOR ~/.config/ayo/agents/@my-agent/config.json
```

### Removing an Agent

```bash
rm -rf ~/.config/ayo/agents/@agent-name
```

### Backing Up Agents

```bash
cp -r ~/.config/ayo/agents ~/ayo-agents-backup
```
