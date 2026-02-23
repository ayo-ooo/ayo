---
id: ayo-la11
status: open
deps: [ayo-7dui]
links: []
created: 2026-02-23T23:13:19Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-pv3a
tags: [agents, config]
---
# Implement ayo.json loader for agents

Update agent loading to read ayo.json instead of config.json. Support both during migration period with deprecation warning for config.json. Update internal/agent/agent.go loadAgentConfig().

