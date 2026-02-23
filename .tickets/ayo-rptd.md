---
id: ayo-rptd
status: open
deps: [ayo-q841]
links: []
created: 2026-02-23T23:15:47Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-sqad
tags: [scheduler]
---
# Implement one-time job support

Add support for one-time scheduled jobs using gocron's OneTimeJob. Trigger config: type: once, schedule.at: '2026-02-24T14:00:00Z'. Job is removed from DB after execution. Add 'ayo trigger once' CLI subcommand.

