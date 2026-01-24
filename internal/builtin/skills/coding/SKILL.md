---
name: coding
description: |
  Skill for source code creation and modification tasks.
  Check <delegate_context> to see if a coding delegate is configured.
  If yes, delegate to it. If no, handle coding tasks directly with bash.
metadata:
  author: ayo
  version: "6.0"
---

# Coding Skill

## Check Your Delegate Context First

Look at your `<delegate_context>` system message:

- **If a coding delegate is configured** (e.g., `coding: @crush`): Delegate coding tasks to that agent via `agent_call`
- **If NO coding delegate is configured**: Handle coding tasks directly using bash

## Handling Coding Tasks Directly (No Delegate)

When no coding delegate is available, you can write code yourself using bash:

### Writing Files

Use heredocs for multi-line files:
```bash
cat > src/App.js << 'EOF'
import React from 'react';

function App() {
  return <div>Hello World</div>;
}

export default App;
EOF
```

Use echo for single lines:
```bash
echo 'export default function() {}' > src/util.js
```

### Creating Projects

Use scaffolding tools when appropriate:
```bash
npx create-react-app my-app --template typescript
npm create vite@latest my-app -- --template react
go mod init myproject
```

Then modify generated files as needed.

### Best Practices for Direct Coding

1. Create directories first: `mkdir -p src/components`
2. Write complete, working code (not stubs)
3. Use appropriate file extensions
4. Follow the project's existing patterns
5. Run tests/builds after changes to verify

## Delegating Coding Tasks (With Delegate)

When a coding delegate IS configured, use `agent_call`:

```json
{
  "agent": "<coding_agent_from_delegate_context>",
  "prompt": "Create a React todo app with components for adding and displaying todos",
  "model": "your-current-model"
}
```

### Good Delegation Prompts

Structure prompts with:
1. **Clear objective**: What needs to be accomplished
2. **Scope**: Which files or directories are involved
3. **Constraints**: What should NOT be changed
4. **Success criteria**: How to verify completion

## What NOT to Delegate

Always handle directly with bash:
- Running commands (git, npm, go build, etc.)
- Questions about code (answer from knowledge)
- File operations without code content
- Installing dependencies
