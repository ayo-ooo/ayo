---
id: ayo-jfqs
status: completed
deps: [ayo-z9oo, ayo-gn7f]
links: ["epic:build-system-refactor", "blocks:gl1u,seqz,3bw1"]
created: 2026-03-07T21:04:14Z
type: task
priority: 2
assignee: Alex Cabrera
---
# Refactor agent loading for embedded config

✅ COMPLETED: Refactor internal/agent/ to load config from embedded config.toml instead of ~/.config/ayo/agents/

## Implementation Summary
- ✅ Removed old directory-based agent loading from ~/.config/ayo/agents/
- ✅ Implemented TOML-based agent loading from project directories with config.toml
- ✅ Updated internal/agent/Load() to work with config.toml files (build system approach)
- ✅ Removed dependencies on builtin agents, database, and session systems
- ✅ Removed memory package entirely (agents manage own memory in build system)
- ✅ Removed delegate system (not needed in build system)
- ✅ Removed agent creation functions (handled by ayo fresh)
- ✅ Added RemoteAgentsCacheDir() for future remote agent support
- ✅ Updated history renderer to use local Message type
- ✅ Agent package now compiles successfully

## Key Changes
1. **New agent loading priority**: Current dir → agents/ subdir → remote cache
2. **TOML-based config**: Agents defined by config.toml files
3. **Simplified architecture**: Removed framework dependencies
4. **Build system focus**: All functionality supports ayo build workflow

## Testing
- ✅ Agent package compiles without errors
- ✅ All framework dependencies removed
- ✅ Core build system functionality preserved

