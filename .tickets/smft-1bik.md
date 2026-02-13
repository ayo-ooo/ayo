---
id: smft-1bik
status: closed
deps: []
links: []
created: 2026-02-12T23:53:41Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-w6k3
---
# README.md: Simplify Quick Start section

Simplify the Quick Start section to be copy-paste ready.

**File:** README.md
**Section:** Quick Start (around line 20)

**Ensure commands are:**
1. Single install command
2. Single env var export
3. Single run command

**Target format:**
```markdown
## Quick Start

\`\`\`bash
# Install
go install github.com/alexcabrera/ayo/cmd/ayo@latest

# Set API key
export ANTHROPIC_API_KEY="sk-..."

# Start chatting
ayo
\`\`\`
```

**Remove:** The 'With file attachment' example from quick start (move to Examples section).

