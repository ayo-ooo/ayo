---
id: ayo-rx04
status: closed
deps: [ayo-rx02, ayo-rx03]
links: []
created: 2026-02-24T03:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-rx01
tags: [remediation, documentation]
---
# Task: Create docs/tutorials/ (5 tutorials)

## Summary

Create 5 hands-on tutorials in `docs/tutorials/` that guide users through common workflows. Each tutorial should be completable in 30 minutes or less.

## Files to Create

### 1. tutorials/first-agent.md (~300 lines)

**Goal**: Create and customize your first agent

```markdown
# Create Your First Agent

## What You'll Build
A custom agent specialized for code review.

## Prerequisites
- ayo installed and setup complete
- Basic familiarity with concepts.md

## Step 1: Create the Agent
ayo agent new @code-reviewer

## Step 2: Customize system.md
[Edit system prompt for code review focus]

## Step 3: Configure ayo.json
[Enable specific tools, set model]

## Step 4: Add Skills
[Add SKILL.md with code review guidelines]

## Step 5: Test Your Agent
ayo @code-reviewer "Review this function..."

## Complete Example
[Full working agent directory]

## Troubleshooting
[Common issues and solutions]
```

### 2. tutorials/squads.md (~350 lines)

**Goal**: Set up a multi-agent squad

```markdown
# Multi-Agent Coordination with Squads

## What You'll Build
A development squad with frontend and backend agents.

## Step 1: Create the Squad
ayo squad create dev-team

## Step 2: Write SQUAD.md
[Define roles and coordination rules]

## Step 3: Add Agents
[Configure agent roster in ayo.json]

## Step 4: Run a Task
ayo squad run dev-team "Build a REST API"

## Step 5: Monitor Progress
[Check tickets, view output]

## Complete Example
[Full working squad directory]
```

### 3. tutorials/triggers.md (~300 lines)

**Goal**: Set up event-driven agents

```markdown
# Event-Driven Agents with Triggers

## What You'll Build
An agent that runs on a schedule.

## Step 1: Create a Cron Trigger
ayo trigger add --cron "0 9 * * *" --agent @daily-report

## Step 2: Test the Trigger
ayo trigger fire daily-report

## Step 3: Add a File Watch Trigger
ayo trigger add --watch ./src --agent @linter

## Step 4: Monitor History
ayo trigger history

## Complete Example
[Trigger configurations]
```

### 4. tutorials/memory.md (~250 lines)

**Goal**: Use the memory system

```markdown
# Memory System Deep Dive

## What You'll Build
An agent that remembers your preferences.

## Step 1: Enable Memory
[Configure memory in ayo.json]

## Step 2: Create Memories
ayo memory add "I prefer TypeScript over JavaScript"

## Step 3: Query Memories
ayo memory search "preferences"

## Step 4: Use Memory Scopes
[Global vs agent vs path scopes]

## Step 5: Export/Backup
ayo memory export > backup.json

## Complete Example
[Memory configuration and usage]
```

### 5. tutorials/plugins.md (~300 lines)

**Goal**: Create a custom plugin

```markdown
# Creating Plugins

## What You'll Build
A plugin with a custom tool.

## Step 1: Create Plugin Directory
mkdir my-plugin && cd my-plugin

## Step 2: Write manifest.json
[Plugin manifest structure]

## Step 3: Create a Tool
[External tool definition]

## Step 4: Install the Plugin
ayo plugin install ./my-plugin

## Step 5: Use in Agents
[Enable the tool in agent config]

## Complete Example
[Full plugin directory]
```

## Acceptance Criteria

- [ ] All 5 tutorial files exist in `docs/tutorials/`
- [ ] Each tutorial completable in < 30 minutes
- [ ] All code examples tested and working
- [ ] Consistent structure across tutorials
- [ ] Clear before/after states
- [ ] Troubleshooting section in each
- [ ] Cross-references to relevant guides
