---
id: ayo-u3l6
status: closed
deps: [ayo-9k8m, ayo-n88v, ayo-oxj6, ayo-akqw, ayo-7rj8, ayo-158p]
links: []
created: 2026-02-24T01:02:53Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-xfu3
tags: [verification, e2e]
---
# Phase 5 E2E verification

End-to-end verification that squads work as first-class coordination primitive.

## Prerequisites

All Phase 5 tickets complete:
- Lead semantics (ayo-9k8m)
- Dispatch routing (ayo-n88v)
- I/O schema enforcement (ayo-oxj6)
- Ticket tools (ayo-akqw)

## Verification Checklist

### Squad Creation
- [ ] `ayo squad create my-squad` creates sandbox
- [ ] ayo.json with squad config loads
- [ ] SQUAD.md used for constitution
- [ ] Agents listed in config are available

### Lead Agent
- [ ] Messages to squad go to lead first
- [ ] Lead can route to other agents
- [ ] Lead receives all squad input
- [ ] `input_accepts` overrides routing if set

### Dispatch Routing
- [ ] Lead can call other squad agents
- [ ] Agents see each other's files
- [ ] Each agent runs as own Unix user
- [ ] Handoff between agents works

### I/O Schema
- [ ] Input schema validates incoming messages
- [ ] Invalid input rejected with clear error
- [ ] Output schema validates final output
- [ ] Schemas defined in ayo.json

### Ticket Coordination
- [ ] `tk` tool available to squad agents
- [ ] Tickets visible in /.tickets/
- [ ] Agents can create/update tickets
- [ ] Ticket assignment works

## Acceptance Criteria

Multi-agent squad completes a coordinated task successfully.
