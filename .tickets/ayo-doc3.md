---
id: ayo-doc3
status: closed
deps: [ayo-doc1]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-docs
tags: [documentation, phase9]
---
# Task: Write Core Concepts Guide

## Summary

Write concepts.md that explains ayo's mental model. This is the foundational document that helps users understand how everything fits together.

## Structure

```markdown
# Core Concepts

## Overview
Brief explanation of what ayo is and why it exists.

## Agents
- What is an agent?
- Agent directory structure
- system.md and ayo.json
- Built-in @ayo agent
- Custom agents

## Sandboxes
- What is a sandbox?
- Why isolation matters
- @ayo sandbox (shared)
- Squad sandboxes (isolated)
- File system layout
- ayod (in-sandbox daemon)

## Squads
- What is a squad?
- SQUAD.md (orchestrator's system prompt)
- ayo.json (configuration)
- Squad lead and coordination
- Tickets and task handoff

## Memory
- What is memory?
- Memory categories (preference, fact, correction, pattern)
- Memory scopes (global, agent, path)
- Zettelkasten notes
- How memory is used

## Triggers
- What is a trigger?
- Cron triggers
- File watch triggers
- Event triggers (plugins)
- Ambient agents

## Tools
- What are tools?
- Built-in tools
- External tools
- Tool permissions

## Plugins
- What is a plugin?
- Plugin components
- Installing plugins
- Creating plugins

## Permissions
- file_request flow
- --no-jodas mode
- Trust levels
- Guardrails
```

## Requirements

- Clear, jargon-free explanations
- Diagrams where helpful (ASCII art)
- Cross-references to detailed docs
- Progressive disclosure (simple → complex)

## Success Criteria

- [ ] User understands ayo mental model after reading
- [ ] All concepts are accurate to implementation
- [ ] Clear path from concepts to tutorials
- [ ] No orphan concepts (all referenced elsewhere)

---

*Created: 2026-02-23*
