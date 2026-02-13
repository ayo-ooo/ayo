---
id: smft-ega3
status: closed
deps: [smft-20kf]
links: []
created: 2026-02-12T23:53:21Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-w6k3
---
# AGENTS.md: Add code conventions section

Add a code conventions section to AGENTS.md.

**File:** AGENTS.md
**Insert after:** Common Commands section

**Content to add:**
```markdown
## Code Conventions

- Use \`globalOutput\` from root.go for JSON/quiet flag support
- All CLI commands inherit \`--json\` and \`--quiet\` flags
- Sandbox operations use \`providers.SandboxProvider\` interface
- Daemon communication: JSON-RPC over Unix socket
- Background tasks use trigger engine, not goroutines
- This project does NOT use Docker
```

