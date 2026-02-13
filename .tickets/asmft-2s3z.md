---
id: asmft-2s3z
status: closed
deps: [asmft-socf]
links: []
created: 2026-02-13T14:35:41Z
type: task
priority: 1
assignee: Alex Cabrera
parent: asmft-ew9b
tags: [phase1, run]
---
# Route @ayo execution to dedicated sandbox

Modify internal/run/run.go to detect when running @ayo agent and route execution to the dedicated @ayo sandbox instead of pool. Add isAyoAgent() check.

