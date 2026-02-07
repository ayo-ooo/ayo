---
id: ayo-41ms
status: closed
deps: [ayo-15ou]
links: []
created: 2026-02-05T18:52:20Z
type: feature
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, identity]
---
# Add Agent Identity System - user accounts in sandbox

Create agent user accounts inside sandbox containers. Each agent gets a dedicated user (e.g., 'ayo' with UID 1000) with home directory. Agent config specifies sandbox user identity.

## Acceptance Criteria

- Agents run as dedicated user, not root
- User created at sandbox startup
- Home directory at /home/{agent}
- Agent config has sandbox.user field
- Backward compatible (defaults to root if not set)

