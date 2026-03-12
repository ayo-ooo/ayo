# Ayo - AI Agent Build System

Ayo is a build system for compiling AI agent definitions into standalone, distributable executable binaries.

## Features

- **Pure Build System**: Compiles agent definitions to standalone binaries - no runtime framework required
- **Cross-Platform**: Build for Linux, macOS, Windows, and more
- **Build Caching**: Intelligent caching to avoid rebuilding unchanged agents
- **Development Mode**: Hot reload with automatic rebuilds on file changes
- **Package Management**: Create distributable archives with checksums
- **Version Management**: Semantic versioning with git integration
- **Skills & Tools**: Modular system for agent capabilities

## Installation

### From Source

```bash
git clone https://github.com/ayo-ooo/ayo.git
cd ayo/ayo
go build -o ayo ./cmd/ayo
sudo mv ayo /usr/local/bin/
```

### Using Homebrew (macOS)

```bash
brew install ayo-ooo/tap/ayo
```

## Quick Start

### Create a New Agent

**Important**: After building from source, either:
1. Use `./ayo` from the ayo/ayo directory, OR
2. Run `sudo mv ayo /usr/local/bin/` to install system-wide

```bash
# If ayo is in PATH (after installing):
ayo fresh my-agent

# If using local build from ayo/ayo directory:
./ayo fresh my-agent

cd my-agent
```

### Create a Team Project

```bash
# Create a new team project
ayo fresh my-team
```

### Add Agent to Team

```bash
cd my-team
ayo add-agent . reviewer
```

### Build Your Agent

```bash
# From the agent directory:
ayo build .

# Or if using local build from ayo/ayo:
../../ayo build .
```

### Run Your Agent

```bash
./.build/bin/my-agent
```

### Development Mode

```bash
# Watch for changes and automatically rebuild
ayo dev .

# Watch, rebuild, and run after each build
ayo dev . --run
```

## Commands

### `ayo fresh <name>`

Create a new agent project with a template.

```bash
ayo fresh my-agent
```

Creates:
- `config.toml` - Agent configuration
- `prompts/system.txt` - System prompt
- `prompts/user.txt` - User prompt template
- `skills/` - Custom skills
- `tools/` - Custom tools

### `ayo add-agent <team> <name>`

Add a new agent to an existing team project.

```bash
# Add agent to team
cd my-team
ayo add-agent . reviewer

# Creates config.toml for the new agent in agents/reviewer/
```

### `ayo build <directory>`

Build an agent executable.

```bash
# Build for current platform
ayo build my-agent

# Build with specific output path
ayo build my-agent -o /tmp/my-agent

# Build for specific platform
ayo build my-agent --target-os linux --target-arch amd64

# Build for all platforms
ayo build my-agent --all
```

**Options:**
- `-o, --output <path>` - Output binary path
- `--target-os <os>` - Target operating system (linux, darwin, windows)
- `--target-arch <arch>` - Target architecture (amd64, arm64)
- `--all` - Build for all common platforms

### `ayo dev <directory>`

Development mode with hot reload.

```bash
# Watch and rebuild on changes
ayo dev my-agent

# Watch, rebuild, and run
ayo dev my-agent --run

# Verbose logging
ayo dev my-agent --verbose
```

**Options:**
- `--run` - Run the agent after each build
- `-v, --verbose` - Enable verbose output

### `ayo package <directory>`

Create distributable archives.

```bash
# Package with version from config
ayo package my-agent

# Package with specific version
ayo package my-agent --version 1.0.0

# Specify archive format
ayo package my-agent --format zip
```

**Options:**
- `-v, --version <version>` - Version string
- `-f, --format <format>` - Archive format (tar.gz, zip, auto)

### `ayo release <directory>`

Manage versions and prepare releases.

```bash
# Bump patch version
ayo release my-agent --bump patch

# Bump minor version
ayo release my-agent --bump minor

# Bump major version
ayo release my-agent --bump major

# Create pre-release
ayo release my-agent --bump patch --pre beta
```

**Options:**
- `--bump <type>` - Version part to bump (major, minor, patch)
- `--pre <identifier>` - Pre-release identifier
- `--build <metadata>` - Build metadata

### `ayo checkit <directory>`

Validate agent configuration.

```bash
ayo checkit my-agent
```

### `ayo clean [directory]`

Clean build artifacts and cache.

```bash
# Clean specific agent
ayo clean my-agent

# Clear build cache
ayo clean --cache
```

**Options:**
- `--cache` - Clear the entire build cache

## Configuration

Agent configuration is defined in `config.toml`:

```toml
[agent]
name = "my-agent"
description = "A helpful AI assistant"
version = "1.0.0"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000

[agent.tools]
allowed = ["file_read", "file_write", "web_search"]

[agent.memory]
enabled = true
scope = "session"

[cli]
mode = "freeform"
description = "Interact with the AI assistant"

[cli.flags]
name = "input"
type = "string"
description = "Your question or request"
required = true

[[build.targets]]
os = "linux"
arch = "amd64"

[[build.targets]]
os = "darwin"
arch = "arm64"
```

### Config Sections

**[agent]** - Core agent settings
- `name` - Agent name (required)
- `description` - Agent description (required)
- `version` - Semantic version (optional)
- `model` - Model to use (required): gpt-*, claude-*, o1-*, gemini-*
- `temperature` - Sampling temperature (0.0 - 2.0)
- `max_tokens` - Maximum response tokens

**[agent.tools]** - Tool permissions
- `allowed` - List of allowed tools

**[agent.memory]** - Memory configuration
- `enabled` - Enable memory
- `scope` - Memory scope: "agent" or "session"

**[cli]** - CLI interface
- `mode` - CLI mode: "freeform", "structured", or "hybrid"
- `description` - CLI description
- `flags` - Custom CLI flags

**[build]** - Build configuration
- `targets` - Build targets for cross-platform builds

## Directory Structure

```
my-agent/
├── config.toml          # Agent configuration
├── prompts/             # Prompt templates
│   ├── system.txt       # System prompt
│   └── user.txt         # User prompt template
├── skills/              # Custom skills
│   └── my_skill.go      # Skill implementation
├── tools/               # Custom tools
│   └── my_tool.go       # Tool implementation
├── .build/              # Build output
│   └── bin/             # Compiled binaries
└── releases/            # Packaged releases
```

## Cross-Platform Building

Build for multiple platforms in one command:

```bash
# Build for all common platforms
ayo build my-agent --all
```

Or define custom build targets in `config.toml`:

```toml
[[build.targets]]
os = "linux"
arch = "amd64"

[[build.targets]]
os = "linux"
arch = "arm64"

[[build.targets]]
os = "darwin"
arch = "amd64"

[[build.targets]]
os = "darwin"
arch = "arm64"

[[build.targets]]
os = "windows"
arch = "amd64"
```

## Build Caching

Ayo automatically caches builds to speed up subsequent builds:

- Cache location: `~/.cache/ayo/`
- Cache key: Hash of config, prompts, skills, tools, and target platform
- Clear cache: `ayo clean --cache`

## Version Management

Use semantic versioning with git integration:

```bash
# Bump patch version (1.0.0 -> 1.0.1)
ayo release my-agent --bump patch

# Bump minor version (1.0.0 -> 1.1.0)
ayo release my-agent --bump minor

# Bump major version (1.0.0 -> 2.0.0)
ayo release my-agent --bump major

# Create pre-release (1.0.0-beta)
ayo release my-agent --bump patch --pre beta
```

This updates:
- `config.toml` version field
- Git tag (for non-pre-release versions)
- `CHANGELOG.md` with new section

## Packaging

Create distributable packages:

```bash
# Package for distribution
ayo package my-agent

# Creates:
# releases/my-agent-1.0.0-linux-amd64.tar.gz
# releases/my-agent-1.0.0-darwin-arm64.tar.gz
# releases/my-agent-1.0.0.sha256
```

Verify checksums:

```bash
cd releases
sha256sum -c my-agent-1.0.0.sha256
```

## Development Workflow

1. **Create**: `ayo fresh my-agent`
2. **Develop**: Edit config.toml, prompts, skills, tools
3. **Dev Mode**: `ayo dev . --run` (automatic rebuilds)
4. **Test**: `./.build/bin/my-agent`
5. **Version**: `ayo release . --bump patch`
6. **Build**: `ayo build . --all`
7. **Package**: `ayo package .`
8. **Release**: `git push origin main --tags`

## Examples

### Simple Q&A Agent

```toml
[agent]
name = "qa-bot"
description = "Simple Q&A assistant"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Ask a question"
```

### Structured Data Processing

```toml
[agent]
name = "data-processor"
description = "Process structured data"
model = "gpt-4"

[cli]
mode = "structured"
description = "Process input data"

[cli.flags]
name = "format"
type = "string"
description = "Output format (json, csv)"
required = true
```

### File Operations

```toml
[agent]
name = "file-organizer"
description = "Organize files"
model = "gpt-4"

[agent.tools]
allowed = ["file_read", "file_write", "file_list"]

[cli]
mode = "freeform"
description = "Organize files"
```

## Troubleshooting

### Unknown Command Error (e.g., "Unknown command 'fresh'")

If you get an error like "Unknown command 'fresh'" or a message about sending prompts to agents, you're running an old ayo binary.

```bash
# You're running the old ayo (CLI framework) - check which ayo is being used
which ayo

# Build and use the new ayo (build system)
cd /path/to/ayo/ayo
go build -o ayo ./cmd/ayo

# Test it works
./ayo --version
./ayo --help

# Then use it to create agents
./ayo fresh my-agent

# OR install system-wide
sudo mv ayo /usr/local/bin/
# Now you can use: ayo fresh my-agent
```

### Build Fails

- Ensure Go is installed: `go version`
- Check config.toml is valid: `ayo checkit .`
- Enable verbose output: `ayo build . -v`

### Missing Module Root

```bash
# Build ayo from its source directory
cd /path/to/ayo/ayo
go build -o ayo ./cmd/ayo
```

### Cache Issues

```bash
# Clear the build cache
ayo clean --cache

# Build again
ayo build .
```

## Contributing

Contributions are welcome! Please read our contributing guidelines.

## License

MIT License - see LICENSE file for details.

## Support

- GitHub: https://github.com/ayo-ooo/ayo
- Issues: https://github.com/ayo-ooo/ayo/issues
