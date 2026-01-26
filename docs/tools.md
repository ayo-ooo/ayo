# Tools

Tools give agents the ability to take actions. Each agent specifies which tools it can use in its configuration.

## Built-in Tools

| Tool | Description |
|------|-------------|
| `bash` | Execute shell commands (default) |
| `todo` | Track multi-step tasks with status updates |
| `memory` | Search, store, and manage memories |
| `agent_call` | Delegate tasks to other agents |

## Tool Categories

Ayo supports **tool categories** - semantic slots that can be filled by different tool implementations. This allows users to swap tools without modifying agent configurations.

| Category | Default | Description |
|----------|---------|-------------|
| `planning` | `todo` | Task tracking during execution |
| `shell` | `bash` | Command execution |
| `search` | (none) | Web search (requires plugin) |

**Resolution order:**
1. Check `default_tools` in config for user override
2. Use built-in default if category has one
3. Fall back to literal tool name

**Configuration:**
```json
// ~/.config/ayo/ayo.json
{
  "default_tools": {
    "search": "searxng"     // Set default for category with no built-in
  }
}
```

**Agent config:**
```json
{
  "allowed_tools": ["bash", "planning"]  // "planning" resolves to "todo"
}
```

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

## Todo Tool

The `todo` tool enables agents to track multi-step tasks with status updates. It's the default implementation for the `planning` category.

### Agent Configuration

```json
{
  "allowed_tools": ["bash", "todo"]
}
```

Or use the category:

```json
{
  "allowed_tools": ["bash", "planning"]
}
```

### Parameters

```json
{
  "todos": [
    {
      "content": "What needs to be done (imperative form)",
      "active_form": "Present continuous form (e.g., 'Running tests')",
      "status": "pending | in_progress | completed"
    }
  ]
}
```

### Example

```json
{
  "todos": [
    {
      "content": "Implement user authentication",
      "active_form": "Implementing user authentication",
      "status": "completed"
    },
    {
      "content": "Write tests for auth module",
      "active_form": "Writing tests for auth module",
      "status": "in_progress"
    },
    {
      "content": "Add documentation",
      "active_form": "Adding documentation",
      "status": "pending"
    }
  ]
}
```

### Todo States

| State | Description |
|-------|-------------|
| `pending` | Not yet started |
| `in_progress` | Currently working on |
| `completed` | Finished successfully |

### Rules

- Each todo needs both `content` (imperative) and `active_form` (present continuous)
- Exactly ONE todo should be `in_progress` at any time
- The full todo list is provided on each call (replacement, not incremental)
- Todos persist within the session

### Storage

Todo data is stored in a dedicated SQLite database at `~/.local/share/ayo/tools/todo/todo.db`, keyed by session ID.

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

## Search Tool (Category)

The `search` category resolves to a configured concrete tool.

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
┌──────────────────────────────────────────────────┐
│ === RUN   TestExample                            │
│ --- PASS: TestExample (0.00s)                    │
│ PASS                                             │
└──────────────────────────────────────────────────┘
✓ Running test suite (1.2s)
```

## Tool Timeouts

Default timeouts:

| Tool | Default Timeout |
|------|-----------------|
| `bash` | 30 seconds |
| `todo` | N/A (instant) |
| `memory` | 30 seconds |
| `agent_call` | No timeout |

Agents can override with `timeout_seconds` parameter.
