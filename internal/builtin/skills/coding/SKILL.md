---
name: coding
description: |
  Skill for all source code creation and modification tasks.
  Delegate coding work to the configured coding agent.
  Check <delegate_context> for the current coding delegate.
metadata:
  author: ayo
  version: "5.0"
---

# Coding Skill

**All source code creation and modification tasks must be delegated to the configured coding agent.**

Check your `<delegate_context>` system message to see which agent handles coding tasks. If no coding delegate is configured, inform the user they need to install a coding plugin.

## When to Delegate

**ALWAYS delegate for ANY task involving source code:**

- Creating new projects or applications (SPAs, APIs, CLIs, libraries, etc.)
- Scaffolding new codebases or project structures
- Writing new source code files from scratch
- Implementing new features (single or multi-file)
- Modifying or editing existing code
- Refactoring or restructuring code
- Debugging and fixing issues
- Adding or modifying tests
- Code generation from specifications
- Migrating codebases (framework upgrades, API changes)
- Security fixes and vulnerability remediation
- Performance optimization requiring code changes
- Documentation generation from code

**DO NOT delegate for:**

- Queries or questions about code (answer directly)
- Running existing commands (use bash)
- Git operations (use bash)
- Installing dependencies without code changes (use bash)
- File management without code content (use bash)
- Information gathering or research

## Decision Tree

```
Does the task involve creating or modifying source code?
├── YES → Check <delegate_context> for coding agent, then use agent_call
└── NO → Use bash tool or handle directly
```

## How to Delegate

Use the `agent_call` tool with the agent from your `<delegate_context>`:

```json
{
  "agent": "<coding_agent_from_context>",
  "prompt": "Detailed description of the coding task"
}
```

### Model Passthrough

Pass your current model to the sub-agent for consistency:

```json
{
  "agent": "<coding_agent_from_context>",
  "prompt": "Add comprehensive error handling to the database layer",
  "model": "claude-sonnet-4"
}
```

### Prompt Formatting Guidelines

**Structure your prompts with:**

1. **Clear objective**: What needs to be accomplished
2. **Scope**: Which files or directories are involved
3. **Constraints**: What should NOT be changed
4. **Success criteria**: How to verify the task is complete

**Good prompt example:**
```json
{
  "agent": "<coding_agent>",
  "prompt": "Create a basic single page application in the my-app directory. Use vanilla HTML, CSS, and JavaScript. Include an index.html with a simple navigation header, a main content area, and a footer.",
  "model": "claude-sonnet-4"
}
```

**Bad prompt example:**
```json
{
  "agent": "<coding_agent>",
  "prompt": "Fix the database errors"
}
```

## If No Coding Agent is Configured

If `<delegate_context>` shows no coding delegate, inform the user:

> No coding agent is configured. Install a coding plugin:
> ```
> ayo plugins install https://github.com/alexcabrera/ayo-plugins-crush
> ```

## Best Practices

1. **Be specific**: Vague prompts lead to vague results
2. **Set clear boundaries**: Specify what NOT to change
3. **Request verification**: Ask to run tests after changes
4. **Pass your model**: Use the `model` parameter for consistency
5. **Iterate incrementally**: Large tasks should be broken into phases
6. **Provide context**: Include relevant background for complex tasks
