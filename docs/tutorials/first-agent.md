# Tutorial: Create Your First Agent

Build a specialized code review agent from scratch. By the end of this tutorial, you'll have a working agent that reviews code with a focus on security.

**Time**: ~20 minutes  
**Prerequisites**: [Getting Started](../getting-started.md) complete

## What You'll Build

A `@reviewer` agent that:
- Reviews code for security vulnerabilities
- Checks for common bugs and anti-patterns
- Provides actionable feedback with line numbers

## Step 1: Create the Agent

```bash
ayo agent create @reviewer
```

This creates the agent directory at `~/.config/ayo/agents/@reviewer/`:

```
@reviewer/
├── config.json
└── system.md
```

## Step 2: Define the System Prompt

Edit `~/.config/ayo/agents/@reviewer/system.md`:

```markdown
# Code Review Agent

You are a senior code reviewer specializing in security and quality.

## Review Focus

When reviewing code, prioritize:

1. **Security Vulnerabilities**
   - SQL injection
   - XSS vulnerabilities
   - Authentication bypasses
   - Sensitive data exposure
   - Input validation issues

2. **Common Bugs**
   - Null pointer dereferences
   - Off-by-one errors
   - Resource leaks
   - Race conditions
   - Error handling gaps

3. **Code Quality**
   - Unclear variable names
   - Complex nested logic
   - Duplicated code
   - Missing documentation for public APIs

## Response Format

Structure your reviews as:

```
## Summary
[1-2 sentence overview]

## Critical Issues
[Security or correctness problems]

## Suggestions
[Non-blocking improvements]

## Positive Notes
[What's done well]
```

## Guidelines

- Be specific: cite file names and line numbers
- Be constructive: explain why something is an issue
- Be concise: focus on the most important points
- Prioritize security over style nitpicks
```

## Step 3: Configure the Agent

Edit `~/.config/ayo/agents/@reviewer/config.json`:

```json
{
  "model": "claude-sonnet-4-20250514",
  "description": "Security-focused code reviewer",
  "allowed_tools": [
    "bash",
    "view",
    "glob",
    "grep"
  ],
  "memory": {
    "enabled": true,
    "scope": "agent"
  }
}
```

### Configuration Explained

| Field | Purpose |
|-------|---------|
| `model` | LLM model for this agent |
| `description` | Shows in `ayo agent list` |
| `allowed_tools` | Tools the agent can use (no edit = read-only) |
| `memory` | Enable agent-specific memories |

## Step 4: Test the Agent

Verify the agent is available:

```bash
ayo agent list
```

You should see `@reviewer` in the list.

Show agent details:

```bash
ayo agent show @reviewer
```

## Step 5: Run Your First Review

Review a single file:

```bash
ayo @reviewer "Review this file for security issues" -a main.go
```

Review a directory:

```bash
ayo @reviewer "Review the auth package" -a ./pkg/auth/
```

Review recent changes:

```bash
git diff HEAD~3 | ayo @reviewer "Review these changes"
```

## Step 6: Add Skills (Optional)

Skills are markdown files that give agents additional knowledge. Create a OWASP reference skill:

Create `~/.config/ayo/agents/@reviewer/skills/owasp.md`:

```markdown
# OWASP Top 10 Reference

## A01:2021 - Broken Access Control
- Missing function-level access control
- IDOR vulnerabilities
- Privilege escalation

## A02:2021 - Cryptographic Failures
- Weak encryption algorithms
- Hardcoded secrets
- Missing TLS

## A03:2021 - Injection
- SQL injection (parameterize queries!)
- Command injection
- LDAP injection

## A07:2021 - Cross-Site Scripting (XSS)
- Reflected XSS in user input
- Stored XSS in database content
- DOM-based XSS

When reviewing code, check for these vulnerability categories.
```

Reference the skill in `config.json`:

```json
{
  "skills": ["./skills/owasp.md"]
}
```

## Complete Example

Your final `@reviewer` directory:

```
@reviewer/
├── config.json
├── system.md
└── skills/
    └── owasp.md
```

**config.json**:
```json
{
  "model": "claude-sonnet-4-20250514",
  "description": "Security-focused code reviewer",
  "allowed_tools": ["bash", "view", "glob", "grep"],
  "memory": {
    "enabled": true,
    "scope": "agent"
  },
  "skills": ["./skills/owasp.md"]
}
```

## Troubleshooting

### Agent not found after creation

```bash
# Verify the directory exists
ls ~/.config/ayo/agents/@reviewer/

# Check for syntax errors in config.json
cat ~/.config/ayo/agents/@reviewer/config.json | jq .
```

### Agent using wrong model

The model resolution order is:
1. Agent `config.json`
2. Global config `~/.config/ayo/config.json`
3. Environment variable `AYO_MODEL`
4. Default (claude-sonnet-4-20250514)

### Skills not loading

Ensure the path in `skills` is relative to the agent directory:

```json
{
  "skills": ["./skills/owasp.md"]
}
```

## Next Steps

- [Multi-Agent Squads](squads.md) - Combine reviewers with other agents
- [Triggers](triggers.md) - Auto-run on git commits
- [Memory](memory.md) - Remember project-specific conventions

---

*You've created your first custom agent! Continue to [Multi-Agent Squads](squads.md).*
