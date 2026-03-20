# Examples Gallery

Browse complete working examples demonstrating Ayo features.

## Overview

Each example is a complete, buildable agent demonstrating specific features. Examples are located in `examples/` in the Ayo repository.

## Examples

|| Example | Features | Complexity |
|---------|----------|------------|
|| [echo](#echo) | Basic string I/O | Minimal |
|| [status-check](#status-check) | Output-only schema | Basic |
|| [summarize](#summarize) | Input + output schemas | Basic |
|| [translate](#translate) | Custom CLI flags | Basic |
|| [code-review](#code-review) | File input, enums, integers | Intermediate |
|| [research](#research) | Prompt templates | Intermediate |
|| [task-runner](#task-runner) | Skills directory | Intermediate |
|| [notifier](#notifier) | Hooks directory | Intermediate |
|| [data-pipeline](#data-pipeline) | All features combined | Advanced |

---

## echo

Minimal agent with string I/O.

**Features:**
- Basic configuration
- Simple system prompt
- Single string input

**Files:**
```
examples/echo/
├── config.toml
├── system.md
└── .gitignore
```

**Usage:**
```bash
./echo "Hello, World!"
```

---

## status-check

Output-only schema (no inputs).

**Features:**
- Output schema without input
- Model requirements
- Default parameters

**Files:**
```
examples/status-check/
├── config.toml
├── system.md
├── output.jsonschema
└── .gitignore
```

**Usage:**
```bash
./status-check
```

---

## summarize

Input and output schemas with multiple types.

**Features:**
- String input
- Structured JSON output
- Nested objects

**Files:**
```
examples/summarize/
├── config.toml
├── system.md
├── input.jsonschema
├── output.jsonschema
└── .gitignore
```

**Usage:**
```bash
# JSON input
./summarize '{"text": "Long text to summarize..."}'

# Flag override
./summarize --text "Long text to summarize..."

# File input
./summarize input.json
```

---

## translate

Custom CLI flag names.

**Features:**
- Custom flag names with `flag` property
- Default values
- Optional fields

**Files:**
```
examples/translate/
├── config.toml
├── system.md
├── input.jsonschema
└── .gitignore
```

**Usage:**
```bash
# JSON input
./translate '{"text": "Hello", "to": "spanish"}'

# Flag overrides
./translate --text "Hello" --to spanish

# Stdin
echo '{"text": "Hello"}' | ./translate - --to spanish
```

---

## code-review

File handling and complex types.

**Features:**
- `file` property for file content loading
- Enum constraints
- Integer types
- Array output

**Files:**
```
examples/code-review/
├── config.toml
├── system.md
├── input.jsonschema
├── output.jsonschema
└── .gitignore
```

**Usage:**
```bash
# JSON input
./code-review '{"file": "main.go"}'

# Flag overrides
./code-review --file main.go --language python --severity error
```

---

## research

Prompt templates with functions.

**Features:**
- `prompt.tmpl` template file
- `file` function for loading content
- `env` function for environment variables
- Conditional logic

**Files:**
```
examples/research/
├── config.toml
├── system.md
├── prompt.tmpl
├── input.jsonschema
├── output.jsonschema
└── .gitignore
```

**Usage:**
```bash
# JSON input
./research '{"topic": "quantum computing"}'

# With environment variable
RESEARCH_DEPTH=deep ./research --topic "AI"
```

---

## task-runner

Skills directory with multiple skills.

**Features:**
- Skills directory structure
- Multiple skill modules
- Skill descriptions in system prompt

**Files:**
```
examples/task-runner/
├── config.toml
├── system.md
├── input.jsonschema
├── output.jsonschema
├── skills/
│   ├── plan/SKILL.md
│   ├── execute/SKILL.md
│   └── review/SKILL.md
└── .gitignore
```

**Usage:**
```bash
# JSON input
./task-runner '{"task": "Create a REST API"}'

# Flag overrides
./task-runner --task "Build a CLI tool"
```

---

## notifier

Hooks for lifecycle events.

**Features:**
- Hooks directory
- Multiple hook types
- Shell script hooks

**Files:**
```
examples/notifier/
├── config.toml
├── system.md
├── input.jsonschema
├── hooks/
│   ├── agent-start
│   ├── agent-finish
│   └── agent-error
└── .gitignore
```

**Usage:**
```bash
# JSON input
./notifier '{"message": "Build complete", "channel": "slack"}'

# Flag overrides
./notifier --message "Alert!" --urgency critical
```

---

## data-pipeline

Comprehensive example with all features.

**Features:**
- All input types (string, integer, boolean, enum)
- Complex nested schemas
- Skills + Hooks
- Prompt templates
- Model requirements
- Default parameters

**Files:**
```
examples/data-pipeline/
├── config.toml
├── system.md
├── prompt.tmpl
├── input.jsonschema
├── output.jsonschema
├── skills/
│   ├── extract/SKILL.md
│   ├── transform/SKILL.md
│   └── validate/SKILL.md
├── hooks/
│   ├── agent-start
│   ├── step-start
│   ├── step-finish
│   └── agent-finish
└── .gitignore
```

**Usage:**
```bash
# JSON input
./data-pipeline '{"source": "data.json", "target_schema": "user-schema"}'

# Flag overrides
./data-pipeline --source data.json --schema user-schema --output-format csv

# With environment variable
PIPELINE_MODE=production ./data-pipeline --source input.csv
```

---

## Building Examples

```bash
# Build any example
cd examples/<name>
ayo runthat .

# Or from the repository root
ayo runthat examples/<name>

# Run
./<name> --help
```

## Input Patterns

All examples support JSON input as the primary input mechanism:

```bash
# Inline JSON
./agent '{"field": "value"}'

# From file
./agent input.json

# From stdin
echo '{"field": "value"}' | ./agent -

# Flag overrides (combine with JSON or use alone)
./agent --field value
```

## Contributing

To add a new example:

1. Create `examples/<name>/`
2. Add required files (config.toml, system.md)
3. Add optional files (schemas, templates, skills, hooks)
4. Test with `ayo runthat examples/<name>`
5. Update this gallery documentation
