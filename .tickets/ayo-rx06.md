---
id: ayo-rx06
status: closed
deps: [ayo-rx03]
links: []
created: 2026-02-24T03:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-rx01
tags: [remediation, documentation]
---
# Task: Create docs/reference/ (5 reference documents)

## Summary

Create 5 technical reference documents in `docs/reference/` for developers and power users. These should be authoritative, complete, and precise.

## Files to Create

### 1. reference/cli.md (~500 lines)

Complete CLI command reference:

```markdown
# CLI Reference

## Global Flags
--json          Output as JSON
--quiet         Minimal output
--config PATH   Config file path
--no-jodas      Auto-approve all file requests

## Commands

### ayo
Interactive mode or single prompt.
ayo [prompt]

### ayo agent
Agent management.
ayo agent list
ayo agent show <name>
ayo agent new <name>

### ayo squad
Squad management.
ayo squad list
ayo squad create <name>
ayo squad destroy <name>
ayo squad run <name> [prompt]
ayo squad shell <name>

### ayo memory
Memory management.
ayo memory list
ayo memory add <content>
ayo memory search <query>
ayo memory show <id>
ayo memory delete <id>
ayo memory export [file]
ayo memory import <file>

### ayo trigger
Trigger management.
ayo trigger list
ayo trigger add [flags]
ayo trigger remove <name>
ayo trigger fire <name>
ayo trigger history

### ayo plugin
Plugin management.
ayo plugin list
ayo plugin install <name|path>
ayo plugin remove <name>
ayo plugin search <query>

### ayo daemon
Daemon control.
ayo daemon start
ayo daemon stop
ayo daemon status
ayo daemon restart

### ayo doctor
System health check.

### ayo backup
Backup operations.
ayo backup create
ayo backup restore <file>

### ayo audit
Audit log viewer.
ayo audit list
ayo audit show <id>

## Exit Codes
0  Success
1  General error
2  Invalid arguments
3  Permission denied

## Environment Variables
AYO_HOME        Base directory
AYO_CONFIG      Config file path
AYO_PROVIDER    Default LLM provider
AYO_MODEL       Default model
```

### 2. reference/ayo-json.md (~400 lines)

Complete ayo.json schema reference:

```markdown
# ayo.json Reference

## Agent Configuration

{
  "provider": "anthropic",
  "model": "claude-sonnet-4-20250514",
  "tools": ["bash", "view", "edit", ...],
  "skills": ["./skills/SKILL.md"],
  "memory": {
    "enabled": true,
    "scope": "agent"
  },
  "permissions": {
    "file_request": true,
    "auto_approve": false
  },
  "planners": {
    "near_term": "ayo-todos",
    "long_term": "ayo-tickets"
  },
  "sandbox": {
    "provider": "applecontainer",
    "resources": {}
  }
}

## Squad Configuration

{
  "agents": {
    "@frontend": { ... },
    "@backend": { ... }
  },
  "planners": { ... },
  "triggers": { ... }
}

[Full field documentation for each]
```

### 3. reference/prompts.md (~250 lines)

Externalized prompts reference:

```markdown
# Prompts Reference

## Directory Structure
~/.local/share/ayo/prompts/
├── defaults/
│   ├── system/
│   │   └── base.md
│   ├── guardrails/
│   │   └── default.md
│   └── sandwich/
│       ├── prefix.md
│       └── suffix.md
└── overrides/
    └── ...

## Prompt Injection Order
1. System base prompt
2. Guardrails prefix
3. Agent system.md
4. Squad constitution (if in squad)
5. Skills
6. Guardrails suffix

## Creating Custom Prompts
[Override mechanisms]

## Plugin Prompt Overrides
[How plugins can provide prompts]
```

### 4. reference/rpc.md (~300 lines)

Daemon RPC reference:

```markdown
# Daemon RPC Reference

## Connection
Unix socket: ~/.local/share/ayo/daemon.sock
Protocol: JSON-RPC 2.0

## Methods

### agent.run
Execute agent with prompt.

### squad.dispatch
Dispatch to squad agent.

### trigger.add
Register new trigger.

### memory.store
Store memory.

### memory.search
Search memories.

[Full method documentation with request/response schemas]

## Error Codes
-32600  Invalid request
-32601  Method not found
-32602  Invalid params
-32603  Internal error
-32700  Parse error

## Example Calls
[curl/netcat examples]
```

### 5. reference/plugins.md (~350 lines)

Plugin manifest reference:

```markdown
# Plugin Reference

## manifest.json Schema

{
  "name": "plugin-name",
  "version": "1.0.0",
  "description": "Plugin description",
  "author": "Author Name",
  "license": "MIT",
  "components": {
    "agents": {
      "@agent-name": {
        "path": "./agents/agent-name"
      }
    },
    "tools": {
      "tool-name": {
        "path": "./tools/tool-name"
      }
    },
    "skills": { ... },
    "squads": { ... },
    "triggers": { ... },
    "planners": { ... }
  },
  "dependencies": {
    "ayo": ">=1.0.0"
  }
}

## Component Types
[Detailed documentation for each]

## Resolution Order
1. User components (~/.config/ayo/)
2. Installed plugins (~/.local/share/ayo/plugins/)
3. Built-in components

## Publishing Plugins
[How to publish to registry]
```

## Acceptance Criteria

- [ ] All 5 reference files exist in `docs/reference/`
- [ ] Technically accurate to implementation
- [ ] Complete coverage of all options
- [ ] Code-level precision
- [ ] Examples verified working
- [ ] Covers edge cases
