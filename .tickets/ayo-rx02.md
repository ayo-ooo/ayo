---
id: ayo-rx02
status: closed
deps: []
links: []
created: 2026-02-24T03:00:00Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ayo-rx01
tags: [remediation, documentation]
---
# Task: Create docs/getting-started.md

## Summary

Create the primary onboarding document that takes a new user from zero to running their first agent.

## Requirements

Target: ~200 lines, completable in < 5 minutes

## Structure

```markdown
# Getting Started

## Prerequisites
- macOS 26+ (Apple Container) or Linux (systemd-nspawn)
- Go 1.22+ (for building)
- LLM API key (Anthropic, OpenAI, or Vertex AI)

## Installation

### macOS (Homebrew)
brew install ayo/tap/ayo

### Linux (Binary)
curl -sSL https://... | sh

### From Source
git clone https://github.com/alexcabrera/ayo
cd ayo && go install ./cmd/ayo/...

## Setup

Run the setup wizard:
ayo setup

This will:
- Configure your LLM provider
- Create the default sandbox
- Start the daemon

## Your First Prompt

ayo "Hello, what can you do?"

## Interactive Mode

ayo

## Creating an Agent

ayo agent new @my-helper

## Next Steps

- [Core Concepts](concepts.md)
- [Create Your First Agent](tutorials/first-agent.md)
- [Multi-Agent Squads](tutorials/squads.md)
```

## Acceptance Criteria

- [ ] File exists at `docs/getting-started.md`
- [ ] Installation commands work on macOS and Linux
- [ ] Setup wizard completes successfully
- [ ] First prompt executes correctly
- [ ] All linked pages exist
