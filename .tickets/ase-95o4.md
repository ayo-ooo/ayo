---
id: ase-95o4
status: closed
deps: []
links: []
created: 2026-02-06T04:08:27Z
type: epic
priority: 0
assignee: Alex Cabrera
---
# Sandbox-First Architecture

Transform ayo to run all agent execution in a persistent Alpine sandbox. Agents are Unix users with home directories, communicate via IRC, and work on mounted host files. The daemon manages sandbox lifecycle, triggers, and autonomous agent execution.

## Design

## Core Concepts

1. **Persistent Sandbox**: One Alpine container that lives forever, agents accumulate state
2. **Agent-as-User**: Each agent is a Unix user with /home/{agent}
3. **IRC for IPC**: ngircd runs in sandbox, agents communicate via channels
4. **Host-Side Intelligence**: LLM calls, memory, orchestration stay on host
5. **Daemon-Driven**: ayo daemon manages sandbox, triggers, and background agents
6. **Git Sync**: Sandbox state syncs to git for backup and multi-machine support
7. **Mount Permissions**: Explicit grants for host filesystem access

## Key Changes
- Sandbox enabled by default for all agents
- Daemon required for full functionality
- IRC server in sandbox base image
- New trigger system for autonomous execution
- Memory queries bridged from sandbox to host

## Acceptance Criteria

- All agents execute in sandbox by default
- Agents have persistent home directories
- IRC server enables inter-agent communication
- Daemon manages sandbox lifecycle and triggers
- Git sync enables backup and multi-machine use
- Mount system provides controlled host access
- CLI works for both humans and agents

