---
id: smft-c9sx
status: closed
deps: [smft-vx4b]
links: []
created: 2026-02-12T23:53:10Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-w6k3
---
# AGENTS.md: Add key file locations table

Add a key file locations section to AGENTS.md.

**File:** AGENTS.md
**Insert after:** Documentation Map section

**Content to add:**
```markdown
## Key Directories

| Path | Purpose |
|------|---------|
| cmd/ayo/ | CLI entry points |
| internal/agent/ | Agent loading, config, identity |
| internal/sandbox/ | Container management |
| internal/daemon/ | Background service |
| internal/providers/ | LLM API integrations |
| internal/memory/ | Persistent knowledge |
| internal/flows/ | Workflow execution |
| internal/share/ | Host directory sharing |
| internal/tools/ | Tool implementations |
| internal/ui/ | TUI components |
```

