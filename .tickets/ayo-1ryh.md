---
id: ayo-1ryh
status: open
deps: []
links: []
created: 2026-02-23T22:15:11Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [removal, flows]
---
# Remove YAML flow executor

Delete internal/flows/yaml_executor.go and yaml_validate.go. Keep flow discovery and DAG inspection capabilities. Remove YAML flow execution path from daemon.

