---
id: ayo-rx05
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
# Task: Create docs/guides/ (6 configuration guides)

## Summary

Create 6 configuration guides in `docs/guides/` that explain how to configure each major component in detail.

## Files to Create

### 1. guides/agents.md (~400 lines)

Complete agent configuration reference:
- Directory structure (`@agent-name/`)
- `system.md` format and best practices
- `ayo.json` agent schema (all fields)
- Tools configuration
- Memory settings
- Sandbox options
- Permissions (`file_request`, `auto_approve`)
- Skills integration
- Example: complete agent configuration

### 2. guides/squads.md (~400 lines)

Complete squad configuration reference:
- Directory structure (`~/.local/share/ayo/sandboxes/squads/{name}/`)
- `SQUAD.md` format and semantics
- YAML frontmatter options
- `ayo.json` squad schema
- Agent roster configuration
- Planners (near-term, long-term)
- Sandbox options
- Memory sharing between agents
- Ticket-based coordination
- Example: complete squad configuration

### 3. guides/triggers.md (~350 lines)

Complete trigger configuration reference:
- Trigger types (cron, interval, one_time, file_watch)
- gocron v2 features and syntax
- Cron expression examples
- Schedule aliases (`@daily`, `@weekly`)
- File watch patterns
- Trigger persistence
- Output handling
- Error handling and retries
- CLI commands (`ayo trigger add/list/remove/fire`)
- Example: various trigger configurations

### 4. guides/tools.md (~300 lines)

Built-in and external tools reference:
- Complete built-in tools list:
  - bash, view, edit, glob, grep
  - memory_store, memory_search
  - file_request, publish
  - delegate, human_input
- Tool configuration in ayo.json
- External tool format (JSON schema)
- Tool permissions model
- Creating custom tools
- Tool execution environment
- Example: custom tool definition

### 5. guides/sandbox.md (~350 lines)

Sandbox architecture and configuration:
- Supported providers:
  - Apple Container (macOS 26+)
  - systemd-nspawn (Linux)
- Provider configuration
- File system layout inside sandbox
- ayod (in-sandbox daemon)
- Resource limits
- Network configuration
- Mount points (`/mnt/{username}`, `/output`)
- Sandbox lifecycle
- Debugging sandbox issues
- Example: custom sandbox configuration

### 6. guides/security.md (~300 lines)

Security model and guardrails:
- Sandbox isolation principles
- file_request approval flow
- `--no-jodas` mode risks and benefits
- Trust levels
- Permission precedence
- Externalized prompts (guardrails, sandwich)
- Audit logging (`~/.local/share/ayo/audit.log`)
- Best practices for secure agent deployment
- Example: secure agent configuration

## Acceptance Criteria

- [ ] All 6 guide files exist in `docs/guides/`
- [ ] Every configuration option documented
- [ ] Default values specified for all options
- [ ] Validation rules explained
- [ ] Examples for each major option
- [ ] Error message explanations
- [ ] Cross-references to tutorials and reference docs
