---
id: ayo-erxf
status: closed
deps: []
links: ["epic:evaluation"]
created: 2026-03-07T21:04:42Z
type: task
priority: 3
assignee: Alex Cabrera
---
# Remove capabilities and guardrails if unused

EVALUATION COMPLETED:
- Capabilities package: Provides agent capability indexing and search
- Guardrails package: Provides security guardrails using sandwich pattern
- Usage: Both packages are actively used in the codebase
- Decision: KEEP both packages as they provide essential functionality
- Rationale: Capabilities useful for agent discovery, Guardrails essential for security
- Action: No removal, mark as core build system components

Evaluate and potentially remove internal/capabilities/ and internal/guardrails/ if not needed for build system

