---
id: ayo-899j
status: open
deps: [ayo-wt6w]
links: []
created: 2026-02-23T23:16:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-sqad
tags: [scheduler, cli]
---
# Add job monitoring interface

Add 'ayo trigger status' command showing: active jobs, next run times, last run status, run history. Store run history in SQLite (last 100 runs per job). Use gocron event listeners to track job execution.

