---
id: ayo-bs1
status: done
deps: []
links: []
created: 2026-03-11T18:00:00Z
type: epic
priority: 1
assignee: Alex Cabrera
tags: [build-system, cleanup, plugin-removal]
---
# Phase 1: Remove Plugin System

Remove the entire plugin system infrastructure as part of the build system transformation. Skills and tools will be managed directly via directories instead of plugins.

## Context

The new build system eliminates the plugin architecture in favor of:
- **Skills**: Direct skills/ directory using agentskills.io format
- **Tools**: Direct tools/ directory with executable binaries
- **No Plugin Registry**: No manifests, no installation commands

This simplifies the architecture and aligns with the "pure build system" vision.

## Scope

### Code to Remove

1. **Plugin Loading**
   - internal/plugins/ package (if exists)
   - Plugin registry and discovery code
   - Plugin manifest parsing

2. **CLI Commands**
   - `ayo plugins install`
   - `ayo plugins list`
   - `ayo plugins remove`
   - `ayo plugins update`
   - Related command files in cmd/ayo/

3. **Documentation**
   - .docs/PLUGINS.md
   - Plugin-related sections in README.md
   - Plugin examples in docs/

4. **Configuration**
   - Plugin-related config options
   - Plugin paths in internal/paths/
   - Plugin registry storage

## Deliverables

- [ ] Plugin loading code completely removed
- [ ] All plugin CLI commands removed
- [ ] Plugin documentation removed
- [ ] No plugin-related config options remain
- [ ] All tests pass after removal
- [ ] No compilation errors or warnings

## Out of Scope

- Keep the ayo-plugins-* directories in parent (these are examples/migration references)
- Keep internal/skills/ (this is different from plugin skills)
- Keep internal/tools/ (this is different from plugin tools)

## Acceptance Criteria

1. Running `ayo plugins` shows "unknown command"
2. No references to plugins in internal code
3. Codebase compiles cleanly
4. All existing tests pass
5. Documentation no longer mentions plugins (except historical notes)

## Risks

- **Breaking Change**: Existing users with plugins will need to migrate
  - **Mitigation**: Provide migration guide in Phase 10

## Dependencies

None - this can start immediately

## Subtasks

- [ ] Remove plugin loading code
- [ ] Remove plugin CLI commands
- [ ] Remove plugin documentation
- [ ] Update configuration code
- [ ] Update paths package
- [ ] Run full test suite
- [ ] Verify no plugin references remain
