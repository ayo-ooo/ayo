# Requirements Clarification

## Q1: config.toml Structure

### Question: What fields in config.toml?

**Answer:**
```toml
[agent]
name = "my-agent"
version = "1.0.0"
description = "An agent that does X"

[model]
# Required features (filters available models)
requires_structured_output = false
requires_tools = true
requires_vision = false

# Suggested models (in preference order)
suggested = ["claude-sonnet-4-6", "gpt-4o", "gemini-2.5-pro"]
default = "claude-sonnet-4-6"

[defaults]
temperature = 0.7
max_tokens = 4096
```

Removed `min_context_window` - not needed.

---

## Q2: Hooks Architecture

### Question: How should hooks work?

**Answer: Two-layer hooks, like `.git/hooks/`**

**Project structure:**
```
hooks/
  agent-start        # executable
  agent-finish       # executable
  agent-error        # executable
  step-start         # executable
  step-finish        # executable
  text-start         # executable
  text-delta         # executable
  text-end           # executable
  reasoning-start    # executable
  reasoning-delta    # executable
  reasoning-end      # executable
  tool-input-start   # executable
  tool-input-delta   # executable
  tool-input-end     # executable
  tool-call          # executable
  tool-result        # executable
  source             # executable
  stream-finish      # executable
  warnings           # executable
```

**Execution model:**
- Hooks are blocking observers - flow pauses until hook exits
- Hook receives event JSON on stdin
- Exit code ignored (observation only, no modification)
- If hook file doesn't exist or isn't executable, skip silently

**Two layers (both always run):**
1. Embedded hooks (from project `hooks/`) - run first, CANNOT be disabled
2. User hooks (from `~/.config/agents/<agent>.toml` paths) - run second

**User config (can only ADD, not disable):**
```toml
[hooks]
tool-call = "/home/user/my-logger.sh"    # runs AFTER embedded tool-call
agent-start = "/home/user/metrics.sh"    # runs AFTER embedded agent-start
```

---

## Q3: Prompt Template + Input Schema

### Question: How do prompt.tmpl and input.jsonschema interact?

**Answer:**

| input.jsonschema | prompt.tmpl | Behavior |
|------------------|-------------|----------|
| ✓ | ✓ | CLI parses args → template renders with structured data |
| ✓ | ✗ | JSON input (conforming to schema) sent directly as prompt |
| ✗ | ✓ | Template uses `{{ .input }}` for freeform text |
| ✗ | ✗ | Single freeform text positional arg |

---

## Q4: Output Schema

### Question: How does output.jsonschema work?

**Answer:**
- If present, agent uses Fantasy's structured output (`object.Generate[T]`)
- Output schema passed to model for constrained generation
- Validated JSON to stdout (pipeable)
- `--output path.json` flag writes to file in addition to stdout

---

## Q5: File Path Handling

### Question: How are file paths handled in schemas?

**Answer:**
- `x-cli-file: true` validates path exists
- Template receives path string: `{{ .source }}`
- Agent's skills provide file access (view, edit, glob, etc.)
- No automatic file reading - agent handles via tools

---

## Q6: Skills Integration

### Question: How are skills embedded and discovered?

**Answer:**
- Skills embedded via `embed.FS` at build time
- All files bundled (SKILL.md, scripts/, references/, assets/)
- At runtime, skill paths appended to system message
- Progressive disclosure: catalog → instructions → resources
- Agent reads via standard file tools

---

## Q7: First-Run Model Selection

### Question: First-run TUI flow?

**Answer:**
1. Check `~/.config/agents/<agent-name>.toml`
2. If not found, launch Bubbletea TUI:
   - Scan environment for API keys
   - Show providers with detected credentials
   - Filter models by agent requirements
   - User selects provider + model
   - Ollama: check install, offer to install, pull model
   - Save configuration
3. Run agent

**Non-interactive:**
- `--provider` + `--model` flags skip TUI
- `--config` flag for custom config path
