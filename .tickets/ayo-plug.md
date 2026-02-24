---
id: ayo-plug
status: closed
deps: [ayo-xfu3]
links: []
created: 2026-02-23T12:00:00Z
type: epic
priority: 1
assignee: Alex Cabrera
tags: [gtm, phase8, plugins]
---
# Epic: Plugin System Expansion

## Summary

Expand ayo's plugin system to support additional component types beyond agents, tools, skills, and delegates. The current plugin architecture is solid but limited—this epic extends it to cover squads, triggers, sandbox configurations, and completes the external planner loading that currently has a TODO placeholder.

## Context

### Current Plugin Capabilities

Plugins can currently provide:
- `agents` - Agent definitions (@agent-name)
- `skills` - SKILL.md files with instructions
- `tools` - External command definitions (tool.json)
- `delegates` - Task routing (e.g., "coding" -> @crush)
- `default_tools` - Tool alias mappings (e.g., "search" -> "searxng")
- `providers` - memory, sandbox, embedding, observer types
- `planners` - near-term/long-term planning tools

### What's Missing

1. **Squad plugins** - No way to package reusable squad definitions
2. **Trigger plugins** - Triggers are hardcoded in daemon's TriggerEngine
3. **External planner loading** - Has TODO at `internal/plugins/planners.go:100-107`
4. **Sandbox config plugins** - No alternative harness/container configurations

### Why This Matters

- MCP ecosystem shows strong demand for communication triggers (IMAP, webhooks, etc.)
- Squads need to be shareable/packageable like agents
- Users should be able to add custom trigger types without modifying core code
- Plugin architecture is already extensible—just needs more component types

## MCP Ecosystem Research Insights

From analyzing 500+ MCP servers:

**Most popular/useful categories for ayo plugins**:
1. **Communication**: IMAP/SMTP, Slack, Discord, WhatsApp, Telegram
2. **Productivity**: Calendar, Task management, Note-taking
3. **Data**: Database connectors, Cloud storage
4. **Search**: Web search, Vector databases
5. **Automation**: Browser automation, CI/CD

**Key patterns for trigger plugins**:
- IMAP email listeners (poll-based or IDLE)
- Webhook receivers (HTTP endpoints)
- File watchers (already in ayo)
- Calendar event triggers
- Database change listeners

## Child Tickets

| Ticket | Title | Priority |
|--------|-------|----------|
| `ayo-plex` | External planner loading completion | high |
| `ayo-plsq` | Squad plugin support | high |
| `ayo-pltg` | Trigger plugin architecture | high |
| `ayo-plsb` | Sandbox config plugins | medium |
| `ayo-plrg` | Plugin registry improvements | medium |
| `ayo-plgv` | Phase 8 E2E verification | high |

## Dependencies

- Depends on: `ayo-sqad` (squad polish), `ayo-6h19` (foundation)
- Blocks: Sibling plugin repository epics

## Success Criteria

- [ ] Plugins can define squads with SQUAD.md + ayo.json
- [ ] Plugins can register custom trigger types
- [ ] External planners load via Go plugin (.so) mechanism
- [ ] Sandbox configs can be bundled in plugins
- [ ] At least one production plugin uses each new capability

## Technical Notes

### Manifest Schema Extension

```json
{
  "name": "my-squad-plugin",
  "version": "1.0.0",
  "components": {
    "squads": {
      "dev-team": {
        "path": "squads/dev-team",
        "description": "Full-stack development squad"
      }
    },
    "triggers": {
      "imap": {
        "path": "triggers/imap",
        "type": "poll",
        "description": "IMAP email trigger"
      }
    },
    "sandbox_configs": {
      "gpu-enabled": {
        "path": "sandboxes/gpu",
        "description": "GPU-accelerated sandbox"
      }
    }
  }
}
```

### Trigger Plugin Interface

```go
type TriggerPlugin interface {
    Name() string
    Type() string  // "poll", "push", "watch"
    Init(ctx context.Context, config map[string]any) error
    Start(ctx context.Context, callback TriggerCallback) error
    Stop() error
}

type TriggerCallback func(event TriggerEvent) error
```

---

*Created: 2026-02-23*
