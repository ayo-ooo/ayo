---
id: ayo-snst
status: open
deps: [ayo-q841]
links: []
created: 2026-02-23T23:15:51Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-sqad
tags: [scheduler]
---
# Implement duration/interval jobs

Add support for duration-based scheduling. Trigger config: type: interval, schedule.every: '5m'. Support units: s, m, h, d. Add singleton option to prevent overlap. Use gocron's DurationJob.

