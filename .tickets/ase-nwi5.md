---
id: ase-nwi5
status: closed
deps: [ase-yqtq, ase-deny]
links: []
created: 2026-02-09T03:25:08Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-gw5j
---
# Update @ayo builtin agent for orchestration

## Background

@ayo is the executive agent that orchestrates other agents. With the new orchestration system, @ayo needs updated skills and instructions for:
- Creating new agents
- Refining agents it created
- Using flows
- Finding agents by capability
- Communicating via Matrix

## Why This Matters

@ayo's behavior is defined by its AGENT.md and associated skills. Without updates, @ayo won't know it can:
- Create specialized agents
- Use the find_agent tool
- Create and run flows
- Communicate with other agents via Matrix

## Implementation Details

### Files to Update

1. `internal/builtin/agents/ayo/AGENT.md` - Main system prompt

2. `internal/builtin/skills/ayo/orchestration.md` - New skill for orchestration

### Updated AGENT.md Content

Add sections for:

```markdown
## Agent Orchestration

You are the executive agent responsible for coordinating other agents to complete complex tasks.

### Discovering Agents

Use the `find_agent` tool to discover agents capable of performing specific tasks:
\`\`\`
<find_agent task="review code for security issues" count="3"/>
\`\`\`

### Creating Specialized Agents

When you notice a pattern that would benefit from a dedicated agent:
1. Verify no existing agent has this capability
2. Use `ayo agents create` to create the agent
3. Track the new agent's performance

Only create agents when:
- Similar context used 3+ times
- User explicitly requests
- No existing agent matches the need

### Refining Agents

For agents you created, you can refine their prompts based on performance:
\`\`\`bash
ayo agents refine <agent-name> --prompt "Updated system prompt..."
\`\`\`

### Using Flows

When a multi-step pattern proves successful, save it as a flow:
\`\`\`bash
ayo flows create --from-session <session-id>
\`\`\`

### Communication

Communicate with other agents via Matrix:
\`\`\`bash
ayo chat send '#session-room' '@agent message'
ayo chat read '#session-room' --follow
\`\`\`
```

### New Tools for @ayo

| Tool | Purpose |
|------|---------|
| find_agent | Discover agents by capability |
| create_agent | Create a new specialized agent |
| refine_agent | Update an agent's system prompt |
| create_flow | Save a pattern as a reusable flow |

### Decision Framework

Add decision tree for when to:
- Use existing agent
- Add invocation context
- Create new agent
- Create flow

## Acceptance Criteria

- [ ] AGENT.md updated with orchestration instructions
- [ ] orchestration.md skill file created
- [ ] find_agent tool documented
- [ ] Agent creation guidelines included
- [ ] Agent refinement guidelines included
- [ ] Flow creation guidelines included
- [ ] Matrix communication documented
- [ ] Decision framework for agent vs context vs flow

