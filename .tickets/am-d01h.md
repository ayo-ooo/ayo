---
id: am-d01h
status: closed
deps: []
links: []
created: 2026-02-18T03:12:44Z
type: epic
priority: 2
assignee: Alex Cabrera
---
# @ayo Squad Lead System

Implement @ayo-in-squad as the default squad lead. Squad-scoped @ayo adopts the squad constitution and mediates internal coordination. Clear separation between @ayo-main (host orchestrator) and @ayo-in-squad (internal mediator).

## Acceptance Criteria

- @ayo-in-squad inherits squad constitution
- Squad lead intercepts all input unless designated agent specified
- @ayo-in-squad cannot reach outside squad (no cross-squad privileges)
- Escalation creates ticket for squad lead
- Clear identity scoping (no @ayo proliferation confusion)

