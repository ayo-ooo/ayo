# ayo squad

Manage agent squads—isolated team sandboxes where multiple agents collaborate.

## Synopsis

```
ayo squad <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `create` | Create a new squad |
| `list` | List all squads |
| `show` | Show squad details |
| `start` | Start a squad's sandbox |
| `stop` | Stop a squad's sandbox |
| `delete` | Delete a squad |
| `exec` | Execute command in squad sandbox |

---

## ayo squad create

Create a new squad with a SQUAD.md constitution.

### Synopsis

```
ayo squad create <name> [flags]
```

### Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--agents` | `-a` | strings | Initial agents to add |
| `--template` | `-t` | string | Template to use |

### Example

```bash
$ ayo squad create dev-team -a @backend,@frontend,@qa
Created squad: dev-team
Path: /Users/user/.local/share/ayo/sandboxes/squads/dev-team
Edit SQUAD.md to define team constitution.

$ $EDITOR ~/.local/share/ayo/sandboxes/squads/dev-team/SQUAD.md
```

### JSON Output

```json
{
  "name": "dev-team",
  "path": "/Users/user/.local/share/ayo/sandboxes/squads/dev-team",
  "agents": ["@backend", "@frontend", "@qa"]
}
```

---

## ayo squad list

List all squads.

### Synopsis

```
ayo squad list [flags]
```

### Example

```bash
$ ayo squad list
NAME        STATUS    AGENTS              CREATED
dev-team    running   @backend,@frontend  2026-02-10
qa-team     stopped   @qa,@tester         2026-02-08
```

---

## ayo squad show

Show detailed squad information.

### Synopsis

```
ayo squad show <name>
```

### Example

```bash
$ ayo squad show dev-team
Name:      dev-team
Status:    running
Created:   2026-02-10T09:00:00Z

Agents:
  - @backend
  - @frontend
  - @qa

Paths:
  SQUAD.md:    /Users/user/.local/share/ayo/sandboxes/squads/dev-team/SQUAD.md
  Tickets:     /Users/user/.local/share/ayo/sandboxes/squads/dev-team/.tickets/
  Workspace:   /Users/user/.local/share/ayo/sandboxes/squads/dev-team/workspace/

Active Tickets:
  smft-a1b2   in_progress   Implement auth      @backend
  smft-c3d4   open          Design UI           @frontend
```

---

## ayo squad start

Start a squad's sandbox.

### Synopsis

```
ayo squad start <name>
```

### Example

```bash
$ ayo squad start dev-team
Starting squad dev-team...
Sandbox started: sbx_x7k9m2p1
```

---

## ayo squad stop

Stop a squad's sandbox.

### Synopsis

```
ayo squad stop <name>
```

### Example

```bash
$ ayo squad stop dev-team
Stopping squad dev-team...
Sandbox stopped
```

---

## ayo squad delete

Delete a squad and its sandbox.

### Synopsis

```
ayo squad delete <name> [flags]
```

### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--force` | bool | Skip confirmation |

### Example

```bash
$ ayo squad delete old-team
Delete squad old-team? This will remove all data. [y/N] y
Deleted squad old-team
```

---

## ayo squad exec

Execute a command in a squad's sandbox.

### Synopsis

```
ayo squad exec <name> <command>
```

### Example

```bash
$ ayo squad exec dev-team "ls workspace/"
main.go
README.md
go.mod
```

---

## Squad Directory Structure

```
~/.local/share/ayo/sandboxes/squads/dev-team/
├── SQUAD.md          # Team constitution (injected into all agents)
├── .tickets/         # Coordination tickets
│   ├── smft-a1b2.md
│   └── smft-c3d4.md
├── workspace/        # Shared code workspace
└── agent-homes/      # Per-agent directories
    ├── @backend/
    ├── @frontend/
    └── @qa/
```

## SQUAD.md Constitution

The SQUAD.md file defines the team's mission, roles, and coordination rules. It's automatically injected into every agent's system prompt:

```markdown
# Development Team

## Mission
Build and ship a high-quality authentication system.

## Roles

### @backend
- API implementation
- Database design
- Security

### @frontend
- UI components
- User experience
- Accessibility

### @qa
- Test planning
- Integration testing
- Bug reporting

## Coordination

1. Use tickets for all work items
2. Add dependencies between related tasks
3. Close tickets when work is verified
4. Note blockers immediately
```

## See Also

- [Squads Guide](../squads.md) - Conceptual overview
- [ayo ticket](cli-ticket.md) - Task coordination
