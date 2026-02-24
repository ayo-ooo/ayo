---
id: ayo-rx07
status: closed
deps: [ayo-rx05, ayo-rx06]
links: []
created: 2026-02-24T03:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-rx01
tags: [remediation, documentation]
---
# Task: Create docs/advanced/ (3 documents)

## Summary

Create 3 advanced documentation files in `docs/advanced/` for developers who want deep understanding of ayo internals.

## Files to Create

### 1. advanced/architecture.md (~500 lines)

System architecture deep dive:

```markdown
# Architecture

## Overview

Ayo follows a host/sandbox architecture where:
- Host process: LLM calls, memory, orchestration
- Sandbox container: Command execution, file operations

## Component Diagram

┌─────────────────────────────────────────────────────┐
│ Host Process                                         │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐         │
│  │ CLI (ayo) │ │ Daemon    │ │ Providers │         │
│  └─────┬─────┘ └─────┬─────┘ └─────┬─────┘         │
│        │             │             │                 │
│        ▼             ▼             ▼                 │
│  ┌─────────────────────────────────────────────┐   │
│  │ Orchestration Layer                          │   │
│  │  - Agent loading                             │   │
│  │  - Memory integration                        │   │
│  │  - Trigger engine                            │   │
│  │  - Squad dispatch                            │   │
│  └───────────────────┬─────────────────────────┘   │
└──────────────────────┼──────────────────────────────┘
                       │ JSON-RPC
┌──────────────────────┼──────────────────────────────┐
│ Sandbox              │                               │
│  ┌───────────────────▼───────────────────────┐     │
│  │ ayod (in-sandbox daemon)                   │     │
│  │  - Command execution                       │     │
│  │  - File operations                         │     │
│  │  - Tool responses                          │     │
│  └───────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────┘

## Daemon Architecture

The daemon (ayo daemon) is the central coordinator:
- Manages sandbox lifecycle
- Routes agent requests
- Executes triggers
- Handles file_request approvals

### Event Loop
[Trigger engine, polling, etc.]

### RPC Protocol
[JSON-RPC over Unix socket]

## ayod Protocol

The in-sandbox daemon (ayod) provides:
- bash execution
- File read/write
- Environment setup

### Communication
[JSON-RPC protocol details]

## LLM Integration

### Provider Interface
[How providers are abstracted]

### Streaming
[How responses are streamed]

## Memory Subsystem

### SQLite Index
[Memory storage schema]

### Zettelkasten Files
[Markdown storage for human readability]

### Embedding Generation
[How embeddings are created and stored]

## Plugin System

### Loading Mechanism
[Discovery, validation, registration]

### Resolution Order
[User > Plugin > Builtin]

## Trigger Engine

### gocron v2 Integration
[Scheduler implementation]

### File Watch
[fsnotify integration]

### Persistence
[SQLite job store]
```

### 2. advanced/extending.md (~400 lines)

Guide to extending ayo:

```markdown
# Extending Ayo

## Creating Sandbox Providers

### Provider Interface
type SandboxProvider interface {
    Create(config SandboxConfig) error
    Destroy(name string) error
    Exec(name, cmd string) (string, error)
    ...
}

### Implementation Example
[Complete provider example]

### Registration
[How to register a provider]

## Custom Embedding Providers

### Interface
type EmbeddingProvider interface {
    Embed(text string) ([]float32, error)
    BatchEmbed(texts []string) ([][]float32, error)
}

### Implementation
[Example with local model]

## Custom Memory Providers

### Interface
[Memory storage interface]

### Implementation
[Example with different backend]

## Custom Planners

### Planner Interface
type PlannerPlugin interface {
    Name() string
    Type() PlannerType
    Initialize(ctx context.Context) error
    GetTools() []Tool
    ...
}

### Near-term vs Long-term
[Difference and use cases]

### Implementation
[Example planner]

## Trigger Plugins

### Interface
type TriggerPlugin interface {
    Type() TriggerType
    Start(callback TriggerCallback) error
    Stop() error
    ...
}

### Implementation
[Example trigger plugin]

## Contributing to Core

### Development Setup
git clone ...
go build ./cmd/ayo/...
go test ./...

### Code Organization
[Key directories and their purpose]

### Pull Request Process
[Guidelines for contributing]
```

### 3. advanced/troubleshooting.md (~350 lines)

Common issues and debugging:

```markdown
# Troubleshooting

## Diagnostic Tools

### ayo doctor
Comprehensive health check:
ayo doctor

Output explains:
- Daemon status
- Sandbox provider availability
- LLM connectivity
- Configuration issues

### Debug Logging
AYO_DEBUG=1 ayo ...

### Daemon Logs
tail -f ~/.local/share/ayo/daemon.log

## Common Issues

### Daemon Won't Start

**Symptom**: `ayo daemon start` hangs or errors

**Solutions**:
1. Check socket: `ls -la ~/.local/share/ayo/daemon.sock`
2. Remove stale socket: `rm ~/.local/share/ayo/daemon.sock`
3. Check port conflicts: `lsof -i :8080`
4. Check logs: `cat ~/.local/share/ayo/daemon.log`

### Sandbox Creation Fails

**Symptom**: "failed to create sandbox"

**Solutions**:
1. macOS: Ensure macOS 26+ for Apple Container
2. Linux: Check systemd-nspawn: `which systemd-nspawn`
3. Permissions: Check container runtime permissions
4. Resources: Check disk space and memory

### Agent Not Found

**Symptom**: "agent @name not found"

**Solutions**:
1. List agents: `ayo agent list`
2. Check path: `ls ~/.config/ayo/agents/`
3. Check plugin: `ayo plugin list`
4. Name case: Agent names are case-sensitive

### Memory Search Returns Nothing

**Symptom**: `ayo memory search` returns empty

**Solutions**:
1. Check memories exist: `ayo memory list`
2. Check embedding provider: `ayo doctor`
3. Rebuild index: `ayo memory reindex`

### file_request Not Working

**Symptom**: Agent can't write to host

**Solutions**:
1. Check approval: Look for terminal prompt
2. Check --no-jodas: Not set when needed
3. Check permissions: `cat ~/.config/ayo/config.json`
4. Check audit log: `ayo audit list`

### Trigger Not Firing

**Symptom**: Scheduled trigger doesn't run

**Solutions**:
1. Check daemon running: `ayo daemon status`
2. List triggers: `ayo trigger list`
3. Check cron expression: `ayo help cron`
4. Manual test: `ayo trigger fire <name>`
5. Check history: `ayo trigger history`

### Plugin Conflicts

**Symptom**: Strange behavior after installing plugin

**Solutions**:
1. List plugins: `ayo plugin list`
2. Check for duplicates: Same agent/tool in multiple plugins
3. Resolution order: User > Plugin > Builtin
4. Remove plugin: `ayo plugin remove <name>`

## Performance Tuning

### Slow First Response
[Sandbox warm-up, model loading]

### High Memory Usage
[Memory limits, cleanup]

### Trigger Latency
[Polling intervals, event batching]

## Debug Scripts

Located in `debug/`:
- `system-info.sh` - Host system information
- `sandbox-status.sh` - Container status
- `daemon-status.sh` - Service status

## Getting Help

- GitHub Issues: [link]
- Documentation: [link]
- Community: [link]
```

## Acceptance Criteria

- [ ] All 3 advanced files exist in `docs/advanced/`
- [ ] Architecture accurately described
- [ ] Extension points fully documented
- [ ] Troubleshooting covers real issues
- [ ] Useful for contributors
- [ ] Code references accurate
