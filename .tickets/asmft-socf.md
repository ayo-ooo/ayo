---
id: asmft-socf
status: closed
deps: [asmft-v2t3]
links: []
created: 2026-02-13T14:35:41Z
type: task
priority: 1
assignee: Alex Cabrera
parent: asmft-ew9b
tags: [phase1, sandbox]
---
# Create EnsureAyoSandbox in sandbox package

Add EnsureAyoSandbox() function to internal/sandbox/ that creates the dedicated @ayo sandbox with persistent /home/ayo/ and /output/ mount. Similar to existing AppleProvider.Create but with specific ayo config.

