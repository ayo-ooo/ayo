---
id: ayo-doc5
status: open
deps: [ayo-doc3]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-docs
tags: [documentation, phase9]
---
# Task: Write Configuration Guides

## Summary

Write 6 configuration guides in `docs/guides/` that explain how to configure each major component.

## Guides

### 1. agents.md
Complete agent configuration reference:
- Directory structure
- system.md format
- ayo.json schema
- Tools configuration
- Memory settings
- Sandbox options
- Permissions

### 2. squads.md
Complete squad configuration reference:
- Directory structure
- SQUAD.md format and semantics
- ayo.json squad schema
- Agent roster
- Planners configuration
- Sandbox options
- Memory sharing

### 3. triggers.md
Complete trigger configuration reference:
- Trigger types (cron, file, event)
- gocron v2 features
- Schedule syntax
- Trigger persistence
- Output handling
- Error handling

### 4. tools.md
Built-in and external tools reference:
- Built-in tools list
- Tool configuration
- External tool format
- Tool permissions
- Creating custom tools

### 5. sandbox.md
Sandbox architecture and configuration:
- Apple Container setup
- systemd-nspawn setup
- File system layout
- ayod operations
- Resource limits
- Network configuration

### 6. security.md
Security model and guardrails:
- Sandbox isolation
- file_request flow
- --no-jodas mode
- Trust levels
- Externalized prompts
- Audit logging

## Requirements

- Every configuration option documented
- Default values specified
- Validation rules explained
- Examples for each option
- Error message explanations

## Success Criteria

- [ ] All 6 guides complete
- [ ] All config options covered
- [ ] Examples tested
- [ ] Cross-references working

---

*Created: 2026-02-23*
