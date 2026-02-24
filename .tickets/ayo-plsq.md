---
id: ayo-plsq
status: open
deps: [ayo-pv3a]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-plug
tags: [plugins, squads]
---
# Task: Squad Plugin Support

## Summary

Enable plugins to provide reusable squad definitions, complete with SQUAD.md constitutions, ayo.json configurations, and bundled agents/skills. This makes squads packageable and shareable like agents.

## Context

Currently squads are only defined locally in:
```
~/.local/share/ayo/sandboxes/squads/{name}/
├── SQUAD.md
├── ayo.json
├── workspace/
└── .tickets/
```

But there's no way to:
- Package a squad definition in a plugin
- Share squad templates between users
- Distribute pre-configured team structures

## Technical Approach

### Plugin Directory Structure

```
my-squad-plugin/
├── manifest.json
├── squads/
│   └── dev-team/
│       ├── SQUAD.md           # Constitution
│       ├── ayo.json           # Squad config
│       └── agents/            # Bundled agents (optional)
│           ├── @architect/
│           └── @developer/
└── skills/                    # Shared skills
    └── code-review.md
```

### Manifest Schema

```json
{
  "name": "my-squad-plugin",
  "version": "1.0.0",
  "components": {
    "squads": {
      "dev-team": {
        "path": "squads/dev-team",
        "description": "Full-stack development squad with architect and developer",
        "agents": ["@architect", "@developer"],
        "planners": {
          "near_term": "ayo-todos",
          "long_term": "ayo-tickets"
        }
      }
    }
  }
}
```

### Squad Installation

When a squad plugin is installed:

1. Squad definition is registered in the plugin registry
2. SQUAD.md and ayo.json are read from plugin path
3. Bundled agents become available (prefixed with squad name?)
4. On first use (`ayo #dev-team "task"`), sandbox is created from template

### Squad Resolution Order

1. Local squads (`~/.local/share/ayo/sandboxes/squads/{name}/`)
2. Plugin squads (from manifest)
3. Error if not found

## Implementation Steps

1. [ ] Add `squads` component type to manifest schema
2. [ ] Update manifest loader to parse squad definitions
3. [ ] Implement squad resolution from plugins in `internal/squads/`
4. [ ] Handle bundled agents within squad plugins
5. [ ] Support squad template instantiation
6. [ ] Add `ayo squad create --from-plugin` command
7. [ ] Update documentation

## Dependencies

- Depends on: `ayo-pv3a` (unified ayo.json schema)
- Blocks: Plugin repository squads

## Acceptance Criteria

- [ ] Plugins can define squads with SQUAD.md + ayo.json
- [ ] Squad plugins can bundle their own agents
- [ ] `ayo squad list` shows plugin squads
- [ ] Plugin squads can be instantiated locally
- [ ] Documentation covers squad plugin development

## Files to Modify

- `internal/plugins/manifest.go` - Add squads component type
- `internal/plugins/registry.go` - Register discovered squads
- `internal/squads/loader.go` - Load from plugins
- `cmd/ayo/squad.go` - Add --from-plugin flag

## Design Decisions

### Bundled vs Referenced Agents

Two options for squad agents:

1. **Bundled**: Agents defined within squad plugin directory
   - Pro: Self-contained, portable
   - Con: Duplication if agent used elsewhere

2. **Referenced**: Squad references external agents by name
   - Pro: Reuse existing agents
   - Con: Dependency management

**Decision**: Support both. Bundled agents for portability, referenced for reuse.

### Squad Instantiation

When user first invokes a plugin squad:
1. Create sandbox from plugin template
2. Copy SQUAD.md and ayo.json to sandbox
3. Install bundled agents if any
4. Initialize planners

---

*Created: 2026-02-23*
