---
id: ase-fked
status: closed
deps: []
links: []
created: 2026-02-09T03:03:32Z
type: epic
priority: 2
assignee: Alex Cabrera
parent: ase-zlew
---
# Trust Levels and Guardrails

Implement trust levels for agents and harden guardrails against jailbreaking. Unrestricted agents are invisible to @ayo and can only be invoked directly.

## Acceptance Criteria

- Three trust levels: sandboxed, privileged, unrestricted
- Guardrails use PREFIX/SUFFIX sandwich pattern
- Unrestricted agents invisible to @ayo capability discovery
- Plugin scanner detects adversarial prompts
- Clear display of trust levels in CLI

