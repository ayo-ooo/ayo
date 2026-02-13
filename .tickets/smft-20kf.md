---
id: smft-20kf
status: closed
deps: [smft-c9sx]
links: []
created: 2026-02-12T23:53:15Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-w6k3
---
# AGENTS.md: Add common commands section

Add a common commands section to AGENTS.md.

**File:** AGENTS.md
**Insert after:** Key Directories section

**Content to add:**
```markdown
## Common Commands

**Build:** \`go build ./cmd/ayo/...\`

**Test:** \`go test ./... -count=1\`

**Run locally:** \`go run ./cmd/ayo/... [args]\`

**Lint:** \`golangci-lint run\`
```

