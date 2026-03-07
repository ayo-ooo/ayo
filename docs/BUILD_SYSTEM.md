# Ayo Build System

Ayo has been transformed from a CLI framework for managing agents into a **pure build system** for creating standalone executable agents and teams.

## Overview

With the new build system, you can:

- Define agents as repository projects (not just `~/.config/ayo/agents/`)
- Compile agent definitions into single executable binaries
- Run agents independently without the `ayo` CLI prefix
- Distribute agents as standalone binaries like any CLI tool

## Quick Start

### 1. Initialize a new agent project

```bash
# Create a new agent with the standard template
ayo init myreviewer --template standard

# Or use simple/advanced templates
ayo init myagent --template simple
ayo init myagent --template advanced
```

This creates:
```
myreviewer/
├── config.toml          # Agent configuration
├── skills/              # Agent-specific skills
├── tools/               # Custom Go tools
└── prompts/             # Prompt templates
    └── system.md        # System prompt
```

### 2. Customize your agent

Edit `config.toml` to define:
- Agent name, description, and model
- CLI interface (flags, modes)
- Input/output schemas
- Tool access
- Memory settings
- Sandbox configuration

### 3. Validate configuration

```bash
ayo validate myreviewer
```

### 4. Build the executable

```bash
# Build in current directory
cd myreviewer
ayo build .

# Or specify directory
ayo build myreviewer

# Build for specific platform
ayo build myreviewer --target-os linux --target-arch amd64
```

This produces a standalone binary: `myreviewer`

### 5. Run your agent

```bash
# Run directly
./myreviewer

# Or move to PATH and run from anywhere
sudo mv myreviewer /usr/local/bin/
myreviewer
```

## Configuration (config.toml)

The `config.toml` file contains all agent configuration in a single, human-readable format.

### Basic Structure

```toml
[agent]
name = "myreviewer"
description = "Security-focused code reviewer"
model = "claude-3-5-sonnet"

[cli]
mode = "hybrid"
description = "Review code for security issues"

[cli.flags]

[agent.tools]
allowed = ["bash", "file_read", "file_write", "git"]

[agent.memory]
enabled = true
scope = "agent"

[agent.sandbox]
network = false
host_path = "."

[triggers]
watch = []
schedule = ""
events = []
```

### Sections

#### [agent] - Agent Identity

- `name` (string, required): Agent name (used for binary name)
- `description` (string, required): Agent description
- `model` (string, required): Default LLM model

#### [agent.tools] - Tool Access

- `allowed` (array of strings): Tools the agent can use

Available tools:
- `bash` - Execute shell commands
- `file_read` - Read files
- `file_write` - Write files
- `git` - Git operations
- `web_search` - Search the web (requires network access)

#### [agent.memory] - Memory Settings

- `enabled` (bool): Enable agent memory
- `scope` (string): Memory scope - "agent" | "session" | "global"

#### [agent.sandbox] - Sandbox Configuration

- `network` (bool): Enable network access
- `host_path` (string): Host filesystem path to mount

#### [cli] - Command-Line Interface

- `mode` (string): CLI mode - "structured" | "freeform" | "hybrid"
- `description` (string): CLI description shown in help
- `flags` (map): CLI flag definitions (see below)

#### [cli.flags] - CLI Flags

Define individual CLI flags:

```toml
[cli.flags]
  [cli.flags.repo]
  name = "repo"
  type = "string"
  short = "r"
  position = 0
  required = true
  description = "Repository path"

  [cli.flags.files]
  name = "files"
  type = "string"
  short = "f"
  multiple = true
  position = 1
  description = "Files to review"
```

Flag properties:
- `name` (string, required): Flag name
- `type` (string, required): Type - "string" | "int" | "float" | "bool" | "array"
- `short` (string): Short flag (e.g., "r" for -r)
- `position` (int): Positional argument index (>= 0)
- `required` (bool): Whether flag is required
- `multiple` (bool): Accept multiple values (for arrays)
- `description` (string): Flag description
- `default` (any): Default value

#### [input] - Input Schema

Define expected input format using JSON Schema:

```toml
[input]
[input.schema]
type = "object"
required = ["repo"]

[input.schema.properties]
  [input.schema.properties.repo]
  type = "string"

  [input.schema.properties.files]
  type = "array"
  items = { type = "string" }
```

#### [output] - Output Schema

Define expected output format using JSON Schema:

```toml
[output]
[output.schema]
type = "object"
required = ["issues", "summary"]

[output.schema.properties]
  [output.schema.properties.issues]
  type = "array"

  [output.schema.properties.summary]
  type = "string"
```

#### [prompts] - Prompt Templates

```toml
[prompts]
system = "You are a security expert..."
user = "Please review: {input}"
```

#### [triggers] - Automatic Execution

```toml
[triggers]
watch = ["./src/**/*.go", "./pkg/**/*.go"]
schedule = "0 * * * *"  # Every hour
events = ["file_created", "file_modified"]
```

## CLI Modes

### Structured Mode

Only structured flags are accepted:

```toml
[cli]
mode = "structured"
```

Usage:
```bash
myagent --repo . --files main.go,auth.go
```

### Freeform Mode

Only freeform text prompts are accepted:

```toml
[cli]
mode = "freeform"
```

Usage:
```bash
myagent "review main.go for security issues"
```

### Hybrid Mode (Recommended)

Both structured flags and freeform text:

```toml
[cli]
mode = "hybrid"
```

Usage:
```bash
# Structured only
myagent --repo . --files main.go

# Freeform only
myagent "review main.go"

# Mixed
myagent --repo . "review main.go"
```

## Templates

### Simple Template

Minimal configuration:
- Freeform CLI mode
- Basic tools (bash, file_read, file_write)
- Agent-scoped memory
- No network access

### Standard Template (Recommended)

Balanced configuration:
- Hybrid CLI mode
- Extended tools (includes git)
- Agent-scoped memory
- No network access
- Example skill included

### Advanced Template

Full-featured configuration:
- Structured CLI mode
- All tools (includes web_search)
- Session-scoped memory
- Network access enabled

## Project Structure

### Minimal Agent

```
myagent/
└── config.toml
```

### Standard Agent

```
myagent/
├── config.toml
├── skills/
│   └── custom/
│       └── SKILL.md
├── tools/
│   └── mytool.go
└── prompts/
    └── system.md
```

### Multi-Agent Team

```
myteam/
├── team.toml
└── workspace/
    └── shared-workspace
```

## Commands

### ayo init

Initialize a new agent project:

```bash
ayo init [directory] [flags]

Flags:
  -d, --description string   Agent description
  -m, --model string        Default model (default: claude-3-5-sonnet)
  --template string          Template: simple, standard, advanced (default: standard)
  --name string              Agent name (default: directory name)
```

### ayo validate

Validate agent configuration:

```bash
ayo validate [directory] [flags]

Flags:
  -v, --verbose    Show detailed validation output
```

### ayo build

Build standalone executable:

```bash
ayo build [directory] [flags]

Flags:
  -o, --output string      Output binary path (default: <agent-name>)
  --target-os string       Target OS (default: current OS)
  --target-arch string     Target architecture (default: current arch)
```

## Examples

### Code Reviewer Agent

config.toml:
```toml
[agent]
name = "myreviewer"
description = "Security-focused code reviewer"
model = "claude-3-5-sonnet"

[cli]
mode = "hybrid"
description = "Review code for security issues"

[cli.flags]
  [cli.flags.repo]
  name = "repo"
  type = "string"
  short = "r"
  position = 0
  required = true
  description = "Repository path"

  [cli.flags.files]
  name = "files"
  type = "array"
  short = "f"
  description = "Files to review"

[agent.tools]
allowed = ["bash", "file_read", "file_write", "git"]

[agent.memory]
enabled = true
scope = "session"

[agent.sandbox]
network = false
host_path = "."
```

Usage:
```bash
# Build
ayo build myreviewer

# Run with structured flags
./myreviewer --repo . --files main.go auth.go

# Run with freeform
./myreviewer "review main.go for SQL injection vulnerabilities"

# Mixed
./myreviewer --repo . "review auth.go"
```

### File Transformer Agent

config.toml:
```toml
[agent]
name = "transform"
description = "Transform and process files"
model = "gpt-4"

[cli]
mode = "structured"
description = "Transform files according to rules"

[cli.flags]
  [cli.flags.input]
  name = "input"
  type = "string"
  short = "i"
  required = true
  description = "Input file"

  [cli.flags.output]
  name = "output"
  type = "string"
  short = "o"
  required = true
  description = "Output file"

  [cli.flags.format]
  name = "format"
  type = "string"
  default = "json"
  description = "Output format"

[agent.tools]
allowed = ["bash", "file_read", "file_write"]

[input]
[input.schema]
type = "object"
required = ["input", "output"]
```

Usage:
```bash
# Build
ayo build transform

# Run
./transform --input data.json --output data.yml --format yaml
```

## Distribution

### Single Binary

```bash
# Build
ayo build myagent

# Distribute
tar -czf myagent.tar.gz myagent
scp myagent.tar.gz user@server:
```

### Cross-Platform Builds

```bash
# Linux AMD64
ayo build myagent --target-os linux --target-arch amd64 -o myagent-linux-amd64

# macOS ARM64
ayo build myagent --target-os darwin --target-arch arm64 -o myagent-darwin-arm64

# Windows AMD64
ayo build myagent --target-os windows --target-arch amd64 -o myagent.exe
```

## Migration from Old Ayo

The build system is a breaking change from the previous directory-based agent management. Here's how to migrate:

### Old Workflow

```bash
# Create agent in ~/.config/ayo/agents/
ayo agent create myreviewer

# Run with ayo prefix
ayo @myreviewer "review this code"
```

### New Workflow

```bash
# Create as standalone project
ayo init myreviewer

# Build executable
cd myreviewer
ayo build .

# Run directly
./myreviewer "review this code"
```

### Converting Existing Agents

1. Copy agent directory to new location
2. Convert `config.json`, `input.jsonschema`, `output.jsonschema` to `config.toml`
3. Add CLI configuration section
4. Build new executable
5. Remove old agent: `ayo agent rm myreviewer`

## Advanced Topics

### Custom Tools

Create custom Go tools in the `tools/` directory:

```go
// tools/mytool.go
package main

import "fmt"

func MyCustomTool(input string) (string, error) {
    // Custom logic
    return result, nil
}
```

### Custom Skills

Create agent-specific skills in `skills/`:

```
skills/
└── custom/
    └── SKILL.md
```

SKILL.md:
```markdown
# Custom Skill

## Behavior

When processing requests:

1. Analyze the context
2. Apply custom logic
3. Provide tailored responses
```

### Environment Variables

Runtime configuration via environment variables:

```bash
AYO_MODEL=gpt-4 ./myagent "task"
AYO_DEBUG=1 ./myagent "task"
AYO_DATA_DIR=/custom/path ./myagent "task"
```

## Troubleshooting

### Build Fails

```bash
# Validate configuration first
ayo validate myagent --verbose

# Check for syntax errors
# Common issues:
# - Missing required fields
# - Invalid JSON Schema syntax
# - Unknown flag types
```

### Binary Won't Run

```bash
# Make executable
chmod +x myagent

# Check architecture
file myagent
# Should match your system (e.g., x86_64 for Linux AMD64)

# Test basic execution
./myagent --help
```

### Configuration Not Loaded

```bash
# Ensure config.toml is present
ls myagent/config.toml

# Validate syntax
ayo validate myagent
```

## Best Practices

1. **Start with simple template**: Use `--template simple` for basic agents
2. **Validate frequently**: Run `ayo validate` before building
3. **Test locally**: Build and test before distributing
4. **Use semantic versioning**: Version your agent releases
5. **Document custom tools**: Add README in tools/ directory
6. **Keep configs readable**: Use comments in TOML files
7. **Test CLI modes**: Verify all three modes work (if using hybrid)
8. **Check schemas**: Validate input/output schemas before building
9. **Test on target platform**: Build for the platform you'll run on
10. **Use environment variables**: For runtime configuration

## Evals (Automated Testing)

The build system includes support for automated testing of agent outputs using LLM-based evaluation (evals).

### Overview

Evals allow you to:
- Define test cases with expected outputs
- Run agent against test inputs
- Use an LLM to score actual vs expected output
- Get pass/fail results with detailed reasoning

### Configuration

Enable evals in your `config.toml`:

```toml
[evals]
enabled = true
file = "evals.csv"
judge_model = "claude-3-5-sonnet"
criteria = "Evaluate correctness, completeness, and clarity of the response"
```

### CSV Format

Create `evals.csv` with the following columns:

| Column | Required | Description |
|--------|----------|-------------|
| description | No | Human-readable test case name |
| input | Yes | JSON input to agent |
| expected | Yes | JSON expected output |
| criteria | No | Override default criteria for this test |

Example `evals.csv`:

```csv
description,input,expected,criteria
Simple addition,"{\"a\": 1, \"b\": 2}","{\"sum\": 3}","Check arithmetic correctness"
Error handling,"{\"operation\": \"divide\", \"a\": 1, \"b\": 0}","{\"error\": \"division by zero\"}","Handle invalid operations"
```

### Running Evals

```bash
# Validate and run evals
ayo checkit myagent

# Run only evals (skip other validation)
ayo checkit --evals-only myagent

# Use custom threshold (default: 7.0)
ayo checkit --evals-threshold 8.0 myagent

# Show detailed reasoning for each test
ayo checkit --verbose myagent
```

### Interpreting Results

Each test receives a score from 0-10:
- **10**: Perfect match, completely correct
- **7-9**: Minor differences, functionally correct
- **4-6**: Partially correct, significant issues
- **1-3**: Major errors, mostly incorrect
- **0**: Completely incorrect or failed

Exit codes:
- **0**: All tests passed (score >= threshold)
- **1**: One or more tests failed
- **2**: Configuration or CSV parsing error

### Notes

- The judge uses your configured LLM provider and model
- Tests run in-process; actual agent execution is a placeholder (returns mock results)
- For accurate results, ensure your judge model is reliable and unbiased
- Use `--verbose` flag to see the judge's reasoning for each test

## FAQ

**Q: Can I still use the old `ayo @agent` syntax?**

A: No, agents are now standalone executables. Use `./agent` instead.

**Q: What happened to squads?**

A: Multi-agent teams are supported via `team.toml` files. Use `ayo build --team`.

**Q: Can I have multiple agents in one project?**

A: No, each agent should have its own project with a `config.toml`. Use teams for multi-agent coordination.

**Q: How do I update an agent?**

A: Edit `config.toml` and rebuild: `ayo build .`

**Q: Can I distribute agents as source?**

A: Yes, distribute the entire project directory and let users build themselves.

**Q: What about the daemon?**

A: Not needed for standalone agents. Agents manage their own lifecycle.

**Q: Can I use plugins?**

A: Plugins are not yet supported in the build system. Use custom tools instead.

**Q: How do I debug a built agent?**

A: Use `AYO_DEBUG=1` environment variable and check logs.

## Future Enhancements

Planned features:

- [ ] Multi-agent team orchestration
- [ ] Plugin system for custom tool distributions
- [ ] Agent marketplace/sharing
- [ ] Automatic dependency management
- [x] Integrated testing framework (evals - implemented!)
- [ ] Versioning and update system
- [ ] Remote agent execution
- [ ] Agent composition and chaining
- [ ] Performance profiling tools
- [ ] Visual configuration editor
- [ ] Full agent execution in evals (currently uses placeholder)
