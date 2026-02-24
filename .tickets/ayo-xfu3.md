---
id: ayo-xfu3
status: closed
deps: [ayo-sqad]
links: []
created: 2026-02-23T22:14:55Z
type: epic
priority: 0
assignee: Alex Cabrera
tags: [gtm, phase5]
---
# Phase 5: Squad Polish

Make squads a first-class coordination primitive.

## Goals

- Clarify squad lead semantics (who routes, who decides)
- Implement squad dispatch routing (messages go to right agent)
- Add I/O schema enforcement (validate input/output contracts)
- Polish ticket tools for squad agents
- Add squad agent shell access for debugging

## Key Decisions

1. **Lead as router**: Squad lead receives all incoming messages, decides routing
2. **Schema enforcement**: Optional but encouraged for production squads
3. **Ticket visibility**: All squad agents can see/modify tickets

## Child Tickets

### Squad Coordination
- `ayo-9k8m`: Clarify squad lead agent semantics
- `ayo-n88v`: Implement squad dispatch routing
- `ayo-oxj6`: Add I/O schema enforcement for squads
- `ayo-akqw`: Polish ticket tools for squad agents
- `ayo-158p`: Add squad agent shell access

### @ayo Routing
- `ayo-rout`: Implement @ayo smart routing (agent vs squad selection)

