# ayo ticket

Ticket-based coordination for multi-agent workflows. Tickets are markdown files with YAML frontmatter that provide persistent, auditable task tracking.

## Synopsis

```
ayo ticket <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List tickets |
| `create` | Create a new ticket |
| `show` | Show ticket details |
| `start` | Set ticket to in_progress |
| `close` | Mark ticket as done |
| `reopen` | Reopen a closed ticket |
| `block` | Mark ticket as blocked |
| `assign` | Assign ticket to an agent |
| `note` | Add a note to a ticket |
| `ready` | List tickets ready to work on |
| `blocked` | List blocked tickets |
| `dep` | Manage dependencies |
| `delete` | Delete a ticket |

---

## ayo ticket list

List tickets in the current session.

### Synopsis

```
ayo ticket list [flags]
```

### Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--session` | `-s` | string | Session ID |
| `--status` | | string | Filter: open, in_progress, blocked, closed |
| `--assignee` | `-a` | string | Filter by assignee (e.g., @coder) |
| `--type` | `-t` | string | Filter: epic, feature, task, bug, chore |

### Examples

```bash
$ ayo ticket list
ID          STATUS       TITLE                    ASSIGNEE
smft-a1b2   open         Implement auth           @backend
smft-c3d4   in_progress  Design UI                @frontend
smft-e5f6   blocked      Integration tests        @qa
```

```bash
$ ayo ticket list --status open --assignee @backend
ID          STATUS  TITLE             ASSIGNEE
smft-a1b2   open    Implement auth    @backend
```

### JSON Output

```bash
$ ayo ticket list --json
```

```json
{
  "tickets": [
    {
      "id": "smft-a1b2",
      "status": "open",
      "title": "Implement auth",
      "assignee": "@backend",
      "priority": 2,
      "deps": [],
      "created": "2026-02-12T10:30:00Z"
    }
  ]
}
```

---

## ayo ticket create

Create a new ticket.

### Synopsis

```
ayo ticket create <title> [flags]
```

### Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--assignee` | `-a` | string | Assign to agent |
| `--priority` | `-p` | int | Priority 0-4 (0=highest, default 2) |
| `--type` | `-t` | string | Type: epic, feature, task, bug, chore |
| `--deps` | | strings | Dependency ticket IDs |
| `--parent` | | string | Parent ticket ID |
| `--tags` | | strings | Tags |

### Examples

```bash
$ ayo ticket create "Implement user authentication"
Created ticket: smft-a1b2
```

```bash
$ ayo ticket create "Test auth flow" -a @qa --deps smft-a1b2 -p 1
Created ticket: smft-c3d4
```

```bash
$ ayo ticket create "Security audit" -t chore --tags security,compliance
Created ticket: smft-e5f6
```

### JSON Output

```json
{
  "id": "smft-a1b2",
  "path": "/Users/user/.local/share/ayo/sessions/ses_x/tickets/smft-a1b2.md"
}
```

---

## ayo ticket show

Show ticket details.

### Synopsis

```
ayo ticket show <id>
```

### Example

```bash
$ ayo ticket show smft-a1b2
# Implement user authentication

ID:        smft-a1b2
Status:    in_progress
Type:      task
Priority:  P2
Assignee:  @backend
Created:   2026-02-12T10:30:00Z
Started:   2026-02-12T11:00:00Z

## Description

Add JWT-based authentication to the API server.

## Notes

### 2026-02-12T11:30:00Z
Started working on auth middleware.
```

### JSON Output

```json
{
  "id": "smft-a1b2",
  "status": "in_progress",
  "type": "task",
  "title": "Implement user authentication",
  "description": "Add JWT-based authentication to the API server.",
  "priority": 2,
  "assignee": "@backend",
  "deps": [],
  "tags": [],
  "created": "2026-02-12T10:30:00Z",
  "started": "2026-02-12T11:00:00Z",
  "notes": [
    {
      "timestamp": "2026-02-12T11:30:00Z",
      "content": "Started working on auth middleware."
    }
  ]
}
```

---

## ayo ticket start

Set a ticket to in_progress status.

### Synopsis

```
ayo ticket start <id>
```

### Example

```bash
$ ayo ticket start smft-a1b2
Ticket smft-a1b2 status: in_progress
```

---

## ayo ticket close

Mark a ticket as completed.

### Synopsis

```
ayo ticket close <id>
```

### Example

```bash
$ ayo ticket close smft-a1b2
Ticket smft-a1b2 status: closed
```

---

## ayo ticket reopen

Reopen a closed ticket.

### Synopsis

```
ayo ticket reopen <id>
```

### Example

```bash
$ ayo ticket reopen smft-a1b2
Ticket smft-a1b2 status: open
```

---

## ayo ticket block

Mark a ticket as blocked.

### Synopsis

```
ayo ticket block <id>
```

### Example

```bash
$ ayo ticket block smft-a1b2
Ticket smft-a1b2 status: blocked
```

---

## ayo ticket assign

Assign a ticket to an agent.

### Synopsis

```
ayo ticket assign <id> <agent>
```

### Example

```bash
$ ayo ticket assign smft-a1b2 @backend
Ticket smft-a1b2 assigned to @backend
```

---

## ayo ticket note

Add a timestamped note to a ticket.

### Synopsis

```
ayo ticket note <id> <content>
```

### Example

```bash
$ ayo ticket note smft-a1b2 "Completed middleware, working on token generation"
Note added to smft-a1b2
```

---

## ayo ticket ready

List tickets that are ready to work on (all dependencies resolved).

### Synopsis

```
ayo ticket ready [flags]
```

### Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--assignee` | `-a` | string | Filter by assignee |

### Example

```bash
$ ayo ticket ready -a @backend
ID          PRIORITY  TITLE                    
smft-a1b2   P2        Implement auth           
smft-g7h8   P3        Refactor database layer  
```

---

## ayo ticket blocked

List tickets blocked on dependencies.

### Synopsis

```
ayo ticket blocked [flags]
```

### Example

```bash
$ ayo ticket blocked
ID          BLOCKED_ON  TITLE
smft-e5f6   smft-a1b2   Integration tests
```

---

## ayo ticket dep

Manage ticket dependencies.

### Commands

| Command | Description |
|---------|-------------|
| `add <id> <dep-id>` | Add a dependency |
| `remove <id> <dep-id>` | Remove a dependency |
| `tree <id>` | Show dependency tree |

### Examples

```bash
$ ayo ticket dep add smft-e5f6 smft-a1b2
Dependency added: smft-e5f6 depends on smft-a1b2

$ ayo ticket dep tree smft-e5f6
smft-e5f6: Integration tests
├── smft-a1b2: Implement auth [closed]
└── smft-c3d4: Design UI [in_progress]
```

---

## ayo ticket delete

Delete a ticket.

### Synopsis

```
ayo ticket delete <id>
```

### Example

```bash
$ ayo ticket delete smft-a1b2
Deleted ticket smft-a1b2
```

---

## Ticket File Format

Tickets are stored as markdown files with YAML frontmatter:

```markdown
---
id: smft-a1b2
status: open
type: task
priority: 2
assignee: "@backend"
deps: []
tags: [auth, backend]
created: 2026-02-12T10:30:00Z
---
# Implement user authentication

Add JWT-based authentication to the API server.

## Notes

### 2026-02-12T11:30:00Z
Started working on auth middleware.
```

## See Also

- [Tickets Guide](../tickets.md) - Conceptual overview
- [ayo squad](cli-squad.md) - Team sandboxes
