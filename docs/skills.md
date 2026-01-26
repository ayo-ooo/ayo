# Skills

Skills are reusable instruction sets that teach agents specialized tasks. They follow the [agentskills spec](https://agentskills.org).

## Built-in Skills

| Skill | Description |
|-------|-------------|
| `ayo` | CLI documentation for programmatic use |
| `debugging` | Systematic debugging techniques |
| `coding` | Source code creation guidelines |
| `memory` | Memory tool usage guidelines |
| `plugins` | Plugin management guidance |
| `agent-discovery` | Discovering and inspecting agents |
| `flows` | Composable agent pipelines and automation |

## Skill Discovery

Skills are discovered from multiple sources (in priority order):

1. **Agent-specific** - In agent's `skills/` directory
2. **User shared** - `~/.config/ayo/skills/`
3. **Built-in** - `~/.local/share/ayo/skills/`
4. **Plugin-provided** - In installed plugins

First match wins, allowing overrides.

## Managing Skills

### List Skills

```bash
# List all skills
ayo skills list

# Filter by source
ayo skills list --source=built-in
ayo skills list --source=user
```

### Show Details

```bash
ayo skills show debugging
```

### Validate

```bash
ayo skills validate ./path/to/skill
```

### Update Built-ins

```bash
ayo skills update
ayo skills update --force
```

## Creating Skills

### Interactive

```bash
# Create in current directory
ayo skills create my-skill

# Create in shared directory
ayo skills create my-skill --shared
```

### Skill Structure

```
my-skill/
├── SKILL.md            # Required: skill definition
├── scripts/            # Optional: executable code
├── references/         # Optional: additional documentation
└── assets/             # Optional: templates, data files
```

### SKILL.md Format

```markdown
---
name: my-skill
description: |
  What this skill does and when to use it.
  This is shown to the agent to help it decide when to apply the skill.
metadata:
  author: your-name
  version: "1.0"
compatibility: Requires bash and common Unix utilities
---

# Skill Title

Detailed instructions for the agent on how to use this skill.

## When to Use

- Scenario 1
- Scenario 2

## How to Use

Step-by-step guidance...

## Examples

### Example 1

```bash
command example
```
```

### Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Skill identifier (1-64 chars, lowercase, hyphens ok) |
| `description` | Yes | When to use this skill (1-1024 chars) |
| `metadata` | No | Key-value pairs (author, version, etc.) |
| `compatibility` | No | Environment requirements (max 500 chars) |

**Note:** The `name` must match the directory name.

## Agent Configuration

### Attaching Skills

In agent's `config.json`:

```json
{
  "skills": ["debugging", "coding"]
}
```

### Excluding Skills

```json
{
  "exclude_skills": ["unwanted-skill"]
}
```

### Ignoring Sources

```json
{
  "ignore_builtin_skills": true,
  "ignore_shared_skills": true
}
```

## Agent-Specific Skills

Create skills inside an agent's directory:

```
@myagent/
├── config.json
├── system.md
└── skills/
    └── my-private-skill/
        └── SKILL.md
```

These skills are only available to that agent and take priority over shared/built-in skills.

## Example: Custom Debugging Skill

```markdown
---
name: python-debugging
description: |
  Python-specific debugging techniques.
  Use when debugging Python code, tracebacks, or pytest failures.
metadata:
  author: devteam
  version: "1.0"
compatibility: Requires Python 3.8+ and pytest
---

# Python Debugging

## Quick Diagnostics

When encountering a Python error:

1. Read the full traceback bottom-to-top
2. Identify the exception type and message
3. Locate the failing line

## Common Issues

### ImportError

```bash
# Check if module is installed
pip show module-name

# Check Python path
python -c "import sys; print(sys.path)"
```

### Pytest Failures

```bash
# Run with verbose output
pytest -v path/to/test.py

# Run specific test
pytest path/to/test.py::test_function -v

# Show print statements
pytest -s path/to/test.py
```

## Debugging Commands

```bash
# Interactive debugger
python -m pdb script.py

# Post-mortem debugging
python -c "import pdb; pdb.pm()"
```
```

## Skills in Plugins

Plugins can provide shared skills:

```
ayo-plugins-example/
├── manifest.json
└── skills/
    └── my-skill/
        └── SKILL.md
```

In `manifest.json`:

```json
{
  "skills": ["my-skill"]
}
```

## Tool Requirements

Skills can specify tool requirements (experimental):

```yaml
---
name: my-skill
description: Skill requiring specific tools
allowed-tools:
  - bash
  - my-custom-tool
---
```

The agent must have these tools in `allowed_tools` to use the skill.
