# Sessions

Sessions persist conversation history to a local SQLite database, allowing you to continue previous conversations.

## Storage

Sessions are stored in `~/.local/share/ayo/ayo.db`.

## Session Lifecycle

1. **Created** when chat starts (interactive or single prompt)
2. **Messages** persisted as the conversation progresses
3. **Session ID** displayed after each interaction
4. **Can be resumed** with `ayo sessions continue`

## Commands

### List Sessions

```bash
# List recent sessions
ayo sessions list

# Filter by agent
ayo sessions list --agent @ayo

# Limit results
ayo sessions list --limit 20
```

Output:

```
Sessions
------------------------------------------------------------

4443df27  @ayo
  research the latest developments...  (5 msgs)  2 hours ago

0377340f  @research
  Minnesota Recent News: Political...  (3 msgs)  2 hours ago

2 sessions
```

### Show Session

```bash
# Show session details and conversation
ayo sessions show 4443df27
```

Accepts full ID or prefix.

Output:

```
Session Details
────────────────────────────────────────────────────────────

ID: 4443df27-0b01-4103-a543-77342123d532
Agent: @ayo
Title: research the latest developments in minnesota
Messages: 5
Created: Jan 24, 2026 10:09 PM
Updated: Jan 24, 2026 10:15 PM

Conversation
────────────────────────────────────────────────────────────

> research the latest developments in minnesota

→ @ayo

I'll search for recent news about Minnesota...
```

### Continue Session

```bash
# Interactive picker for recent sessions
ayo sessions continue

# Continue specific session by ID prefix
ayo sessions continue 4443df27

# Search by title
ayo sessions continue "debugging"
```

When continued:
- Previous conversation history is loaded
- New messages append to the session
- Plans are restored if present

### Delete Session

```bash
# With confirmation prompt
ayo sessions delete 4443df27

# Skip confirmation
ayo sessions delete 4443df27 --force
```

## Session Sources

Sessions track where the conversation originated:

| Source | Description |
|--------|-------------|
| `ayo` | Direct ayo interaction |
| `crush` | Crush tool invocation |
| `crush-via-ayo` | Crush called through ayo delegation |

Filter by source:

```bash
ayo sessions list --source crush-via-ayo
```

## Session Data

Each session stores:

| Field | Description |
|-------|-------------|
| `id` | Unique UUID |
| `agent_handle` | Agent used (e.g., `@ayo`) |
| `title` | Auto-generated from first prompt |
| `created_at` | When session started |
| `updated_at` | Last activity |
| `source` | Origin (ayo, crush, etc.) |
| `plan` | Current plan state (JSON) |

## Messages

Each message in a session includes:

| Field | Description |
|-------|-------------|
| `role` | `user`, `assistant`, or `system` |
| `content` | Message text |
| `created_at` | Timestamp |

## Plans in Sessions

When an agent uses the `plan` tool, the plan is stored on the session:

```bash
# View session to see plan
ayo sessions show abc123

# Continue session - plan is restored
ayo sessions continue abc123
```

The plan persists across:
- Session continuation
- Agent interruption (Ctrl+C)
- Tool failures

## Common Workflows

### Continue Recent Work

```bash
# See what you were working on
ayo sessions list --limit 5

# Pick up where you left off
ayo sessions continue
```

### Review Past Conversations

```bash
# Find session about a topic
ayo sessions list | grep "authentication"

# View the full conversation
ayo sessions show abc123
```

### Clean Up Old Sessions

```bash
# List old sessions
ayo sessions list --limit 100

# Delete specific ones
ayo sessions delete old-id-1 --force
ayo sessions delete old-id-2 --force
```

## Tips

- Sessions are created automatically - no explicit "save" needed
- Use ID prefixes (first 8 chars) for convenience
- The `--debug` flag shows additional session metadata
- Session continuation preserves conversation context for the LLM
