---
id: ayo-z91f
status: closed
deps: [ayo-12vf]
links: []
created: 2026-02-05T18:53:01Z
type: task
priority: 3
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, monitoring]
---
# Sandbox Stats and Resource Monitoring

Show sandbox resource usage: CPU, memory, disk, network. Uses container stats command.

## Acceptance Criteria

- ayo sandbox stats <id> shows resources
- Live updating with --watch
- Shows CPU %, memory usage, disk I/O
- Works with Apple Container stats command

