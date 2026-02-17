# ayo sandbox

Manage agent sandboxes—isolated containers for command execution.

## Synopsis

```
ayo sandbox <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List active sandboxes |
| `exec` | Execute command in sandbox |
| `shell` | Open interactive shell |
| `stop` | Stop a sandbox |
| `service` | Manage the sandbox daemon |
| `users` | List sandbox users |

---

## ayo sandbox list

List all active sandboxes.

```bash
$ ayo sandbox list
ID              STATUS    AGENT       AGE
sbx_a1b2c3d4    running   @backend    2h
sbx_e5f6g7h8    running   @frontend   1h
```

---

## ayo sandbox exec

Execute a command in a sandbox.

```bash
$ ayo sandbox exec <command>
$ ayo sandbox exec --id <sandbox-id> <command>
```

### Example

```bash
$ ayo sandbox exec "ls -la"
$ ayo sandbox exec --id sbx_a1b2c3d4 "cat /workspace/main.go"
```

---

## ayo sandbox shell

Open an interactive shell in a sandbox.

```bash
$ ayo sandbox shell [--id <sandbox-id>]
```

---

## ayo sandbox stop

Stop a running sandbox.

```bash
$ ayo sandbox stop <id>
```

---

## ayo sandbox service

Manage the background daemon that provides sandbox services.

### Commands

```bash
ayo sandbox service start [-f]   # Start daemon (foreground with -f)
ayo sandbox service stop         # Stop daemon
ayo sandbox service status       # Check status
ayo sandbox service restart      # Restart daemon
```

### Example

```bash
$ ayo sandbox service status
Daemon:     running
PID:        12345
Socket:     /Users/user/.local/share/ayo/daemon.sock
Uptime:     2h 15m
Sandboxes:  3 active
```

---

## Sandbox Providers

| Provider | Platform | Description |
|----------|----------|-------------|
| Apple Container | macOS 26+ | Native macOS containers |
| systemd-nspawn | Linux | Lightweight Linux containers |

Check available providers:

```bash
$ ayo doctor
...
Sandbox: Apple Container (darwin)
...
```

## See Also

- [ayo share](cli-share.md) - Share directories with sandboxes
- [ayo squad](cli-squad.md) - Team sandboxes
