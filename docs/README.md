# Ayo Documentation

Build AI-powered CLIs with structured inputs and outputs.

## Overview

Ayo transforms project specifications into standalone command-line agents. Define your agent's behavior through:

- **Schemas**: JSON Schema for inputs and outputs
- **Configuration**: TOML files for agent metadata
- **Prompts**: Markdown system prompts with optional templates
- **Skills**: Reusable skill modules
- **Hooks**: Lifecycle event handlers

## Quick Links

### Getting Started

- [Installation](getting-started/installation.md) - Install Ayo and prerequisites
- [Quickstart](getting-started/quickstart.md) - Build your first agent in 5 minutes
- [First Agent](getting-started/first-agent.md) - Deep dive into agent structure

### Reference

- [Project Structure](reference/project-structure.md) - Directory layout and files
- [Configuration](reference/config.md) - TOML configuration options
- [Input Schema](reference/input-schema.md) - Define CLI inputs
- [Output Schema](reference/output-schema.md) - Structure agent outputs
- [CLI Flags](reference/cli-flags.md) - Customize flag names and positions
- [Prompt Templates](reference/prompt-templates.md) - Dynamic prompt generation
- [Skills](reference/skills.md) - Modular agent capabilities
- [Hooks](reference/hooks.md) - Lifecycle event handlers
- [Generated Code](reference/generated-code.md) - Understanding the output

### Guides

- [Building Agents](guides/building-agents.md) - Best practices for agent design
- [Structured Outputs](guides/structured-outputs.md) - Working with typed responses
- [File Processing](guides/file-processing.md) - Handle file inputs
- [Integrations](guides/integrations.md) - Connect to external services
- [Best Practices](guides/best-practices.md) - Production-ready patterns

### Examples

Browse the [Examples Gallery](examples/README.md) for complete working agents:

| Example | Features Demonstrated |
|---------|----------------------|
| [echo](examples/echo.md) | Minimal agent with string I/O |
| [summarize](examples/summarize.md) | Input and output schemas |
| [translate](examples/translate.md) | Custom CLI flag names |
| [code-review](examples/code-review.md) | File input, enums, integers |
| [research](examples/research.md) | Prompt templates |
| [task-runner](examples/task-runner.md) | Skills directory |
| [notifier](examples/notifier.md) | Hooks directory |
| [data-pipeline](examples/data-pipeline.md) | All features combined |

## How It Works

```
your-agent/
├── config.toml        # Agent configuration
├── system.md          # System prompt
├── input.jsonschema   # Input types (optional)
├── output.jsonschema  # Output types (optional)
├── prompt.tmpl        # Prompt template (optional)
├── skills/            # Skill modules (optional)
└── hooks/             # Event hooks (optional)
```

1. Define your agent with schemas and configuration
2. Run `ayo build` to generate a standalone Go binary
3. Distribute the binary - no runtime dependencies

## Key Features

### Structured I/O

Define inputs and outputs with JSON Schema. Ayo generates type-safe CLIs with proper flag parsing and validation.

### Customizable CLI

Control flag names, positional arguments, short flags, and more through schema extensions.

### Prompt Templates

Create dynamic prompts with template functions for file loading, environment variables, and data interpolation.

### Skills System

Package reusable capabilities as skill modules that agents can invoke.

### Lifecycle Hooks

Execute scripts at key lifecycle events: agent start/finish, step start/finish, errors.

## Getting Help

- Check the [Examples](examples/README.md) for working code
- Read the [Reference](reference/project-structure.md) for detailed specifications
- Open an issue on GitHub for bugs or feature requests
