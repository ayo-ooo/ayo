---
id: ayo-doc2
status: open
deps: [ayo-doc1]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-docs
tags: [documentation, phase9]
---
# Task: Write README.md and Getting Started Guide

## Summary

Write the project README.md and getting-started guide as the primary entry points for new users.

## README.md Structure

```markdown
# Ayo

> CLI framework for managing AI agents in isolated sandboxes

## Features

- Sandboxed execution (Apple Container, systemd-nspawn)
- Multi-agent coordination with Squads
- Persistent memory system
- Flexible triggers (cron, file watch, events)
- Extensible plugin system

## Quick Start

[5-line install and first use]

## Documentation

[Links to docs/]

## Installation

[Platform-specific instructions]

## Contributing

[Contribution guide]

## License

[License info]
```

## getting-started.md Structure

```markdown
# Getting Started

## Installation

### macOS
[Homebrew, binary download]

### Linux
[Package managers, binary download]

## First Agent

### 1. Setup
ayo setup

### 2. Run your first prompt
ayo "Hello, what can you do?"

### 3. Create a custom agent
ayo agent new @my-agent

### 4. Use squads
ayo squad create my-team

## Next Steps

[Links to tutorials]
```

## Requirements

- Installation tested on macOS and Linux
- All commands verified working
- Screenshots/output examples included
- Clear progression from simple to complex

## Success Criteria

- [ ] README.md is comprehensive but scannable
- [ ] getting-started.md takes < 5 minutes to complete
- [ ] All commands in docs work exactly as shown
- [ ] No broken links

---

*Created: 2026-02-23*
