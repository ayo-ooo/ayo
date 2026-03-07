# ayo – Build Standalone AI Agents

**ayo** is a build system for creating standalone, executable AI agents. Define agents as projects, compile them into single binaries, and distribute them like any CLI tool.

## Quick Start

```bash
# Install
go install github.com/alexcabrera/ayo/cmd/ayo@latest

# Initialize a new agent
ayo init myagent

# Validate configuration
ayo checkit myagent

# Build standalone executable
ayo build myagent

# Run your agent
./myagent
```

## What Can You Build?

- **CLI Agents**: Command-line tools for specific tasks (code review, testing, documentation)
- **Service Agents**: Background services with file watching or scheduling
- **Team Agents**: Multi-agent tools with specialized roles and coordination
- **Integration Agents**: Connect systems and automate workflows

## Getting Started

### 1. Create an Agent

```bash
# Standard template (recommended)
ayo init myreviewer --template standard

# Simple template (minimal)
ayo init myagent --template simple

# Advanced template (full features)
ayo init myagent --template advanced
```

### 2. Configure Your Agent

Edit the generated `config.toml`:

```toml
[agent]
name = "myagent"
description = "AI-powered code reviewer"
model = "claude-3-5-sonnet"

[cli]
mode = "hybrid"
description = "Review code for security and quality"

[cli.flags]
severity = { type = "string", description = "Severity level to filter" }

[agent.tools]
allowed = ["bash", "file_read", "git"]
```

### 3. Validate Configuration

```bash
ayo checkit myagent --verbose
```

### 4. Build Executable

```bash
# Build for current platform
cd myagent
ayo build .

# Or specify directory
ayo build myagent

# Cross-compile for other platforms
ayo build myagent --target-os linux --target-arch amd64
```

This produces a standalone binary you can run anywhere.

## Configuration

Your agent is defined entirely in `config.toml`:

- **Agent metadata**: Name, description, model
- **CLI interface**: Flags, modes, commands
- **Input/output schemas**: JSON Schema validation
- **Tool access**: Which tools the agent can use
- **Memory settings**: Persistent context storage
- **Triggers**: File watching, scheduling, events
- **Testing**: Evals configuration for automated testing

See [Build System Documentation](docs/BUILD_SYSTEM.md) for complete configuration reference.

## Testing Your Agent

### Validation

```bash
# Check configuration syntax and structure
ayo checkit myagent

# Detailed output
ayo checkit myagent --verbose
```

### Automated Evals

Define test cases in `evals.csv` and run automated LLM-based evaluation:

```bash
# Run evals with scoring
ayo checkit --evals-only myagent

# Custom threshold (default: 7.0/10)
ayo checkit --evals-threshold 8.0 myagent

# See detailed reasoning
ayo checkit --evals-only --verbose myagent
```

See [Build System - Evals](docs/BUILD_SYSTEM.md#evals-automated-testing) for details.

## Distribution

Your built agents are standalone executables with no ayo dependency:

```bash
# Package for distribution
tar -czf myagent-v1.0.tar.gz myagent README.md

# Users install by extracting and running
tar -xzf myagent-v1.0.tar.gz
chmod +x myagent
./myagent --help

# Or install system-wide
sudo cp myagent /usr/local/bin/
myagent
```

## Advanced Features

### Multi-Agent Teams

Coordinate multiple specialized agents using `team.toml`:

```bash
# Build team executable
ayo build --team myproject

# Team dispatches work to best agent for the task
./myteam "build authentication system"
```

### Custom Tools

Write Go tools in the `tools/` directory for agent-specific functionality:

```go
package tools

import (
    "context"
    "fmt"
)

func AnalyzeCode(ctx context.Context, params struct {
    Language string `json:"language"`
    Code     string `json:"code"`
}) (map[string]any, error) {
    // Your custom logic
    return map[string]any{
        "issues": []string{"potential security flaw"},
    }, nil
}
```

### Input/Output Schemas

Define JSON schemas for type-safe agent interfaces:

```toml
[input]
file = "schema/input.json"

[output]
file = "schema/output.json"
```

## Supported Providers

ayo works with all major LLM providers:

| Provider | Environment Variable | Notes |
|----------|---------------------|-------|
| Anthropic | `ANTHROPIC_API_KEY` | Claude models |
| OpenAI | `OPENAI_API_KEY` | GPT-4 models |
| Google | `GEMINI_API_KEY` | Gemini models |
| OpenRouter | `OPENROUTER_API_KEY` | Multi-provider gateway |
| Azure | `AZURE_OPENAI_API_KEY` | Azure OpenAI |
| Ollama | *(none required)* | Local models |
| ...and 6 more | | See `ayo setup` |

Run `ayo setup` for interactive provider configuration.

## Documentation

### For Users

- **[Build System Guide](docs/BUILD_SYSTEM.md)** – Complete configuration reference
- **[Getting Started](docs/getting-started.md)** – First steps with ayo
- **[Tutorials](docs/tutorials/)** – Step-by-step guides

### For Contributors

Documentation in `./docs/` is for people working on ayo itself:

- **[Architecture](docs/concepts.md)** – System design
- **[Patterns](docs/patterns/)** – Implementation patterns
- **[Reference](docs/reference/)** – API and config schemas
- **[Advanced](docs/advanced/)** – Internals and extending ayo

## Requirements

- **Go 1.24+** (for building from source)
- At least one LLM provider configured
- Optional: macOS 26+ (Tahoe) or Linux with systemd (for squad features)

## License

MIT – see LICENSE file

## Contributing

Contributions welcome! See `./docs` for contributor documentation.
