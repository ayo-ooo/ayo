---
id: ayo-bs2
status: open
deps: []
links: [ayo-bs1]
created: 2026-03-11T18:00:00Z
type: epic
priority: 2
assignee: Alex Cabrera
tags: [build-system, core, executable-generation]
---
# Phase 2: Build System Core

Complete the build system to generate working standalone executables from agent definitions. This is the heart of the build system transformation.

## Context

The current build command (cmd/ayo/build.go) exists but generates only placeholder stubs. We need to:

1. Generate complete main.go with Fantasy initialization
2. Embed all agent resources (config, skills, tools)
3. Implement proper cross-platform build support
4. Ensure generated binaries run independently

## Tasks

### 2.1 Complete main.go Stub Generation
- [ ] Implement proper imports for Fantasy and ayo internals
- [ ] Initialize Fantasy agent with embedded config
- [ ] Parse embedded config.toml
- [ ] Set up CLI argument parsing
- [ ] Implement main execution loop

### 2.2 Implement Resource Embedding
- [ ] Embed config.toml using `//go:embed`
- [ ] Embed prompts/system.md
- [ ] Embed all SKILL.md files from skills/
- [ ] Embed tool binaries from tools/
- [ ] Create resource loader for embedded files

### 2.3 Cross-Platform Build Support
- [ ] Support GOOS/GOARCH flags
- [ ] Add proper Windows extension (.exe)
- [ ] Handle platform-specific paths
- [ ] Set appropriate file permissions

### 2.4 Build System Tests
- [ ] Unit tests for main.go generation
- [ ] Unit tests for resource embedding
- [ ] Integration tests for full build
- [ ] Tests for cross-platform builds
- [ ] Tests for resource loading

## Technical Details

### Generated main.go Structure

```go
package main

import (
    _ "embed"
    "os"
    "github.com/alexcabrera/ayo/internal/build/runtime"
)

//go:embed config.toml
var configToml []byte

//go:embed prompts/system.md
var systemPrompt []byte

//go:embed skills/**/*.md
var skills embed.FS

//go:embed tools/*
var tools embed.FS

func main() {
    runtime.Execute(configToml, systemPrompt, skills, tools)
}
```

### Resource Embedding Strategy

- Use Go's embed package for static files
- Embed as filesystem for skills/tools directories
- Embed as bytes for single files (config, prompts)
- Ensure proper escaping and encoding

## Deliverables

- [ ] Complete main.go generation
- [ ] All resources embedded properly
- [ ] Cross-platform builds work (Linux, macOS, Windows AMD64/ARM64)
- [ ] Generated binaries run standalone
- [ ] Test coverage > 80% for build code
- [ ] Build command documentation updated

## Acceptance Criteria

1. `ayo build my-agent` creates working executable
2. Executable runs without ayo installed
3. All embedded resources load correctly
4. Cross-platform builds produce valid binaries
5. Build process is fast (<10s for typical agent)

## Dependencies

- **ayo-bs1**: Plugin system removed (clean slate)

## Out of Scope

- Input schema to CLI conversion (Phase 5)
- Skills prompt injection (Phase 3)
- Tools execution framework (Phase 4)
- Output schema validation (Phase 6)

## Risks

- **Complexity**: Generating working executables with embedded resources is complex
  - **Mitigation**: Start simple, iterate rapidly, extensive testing
- **Binary Size**: Embedding all resources may increase size
  - **Mitigation**: Test size, optimize if needed (Phase 10)

## Notes

This is the critical path for the entire build system. Other phases depend on this core working.
