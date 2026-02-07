---
id: ayo-0u7e
status: closed
deps: [ayo-41ms]
links: []
created: 2026-02-05T18:52:26Z
type: feature
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, identity, storage]
---
# Persistent Agent Home Directories

Agent home directories persist across sandbox restarts. Store at ~/.local/share/ayo/sandboxes/{agent}/home and mount into container. Contains agent-specific config, history, scratch space.

## Acceptance Criteria

- Home persists across sandbox restarts
- Mounted at /home/{agent} in container
- Contains .config, .cache subdirs
- Separate per-agent storage
- Cleaned up with ayo sandbox prune --homes

