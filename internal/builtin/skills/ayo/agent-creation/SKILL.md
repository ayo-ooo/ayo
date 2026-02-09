# Agent Creation Skill

You have the ability to create specialized agents to delegate tasks that require specific expertise or capabilities.

## When to Create Agents

Create specialized agents when:
1. You've used similar context for the same base task 3+ times
2. The user explicitly asks for a specialized agent
3. A task requires a unique combination of skills no existing agent has
4. The task would benefit from a persistent, reusable personality

## Before Creating

Always check if a similar agent already exists:
```bash
ayo agents list
```

## How to Create

Use the `ayo agents create` command:

```bash
ayo agents create @<handle> \
  --model <model-id> \
  --description "Brief description" \
  --system "System prompt text" \
  --skills skill1,skill2 \
  --tools bash,todo \
  --created-by "@ayo" \
  --creation-reason "Why this agent was created"
```

### Parameters

- `@<handle>`: A short, descriptive name without spaces or dots (e.g., `@science-researcher`, `@code-reviewer`)
- `--model`: The model to use (e.g., `gpt-5.2`, `claude-3.5-sonnet`)
- `--description`: A brief description of what this agent does
- `--system`: The system prompt defining the agent's personality and capabilities
- `--skills`: Comma-separated list of skills to include
- `--tools`: Comma-separated list of tools to allow

### Examples

```bash
# Create a code review specialist
ayo agents create @code-reviewer \
  --model gpt-5.2 \
  --description "Reviews code for bugs, security issues, and best practices" \
  --system "You are an expert code reviewer. Focus on:
- Security vulnerabilities
- Performance issues
- Code clarity and maintainability
- Best practices for the language/framework
Be thorough but constructive in your feedback." \
  --skills coding \
  --tools bash \
  --created-by "@ayo" \
  --creation-reason "User frequently requested code reviews"

# Create a research assistant
ayo agents create @researcher \
  --model claude-3.5-sonnet \
  --description "Researches topics and summarizes findings" \
  --system "You are a research specialist. When given a topic:
1. Break it into key questions
2. Search for authoritative sources
3. Synthesize information
4. Present clear, cited summaries" \
  --skills memory \
  --created-by "@ayo" \
  --creation-reason "Pattern: 5+ research requests in similar domain"
```

## Naming Guidelines

- Keep handles short and descriptive (1-3 words)
- Use hyphens to separate words
- No dots, spaces, or special characters
- Avoid generic names like `@helper` or `@assistant`
- Be specific: `@python-debugger` is better than `@debugger`

## Agent Lifecycle

Agents you create are tracked in the database:
- Usage statistics are recorded (invocations, success/failure rates)
- Confidence scores increase with successful uses
- Prompts can be refined based on feedback
- Underperforming agents can be archived

You can review your created agents:
```bash
ayo agents list --created-by @ayo
```

## Invoking Created Agents

Once created, invoke agents with:
```bash
ayo @<handle> "Your prompt here"
```

Or from within a flow:
```yaml
steps:
  - id: review
    type: agent
    agent: "@code-reviewer"
    prompt: "Review the changes in this PR"
```
