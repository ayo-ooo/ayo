---
name: coding
description: |
  Skill for source code creation and modification tasks.
  Check <delegate_context> for a coding delegate - if configured, you MUST delegate ALL coding to it.
metadata:
  author: ayo
  version: "7.0"
---

# Coding Skill

## CRITICAL: Check Delegate Context First

Look at your `<delegate_context>` system message.

### If a coding delegate IS configured (e.g., `coding: @crush`):

**YOU MUST delegate ALL coding tasks to that agent.** No exceptions.

This includes:
- Writing ANY code file (even trivial ones like hello world)
- Creating new projects or applications
- Modifying or editing existing code
- Debugging and fixing code issues
- Writing or modifying tests
- Refactoring code

**DO NOT write code yourself using bash when a delegate is configured.**

Use agent_call:
```json
{
  "agent": "@crush",
  "prompt": "Create a hello world Go program in /tmp/test"
}
```

### If NO coding delegate is configured:

Handle coding tasks directly using bash:

```bash
cat > main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello, world!")
}
EOF
```

## What You Can Still Do With Bash

Even when a coding delegate is configured, use bash for:
- Running code (go run, npm start, python script.py)
- Building projects (go build, npm run build)
- Running tests (go test, npm test)
- Git operations
- Installing dependencies
- File operations that don't involve writing code

## Summary

```
Is there a coding delegate in <delegate_context>?
├── YES → Use agent_call for ALL code writing
└── NO  → Use bash with heredocs to write code
```
