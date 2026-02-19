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
ayo agents list [--flags]
```

| Flag | Description |
|------|-------------|
| `--trust` | Filter by trust level: sandboxed, privileged, unrestricted |
| `--type` | Filter by type: user, builtin, created |

**Output columns:**
- HANDLE - Agent handle (with @ prefix)
- DESCRIPTION - Brief description
- TRUST - Trust level (color-coded: green=sandboxed, yellow=privileged, red=unrestricted)
- TYPE - Source: user, builtin, or created (by @ayo)

**Examples:**

```bash
ayo agents list                    # List all agents
ayo agents list --trust sandboxed  # Only sandboxed agents
ayo agents list --type user        # Only user-created agents
```

### ayo agents show

Show agent details including trust level and metadata.

```bash
ayo agents show <handle>
```

Displays:
- Handle, model, and description
- Trust level with explanation
- Tools and skills configuration
- Input/output schemas if defined
- Creation metadata (for @ayo-created agents)

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

### ayo agents rm

Remove a user agent.

```bash
ayo agents rm <handle> [--force]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation prompt |

### ayo agents refine

Refine an agent's configuration through conversation.

```bash
ayo agents refine <handle>
```

Opens an interactive session with @ayo to improve the agent's system prompt, tools, or skills based on feedback.

### ayo agents promote

Promote an @ayo-created agent to a permanent user agent.

```bash
ayo agents promote <handle>
```

Moves the agent from the temporary created-agents directory to the user agents directory.

### ayo agents archive

Archive an agent (soft delete).

```bash
ayo agents archive <handle>
```

Marks the agent as archived. It won't appear in listings but can be restored.

### ayo agents unarchive

Restore an archived agent.

```bash
ayo agents unarchive <handle>
```

### ayo agents capabilities

Show or search agent capabilities.

```bash
ayo agents capabilities <handle> [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--search` | `-s` | Search query for capability matching |
| `--json` | | Output in JSON format |

#### ayo agents capabilities refresh

Regenerate capability embeddings.

```bash
ayo agents capabilities refresh <handle>
```

### ayo agents status

Show agent status (running sessions, wake state).

```bash
ayo agents status [handle]
```

Without a handle, shows status of all agents.

### ayo agents wake

Wake an agent (start background session).

```bash
ayo agents wake <handle>
```

### ayo agents sleep

Put an agent to sleep (end background session).

```bash
ayo agents sleep <handle>
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

Flows come in two types:
- **Shell flows** (`.sh`): Bash scripts with JSON I/O
- **YAML flows** (`.yaml`): Declarative multi-step workflows with dependencies and parallel execution

### ayo flows list

List all available flows.

```bash
ayo flows list [--json]
```

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |

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
| `--yaml` | Create a YAML flow instead of shell flow |
| `--force` | Overwrite if exists |

### ayo flows run

Execute a flow.

```bash
ayo flows run <name> [input] [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--input` | `-i` | Input file path |
| `--param` | `-p` | Set parameter (key=value, repeatable) |
| `--timeout` | `-t` | Timeout in seconds (default 300) |
| `--validate` | | Validate input only, don't run |
| `--no-history` | | Don't record run in history |

**Input sources:**
- Argument: `ayo flows run myflow '{"key": "value"}'`
- Stdin: `echo '{"key": "value"}' | ayo flows run myflow`
- File: `ayo flows run myflow -i data.json`
- Parameters: `ayo flows run myflow --param key=value`

### ayo flows validate

Validate a flow file or directory.

```bash
ayo flows validate <path>
```

Checks:
- YAML/shell syntax
- Step dependencies form a valid DAG
- Template variable references are valid
- Required parameters are defined

### ayo flows history

Show flow run history.

```bash
ayo flows history [--flags]
```

| Flag | Description |
|------|-------------|
| `--flow` | Filter by flow name |
| `--status` | Filter by status (success, failed) |
| `--limit` | Maximum results (default 20) |
| `--json` | Output in JSON format |

### ayo flows stats

Show flow execution statistics.

```bash
ayo flows stats [flow-name]
```

Displays:
- Total runs
- Success/failure rate
- Average duration
- Last run time

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

## ayo squad

Manage squads - isolated sandboxes where multiple agents collaborate. Each squad has a `SQUAD.md` constitution that defines the team's mission, roles, and coordination rules.

### ayo squad create

Create a new squad.

```bash
ayo squad create <name> [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--description` | `-d` | Squad description |
| `--agents` | `-a` | Comma-separated agent handles |
| `--workspace` | `-w` | Host path to mount as workspace |
| `--output` | `-o` | Host path for output sync |
| `--image` | `-i` | Container image |
| `--ephemeral` | `-e` | Destroy sandbox after session |

Creates the squad with a default `SQUAD.md` template in:
```
~/.local/share/ayo/sandboxes/squads/{name}/SQUAD.md
```

**Examples:**

```bash
# Basic squad
ayo squad create alpha

# With agents
ayo squad create feature -a @backend,@frontend,@qa

# Full options
ayo squad create my-team \
  -d "Auth implementation team" \
  -a @backend,@qa \
  -w ~/Code/project \
  -o /tmp/output
```

### ayo squad list

List all squads.

```bash
ayo squad list [--flags]
```

| Flag | Description |
|------|-------------|
| `--json` | JSON output |

### ayo squad show

Show squad details.

```bash
ayo squad show <name>
```

Displays:
- Squad configuration
- Agent list
- SQUAD.md content summary
- Sandbox status

### ayo squad start

Start a squad's sandbox.

```bash
ayo squad start <name>
```

### ayo squad stop

Stop a squad's sandbox.

```bash
ayo squad stop <name>
```

### ayo squad destroy

Destroy a squad.

```bash
ayo squad destroy <name> [--flags]
```

| Flag | Description |
|------|-------------|
| `--delete-data` | Also delete workspace and context |

### ayo squad add-agent

Add an agent to a squad.

```bash
ayo squad add-agent <squad> <@agent>
```

### ayo squad remove-agent

Remove an agent from a squad.

```bash
ayo squad remove-agent <squad> <@agent>
```

### ayo squad ticket

Manage tickets within a squad.

```bash
ayo squad ticket <squad> create "title" [--flags]
ayo squad ticket <squad> list
ayo squad ticket <squad> show <id>
ayo squad ticket <squad> start <id>
ayo squad ticket <squad> close <id>
```

Tickets are stored in `~/.local/share/ayo/sandboxes/squads/{squad}/.tickets/`.

### SQUAD.md Constitution

Each squad has a `SQUAD.md` file that serves as the team constitution. Edit it to define:

```markdown
# Squad: {name}

## Mission
{What the team is trying to accomplish}

## Context
{Technical constraints, dependencies, deadlines}

## Agents
### @backend
**Role**: Implementation
**Responsibilities**:
- Implement API endpoints
- Write database migrations

### @qa
**Role**: Quality assurance
**Responsibilities**:
- Review code changes
- Write test cases

## Coordination
{How agents hand off work, dependency rules}

## Guidelines
{Team-specific rules, coding style, review process}
```

When agents start in the squad, this constitution is injected into their system prompt so all team members share the same context.

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

## ayo ticket

Ticket-based coordination for multi-agent workflows. See [Tickets](tickets.md) for full documentation.

Tickets are markdown files with YAML frontmatter that provide persistent, auditable task tracking between agents.

### ayo ticket list

List tickets in a session.

```bash
ayo ticket list -s <session> [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--session` | `-s` | Session ID (required) |
| `--status` | | Filter by status: pending, in_progress, blocked, closed |
| `--assignee` | `-a` | Filter by assignee (e.g., @coder) |
| `--type` | `-t` | Filter by type: task, bug, feature, subtask |
| `--json` | | Output in JSON format |

**Examples:**

```bash
ayo ticket list -s my-session
ayo ticket list -s my-session --status in_progress
ayo ticket list -s my-session -a @coder
```

### ayo ticket create

Create a new ticket.

```bash
ayo ticket create <title> -s <session> [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--session` | `-s` | Session ID (required) |
| `--assignee` | `-a` | Assign to agent |
| `--priority` | `-p` | Priority 0-3 (0=critical, 2=default) |
| `--type` | `-t` | Type: task, bug, feature, subtask |
| `--deps` | | Comma-separated dependency ticket IDs |
| `--parent` | | Parent ticket ID (for subtasks) |
| `--ref` | | External reference (e.g., github:org/repo#123) |
| `--json` | | Output in JSON format |

**Examples:**

```bash
ayo ticket create "Implement auth" -s my-session
ayo ticket create "Fix login bug" -s my-session -a @debugger -p 1 --type bug
ayo ticket create "Deploy" -s my-session --deps auth-impl,tests
```

### ayo ticket show

Show ticket details.

```bash
ayo ticket show <ticket-id> -s <session>
```

Displays full ticket information including:
- Status, type, priority
- Assignee and timestamps
- Dependencies and links
- Description and notes

### ayo ticket start

Start working on a ticket (sets status to in_progress).

```bash
ayo ticket start <ticket-id> -s <session>
```

### ayo ticket close

Close a ticket (sets status to closed).

```bash
ayo ticket close <ticket-id> -s <session>
```

### ayo ticket reopen

Reopen a closed ticket (sets status to pending).

```bash
ayo ticket reopen <ticket-id> -s <session>
```

### ayo ticket block

Mark a ticket as blocked.

```bash
ayo ticket block <ticket-id> -s <session>
```

### ayo ticket assign

Assign a ticket to an agent.

```bash
ayo ticket assign <ticket-id> <agent> -s <session>
```

**Examples:**

```bash
ayo ticket assign proj-a1b2 @coder -s my-session
ayo ticket assign proj-a1b2 "" -s my-session  # Unassign
```

### ayo ticket note

Add a note to a ticket.

```bash
ayo ticket note <ticket-id> <content> -s <session>
```

**Examples:**

```bash
ayo ticket note proj-a1b2 "Completed login endpoint" -s my-session
ayo ticket note proj-a1b2 "Blocked: waiting for API spec" -s my-session
```

### ayo ticket ready

List tickets ready to work on (dependencies resolved).

```bash
ayo ticket ready -s <session> [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--assignee` | `-a` | Filter by assignee |
| `--json` | | Output in JSON format |

### ayo ticket blocked

List tickets blocked on dependencies.

```bash
ayo ticket blocked -s <session> [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--assignee` | `-a` | Filter by assignee |
| `--json` | | Output in JSON format |

### ayo ticket dep

Manage ticket dependencies.

#### ayo ticket dep add

Add a dependency to a ticket.

```bash
ayo ticket dep add <ticket-id> <dep-id> -s <session>
```

The system prevents circular dependencies.

#### ayo ticket dep remove

Remove a dependency from a ticket.

```bash
ayo ticket dep remove <ticket-id> <dep-id> -s <session>
```

### ayo ticket delete

Delete a ticket.

```bash
ayo ticket delete <ticket-id> -s <session> [--force]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation |

---

## ayo share

Share host directories with sandboxed agents.

Shares create symlinks in a workspace directory that is mounted into sandboxes. Changes take effect immediately without requiring sandbox restart. Shared directories appear at `/workspace/{name}` inside the sandbox.

### ayo share add

Share a host directory.

```bash
ayo share add <path> [--flags]
```

| Flag | Description |
|------|-------------|
| `--as` | Custom name for the share |
| `--session` | Remove share when session ends |

**Features:**
- Paths are resolved to absolute paths
- Supports `~/` home directory expansion
- Name is derived from path basename unless `--as` is specified

**Examples:**

```bash
ayo share add ~/Code/myproject           # Share with auto-generated name
ayo share add . --as project             # Share current directory as 'project'
ayo share add ~/data --session           # Share for this session only
```

**Output:**
```
✓ Shared /Users/you/project → /workspace/project
```

### ayo share list

List all shares.

```bash
ayo share list [--json]
```

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |

Aliases: `ayo share ls`

**Output:**
```
  Shares
  ──────────────────────────────────────────────────

  ● project → /workspace/project
    /Users/you/project  2 hours ago

  ○ data → /workspace/data (session)
    /Users/you/data  10 minutes ago

  Access at /workspace/{name} inside sandbox
```

Session shares (temporary) are marked with ○, permanent shares with ●.

### ayo share rm

Remove a share.

```bash
ayo share rm [name|path]
ayo share rm --all
```

| Flag | Description |
|------|-------------|
| `--all` | Remove all shares |

Aliases: `ayo share remove`

**Examples:**

```bash
ayo share rm project              # Remove by name
ayo share rm ~/Code/project       # Remove by path
ayo share rm --all                # Remove all shares
```

**Output:**
```
✓ Removed share: project
✓ Removed 3 share(s)
```

---

## ayo trigger

Manage automated triggers (cron schedules and file watchers).

> **Note:** The singular `trigger` is preferred. `triggers` is still supported as an alias.

### ayo trigger list

List all triggers.

```bash
ayo trigger list
```

### ayo trigger schedule

Create a scheduled (cron) trigger.

```bash
ayo trigger schedule <agent> <schedule> [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--prompt` | `-p` | Prompt to send to agent when triggered |

The schedule can be either cron syntax or natural language:

**Natural language examples:**
- `"every hour"`
- `"every day at 9am"`
- `"every monday at 3pm"`
- `"daily"`, `"hourly"`, `"weekly"`
- `"every 5 minutes"`

**Cron syntax** (6 fields with seconds):
- `"0 0 * * * *"` - every hour
- `"0 30 9 * * *"` - every day at 9:30am
- `"0 0 9 * * MON"` - every Monday at 9am

**Examples:**

```bash
# Natural language
ayo trigger schedule @backup "every hour"
ayo trigger schedule @reports "every day at 9am" --prompt "Generate daily report"
ayo trigger schedule @cleanup "every monday at 3pm"

# Cron syntax
ayo trigger schedule @backup "0 0 * * * *"
ayo trigger schedule @weekly "0 0 9 * * MON"
```

### ayo trigger watch

Create a filesystem watch trigger.

```bash
ayo trigger watch <path> <agent> [patterns...] [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--prompt` | `-p` | Prompt to send to agent when triggered |
| `--recursive` | `-r` | Watch subdirectories |
| `--events` | | Events to trigger on: create, modify, delete |

**Examples:**

```bash
# Watch directory for any changes
ayo trigger watch ./src @build

# Watch for specific file patterns
ayo trigger watch ./src @build "*.go" "*.mod"

# Watch recursively with events filter
ayo trigger watch ./docs @docs "*.md" --recursive --events modify,create
```

### ayo trigger show

Show trigger details.

```bash
ayo trigger show [id]
```

If ID is omitted, shows an interactive picker. Supports prefix matching.

### ayo trigger test

Manually fire a trigger.

```bash
ayo trigger test [id]
```

This will wake the associated agent just as if the trigger had fired naturally.

### ayo trigger enable

Enable a disabled trigger.

```bash
ayo trigger enable [id]
```

### ayo trigger disable

Disable a trigger without removing it.

```bash
ayo trigger disable [id]
```

### ayo trigger rm

Remove a trigger.

```bash
ayo trigger rm [id] [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation |

Aliases: `remove`, `delete`

---

## ayo backup

Manage backups of sandbox state, config, and data.

Backups include:
- Sandbox state (agent homes, shared files)
- Config (~/.config/ayo/)
- Data (~/.local/share/ayo/ except sandbox and backups)

### ayo backup create

Create a new backup.

```bash
ayo backup create [--flags]
```

| Flag | Description |
|------|-------------|
| `--name` | Backup name (default: timestamp) |
| `--json` | Output in JSON format |

### ayo backup list

List all backups.

```bash
ayo backup list [--json]
```

Aliases: `ls`

### ayo backup show

Show backup details.

```bash
ayo backup show <name>
```

### ayo backup restore

Restore from a backup.

```bash
ayo backup restore <name> [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation |
| `--dry-run` | | Preview changes without applying |

### ayo backup export

Export backup to portable archive.

```bash
ayo backup export <name> <destination>
```

Creates a .tar.gz archive that can be transferred to another machine.

### ayo backup import

Import backup from archive.

```bash
ayo backup import <archive>
```

### ayo backup prune

Clean old auto-backups.

```bash
ayo backup prune [--flags]
```

| Flag | Description |
|------|-------------|
| `--keep` | Number of backups to keep (default: 5) |
| `--force` | Skip confirmation |

---

## ayo sync

Synchronize ayo configuration and data with remote storage.

### ayo sync init

Initialize sync for current directory.

```bash
ayo sync init
```

### ayo sync status

Show sync status.

```bash
ayo sync status
```

### ayo sync remote

Manage sync remotes.

```bash
ayo sync remote
```

#### ayo sync remote add

Add a sync remote.

```bash
ayo sync remote add <name> <url>
```

#### ayo sync remote show

Show remote details.

```bash
ayo sync remote show <name>
```

### ayo sync push

Push local changes to remote.

```bash
ayo sync push [remote]
```

### ayo sync pull

Pull remote changes to local.

```bash
ayo sync pull [remote]
```

---

## ayo sandbox service status

Show sandbox service status.

```bash
ayo sandbox service status
```

Displays:
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
