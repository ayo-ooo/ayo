---
name: coding
description: |
  Skill for all source code creation and modification tasks.
  Delegate coding work to the configured coding agent (default: @crush from ayo-plugins-crush).
  This skill requires a coding plugin to be installed.
metadata:
  author: ayo
  version: "4.0"
---

# Coding Skill

**This skill requires a coding plugin to be installed.**

By default, ayo delegates coding tasks to `@crush` (provided by `ayo-plugins-crush`). 
If no coding plugin is installed, coding tasks cannot be delegated.

## Installation

Install a coding plugin:

```bash
ayo plugins install alexcabrera/crush
```

This installs the `@crush` agent which uses Crush for complex source code tasks.

## Configuration

The coding agent can be configured at three levels (highest priority first):

1. **Directory level** (`.ayo.json` in project directory):
```json
{
  "delegates": {
    "coding": "@crush"
  }
}
```

2. **Agent level** (in agent's `config.json`):
```json
{
  "delegates": {
    "coding": "@crush"
  }
}
```

3. **Global level** (`~/.config/ayo/ayo.json`):
```json
{
  "delegates": {
    "coding": "@crush"
  }
}
```

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
├── YES → Delegate to coding agent (@crush)
└── NO → Use bash tool or handle directly
```

## How to Delegate

Use the `agent_call` tool to delegate coding tasks:

```json
{
  "agent": "@crush",
  "prompt": "Create a basic React single page application in the test-app directory with a home page and about page"
}
```

### Model Passthrough

Pass your current model to the sub-agent for consistency:

```json
{
  "agent": "@crush",
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

**Good prompt examples:**

Creating a new project:
```json
{
  "agent": "@crush",
  "prompt": "Create a basic single page application in the my-app directory. Use vanilla HTML, CSS, and JavaScript. Include an index.html with a simple navigation header, a main content area, and a footer.",
  "model": "claude-sonnet-4"
}
```

Modifying existing code:
```json
{
  "agent": "@crush",
  "prompt": "Add comprehensive error handling to the database connection logic in internal/db/. Wrap all database calls with proper error context. Do NOT modify the connection pool configuration.",
  "model": "claude-sonnet-4"
}
```

**Bad prompt example:**
```json
{
  "agent": "@crush",
  "prompt": "Fix the database errors"
}
```

### Scope Setting

| Scope | How to Specify |
|-------|----------------|
| New project | `"in the my-app directory"` or `"in a new directory called my-app"` |
| Single file | `"in api/handlers/user.go"` |
| Directory | `"in the internal/auth/ directory"` |
| Multiple files | `"in user.go, auth.go, and session.go"` |
| Project-wide | `"across the entire codebase"` (use sparingly) |

### Constraint Specification

Always specify what should NOT be modified:

- `"Do not modify any test files"`
- `"Preserve the existing public API"`
- `"Keep backwards compatibility with v1 endpoints"`
- `"Do not change the database schema"`

## Understanding Results

When the coding agent completes, you receive:

1. **Summary**: What was accomplished
2. **Files modified**: List of changed files
3. **Test results**: Whether tests pass (if applicable)
4. **Any issues encountered**: Warnings or errors

### Success Indicators

The delegation succeeded if:
- No error messages in output
- Agent confirms completion with specific details
- Modified files match the expected scope

### When to Iterate

Retry with a refined prompt if:
- The scope was misunderstood
- Changes were incomplete
- Tests are failing due to missed edge cases
- Output doesn't match expectations

## Best Practices

1. **Be specific**: Vague prompts lead to vague results
2. **Set clear boundaries**: Specify what NOT to change
3. **Request verification**: Ask to run tests after changes
4. **Pass your model**: Use the `model` parameter for consistency
5. **Iterate incrementally**: Large tasks should be broken into phases
6. **Provide context**: Include relevant background for complex tasks

## Example Delegations

### Creating a New Project
```json
{
  "agent": "@crush",
  "prompt": "Create a new Go CLI application in the my-cli directory. Use cobra for command handling. Include a root command with version flag, and subcommands for 'init' and 'run'. Add a Makefile with build and test targets.",
  "model": "claude-sonnet-4"
}
```

### Adding a New Feature
```json
{
  "agent": "@crush",
  "prompt": "Add a rate limiting middleware to the API server: Create internal/middleware/ratelimit.go using a token bucket algorithm. Add configuration options in config/config.go. Apply to all API endpoints.",
  "model": "claude-sonnet-4"
}
```

### Debugging an Issue
```json
{
  "agent": "@crush",
  "prompt": "Debug and fix the memory leak in the WebSocket handler. The issue is in internal/ws/handler.go - connections are not being properly cleaned up on disconnect.",
  "model": "claude-sonnet-4"
}
```

### Refactoring Code
```json
{
  "agent": "@crush",
  "prompt": "Refactor the user service to use the repository pattern: Extract database operations from internal/user/service.go into a new Repository interface and implementation.",
  "model": "claude-sonnet-4"
}
```
