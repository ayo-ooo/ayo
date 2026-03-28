# ayo-manage: Manage the ayo agent registry

Use this skill when the user wants to list available agents, get details about a specific agent, register new agents, or remove agents from the registry.

## Overview

Ayo maintains a registry of agents at `~/.config/ayo/registry.toml`. The registry tracks agent names, versions, descriptions, types, and paths to their source and compiled binaries.

## Commands

### List all agents

```bash
# Human-readable table
ayo list

# Filter by type
ayo list --type tool
ayo list --type conversational

# Machine-readable JSON
ayo list --json
```

### JSON output format for `ayo list --json`

```json
[
  {
    "Name": "code-reviewer",
    "Version": "1.0.0",
    "Description": "Reviews code for bugs and style",
    "SourcePath": "/path/to/code-reviewer",
    "BinaryPath": "/path/to/code-reviewer-binary",
    "Type": "tool",
    "RegisteredAt": "2026-03-28T12:00:00Z"
  }
]
```

### Describe an agent

```bash
# Human-readable details
ayo describe <agent-name>

# Machine-readable JSON (includes schemas, skills, hooks)
ayo describe <agent-name> --json
```

### JSON output format for `ayo describe --json`

```json
{
  "ayo_version": 1,
  "name": "code-reviewer",
  "version": "1.0.0",
  "description": "Reviews code for bugs and style",
  "type": "tool",
  "input_schema": { "type": "object", "properties": { ... } },
  "output_schema": { "type": "object", "properties": { ... } },
  "skills": ["analyze", "report"],
  "hooks": ["agent-start", "agent-finish"],
  "source_path": "/path/to/source",
  "binary_path": "/path/to/binary",
  "invocation": {
    "direct": "/path/to/binary",
    "via_ayo": "ayo run code-reviewer",
    "piped": "echo '{...}' | ayo run code-reviewer"
  }
}
```

### Register an agent

```bash
# Register from a project directory
ayo register ./my-agent

# Register from a compiled binary
ayo register /usr/local/bin/my-agent

# Build and register in one step (preferred)
ayo runthat ./my-agent --register
```

### Remove an agent

```bash
# Remove from registry only
ayo remove <agent-name>

# Remove and delete the binary
ayo remove <agent-name> --delete-binary
```

## Common Workflows

### "What agents do I have?"

```bash
ayo list
```

Or programmatically:

```bash
ayo list --json | jq '.[].Name'
```

### "Tell me about the summarize agent"

```bash
ayo describe summarize
```

For programmatic use (e.g., to understand input schema before invoking):

```bash
ayo describe summarize --json | jq '.input_schema.properties'
```

### "Register this agent I just built"

After building:

```bash
ayo runthat ./my-agent --register
```

Or register an existing binary:

```bash
ayo register ./my-agent-binary
```

### "Clean up stale agents"

List agents and check which binaries still exist:

```bash
ayo list --json | jq '.[] | select(.BinaryPath != "") | .BinaryPath' | while read path; do
  [ -f "$path" ] || echo "Missing: $path"
done
```

Remove stale entries:

```bash
ayo remove old-agent
```

### "What tool agents can process my data?"

```bash
ayo list --json --type tool | jq '.[].Name'
```

Then inspect each one:

```bash
ayo describe data-formatter --json | jq '.input_schema'
```
