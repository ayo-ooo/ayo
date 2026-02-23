---
id: ayo-y0x1
status: open
deps: [ayo-q841]
links: []
created: 2026-02-23T23:15:56Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-sqad
tags: [scheduler]
---
# Implement daily/weekly/monthly jobs

Add support for friendly schedule syntax. Daily: times: ['09:00'], days: [monday, tuesday]. Weekly/Monthly variants. Use gocron's fluent API (DailyJob, WeeklyJob, MonthlyJob). Parse YAML config into gocron job definitions.

