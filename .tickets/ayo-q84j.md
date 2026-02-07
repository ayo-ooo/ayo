---
id: ayo-q84j
status: open
deps: [ayo-1rw2]
links: []
created: 2026-02-05T18:52:50Z
type: feature
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, files, tools]
---
# File Request Tool: agent requests files from host

New tool allowing agents to request files from host filesystem. User prompted to approve. Creates secure escape hatch for agents to access host files on demand.

## Acceptance Criteria

- request_file tool available to sandboxed agents
- User prompted with file path for approval
- Approved files copied into sandbox workspace
- Denied requests return clear error to agent
- Audit log of file requests

