---
id: ayo-path
title: Consistent path and config naming
status: open
priority: high
assignee: "@ayo"
tags: [docs, gtm, polish]
created: 2026-02-24
---

# Consistent path and config naming

Documentation uses inconsistent path references and config file naming.

## Problem

1. Agent directories sometimes shown with `@` prefix, sometimes without
2. Config files called both `config.json` and `ayo.json`
3. Path references differ between docs

## Path Inconsistencies

### Agent Directory Names

| File | Shows | Actual |
|------|-------|--------|
| `docs/concepts.md:47-48` | `@agent-name/` | Directories use `@` prefix |
| `docs/tutorials/first-agent.md:20-21` | `~/.config/ayo/agents/@reviewer/` | Correct |
| `docs/reference/plugins.md:17-20` | `agents/my-agent/` | Missing `@` |

**Decision**: Agent directories **DO** use `@` prefix: `~/.config/ayo/agents/@myagent/`

### Config File Naming

| Usage | File | Notes |
|-------|------|-------|
| `config.json` | Most examples | Legacy format |
| `ayo.json` | New format | Has `$schema` and namespaced sections |

**Decision**: 
- New agents use `ayo.json` (preferred)
- Legacy `config.json` still supported
- Docs should show `ayo.json` as primary

### Path Locations

| Location | Purpose |
|----------|---------|
| `~/.config/ayo/` | User configuration (agents, config) |
| `~/.config/ayo/agents/` | User-created agents |
| `~/.config/ayo/ayo.json` | Global configuration |
| `~/.local/share/ayo/` | Data (built-in agents, memories, sessions) |
| `~/.local/share/ayo/agents/` | Built-in agents |
| `./.config/ayo/` | Project-local configuration |
| `./.config/ayo/agents/` | Project-local agents |

## Files to Update

### Agent Path References

| File | Line | Change |
|------|------|--------|
| `docs/reference/plugins.md:17-20` | Add `@` prefix to agent dirs |
| `docs/reference/cli.md` | Verify path examples |

### Config File References

| File | Line | Change |
|------|------|--------|
| `README.md:50-52` | Show `ayo.json` format |
| `docs/guides/agents.md` | Use `ayo.json` examples |
| `docs/tutorials/first-agent.md` | Use `ayo.json` examples |

### Memory Config Format

| File | Line | Issue |
|------|------|-------|
| `docs/memory.md:119-126` | Shows TOML format but ayo uses JSON |

## Acceptance Criteria

- [ ] All agent directory examples use `@` prefix
- [ ] All config examples use `ayo.json` format
- [ ] Path locations table in getting-started.md
- [ ] No TOML config examples (use JSON)
- [ ] Project-local vs user vs data directories clearly explained

## Dependencies

None

## Notes

The code supports both `config.json` and `ayo.json` (see `agent.go:loadAgentConfig`).
`ayo.json` is preferred for new agents with `$schema` support.
