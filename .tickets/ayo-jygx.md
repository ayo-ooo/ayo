---
id: ayo-jygx
status: open
deps: []
links: []
created: 2026-02-12T19:46:32Z
type: bug
priority: 1
assignee: Alex Cabrera
parent: ayo-7j3v
tags: [tests, cli]
---
# Fix TestAgentsCommandStructure test failure

Test expects 'agents' command with 'agent' alias but command uses 'agent' with no alias. Either fix the test or fix the command to match expected behavior.

## Acceptance Criteria

go test ./cmd/ayo/... -run TestAgentsCommandStructure passes

