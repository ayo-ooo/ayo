---
id: ayo-e2e10
status: open
deps: [ayo-e2e9]
links: []
created: 2026-02-24T14:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-e2e1
tags: [gtm, documentation, testing, e2e]
---
# Task: E2E Section 8 - Plugins

## Summary

Write Section 8 of the E2E Manual Testing Guide covering the plugin system, configuration, and management.

## Content Requirements

### List Plugins
```bash
./ayo plugin list

# Expected: Shows installed plugins with:
# - Name
# - Version
# - Status (enabled/disabled)
# - Type (planner, tool, provider, etc.)
```

### Plugin Info
```bash
# Get details on a plugin
./ayo plugin info ayo-tickets

# Expected:
# - Full description
# - Configuration options
# - Dependencies
# - Author/source
```

### Built-in Plugins
```bash
# List built-in plugins
./ayo plugin list --builtin

# Expected built-in plugins:
# - ayo-tickets (long-term planner)
# - ayo-todos (near-term planner)
# - ayo-memory (memory system)
# - etc.
```

### Plugin Enable/Disable
```bash
# Disable a plugin
./ayo plugin disable ayo-todos

# Verify disabled
./ayo plugin list
# Expected: ayo-todos shows disabled

# Re-enable
./ayo plugin enable ayo-todos

# Verify enabled
./ayo plugin list
```

### Plugin Configuration
```bash
# View plugin config
./ayo plugin config ayo-tickets

# Expected: Shows current configuration

# Set config (if applicable)
./ayo plugin config ayo-tickets --set prefix=PRJ
```

### Plugin in Agent Config
```bash
# View how plugins are assigned to agents
./ayo agents show @ayo

# Verify plugin configuration in agent JSON
cat ~/.local/share/ayo/agents/ayo.json | jq '.plugins'
```

### Plugin in Squad Config
```bash
# SQUAD.md can specify planners
cat << 'EOF'
---
planners:
  near_term: "ayo-todos"
  long_term: "ayo-tickets"
---
EOF

# This overrides agent defaults for squad context
```

### Custom Plugin Installation (if applicable)
```bash
# Install from path
./ayo plugin install /path/to/custom-plugin

# Or from registry
./ayo plugin install ayo-custom-tool

# Verify installation
./ayo plugin list
```

### Plugin Removal (for custom plugins)
```bash
./ayo plugin rm custom-plugin

# Built-in plugins cannot be removed
./ayo plugin rm ayo-tickets
# Expected: Error - cannot remove built-in plugin
```

### Verification Criteria
- [ ] Plugin listing works
- [ ] Plugin info shows details
- [ ] Plugin enable/disable works
- [ ] Plugin configuration accessible
- [ ] Plugins correctly used by agents
- [ ] Squad-level plugin overrides work

## Acceptance Criteria

- [ ] Section written in guide
- [ ] All plugin commands documented
- [ ] Built-in plugins listed
- [ ] Configuration options documented
- [ ] Agent/squad plugin integration explained
