# ayo share

Share host directories with sandboxed agents.

## Synopsis

```
ayo share <command> [path] [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `add` | Share a directory |
| `list` | List shared directories |
| `remove` | Remove a share |

---

## ayo share add

Share a host directory with sandboxes.

```bash
$ ayo share add <path> [flags]
```

### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--name` | string | Name for the share (default: directory name) |
| `--readonly` | bool | Mount as read-only |

### Examples

```bash
$ ayo share add ~/Code/myproject
Shared ~/Code/myproject as 'myproject'
Available in sandbox at: /workspace/myproject

$ ayo share add ~/data --name datasets --readonly
Shared ~/data as 'datasets' (read-only)
```

---

## ayo share list

List all shared directories.

```bash
$ ayo share list
```

### Example

```bash
$ ayo share list
NAME        HOST PATH            SANDBOX PATH              MODE
myproject   ~/Code/myproject     /workspace/myproject      rw
datasets    ~/data               /workspace/datasets       ro
```

---

## ayo share remove

Remove a shared directory.

```bash
$ ayo share remove <name>
```

### Example

```bash
$ ayo share remove datasets
Removed share 'datasets'
```

---

## How Sharing Works

When you share a directory:

1. The daemon creates a bind mount from host to sandbox
2. Agents can read/write files (unless read-only)
3. Changes are reflected immediately on both sides
4. Share persists across sandbox restarts

## See Also

- [ayo sandbox](cli-sandbox.md) - Sandbox management
