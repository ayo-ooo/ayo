---
id: ayo-e2e1
status: open
deps: []
links: []
created: 2026-02-24T14:00:00Z
type: epic
priority: 0
assignee: Alex Cabrera
tags: [gtm, documentation, testing, e2e]
---
# Epic: E2E Manual Human Testing Guide

## Summary

Create a comprehensive end-to-end manual testing guide that validates every aspect of the ayo system. This guide is designed as a sequential walkthrough that builds on itself - if any step fails, the tester returns to a clean state and restarts.

## Background

The existing `MANUAL_TESTING.md` (645 lines) provides useful command references but lacks:
- Sequential dependency between tests
- Clean state management
- Comprehensive coverage of all features
- Verification criteria at each step
- Escalating complexity progression

## New Guide: `docs/_E2E_MANUAL_HUMAN_TESTING_GUIDE.md`

The new guide will be:
1. **Sequential**: Each section depends on previous sections succeeding
2. **Comprehensive**: Covers every user-facing feature
3. **Verifiable**: Clear pass/fail criteria at each step
4. **Recoverable**: Clean state script to restart from scratch
5. **Progressive**: Starts simple, builds complexity

## Sections / Children

| Ticket | Section | Dependencies |
|--------|---------|--------------|
| ayo-e2e2 | Clean State & Prerequisites | None |
| ayo-e2e3 | Build, Installation & Setup | Clean state |
| ayo-e2e4 | Agent Management | Setup complete |
| ayo-e2e5 | Sessions & Chat | Agents working |
| ayo-e2e6 | Memory System | Chat working |
| ayo-e2e7 | Tickets & Planning | Sessions working |
| ayo-e2e8 | Squads & Coordination | Tickets working |
| ayo-e2e9 | Triggers & Scheduling | Squads working |
| ayo-e2e10 | Plugins | Core systems working |
| ayo-e2e11 | Full Orchestration Scenarios | All systems working |
| ayo-e2e12 | Error Handling & Edge Cases | All systems working |
| ayo-e2e13 | Cleanup & Final Verification | All tests passed |

## Guide Structure

```markdown
# E2E Manual Human Testing Guide

## Overview
- Purpose: Validate complete system functionality
- Approach: Sequential, builds on itself
- Recovery: If any step fails, run clean state script, restart

## Section 0: Clean State & Prerequisites
- System requirements (macOS 26+, Go 1.24+, etc.)
- Environment verification
- Clean state script (remove all ayo data)
- Provider configuration (API keys, model selection)

## Section 1: Build, Installation & Setup
- Build from source
- Run `ayo setup`
- Verify configuration files created
- Start daemon, verify status
- Run `ayo doctor`

## Section 2: Agent Management
- List default agents
- View agent details
- Create custom agent
- Configure agent (edit JSON)
- Remove agent
- Verify agent inheritance

## Section 3: Sessions & Chat
- Interactive chat (enter/exit)
- Single prompt (quoted)
- Session continuation (-c)
- File attachments (-a)
- Multi-turn context
- Session listing and management

## Section 4: Memory System
- Store facts
- Search memories
- List memories
- Memory context in conversations
- Memory deletion

## Section 5: Tickets & Planning
- Create tickets
- Ticket dependencies
- Ready/blocked queries
- Ticket workflow (start → close)
- Ticket assignment
- Priority management

## Section 6: Squads & Coordination
- Create squad
- Edit SQUAD.md constitution
- Add/remove agents
- Start/stop squad
- Squad ticket management
- Dispatch to squad (#squad-name)
- Verify workspace output
- Multi-agent coordination

## Section 7: Triggers & Scheduling
- Schedule triggers (cron)
- Watch triggers (file)
- Trigger listing
- Trigger execution verification
- Trigger removal

## Section 8: Plugins
- List plugins
- Plugin info
- Disable/enable plugins
- Plugin configuration
- Custom plugin installation

## Section 9: Full Orchestration Scenarios
- Complex multi-agent task
- End-to-end feature build
- Cross-squad coordination (via @ayo)
- Autonomous workflow completion

## Section 10: Error Handling & Edge Cases
- Invalid commands
- Daemon not running
- Unknown agent
- Network errors
- Concurrent operations
- Recovery scenarios

## Section 11: Cleanup & Final Verification
- Destroy test artifacts
- Stop daemon
- Remove test data
- Verify clean state
- Final checklist
```

## Acceptance Criteria

- [ ] Guide exists at `docs/_E2E_MANUAL_HUMAN_TESTING_GUIDE.md`
- [ ] All 12 sections written with step-by-step instructions
- [ ] Each step has clear verification criteria
- [ ] Clean state script provided and tested
- [ ] `MANUAL_TESTING.md` removed
- [ ] Guide successfully followed start-to-finish on clean macOS system
