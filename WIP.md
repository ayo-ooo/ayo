# Work In Progress: Plugin System

## Status: Implementation Complete, Pending Testing

The plugin system has been fully implemented. All tests pass. Ready for manual integration testing.

## What Was Done

### Core Plugin Infrastructure (`internal/plugins/`)

| File | Purpose |
|------|---------|
| `manifest.go` | Plugin manifest schema, validation, URL parsing |
| `registry.go` | Installed plugins tracking (`packages.json`) |
| `install.go` | Git clone, local install, dependency checking |
| `update.go` | Version-pinned updates, dry-run support |
| `remove.go` | Plugin uninstallation |
| `tools.go` | External tool definition schema (`tool.json`) |
| `resolve.go` | Conflict detection and resolution |

### Delegation System (`internal/delegates/`)

- 3-layer priority: `.ayo.json` (directory) > agent config > global config
- Task types: `coding`, `research`, `debug`, `test`, `docs`
- `Delegates` field added to both agent and global config

### Integration Changes

- `internal/paths/paths.go` - Added `PluginsDir()`, `AllPluginAgentsDirs()`, `AllPluginSkillsDirs()`, `FindDirectoryConfig()`
- `internal/agent/agent.go` - Plugin agent discovery in `ListHandles()` and `Load()`
- `internal/skills/discover.go` - Plugin skill discovery with `SourcePlugin`
- `internal/run/external_tools.go` - External tool executor implementing `fantasy.AgentTool`
- `internal/run/fantasy_tools.go` - `loadExternalTool()` for dynamic plugin tool loading
- `internal/config/config.go` - Added `Delegates` field

### CLI Commands (`cmd/ayo/plugins.go`)

```bash
ayo plugins install <ref>     # Install from git or --local
ayo plugins list              # List installed plugins
ayo plugins show <name>       # Show plugin details  
ayo plugins update [name]     # Update plugins (--dry-run supported)
ayo plugins remove <name>     # Remove plugin (--yes to skip confirm)
```

### Removed Built-ins

- `internal/builtin/agents/@ayo.coding/` - Deleted (now via plugin)
- `internal/crush/` - Deleted (no longer needed)
- `NewCrushTool()` in `fantasy_tools.go` - Removed
- `CrushParams` struct - Removed

### Updated Skills

- `coding` skill - Now references `@crush` from plugin, not `@ayo.coding`
- `ayo` skill - Added plugins commands documentation
- New `plugins` skill - Plugin management help

### Documentation

- `AGENTS.md` - Full plugin system documentation, delegation system, updated crush integration

## Crush Plugin Repository

Located at `/Users/acabrera/Code/ayo-plugins-crush/` (committed but not pushed):

```
ayo-plugins-crush/
├── manifest.json              # name: "crush", version: "1.0.0"
├── README.md                  # Installation and usage docs
├── agents/@crush/
│   ├── config.json           # allowed_tools: ["crush"]
│   └── system.md             # System prompt
├── skills/crush-coding/
│   └── SKILL.md              # Crush usage guidance
└── tools/crush/
    └── tool.json             # Maps to `crush run --quiet`
```

**Needs:** Push to `https://github.com/alexcabrera/ayo-plugins-crush`

## Testing Checklist

- [x] All unit tests pass (`go test ./...`)
- [x] Build succeeds (`go build ./cmd/ayo`)
- [ ] Manual test: `ayo plugins install alexcabrera/crush`
- [ ] Manual test: `ayo plugins list`
- [ ] Manual test: `ayo @crush "simple task"`
- [ ] Manual test: Delegation via `.ayo.json`
- [ ] Manual test: `ayo plugins update`
- [ ] Manual test: `ayo plugins remove crush`

## Next Steps

1. Push crush plugin repo to GitHub
2. Run `./install.sh` to rebuild ayo with new builtins
3. Test full plugin lifecycle manually
4. Consider adding interactive conflict resolution prompts (currently just detects conflicts)

## Key Design Decisions

| Decision | Choice |
|----------|--------|
| Package naming | `ayo-plugins-<name>` |
| Tool distribution | Plugins bring own tools via `tool.json` |
| Delegation priority | Directory > Agent > Global config |
| Version management | Pinned versions, explicit upgrade |
| Built-in crush | Completely removed, now plugin-only |

## Files Changed in This Branch

```
cmd/ayo/plugins.go                    # New - CLI commands
cmd/ayo/root.go                       # Added plugins command
internal/plugins/*.go                 # New - plugin system
internal/delegates/delegates.go       # New - delegation resolution
internal/delegates/delegates_test.go  # New - tests
internal/plugins/*_test.go            # New - tests
internal/agent/agent.go               # Plugin agent discovery
internal/skills/skills.go             # Added SourcePlugin
internal/skills/discover.go           # Plugin skill discovery
internal/config/config.go             # Added Delegates field
internal/paths/paths.go               # Plugin paths
internal/run/fantasy_tools.go         # Removed crush, added plugin loading
internal/run/external_tools.go        # New - external tool executor
internal/builtin/install.go           # Bumped version to 13
internal/builtin/agents/@ayo.coding/  # Deleted
internal/builtin/skills/coding/       # Updated for plugins
internal/builtin/skills/ayo/          # Added plugins docs
internal/builtin/skills/plugins/      # New - plugin management skill
AGENTS.md                             # Plugin system documentation
```
