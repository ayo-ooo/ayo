---
id: ayo-cmds
title: CLI command documentation accuracy
status: open
priority: critical
assignee: "@ayo"
tags: [docs, gtm, polish]
created: 2026-02-24
---

# CLI command documentation accuracy

Audit and fix all CLI command documentation to match actual implementation.

## Problem

Documentation uses inconsistent command names (singular vs plural) and some commands are documented incorrectly or missing entirely.

## Command Naming Issues

### Singular vs Plural Inconsistency

The CLI accepts both forms but docs should be consistent:

| Documented | Actual Primary | Also Works |
|------------|---------------|------------|
| `ayo agents` | `ayo agent` | `ayo agents` (alias) |
| `ayo triggers` | `ayo trigger` | `ayo triggers` (alias) |
| `ayo service` | `ayo sandbox service` | - |

**Decision**: Use **singular** forms as they are the primary commands.

### Files to Update

| File | Issue |
|------|-------|
| `README.md:57` | `ayo agents create` → `ayo agent create` |
| `README.md:144-151` | `ayo triggers schedule/watch` → `ayo trigger schedule/watch` |
| `docs/reference/cli.md:26` | `ayo service` → `ayo sandbox service` |
| `docs/reference/cli.md:70-100` | Mix of `ayo agents` and `ayo agent` |
| `docs/guides/triggers.md:63-68` | Uses `ayo trigger` (correct) |
| `docs/tutorials/triggers.md:201-215` | Uses `ayo trigger` (correct) |
| `docs/advanced/troubleshooting.md:85-93` | `ayo daemon start` → `ayo sandbox service start` |

## Missing Commands in docs/reference/cli.md

Commands that exist but aren't documented:

| Command | File | Description |
|---------|------|-------------|
| `ayo flow` | `flows.go` | Manage multi-step workflows |
| `ayo sync` | `sync.go` | Manage sandbox synchronization |
| `ayo share` | `share.go` | Share host directories with sandboxes |
| `ayo planner` | `planner.go` | Manage planner plugins |
| `ayo session` | `session.go` | Manage conversation sessions |
| `ayo skill` | `skills.go` | Manage agent skills |
| `ayo index` | `index.go` | Manage entity index |
| `ayo backup` | `backup.go` | Manage sandbox backups |
| `ayo audit` | `audit.go` | View file modification audit logs |
| `ayo notifications` | `notifications.go` | View trigger notifications |
| `ayo migrate` | `migrate.go` | Migrate configurations |
| `ayo plugin` | `plugins.go` | Manage plugins |
| `ayo ticket` | `tickets.go` | Manage task tickets |

## Missing Agent Subcommands

Agent commands not documented:

| Command | Description |
|---------|-------------|
| `ayo agent archive` | Archive an @ayo-created agent |
| `ayo agent unarchive` | Unarchive an agent |
| `ayo agent promote` | Promote @ayo agent to user-owned |
| `ayo agent refine` | Refine an agent's system prompt |
| `ayo agent capabilities` | Show agent capabilities |
| `ayo agent wake` | Start an agent session |
| `ayo agent sleep` | Stop an agent session |
| `ayo agent status` | Show active agent sessions |

## Incorrect Flag Documentation

| File | Issue |
|------|-------|
| `README.md:57` | `--description` → `-d` |

## Acceptance Criteria

- [ ] All commands use singular form consistently (`agent`, `trigger`, `squad`)
- [ ] `ayo service` references changed to `ayo sandbox service`
- [ ] All implemented commands documented in cli.md
- [ ] All agent subcommands documented
- [ ] All flags match actual implementation

## Dependencies

None

## Notes

Run `go run ./cmd/ayo/... --help` to verify command structure.
