---
id: ase-7rvn
status: closed
deps: [ase-syl9]
links: []
created: 2026-02-09T03:05:30Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-qnjh
---
# Add Matrix instructions to guardrails

Update the guardrails PREFIX/SUFFIX to include instructions for inter-agent Matrix communication.

## Background

Agents need to know how to communicate with other agents via Matrix. This is part of the safety guardrails so it's always present and cannot be overridden by agent system prompts.

## Guardrails addition

Add to SUFFIX (after agent system prompt):

```markdown
## Inter-Agent Communication

You are part of a multi-agent system coordinated by @ayo. You communicate via Matrix chat.

### Commands
- \`ayo chat read '#session-{id}'\` - Check for messages from other agents
- \`ayo chat send '#session-{id}' 'your message'\` - Send a message  
- \`ayo chat who '#session-{id}'\` - See who else is in the session

### Protocol
1. When you start a task, check for context: \`ayo chat read '#session-{id}' --limit 10\`
2. Report progress: \`ayo chat send '#session-{id}' 'Starting summarization...'\`
3. Report completion with structured output:
   \`ayo chat send '#session-{id}' '{"status": "complete", "output": {...}}'\`
4. If you encounter an error, attempt recovery before reporting failure
5. Check for messages periodically during long tasks

### Your Session Room
{session_room}

### Your Identity  
You are {agent_handle}. Messages you send appear as from you.
```

## Implementation

1. Find existing guardrails implementation
2. Add Matrix instructions section
3. Template in session_room and agent_handle at runtime
4. Only include if agent is in an orchestrated session (has AYO_SESSION_ROOM)

## Files to modify

- internal/agent/guardrails.go (or wherever guardrails are applied)
- Ensure session context is available when building prompt

## Acceptance Criteria

- Matrix instructions included in guardrails
- Session room templated correctly
- Agent handle templated correctly
- Only included when in orchestrated session

