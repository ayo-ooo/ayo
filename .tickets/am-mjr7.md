---
id: am-mjr7
status: closed
deps: []
links: []
created: 2026-02-02T02:56:14Z
type: task
priority: 1
assignee: Alex Cabrera
---
# Consolidate StreamHandler and StreamWriter interfaces

The run package has duplicate interface patterns:
- StreamHandler (callback-based) in handler.go
- StreamWriter (push-based) in stream_writer.go
- PrintStreamHandler and PrintWriter are near-identical implementations

StreamHandler is marked deprecated. Consolidate to StreamWriter only and remove the deprecated handler pattern.

