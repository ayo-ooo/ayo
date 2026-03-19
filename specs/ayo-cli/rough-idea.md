# Ayo CLI - Rough Idea

A tool for compiling AI agent definitions into standalone, dependency-free CLI executables backed by the Fantasy multi-provider, multi-model abstraction layer.

## Agent Definition Structure

```
<agent-name>/
  - config.toml        # Required: metadata and configuration
  - system.md          # Required: system message governing agent behavior
  - prompt.tmpl        # Optional: Go template for input-to-prompt rendering
  - input.jsonschema   # Optional: defines CLI inputs and generates arguments
  - output.jsonschema  # Optional: defines structured output schema
  - skills/            # Optional: Agent Skills compatible packages
  - hooks/             # Optional: lifecycle hooks (git-hooks style)
```

## Key Components

### ayo CLI Commands
- `ayo fresh`: Create new agent project from template
- `ayo build`: Compile project into standalone executable
- `ayo checkit`: Validate project integrity before compilation

### Generated Executables
- Fully self-contained (no runtime dependencies)
- First-run model selection via TUI (Bubbletea/Bubbles)
- Configuration stored in `~/.config/agents/<agent-name>.toml`
- Uses Catwalk for provider/model configuration

### Model Selection
- Detect available providers via environment variables
- TUI-based provider and model selection
- Support for local models via Ollama
- Agent specifies model requirements in config.toml

### Non-Interactive First
- All functionality available non-interactively
- Required for agent composition

## Technology Stack
- Go (native compilation)
- Fantasy (multi-provider AI abstraction)
- Catwalk (provider/model registry)
- Bubbletea + Bubbles (TUI)
- Fang + Cobra (CLI)
- Agent Skills standard (skills packages)
