# ayo memory

Manage semantic memory for persistent facts and preferences across sessions.

## Synopsis

```
ayo memory <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `store` | Store a fact or preference |
| `search` | Search memories |
| `list` | List all memories |
| `delete` | Delete a memory |
| `clear` | Clear all memories |

---

## ayo memory store

Store a fact or preference in memory.

```bash
$ ayo memory store <fact> [--category <cat>]
```

### Example

```bash
$ ayo memory store "I prefer TypeScript over JavaScript"
Stored memory: mem_a1b2c3

$ ayo memory store "API key is in ~/.secrets/api.key" --category config
Stored memory: mem_d4e5f6
```

---

## ayo memory search

Search memories semantically.

```bash
$ ayo memory search <query> [--limit <n>]
```

### Example

```bash
$ ayo memory search "programming preferences"
RELEVANCE  ID          FACT
0.92       mem_a1b2c3  I prefer TypeScript over JavaScript
0.78       mem_g7h8i9  Use 2-space indentation
0.71       mem_j1k2l3  Always include unit tests
```

---

## ayo memory list

List all stored memories.

```bash
$ ayo memory list [--category <cat>]
```

### Example

```bash
$ ayo memory list
ID          CATEGORY    CREATED              FACT
mem_a1b2c3  preference  2026-02-10 09:00:00  I prefer TypeScript...
mem_d4e5f6  config      2026-02-10 09:05:00  API key is in...
```

---

## ayo memory delete

Delete a specific memory.

```bash
$ ayo memory delete <id>
```

---

## ayo memory clear

Clear all memories.

```bash
$ ayo memory clear [--category <cat>] [--force]
```

### Example

```bash
$ ayo memory clear --category temp --force
Cleared 5 memories from category 'temp'
```

---

## Memory Categories

| Category | Description |
|----------|-------------|
| `preference` | User preferences |
| `fact` | General facts |
| `config` | Configuration info |
| `project` | Project-specific info |
| `custom` | User-defined |

## How Agents Use Memory

When an agent starts, relevant memories are automatically retrieved and included in context. Agents can also explicitly search memory during conversation.

## See Also

- [Memory Guide](../memory.md) - Conceptual overview
