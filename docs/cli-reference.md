# CLI Reference

Complete command reference for ayo.

## Root Command

```bash
ayo [command] [@agent] [prompt] [--flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--attachment` | `-a` | File attachments (repeatable) |
| `--config` | | Path to config file |
| `--debug` | | Show debug output including raw tool payloads |
| `--model` | `-m` | Model to use (overrides config default) |
| `--help` | `-h` | Help for ayo |
| `--version` | `-v` | Show version |

### Examples

```bash
# Interactive chat with default agent
ayo

# Interactive chat with specific agent
ayo @ayo

# Single prompt
ayo "explain this error"

# With file attachment
ayo -a main.go "review this code"

# Multiple attachments
ayo -a file1.txt -a file2.txt "compare these"
```

---

## ayo agents

Manage agents.

### ayo agents list

List all available agents.

```bash
ayo agents list
```

### ayo agents show

Show agent details.

```bash
ayo agents show <handle>
```

### ayo agents create

Create a new agent.

```bash
ayo agents create @handle [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--model` | `-m` | Model to use |
| `--description` | `-d` | Brief description |
| `--system` | `-s` | System prompt text |
| `--system-file` | `-f` | Path to system prompt file |
| `--tools` | `-t` | Allowed tools (comma-separated) |
| `--skills` | | Skills to include |
| `--exclude-skills` | | Skills to exclude |
| `--ignore-builtin-skills` | | Don't load built-in skills |
| `--ignore-shared-skills` | | Don't load user shared skills |
| `--input-schema` | | JSON schema for stdin input |
| `--output-schema` | | JSON schema for stdout output |
| `--no-guardrails` | | Disable safety guardrails |

**Examples:**

```bash
# Show help (no arguments)
ayo agents create

# Minimal agent
ayo agents create @helper -m gpt-5.2

# With system file
ayo agents create @reviewer -m gpt-5.2 -f system.md

# Full options
ayo agents create @debugger \
  -m gpt-5.2 \
  -d "Debugging specialist" \
  -f system.md \
  -t bash,agent_call \
  --skills debugging
```

**Conversational alternative:**

```bash
ayo "help me create an agent for code review"
```

### ayo agents update

Update built-in agents.

```bash
ayo agents update [--force]
```

---

## ayo skills

Manage skills.

### ayo skills list

List all available skills.

```bash
ayo skills list
```

### ayo skills show

Show skill details.

```bash
ayo skills show <name>
```

### ayo skills create

Create a new skill.

```bash
ayo skills create <name> [--flags]
```

| Flag | Description |
|------|-------------|
| `--shared` | Create in shared skills directory |

### ayo skills validate

Validate a skill directory.

```bash
ayo skills validate <path>
```

### ayo skills update

Update built-in skills.

```bash
ayo skills update [--force]
```

---

## ayo sessions

Manage conversation sessions.

### ayo sessions list

List conversation sessions.

```bash
ayo sessions list [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--agent` | `-a` | Filter by agent handle |
| `--source` | `-s` | Filter by source (ayo, crush, crush-via-ayo) |
| `--limit` | `-n` | Maximum results (default 20) |

### ayo sessions show

Show session details and conversation.

```bash
ayo sessions show <session-id>
```

### ayo sessions continue

Continue a previous session.

```bash
ayo sessions continue [session-id]
```

Without ID:
- Shows interactive picker by default
- Use `--latest` for headless automation

| Flag | Short | Description |
|------|-------|-------------|
| `--latest` | `-l` | Continue most recent session without prompting |
| `--debug` | | Show debug output |

### ayo sessions delete

Delete a session.

```bash
ayo sessions delete <session-id> [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Delete without confirmation |

---

## ayo memory

Manage agent memories.

### ayo memory list

List memories.

```bash
ayo memory list [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--agent` | `-a` | Filter by agent |
| `--category` | `-c` | Filter by category |
| `--limit` | `-n` | Maximum results (default 50) |
| `--json` | | JSON output |

### ayo memory search

Search memories semantically.

```bash
ayo memory search <query> [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--agent` | `-a` | Filter by agent |
| `--threshold` | `-t` | Similarity threshold (0-1, default 0.3) |
| `--limit` | `-n` | Maximum results (default 10) |
| `--json` | | JSON output |

### ayo memory show

Show memory details.

```bash
ayo memory show <id>
```

### ayo memory store

Store a new memory.

```bash
ayo memory store <content> [--flags]
```

| Flag | Description |
|------|-------------|
| `-c`, `--category` | Category: preference, fact, correction, pattern (auto-detected if not specified) |
| `-a`, `--agent` | Agent handle for scoping the memory |
| `-p`, `--path` | Path scope for this memory |
| `--json` | Output in JSON format |

### ayo memory forget

Forget a memory (soft delete).

```bash
ayo memory forget <id> [--force]
```

### ayo memory stats

Show memory statistics.

```bash
ayo memory stats
```

### ayo memory clear

Clear all memories.

```bash
ayo memory clear [--flags]
```

| Flag | Description |
|------|-------------|
| `--agent` | Clear for specific agent only |
| `--force` | Skip confirmation |

---

## ayo plugins

Manage plugins.

### ayo plugins install

Install a plugin from git.

```bash
ayo plugins install <git-url> [--flags]
```

| Flag | Description |
|------|-------------|
| `--force` | Reinstall/overwrite |
| `--local` | Install from local directory |
| `--yes` | Skip prompts |

**Examples:**

```bash
# From GitHub
ayo plugins install https://github.com/user/ayo-plugins-name

# Local development
ayo plugins install --local ./my-plugin
```

### ayo plugins list

List installed plugins.

```bash
ayo plugins list
```

### ayo plugins show

Show plugin details.

```bash
ayo plugins show <name>
```

### ayo plugins update

Update plugins.

```bash
ayo plugins update [name] [--flags]
```

| Flag | Description |
|------|-------------|
| `--force` | Force update |
| `--dry-run` | Check without applying |

### ayo plugins remove

Remove a plugin.

```bash
ayo plugins remove <name> [--yes]
```

---

## ayo flows

Manage flows - composable agent pipelines.

### ayo flows list

List all available flows.

```bash
ayo flows list
```

### ayo flows show

Show flow details.

```bash
ayo flows show <name>
```

### ayo flows new

Create a new flow.

```bash
ayo flows new <name> [--flags]
```

| Flag | Description |
|------|-------------|
| `--project` | Create in project directory (.ayo/flows/) |
| `--with-schemas` | Create with input/output schemas |
| `--force` | Overwrite if exists |

### ayo flows run

Execute a flow.

```bash
ayo flows run <name> [input] [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--input` | `-i` | Input file path |
| `--timeout` | `-t` | Timeout in seconds (default 300) |
| `--validate` | | Validate input only, don't run |
| `--no-history` | | Don't record run in history |

**Input sources:**
- Argument: `ayo flows run myflow '{"key": "value"}'`
- Stdin: `echo '{"key": "value"}' | ayo flows run myflow`
- File: `ayo flows run myflow -i data.json`

### ayo flows validate

Validate a flow file or directory.

```bash
ayo flows validate <path>
```

### ayo flows history

Show flow run history.

```bash
ayo flows history [--flow=<name>]
```

### ayo flows replay

Replay a flow run with its original input.

```bash
ayo flows replay <run-id>
```

---

## ayo chain

Explore and validate agent chaining.

### ayo chain ls

List chainable agents.

```bash
ayo chain ls [--json]
```

### ayo chain inspect

Show agent schemas.

```bash
ayo chain inspect <agent> [--json]
```

### ayo chain from

Find agents that can receive this agent's output.

```bash
ayo chain from <agent>
```

### ayo chain to

Find agents whose output this agent can receive.

```bash
ayo chain to <agent>
```

### ayo chain validate

Validate JSON against agent's input schema.

```bash
ayo chain validate <agent> [json]
```

JSON can be provided as argument or via stdin.

### ayo chain example

Generate example input for an agent.

```bash
ayo chain example <agent>
```

---

## ayo sandbox

Manage sandboxed execution environments.

### ayo sandbox service

Manage the sandbox background service.

```bash
ayo sandbox service start [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--foreground` | `-f` | Run in foreground for debugging |

```bash
ayo sandbox service stop
ayo sandbox service status
```

### ayo sandbox list

List active sandboxes.

```bash
ayo sandbox list
```

Output includes sandbox ID, name, status, and age.

### ayo sandbox show

Show sandbox details.

```bash
ayo sandbox show [--id <id>]
```

Without `--id`:
- With 1 sandbox: auto-selects it
- With multiple: shows interactive picker

### ayo sandbox exec

Execute command in a sandbox.

```bash
ayo sandbox exec [--id <id>] [--user <user>] [--workdir <dir>] <command> [args...]
```

Flags must come before the command. After the first non-flag argument, everything is passed to the command.

| Flag | Short | Description |
|------|-------|-------------|
| `--id` | | Sandbox ID (uses picker if not specified) |
| `--user` | `-u` | Run as specified user |
| `--workdir` | `-w` | Working directory inside container |

**Examples:**

```bash
ayo sandbox exec ls -la
ayo sandbox exec --user ayo whoami
ayo sandbox exec --id abc123 cat /etc/os-release
ayo sandbox exec sh -c "echo hello > /tmp/test.txt"
```

### ayo sandbox login

Open interactive shell in sandbox.

```bash
ayo sandbox login [--id <id>] [--as <agent>]
```

| Flag | Description |
|------|-------------|
| `--id` | Sandbox ID (uses picker if not specified) |
| `--as` | Login as agent user (e.g., `--as @ayo`) |

### ayo sandbox shell

Open line-mode shell in sandbox (for non-TTY environments).

```bash
ayo sandbox shell [--id <id>] [--as <agent>]
```

### ayo sandbox push

Copy file or directory to sandbox.

```bash
ayo sandbox push <local-path> <container-path> [--id <id>]
```

### ayo sandbox pull

Copy file or directory from sandbox.

```bash
ayo sandbox pull <container-path> <local-path> [--id <id>]
```

### ayo sandbox diff

Show differences between sandbox and host.

```bash
ayo sandbox diff <sandbox-path> <host-path> [--id <id>]
```

### ayo sandbox sync

Sync changes from sandbox back to host.

```bash
ayo sandbox sync <sandbox-path> <host-path> [--id <id>]
```

| Flag | Description |
|------|-------------|
| `--id` | Sandbox ID (uses picker if not specified) |
| `--dry-run` | Preview changes without applying |

### ayo sandbox start

Start a stopped sandbox.

```bash
ayo sandbox start [--id <id>]
```

### ayo sandbox stop

Stop a running sandbox.

```bash
ayo sandbox stop [--id <id>] [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--id` | | Sandbox ID (uses picker if not specified) |
| `--force` | `-f` | Force kill immediately |
| `--timeout` | `-t` | Seconds to wait before force kill |

### ayo sandbox prune

Remove stopped sandboxes.

```bash
ayo sandbox prune [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation prompt |
| `--all` | `-a` | Also stop and remove running sandboxes |
| `--homes` | | Also remove persistent agent home directories |

### ayo sandbox logs

View sandbox container logs.

```bash
ayo sandbox logs [--id <id>] [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--id` | | Sandbox ID (uses picker if not specified) |
| `--follow` | `-f` | Follow log output |
| `--tail` | `-n` | Number of lines to show from end |

### ayo sandbox stats

Show resource usage statistics.

```bash
ayo sandbox stats [--id <id>]
```

### ayo sandbox join

Add an agent to an existing sandbox.

```bash
ayo sandbox join <agent> [--id <id>]
```

### ayo sandbox users

List agents in a sandbox.

```bash
ayo sandbox users [--id <id>]
```

---

## ayo mount

Manage persistent filesystem access for sandboxed agents.

Grants persist across sessions and allow agents to access host filesystem paths. Project-level mounts (`.ayo.json`) and session mounts (`--mount` flag) can only restrict access to paths already granted here—they cannot grant new access.

### ayo mount add

Grant filesystem access to a path.

```bash
ayo mount add <path> [--flags]
```

| Flag | Description |
|------|-------------|
| `--ro` | Grant read-only access (default is read-write) |
| `--json` | Output in JSON format |

Aliases: `ayo mount grant`

**Features:**
- Paths are resolved to absolute paths
- Supports `~/` home directory expansion
- Warns if path doesn't exist (but still grants for future use)

**Examples:**

```bash
ayo mount add .                  # Current directory, read-write
ayo mount add ~/Documents --ro   # Read-only access to Documents
ayo mount add /tmp/project       # Absolute path
```

**Output:**
```
✓ Granted readwrite access to /Users/you/project
```

### ayo mount list

List all filesystem grants.

```bash
ayo mount list [--json]
```

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |

Aliases: `ayo mount ls`

**Output:**
```
PATH              MODE       GRANTED
/Users/you/proj   readwrite  2026-02-07
/Users/you/docs   readonly   2026-02-06
```

**JSON Output:**
```json
[
  {"path": "/Users/you/proj", "mode": "readwrite", "granted_at": "2026-02-07T10:30:00Z"},
  {"path": "/Users/you/docs", "mode": "readonly", "granted_at": "2026-02-06T14:00:00Z"}
]
```

### ayo mount rm

Remove filesystem access.

```bash
ayo mount rm <path> [--flags]
ayo mount rm --all
```

| Flag | Description |
|------|-------------|
| `--all` | Remove all grants |
| `--json` | Output in JSON format |

Aliases: `ayo mount revoke`

**Examples:**

```bash
ayo mount rm ~/Documents         # Remove specific grant
ayo mount rm --all               # Remove all grants
```

**Output:**
```
✓ Revoked access to /Users/you/Documents
✓ Revoked 3 grant(s)
```

### Mount Hierarchy

Mounts work in a three-tier hierarchy:

1. **Global grants** (`ayo mount add`) - Defines maximum accessible paths
2. **Project mounts** (`.ayo.json` `mounts` array) - Restricts to project-relevant paths
3. **Session mounts** (`--mount` flag) - Further restricts for specific sessions

Each tier can only narrow access, never expand it. A project cannot access paths not granted globally.

---

## ayo triggers

Manage automated triggers (cron schedules and file watchers).

### ayo triggers list

List all triggers.

```bash
ayo triggers list
```

### ayo triggers add

Add a new trigger.

```bash
ayo triggers add --type <type> --agent <agent> --prompt <prompt> [--flags]
```

| Flag | Description |
|------|-------------|
| `--type` | Trigger type: `cron` or `watch` |
| `--agent` | Agent to invoke (e.g., `@ayo`) |
| `--prompt` | Prompt to send to agent |
| `--schedule` | Cron schedule (for cron type) |
| `--path` | Path to watch (for watch type) |
| `--patterns` | File patterns to match (for watch type) |

**Examples:**

```bash
# Cron trigger every 5 minutes
ayo triggers add --type cron --agent @ayo --schedule "*/5 * * * *" --prompt "Check status"

# Watch trigger for text files
ayo triggers add --type watch --agent @ayo --path /tmp --patterns "*.txt" --prompt "File changed"
```

### ayo triggers show

Show trigger details.

```bash
ayo triggers show <id>
```

### ayo triggers test

Manually fire a trigger.

```bash
ayo triggers test <id>
```

### ayo triggers enable

Enable a disabled trigger.

```bash
ayo triggers enable <id>
```

### ayo triggers disable

Disable a trigger.

```bash
ayo triggers disable <id>
```

### ayo triggers rm

Remove a trigger.

```bash
ayo triggers rm <id>
```

---

## ayo status

Show system and service status.

```bash
ayo status
```

Displays:
- CLI version
- Service status (running/not running)
- Service PID and uptime
- Sandbox pool statistics

---

## ayo setup

Complete ayo setup.

```bash
ayo setup [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Overwrite modifications without prompting |

This command:
- Installs built-in agents and skills
- Creates user directories
- Checks Ollama models

---

## ayo doctor

Check system health and dependencies.

```bash
ayo doctor [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--verbose` | `-v` | Show detailed output including model list |

Checks:
- Ayo version
- Config file
- Database connection
- Ollama service and models
- Default model configuration

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `OPENROUTER_API_KEY` | OpenRouter API key |
| `GOOGLE_API_KEY` | Google AI API key |
| `OLLAMA_HOST` | Ollama server URL (default: localhost:11434) |

---

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 130 | Interrupted (Ctrl+C) |
