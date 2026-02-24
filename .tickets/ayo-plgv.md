---
id: ayo-plgv
status: closed
deps: [ayo-plex, ayo-plsq, ayo-pltg, ayo-plsb, ayo-plrg]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-plug
tags: [gtm, phase8, e2e]
---
# Phase 8 E2E: Plugin System Verification

End-to-end verification that the expanded plugin system works correctly.

## Test Scenarios

### External Planner Loading
```bash
# Create a test external planner plugin
ayo plugin install ./test-fixtures/mock-planner-plugin

# Verify it's registered
ayo planner list | grep "test-planner"

# Use it in a session
ayo --planner-near=test-planner "plan my day"

# Verify state directory created
ls ~/.local/share/ayo/sandboxes/ayo/.planner.test-planner/
```

### Squad Plugins
```bash
# Install squad from plugin
ayo plugin install ./test-fixtures/squad-plugin

# List shows plugin squad
ayo squad list | grep "plugin-squad"

# Instantiate plugin squad
ayo squad create test-squad --from=plugin-squad

# Verify SQUAD.md and ayo.json copied
cat ~/.local/share/ayo/sandboxes/squads/test-squad/SQUAD.md
cat ~/.local/share/ayo/sandboxes/squads/test-squad/ayo.json
```

### Trigger Plugins
```bash
# Install trigger plugin
ayo plugin install ./test-fixtures/mock-trigger-plugin

# List trigger types includes new type
ayo trigger types | grep "mock-trigger"

# Create trigger using plugin type
ayo trigger add test-trigger --type=mock-trigger --config '{"interval": "1m"}'

# Verify trigger registered
ayo trigger list | grep "test-trigger"

# Stop and cleanup
ayo trigger remove test-trigger
```

### Sandbox Config Plugins
```bash
# Install sandbox config plugin
ayo plugin install ./test-fixtures/sandbox-config-plugin

# List shows plugin config
ayo sandbox configs | grep "gpu-enabled"

# Create sandbox with plugin config
ayo sandbox create test-sandbox --config=gpu-enabled

# Verify config applied
ayo sandbox inspect test-sandbox | grep "gpu"
```

### Plugin Registry
```bash
# Search across all component types
ayo plugin search "email"

# Validate all installed plugins
ayo plugin validate

# Show dependency tree
ayo plugin deps ayo-plugins-imap
```

## Success Criteria

- [ ] External .so planner plugins load and function
- [ ] Squad definitions can be installed from plugins
- [ ] Custom trigger types register correctly
- [ ] Sandbox configs apply correctly
- [ ] Plugin registry lists all component types
- [ ] No regressions in existing plugin functionality
- [ ] `ayo doctor` validates plugin health

## Test Fixtures Required

Create test fixtures in `test/fixtures/plugins/`:
- `mock-planner-plugin/` - Minimal external planner
- `squad-plugin/` - Squad with SQUAD.md and ayo.json
- `mock-trigger-plugin/` - Custom trigger type
- `sandbox-config-plugin/` - Alternative sandbox config
