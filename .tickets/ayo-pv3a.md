---
id: ayo-pv3a
status: open
deps: [ayo-whmn]
links: []
created: 2026-02-23T23:13:09Z
type: epic
priority: 0
assignee: Alex Cabrera
tags: [gtm, phase3]
---
# Phase 3: Unified Configuration

Single ayo.json schema for agents and squads. Replace SQUAD.md frontmatter with JSON config.

## Goals

- Design unified ayo.json schema with `agent` and `squad` namespaces
- Implement loaders for both agents and squads
- Migrate SQUAD.md frontmatter to ayo.json
- SQUAD.md becomes pure documentation (no YAML frontmatter)
- Update CLI commands to work with new schema

## Key Decisions

1. **Schema namespace approach**: `agent` and `squad` top-level keys
2. **Versioning**: Include `version` field for future migrations
3. **JSON Schema file**: Publish at `/schemas/ayo.json` for editor support

## Child Tickets

- `ayo-7dui`: Design ayo.json schema with agent namespace
- `ayo-la11`: Implement ayo.json loader for agents
- `ayo-nqyv`: Design ayo.json schema with squad namespace
- `ayo-7jth`: Implement ayo.json loader for squads
- `ayo-mp44`: Migrate SQUAD.md frontmatter to ayo.json

