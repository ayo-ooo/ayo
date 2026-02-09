---
id: ase-yqtq
status: closed
deps: [ase-tnmi]
links: []
created: 2026-02-09T03:06:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-gw5j
---
# Implement @ayo agent creation capability

Enable @ayo to create specialized agents based on task patterns. Created agents get normal top-level names (no dots, no namespacing).

## Background

@ayo creates agents when:
- It recognizes a repeated pattern (similar context used 3+ times)
- User explicitly asks for a specialized agent
- Task requires capabilities no existing agent has

Created agents are tracked in SQLite (ayo_created_agents table) but are otherwise normal agents.

## Implementation

1. Add `ayo agents create` command (internal, used by @ayo):
   ```bash
   ayo agents create --name 'science-researcher' \
     --system-prompt 'You are a science research specialist...' \
     --skills bash,web_search,file_read \
     --sandbox-network true \
     --created-by '@ayo'
   ```

2. Create agent directory structure:
   - ~/.config/ayo/agents/{name}/
   - agent.json with config
   - AGENT.md with system prompt

3. Register in ayo_created_agents table with initial metrics

4. Add @ayo skill/tool for agent creation:
   - Skill provides instructions on when/how to create agents
   - Tool wraps the CLI command

## Agent creation logic (for @ayo's skill):

```markdown
## Creating Specialized Agents

You can create specialized agents when:
1. You've used similar context for the same base task 3+ times
2. The user explicitly asks for a specialized agent
3. A task requires a unique combination of skills

Before creating, check if a similar agent exists:
`ayo agents list`
`ayo agents capabilities --suggest 'the task description'`

To create:
`ayo agents create --name 'descriptive-name' --system-prompt '...' ...`

Keep names concise and descriptive. No dots or special characters.
```

## Files to modify/create

- cmd/ayo/agents.go (add create subcommand)
- internal/agent/create.go (new - creation logic)
- internal/builtin/skills/ayo/agent_creation.md (new skill)

## Acceptance Criteria

- ayo agents create works with all options
- Agent directory created correctly
- Registered in SQLite table
- @ayo skill provides clear guidance
- Created agents work like any other agent

