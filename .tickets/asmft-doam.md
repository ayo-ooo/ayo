---
id: asmft-doam
status: done
deps: [asmft-350z]
links: []
created: 2026-02-13T14:36:34Z
type: task
priority: 1
assignee: Alex Cabrera
parent: asmft-zies
tags: [phase4, notification]
---
# Add ticket completion notification

Modify ticket watcher to detect ticket status changes to 'closed'. Add OnTicketClosed callback that notifies @ayo via RPC or file-based signal.

