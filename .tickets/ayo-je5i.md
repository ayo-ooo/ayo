---
id: ayo-je5i
status: closed
deps: [ayo-q84j]
links: []
created: 2026-02-05T18:52:55Z
type: feature
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, files, tools]
---
# Publish Tool: agent publishes files back to host

New tool allowing agents to publish files from sandbox to host. User prompted to approve with diff preview. Complements request_file for bidirectional file flow.

## Acceptance Criteria

- publish_file tool available to sandboxed agents
- User sees diff before approval
- Approved files written to host
- Supports single file or directory
- Audit log of published files

