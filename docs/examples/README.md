# Examples Gallery

Browse complete working examples demonstrating Ayo features.

## Overview

Each example is a complete, buildable agent demonstrating specific features. Examples are located in `examples/` in the Ayo repository.

## Examples

| Example | Features | Complexity |
|---------|----------|------------|
| [echo](#echo) | Basic string I/O | Minimal |
| [status-check](#status-check) | Output-only schema | Basic |
| [summarize](#summarize) | Input + output schemas | Basic |
| [translate](#translate) | Custom CLI flags | Basic |
| [code-review](#code-review) | File input, enums, integers | Intermediate |
| [research](#research) | Prompt templates | Intermediate |
| [task-runner](#task-runner) | Skills directory | Intermediate |
| [notifier](#notifier) | Hooks directory | Intermediate |
| [data-pipeline](#data-pipeline) | All features combined | Advanced |

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
├── input.jsonschema
└── .gitignore
```

**Usage:**
```bash
./echo "Hello, World!"
```

[Full Documentation](echo.md)

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
./summarize "Long text to summarize..."
./summarize input.txt -o summary.json
```

---

## translate

Custom CLI flag names.

**Features:**
- `x-cli-flag` custom flag names
- `x-cli-short` short flags
- Positional arguments
- Default values

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
./translate "Hello" --to spanish
./translate "Bonjour" -s french -t english
```

[Full Documentation](translate.md)

---

## code-review

File handling and complex types.

**Features:**
- `x-cli-file` file content loading
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
./code-review main.go
./code-review app.py --language python --severity error
```

[Full Documentation](code-review.md)

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
./research "quantum computing"
RESEARCH_DEPTH=deep ./research "AI" --context paper.pdf
```

[Full Documentation](research.md)

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
./task-runner "Create a REST API"
./task-runner "Build a CLI tool" --steps 5
```

[Full Documentation](task-runner.md)

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
./notifier "Build complete" --channel slack
./notifier "Alert!" --urgency critical
```

[Full Documentation](notifier.md)

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
./data-pipeline data.json --schema user-schema --output_format csv
PIPELINE_MODE=production ./data-pipeline input.csv --schema report
```

[Full Documentation](data-pipeline.md)

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

## Contributing

To add a new example:

1. Create `examples/<name>/`
2. Add required files (config.toml, system.md)
3. Add optional files (schemas, templates, skills, hooks)
4. Test with `ayo runthat examples/<name>`
5. Add documentation in `docs/examples/<name>.md`
