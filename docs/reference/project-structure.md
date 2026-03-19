# Project Structure

Ayo projects follow a conventional directory structure.

## Minimal Project

The simplest agent requires two files:

```
my-agent/
├── config.toml    # Required: configuration
└── system.md      # Required: system prompt
```

## Full Project

A complete project with all features:

```
my-agent/
├── config.toml           # Configuration
├── system.md             # System prompt
├── input.jsonschema      # Input types
├── output.jsonschema     # Output types
├── prompt.tmpl           # Prompt template
├── skills/               # Skill modules
│   ├── analyze/
│   │   └── SKILL.md
│   └── transform/
│       └── SKILL.md
├── hooks/                # Event hooks
│   ├── agent-start
│   ├── agent-finish
│   ├── agent-error
│   ├── step-start
│   └── step-finish
└── .gitignore           # Ignore generated files
```

## File Details

### config.toml

Agent metadata and model configuration. See [Configuration](config.md).

### system.md

Markdown file containing the system prompt. This defines the agent's behavior and capabilities.

### input.jsonschema

JSON Schema defining input types. When present, Ayo generates CLI flags for each property. See [Input Schema](input-schema.md).

### output.jsonschema

JSON Schema defining output structure. When present, the LLM is instructed to respond with JSON matching this schema. See [Output Schema](output-schema.md).

### prompt.tmpl

Go template for constructing the user prompt. Provides access to input fields and template functions. See [Prompt Templates](prompt-templates.md).

### skills/

Directory containing skill modules. Each skill is a subdirectory with a `SKILL.md` file. See [Skills](skills.md).

### hooks/

Directory containing executable scripts for lifecycle events. See [Hooks](hooks.md).

## Generated Files

When you run `ayo build`, these files are generated:

```
my-agent/
├── main.go              # Entry point
├── cli.go               # CLI parsing (if input.jsonschema)
├── types.go             # Input/output types
├── embed.go             # Embedded files
├── hooks.go             # Hook runner (if hooks/)
├── config-embed.go      # Embedded config
├── go.mod               # Go module file
├── go.sum               # Dependencies
└── my-agent             # Compiled binary
```

Add these to `.gitignore`:

```gitignore
# Generated files
main.go
cli.go
types.go
embed.go
hooks.go
config-embed.go
go.mod
go.sum

# Binary
my-agent
*.json
```

## Build Output

The `ayo build` command:

1. Validates the project structure
2. Parses schemas and configuration
3. Generates Go source files
4. Compiles a standalone binary

The resulting binary:

- Has no runtime dependencies
- Embeds all configuration and prompts
- Can be distributed as a single file
- Runs on the target platform (Linux, macOS, Windows)

## Next Steps

- [Configuration](config.md) - Configuration options
- [Input Schema](input-schema.md) - Define inputs
- [Output Schema](output-schema.md) - Define outputs
