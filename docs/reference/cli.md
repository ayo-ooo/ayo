# CLI Reference

Complete command-line reference for ayo.

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | | Output in JSON format |
| `--quiet` | `-q` | Suppress non-essential output |
| `--no-jodas` | `-y` | Auto-approve file modifications |
| `--config PATH` | | Path to config file |
| `--help` | `-h` | Show help |

## Commands Overview

| Command | Description |
|---------|-------------|
| `ayo [prompt]` | Chat with @ayo |
| `ayo @agent [prompt]` | Chat with specific agent |
| `ayo "#squad" [prompt]` | Send task to squad |
| `ayo agent` | Manage agents |
| `ayo squad` | Manage squads |
| `ayo trigger` | Manage triggers |
| `ayo memory` | Manage memories |
| `ayo session` | Manage sessions |
| `ayo service` | Control background service |
| `ayo doctor` | Check system health |
| `ayo audit` | View file modification logs |
| `ayo backup` | Manage sandbox backups |
| `ayo plugin` | Manage plugins |
| `ayo sandbox` | Manage sandboxes |
| `ayo setup` | Initial setup wizard |

---

## ayo

Interactive chat or single prompt.

```bash
ayo [prompt]
ayo @agent [prompt]
ayo "#squad" [prompt]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--attach FILE` | `-a` | Attach file to prompt |
| `--continue` | `-c` | Continue most recent session |
| `--session ID` | `-s` | Continue specific session |

### Examples

```bash
ayo "explain this code"
ayo @reviewer "review my changes"
ayo "#dev-team" "build auth feature"
ayo -a main.go "fix the bug"
ayo -c "also add tests"
```

---

## ayo agent

Manage AI agents.

### agent list

List all available agents.

```bash
ayo agent list [--json]
```

**Output**:
```
NAME          DESCRIPTION              SOURCE
@ayo          General purpose agent    builtin
@reviewer     Code review specialist   user
```

### agent create

Create a new agent.

```bash
ayo agent create @name [--template TEMPLATE]
```

**Templates**: `default`, `reviewer`, `assistant`

### agent show

Show agent details.

```bash
ayo agent show @name
```

### agent rm

Remove an agent.

```bash
ayo agent rm @name [--force]
```

### agent status

Show active agent sessions.

```bash
ayo agent status
```

### agent wake

Start an agent session.

```bash
ayo agent wake @name
```

### agent sleep

Stop an agent session.

```bash
ayo agent sleep @name
```

---

## ayo squad

Manage squads for multi-agent coordination.

### squad list

List all squads.

```bash
ayo squad list [--json]
```

### squad create

Create a new squad.

```bash
ayo squad create NAME [--agents @a,@b,@c]
```

### squad show

Show squad details.

```bash
ayo squad show NAME
```

### squad destroy

Destroy a squad and its data.

```bash
ayo squad destroy NAME [--force]
```

### squad start

Start a squad's sandbox.

```bash
ayo squad start NAME
```

### squad stop

Stop a squad's sandbox.

```bash
ayo squad stop NAME
```

### squad add-agent

Add an agent to a squad.

```bash
ayo squad add-agent SQUAD @agent
```

### squad remove-agent

Remove an agent from a squad.

```bash
ayo squad remove-agent SQUAD @agent
```

### squad shell

Open interactive shell in squad sandbox.

```bash
ayo squad shell SQUAD [@agent]
```

### squad ticket

Manage squad tickets.

```bash
ayo squad ticket SQUAD list
ayo squad ticket SQUAD create TITLE [--assignee @agent]
ayo squad ticket SQUAD show ID
ayo squad ticket SQUAD ready
```

---

## ayo trigger

Manage triggers for automated execution.

### trigger list

List all triggers.

```bash
ayo trigger list [--json]
```

### trigger create

Create a new trigger.

```bash
# Cron trigger
ayo trigger schedule NAME --cron "0 9 * * *" --agent @name --prompt "..."

# Watch trigger
ayo trigger schedule NAME --watch PATH --pattern "*.go" --agent @name --prompt "..."

# Interval trigger
ayo trigger schedule NAME --interval 30m --agent @name --prompt "..."

# One-time trigger
ayo trigger schedule NAME --once "2024-12-25 09:00" --agent @name --prompt "..."
```

**Flags**:

| Flag | Description |
|------|-------------|
| `--cron EXPR` | Cron expression |
| `--watch PATH` | Directory to watch |
| `--interval DUR` | Interval duration |
| `--once DATETIME` | One-time execution |
| `--agent @name` | Agent to invoke |
| `--prompt TEXT` | Prompt to send |
| `--pattern GLOB` | File patterns (for watch) |
| `--debounce DUR` | Debounce duration |
| `--singleton` | Prevent overlapping |

### trigger show

Show trigger details.

```bash
ayo trigger show NAME
```

### trigger enable/disable

Enable or disable a trigger.

```bash
ayo trigger enable NAME
ayo trigger disable NAME
```

### trigger remove

Remove a trigger.

```bash
ayo trigger remove NAME
```

### trigger fire

Manually execute a trigger.

```bash
ayo trigger fire NAME
```

### trigger history

View execution history.

```bash
ayo trigger history [NAME] [--limit N] [--status STATUS]
```

---

## ayo memory

Manage persistent memories.

### memory list

List memories.

```bash
ayo memory list [--category CAT] [--scope SCOPE] [--json]
```

### memory store

Store a new memory.

```bash
ayo memory store CONTENT [--category CAT] [--scope SCOPE]
```

**Categories**: `preference`, `fact`, `correction`, `pattern`
**Scopes**: `global`, `agent`, `path`, `squad`

### memory search

Search memories semantically.

```bash
ayo memory search QUERY [--limit N]
```

### memory show

Show memory details.

```bash
ayo memory show ID
```

### memory forget

Soft-delete a memory.

```bash
ayo memory forget ID
```

### memory link

Link two memories.

```bash
ayo memory link ID1 ID2
```

### memory export

Export memories to JSON.

```bash
ayo memory export [FILE] [--include-embeddings]
```

### memory import

Import memories from JSON.

```bash
ayo memory import FILE
```

### memory reindex

Rebuild search index.

```bash
ayo memory reindex
```

### memory stats

Show memory statistics.

```bash
ayo memory stats
```

---

## ayo service

Control the background service.

### service start

Start the service.

```bash
ayo service start
```

### service stop

Stop the service.

```bash
ayo service stop
```

### service status

Show service status.

```bash
ayo service status [--json]
```

### service restart

Restart the service.

```bash
ayo service restart
```

---

## ayo doctor

Check system health and dependencies.

```bash
ayo doctor [--json]
```

**Checks**:
- Daemon status
- Sandbox provider availability
- LLM connectivity
- Configuration validity
- Disk space
- Permissions

---

## ayo audit

View file modification audit logs.

### audit list

List audit entries.

```bash
ayo audit list [--agent @name] [--path PATH] [--since DURATION]
```

### audit show

Show audit entry details.

```bash
ayo audit show ID
```

---

## ayo backup

Manage backups.

### backup create

Create a backup.

```bash
ayo backup create [FILE]
```

### backup restore

Restore from backup.

```bash
ayo backup restore FILE
```

### backup list

List available backups.

```bash
ayo backup list
```

---

## ayo plugin

Manage plugins.

### plugin list

List installed plugins.

```bash
ayo plugin list [--json]
```

### plugin install

Install a plugin.

```bash
ayo plugin install NAME|PATH|URL
```

### plugin remove

Remove a plugin.

```bash
ayo plugin remove NAME
```

### plugin show

Show plugin details.

```bash
ayo plugin show NAME
```

---

## ayo sandbox

Direct sandbox management.

### sandbox shell

Open shell in sandbox.

```bash
ayo sandbox shell @ayo
```

### sandbox exec

Execute command in sandbox.

```bash
ayo sandbox exec @ayo "command"
```

### sandbox reset

Reset sandbox to clean state.

```bash
ayo sandbox reset @ayo
```

---

## ayo session

Manage conversation sessions.

### session list

List recent sessions.

```bash
ayo session list [--limit N] [--agent @name] [--json]
```

### session show

Show session details.

```bash
ayo session show SESSION_ID
```

### session resume

Resume a session.

```bash
ayo -s SESSION_ID "continue..."
```

### session delete

Delete a session.

```bash
ayo session delete SESSION_ID
```

---

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Permission denied |
| 4 | Resource not found |
| 5 | Operation timeout |

---

## Environment Variables

### Configuration

| Variable | Description |
|----------|-------------|
| `AYO_HOME` | Base directory (default: `~/.local/share/ayo`) |
| `AYO_CONFIG` | Config file path |
| `AYO_PROVIDER` | Default LLM provider |
| `AYO_MODEL` | Default model |
| `AYO_DEBUG` | Enable debug logging |
| `XDG_CONFIG_HOME` | Base config directory (default: `~/.config`) |
| `XDG_DATA_HOME` | Base data directory (default: `~/.local/share`) |

### LLM Provider API Keys

| Variable | Provider |
|----------|----------|
| `ANTHROPIC_API_KEY` | Anthropic (Claude) |
| `OPENAI_API_KEY` | OpenAI (GPT) |
| `GEMINI_API_KEY` | Google (Gemini) |
| `OPENROUTER_API_KEY` | OpenRouter |
| `AZURE_OPENAI_API_KEY` | Azure OpenAI |
| `GROQ_API_KEY` | Groq |
| `DEEPSEEK_API_KEY` | DeepSeek |
| `CEREBRAS_API_KEY` | Cerebras |
| `XAI_API_KEY` | xAI (Grok) |
| `TOGETHER_API_KEY` | Together.ai |

### Service URLs

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama endpoint |

---

## Shell Completion

### Bash

```bash
ayo completion bash > /etc/bash_completion.d/ayo
```

### Zsh

```bash
ayo completion zsh > "${fpath[1]}/_ayo"
```

### Fish

```bash
ayo completion fish > ~/.config/fish/completions/ayo.fish
```
