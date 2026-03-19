# Skills

Package reusable capabilities as skill modules.

## Overview

Skills are modular components that extend an agent's capabilities. Each skill is defined in a `SKILL.md` file within a subdirectory of the `skills/` directory.

## Directory Structure

```
my-agent/
├── skills/
│   ├── analyze/
│   │   └── SKILL.md
│   ├── transform/
│   │   └── SKILL.md
│   └── validate/
│       └── SKILL.md
```

## Skill Definition

Each `SKILL.md` file describes a skill:

```markdown
# Skill Name

Brief description of what the skill does.

## Purpose

Detailed explanation of the skill's purpose and use cases.

## Usage

When and how to invoke this skill.

## Parameters

- param1: Description
- param2: Description

## Output

Description of expected output.
```

## How Skills Work

1. Skills are embedded in the generated binary
2. The system prompt includes skill descriptions
3. The LLM learns about available skills from the prompt
4. The LLM decides when to invoke skills based on context

## Example: Task Runner

From the task-runner example:

**skills/plan/SKILL.md**:
```markdown
# Plan Skill

Break down complex tasks into executable steps.

## Purpose

Analyze a task request and create a structured execution plan with ordered steps.

## Usage

Invoke this skill when:
- Starting a new task
- Breaking down complex requests
- Creating execution strategies

## Output

Returns a structured plan with:
- Ordered steps
- Dependencies between steps
- Estimated effort per step
- Success criteria
```

**skills/execute/SKILL.md**:
```markdown
# Execute Skill

Execute a single step from a plan.

## Purpose

Perform the actual work defined in a plan step.

## Usage

Invoke this skill when:
- Running a planned step
- Implementing a solution
- Performing operations

## Parameters

- step: The step to execute
- context: Relevant context from previous steps

## Output

Returns execution results and any artifacts produced.
```

**skills/review/SKILL.md**:
```markdown
# Review Skill

Evaluate execution results and validate quality.

## Purpose

Review completed work against requirements and quality standards.

## Usage

Invoke this skill when:
- A step completes
- Checking work quality
- Validating outputs

## Output

Returns:
- Pass/fail status
- Issues found
- Improvement suggestions
```

## System Prompt Integration

When skills are present, include them in your system prompt:

```markdown
# Task Runner Agent

You are a task execution agent.

## Available Skills

You have access to these skills:

- **plan**: Break down tasks into steps
- **execute**: Execute individual steps
- **review**: Validate completed work

## Workflow

1. Use the `plan` skill to analyze requests
2. Use the `execute` skill for each step
3. Use the `review` skill to validate results
```

## Skill Discovery

Ayo automatically discovers skills in the `skills/` directory:

- Each subdirectory is a skill
- Each skill must have a `SKILL.md` file
- Skill name is the directory name

## Best Practices

1. **Single responsibility**: Each skill does one thing well
2. **Clear documentation**: Describe when and how to use
3. **Consistent structure**: Use the same format for all skills
4. **Meaningful names**: Use verb-based names (analyze, transform, validate)

## Limitations

- Skills are informational only - they guide the LLM, not executable code
- No automatic skill invocation - the LLM decides when to use them
- Skills share the same context as the main agent

## Next Steps

- [Hooks](hooks.md) - React to lifecycle events
- [Examples](../examples/task-runner.md) - See the task-runner example
