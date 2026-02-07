---
id: nd-epic
status: done
deps: []
links: []
created: 2026-02-04T14:00:00Z
type: epic
priority: 1
assignee: ""
---
# Replace Docker with Native Container Solutions

## Overview

Remove Docker dependency entirely from the ayo codebase. Replace with platform-native container solutions:

- **macOS**: Apple Container (github.com/apple/container) - requires macOS 26+, Apple Silicon
- **Linux**: LXC/LXD or similar low-level native solution (distribution-specific)

## Rationale

Docker is a heavy dependency with significant overhead. Native container solutions provide:
- Lower resource usage
- Faster startup times
- Better integration with OS-level security (macOS Sandbox, Linux namespaces)
- No daemon requirement (Apple Container is self-contained)
- Simpler installation (system-native or single binary)

## Current State

The codebase has three sandbox providers:
1. `NoneProvider` - Host execution (no isolation)
2. `AppleProvider` - Apple Container (partially implemented, needs work)
3. `DockerProvider` - Docker containers (TO BE REMOVED)

### Files Affected

**Core sandbox code:**
- `internal/sandbox/docker.go` - DELETE entirely
- `internal/sandbox/apple.go` - UPDATE to align with real Apple Container CLI
- `internal/sandbox/none.go` - KEEP as fallback
- `internal/sandbox/pool.go` - UPDATE provider selection logic
- `internal/sandbox/mounts/mounts.go` - REMOVE DockerMountArgs, keep AppleContainerMountArgs

**Tests:**
- `internal/sandbox/sandbox_test.go` - REMOVE all Docker tests
- `internal/sandbox/mounts/mounts_test.go` - REMOVE Docker mount tests

**Documentation:**
- `internal/sandbox/images/busybox.md` - REWRITE for Apple Container
- `internal/sandbox/images/Dockerfile` - DELETE (no Docker)
- `AGENTS.md` - UPDATE sandbox documentation

**Install script:**
- `install.sh` - REMOVE setup_docker(), UPDATE setup_apple_container()

**Integration:**
- `internal/integration/harness.go` - REMOVE Docker harness
- `internal/integration/harness_test.go` - REMOVE Docker tests

**Config:**
- Config references to "docker" provider type

## Implementation Phases

### Phase 1: Research Apple Container API
- Study `.read-only/container/` codebase
- Document CLI commands and expected behavior
- Identify gaps in current AppleProvider implementation

### Phase 2: Remove Docker Provider
- Delete docker.go and all Docker tests
- Remove Docker references from install.sh
- Update documentation

### Phase 3: Fix Apple Container Provider
- Align apple.go with real Apple Container CLI (from .read-only/container)
- Test on macOS 26+ with Apple Silicon
- Implement proper error handling and status checking

### Phase 4: Linux Provider Research
- Research LXC/LXD for Linux native containers
- Design LinuxProvider interface
- Document distribution-specific requirements

### Phase 5: Testing & Documentation
- Update all tests to use mock provider in CI
- Update AGENTS.md and README.md
- Create platform-specific setup guides

## Success Criteria

1. No Docker references in codebase (except .read-only/)
2. Apple Container works on macOS 26+ Apple Silicon
3. NoneProvider works on all platforms
4. Clear documentation for each platform
5. All tests pass
