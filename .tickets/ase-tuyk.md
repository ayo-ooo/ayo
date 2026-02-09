---
id: ase-tuyk
status: closed
deps: []
links: []
created: 2026-02-09T03:06:28Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-gw5j
---
# Implement invocation context (ephemeral instructions)

Enable passing task-specific context when invoking an agent, without creating a new agent.

## Background

'Ephemeral agents' are really just invocation context - instructions passed to an existing agent for a specific task. This avoids agent sprawl while still allowing specialization.

Example:
```
@ayo invoking @researcher:
  'Focus on quantum computing papers for this task.
   Prioritize recent publications from 2025-2026.'
```

The @researcher agent receives this as additional context, adapts behavior, but remains @researcher.

## Implementation

1. When @ayo orchestrates via Matrix, the message IS the invocation context:
   ```
   @ayo → #session-xyz:
   '@researcher Focus on quantum computing papers. 
    Prioritize 2025-2026. Here is the task: {actual task}'
   ```

2. No code changes needed for basic case - it's just prompting

3. For flow steps, add optional 'context' field:
   ```yaml
   steps:
     - id: research
       agent: '@researcher'
       context: 'Focus on quantum computing papers'
       prompt: 'Find recent advances in {topic}'
   ```

4. Context is prepended to prompt when invoking agent

## Documentation

Add to @ayo skill:
```markdown
## Invocation Context vs New Agents

Before creating a new agent, consider if invocation context is enough:

INVOCATION CONTEXT (prefer this):
- One-off specialization
- Simple focus adjustment
- First few times trying something

NEW AGENT (only when needed):
- Pattern used 3+ times successfully
- Requires persistent refinement
- User explicitly requests

To use context, just include instructions in your message to the agent:
'@summarizer For this task, focus on technical details and use bullet points.'
```

## Files to modify

- internal/flows/execute.go (handle context field)
- internal/builtin/skills/ayo/SKILL.md (document pattern)

## Acceptance Criteria

- Context field works in flow steps
- Context prepended to prompt correctly
- @ayo skill explains when to use context vs create agent

