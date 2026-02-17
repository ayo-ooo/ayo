# ayo session

Manage conversation sessions for persistence and resumption.

## Synopsis

```
ayo session <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List sessions |
| `show` | Show session details |
| `resume` | Resume a session |
| `delete` | Delete a session |
| `export` | Export session to file |

---

## ayo session list

List all sessions.

```bash
$ ayo session list [--limit <n>] [--agent <name>]
```

### Example

```bash
$ ayo session list --limit 5
ID            AGENT      STARTED              MESSAGES
ses_x7k9m2p1  @ayo       2026-02-12 10:30:00  12
ses_a1b2c3d4  @reviewer  2026-02-12 09:15:00  8
ses_e5f6g7h8  @writer    2026-02-11 16:45:00  24
```

---

## ayo session show

Show session details and conversation history.

```bash
$ ayo session show <id> [--messages <n>]
```

### Example

```bash
$ ayo session show ses_x7k9m2p1 --messages 3
Session:   ses_x7k9m2p1
Agent:     @ayo
Started:   2026-02-12T10:30:00Z
Messages:  12

--- Conversation (last 3) ---

[You] Can you help with authentication?
[@ayo] I'd be happy to help with authentication. What approach...
[You] JWT tokens please
```

---

## ayo session resume

Resume an existing session (alias for `ayo -s <id>`).

```bash
$ ayo session resume <id>
```

---

## ayo session delete

Delete a session.

```bash
$ ayo session delete <id> [--force]
```

---

## ayo session export

Export session to a file.

```bash
$ ayo session export <id> [--format <fmt>] [--output <path>]
```

### Formats

- `json` - Full session data
- `markdown` - Human-readable transcript
- `jsonl` - Line-delimited JSON

### Example

```bash
$ ayo session export ses_x7k9m2p1 --format markdown -o conversation.md
Exported to conversation.md
```

## See Also

- [Sessions Guide](../sessions.md) - Conceptual overview
