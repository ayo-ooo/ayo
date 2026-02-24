# Tools Configuration Guide

Complete reference for built-in tools and creating custom tools.

## Built-in Tools

### Shell & Files

| Tool | Description | Execution |
|------|-------------|-----------|
| `bash` | Execute shell commands | Sandbox |
| `view` | Read file contents | Sandbox |
| `edit` | Modify files | Sandbox |
| `glob` | Find files by pattern | Sandbox |
| `grep` | Search file contents | Sandbox |

### Memory

| Tool | Description | Execution |
|------|-------------|-----------|
| `memory_store` | Save a memory | Host |
| `memory_search` | Find relevant memories | Host |

### Host Interaction

| Tool | Description | Execution |
|------|-------------|-----------|
| `file_request` | Request host file access | Bridge |
| `publish` | Write to /output (no approval) | Bridge |

### Agent Interaction

| Tool | Description | Execution |
|------|-------------|-----------|
| `delegate` | Call another agent | Host |
| `human_input` | Ask user for input | Host |

### Planning

| Tool | Description | Execution |
|------|-------------|-----------|
| `todo_add` | Add task to todo list | Host |
| `todo_update` | Update task status | Host |
| `ticket_create` | Create a ticket | Host |
| `ticket_update` | Update ticket | Host |

## Tool Configuration

### Enable/Disable Tools

In agent `config.json`:

```json
{
  "allowed_tools": ["bash", "view", "edit"],
  "disabled_tools": ["delegate"]
}
```

### Default Tools

If `allowed_tools` is not specified, all tools are available.

### Tool Precedence

1. `disabled_tools` takes precedence over `allowed_tools`
2. Agent config overrides global config

## Tool Details

### bash

Execute shell commands in the sandbox.

**Parameters**:
```json
{
  "command": "string (required)",
  "working_dir": "string (optional)",
  "timeout": "string (optional, e.g., '30s')"
}
```

**Example**:
```json
{
  "command": "go build ./...",
  "working_dir": "/workspace",
  "timeout": "60s"
}
```

### view

Read file contents.

**Parameters**:
```json
{
  "path": "string (required)",
  "start_line": "integer (optional)",
  "end_line": "integer (optional)"
}
```

**Example**:
```json
{
  "path": "/workspace/main.go",
  "start_line": 10,
  "end_line": 50
}
```

### edit

Modify file contents.

**Parameters**:
```json
{
  "path": "string (required)",
  "old_string": "string (required)",
  "new_string": "string (required)"
}
```

**Example**:
```json
{
  "path": "/workspace/main.go",
  "old_string": "fmt.Println(\"hello\")",
  "new_string": "fmt.Println(\"world\")"
}
```

### glob

Find files by pattern.

**Parameters**:
```json
{
  "pattern": "string (required)",
  "root": "string (optional)"
}
```

**Example**:
```json
{
  "pattern": "**/*.go",
  "root": "/workspace"
}
```

### grep

Search file contents.

**Parameters**:
```json
{
  "pattern": "string (required)",
  "path": "string (optional)",
  "recursive": "boolean (optional)"
}
```

**Example**:
```json
{
  "pattern": "func main",
  "path": "/workspace",
  "recursive": true
}
```

### file_request

Request to write a file to the host.

**Parameters**:
```json
{
  "path": "string (required)",
  "content": "string (required)",
  "reason": "string (optional)"
}
```

**Example**:
```json
{
  "path": "~/Projects/app/main.go",
  "content": "package main...",
  "reason": "Fixed authentication bug"
}
```

**Behavior**:
1. User sees approval prompt
2. User approves/denies
3. If approved, file is written

### publish

Write to /output (syncs to host without approval).

**Parameters**:
```json
{
  "filename": "string (required)",
  "content": "string (required)"
}
```

**Example**:
```json
{
  "filename": "report.md",
  "content": "# Report\n..."
}
```

File appears at `~/.local/share/ayo/output/report.md`.

### delegate

Call another agent.

**Parameters**:
```json
{
  "agent": "string (required)",
  "prompt": "string (required)",
  "context": "object (optional)"
}
```

**Example**:
```json
{
  "agent": "@reviewer",
  "prompt": "Review this code",
  "context": {
    "file": "main.go",
    "changes": "..."
  }
}
```

### human_input

Ask the user for input.

**Parameters**:
```json
{
  "prompt": "string (required)",
  "type": "string (optional: text, choice, confirm)",
  "options": "array (for choice type)"
}
```

**Example**:
```json
{
  "prompt": "Which database should I use?",
  "type": "choice",
  "options": ["PostgreSQL", "MySQL", "SQLite"]
}
```

## Execution Contexts

| Context | Location | Tools |
|---------|----------|-------|
| `host` | Your machine | memory, delegate, human_input |
| `sandbox` | Container | bash, view, edit, glob, grep |
| `bridge` | Cross-boundary | file_request, publish |

## Creating Custom Tools

### External Tool Format

Create a directory with `tool.json` and execution script:

```
my-tool/
├── tool.json
└── run.sh
```

### tool.json Schema

```json
{
  "name": "my-tool",
  "description": "What the tool does",
  "execution": "sandbox",
  "parameters": {
    "type": "object",
    "properties": {
      "input": {
        "type": "string",
        "description": "Input parameter"
      },
      "options": {
        "type": "object",
        "description": "Optional settings"
      }
    },
    "required": ["input"]
  }
}
```

### Execution Scripts

**run.sh** (bash):
```bash
#!/bin/bash
INPUT="$1"
OPTIONS="$2"

# Tool logic here
echo "Processing: $INPUT"
```

**run.py** (Python):
```python
#!/usr/bin/env python3
import sys
import json

input_data = sys.argv[1]
options = json.loads(sys.argv[2]) if len(sys.argv) > 2 else {}

# Tool logic here
print(f"Processing: {input_data}")
```

### Installation

Place in `~/.config/ayo/tools/my-tool/` or include in a plugin.

### Using Custom Tools

Enable in agent config:

```json
{
  "allowed_tools": ["bash", "view", "my-tool"]
}
```

## Tool Permissions

### Blocked Patterns

Some paths are always blocked:

```go
DefaultBlockedPatterns = []string{
    ".git/*",
    ".env*",
    "**/secrets/*",
    "**/*.key",
    "**/*.pem",
    "**/id_rsa*",
    "**/.ssh/*",
}
```

### Auto-Approve Patterns

In agent config:

```json
{
  "permissions": {
    "auto_approve_patterns": [
      "./build/*",
      "./dist/*",
      "./tmp/*"
    ]
  }
}
```

### Trust Levels

| Level | Tools Available |
|-------|-----------------|
| `sandboxed` | Sandbox tools only |
| `privileged` | Sandbox + file_request |
| `unrestricted` | All tools, no sandbox |

## Tool Categories

Tools can be assigned to semantic categories:

| Category | Default | Description |
|----------|---------|-------------|
| `shell` | `bash` | Command execution |
| `plan` | Plugin-provided | Task planning |
| `search` | Plugin-provided | Code search |

### Overriding Categories

In config:

```json
{
  "tool_categories": {
    "shell": "bash",
    "plan": "my-planner",
    "search": "my-search"
  }
}
```

## Stateful Tools

### SQLite Storage

Stateful tools store data in:
```
~/.local/share/ayo/tools/{name}/{name}.db
```

### Creating Stateful Tools

Implement the `StatefulTool` interface:

```go
type StatefulTool interface {
    Name() string
    Initialize(db *sql.DB) error
    Execute(params map[string]interface{}) (interface{}, error)
}
```

## Troubleshooting

### Tool not found

```bash
# List available tools
ayo agent show @name | grep tools

# Check tool directory
ls ~/.config/ayo/tools/
```

### Tool execution fails

Check execution script:
```bash
# Make executable
chmod +x ~/.config/ayo/tools/my-tool/run.sh

# Test manually
cd ~/.config/ayo/tools/my-tool
./run.sh "test input"
```

### Permission denied

Check tool execution context and trust level:
- Sandbox tools run inside container
- Host tools need appropriate permissions
