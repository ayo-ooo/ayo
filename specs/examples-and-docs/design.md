# Example Projects & Documentation Suite Design

## Overview

This design specifies a comprehensive set of example Ayo projects that demonstrate all framework features, paired with a documentation suite that uses these examples as teaching tools. The examples progress from minimal to complex, each focusing on specific features.

## Example Projects

### Project 1: `echo` (Minimal)

**Purpose**: Absolute simplest working agent - text in, text out.

**Features Demonstrated**:
- Basic project structure (config.toml, system.md)
- No input schema (freeform text input)
- No output schema (text response)
- No CLI flags beyond defaults

**Files**:
```
examples/echo/
├── config.toml
├── system.md
└── .gitignore
```

**config.toml**:
```toml
[agent]
name = "echo"
version = "1.0.0"
description = "Echoes back what you say with enthusiasm"

[model]
suggested = ["anthropic/claude-3.5-sonnet", "openai/gpt-4o"]

[defaults]
temperature = 0.7
```

**Usage**:
```bash
echo "Hello world"
# Output: HELLO WORLD! 🎉
```

---

### Project 2: `status-check` (No Input / Autonomous)

**Purpose**: Demonstrates an agent that runs autonomously without any input - uses hardcoded prompt to perform actions.

**Features Demonstrated**:
- No input schema (no input.jsonschema)
- Hardcoded behavior in system.md or prompt.tmpl
- Zero-argument execution
- Useful for scheduled/automated tasks

**Files**:
```
examples/status-check/
├── config.toml
├── system.md
├── output.jsonschema
└── .gitignore
```

**config.toml**:
```toml
[agent]
name = "status-check"
version = "1.0.0"
description = "Check system status and generate a health report"

[model]
suggested = ["anthropic/claude-3.5-sonnet", "openai/gpt-4o"]

[defaults]
temperature = 0.3
```

**system.md**:
```markdown
# System Status Checker

You are a system status monitoring agent. When invoked, you:

1. Check the current date and time
2. Generate a simulated status report covering:
   - System health (CPU, memory, disk)
   - Recent activity summary
   - Recommendations for maintenance
3. Output structured JSON with your findings

You run autonomously without user input - your task is fixed and predetermined.
```

**Usage**:
```bash
status-check
# Output: {"status": "healthy", "checks": [...], "recommendations": [...], "timestamp": "..."}
```

---

### Project 3: `summarize` (Structured Output)

**Purpose**: Demonstrates structured output schema with type-safe results.

**Features Demonstrated**:
- Input schema with positional argument (input.jsonschema)
- Output schema (output.jsonschema)
- Structured JSON response
- Multiple output fields

**Files**:
```
examples/summarize/
├── config.toml
├── system.md
├── input.jsonschema
├── output.jsonschema
└── .gitignore
```

**input.jsonschema**:
```json
{
  "type": "object",
  "properties": {
    "text": {
      "type": "string",
      "description": "Text content to summarize",
      "x-cli-position": 1
    }
  },
  "required": ["text"]
}
```

**output.jsonschema**:
```json
{
  "type": "object",
  "properties": {
    "title": {
      "type": "string",
      "description": "A concise title for the content"
    },
    "summary": {
      "type": "string",
      "description": "A 2-3 sentence summary"
    },
    "key_points": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Key points extracted from the content"
    },
    "word_count": {
      "type": "integer",
      "description": "Approximate word count of original"
    }
  },
  "required": ["title", "summary", "key_points"]
}
```

**Usage**:
```bash
summarize < article.txt
# Output: {"title": "...", "summary": "...", "key_points": [...], "word_count": 500}
```

---

### Project 3: `translate` (CLI Flags)

**Purpose**: Demonstrates input schema with CLI flag customization.

**Features Demonstrated**:
- Input schema (input.jsonschema)
- CLI flag generation
- Custom flag names (x-cli-flag)
- Short flags (x-cli-short)
- Required fields
- Default values

**Files**:
```
examples/translate/
├── config.toml
├── system.md
├── input.jsonschema
└── .gitignore
```

**input.jsonschema**:
```json
{
  "type": "object",
  "properties": {
    "text": {
      "type": "string",
      "description": "Text to translate",
      "x-cli-position": 1
    },
    "from": {
      "type": "string",
      "description": "Source language",
      "x-cli-flag": "source-language",
      "x-cli-short": "-s",
      "default": "auto"
    },
    "to": {
      "type": "string",
      "description": "Target language",
      "x-cli-short": "-t"
    },
    "formal": {
      "type": "boolean",
      "description": "Use formal tone",
      "default": false
    }
  },
  "required": ["to"]
}
```

**Usage**:
```bash
translate "Hello" --to spanish
translate "Bonjour" -s french -t english
translate "Hi" --to german --formal
```

---

### Project 4: `code-review` (File Input + Complex Types)

**Purpose**: Demonstrates file handling and complex input/output schemas.

**Features Demonstrated**:
- File input (x-cli-file)
- Multiple input types (string, boolean, integer, enum)
- Nested object schemas
- Array outputs
- Enum constraints

**Files**:
```
examples/code-review/
├── config.toml
├── system.md
├── input.jsonschema
├── output.jsonschema
└── .gitignore
```

**input.jsonschema**:
```json
{
  "type": "object",
  "properties": {
    "file": {
      "type": "string",
      "description": "Path to code file",
      "x-cli-position": 1,
      "x-cli-file": true
    },
    "language": {
      "type": "string",
      "description": "Programming language",
      "enum": ["go", "python", "javascript", "typescript", "rust"]
    },
    "severity": {
      "type": "string",
      "description": "Minimum severity level to report",
      "enum": ["info", "warning", "error", "critical"],
      "default": "warning"
    },
    "max_issues": {
      "type": "integer",
      "description": "Maximum number of issues to report",
      "default": 10
    },
    "include_suggestions": {
      "type": "boolean",
      "description": "Include fix suggestions",
      "default": true
    }
  },
  "required": ["file"]
}
```

**output.jsonschema**:
```json
{
  "type": "object",
  "properties": {
    "issues": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "line": { "type": "integer" },
          "severity": { "type": "string" },
          "message": { "type": "string" },
          "suggestion": { "type": "string" }
        }
      }
    },
    "summary": {
      "type": "string"
    },
    "score": {
      "type": "integer",
      "description": "Code quality score 0-100"
    }
  }
}
```

**Usage**:
```bash
code-review main.go
code-review app.py --language python --severity error
code-review Component.tsx -o review.json
```

---

### Project 5: `research` (Prompt Templates)

**Purpose**: Demonstrates prompt template functions and dynamic prompts.

**Features Demonstrated**:
- Prompt template (prompt.tmpl)
- Template functions: json, file, env, upper, lower, title, trim
- Conditional logic in templates
- Data interpolation

**Files**:
```
examples/research/
├── config.toml
├── system.md
├── prompt.tmpl
├── input.jsonschema
├── output.jsonschema
└── .gitignore
```

**prompt.tmpl**:
```gotemplate
Research topic: {{.topic}}

{{if .context}}Additional context from file:
{{file .context}}
{{end}}

{{if .env "RESEARCH_DEPTH"}}Depth level: {{.env "RESEARCH_DEPTH"}}
{{end}}

Focus areas:
{{range .focus_areas}}
- {{. | title}}
{{end}}

Output format: JSON with structured findings.
```

**Usage**:
```bash
RESEARCH_DEPTH=deep research "quantum computing" --focus-areas applications,theory
```

---

### Project 6: `task-runner` (Skills)

**Purpose**: Demonstrates skill integration for complex multi-step workflows.

**Features Demonstrated**:
- Skills directory structure
- Multiple skills with descriptions
- Skill invocation in system prompt

**Files**:
```
examples/task-runner/
├── config.toml
├── system.md
├── input.jsonschema
├── output.jsonschema
├── skills/
│   ├── plan/
│   │   └── SKILL.md
│   ├── execute/
│   │   └── SKILL.md
│   └── review/
│       └── SKILL.md
└── .gitignore
```

**skills/plan/SKILL.md**:
```markdown
# Task Planning

Break down complex tasks into actionable steps.

## When to use
- Multi-step workflows
- Complex objectives
- Tasks requiring coordination

## Approach
1. Analyze the objective
2. Identify dependencies
3. Create ordered steps
4. Define success criteria
```

**Usage**:
```bash
task-runner "Deploy the application to production"
```

---

### Project 7: `notifier` (Hooks)

**Purpose**: Demonstrates lifecycle hooks for external integrations.

**Features Demonstrated**:
- Hook scripts
- Agent lifecycle events
- Hook payload structure
- Error handling hooks

**Files**:
```
examples/notifier/
├── config.toml
├── system.md
├── input.jsonschema
├── hooks/
│   ├── agent-start
│   ├── agent-finish
│   ├── agent-error
│   └── tool-call
└── .gitignore
```

**hooks/agent-start**:
```bash
#!/bin/bash
# Send notification when agent starts
PAYLOAD=$(cat)
EVENT=$(echo "$PAYLOAD" | jq -r '.event')
curl -X POST "https://hooks.example.com/start" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD"
```

**hooks/agent-finish**:
```bash
#!/bin/bash
# Log completion
PAYLOAD=$(cat)
TIMESTAMP=$(echo "$PAYLOAD" | jq -r '.timestamp')
echo "[$TIMESTAMP] Agent completed successfully" >> /var/log/agent.log
```

---

### Project 8: `data-pipeline` (Full-Featured)

**Purpose**: Comprehensive example combining all features.

**Features Demonstrated**:
- All input types
- Complex nested schemas
- Skills + Hooks
- Prompt templates
- Model requirements
- Default parameters

**Files**:
```
examples/data-pipeline/
├── config.toml
├── system.md
├── prompt.tmpl
├── input.jsonschema
├── output.jsonschema
├── skills/
│   ├── extract/
│   │   └── SKILL.md
│   ├── transform/
│   │   └── SKILL.md
│   └── validate/
│       └── SKILL.md
├── hooks/
│   ├── agent-start
│   ├── step-start
│   ├── step-finish
│   └── agent-finish
└── .gitignore
```

**config.toml**:
```toml
[agent]
name = "data-pipeline"
version = "1.0.0"
description = "ETL pipeline orchestrator with AI-powered transformations"

[model]
requires_structured_output = true
requires_tools = false
requires_vision = false
suggested = ["anthropic/claude-3.5-sonnet", "openai/gpt-4o"]
default = "anthropic/claude-3.5-sonnet"

[defaults]
temperature = 0.3
max_tokens = 4096
```

---

## Documentation Suite

### Structure

```
docs/
├── README.md                 # Landing page
├── getting-started/
│   ├── installation.md
│   ├── quickstart.md         # Uses echo example
│   └── first-agent.md
├── reference/
│   ├── project-structure.md
│   ├── config.md
│   ├── input-schema.md       # Uses translate, code-review examples
│   ├── output-schema.md      # Uses summarize example
│   ├── cli-flags.md          # Uses translate example
│   ├── prompt-templates.md   # Uses research example
│   ├── skills.md             # Uses task-runner example
│   ├── hooks.md              # Uses notifier example
│   └── generated-code.md
├── guides/
│   ├── building-agents.md
│   ├── structured-outputs.md
│   ├── file-processing.md
│   ├── integrations.md
│   └── best-practices.md
├── examples/
│   ├── README.md             # Examples gallery
│   ├── echo.md
│   ├── summarize.md
│   ├── translate.md
│   ├── code-review.md
│   ├── research.md
│   ├── task-runner.md
│   ├── notifier.md
│   └── data-pipeline.md
└── api/
    └── generated-types.md    # Auto-documented types
```

### Document Specifications

#### 1. `getting-started/quickstart.md`

**Goal**: Get users running in 5 minutes

**Structure**:
1. Prerequisites (Go 1.21+, API key)
2. Install ayo
3. Create first project (echo)
4. Build and run
5. See it work

**Example Usage**:
```bash
# Install
go install github.com/charmbracelet/ayo/cmd/ayo@latest

# Create
mkdir my-agent && cd my-agent
ayo new

# Configure (show minimal config.toml)
# Build
ayo build

# Run
./my-agent "Hello"
```

---

#### 2. `reference/input-schema.md`

**Goal**: Complete reference for input.jsonschema

**Structure**:
1. Overview
2. Basic types (string, integer, number, boolean)
3. Required fields
4. Default values
5. Enum constraints
6. CLI extensions
   - x-cli-position (positional args)
   - x-cli-flag (custom flag name)
   - x-cli-short (short flag)
   - x-cli-file (file content loading)
7. Complex types (arrays, nested objects)
8. Examples from translate and code-review

**Example Snippet**:
```markdown
### CLI Position

Positional arguments are specified with `x-cli-position`:

\`\`\`json
{
  "properties": {
    "file": {
      "type": "string",
      "x-cli-position": 1
    }
  }
}
\`\`\`

Generates:
\`\`\`go
func buildInput(args []string) Input {
    var input Input
    if len(args) > 0 {
        input.File = args[0]
    }
    return input
}
\`\`\`
```

---

#### 3. `reference/hooks.md`

**Goal**: Complete reference for hook system

**Structure**:
1. Overview
2. Available hook types
3. Hook payload format
4. Creating hook scripts
5. Hook execution order
6. Error handling
7. Example from notifier

**Hook Types Table**:
| Hook | When | Payload |
|------|------|---------|
| agent-start | Agent begins execution | `{}` |
| agent-finish | Agent completes successfully | `{result: ...}` |
| agent-error | Agent encounters error | `{error: ...}` |
| step-start | Before each step | `{step: ...}` |
| step-finish | After each step | `{step: ..., result: ...}` |
| text-start | Text streaming begins | `{}` |
| text-delta | Text chunk received | `{delta: ...}` |
| text-end | Text streaming ends | `{full_text: ...}` |
| tool-call | Tool is called | `{tool: ..., args: ...}` |
| tool-result | Tool returns result | `{tool: ..., result: ...}` |

---

#### 4. `reference/prompt-templates.md`

**Goal**: Reference for template functions

**Functions Table**:
| Function | Description | Example |
|----------|-------------|---------|
| `json` | Serialize to JSON | `{{.data \| json}}` |
| `file` | Read file contents | `{{file "/path/to/file"}}` |
| `env` | Get environment variable | `{{env "API_KEY"}}` |
| `upper` | Uppercase string | `{{.name \| upper}}` |
| `lower` | Lowercase string | `{{.name \| lower}}` |
| `title` | Title case string | `{{.title \| title}}` |
| `trim` | Trim whitespace | `{{.text \| trim}}` |

---

#### 5. `examples/README.md` (Gallery)

**Goal**: Quick navigation to all examples

**Structure**:
```markdown
# Examples Gallery

## By Complexity

| Example | Input | Output | Skills | Hooks | Templates | Autonomous |
|---------|-------|--------|--------|-------|-----------|------------|
| [echo](./echo.md) | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| [status-check](./status-check.md) | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ |
| [summarize](./summarize.md) | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| [translate](./translate.md) | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| [code-review](./code-review.md) | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| [research](./research.md) | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ |
| [task-runner](./task-runner.md) | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| [notifier](./notifier.md) | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ |
| [data-pipeline](./data-pipeline.md) | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |

## By Use Case

### Text Processing
- **echo**: Simplest text transformation
- **summarize**: Extract structured summaries from text
- **translate**: Language translation with options

### Automation & Monitoring
- **status-check**: Autonomous health checks with no input required

### Code Analysis
- **code-review**: Automated code review with file input

### Research & Data
- **research**: Research assistant with templates
- **data-pipeline**: Full ETL orchestration

### Orchestration
- **task-runner**: Multi-step task planning
- **notifier**: Event-driven notifications
```

---

## Implementation Protocol

When building examples, if any error or unexpected behavior is encountered:

### 1. STOP Immediately
- Do not continue with other examples
- Do not workaround or skip the issue

### 2. RESEARCH
- Read error messages completely
- Examine relevant source code in `internal/`
- Check generated code in build output
- Search for similar patterns in existing tests
- Document findings

### 3. PLAN
- Identify root cause
- Design minimal fix
- Consider edge cases
- Document the plan

### 4. VALIDATE
- Review plan against codebase patterns
- Check for unintended side effects
- Verify fix doesn't break existing tests

### 5. EXECUTE
- Implement the fix
- Make minimal necessary changes

### 6. TEST
- Run affected package tests
- Run full test suite
- Verify the original error is resolved

### 7. VERIFY
- Re-run the failing example build
- Confirm it now works as expected
- Document what was fixed

### 8. RESUME
- Return to the example that caused the error
- Continue from that point

This loop continues until all 9 examples build and run successfully. Only then proceed to documentation.

---

## Acceptance Criteria

### Examples

- [ ] All 9 examples build successfully with `ayo build`
- [ ] Each example runs without errors
- [ ] Examples demonstrate unique feature combinations
- [ ] Examples progress from simple to complex
- [ ] Each example has a clear README.md

### Documentation

- [ ] All reference docs cover their topic completely
- [ ] Code examples are copy-pasteable and working
- [ ] Examples are cross-referenced in relevant docs
- [ ] Getting started guide works end-to-end
- [ ] API reference is auto-generated from code

## Testing Strategy

1. **Build Tests**: Each example must compile
2. **Schema Validation**: All JSON schemas are valid
3. **Hook Tests**: Hook scripts are executable
4. **Template Tests**: Templates parse without errors
5. **Documentation Links**: All links are valid
6. **Code Examples**: All code blocks are tested

## Implementation Order

1. Create `examples/` directory structure
2. Implement examples 1-4 (core features)
3. Write getting-started docs
4. Implement examples 5-8 (advanced features)
5. Write reference docs
6. Write example gallery
7. Write guides
8. Add integration tests
