---
id: ayo-rx03
status: closed
deps: []
links: []
created: 2026-02-24T03:00:00Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ayo-rx01
tags: [remediation, documentation]
---
# Task: Create docs/concepts.md

## Summary

Create the foundational concepts document that explains ayo's mental model. This is the most important document for understanding how everything fits together.

## Requirements

Target: ~400 lines, provides conceptual foundation

## Structure

```markdown
# Core Concepts

## What is Ayo?

Ayo is a CLI framework for running AI agents in isolated sandbox environments. Key principles:
- Agents run in containers, not on your host
- Explicit permission for host modifications
- Multi-agent coordination through Squads
- Persistent memory across sessions

## Agents

### What is an Agent?
An agent is an AI assistant configured for a specific task...

### Agent Structure
@agent-name/
├── ayo.json       # Configuration
├── system.md      # System prompt
├── tools/         # Custom tools
└── skills/        # Knowledge files

### The Default @ayo Agent
Ships with ayo, general-purpose assistant...

## Sandboxes

### What is a Sandbox?
Isolated container where agents execute...

### Why Isolation?
Security: agents can't access host files directly...

### Sandbox Types
- @ayo sandbox: Shared, persistent
- Squad sandboxes: Per-squad, isolated

### File System Layout
~/.local/share/ayo/
├── sandboxes/     # Container data
├── output/        # Agent outputs
└── ...

## Squads

### What is a Squad?
A team of agents working together in a shared sandbox...

### SQUAD.md
The orchestrator's system prompt that defines roles...

### Coordination
Agents coordinate through tickets...

## Memory

### What is Memory?
Persistent knowledge that agents remember...

### Memory Categories
- preference: User preferences
- fact: Factual information
- correction: Behavior corrections
- pattern: Observed patterns

### Memory Scopes
- global: All agents
- agent: Specific agent
- path: Directory-specific
- squad: Squad members only

## Triggers

### What is a Trigger?
An event that starts an agent automatically...

### Trigger Types
- cron: Schedule-based
- interval: Time-based
- file_watch: File changes
- event: Custom events

## Tools

### What are Tools?
Functions agents can call...

### Built-in Tools
bash, view, edit, memory_store, file_request...

### Permissions
Tools require explicit enablement...

## Plugins

### What is a Plugin?
Extension that adds agents, tools, triggers...

### Installing Plugins
ayo plugin install <name>

## Permissions Model

### file_request Flow
1. Agent wants to write to host
2. Calls file_request tool
3. User sees approval prompt
4. User approves/denies
5. File written (or rejected)

### --no-jodas Mode
Auto-approve all file requests...

### Trust Levels
- Sandbox: Full access
- Host read: Read-only mount
- Host write: Requires approval
```

## Acceptance Criteria

- [ ] File exists at `docs/concepts.md`
- [ ] All concepts accurate to implementation
- [ ] Provides clear mental model
- [ ] Links to detailed guides
- [ ] No orphan concepts
