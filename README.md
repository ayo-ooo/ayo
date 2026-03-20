# Ayo

Build AI-powered CLIs in minutes, not hours.

Ayo transforms simple project specifications into standalone command-line agents. Define your agent's behavior with schemas and prompts, then compile it into a distributable binary — no runtime dependencies.

## The Basics

Create a project, define your agent, build it:

```bash
# Scaffold a new agent
ayo fresh my-agent && cd my-agent

# Define behavior (edit system.md)
echo "You are a helpful coding assistant." > system.md

# Build and run
ayo runthat .
./my-agent "Write a haiku about recursion"
```

That's it. You now have a standalone CLI that calls an LLM with your system prompt.

## Installation

**Homebrew:**
```bash
brew install ayo-ooo/tap/ayo
```

**Go install:**
```bash
go install github.com/ayo-ooo/ayo/cmd/ayo@latest
```

**Prerequisites:**
- Go 1.21+ (for building generated agents)

Configure your API key for your preferred provider:

| Environment Variable        | Provider                                           |
| --------------------------- | -------------------------------------------------- |
| `ANTHROPIC_API_KEY`         | Anthropic                                          |
| `OPENAI_API_KEY`            | OpenAI                                             |
| `GEMINI_API_KEY`            | Google Gemini                                      |
| `GROQ_API_KEY`              | Groq                                               |
| `OPENROUTER_API_KEY`        | OpenRouter                                         |
| `CEREBRAS_API_KEY`          | Cerebras                                           |
| `HF_TOKEN`                  | Hugging Face Inference                             |

## Tutorial: Building a Translation Agent

Let's build a translation CLI that takes text and a target language.

### 1. Create the Project

```bash
ayo fresh translate && cd translate
```

### 2. Define Input Schema

Create `input.jsonschema` to define what the agent accepts:

```json
{
  "type": "object",
  "properties": {
    "text": {
      "type": "string",
      "description": "Text to translate",
      "x-cli-position": 1
    },
    "to": {
      "type": "string",
      "description": "Target language",
      "x-cli-short": "-t"
    }
  },
  "required": ["text", "to"]
}
```

The `x-cli-position` makes `text` a positional argument. The `x-cli-short` adds `-t` as a shorthand.

### 3. Define Output Schema (Optional)

Create `output.jsonschema` to structure responses:

```json
{
  "type": "object",
  "properties": {
    "translation": { "type": "string" },
    "detected_language": { "type": "string" }
  }
}
```

### 4. Write the System Prompt

Edit `system.md`:

```markdown
You are a professional translator.

Translate the given text to the target language.
Detect the source language automatically.
Respond with accurate, natural-sounding translations.
```

### 5. Build and Run

```bash
ayo runthat .
./translate "Hello, world!" -t spanish
```

Output:
```json
{
  "translation": "¡Hola, mundo!",
  "detected_language": "english"
}
```

## Project Structure

```
my-agent/
├── config.toml        # Agent metadata and model settings
├── system.md          # System prompt (required)
├── input.jsonschema   # Define inputs and CLI flags
├── output.jsonschema  # Structure responses (optional)
├── prompt.tmpl        # Dynamic prompt template (optional)
├── skills/            # Reusable skill modules (optional)
└── hooks/             # Lifecycle event handlers (optional)
```

## Commands

| Command | Description |
|---------|-------------|
| `ayo fresh <name>` | Create a new agent project |
| `ayo runthat [path]` | Compile agent into standalone executable |
| `ayo checkit [path]` | Validate an agent project |
| `ayo --version` | Show version |
| `ayo --help` | Show help |

## Input Schema

Define inputs using JSON Schema with CLI extensions:

```json
{
  "type": "object",
  "properties": {
    "file": {
      "type": "string",
      "description": "File to process",
      "x-cli-position": 1,
      "x-cli-file": true
    },
    "format": {
      "type": "string",
      "description": "Output format",
      "enum": ["json", "yaml", "text"],
      "default": "text",
      "x-cli-short": "-f"
    },
    "verbose": {
      "type": "boolean",
      "description": "Enable verbose output",
      "default": false
    }
  },
  "required": ["file"]
}
```

**CLI Extensions:**

| Extension | Purpose |
|-----------|---------|
| `x-cli-position` | Make a positional argument (1-indexed) |
| `x-cli-flag` | Custom flag name |
| `x-cli-short` | Short flag (e.g., `-f`) |
| `x-cli-file` | Load file content into field |

Generates:
```bash
./agent <file> [-f json] [--verbose]
```

## Output Schema

Structure LLM responses with JSON Schema:

```json
{
  "type": "object",
  "properties": {
    "summary": { "type": "string" },
    "key_points": {
      "type": "array",
      "items": { "type": "string" }
    },
    "confidence": {
      "type": "number",
      "minimum": 0,
      "maximum": 1
    }
  }
}
```

The LLM is instructed to respond with JSON matching this schema.

## Prompt Templates

Create dynamic prompts with `prompt.tmpl`:

```gotemplate
Analyze the following {{.format}} file:

{{file .file}}

{{if .verbose}}Provide detailed analysis with line numbers.{{end}}
```

**Template Functions:**

| Function | Description |
|----------|-------------|
| `{{.field}}` | Access input field |
| `{{file "path"}}` | Load file contents |
| `{{env "VAR"}}` | Get environment variable |
| `{{upper .text}}` | Uppercase |
| `{{lower .text}}` | Lowercase |
| `{{json .data}}` | JSON encode |

## Skills

Package reusable capabilities in `skills/`:

```
skills/
└── analyze/
    └── SKILL.md
```

Skills are injected into the system prompt, giving the LLM additional capabilities it can invoke.

## Hooks

Execute scripts at lifecycle events:

```
hooks/
├── agent-start      # Before agent runs
├── agent-finish     # After agent completes
├── agent-error      # On error
├── step-start       # Before each step
└── step-finish      # After each step
```

Hooks receive context via environment variables and can modify behavior.

## Examples

The [examples/](examples/) directory contains complete working agents:

| Example | Features |
|---------|----------|
| [echo](examples/echo/) | Minimal agent |
| [translate](examples/translate/) | Input/output schemas, custom flags |
| [summarize](examples/summarize/) | Structured output |
| [code-review](examples/code-review/) | File input, enums, integers |
| [research](examples/research/) | Prompt templates |
| [task-runner](examples/task-runner/) | Skills directory |
| [notifier](examples/notifier/) | Hooks directory |
| [data-pipeline](examples/data-pipeline/) | All features combined |

## What Happens Under the Hood

When you run `ayo runthat`:

1. Reads your `config.toml`, `system.md`, and schemas
2. Generates a Go program with:
   - CLI flag parsing based on `input.jsonschema`
   - Type-safe input validation
   - LLM client configured for your provider
   - Structured output parsing if `output.jsonschema` exists
3. Compiles it into a single static binary
4. No runtime dependencies — just distribute the executable

## Documentation

- [Installation](docs/getting-started/installation.md)
- [First Agent](docs/getting-started/first-agent.md)
- [Input Schema Reference](docs/reference/input-schema.md)
- [Output Schema Reference](docs/reference/output-schema.md)
- [CLI Flags](docs/reference/cli-flags.md)
- [Prompt Templates](docs/reference/prompt-templates.md)
- [Skills](docs/reference/skills.md)
- [Hooks](docs/reference/hooks.md)

## Contributing

We welcome contributions! See the [examples/](examples/) directory for patterns and conventions.

## Feedback

- Open an issue for bugs or feature requests
- Start a discussion for questions

## License

[MIT](LICENSE)
