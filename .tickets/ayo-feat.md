---
id: ayo-feat
title: Document missing CLI features
status: closed
priority: medium
assignee: "@ayo"
tags: [docs, gtm, polish]
created: 2026-02-24
---

# Document missing CLI features

Many implemented CLI features are not documented.

## Problem

The CLI reference only covers basic commands. Many features exist but aren't documented.

## Missing Command Documentation

### Top-Level Commands

| Command | File | Description | Priority |
|---------|------|-------------|----------|
| `ayo flow` | `flows.go` | Multi-step workflows | High |
| `ayo session` | `session.go` | Session management | High |
| `ayo skill` | `skills.go` | Skill management | High |
| `ayo ticket` | `tickets.go` | Task tickets | High |
| `ayo sync` | `sync.go` | Sandbox sync | Medium |
| `ayo share` | `share.go` | Host directory sharing | Medium |
| `ayo planner` | `planner.go` | Planner plugins | Medium |
| `ayo plugin` | `plugins.go` | Plugin management | Medium |
| `ayo backup` | `backup.go` | Sandbox backups | Low |
| `ayo audit` | `audit.go` | File modification logs | Low |
| `ayo notifications` | `notifications.go` | Trigger notifications | Low |
| `ayo index` | `index.go` | Entity index | Low |
| `ayo migrate` | `migrate.go` | Config migration | Low |

### Agent Subcommands

| Command | Description | Priority |
|---------|-------------|----------|
| `ayo agent wake` | Start agent session | High |
| `ayo agent sleep` | Stop agent session | High |
| `ayo agent status` | Show active sessions | High |
| `ayo agent archive` | Archive @ayo agent | Medium |
| `ayo agent unarchive` | Unarchive agent | Medium |
| `ayo agent promote` | Promote to user-owned | Medium |
| `ayo agent refine` | Refine system prompt | Medium |
| `ayo agent capabilities` | Show capabilities | Low |

### Trigger Subcommands

| Command | Description | Priority |
|---------|-------------|----------|
| `ayo trigger history` | Show run history | Medium |
| `ayo trigger test` | Fire manually | Medium |
| `ayo trigger types` | List trigger types | Low |

### Sandbox Subcommands

| Command | Description | Priority |
|---------|-------------|----------|
| `ayo sandbox login` | Interactive shell | High |
| `ayo sandbox exec` | Run command | High |
| `ayo sandbox push/pull` | File transfer | High |
| `ayo sandbox diff` | Compare sandbox/host | Medium |
| `ayo sandbox stats` | Resource usage | Low |
| `ayo sandbox users` | List agents in sandbox | Low |

## Documentation Structure Proposal

### Option A: Expand cli.md

Add sections for each command family.

### Option B: Separate Reference Pages

```
docs/reference/
├── cli.md              (overview)
├── cli-agent.md        (agent commands)
├── cli-squad.md        (squad commands)
├── cli-trigger.md      (trigger commands)
├── cli-sandbox.md      (sandbox commands)
├── cli-flow.md         (flow commands)
├── cli-session.md      (session commands)
└── cli-tools.md        (utility commands)
```

## Acceptance Criteria

- [ ] All top-level commands documented
- [ ] All agent subcommands documented
- [ ] Flow system documented
- [ ] Session management documented
- [ ] Sandbox commands documented
- [ ] Each command has examples

## Dependencies

- ayo-cmds (command accuracy first)

## Notes

Start with high priority commands that users will need most.
