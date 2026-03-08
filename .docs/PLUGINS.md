# Ayo Build System - Plugin Architecture

This document explains how the plugin system works in the Ayo build system, including plugin types, installation, configuration, and usage patterns.

---

## Table of Contents

1. [Plugin Overview](#plugin-overview)
2. [Plugin Types](#plugin-types)
3. [Plugin Structure](#plugin-structure)
4. [Manifest Specification](#manifest-specification)
5. [Installation Process](#installation-process)
6. [Configuration Integration](#configuration-integration)
7. [Tool Categories](#tool-categories)
8. [Creating Plugins](#creating-plugins)
9. [Plugin Examples](#plugin-examples)
10. [Best Practices](#best-practices)

---

## Plugin Overview

Plugins extend Ayo's functionality by providing additional agents, skills, tools, and providers. In the build system, plugins are distributed as Git repositories and installed into the local plugin directory.

### Key Characteristics

- **Git-based distribution**: Plugins are hosted in Git repositories
- **Manifest-driven**: Each plugin has a `manifest.json` defining its contents
- **Isolated execution**: Plugins run in their own context
- **Configuration integration**: Plugins can define defaults and aliases
- **Dependency management**: Plugins declare their requirements

### Plugin Repository Structure

```
github.com/ayo-plugins/{plugin-name}/
├── manifest.json          # Plugin manifest (required)
├── agents/                # Agent definitions
│   └── {agent-name}/      # Individual agents
│       ├── config.toml    # Agent configuration
│       ├── prompts/      # Prompt templates
│       └── skills/        # Agent skills
├── skills/                # Shared skills
│   └── {skill-name}/      # Individual skills
├── tools/                 # Tool implementations
│   └── {tool-name}/       # Individual tools
├── providers/             # Provider implementations
│   └── {provider-type}/   # Type-specific providers
└── scripts/                # Optional installation scripts
```

---

## Plugin Types

### 1. Agent Plugins

Provide pre-configured agents that can be added to projects.

**Example**: `ayo-plugins-code-reviewer` - Provides a code review agent

**Manifest**:
```json
{
  "name": "code-reviewer",
  "version": "1.0.0",
  "description": "Code review and analysis agent",
  "agents": ["reviewer"],
  "delegates": {
    "code_review": "@reviewer"
  }
}
```

### 2. Skill Plugins

Provide reusable skills that agents can use.

**Example**: `ayo-plugins-security` - Provides security analysis skills

**Manifest**:
```json
{
  "name": "security",
  "version": "1.0.0",
  "description": "Security analysis skills",
  "skills": ["security-audit", "vulnerability-scan"]
}
```

### 3. Tool Plugins

Provide additional tools that agents can use.

**Example**: `ayo-plugins-websearch` - Provides web search tools

**Manifest**:
```json
{
  "name": "websearch",
  "version": "1.0.0",
  "description": "Web search capabilities",
  "tools": ["google-search", "github-search"],
  "default_tools": {
    "search": "google-search"
  }
}
```

### 4. Provider Plugins

Provide implementations of core services:

- **Memory providers**: Different memory backends
- **Sandbox providers**: Different sandbox environments
- **Embedding providers**: Different embedding models
- **Observer providers**: Different monitoring systems

**Example**: `ayo-plugins-postgres-memory`

**Manifest**:
```json
{
  "name": "postgres-memory",
  "version": "1.0.0",
  "description": "PostgreSQL memory provider",
  "providers": [
    {
      "name": "postgres",
      "type": "memory",
      "description": "PostgreSQL-based memory storage",
      "entry_point": "providers/memory/postgres.so"
    }
  ]
}
```

### 5. Planner Plugins

Provide coordination and planning strategies:

- **Near-term planners**: Session-scoped task management
- **Long-term planners**: Persistent ticket systems

**Example**: `ayo-plugins-jira`

**Manifest**:
```json
{
  "name": "jira",
  "version": "1.0.0",
  "description": "Jira integration for ticket management",
  "planners": [
    {
      "name": "jira",
      "type": "long",
      "description": "Jira-based ticket system",
      "entry_point": "planners/jira.so"
    }
  ]
}
```

---

## Plugin Structure

### Required Files

1. **manifest.json**: Plugin manifest with metadata and contents
2. **agents/**: Directory containing agent definitions (if plugin provides agents)
3. **skills/**: Directory containing skill definitions (if plugin provides skills)
4. **tools/**: Directory containing tool implementations (if plugin provides tools)

### Optional Files

- **providers/**: Provider implementations
- **planners/**: Planner implementations
- **scripts/post-install.sh**: Post-installation script
- **README.md**: Plugin documentation
- **LICENSE**: License file

### Agent Structure

```
agents/{agent-name}/
├── config.toml          # Agent configuration
├── prompts/            # Prompt templates
│   ├── system.md       # System prompt
│   └── user.md         # User prompt template
├── skills/             # Agent-specific skills
│   └── {skill-name}/   # Individual skills
└── tools/              # Agent-specific tools
```

### Tool Structure

```
tools/{tool-name}/
├── main.go             # Go implementation
├── config.json         # Default configuration
└── README.md           # Tool documentation
```

---

## Manifest Specification

### Complete Manifest Example

```json
{
  "name": "example-plugin",
  "version": "1.0.0",
  "description": "Example plugin demonstrating all features",
  "author": "Ayo Team",
  "repository": "https://github.com/ayo-plugins/example-plugin",
  "license": "MIT",
  "agents": ["example-agent"],
  "skills": ["example-skill"],
  "tools": ["example-tool"],
  "delegates": {
    "example_task": "@example-agent"
  },
  "default_tools": {
    "example_category": "example-tool"
  },
  "dependencies": {
    "binaries": [
      {
        "name": "jq",
        "install_hint": "Install with: brew install jq",
        "install_url": "https://stedolan.github.io/jq/download/"
      }
    ]
  },
  "post_install": "scripts/setup.sh",
  "ayo_version": ">=0.2.0",
  "providers": [
    {
      "name": "example-memory",
      "type": "memory",
      "description": "Example memory provider",
      "entry_point": "providers/memory/example.so"
    }
  ],
  "planners": [
    {
      "name": "example-planner",
      "type": "near",
      "description": "Example near-term planner",
      "entry_point": "planners/example.so"
    }
  ],
  "sandbox_configs": [
    {
      "name": "example-sandbox",
      "description": "Example sandbox configuration"
    }
  ],
  "squads": [
    {
      "name": "example-squad",
      "description": "Example squad configuration",
      "agents": ["example-agent"]
    }
  ],
  "triggers": [
    {
      "name": "example-trigger",
      "category": "poll",
      "description": "Example trigger",
      "entry_point": "triggers/example"
    }
  ]
}
```

### Manifest Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | ✅ | Plugin identifier (lowercase alphanumeric with hyphens) |
| `version` | string | ✅ | Semantic version (e.g., "1.0.0") |
| `description` | string | ✅ | Brief description of plugin |
| `author` | string | ❌ | Plugin author or organization |
| `repository` | string | ❌ | Git repository URL |
| `license` | string | ❌ | SPDX license identifier |
| `agents` | string[] | ❌ | List of agent handles provided |
| `skills` | string[] | ❌ | List of skill names provided |
| `tools` | string[] | ❌ | List of tool names provided |
| `delegates` | object | ❌ | Task type to agent handle mappings |
| `default_tools` | object | ❌ | Tool category to implementation mappings |
| `dependencies` | object | ❌ | External requirements |
| `post_install` | string | ❌ | Post-installation script path |
| `ayo_version` | string | ❌ | Minimum Ayo version required |
| `providers` | ProviderDef[] | ❌ | Provider implementations |
| `planners` | PlannerDef[] | ❌ | Planner implementations |
| `sandbox_configs` | SandboxConfigDef[] | ❌ | Sandbox configurations |
| `squads` | SquadDef[] | ❌ | Squad definitions |
| `triggers` | TriggerDef[] | ❌ | Trigger type definitions |

---

## Installation Process

### Installation Command

```bash
# Install a plugin from Git repository
ao plugins install https://github.com/ayo-plugins/plugin-name.git

# Install with custom name
ao plugins install https://github.com/ayo-plugins/plugin-name.git --as custom-name

# Force reinstall
ao plugins install https://github.com/ayo-plugins/plugin-name.git --force

# Skip dependency check
ao plugins install https://github.com/ayo-plugins/plugin-name.git --skip-deps
```

### Installation Steps

1. **Parse plugin reference**: Extract Git URL and plugin name
2. **Check requirements**: Verify Git is installed
3. **Load registry**: Check if plugin already installed
4. **Clone repository**: Download plugin to local directory
5. **Security scan**: Run security checks on plugin code
6. **Validate manifest**: Check manifest structure and contents
7. **Check dependencies**: Verify required binaries are available
8. **Run post-install script**: Execute any setup scripts
9. **Update registry**: Add plugin to installed plugins list

### Installation Locations

```
~/.local/share/ayo/
├── plugins/                # Installed plugins
│   └── {plugin-name}/     # Individual plugins
│       ├── manifest.json  # Plugin manifest
│       ├── agents/        # Agent definitions
│       ├── skills/        # Skill definitions
│       └── tools/          # Tool implementations
└── registry.json          # Installed plugins registry
```

---

## Configuration Integration

### Plugin Configuration Files

Plugins can provide default configurations that get merged with project configurations.

**Example**: Tool category configuration

```toml
# In project's config.toml
[default_tools]
search = "google-search"  # Uses tool from websearch plugin
plan = "jira"            # Uses planner from jira plugin
```

### Agent Integration

After installing a plugin, its agents can be added to projects:

```bash
# Add plugin agent to project
ao add-agent my-project plugin-agent

# The agent configuration is copied from plugin to project
# Agent can then be customized for the specific project
```

### Tool Integration

Plugin tools become available through tool categories:

```bash
# List available tool categories
ao tools categories

# Show current tool mappings
ao config show default_tools

# Use a plugin tool
ao chat my-project --tool search
```

---

## Tool Categories

Tool categories provide abstraction over specific tool implementations, allowing plugins to provide alternatives.

### Built-in Categories

| Category | Default | Description | Plugin Example |
|----------|---------|-------------|----------------|
| `shell` | `bash` | Command execution | `ayo-plugins-zsh` |
| `search` | (none) | Web search | `ayo-plugins-websearch` |
| `plan` | (none) | Project planning | `ayo-plugins-jira` |

### Category Resolution

```
User requests: "search"
    ↓
Check config.default_tools["search"]
    ↓
If found: use configured tool
    ↓
If not found: use builtin default
    ↓
If no default: return error
```

### Example: Web Search Configuration

```bash
# Install web search plugin
ao plugins install https://github.com/ayo-plugins/websearch.git

# Configure search category to use google-search tool
ao config set default_tools.search google-search

# Now agents can use search without knowing implementation
ao chat my-project "Search the web for AI trends"
```

---

## Creating Plugins

### Plugin Development Workflow

1. **Create repository**: `ayo-plugins-{name}` format
2. **Write manifest**: Define plugin contents in `manifest.json`
3. **Implement features**: Create agents, skills, or tools
4. **Test locally**: Install and test plugin
5. **Document**: Add README and usage examples
6. **Publish**: Push to Git repository

### Plugin Development Example

```bash
# 1. Create plugin repository
mkdir ayo-plugins-my-plugin
cd ayo-plugins-my-plugin

# 2. Create manifest.json
cat > manifest.json <<EOF
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "My custom plugin",
  "tools": ["my-tool"],
  "default_tools": {
    "my-category": "my-tool"
  }
}
EOF

# 3. Create tool implementation
mkdir -p tools/my-tool
cat > tools/my-tool/main.go <<'EOF'
package main

import (
    "context"
    "fmt"
    "github.com/alexcabrera/ayo/internal/tools"
)

type MyTool struct {}

func (t *MyTool) Name() string {
    return "my-tool"
}

func (t *MyTool) Description() string {
    return "My custom tool"
}

func (t *MyTool) Execute(ctx context.Context, input tools.Input) (tools.Output, error) {
    result := fmt.Sprintf("Processed: %s", input.Data)
    return tools.Output{Data: result}, nil
}

func init() {
    tools.Register("my-tool", &MyTool{})
}
EOF

# 4. Test locally
go build -o my-tool.so tools/my-tool/main.go

# 5. Install and test
ao plugins install file://$(pwd)
ao tools list | grep my-tool
```

### Plugin Development Best Practices

1. **Semantic Versioning**: Follow semver for plugin versions
2. **Dependency Declaration**: List all required binaries
3. **Security**: Include security scanning in CI/CD
4. **Documentation**: Provide clear usage examples
5. **Backward Compatibility**: Maintain API stability
6. **Testing**: Include integration tests
7. **Performance**: Optimize tool execution

---

## Plugin Examples

### 1. Web Search Plugin

**Repository**: `ayo-plugins-websearch`

**Manifest**:
```json
{
  "name": "websearch",
  "version": "1.0.0",
  "description": "Web search capabilities using multiple engines",
  "tools": ["google-search", "bing-search", "duckduckgo-search"],
  "default_tools": {
    "search": "google-search"
  },
  "dependencies": {
    "binaries": ["curl"]
  }
}
```

**Usage**:
```bash
# Install plugin
ao plugins install https://github.com/ayo-plugins/websearch.git

# Configure search category
ao config set default_tools.search google-search

# Use in agent
ao chat my-agent "Search for Go programming best practices"
```

### 2. Jira Integration Plugin

**Repository**: `ayo-plugins-jira`

**Manifest**:
```json
{
  "name": "jira",
  "version": "2.0.0",
  "description": "Jira integration for ticket management",
  "planners": [
    {
      "name": "jira",
      "type": "long",
      "description": "Jira-based ticket system"
    }
  ],
  "tools": ["jira-create", "jira-update", "jira-search"],
  "dependencies": {
    "binaries": ["jq"]
  }
}
```

**Usage**:
```bash
# Install plugin
ao plugins install https://github.com/ayo-plugins/jira.git

# Configure planner
ao config set planners.long_term jira

# Create Jira ticket via agent
ao chat my-agent "Create Jira ticket for login bug"
```

### 3. Code Analysis Plugin

**Repository**: `ayo-plugins-code-analysis`

**Manifest**:
```json
{
  "name": "code-analysis",
  "version": "1.5.0",
  "description": "Advanced code analysis tools",
  "agents": ["code-reviewer", "security-analyst"],
  "tools": ["static-analysis", "coverage-report", "dependency-check"],
  "delegates": {
    "code_review": "@code-reviewer",
    "security_scan": "@security-analyst"
  },
  "dependencies": {
    "binaries": ["gosec", "staticcheck"]
  }
}
```

**Usage**:
```bash
# Install plugin
ao plugins install https://github.com/ayo-plugins/code-analysis.git

# Add agents to project
ao add-agent my-project code-reviewer
ao add-agent my-project security-analyst

# Use analysis tools
ao chat my-project "Analyze this code for security issues" --agent security-analyst
```

---

## Best Practices

### Plugin Development

1. **Single Responsibility**: Each plugin should focus on one specific function
2. **Clear Documentation**: Provide examples and usage patterns
3. **Version Compatibility**: Specify minimum Ayo version requirements
4. **Dependency Management**: Document all requirements clearly
5. **Security**: Follow secure coding practices
6. **Performance**: Optimize for fast execution
7. **Error Handling**: Provide clear error messages

### Plugin Usage

1. **Start Small**: Install one plugin at a time
2. **Test Thoroughly**: Verify plugin works before production use
3. **Configure Properly**: Set up tool categories and defaults
4. **Monitor Dependencies**: Ensure required binaries are available
5. **Update Regularly**: Keep plugins current
6. **Backup Configs**: Save configuration before updates
7. **Isolate Testing**: Test plugins in separate environments

### Plugin Management

```bash
# List installed plugins
ao plugins list

# Show plugin details
ao plugins show plugin-name

# Update plugin
ao plugins update plugin-name

# Remove plugin
ao plugins remove plugin-name

# Check for updates
ao plugins check-updates
```

---

## Plugin Architecture Diagram

```
User
  ↓
Command: ayo plugins install https://github.com/ayo-plugins/example.git
  ↓
Plugin System
  ↓
1. Parse URL → (git URL, plugin name)
  ↓
2. Check Git availability
  ↓
3. Clone repository to ~/.local/share/ayo/plugins/example/
  ↓
4. Load and validate manifest.json
  ↓
5. Check dependencies (binaries, other plugins)
  ↓
6. Run security scan
  ↓
7. Execute post-install script (if any)
  ↓
8. Update registry
  ↓
Plugin Installed
  ↓
Configuration Integration
  ↓
- Agents available via ayo add-agent
- Tools available via tool categories
- Providers available via config
- Planners available via config
  ↓
Ready for Use
```

---

## Troubleshooting

### Common Plugin Issues

**Issue**: Plugin installation fails with "git not found"
**Solution**: Install Git and ensure it's in PATH

**Issue**: Plugin fails dependency check
**Solution**: Install required binaries or use `--skip-deps` (not recommended)

**Issue**: Plugin tools not available
**Solution**: Check tool category configuration and manifest

**Issue**: Plugin agents not showing up
**Solution**: Verify agent names in manifest match directory structure

**Issue**: Security scan fails
**Solution**: Review plugin code and fix security issues

### Debugging Commands

```bash
# Show plugin registry
ao plugins registry

# Validate plugin manifest
ao plugins validate /path/to/plugin

# Check plugin dependencies
ao plugins deps plugin-name

# Show plugin configuration
ao plugins config plugin-name

# Test plugin tool
ao tools test plugin-tool
```

---

## Future Plugin Features

### Planned Enhancements

1. **Plugin Marketplace**: Central repository for discovering plugins
2. **Automatic Updates**: Background update checking
3. **Plugin Signing**: Cryptographic verification of plugins
4. **Sandboxed Execution**: Isolated plugin execution environments
5. **Dependency Resolution**: Automatic dependency installation
6. **Plugin Bundles**: Group related plugins together
7. **Version Constraints**: Better version compatibility management
8. **Plugin Metrics**: Usage tracking and performance monitoring

### Experimental Features

- **Dynamic Loading**: Load plugins at runtime without restart
- **Plugin Isolation**: Run plugins in separate processes
- **Plugin API**: Standardized plugin development API
- **Plugin Testing**: Built-in plugin test framework

---

## Conclusion

The Ayo plugin system provides a powerful extension mechanism that allows the build system to be customized and extended without modifying core functionality. Plugins enable:

- **Custom Agents**: Add specialized agents for specific tasks
- **Additional Tools**: Extend tool capabilities with new implementations
- **Alternative Providers**: Use different backends for core services
- **Advanced Planning**: Integrate with external ticket systems
- **Domain-Specific Skills**: Add industry-specific knowledge

By leveraging the plugin system, users can tailor Ayo to their specific needs while maintaining a clean, modular architecture. The Git-based distribution ensures easy sharing and version control, while the manifest system provides structured metadata for discovery and integration.
