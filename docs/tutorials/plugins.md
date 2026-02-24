# Tutorial: Creating Plugins

Build a custom plugin with agents and tools. By the end, you'll know how to package and distribute ayo extensions.

**Time**: ~25 minutes  
**Prerequisites**: [First Agent Tutorial](first-agent.md) complete

## What You'll Build

A `@devtools` plugin with:
- A `@formatter` agent for code formatting
- A custom `prettier` tool
- A skill with formatting rules

## Step 1: Create the Plugin Directory

```bash
mkdir -p ~/my-plugins/devtools
cd ~/my-plugins/devtools
```

## Step 2: Write the Manifest

Create `manifest.json`:

```json
{
  "name": "@acme/devtools",
  "version": "1.0.0",
  "description": "Development tools including formatters and linters",
  "author": "Your Name",
  "license": "MIT",
  "components": {
    "agents": {
      "@formatter": {
        "path": "./agents/formatter"
      }
    },
    "tools": {
      "prettier": {
        "path": "./tools/prettier"
      }
    },
    "skills": {
      "formatting-rules": {
        "path": "./skills/formatting-rules.md"
      }
    }
  },
  "dependencies": {
    "ayo": ">=1.0.0"
  }
}
```

### Manifest Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique plugin identifier |
| `version` | Yes | Semantic version |
| `description` | No | Short description |
| `components` | Yes | Agents, tools, skills, etc. |
| `dependencies` | No | Required ayo version |

## Step 3: Create the Agent

Create the agent directory:

```bash
mkdir -p agents/formatter
```

**agents/formatter/config.json**:
```json
{
  "model": "claude-sonnet-4-20250514",
  "description": "Code formatting specialist",
  "allowed_tools": [
    "bash",
    "view",
    "edit",
    "prettier"
  ],
  "skills": [
    "formatting-rules"
  ]
}
```

**agents/formatter/system.md**:
```markdown
# Code Formatter Agent

You format code according to project standards.

## Responsibilities

- Format code files to consistent style
- Apply language-specific conventions
- Respect existing project configurations (.prettierrc, etc.)

## Process

1. Identify the language/framework
2. Check for existing config files
3. Apply formatting using appropriate tools
4. Verify the result

## Guidelines

- Preserve semantic meaning (formatting only)
- Respect existing project conventions
- Report any files that couldn't be formatted
```

## Step 4: Create the Tool

Create the tool directory:

```bash
mkdir -p tools/prettier
```

**tools/prettier/tool.json**:
```json
{
  "name": "prettier",
  "description": "Format code using Prettier",
  "execution": "sandbox",
  "parameters": {
    "type": "object",
    "properties": {
      "file": {
        "type": "string",
        "description": "Path to the file to format"
      },
      "write": {
        "type": "boolean",
        "description": "Write formatted output to file",
        "default": false
      },
      "parser": {
        "type": "string",
        "description": "Parser to use (typescript, babel, css, etc.)",
        "enum": ["typescript", "babel", "css", "html", "markdown", "json"]
      }
    },
    "required": ["file"]
  }
}
```

**tools/prettier/run.sh**:
```bash
#!/bin/bash
set -e

FILE="$1"
WRITE="$2"
PARSER="$3"

ARGS="--parser ${PARSER:-babel}"

if [ "$WRITE" = "true" ]; then
  ARGS="$ARGS --write"
fi

npx prettier $ARGS "$FILE"
```

Make it executable:
```bash
chmod +x tools/prettier/run.sh
```

### Tool Schema

| Field | Description |
|-------|-------------|
| `name` | Tool identifier |
| `description` | Shown to agent |
| `execution` | `sandbox`, `host`, or `bridge` |
| `parameters` | JSON Schema for inputs |

## Step 5: Create the Skill

Create the skill:

```bash
mkdir -p skills
```

**skills/formatting-rules.md**:
```markdown
# Formatting Rules

## JavaScript/TypeScript

- Use single quotes for strings
- 2-space indentation
- Trailing commas in multiline
- Semicolons required
- 100 character line width

## CSS

- 2-space indentation
- One selector per line
- Opening brace on same line

## Markdown

- 80 character line width for prose
- Fenced code blocks with language
- Blank line before headings

## JSON

- 2-space indentation
- No trailing commas
- Keys in alphabetical order (where meaningful)
```

## Step 6: Plugin Directory Structure

Your complete plugin:

```
devtools/
├── manifest.json
├── agents/
│   └── formatter/
│       ├── config.json
│       └── system.md
├── tools/
│   └── prettier/
│       ├── tool.json
│       └── run.sh
└── skills/
    └── formatting-rules.md
```

## Step 7: Install the Plugin

### From Local Directory

```bash
ayo plugin install ~/my-plugins/devtools
```

### From Git Repository

```bash
ayo plugin install https://github.com/acme/devtools
```

### Verify Installation

```bash
ayo plugin list
```

Output:
```
NAME              VERSION   COMPONENTS
@acme/devtools    1.0.0     agents: 1, tools: 1, skills: 1
```

## Step 8: Use the Plugin

### Use the Agent

```bash
ayo @formatter "Format all TypeScript files in ./src"
```

### Use the Tool Directly

The tool is available to any agent with it enabled:

```json
{
  "allowed_tools": ["prettier"]
}
```

### Use the Skill

Reference in agent config:

```json
{
  "skills": ["formatting-rules"]
}
```

## Adding More Component Types

### Triggers

**manifest.json**:
```json
{
  "components": {
    "triggers": {
      "format-on-save": {
        "path": "./triggers/format-on-save"
      }
    }
  }
}
```

**triggers/format-on-save/trigger.yaml**:
```yaml
name: format-on-save
type: watch
config:
  patterns:
    - "*.ts"
    - "*.tsx"
  events:
    - modify
  debounce: 2s
agent: "@formatter"
prompt: "Format the changed file: {{file}}"
```

### Planners

**manifest.json**:
```json
{
  "components": {
    "planners": {
      "code-review-planner": {
        "path": "./planners/code-review",
        "type": "near_term"
      }
    }
  }
}
```

### Squads

**manifest.json**:
```json
{
  "components": {
    "squads": {
      "formatting-team": {
        "path": "./squads/formatting-team"
      }
    }
  }
}
```

## Plugin Management

### List Installed Plugins

```bash
ayo plugin list
```

### Show Plugin Details

```bash
ayo plugin show @acme/devtools
```

### Update a Plugin

```bash
ayo plugin update @acme/devtools
```

### Remove a Plugin

```bash
ayo plugin remove @acme/devtools
```

## Resolution Order

When loading components, ayo checks in order:

1. **User-defined**: `~/.config/ayo/`
2. **Installed plugins**: `~/.local/share/ayo/plugins/`
3. **Built-in**: Bundled with ayo

This allows users to override plugin behavior.

## Publishing Plugins

### GitHub

Push to a GitHub repository:

```bash
git init
git add .
git commit -m "Initial release"
git remote add origin https://github.com/acme/devtools
git push -u origin main
```

Users install with:
```bash
ayo plugin install https://github.com/acme/devtools
```

### Plugin Registry (Coming Soon)

```bash
ayo plugin publish
```

## Troubleshooting

### Plugin not loading

Check manifest syntax:
```bash
cat manifest.json | jq .
```

### Agent not found after install

```bash
# Check plugin is installed
ayo plugin list | grep devtools

# Check agent is registered
ayo agents list | grep formatter
```

### Tool not available

Ensure the agent has the tool enabled:

```json
{
  "allowed_tools": ["prettier"]
}
```

### Skill not loading

Check the skill path is correct and file exists:

```bash
ls ~/.local/share/ayo/plugins/@acme/devtools/skills/
```

## Next Steps

- [Agents Guide](../guides/agents.md) - Advanced agent configuration
- [Tools Guide](../guides/tools.md) - Creating complex tools
- [Reference: Plugins](../reference/plugins.md) - Complete manifest schema

---

*You've created your first plugin! Share it with the community.*
