---
id: asmft-8nej
status: done
deps: [asmft-hq0a]
links: []
created: 2026-02-13T14:36:19Z
type: task
priority: 1
assignee: Alex Cabrera
parent: asmft-y8gw
tags: [phase3, tickets]
---
# Modify ticket service for squad paths

Update internal/tickets/service.go to support squad-based ticket directories. Add SquadTicketsDir(squad) path resolution. Tickets in squads go to /.tickets/ inside squad sandbox.

