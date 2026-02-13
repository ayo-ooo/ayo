---
id: asmft-frxa
status: done
deps: [asmft-61lu]
links: []
created: 2026-02-13T14:36:34Z
type: task
priority: 1
assignee: Alex Cabrera
parent: asmft-zies
tags: [phase4, rpc]
---
# Implement notify agents handler

Add handleSquadNotifyAgents to server.go. Spawns agent sessions for all agents with assigned tickets in the squad. Returns session IDs.

