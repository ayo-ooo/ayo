---
id: ayo-plrg
status: open
deps: [ayo-plex, ayo-plsq, ayo-pltg, ayo-plsb]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-plug
tags: [plugins, registry]
---
# Task: Plugin Registry Improvements

## Summary

Improve the plugin registry to handle the expanded component types (squads, triggers, sandbox configs) and provide better discovery, validation, and dependency resolution.

## Context

The current registry (`internal/plugins/registry.go`) handles:
- Agents, skills, tools
- Delegates, default_tools
- Providers

With the expansion to squads, triggers, and sandbox configs, the registry needs:
- More component type support
- Better dependency tracking
- Validation for new types
- Improved listing/discovery

## Technical Approach

### Registry Structure

```go
type Registry struct {
    plugins       map[string]*Plugin
    agents        map[string]*AgentRef
    skills        map[string]*SkillRef
    tools         map[string]*ToolRef
    delegates     map[string]*DelegateRef
    defaultTools  map[string]string
    providers     map[string]*ProviderRef
    // New types
    squads        map[string]*SquadRef
    triggers      map[string]*TriggerRef
    sandboxCfgs   map[string]*SandboxConfigRef
    planners      map[string]*PlannerRef
}

type SquadRef struct {
    Name        string
    Plugin      string
    Path        string
    Description string
    Agents      []string
}

type TriggerRef struct {
    Name        string
    Plugin      string
    Path        string
    Type        string  // poll, push, watch
    ConfigSchema map[string]any
}

type SandboxConfigRef struct {
    Name         string
    Plugin       string
    Path         string
    Requirements []string
}

type PlannerRef struct {
    Name   string
    Plugin string
    Path   string
    Type   string  // near, long
}
```

### Discovery Improvements

```go
// Discovery results by type
func (r *Registry) ListSquads() []*SquadRef
func (r *Registry) ListTriggers() []*TriggerRef
func (r *Registry) ListSandboxConfigs() []*SandboxConfigRef
func (r *Registry) ListPlanners() []*PlannerRef

// Search across all types
func (r *Registry) Search(query string) []ComponentRef
```

### Dependency Resolution

```go
type Dependency struct {
    Type    string  // "agent", "skill", "tool", etc.
    Name    string
    Version string  // Optional
}

type Plugin struct {
    // ...existing fields
    Dependencies []Dependency
    Conflicts    []string  // Plugin names that conflict
}

func (r *Registry) ResolveDependencies(plugin string) ([]string, error)
func (r *Registry) CheckConflicts(plugin string) []string
```

### Validation

```go
func (r *Registry) ValidatePlugin(p *Plugin) []ValidationError

type ValidationError struct {
    Field   string
    Message string
    Severity string  // "error", "warning"
}
```

## Implementation Steps

1. [ ] Add new component maps to Registry struct
2. [ ] Update `loadPlugin()` to handle new types
3. [ ] Implement list functions for new types
4. [ ] Add cross-type search functionality
5. [ ] Implement dependency resolution
6. [ ] Add conflict detection
7. [ ] Improve validation for all types
8. [ ] Update `ayo plugin list` to show all types
9. [ ] Add `ayo plugin search` command
10. [ ] Update tests

## Dependencies

- Depends on: `ayo-plex`, `ayo-plsq`, `ayo-pltg`, `ayo-plsb` (new component types)
- Blocks: None

## Acceptance Criteria

- [ ] Registry handles all new component types
- [ ] `ayo plugin list` shows squads, triggers, sandbox configs
- [ ] Dependency resolution works for cross-plugin refs
- [ ] Conflicts are detected and reported
- [ ] Validation catches schema errors

## Files to Modify

- `internal/plugins/registry.go` - Add new types
- `internal/plugins/manifest.go` - Expand schema
- `cmd/ayo/plugin.go` - Update list command

## Performance Considerations

- Registry is loaded once at startup
- Consider lazy loading for large plugin sets
- Cache validated manifests

---

*Created: 2026-02-23*
