# Tools

Tools give agents the ability to take actions. Each agent specifies which tools it can use in its configuration.

## Built-in Tools

| Tool | Description |
|------|-------------|
| `bash` | Execute shell commands (default) |
| `plan` | Track multi-step tasks with status updates |
| `memory` | Search, store, and manage memories |
| `agent_call` | Delegate tasks to other agents |

## Bash Tool

The `bash` tool is the default and primary tool for most agents. It executes shell commands and returns output.

### Agent Configuration

```json
{
  "allowed_tools": ["bash"]
}
```

### How It Works

When the agent runs a command:

1. Agent specifies `command` and `description`
2. UI shows spinner with description
3. Command executes in project directory
4. Output displayed in styled box
5. Success/failure shown with elapsed time

### Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `command` | Yes | Shell command to execute |
| `description` | Yes | Human-readable description for UI |
| `timeout_seconds` | No | Command timeout (default: 30s) |
| `working_dir` | No | Working directory (scoped to project) |

### Security

- Commands run in the project directory
- Dangerous commands trigger guardrail warnings
- Long-running commands timeout after 30s (configurable)

## Plan Tool

The `plan` tool enables agents to track multi-step tasks with status updates.

### Agent Configuration

```json
{
  "allowed_tools": ["bash", "plan"]
}
```

**Note:** The `planning` skill is automatically attached when the plan tool is enabled.

### Hierarchical Structure

Plans support three levels:

1. **Phases** (optional) - High-level stages
2. **Tasks** (required) - Units of work
3. **Todos** (optional) - Atomic sub-items within tasks

### Task-Only Plan

```json
{
  "tasks": [
    {
      "content": "Implement user authentication",
      "active_form": "Implementing user authentication",
      "status": "in_progress",
      "todos": [
        {
          "content": "Add login endpoint",
          "active_form": "Adding login endpoint",
          "status": "completed"
        },
        {
          "content": "Add logout endpoint",
          "active_form": "Adding logout endpoint",
          "status": "pending"
        }
      ]
    },
    {
      "content": "Write tests",
      "active_form": "Writing tests",
      "status": "pending"
    }
  ]
}
```

### Plan with Phases

```json
{
  "phases": [
    {
      "name": "Phase 1: Setup",
      "status": "completed",
      "tasks": [
        {
          "content": "Initialize project",
          "active_form": "Initializing project",
          "status": "completed"
        }
      ]
    },
    {
      "name": "Phase 2: Implementation",
      "status": "in_progress",
      "tasks": [
        {
          "content": "Build core features",
          "active_form": "Building core features",
          "status": "in_progress"
        }
      ]
    }
  ]
}
```

### Task States

| State | Description |
|-------|-------------|
| `pending` | Not yet started |
| `in_progress` | Currently working on |
| `completed` | Finished successfully |

### Rules

- Each task needs both `content` (imperative) and `active_form` (present continuous)
- Exactly ONE item should be `in_progress` at any time
- Cannot mix phases and top-level tasks
- Plans persist across session resumption

## Memory Tool

The `memory` tool allows agents to search, store, and manage user memories.

### Agent Configuration

```json
{
  "allowed_tools": ["bash", "memory"]
}
```

### Operations

| Operation | Description |
|-----------|-------------|
| `search` | Find relevant memories semantically |
| `store` | Save new information |
| `list` | Show all memories |
| `forget` | Remove a memory |

### Example Usage (by agent)

```
Agent: I'll search your memories for coding preferences.
[uses memory tool: search "coding preferences"]

Agent: I'll remember that you prefer TypeScript.
[uses memory tool: store "User prefers TypeScript for frontend development"]
```

Memory storage is asynchronous - the agent continues immediately while the memory stores in the background.

## Agent Call Tool

The `agent_call` tool enables delegation to other agents.

### Agent Configuration

```json
{
  "allowed_tools": ["bash", "agent_call"]
}
```

### Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `agent` | Yes | Agent handle (e.g., `@crush`) |
| `prompt` | Yes | Task to delegate |

### Example

```json
{
  "agent": "@crush",
  "prompt": "Refactor the authentication module to use JWT tokens"
}
```

## Search Tool (Alias)

The `search` tool is a **tool alias** that resolves to a configured concrete tool.

### Agent Configuration

```json
{
  "allowed_tools": ["bash", "search"]
}
```

### Global Configuration

In `~/.config/ayo/ayo.json`:

```json
{
  "default_tools": {
    "search": "searxng"
  }
}
```

Without a configured search provider, the search tool is not available.

## External Tools (via Plugins)

Plugins can provide additional tools that wrap external CLI commands.

### Example: Installing a Tool via Plugin

```bash
ayo plugins install https://github.com/user/ayo-plugins-searxng
```

### Tool Definition (tool.json)

Plugins define tools in `tools/<name>/tool.json`:

```json
{
  "name": "my-tool",
  "description": "What this tool does",
  "command": "my-binary",
  "args": ["--flag"],
  "parameters": [
    {
      "name": "input",
      "description": "Input text",
      "type": "string",
      "required": true
    }
  ],
  "timeout": 60,
  "working_dir": "inherit",
  "depends_on": ["my-binary"]
}
```

See [Plugins](plugins.md) for complete documentation.

## UI Feedback

When tools execute, the UI shows:

1. **Spinner** with tool description
2. **Output** in styled box (expandable if long)
3. **Status** with elapsed time

```
◐ Running test suite...
┌──────────────────────────────────────────┐
│ === RUN   TestExample                    │
│ --- PASS: TestExample (0.00s)            │
│ PASS                                     │
└──────────────────────────────────────────┘
✓ Running test suite (1.2s)
```

## Tool Timeouts

Default timeouts:

| Tool | Default Timeout |
|------|-----------------|
| `bash` | 30 seconds |
| `plan` | N/A (instant) |
| `memory` | 30 seconds |
| `agent_call` | No timeout |

Agents can override with `timeout_seconds` parameter.
