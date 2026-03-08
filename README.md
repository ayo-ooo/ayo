# ayo is a build system for your agentic homies

Use `ayo` to store agents as source code and compile them into standalone executables.

```
my-agent/
├── config.toml          # Agent configuration
├── agents/
│   └── main/
│       ├── config.toml  # Agent-specific config
│       ├── prompts/
│       │   └── system.md # System prompt
│       ├── skills/
│       │   └── custom/
│       │       └── SKILL.md
│       └── tools/
│           └── custom.go # Custom tools
├── workspace/           # Shared workspace
├── team.toml           # Team configuration (if multi-agent)
└── SQUAD.md            # Team constitution (if multi-agent)
```

```bash
# Build your agent
ayo build my-agent

# Execute the compiled agent
./my-agent
```

```bash
# Interactive session with your agent
./my-agent --interactive
```

---

## Agent Projects

Ayo projects store agents as source code in a structured directory format. Each project can contain one or more agents.

### Single-Agent Projects

```
my-agent/
├── config.toml          # Main configuration
└── agents/
    └── main/
        ├── config.toml  # Agent config
        ├── prompts/
        │   └── system.md # System prompt
        ├── skills/
        │   └── custom/
        │       └── SKILL.md
        └── tools/
            └── custom.go # Custom tools
```

### Multi-Agent (Team) Projects

When you add a second agent, ayo automatically promotes your project to a team format:

```
my-team/
├── config.toml          # Main configuration
├── team.toml           # Team configuration
├── SQUAD.md            # Team constitution
├── workspace/           # Shared workspace
└── agents/
    ├── agent1/
    │   ├── config.toml
    │   ├── prompts/
    │   └── system.md
    └── agent2/
        ├── config.toml
        ├── prompts/
        └── system.md
```

## Configuration Files

### `config.toml` - Main Configuration

```toml
[agent]
name = "my-agent"
description = "My AI assistant"
model = "claude-3-5-sonnet"

[cli]
mode = "hybrid"  # freeform, hybrid, or structured
description = "My agent CLI"

[agent.tools]
allowed = ["bash", "file_read", "file_write", "git"]

[agent.memory]
enabled = true
scope = "agent"  # agent or session

[agent.sandbox]
network = false
host_path = ".."

[triggers]
watch = []
schedule = ""
events = []
```

### `team.toml` - Team Configuration

```toml
[team]
name = "my-team"
description = "My agent team"
coordination = "sequential"

[agents]
[agents.agent1]
path = "agents/agent1"

[agents.agent2]
path = "agents/agent2"

[workspace]
shared_path = "workspace"
output_path = "workspace/results"

[coordination]
strategy = "round-robin"
max_iterations = 5
```

### `SQUAD.md` - Team Constitution

```markdown
# Team: my-team

## Mission

[Describe what this team is trying to accomplish]

## Context

[Background information all agents need]

## Agents

### agent1
**Role**: [Define this agent's role]
**Responsibilities**:
- [Responsibility 1]
- [Responsibility 2]

### agent2
**Role**: [Define this agent's role]
**Responsibilities**:
- [Responsibility 1]
- [Responsibility 2]

## Coordination

[How agents should work together]

## Guidelines

[Specific rules for this team]
```

## File System Layout

```
~/.local/share/ayo/
├── cache/              # Remote agent cache
├── builds/             # Built executables
└── projects/           # Project templates

./my-project/
├── config.toml          # Project configuration
├── team.toml           # Team configuration (if multi-agent)
├── SQUAD.md            # Team constitution (if multi-agent)
├── workspace/           # Shared workspace
└── agents/
    └── agent-name/
        ├── config.toml  # Agent configuration
        ├── prompts/     # Prompt templates
        ├── skills/       # Agent skills
        └── tools/        # Custom tools
```

---

## Commands

### `ayo fresh` (alias: `new`, `init`)

Initialize a new agent project.

```bash
# Create a new agent project
ayo fresh my-agent

# With description
ayo fresh my-agent --description "My AI assistant"

# With specific model
ayo fresh my-agent --model claude-3-5-sonnet
```

### `ayo checkit` (alias: `validate`)

Validate agent configuration and project structure.

```bash
# Validate current project
ayo checkit

# Validate specific project
ayo checkit ./my-agent

# Verbose output
ayo checkit --bars
```

### `ayo dunn` (alias: `build`, `done`, `compile`)

Build agent or team executable.

```bash
# Build current project
ayo dunn

# Build specific project
ayo dunn ./my-agent

# Specify output path
ayo dunn ./my-agent --output ./bin/my-agent

# Cross-compile for Linux
ayo dunn ./my-agent --target-os linux --target-arch amd64
```

### `ayo add-agent`

Add an agent to an existing project.

```bash
# Add agent to current project
ayo add-agent reviewer

# Add agent to specific project
ayo add-agent ./my-project reviewer

# With description and model
ayo add-agent ./my-project security-agent \
  --description "Security analysis agent" \
  --model gpt-4-turbo

# Using template
ayo add-agent ./my-project code-analyzer --template advanced
```

### `ayo chat`

Interactive chat with your agent.

```bash
# Chat with agent
ayo chat

# Chat with specific agent
ayo chat ./my-agent

# Chat with team
ayo chat ./my-team
```

### `ayo doctor`

Check system requirements and diagnose issues.

```bash
# Run system checks
ayo doctor

# Check specific component
ayo doctor --check go

# Verbose output
ayo doctor --bars
```

### `ayo flows`

Manage and execute workflows.

```bash
# List available flows
ayo flows list

# Execute a flow
ayo flows run my-flow

# Create new flow
ayo flows new my-flow
```

### `ayo` (default)

Show help and version information.

```bash
# Show help
ayo --help

# Show version
ayo --version

# Verbose output
ayo --bars
```

---

## Configuration Options

### Agent Configuration (`config.toml`)

```toml
[agent]
name = "my-agent"          # Agent name
description = "..."       # Agent description
model = "claude-3-5-sonnet" # LLM model to use

[cli]
mode = "hybrid"            # CLI mode: freeform, hybrid, structured
description = "..."       # CLI description

[agent.tools]
allowed = ["bash", "file_read", "file_write", "git"]  # Allowed tools

[agent.memory]
enabled = true              # Enable memory
scope = "agent"            # Memory scope: agent or session

[agent.sandbox]
network = false            # Network access
host_path = ".."           # Host path mapping

[triggers]
watch = []                  # Files to watch
schedule = ""              # Cron schedule
events = []                # Events to trigger on
```

### Build Options

```bash
# Build for current platform
ayo dunn ./my-agent

# Cross-compile for Linux AMD64
ayo dunn ./my-agent --target-os linux --target-arch amd64

# Build for Windows ARM64
ayo dunn ./my-agent --target-os windows --target-arch arm64

# Specify output directory
ayo dunn ./my-agent --output ./dist/my-agent
```

### Team Coordination

Team projects use `team.toml` to configure how agents work together:

```toml
[coordination]
strategy = "round-robin"    # Coordination strategy
max_iterations = 5          # Maximum iterations

[team]
coordination = "sequential" # Overall coordination mode
```

---

## Workflow

### 1. Create a new agent

```bash
ayo fresh my-agent
cd my-agent
```

### 2. Customize your agent

Edit `config.toml` and `agents/main/config.toml` to configure your agent.

### 3. Add skills and tools

Add custom skills in `agents/main/skills/` and tools in `agents/main/tools/`.

### 4. Build your agent

```bash
ayo dunn
```

### 5. Run your agent

```bash
./my-agent
```

### 6. Add more agents (optional)

```bash
ayo add-agent reviewer
ayo add-agent security-analyst
```

Your project is automatically promoted to team format when you add the second agent.

### 7. Build and run your team

```bash
ayo dunn
./my-team
```

---

## Advanced Features

### Remote Agents

```bash
# Add remote agent from Git repository
ayo add-agent https://github.com/user/agent-repo.git

# Update remote agents
ayo update
```

### Custom Templates

```bash
# Create agent using custom template
ayo fresh my-agent --template advanced

# List available templates
ayo templates list
```

### Cross-Platform Builds

```bash
# Build for multiple platforms
ayo dunn ./my-agent --target-os linux --target-arch amd64
ayodunn ./my-agent --target-os windows --target-arch arm64
ayodunn ./my-agent --target-os darwin --target-arch arm64
```

### Environment Variables

```bash
# Set model via environment variable
AYO_MODEL=gpt-4-turbo ./my-agent

# Set API key
AYO_OPENAI_API_KEY=sk-... ./my-agent
```

---

## Troubleshooting

### Common Issues

**Build fails with missing dependencies:**
```bash
go mod tidy
ayo dunn
```

**Agent not responding:**
```bash
ayo doctor
ayo checkit
```

**Permission issues:**
```bash
chmod +x ./my-agent
./my-agent
```

### Debugging

```bash
# Verbose output
ayo --bars dunn

# Debug specific component
ayo doctor --check go

# Show configuration
ayo config show
```

---

## Examples

### Simple Agent

```bash
# Create and run a simple agent
ayo fresh greeting-agent
ayo dunn greeting-agent
./greeting-agent --prompt "Hello, how are you?"
```

### Code Analysis Team

```bash
# Create code analysis team
ayo fresh code-team
ayo add-agent reviewer
ayo add-agent security-analyst
ayo add-agent performance-analyst
ayo dunn code-team
./code-team --code ./my-project
```

### Document Processing

```bash
# Create document processing agent
ayo fresh doc-processor --template advanced
ayo add-agent summarizer
ayo add-agent translator
ayo dunn doc-processor
./doc-processor --input documents/ --output processed/
```

---

## Migration Guide

### From Ayo Framework to Build System

```bash
# Convert existing framework agents to build system
ayo migrate my-agent

# Build converted agent
ayo dunn my-agent
```

### Project Structure Changes

```
# Old framework structure
~/.config/ayo/agents/my-agent/

# New build system structure
./my-agent/
├── config.toml
└── agents/
    └── main/
        └── config.toml
```

---

## License

MIT License - See LICENSE file for details.

---

## Support

For issues, questions, or contributions:
- GitHub Issues: https://github.com/alexcabrera/ayo/issues
- Documentation: https://github.com/alexcabrera/ayo/docs
- Community: https://github.com/alexcabrera/ayo/discussions
