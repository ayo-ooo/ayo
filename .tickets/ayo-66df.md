---
id: ayo-66df
status: open
deps: []
links: []
created: 2026-02-23T22:15:34Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-whmn
tags: [sandbox, filesystem]
---
# Implement /output safe write zone

Create /output/{session}/ directory in sandbox that maps to ~/.local/share/ayo/output/{session}/ on host. Agents can write freely here without approval. Good for generated reports, artifacts, code output.

