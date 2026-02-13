---
id: smft-x7jx
status: closed
deps: [smft-1bik]
links: []
created: 2026-02-12T23:53:47Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-w6k3
---
# README.md: Add 'What can you do with ayo?' section

Add a new section after Quick Start showing practical examples.

**File:** README.md
**Insert after:** Quick Start section

**Content to add:**
```markdown
## What Can You Do?

\`\`\`bash
# Chat interactively
ayo

# Single task
ayo "help me debug this test"

# Review a file
ayo -a main.go "review this code"

# Continue a conversation
ayo -c "what about edge cases?"

# Use a specialized agent
ayo @reviewer "check for security issues"
\`\`\`
```

