---
id: ayo-gtm2
status: open
deps: []
links: []
created: 2026-02-24T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-gtm1
tags: [gtm, refactoring]
---
# Task: Split internal/agent/agent.go

## Summary

`internal/agent/agent.go` is 1320 lines. Split into logical modules for maintainability.

## Current Structure Analysis

The file contains:
1. **Type definitions** - Agent struct, config types (~200 lines)
2. **Loading functions** - LoadAgent, LoadAgentFromDir, etc. (~300 lines)
3. **Memory integration** - Memory-related methods (~150 lines)
4. **Configuration** - GetSystemPrompt, GetGuardrails, etc. (~200 lines)
5. **Identity** - Handle parsing, trust levels (~100 lines)
6. **Tools** - Tool configuration, capabilities (~200 lines)
7. **Validation** - Config validation, defaults (~170 lines)

## Proposed Split

| New File | Contents | Est. Lines |
|----------|----------|------------|
| `agent.go` | Core Agent struct, basic methods | ~200 |
| `loading.go` | LoadAgent*, FindAgent, directory scanning | ~300 |
| `config.go` | Config types, validation, defaults | ~250 |
| `identity.go` | Handle parsing, trust levels, identity | ~150 |
| `memory.go` | Memory integration methods | ~200 |
| `prompts.go` | GetSystemPrompt, GetGuardrails, sandwich | ~220 |

## Implementation Steps

1. [ ] Create `loading.go` - move all Load* and Find* functions
2. [ ] Create `config.go` - move config types and validation
3. [ ] Create `identity.go` - move handle parsing and trust
4. [ ] Create `memory.go` - move memory integration
5. [ ] Create `prompts.go` - move prompt assembly
6. [ ] Update imports in all affected files
7. [ ] Run tests to verify no breakage
8. [ ] Verify no circular imports

## Acceptance Criteria

- [ ] No file > 300 lines
- [ ] All tests pass
- [ ] No circular import errors
- [ ] Clean package structure
