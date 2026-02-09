# Orchestration Skill

This skill enables @ayo to orchestrate other agents effectively.

## Overview

As the executive agent, @ayo coordinates specialized agents to complete complex tasks. This involves:
- Discovering agents by capability
- Creating new specialized agents
- Refining agents based on performance
- Managing agent lifecycle (archive/promote)
- Creating reusable flows

## Tools Available

### find_agent

Find agents capable of performing a task:

```json
{
  "task": "review code for security vulnerabilities",
  "count": 3
}
```

**Returns:**
- Ranked list of agents with similarity scores
- Matching capability for each agent
- Capability description

**When to use:**
- Before delegating a task
- When unsure which agent is best suited
- To verify no existing agent has a capability before creating one

### Agent CLI Commands

**Discovery and Information:**
```bash
ayo agents list                           # List all agents
ayo agents show @agent                    # Agent details
ayo agents capabilities @agent            # Agent capabilities
ayo agents capabilities --search "term"   # Find by capability
```

**Creation:**
```bash
ayo agents create @name -m model -d "desc" -f system.md
```

**Lifecycle:**
```bash
ayo agents promote @old @new              # Promote to user-owned
ayo agents archive @agent                 # Hide from listings
ayo agents unarchive @agent               # Restore
```

## Decision Framework

### Should I Create an Agent?

```
Is this a one-off task?
  → YES: Handle directly or delegate to existing agent
  → NO: Continue...

Does an existing agent handle this?
  → Use find_agent to check
  → YES: Delegate to existing agent
  → NO: Continue...

Has this pattern repeated 3+ times?
  → YES: Create specialized agent
  → NO: Handle directly, track usage

Did user request a dedicated agent?
  → YES: Create it
  → NO: Handle directly
```

### Agent Creation Best Practices

1. **Specific system prompts**: Include domain expertise, constraints, and examples
2. **Appropriate tools**: Only include tools the agent needs
3. **Track creation**: Always use `--created-by "@ayo"` flag
4. **Document reason**: Use `--creation-reason` to explain why

### When to Archive Agents

Archive agents that:
- Have 0 invocations in 30+ days
- Have low confidence scores (< 0.3)
- Were superseded by better agents
- Were created for one-time experiments

### When to Promote Agents

Promote agents when:
- User wants ownership
- Agent is stable and well-tested
- User wants to customize beyond @ayo's scope

## Flows

Flows are reusable multi-step workflows defined in YAML.

### Creating Flows

When a successful pattern emerges:
```bash
# Create from session history
ayo flows create --from-session SESSION_ID

# Or create manually
cat > ~/.config/ayo/flows/my-flow.yaml << 'EOF'
name: my-flow
description: Description of what this flow does
steps:
  - id: step1
    agent: "@agent-name"
    prompt: "First step prompt"
  - id: step2
    agent: "@another-agent"
    prompt: "Second step using {{ steps.step1.output }}"
EOF
```

### Running Flows

```bash
ayo flows run my-flow
ayo flows run my-flow --input '{"key": "value"}'
```

## Communication Patterns

### Delegating Tasks

```bash
# Simple delegation
ayo @agent "task description"

# With session continuity
ayo @agent -s SESSION_ID "follow-up"
```

### Multi-Agent Collaboration

For complex tasks requiring multiple agents:
1. Use `find_agent` to identify capable agents
2. Break task into subtasks
3. Delegate each subtask to appropriate agent
4. Aggregate results

## Metrics and Refinement

Track agent performance:
- **Invocation count**: How often the agent is used
- **Success rate**: Percentage of successful completions
- **Confidence score**: System's confidence in agent capability

### When to Refine Agents

Refine agents you created when:
- User corrects the agent's output and the fix could be generalized
- Agent consistently produces incorrect format/style
- User expresses preferences that should become default behavior
- Agent misunderstands context that a prompt clarification could fix

### How to Refine

```bash
# Append to existing prompt (preferred for small adjustments)
ayo agents refine @agent-name \
  --append "Additional instruction text here." \
  --note "Reason for this change"

# Replace entire prompt (for major rewrites)
ayo agents refine @agent-name \
  --replace "Complete new system prompt..." \
  --note "Reason for complete rewrite"
```

**Always include a note** explaining why you're making the refinement.

### Refinement Best Practices

1. **Start with append**: Add small clarifications before rewriting
2. **Be specific**: "When discussing X, always do Y" is better than vague instructions
3. **Track patterns**: Wait for 2-3 similar corrections before refining
4. **Document reasons**: Notes help understand refinement history later
5. **Test after refining**: Verify the refinement improved behavior

## Invocation Context

Invocation context provides task-specific instructions to an agent without creating a new agent. This prevents agent sprawl while allowing specialization.

### What is Invocation Context?

Context is additional guidance prepended to the prompt for a specific invocation:

```
@researcher: "Focus on quantum computing papers from 2025-2026.
  Prioritize peer-reviewed sources.
  
  Find recent advances in error correction."
```

The context ("Focus on quantum computing...") shapes this specific task without permanently changing @researcher.

### When to Use Invocation Context

**Prefer invocation context when:**
- One-off specialization ("For this task, focus on X")
- Simple focus adjustment ("Only look at Python files")
- First few times trying a pattern
- Testing whether a specialization helps

**Create a new agent when:**
- Pattern used successfully 3+ times with same context
- Context is too complex to repeat each time
- User explicitly requests a dedicated agent
- Persistent refinement needed

### Using Context in Flows

Flow steps support a `context` field for invocation context:

```yaml
steps:
  - id: research
    type: agent
    agent: "@researcher"
    context: |
      Focus on quantum computing papers.
      Prioritize 2025-2026 publications.
      Prefer peer-reviewed sources.
    prompt: "Find recent advances in {{ params.topic }}"
```

The context is prepended to the prompt, giving the agent task-specific guidance.

### Context in Direct Delegation

When delegating directly, include context as part of your message:

```bash
# Include context in the prompt itself
ayo @summarizer "For this document, focus on technical details and use bullet points.

Here is the content to summarize: ..."
```

### Best Practices

1. **Keep context concise**: 1-3 lines is usually enough
2. **Be directive**: "Do X" is clearer than "Consider doing X"
3. **Test context effectiveness**: If context helps, consider creating an agent after 3+ successful uses
4. **Don't duplicate system prompt**: Context supplements, not replaces, the agent's base capabilities
