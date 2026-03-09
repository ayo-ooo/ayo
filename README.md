# ayo - Build System for AI Agents

**Ayo is a pure build system for creating standalone AI agent executables.**

```bash
# Create a new agent project
ayo fresh my-agent

# Build the agent
ayo build my-agent

# Run the compiled agent
./my-agent "Hello, analyze this code"
```

## Overview

Ayo transforms agent definitions into self-contained, distributable executables. No runtime dependencies, no framework required - just pure agent binaries.

### Key Features

- **Pure Build System**: Compile agents to standalone binaries
- **No Runtime Framework**: Built agents run independently
- **Simple Project Structure**: Single `config.toml` per agent
- **Cross-Platform**: Build for Linux, macOS, Windows
- **Tool Integration**: Use existing CLI tools, no Go required

## 📚 Comprehensive Documentation

Ayo provides a complete operator manual that takes you from novice to expert:

### 🎓 Learning Path

```mermaid
graph LR
    A[Getting Started] --> B[Basic Usage]
    B --> C[Intermediate Techniques]
    C --> D[Advanced Patterns]
    D --> E[Expert Reference]
    E --> F[Troubleshooting]
```

### 📖 Operator Manual

| Level | Guide | Description |
|-------|-------|-------------|
| 🟢 Novice | [Getting Started](docs/operator-manual/01-getting-started.md) | Installation, first agent, basic concepts |
| 🟡 Beginner | [Basic Usage](docs/operator-manual/02-basic-usage.md) | Configuration, tools, I/O patterns, debugging |
| 🟠 Intermediate | [Intermediate Techniques](docs/operator-manual/03-intermediate-techniques.md) | Prompt engineering, skills, memory optimization |
| 🔴 Advanced | [Advanced Patterns](docs/operator-manual/04-advanced-patterns.md) | Multi-agent systems, production deployment, scaling |
| 🟣 Expert | [Expert Reference](docs/operator-manual/05-expert-reference.md) | Internals, design patterns, contributing |
| 🛠 All Levels | [Troubleshooting](docs/operator-manual/06-troubleshooting.md) | Common issues, best practices, optimization |

### 🍳 Cookbook

Practical examples and recipes:
- [Cookbook](docs/COOKBOOK.md) - File processing, web automation, data analysis
- [Patterns](docs/patterns/) - Ticket workers, scheduled agents, watchers

### 📖 Reference

- [Operator Manual](docs/OPERATOR_MANUAL.md) - Comprehensive usage guide
- [Build System](docs/BUILD_SYSTEM.md) - Technical overview
- [Concepts](docs/concepts.md) - Core concepts and architecture

## Quick Start

### 1. Create an agent project

```bash
ayo fresh my-agent
```

This creates:
```
my-agent/
├── config.toml          # Agent configuration
├── skills/             # Agent skills (optional)
├── tools/               # Executable tools (optional)
└── prompts/
    └── system.md       # System prompt
```

### 2. Configure your agent

Edit `config.toml`:

```toml
[agent]
name = "my-agent"
description = "My AI assistant"
model = "claude-3-5-sonnet"

[cli]
mode = "hybrid"        # freeform, hybrid, or structured
description = "My agent CLI"

[agent.tools]
allowed = ["bash", "file_read", "file_write"]
```

### 3. Build your agent

```bash
ayo build my-agent
```

### 4. Run your agent

```bash
./my-agent "Analyze this code"
```

## Project Structure

### Single Agent

```
my-agent/
├── config.toml          # Main configuration
├── skills/             # Agent skills (optional)
│   └── custom/          # Custom skills
│       └── SKILL.md    # Skill definition
├── tools/               # Executable tools (optional)
│   └── mytool           # Any executable program
└── prompts/             # Prompt templates
    └── system.md       # System prompt
```

### Multi-Agent Team

```
my-team/
├── config.toml          # Main configuration
├── agents/             # Multiple agents
│   ├── agent1/
│   │   └── config.toml  # Agent 1 config
│   └── agent2/
│       └── config.toml  # Agent 2 config
├── workspace/           # Shared workspace (optional)
└── team.toml           # Team configuration
```

## Commands

### `ayo fresh`

Create a new agent project.

```bash
# Create agent with defaults
ayo fresh my-agent

# With custom settings
ayo fresh my-agent \
  --description "Code reviewer" \
  --model "gpt-4-turbo" \
  --template advanced
```

### `ayo build`

Compile agent to standalone executable.

```bash
# Build current directory
ayo build

# Build specific project
ayo build ./my-agent

# Cross-compile for Linux
ayo build ./my-agent --target-os linux --target-arch amd64

# Specify output path
ayo build ./my-agent --output ./bin/my-agent
```

### `ayo checkit`

Validate configuration and project structure.

```bash
# Validate current project
ayo checkit

# Validate specific project
ayo checkit ./my-agent

# Verbose output
ayo checkit --verbose
```

### `ayo add-agent`

Add agent to existing team project.

```bash
# Add agent to team
ayo add-agent ./my-team reviewer

# With custom settings
ayo add-agent ./my-team security-agent \
  --description "Security analysis" \
  --model "gpt-4-turbo"
```

## Configuration

### `config.toml`

```toml
[agent]
name = "my-agent"              # Agent name
description = "My AI assistant" # Agent description
model = "claude-3-5-sonnet"     # LLM model

[cli]
mode = "hybrid"                # CLI interaction mode
description = "My agent CLI"   # CLI description

[agent.tools]
allowed = ["bash", "file_read", "file_write"]  # Allowed tools

[agent.memory]
enabled = true                  # Enable memory
scope = "agent"                # Memory scope

[input]
schema = ""                    # JSON schema for input (optional)

[output]
schema = ""                    # JSON schema for output (optional)
```

### CLI Modes

- **freeform**: Natural language conversation
- **hybrid**: Mix of structured and freeform
- **structured**: Strict input/output schemas

## Tools

Ayo uses existing CLI programs as tools. No Go code required.

### Built-in Tools

- `bash`: Execute shell commands
- `file_read`: Read file contents
- `file_write`: Write to files
- `git`: Git operations
- `web_search`: Web search (requires API key)

### Custom Tools

Add any executable to the `tools/` directory:

```bash
# Make a script executable
chmod +x tools/my-custom-tool

# Reference in config.toml
[agent.tools]
allowed = ["bash", "file_read", "my-custom-tool"]
```

## Building Agents

### Basic Build

```bash
ayo build my-agent
```

### Cross-Platform Builds

```bash
# Linux AMD64
ayo build my-agent --target-os linux --target-arch amd64

# Windows ARM64
ayo build my-agent --target-os windows --target-arch arm64

# macOS ARM64
ayo build my-agent --target-os darwin --target-arch arm64
```

### Output Options

```bash
# Specify output directory
ayo build my-agent --output ./dist/my-agent

# Build in current directory
ayo build my-agent --output ./my-agent-bin
```

## Running Agents

### Direct Execution

```bash
./my-agent "Analyze this code"
```

### Interactive Mode

```bash
./my-agent --interactive
```

### Structured Input

```bash
./my-agent --input '{"file": "main.go", "task": "review"}'
```

## Advanced Features

### Input/Output Schemas

Define JSON schemas for structured interaction:

```toml
[input]
schema = '''
{
  "type": "object",
  "properties": {
    "file": {"type": "string"},
    "task": {"type": "string"}
  },
  "required": ["file", "task"]
}
'''

[output]
schema = '''
{
  "type": "object",
  "properties": {
    "result": {"type": "string"},
    "score": {"type": "number"}
  }
}
'''
```

### Environment Variables

```bash
# Override model
AYO_MODEL=gpt-4-turbo ./my-agent

# Set API keys
OPENAI_API_KEY=sk-... ./my-agent
ANTHROPIC_API_KEY=sk-... ./my-agent
```

### Team Coordination

For multi-agent projects, use `team.toml`:

```toml
[team]
name = "my-team"
coordination = "sequential"

[agents]
agent1 = { path = "agents/agent1" }
agent2 = { path = "agents/agent2" }
```

## Examples

### Code Review Agent

```bash
# Create agent
ayo fresh code-reviewer --template advanced

# Configure for code analysis
ayo build code-reviewer

# Review code
./code-reviewer --file main.go --task "review"
```

### Document Processing Team

```bash
# Create team
ayo fresh doc-team

# Add agents
ayo add-agent doc-team summarizer
ayo add-agent doc-team translator

# Build team
ayo build doc-team

# Process documents
./doc-team --input documents/ --output processed/
```

### Data Analysis Agent

```bash
# Create agent with structured I/O
ayo fresh data-analyst

# Configure schemas in config.toml
ayo build data-analyst

# Analyze data
./data-analyst --data data.csv --query "find trends"
```

## Troubleshooting

### Build Issues

```bash
# Clean and rebuild
ayo clean
ayo build

# Verbose output
ayo build --verbose
```

### Permission Issues

```bash
chmod +x ./my-agent
```

### Configuration Errors

```bash
ayo checkit --verbose
```

## Migration from Framework

If migrating from the old Ayo framework:

```bash
# Old structure
~/.config/ayo/agents/my-agent/

# New structure
./my-agent/
├── config.toml
└── prompts/
    └── system.md
```

## License

MIT License - See LICENSE file for details.

## Support

- Issues: https://github.com/alexcabrera/ayo/issues
- Documentation: https://github.com/alexcabrera/ayo/docs
- Discussions: https://github.com/alexcabrera/ayo/discussions
