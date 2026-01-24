---
name: plugins
description: |
  Skill for managing ayo plugins.
  Helps with installing, updating, and removing plugins from git repositories.
metadata:
  author: ayo
  version: "1.0"
---

# Plugin Management Skill

This skill helps manage ayo plugins.

## Plugin System Overview

Plugins extend ayo with additional agents, skills, and tools. They are distributed via git repositories with the naming convention `ayo-plugins-<name>`.

### Directory Structure

Plugins are installed to `~/.local/share/ayo/plugins/`.

A plugin can contain:
- `agents/` - Agent definitions
- `skills/` - Skill definitions  
- `tools/` - External tool definitions
- `manifest.json` - Plugin metadata (required)

## Common Tasks

### Installing a Plugin

```bash
# From GitHub (shorthand)
ayo plugins install owner/name

# From GitHub (full URL)
ayo plugins install https://github.com/owner/ayo-plugins-name

# From local directory (for development)
ayo plugins install --local ./my-plugin
```

### Listing Installed Plugins

```bash
ayo plugins list
```

### Viewing Plugin Details

```bash
ayo plugins show <name>
```

### Updating Plugins

```bash
# Update all plugins
ayo plugins update

# Update specific plugin
ayo plugins update <name>

# Check for updates without applying
ayo plugins update --dry-run
```

### Removing a Plugin

```bash
ayo plugins remove <name>
```

## Popular Plugins

### crush

Provides the `@crush` agent for complex source code tasks.

```bash
ayo plugins install alexcabrera/crush
```

## Creating Plugins

### manifest.json

Every plugin needs a manifest:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "My awesome plugin",
  "author": "your-name",
  "agents": ["@my-agent"],
  "skills": ["my-skill"],
  "tools": ["my-tool"],
  "dependencies": {
    "binaries": ["some-binary"]
  },
  "ayo_version": ">=0.2.0"
}
```

### Repository Naming

Plugin repositories must be named `ayo-plugins-<name>`:
- `ayo-plugins-crush` for the "crush" plugin
- `ayo-plugins-research` for a "research" plugin

### Tool Definitions

External tools are defined in `tools/<name>/tool.json`:

```json
{
  "name": "my-tool",
  "description": "What this tool does",
  "command": "my-binary",
  "args": ["--flag"],
  "parameters": [
    {
      "name": "input",
      "description": "Input text",
      "type": "string",
      "required": true
    }
  ],
  "timeout": 60
}
```

## Troubleshooting

### Plugin not loading

1. Check if installed: `ayo plugins list`
2. Verify manifest: `ayo plugins show <name>`
3. Check for missing dependencies

### Agent not found

1. Verify the plugin provides the agent
2. Check for naming conflicts
3. Try reinstalling: `ayo plugins install --force <name>`

### Tool not available

1. Ensure the binary is installed and in PATH
2. Check tool definition in plugin
3. Verify `allowed_tools` in agent config
