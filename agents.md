# Agent Instructions

**CRITICAL:** After every code change, you MUST add/maintain tests and keep them passing. NEVER reply to the user while tests are failing. Never skip tests.
**CRITICAL:** Do not modify anything under ./.read-only (vendored, read-only). Explore only.
**CRITICAL:** Do not modify anything under `./.local/share/ayo/` (installed built-ins). This directory contains files copied from source by `./install.sh`. To modify built-in agents, skills, or prompts, edit the source files in `internal/builtin/` and run `./install.sh` to reinstall.
**CRITICAL:** Always use `./install.sh` to build the application. This script automatically installs to `.local/bin/` unless on a clean `main` branch in sync with origin. If you cannot use the script, you MUST set `GOBIN=$(pwd)/.local/bin` manually. NEVER install to the standard GOBIN location unless on an unmodified `main` branch that is in sync with `origin/main`.
**CRITICAL:** Never use emojis or unicode glyphs that have inherent colors. This ensures the UI respects user terminal theme preferences. The following is a non-exhaustive list:
	- **Geometric shapes:** `◆ ◇ ● ○ ◐ ◑ ◒ ◓ ◉ ◎ ■ □ ▪ ▫ ▲ △ ▼ ▽ ▶ ▷ ◀ ◁ ▸ ▹`
	- **Box drawing:** `─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼ ═ ║ ╭ ╮ ╯ ╰`
	- **Arrows:** `→ ← ↑ ↓ ↔ ⇒ ⇐ ➜ ➤`
	- **Dingbats/symbols:** `✓ ✗ ❯ ❮ • ‣ ★ ☆ ⋯ ≡`
	- **Braille (spinners):** `⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏`
	- **Block elements:** `█ ▓ ▒ ░ ▀ ▄ ▌ ▐`
- **CRITICAL:** All command examples in documentation (README.md, AGENTS.md, etc.) must work if copy/pasted.

## Documentation Guidelines

- Use real agent handles and skill names that exist (e.g., `@ayo`, `@ayo.research`, `debugging`)
- For commands that create new entities (like `ayo agents create @myagent`), placeholders are acceptable since they will create the entity
- Directory structure diagrams showing hypothetical user content are acceptable (e.g., `@myagent/` to show where user agents go)
- Never use placeholder names like `@agent`, `@myagent`, `@source-agent` in commands that query or operate on existing entities
- Always test example commands before committing documentation changes

## CLI Skill Maintenance

**CRITICAL:** The `ayo` skill at `internal/builtin/skills/ayo/SKILL.md` documents the CLI for use by agents. **This skill MUST be updated whenever the CLI changes.**

When modifying the CLI:
1. Add/remove/modify commands or flags
2. Update `internal/builtin/skills/ayo/SKILL.md` to reflect the changes
3. Ensure all command examples in the skill are accurate and work when copy/pasted
4. Keep the flag tables in sync with actual `--help` output

The skill should document:
- All commands and subcommands with their flags
- Common workflows and examples
- Configuration file format
- Directory structure conventions

## Preferred Libraries (./.read-only)

The `./.read-only` directory contains vendored source code from Charm and related libraries. These are **read-only reference implementations** for illustrative purposes only.

**CRITICAL: These libraries are the REQUIRED stack for this application. Before implementing ANY feature:**
1. **Consult `./.read-only` first** - explore the source to understand patterns and APIs
2. **Use these libraries** as the first-line solution for any applicable problem
3. **Follow the patterns** demonstrated in the reference implementations (crush, glow, soft-serve)
4. **Never reinvent** functionality that exists in these libraries

**IMPORTANT:**
- These sources are snapshots and may be outdated - always verify against live documentation when implementing
- Do NOT modify files in `./.read-only` - explore only
- When in doubt, check the official docs at https://charm.sh/

### Library Reference

| Library | Import | Use For |
|---------|--------|---------|
| **Bubble Tea** | `github.com/charmbracelet/bubbletea` | Interactive TUI apps (Elm Architecture) |
| **Bubbles** | `github.com/charmbracelet/bubbles` | Pre-built TUI components (spinners, inputs, tables) |
| **Lip Gloss** | `github.com/charmbracelet/lipgloss` | Terminal styling (colors, borders, layout) |
| **Glamour** | `github.com/charmbracelet/glamour` | Markdown rendering in terminal |
| **Huh** | `github.com/charmbracelet/huh` | Interactive forms and prompts |
| **Log** | `github.com/charmbracelet/log` | Styled, leveled logging |
| **Harmonica** | `github.com/charmbracelet/harmonica` | Spring physics animations |
| **Fang** | `github.com/charmbracelet/fang` | Cobra CLI enhancements (help, manpages) |
| **Catwalk** | `github.com/charmbracelet/catwalk` | AI provider/model configuration |
| **Fantasy** | `charm.land/fantasy` | Provider-agnostic LLM abstraction (streaming, tools, agents) |

### When to Use Each Library

#### Bubble Tea (`./.read-only/bubbletea`)
**Use for:** Full interactive TUI applications with state management
- Complex multi-screen interfaces
- Real-time updates and event handling
- Keyboard/mouse input handling
- Any app needing Model-View-Update pattern

**Key patterns:**
```go
// Implement tea.Model interface
type model struct { /* state */ }
func (m model) Init() tea.Cmd { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* handle msgs */ }
func (m model) View() string { return /* render with lipgloss */ }

// Run the program
p := tea.NewProgram(model{})
if _, err := p.Run(); err != nil { /* handle */ }
```

#### Bubbles (`./.read-only/bubbles`)
**Use for:** Low-level TUI components when huh doesn't fit or for custom UIs

**IMPORTANT: For forms/wizards, prefer huh. Use bubbles for:**
- Custom scrollable content (viewport)
- Loading indicators in non-form contexts (spinner)
- Complex data display (table, list)
- Specialized inputs not covered by huh

**Component Reference:**

| Component | File | When to Use |
|-----------|------|-------------|
| `viewport` | `viewport/viewport.go` | Scrollable content (logs, docs, previews) |
| `spinner` | `spinner/spinner.go` | Loading indicators in custom UIs |
| `textinput` | `textinput/textinput.go` | Single-line input (prefer huh.NewInput for forms) |
| `textarea` | `textarea/textarea.go` | Multi-line editor (prefer huh.NewText for forms) |
| `table` | `table/table.go` | Tabular data with selection |
| `list` | `list/list.go` | Rich filterable lists (prefer huh.NewSelect for forms) |
| `filepicker` | `filepicker/filepicker.go` | File browser (prefer huh.NewFilePicker for forms) |
| `paginator` | `paginator/paginator.go` | Pagination state ("1/5" or dots) |
| `progress` | `progress/progress.go` | Animated progress bars |
| `timer` | `timer/timer.go` | Countdown timers |
| `stopwatch` | `stopwatch/stopwatch.go` | Elapsed time |
| `help` | `help/help.go` | Keybinding help display |
| `key` | `key/key.go` | Keybinding definitions |

**Viewport (scrollable content):**
```go
vp := viewport.New(width, height)
vp.SetContent(longText)  // Auto-splits by newlines

// In Update:
vp, cmd = vp.Update(msg)

// Key methods:
vp.ScrollDown(n), vp.ScrollUp(n)
vp.PageDown(), vp.PageUp()
vp.GotoTop(), vp.GotoBottom()
vp.AtTop(), vp.AtBottom()
vp.ScrollPercent()  // 0.0-1.0
```

**Spinner pattern:**
```go
s := spinner.New(spinner.WithSpinner(spinner.Dot))

// MUST return Tick in Init
func (m model) Init() tea.Cmd { return m.spinner.Tick }

// Handle TickMsg in Update
case spinner.TickMsg:
    m.spinner, cmd = m.spinner.Update(msg)
```

**Table:**
```go
t := table.New(
    table.WithColumns([]table.Column{
        {Title: "Name", Width: 20},
        {Title: "Value", Width: 10},
    }),
    table.WithRows([]table.Row{
        {"Item A", "100"},
        {"Item B", "200"},
    }),
    table.WithFocused(true),
)
selected := t.SelectedRow()
```

**Key patterns:**
```go
// Embed bubble as field, delegate Update/View
type model struct {
    spinner spinner.Model
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.spinner, cmd = m.spinner.Update(msg)
    return m, cmd
}
```

#### Lip Gloss (`./.read-only/lipgloss`)
**Use for:** Styling terminal output
- Colors (foreground, background, adaptive)
- Text formatting (bold, italic, underline)
- Borders and padding
- Layout (centering, joining text blocks)
- `lipgloss/table` - Styled tables
- `lipgloss/list` - Styled lists
- `lipgloss/tree` - Tree rendering

**Key patterns:**
```go
// Styles are immutable, chain methods
style := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#FF0")).
    Padding(1, 2)
output := style.Render("text")

// Layout
lipgloss.JoinHorizontal(lipgloss.Top, left, right)
lipgloss.JoinVertical(lipgloss.Left, top, bottom)
lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
```

#### Glamour (`./.read-only/glamour`)
**Use for:** Rendering Markdown in terminal
- Displaying README/documentation
- Help text with formatting
- Any Markdown content (API responses, notes)
- Multiple built-in themes (dark, light, dracula, tokyo-night)

**Key patterns:**
```go
// Quick render with style
out, _ := glamour.Render(markdown, "dark")

// Custom renderer
r, _ := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(80),
)
out, _ := r.Render(markdown)
```

#### Huh (`./.read-only/huh`)
**Use for:** Interactive forms and user input collection - THE PREFERRED LIBRARY FOR ALL FORM UIs

**CRITICAL: Use huh for ALL wizard/form UIs instead of building custom Bubble Tea models.**

**When to use huh vs custom Bubble Tea:**
- **Use huh**: Forms, wizards, input collection, selection menus, confirmations, file pickers
- **Use custom Bubble Tea**: Real-time dashboards, games, complex non-linear UIs, streaming content

**Key source files:**
- `./.read-only/huh/form.go` - Core Form type
- `./.read-only/huh/group.go` - Group for multi-step wizards  
- `./.read-only/huh/field_*.go` - All field types
- `./.read-only/huh/theme.go` - Theming system

**All Available Field Types:**

| Field | Use For | Key Options |
|-------|---------|-------------|
| `NewInput()` | Single-line text | `Placeholder`, `Prompt`, `CharLimit`, `EchoMode`, `Suggestions`, `Validate` |
| `NewText()` | Multi-line text | `Lines`, `CharLimit`, `ShowLineNumbers`, `Editor()`, `EditorExtension` |
| `NewSelect[T]()` | Single selection | `Options`, `Height`, `Inline`, `Filtering` |
| `NewMultiSelect[T]()` | Multiple selection | `Options`, `Height`, `Limit`, `Filterable` |
| `NewConfirm()` | Yes/No prompt | `Affirmative`, `Negative`, `Inline` |
| `NewNote()` | Display-only text | `Height`, `Next`, `NextLabel` |
| `NewFilePicker()` | File selection | `CurrentDirectory`, `AllowedTypes`, `ShowHidden`, `DirAllowed` |

**Multi-step wizards with groups:**
```go
form := huh.NewForm(
    // Step 1
    huh.NewGroup(
        huh.NewInput().Title("Name").Value(&name),
        huh.NewInput().Title("Email").Value(&email),
    ).Title("Identity"),
    
    // Step 2  
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("Plan").
            Options(huh.NewOptions("Free", "Pro", "Enterprise")...).
            Value(&plan),
    ).Title("Subscription"),
    
    // Step 3 - Conditional (skip if Free plan)
    huh.NewGroup(
        huh.NewInput().Title("Card Number").Value(&card),
    ).Title("Payment").
      WithHideFunc(func() bool { return plan == "Free" }),
).WithTheme(huh.ThemeCharm())

if err := form.Run(); err != nil { /* handle */ }
```

**Dynamic options based on previous selections:**
```go
var country string
huh.NewSelect[string]().
    Value(&state).
    TitleFunc(func() string {
        if country == "Canada" { return "Province" }
        return "State"
    }, &country).
    OptionsFunc(func() []huh.Option[string] {
        return getStatesFor(country)
    }, &country)
```

**Built-in themes:** `ThemeCharm()`, `ThemeDracula()`, `ThemeCatppuccin()`, `ThemeBase16()`, `ThemeBase()`

**Key examples to study:**
- `./.read-only/huh/examples/burger/main.go` - Full multi-step wizard
- `./.read-only/huh/examples/bubbletea/main.go` - Embedding in Bubble Tea
- `./.read-only/huh/examples/dynamic/` - Dynamic forms
- `./.read-only/huh/examples/conditional/main.go` - Conditional fields
- `./.read-only/huh/examples/filepicker/main.go` - File selection

**Spinner for async operations:**
```go
import "github.com/charmbracelet/huh/spinner"
spinner.New().
    Title("Processing...").
    Action(func() { doWork() }).
    Run()
```

#### Log (`./.read-only/log`)
**Use for:** Application logging
- Leveled logging (debug, info, warn, error, fatal)
- Colored, styled output
- Structured logging with key-value pairs
- JSON/logfmt formatters
- slog compatibility

**Key patterns:**
```go
log.Info("message", "key", value)
log.Error("failed", "err", err)

// Custom logger
logger := log.NewWithOptions(os.Stderr, log.Options{
    Level: log.DebugLevel,
    ReportTimestamp: true,
})
```

#### Fantasy (`./.read-only/fantasy`)
**Use for:** Provider-agnostic LLM abstraction
- Unified API across Anthropic, OpenAI, Google, OpenRouter
- Streaming responses with callbacks
- Tool/function calling
- Agent orchestration with stop conditions

**Key patterns:**
```go
// Create provider and model
provider, _ := openrouter.New(openrouter.WithAPIKey(key))
model, _ := provider.LanguageModel(ctx, "anthropic/claude-3.5-sonnet")

// Create agent with tools
agent := fantasy.NewAgent(model,
    fantasy.WithSystemPrompt("You are helpful."),
    fantasy.WithTools(myTools...),
    fantasy.OnTextDelta(func(delta string) { fmt.Print(delta) }),
)

// Generate with stop condition
result, _ := agent.Generate(ctx, fantasy.AgentCall{
    Prompt: "Hello",
    StopWhen: fantasy.FinishReasonIs(fantasy.FinishReasonEndTurn),
})
```

#### Fang (`./.read-only/fang`)
**Use for:** Enhancing Cobra CLI apps
- Styled help output
- Automatic `--version` flag from build info
- Manpage generation
- Consistent error handling

**Key patterns:**
```go
// Replace cmd.Execute() with fang.Execute()
if err := fang.Execute(ctx, rootCmd); err != nil {
    os.Exit(1)
}
```

### Reference Implementations

#### Crush (`./.read-only/crush`)
**THE primary reference** for AI CLI implementation patterns.

**Key directories to study:**
- `./.read-only/crush/internal/tui/` - Main TUI architecture
- `./.read-only/crush/internal/tui/components/dialogs/` - Dialog system
- `./.read-only/crush/internal/tui/components/chat/splash/` - Onboarding wizard
- `./.read-only/crush/internal/tui/layout/` - Component interfaces

**Patterns to learn from:**

1. **Page-based navigation with lazy loading:**
```go
type appModel struct {
    currentPage  page.PageID
    pages        map[page.PageID]util.Model
    loadedPages  map[page.PageID]bool
}
```

2. **Dialog system with message-based open/close:**
```go
type OpenDialogMsg struct { Model DialogModel }
type CloseDialogMsg struct{}
```

3. **State-based multi-step flows (splash.go):**
```go
type splashCmp struct {
    isOnboarding     bool
    needsProjectInit bool
    needsAPIKey      bool
}
```

4. **Interface-based components:**
```go
type Focusable interface {
    Focus() tea.Cmd
    Blur() tea.Cmd
    IsFocused() bool
}
```

5. **Message-driven transitions:**
```go
util.CmdHandler(ModelSelectedMsg{Model: selected})
```

**NOTE:** Crush builds custom form components instead of using huh. For ayo, prefer huh for forms.

#### Soft Serve (`./.read-only/soft-serve`)
**Reference for:** Complex multi-component TUI
- Wish SSH server integration
- Bubble Tea over SSH
- Git operations
- Database integration

#### Glow (`./.read-only/glow`)
**Reference for:** Markdown TUI browser
- Glamour rendering
- File browser patterns
- GitHub/GitLab fetching

#### Gum (`./.read-only/gum`)
**Reference for:** Exposing Bubbles as CLI commands
- Flag-based configuration
- Shell script integration

### Additional Tools

#### Sequin (`./.read-only/sequin`)
**Use for:** Debugging terminal output
- Decoding ANSI escape sequences
- Inspecting TUI rendering
- Validating golden test files

#### Ultraviolet (`./.read-only/ultraviolet`)
**Use for:** Low-level terminal rendering (advanced)
- Cell-based diffing renderer
- Cross-platform terminal I/O
- Internal use by Bubble Tea v2

#### x/ Packages (`./.read-only/x`)
**Experimental utilities:**
- `x/ansi` - ANSI escape sequence parsing
- `x/term` - Terminal utilities (size, raw mode)
- `x/editor` - Open files in text editors
- `x/exp/golden` - Golden file testing

## Completion Checklist

**Before reporting any task as complete, you MUST:**

1. **Run the full test suite**: `go test ./...`
   - If any test fails, fix it immediately without asking the user
   - Keep iterating until all tests pass
   - Never report completion while tests are failing

2. **Rebuild the binary**: `go install ./cmd/ayo`
   - This ensures the `ayo` command reflects all changes

A task is NOT complete until both steps pass successfully.

A Go-based command line tool for managing local AI agents.

## Features

- Define, manage, and run AI agents
- Built-in agents shipped with the binary
- Interactive chat sessions within the terminal
- Non-interactive single-prompt mode
- Bash tool as default for task execution
- System prompts assembled from prefix, shared, agent, tools, skills, and suffix
- Configurable models via Catwalk

## Configuration

Ayo uses two directories:

**Unix (macOS, Linux):**
- User config: `~/.config/ayo/` (ayo.json, ayo-schema.json, user agents, user skills, prompts)
- Built-in data: `~/.local/share/ayo/` (agents and skills auto-installed on first run)

**Dev mode:** When running from a source checkout (`go run ./cmd/ayo`), built-in data is stored in `{repo}/.ayo/` instead. User config remains at `~/.config/ayo/`.

**Windows:**
- Both: `%LOCALAPPDATA%\ayo\`

```json
// ~/.config/ayo/ayo.json
{
  "$schema": "./ayo-schema.json",
  "agents_dir": "~/.config/ayo/agents",
  "skills_dir": "~/.config/ayo/skills",
  "system_prefix": "~/.config/ayo/prompts/prefix.md",
  "system_suffix": "~/.config/ayo/prompts/suffix.md",
  "default_model": "gpt-4.1",
  "provider": {}
}
```

## Directory Structure

**Production (installed binary):**
```
~/.config/ayo/                    # User configuration (editable)
├── ayo.json                      # Main config file
├── ayo-schema.json               # JSON schema for config (auto-installed)
├── agents/                       # User-defined agents
│   └── @myagent/
│       ├── config.json
│       ├── system.md
│       └── skills/               # Agent-specific skills
├── skills/                       # User-defined shared skills
│   └── my-skill/
│       └── SKILL.md
└── prompts/                      # Custom system prompts
    ├── prefix.md
    ├── system.md
    └── suffix.md

~/.local/share/ayo/               # Built-in data (auto-installed on first run)
├── agents/                       # Built-in agents
│   └── @ayo/
│       ├── config.json
│       ├── system.md
│       └── skills/
├── skills/                       # Built-in shared skills
│   └── debugging/
│       └── SKILL.md
└── .builtin-version              # Version marker
```

**Dev mode (running from source checkout):**
```
~/Code/ayo-skills/                # Your checkout
├── .ayo/                         # Built-in data (project-local)
│   ├── agents/
│   ├── skills/
│   └── .builtin-version
└── ...

~/.config/ayo/                    # User config (shared across all instances)
├── agents/
├── skills/
└── ...
```

This allows multiple dev branches to have isolated built-ins while sharing user-defined agents and skills.

## Loading Priority

**Agents:** User agents (`~/.config/ayo/agents`) take priority over built-in agents (`~/.local/share/ayo/agents`).

**Skills:** Discovery priority (first found wins):
1. Agent-specific skills (in agent's `skills/` directory)
2. User shared skills (`~/.config/ayo/skills`)
3. Built-in skills (`~/.local/share/ayo/skills`)

## Usage

```bash
# Setup (optional - built-ins auto-install on first run)
ayo setup                   # Reinstall built-ins, create user dirs
ayo setup --force           # Overwrite modifications without prompting

# Chat
ayo                         # Start interactive chat with default @ayo agent
ayo "tell me a joke"        # Run single prompt with default @ayo agent
ayo @ayo                   # Start interactive chat session with agent
ayo @ayo "tell me a joke"  # Run single prompt (non-interactive)

# Agents management
ayo agents list             # List available agents
ayo agents show @ayo      # Show agent details
ayo agents create <handle>  # Create new agent
ayo agents dir              # Show agents directories
ayo agents update           # Update built-in agents
ayo agents update --force   # Update without checking for modifications

# Skills management
ayo skills list             # List available skills
ayo skills show <name>      # Show skill details
ayo skills create <name>    # Create new skill
ayo skills validate <path>  # Validate skill directory
ayo skills dir              # Show skills directories
ayo skills update           # Update built-in skills

# Sessions management
ayo sessions list           # List conversation sessions
ayo sessions list -a @ayo   # Filter by agent
ayo sessions show <id>      # Show session details and messages
ayo sessions continue       # Continue a session (interactive picker)
ayo sessions continue <id>  # Continue a specific session
ayo sessions delete <id>    # Delete a session
```

### Default Agent

When no agent is specified, ayo uses the `@ayo` agent (the default built-in agent).

### Interactive Mode

`ayo @agentname` starts an interactive chat session. The conversation continues until you exit with Ctrl+C.

- First Ctrl+C interrupts the current request
- Second Ctrl+C (at prompt) exits the session

### Non-Interactive Mode

`ayo @agentname "Your prompt here"` executes the prompt and exits.

### Sessions

Sessions persist conversation history to a local SQLite database. This enables:
- Resuming previous conversations
- Reviewing past interactions
- Managing conversation history

**Storage location:** `~/.local/share/ayo/ayo.db`

**Session lifecycle:**
1. Session created when chat starts (interactive or single prompt)
2. Messages persisted as the conversation progresses
3. Session ID displayed after each interaction
4. Sessions can be resumed with `ayo sessions continue`

**Common workflows:**

```bash
# List recent sessions
ayo sessions list

# Continue the most recent session (shows picker)
ayo sessions continue

# Continue a specific session by ID prefix
ayo sessions continue abc123

# Search and continue by title
ayo sessions continue "debugging issue"

# View session details
ayo sessions show abc123

# Clean up old sessions
ayo sessions delete abc123
```

## UI Behavior

Both interactive and non-interactive modes share the same UI components:

### Spinner Feedback

Spinners display progress during async operations:

- **LLM calls**: "Thinking..." while waiting, then "✓ Response received (elapsed)"
- **Tool calls**: Shows LLM-provided description (e.g., "Installing dependencies..."), then "✓ Installing dependencies (1.2s)" or "× ... failed (elapsed)"

### Styled Output

- Markdown rendering via glamour with syntax highlighting
- Tool outputs displayed in styled boxes
- Error messages with red styling and icons
- Reasoning/thinking displayed in bordered boxes

### Chat Header (Interactive Only)

Purple styled "Chat with @agentname" header with exit hint.

## Tool System

### Bash Tool

The `bash` tool is the default and primary tool. Agents use it to accomplish any task unless a more specific skill is available.

When calling bash, the LLM must provide:
- `command`: The shell command to execute
- `description`: Human-readable description shown in the spinner (e.g., "Running test suite")

Optional parameters:
- `timeout_seconds`: Command timeout (default 30s)
- `working_dir`: Working directory scoped to project root

### Plan Tool

The `plan` tool enables agents to track multi-step tasks with status updates. Plans are stored per-session as JSON in the database.

**Required skill:** The `planning` skill is automatically attached when the plan tool is enabled.

**Hierarchical structure:**
Plans support three levels of hierarchy:

1. **Phases** (optional): High-level stages of work
   - If used, must have at least 2 phases
   - Each phase must contain at least 1 task

2. **Tasks** (required): Units of work
   - Can exist at top level or within phases
   - Each task needs `content` and `active_form`

3. **Todos** (optional): Atomic sub-items within tasks
   - Use for granular step tracking within a task

**Parameters:**
```json
{
  "tasks": [
    {
      "content": "What needs to be done (imperative form)",
      "active_form": "Present continuous form (e.g., 'Running tests')",
      "status": "pending | in_progress | completed",
      "todos": [
        {
          "content": "Atomic sub-item",
          "active_form": "Doing sub-item",
          "status": "pending | in_progress | completed"
        }
      ]
    }
  ]
}
```

Or with phases:
```json
{
  "phases": [
    {
      "name": "Phase 1: Setup",
      "status": "completed",
      "tasks": [...]
    },
    {
      "name": "Phase 2: Implementation",
      "status": "in_progress",
      "tasks": [...]
    }
  ]
}
```

**Task states:**
- `pending`: Not yet started
- `in_progress`: Currently working on (limit to ONE item at a time across all levels)
- `completed`: Finished successfully

**Rules:**
- Each task/todo requires both `content` (imperative) and `active_form` (present continuous)
- Exactly ONE item should be `in_progress` at any time
- Mark items complete IMMEDIATELY after finishing
- Remove irrelevant items from the list entirely
- Cannot have both phases and top-level tasks (mutually exclusive)

**Storage:** Plans are stored as a JSON column on the sessions table and persist across session resumption.

### Skills

Skills extend agent capabilities by providing domain-specific instructions. Skills follow the [agentskills spec](https://agentskills.org).

Skills are discovered from multiple sources (in priority order):
1. **Agent-specific**: `{agent_dir}/skills/{skill-name}/`
2. **User shared**: `~/.config/ayo/skills/{skill-name}/`
3. **Built-in**: `~/.local/share/ayo/skills/{skill-name}/`

Each skill is a directory containing a `SKILL.md` with YAML frontmatter:

```markdown
---
name: my-skill
description: What this skill does and when to use it.
metadata:
  author: your-name
  version: "1.0"
---

# Skill Instructions

Detailed instructions for the agent...
```

**Required fields:**
- `name`: 1-64 chars, lowercase, hyphens ok, must match directory name
- `description`: 1-1024 chars, describes what the skill does and when to use it

**Optional fields:**
- `compatibility`: Environment requirements (max 500 chars)
- `metadata`: Key-value pairs (author, version, etc.)
- `allowed-tools`: Pre-approved tools (experimental)

**Optional directories:**
- `scripts/`: Executable code
- `references/`: Additional documentation
- `assets/`: Templates, data files

#### Agent Config for Skills

Agents can configure which skills are available:

```json
{
  "skills": ["skill-a", "skill-b"],
  "exclude_skills": ["unwanted-skill"],
  "ignore_builtin_skills": false,
  "ignore_shared_skills": false
}
```

#### Skills CLI Commands

```bash
ayo skills list                  # List all available skills
ayo skills list --source=built-in # Filter by source
ayo skills show <name>           # Show skill details
ayo skills validate <path>       # Validate a skill directory
ayo skills create <name>         # Create new skill from template
ayo skills create <name> --shared # Create in shared skills directory
```

#### Built-in Skills

Built-in skills are embedded in the binary and installed via `ayo setup`:

**Source (in repo):**
- Shared: `internal/builtin/skills/{skill-name}/`
- Agent-specific: `internal/builtin/agents/@ayo/skills/{skill-name}/`

**Installed to:**
- `~/.local/share/ayo/skills/`
- `~/.local/share/ayo/agents/@ayo/skills/`

**Current built-in skills:**
- `debugging` - Systematic debugging techniques
- `planning` - Task decomposition into phases, tasks, and todos (required by plan tool)
- `project-summary` - Project analysis and documentation (for @ayo)

## System Prompt Assembly

Messages are built in order:
1. Combined system (prefix + shared + agent + suffix)
2. Tools prompt (bash instructions)
3. Skills prompt (available skills XML)
4. User message

## Architecture Notes

- **Fantasy provider abstraction**: Uses `charm.land/fantasy` for provider-agnostic LLM calls. Supports OpenAI, Anthropic, Google, OpenRouter, and OpenAI-compatible providers.
- **Agent-based streaming**: Fantasy's `Agent` abstraction handles tool execution and multi-step interactions via callbacks (`OnTextDelta`, `OnToolCall`, `OnToolResult`, etc.)
- UI renders ordered tool outputs with spinner feedback

## Built-in Agents

Built-in agents are embedded in the binary and installed via `ayo setup`.

### Installation

**Source (in repo):** `internal/builtin/agents/{name}/`

**Installed to:** `~/.local/share/ayo/agents/`

**User agents:** `~/.config/ayo/agents/`

User agents take precedence over built-in agents with the same name.

### Structure

Each built-in agent directory contains:
```
internal/builtin/agents/{name}/
├── config.json      # Agent configuration
├── system.md        # System prompt (sandwiched between prefix/suffix)
└── skills/          # Optional agent-specific skills
    └── {skill}/
        └── SKILL.md
```

### Adding a Built-in Agent

1. Create directory: `internal/builtin/agents/{name}/`
2. Add `config.json`:
   ```json
   {
     "description": "Agent description",
     "allowed_tools": ["bash"]
   }
   ```
3. Add `system.md` with the agent's system prompt
4. Optionally add skills in `skills/{skill}/SKILL.md`
5. Bump `Version` constant in `internal/builtin/install.go`
6. The agent is automatically embedded via `//go:embed` and installed on next `ayo setup`

### Current Built-in Agents

- `@ayo` - The default agent, a versatile command-line assistant
- `@ayo.coding` - Coding agent that uses Crush for complex source code tasks
- `@ayo.research` - Research assistant with web search capabilities
- `@ayo.agents` - Agent management agent for creating and managing agents
- `@ayo.skills` - Skill management agent for creating and managing skills

The `ayo` namespace is reserved - users cannot create agents with the `@ayo` handle or `@ayo.` prefix.

## Crush Integration

The `@ayo.coding` agent provides integration with the [Crush](https://github.com/charmbracelet/crush) coding agent. Crush excels at complex source code tasks that require multi-file modifications.

### Prerequisites

Crush must be installed and available in your PATH:
```bash
# Install Crush (see Crush documentation for latest instructions)
go install github.com/charmbracelet/crush@latest
```

The `install.sh` script will prompt to install Crush if it's not found.

### Architecture

- `@ayo` uses the `coding` skill to know when to delegate to `@ayo.coding`
- `@ayo.coding` has ONLY the `crush` tool - it cannot use bash or other tools
- The `crush` tool maps directly to `crush run --quiet` with model passthrough

### Usage

Direct invocation:
```bash
ayo @ayo.coding "Add comprehensive error handling to the database layer"
```

Via @ayo delegation (automatic when appropriate):
```bash
ayo "Refactor the authentication module to use JWT tokens"
# @ayo will delegate this to @ayo.coding if it requires code changes
```

### The Crush Tool

The `crush` tool is available to `@ayo.coding` and maps to `crush run`:

```json
{
  "prompt": "Add input validation to auth handlers",
  "model": "claude-sonnet-4",
  "small_model": "gpt-4.1-mini",
  "working_dir": "./internal/auth"
}
```

| Parameter | Description |
|-----------|-------------|
| `prompt` | The coding task (required) |
| `model` | Model to use (optional, passed from caller) |
| `small_model` | Small model for auxiliary tasks (optional) |
| `working_dir` | Working directory (optional, defaults to project root) |

### Model Passthrough

When `@ayo` delegates to `@ayo.coding`, it can pass the model to use:

```json
{
  "agent": "@ayo.coding",
  "prompt": "Add rate limiting middleware",
  "model": "claude-sonnet-4"
}
```

The `@ayo.coding` agent receives the model in a `<model_context>` system message and passes it to the `crush` tool.

## Versioning

Ayo uses semantic versioning (semver). The CLI version is defined in `internal/version/version.go`.

### Bumping the Version

When releasing a new version:

1. Update the `Version` constant in `internal/version/version.go`
2. Follow semver conventions:
   - **MAJOR** (1.0.0): Breaking changes
   - **MINOR** (0.2.0): New features, backward compatible
   - **PATCH** (0.1.1): Bug fixes, backward compatible

```go
// internal/version/version.go
const Version = "0.2.0"  // Example: bumping minor version
```

### Checking the Version

```bash
ayo --version
# Output: ayo version 0.1.0
```

## Agent Chaining

Agents can be composed via Unix pipes when they have structured input/output schemas. The output of one agent becomes the input to the next.

### Structured I/O Schemas

Agents can define optional JSON schemas:

- `input.jsonschema` - Validates input; agent only accepts JSON matching this schema
- `output.jsonschema` - Structures output; final response is cast to this format

Example agent structure:
```
@my-agent/
├── config.json
├── system.md
├── input.jsonschema    # Optional: structured input
└── output.jsonschema   # Optional: structured output
```

### Piping Agents

```bash
# Chain two agents (code reviewer -> issue reporter)
ayo @ayo.example.chain.code-reviewer '{"repo":".", "files":["main.go"]}' | ayo @ayo.example.chain.issue-reporter
```

**Pipeline behavior:**
- Stdin is piped → agent reads JSON from stdin
- Stdout is piped → UI goes to stderr, raw JSON goes to stdout
- Full UI (spinners, reasoning, tool calls) always visible on stderr

### Schema Compatibility

When piping agents:

1. **Exact match**: Output schema identical to input schema
2. **Structural match**: Output has all required fields of input (superset OK)
3. **Freeform**: Target agent has no input schema (accepts anything)

If schemas are incompatible, validation fails with a clear error.

### Chain Discovery Commands

```bash
# List all chainable agents (have input or output schema)
ayo chain ls

# Show agent's schemas
ayo chain inspect @ayo.debug.structured-io

# Find agents that can receive this agent's output
ayo chain from @ayo.example.chain.code-reviewer

# Find agents whose output this agent can receive
ayo chain to @ayo.example.chain.issue-reporter

# Validate JSON against agent's input schema
ayo chain validate @ayo.debug.structured-io '{"environment": "staging", "service": "api"}'
echo '{"environment": "staging", "service": "api"}' | ayo chain validate @ayo.debug.structured-io

# Generate example input for an agent
ayo chain example @ayo.debug.structured-io
```

### Chain Context

When agents are chained, context is passed via environment variable:
- `AYO_CHAIN_CONTEXT` contains JSON with `depth`, `source`, and `source_description`
- Freeform agents receive a preamble describing the chain context

### Example Chain Agents

Built-in example agents demonstrating chaining:

```bash
# Code reviewer outputs structured findings
ayo @ayo.example.chain.code-reviewer '{"repo":".", "files":["main.go"]}'

# Issue reporter consumes code reviewer output
ayo @ayo.example.chain.code-reviewer '{"repo":".", "files":["main.go"]}' \
  | ayo @ayo.example.chain.issue-reporter
```
