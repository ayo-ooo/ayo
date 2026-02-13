---
id: asmft-34yd
status: done
deps: [asmft-doam]
links: []
created: 2026-02-13T14:36:34Z
type: task
priority: 1
assignee: Alex Cabrera
parent: asmft-zies
tags: [phase4, rpc]
---
# Add squads.wait_completion RPC

Add to protocol.go: squads.wait_completion method with timeout. @ayo calls this to block until all squad tickets are closed or timeout.

