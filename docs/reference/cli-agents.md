# ayo agents

Manage AI agents—list, create, show, and edit agent configurations.

## Synopsis

```
ayo agents <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List all available agents |
| `show` | Show agent details |
| `create` | Create a new agent |
| `edit` | Edit an agent |
| `delete` | Delete an agent |
| `validate` | Validate agent configuration |

---

## ayo agents list

List all available agents.

### Synopsis

```
ayo agents list [flags]
```

### Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--all` | `-a` | bool | Include system agents |

### Example

```bash
$ ayo agents list
NAME        MODEL                      DESCRIPTION
@ayo        claude-sonnet-4-20250514   Default assistant
@reviewer   claude-sonnet-4-20250514   Code review specialist
@writer     gpt-4o                     Technical writer
```

### JSON Output

```json
{
  "agents": [
    {
      "name": "@ayo",
      "model": "claude-sonnet-4-20250514",
      "description": "Default assistant",
      "path": "/Users/user/.config/ayo/agents/@ayo"
    }
  ]
}
```

---

## ayo agents show

Show detailed information about an agent.

### Synopsis

```
ayo agents show <name>
```

### Example

```bash
$ ayo agents show @reviewer
Name:         @reviewer
Model:        claude-sonnet-4-20250514
Description:  Code review specialist
Path:         /Users/user/.config/ayo/agents/@reviewer

Tools:
  - bash
  - read_file
  - write_file

Skills:
  - code-review
  - security

System Prompt:
  You are an expert code reviewer...
```

---

## ayo agents create

Create a new agent interactively or from flags.

### Synopsis

```
ayo agents create <name> [flags]
```

### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--model` | string | Model to use |
| `--description` | string | Agent description |
| `--system` | string | System prompt (or path to .md file) |
| `--tools` | strings | Tools to enable |
| `--skills` | strings | Skills to include |

### Examples

```bash
# Interactive creation
$ ayo agents create @analyst
Creating agent @analyst...
Model [claude-sonnet-4-20250514]: 
Description: Data analysis specialist
Created @analyst at /Users/user/.config/ayo/agents/@analyst
```

```bash
# Non-interactive
$ ayo agents create @helper \
  --model gpt-4o \
  --description "General helper" \
  --tools bash,read_file
Created @helper
```

---

## ayo agents edit

Open an agent's configuration in your editor.

### Synopsis

```
ayo agents edit <name>
```

### Example

```bash
$ ayo agents edit @reviewer
# Opens $EDITOR with @reviewer/config.json
```

---

## ayo agents delete

Delete an agent.

### Synopsis

```
ayo agents delete <name> [flags]
```

### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--force` | bool | Skip confirmation |

### Example

```bash
$ ayo agents delete @oldagent
Delete @oldagent? [y/N] y
Deleted @oldagent
```

---

## ayo agents validate

Validate an agent's configuration.

### Synopsis

```
ayo agents validate <name>
```

### Example

```bash
$ ayo agents validate @reviewer
✓ config.json is valid
✓ system.md exists
✓ Model 'claude-sonnet-4-20250514' is available
✓ All tools are valid
```

---

## Agent Directory Structure

```
@myagent/
├── config.json    # Agent configuration
├── system.md      # System prompt
└── skills/        # Optional: embedded skills
    └── custom.md
```

### config.json

```json
{
  "model": "claude-sonnet-4-20250514",
  "description": "My custom agent",
  "temperature": 0.7,
  "tools": ["bash", "read_file", "write_file"],
  "skills": ["debugging"],
  "sandbox": {
    "enabled": true,
    "image": "default"
  }
}
```

### system.md

```markdown
You are a specialized assistant for...

## Capabilities

- ...

## Guidelines

- ...
```

## See Also

- [Agents Guide](../agents.md) - Conceptual overview
- [ayo skill](cli-skill.md) - Managing skills
