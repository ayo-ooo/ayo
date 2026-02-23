---
id: ayo-ieiy
status: open
deps: []
links: []
created: 2026-02-23T22:15:11Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [removal, daemon]
---
# Remove webhook server from daemon

Delete internal/daemon/webhook_server.go and remove webhook startup/handling from server.go. Keep trigger engine for cron and file watch.

