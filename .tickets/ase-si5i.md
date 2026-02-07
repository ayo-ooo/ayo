---
id: ase-si5i
status: closed
deps: [ase-fb0m]
links: []
created: 2026-02-06T04:10:04Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-ka3q
---
# Lazy agent user creation system

Implement lazy creation of Unix users for agents in the sandbox. When an agent is first invoked, create their user account and home directory.

## Design

## Current State
User creation exists but is tied to sandbox creation. We need per-agent users created lazily.

## Implementation
1. Add EnsureAgentUser(ctx, sandboxID, agentHandle) to SandboxProvider
2. Before executing any command for an agent, ensure their user exists
3. Store created users in sandbox metadata

## User Creation
- adduser -D -s /bin/sh {agent}
- Creates /home/{agent}/
- Copy skeleton dotfiles if agent has sandbox/dotfiles/ directory

## Agent Dotfiles
Agents can define dotfiles in their config:
  {agent_dir}/sandbox/dotfiles/.bashrc
  {agent_dir}/sandbox/dotfiles/.profile
These are copied to /home/{agent}/ on first use.

## Tracking
Maintain list of created users in sandbox metadata to avoid redundant creation.

## Acceptance Criteria

- User created on first agent invocation
- Home directory at /home/{agent}/
- Agent dotfiles copied if present
- Subsequent invocations skip creation

