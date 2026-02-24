# Prompts Reference

Complete reference for externalized prompts and guardrails.

## Overview

Ayo uses externalized prompts for:
- **Guardrails**: Safety constraints
- **Sandwich prompts**: Before/after agent prompts
- **System base**: Core ayo behavior

## Directory Structure

```
~/.local/share/ayo/prompts/
├── defaults/                 # Built-in prompts
│   ├── system/
│   │   └── base.md           # Core system prompt
│   ├── guardrails/
│   │   └── default.md        # Default guardrails
│   └── sandwich/
│       ├── prefix.md         # Before agent prompt
│       └── suffix.md         # After agent prompt
└── overrides/                # User overrides (not created by default)

~/.config/ayo/prompts/        # User customizations
├── guardrails/
│   ├── default.md            # Override default guardrails
│   └── @reviewer.md          # Agent-specific guardrails
└── sandwich/
    ├── prefix.md             # Custom prefix
    └── suffix.md             # Custom suffix
```

## Prompt Injection Order

When an agent runs, prompts are assembled in this order:

1. **System base prompt** (`system/base.md`)
2. **Guardrails prefix** (`guardrails/default.md`)
3. **Agent system.md** (agent's own prompt)
4. **Squad constitution** (SQUAD.md, if in squad)
5. **Skills** (SKILL.md files)
6. **Guardrails suffix** (`sandwich/suffix.md`)

### Visualization

```
┌─────────────────────────────────────┐
│ System Base Prompt                  │ ← Core ayo behavior
├─────────────────────────────────────┤
│ Guardrails Prefix                   │ ← Safety constraints
├─────────────────────────────────────┤
│ Agent system.md                     │ ← Agent personality
├─────────────────────────────────────┤
│ Squad Constitution (if applicable)  │ ← Team context
├─────────────────────────────────────┤
│ Skills                              │ ← Knowledge
├─────────────────────────────────────┤
│ Guardrails Suffix                   │ ← Final constraints
└─────────────────────────────────────┘
```

## Guardrails

### Default Guardrails

`defaults/guardrails/default.md`:

```markdown
# Safety Guidelines

You are an AI assistant operating in a sandboxed environment.

## Core Principles

1. **Transparency**: Always explain what you're doing
2. **Safety**: Never execute harmful commands
3. **Privacy**: Don't access or transmit sensitive data
4. **Permission**: Request approval for significant changes

## Prohibited Actions

- Accessing credentials, API keys, or secrets
- Making network requests to unknown endpoints
- Modifying system files
- Installing untrusted software
- Accessing .git directories directly
- Reading .env or similar secret files

## Required Behaviors

- Explain commands before running them
- Show diffs before file modifications
- Request explicit confirmation for destructive actions
- Report any errors or unexpected behavior
```

### Agent-Specific Guardrails

Create `~/.config/ayo/prompts/guardrails/@agent.md`:

```markdown
# @reviewer Guardrails

In addition to standard guardrails:

## Review-Specific Rules

- Focus on code quality, not personal preferences
- Cite line numbers when pointing out issues
- Prioritize security issues over style
- Be constructive, not critical
```

### Disabling Guardrails

For `unrestricted` agents:

```json
{
  "trust_level": "unrestricted",
  "guardrails": false
}
```

**Warning**: Only disable for trusted, well-tested agents.

## Sandwich Prompts

### Prefix Prompt

`sandwich/prefix.md` - Injected after guardrails, before agent content:

```markdown
## Context

You are working in an isolated sandbox environment.
Your home directory is /home/{agent}/.
The host user's files are mounted read-only at /mnt/{username}/.

## Tools Available

{tools_list}

## Current Working Directory

{cwd}
```

### Suffix Prompt

`sandwich/suffix.md` - Injected at the end:

```markdown
## Reminders

- Always verify your work before reporting completion
- Use file_request for host file modifications
- Store important learnings as memories
```

## System Base Prompt

`system/base.md` - Core ayo behavior:

```markdown
# Ayo Agent

You are an ayo agent - an AI assistant that operates within
a sandboxed environment on the user's machine.

## Environment

You execute commands in a Linux container. Your actions are
isolated from the host system unless you explicitly use the
file_request tool.

## Communication Style

- Be concise and direct
- Show your work (commands, reasoning)
- Ask clarifying questions when needed
- Report progress on long tasks

## Tools

You have access to various tools for interacting with the
environment. Use them appropriately for the task at hand.
```

## Customization

### Override Resolution

1. User overrides (`~/.config/ayo/prompts/`)
2. Plugin prompts
3. System defaults (`~/.local/share/ayo/prompts/defaults/`)

### Creating Overrides

```bash
# Create override directory
mkdir -p ~/.config/ayo/prompts/guardrails

# Copy default and modify
cp ~/.local/share/ayo/prompts/defaults/guardrails/default.md \
   ~/.config/ayo/prompts/guardrails/default.md
   
# Edit to customize
vim ~/.config/ayo/prompts/guardrails/default.md
```

### Agent-Specific Prompts

Name the file with the agent handle:

```
~/.config/ayo/prompts/guardrails/@reviewer.md
~/.config/ayo/prompts/guardrails/@assistant.md
```

## Template Variables

Prompts can include template variables:

| Variable | Description |
|----------|-------------|
| `{agent}` | Current agent handle |
| `{user}` | Host username |
| `{cwd}` | Current working directory |
| `{tools_list}` | Available tools |
| `{date}` | Current date |
| `{time}` | Current time |

### Usage

```markdown
## Current Context

You are @{agent} working for {user}.
Today is {date}.

## Available Tools

{tools_list}
```

## Plugin Prompt Overrides

Plugins can provide prompts:

```
my-plugin/
├── manifest.json
└── prompts/
    ├── guardrails/
    │   └── @my-agent.md
    └── sandwich/
        └── my-suffix.md
```

In `manifest.json`:

```json
{
  "components": {
    "prompts": {
      "guardrails/@my-agent": "./prompts/guardrails/@my-agent.md",
      "sandwich/my-suffix": "./prompts/sandwich/my-suffix.md"
    }
  }
}
```

## Best Practices

### Guardrails

- Keep guardrails concise and clear
- Focus on boundaries, not instructions
- Test changes with various prompts
- Monitor for bypass attempts

### Agent Prompts (system.md)

- Define clear responsibilities
- Include output format examples
- Set behavioral boundaries
- Consider squad context

### Skills

- Keep skills focused (one topic per file)
- Use clear headings
- Include examples
- Keep under 1000 tokens if possible

## Troubleshooting

### Prompts not loading

```bash
# Check prompt exists
ls ~/.config/ayo/prompts/guardrails/

# Check file permissions
stat ~/.config/ayo/prompts/guardrails/default.md

# Verify markdown syntax
cat ~/.config/ayo/prompts/guardrails/default.md
```

### Agent ignoring guardrails

1. Check `guardrails: false` not set in config
2. Check trust level isn't `unrestricted`
3. Verify prompt file exists and is readable
4. Check for syntax errors in prompt

### Template variables not expanding

Ensure variables use correct syntax: `{variable}` not `{{variable}}` or `$variable`.
