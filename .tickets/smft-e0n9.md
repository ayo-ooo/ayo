---
id: smft-e0n9
status: closed
deps: []
links: []
created: 2026-02-12T23:52:53Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-w6k3
---
# AGENTS.md: Replace with new header and overview

Replace the current AGENTS.md content with a new concise header and project overview.

**File:** AGENTS.md

**New content (lines 1-25):**
```markdown
# Ayo Agent Memory

Quick reference for AI coding agents working on the ayo codebase.

## Project Overview

Ayo is a CLI framework for creating, managing, and orchestrating AI agents that operate in isolated sandbox environments.

**Core architecture:**
- Host process: LLM calls, memory, orchestration
- Sandbox container: Command execution, file operations

**Sandbox providers:** Apple Container (macOS 26+), systemd-nspawn (Linux). NOT Docker.

For comprehensive documentation, see `docs/`.
```

**Delete:** All existing content below the header
**This is a full replacement, not an edit**

