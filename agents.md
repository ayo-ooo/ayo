# Agent Instructions

**CRITICAL:** After every code change, you MUST add/maintain tests and keep them passing. NEVER reply to the user while tests are failing. Never skip tests.
**CRITICAL:** Do not modify anything under ./.read-only (vendored, read-only). Explore only.
**CRITICAL:** Do not modify anything under `./.local/share/ayo/` (installed built-ins). This directory contains files copied from source by `./install.sh`. To modify built-in agents, skills, or prompts, edit the source files in `internal/builtin/` and run `./install.sh` to reinstall.
**CRITICAL:** Always use `./install.sh` to build the application. This script automatically installs to `.local/bin/` unless on a clean `main` branch in sync with origin. If you cannot use the script, you MUST set `GOBIN=$(pwd)/.local/bin` manually. NEVER install to the standard GOBIN location unless on an unmodified `main` branch that is in sync with `origin/main`.
**CRITICAL:** If a `PLAN.txt` file exists in the project root, you MUST update it after completing each task to mark completed items and add notes for future agents. This ensures continuity across sessions.
**CRITICAL:** Never use emojis or unicode glyphs that have inherent colors. This ensures the UI respects user terminal theme preferences. The following is a non-exhaustive list:
	- **Geometric shapes:** `◆ ◇ ● ○ ◐ ◑ ◒ ◓ ◉ ◎ ■ □ ▪ ▫ ▲ △ ▼ ▽ ▶ ▷ ◀ ◁ ▸ ▹`
	- **Box drawing:** `─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼ ═ ║ ╭ ╮ ╯ ╰`
	- **Arrows:** `→ ← ↑ ↓ ↔ ⇒ ⇐ ➜ ➤`
	- **Dingbats/symbols:** `✓ ✗ ❯ ❮ • ‣ ★ ☆ ⋯ ≡`
	- **Braille (spinners):** `⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏`
	- **Block elements:** `█ ▓ ▒ ░ ▀ ▄ ▌ ▐`
- **CRITICAL:** All command examples in documentation (README.md, AGENTS.md, etc.) must work if copy/pasted.

## Documentation Guidelines

- Use real agent handles and skill names that exist (e.g., `@ayo`, `debugging`)
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

3. **Update PLAN.txt** (if it exists in project root):
   - Mark completed tasks with `[DONE]` prefix
   - Add any implementation notes or caveats discovered
   - Note any deviations from the original plan
   - This ensures future agents can continue the work seamlessly

A task is NOT complete until all steps pass successfully.

## PLAN.txt Workflow

When a `PLAN.txt` file exists in the project root, follow this workflow:

1. **At session start**: Read `PLAN.txt` and `PLAN_CONTEXT.txt` (if exists) to understand the full context
2. **Before each task**: Check if the task has dependencies that must be completed first
3. **After each task**:
   - Add `[DONE]` marker to the task line
   - Add notes in a sub-bullet if there are important implementation details
   - If a task required different approach than planned, document why
4. **After completing a group**: Add a completion summary at the end of that group's section

**Example update format:**
```markdown
A.01: Add MemoryService field to TUIStreamHandler [DONE]
- Added optional parameter to NewTUIStreamHandler
- Used functional options pattern for backward compatibility
```

## Context Window Management (CRITICAL)

**You MUST monitor your available context window after EACH task.**

### When to Trigger Handoff

If your remaining context drops below **50,000 tokens**, you MUST:

1. **Immediately stop new work** - Do not start the next task
2. **Update PLAN.txt** with:
   - All completed tasks marked `[DONE]`
   - Current task status (in-progress, blocked, etc.)
   - Any important context discovered during implementation
   - The exact next task to resume from
3. **Create a handoff summary** at the end of PLAN.txt:
   ```markdown
   ## Handoff Notes (Session N)
   
   **Stopped at**: Task X.XX
   **Reason**: Context window exhausted
   **Next action**: [Specific task to resume]
   **Critical context**:
   - [Key implementation detail 1]
   - [Key implementation detail 2]
   - [Any blockers or warnings]
   ```
4. **Use the summarization tool** with this exact prompt:
   ```
   Continue implementing the Interactive Chat TUI Redesign for ayo.
   
   IMMEDIATELY read these files to restore context:
   1. ./PLAN.txt - Contains task list and completion status
   2. ./PLAN_CONTEXT.txt - Contains full design context and rationale
   3. ./AGENTS.md - Contains project rules and patterns
   
   Resume from the task marked as next in the Handoff Notes section.
   Do not ask questions - read the files and continue working.
   ```

### Proactive Context Monitoring

- Check context after EVERY task completion
- If below 75,000 tokens, prepare for potential handoff (ensure PLAN.txt is current)
- If below 50,000 tokens, trigger handoff immediately
- NEVER let context exhaustion cause work loss - always save state first

### This Work May Take Hours

This implementation spans 58 tasks across multiple groups. You are expected to:
- Work continuously without stopping for user input
- Handle context limitations via summarization handoffs
- Complete ALL tasks including final license addition
- Only stop when the entire plan is complete OR you hit an actual blocking error

## Project Overview

A Go-based command line tool for managing local AI agents. The design philosophy is **conversational-first**: instead of complex wizards or configuration UIs, ayo relies on natural conversation with `@ayo` to accomplish tasks like creating agents and skills.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              ayo CLI                                        │
│  cmd/ayo/                                                                   │
│  ├── root.go         Entry point, argument parsing                          │
│  ├── agents.go       ayo agents subcommands                                 │
│  ├── skills.go       ayo skills subcommands                                 │
│  ├── sessions.go     ayo sessions subcommands                               │
│  ├── memory.go       ayo memory subcommands                                 │
│  ├── plugins.go      ayo plugins subcommands                                │
│  ├── chain.go        ayo chain subcommands                                  │
│  ├── setup.go        ayo setup command                                      │
│  └── doctor.go       ayo doctor command                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                           Core Packages                                     │
│  internal/                                                                  │
│  ├── agent/          Agent loading, config, validation                      │
│  ├── skills/         Skill discovery, loading, validation                   │
│  ├── run/            Agent execution, tool handling, streaming              │
│  ├── session/        Session persistence, history                           │
│  ├── memory/         Memory storage, search, embeddings                     │
│  ├── builtin/        Embedded agents/skills, installation                   │
│  ├── config/         Config loading, provider setup                         │
│  ├── delegates/      Task delegation routing                                │
│  ├── pipe/           Agent chaining, schema validation                      │
│  ├── plugins/        Plugin registry, installation                          │
│  ├── ui/             Spinners, styles, rendering                            │
│  ├── db/             SQLite database, queries                               │
│  ├── embedding/      Vector embeddings for memory search                    │
│  ├── smallmodel/     Ollama integration for local models                    │
│  └── paths/          Platform-specific path resolution                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                           External Dependencies                             │
│  Fantasy (charm.land/fantasy)     Provider-agnostic LLM abstraction         │
│  Bubble Tea, Lipgloss, Glamour    Terminal UI                               │
│  Huh                              Interactive forms                         │
│  Cobra, Fang                      CLI framework                             │
│  SQLite (sqlc)                    Persistence                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Key Concepts

| Concept | Description |
|---------|-------------|
| **Agent** | AI assistant with custom system prompt and tool access |
| **Skill** | Instruction set attached to agents at runtime |
| **Tool** | Capability (bash, agent_call, task) an agent can use |
| **Session** | Persisted conversation history |
| **Memory** | Facts/preferences learned across sessions |
| **Plugin** | External package providing agents, skills, or tools |

### Data Flow

```
User Input
    │
    ▼
┌─────────────────┐
│  Agent Loading  │  Load config.json, system.md, skills
│  (internal/agent)│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ System Prompt   │  Assemble: env + guardrails + prefix + agent + suffix
│   Assembly      │           + tools + skills
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Fantasy       │  Provider-agnostic LLM calls (streaming, tools)
│   Agent         │  Handles tool execution loop
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Tool Handler  │  Execute bash, delegate to agents, update tasks
│   (internal/run)│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   UI Rendering  │  Spinners, markdown, styled output
│   (internal/ui) │
└─────────────────┘
```

## Features

- Single built-in agent (`@ayo`) handles all tasks including agent/skill management
- Interactive chat sessions within the terminal
- Non-interactive single-prompt mode
- Bash tool as default for task execution
- System prompts assembled from prefix, shared, agent, tools, skills, and suffix
- Configurable models via Catwalk/Fantasy
- Persistent sessions and semantic memory search

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
  "default_model": "gpt-5.2",
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
├── ayo.db                        # Main database (sessions, messages, memories)
├── agents/                       # Built-in agents
│   └── @ayo/
│       ├── config.json
│       ├── system.md
│       └── skills/
├── skills/                       # Built-in shared skills
│   └── debugging/
│       └── SKILL.md
├── tools/                        # Stateful tool data (per-tool databases)
│   └── todo/
│       └── todo.db               # Todo tool's session data
├── plugins/                      # Installed plugins
└── .builtin-version              # Version marker
```

**Dev mode (running from source checkout):**
```
~/Code/ayo/                       # Your checkout
├── .local/share/ayo/             # Built-in data (project-local)
│   ├── agents/
│   ├── skills/
│   └── .builtin-version
├── .config/ayo/                  # Project-local config
│   ├── agents/
│   ├── skills/
│   └── ayo.json
└── ...
```

This allows multiple dev branches to have isolated built-ins and configuration.

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
ayo                         # Start interactive chat with @ayo
ayo "tell me a joke"        # Run single prompt with @ayo
ayo @myagent                # Start interactive chat with specific agent
ayo @myagent "do task"      # Run single prompt with specific agent

# Agent management (conversational approach - recommended)
ayo "help me create an agent for code review"
ayo "show me my agents"

# Agent management (CLI approach)
ayo agents list                        # List available agents
ayo agents show @ayo                   # Show agent details
ayo agents create @myagent -m gpt-5.2  # Create new agent
ayo agents update                      # Update built-in agents

# Skills management
ayo skills list             # List available skills
ayo skills show <name>      # Show skill details
ayo skills create <name>    # Create new skill
ayo skills validate <path>  # Validate skill directory
ayo skills update           # Update built-in skills

# Sessions management
ayo sessions list           # List conversation sessions
ayo sessions list -a @ayo   # Filter by agent
ayo sessions show <id>      # Show session details and messages
ayo sessions continue       # Continue a session (interactive picker)
ayo sessions continue <id>  # Continue a specific session
ayo sessions delete <id>    # Delete a session

# Memory management
ayo memory list             # List all memories
ayo memory list -a @ayo     # Filter by agent
ayo memory search <query>   # Search memories semantically
ayo memory show <id>        # Show memory details
ayo memory store <content>  # Store a new memory
ayo memory forget <id>      # Forget a memory (soft delete)
ayo memory stats            # Show memory statistics
ayo memory clear            # Clear all memories (with confirmation)

# System diagnostics
ayo doctor                  # Check system health and dependencies
ayo doctor -v               # Verbose output with model list

# Flows management
ayo flows list              # List available flows
ayo flows show <name>       # Show flow details
ayo flows run <name> [input] # Execute a flow
ayo flows new <name>        # Create new flow
ayo flows validate <path>   # Validate flow file
ayo flows history           # Show flow run history
ayo flows replay <run-id>   # Replay a flow run
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

### Memory

Memories are persistent facts, preferences, and patterns learned about users that help agents provide more personalized and contextual responses across sessions.

**How it works:**
1. Agents use a small local LLM (ministral-3:3b) to detect memorable information during conversations
2. Memories are stored with vector embeddings (nomic-embed-text) for semantic search
3. Relevant memories are automatically retrieved and injected into system prompts at session start
4. Agents can also use the `memory` tool to search, store, list, or forget memories

**Memory categories (auto-detected):**
- `preference`: User preferences (tools, styles, communication)
- `fact`: Facts about user or project
- `correction`: User corrections to agent behavior
- `pattern`: Observed behavioral patterns

When storing a memory via CLI or tool, the category is automatically detected if not specified.

**Memory scopes:**
- **Global**: Applies to all agents
- **Agent-scoped**: Applies only to specific agent
- **Path-scoped**: Applies to specific project/directory

**Agent configuration:**
```json
{
  "memory": {
    "enabled": true,
    "scope": "hybrid",
    "formation_triggers": {
      "on_correction": true,
      "on_preference": true,
      "on_project_fact": true,
      "explicit_only": false
    },
    "retrieval": {
      "auto_inject": true,
      "threshold": 0.3,
      "max_memories": 10
    }
  }
}
```

**Ollama-based embedding and extraction:**
Memory uses Ollama for both embeddings and intelligent extraction:
- **Embedding model**: nomic-embed-text (default) - for semantic similarity search
- **Small model**: ministral-3:3b (default) - for extracting memorable content and auto-categorization

Both models are installed during `ayo setup`.

**Storage:** Memories are stored in SQLite (`~/.local/share/ayo/ayo.db`) with vector embeddings as BLOBs. Similarity search is performed in Go without requiring external vector databases.

**Memory tool:** Agents with `memory` in their `allowed_tools` can use the memory tool to:
- `search`: Find relevant memories semantically
- `store`: Save new information
- `list`: Show all memories
- `forget`: Remove a memory

### Flows

Flows are composable agent pipelines - shell scripts with structured frontmatter that orchestrate agent calls. They are the unit of work that external systems (cron, CI, webhooks) invoke.

**Flow file format:**
```bash
#!/usr/bin/env bash
# ayo:flow
# name: my-flow
# description: What this flow does

set -euo pipefail
INPUT="${1:-$(cat)}"
echo "$INPUT" | ayo @ayo "Process this and return JSON"
```

**Directory structure:**
- Project flows: `.ayo/flows/`
- User flows: `~/.config/ayo/flows/`
- Built-in flows: `~/.local/share/ayo/flows/`

**CLI commands:**
```bash
ayo flows list                    # List all flows
ayo flows show <name>             # Show flow details
ayo flows run <name> [input]      # Run a flow
ayo flows run <name> -i file.json # Run with input file
ayo flows new <name>              # Create new flow
ayo flows new <name> --with-schemas # With input/output schemas
ayo flows validate <path>         # Validate flow file
ayo flows history                 # Show run history
ayo flows history --flow=myflow   # Filter by flow
ayo flows replay <run-id>         # Replay a previous run
```

**Structured I/O:** Flows can define JSON schemas for type-safe input/output by placing `input.jsonschema` and `output.jsonschema` alongside the flow script.

**Exit codes:**
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Input validation failed |
| 124 | Timeout |

**History retention:** Configured via `flows.history_retention_days` (default 30) and `flows.history_max_runs` (default 1000) in `ayo.json`.

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

### Tool Categories

Ayo supports **tool categories** - semantic slots that can be filled by different tool implementations. This allows users to swap planning tools, shell interpreters, or other tools without modifying agent configurations.

**Defined categories:**

| Category | Default | Description |
|----------|---------|-------------|
| `planning` | `todo` | Task tracking during execution |
| `shell` | `bash` | Command execution |
| `search` | (none) | Web search (requires plugin) |

**Resolution order:**
1. Check `default_tools` in config for user override
2. Use built-in default if category has one
3. Fall back to literal tool name

**Configuration:**
```json
// ~/.config/ayo/ayo.json
{
  "default_tools": {
    "search": "searxng"     // Set default for category with no built-in
  }
}
```

**Agent config:**
```json
{
  "allowed_tools": ["bash", "planning"]  // "planning" resolves to configured tool
}
```

### Bash Tool

The `bash` tool is the default and primary tool. Agents use it to accomplish any task unless a more specific skill is available.

When calling bash, the LLM must provide:
- `command`: The shell command to execute
- `description`: Human-readable description shown in the spinner (e.g., "Running test suite")

Optional parameters:
- `timeout_seconds`: Command timeout (default 30s)
- `working_dir`: Working directory scoped to project root

### Todo Tool (Default Planning)

The `todo` tool is the **default** for the `planning` category. It provides a flat todo list for tracking multi-step tasks.

**Storage:** Todo data is stored in a dedicated SQLite database at `~/.local/share/ayo/tools/todo/todo.db`, keyed by session ID.

**Parameters:**
```json
{
  "todos": [
    {
      "content": "What needs to be done (imperative form)",
      "active_form": "Present continuous form (e.g., 'Running tests')",
      "status": "pending | in_progress | completed"
    }
  ]
}
```

**Todo states:**
- `pending`: Not yet started
- `in_progress`: Currently working on (limit to ONE todo at a time)
- `completed`: Finished successfully

**Rules:**
- Each todo requires both `content` (imperative) and `active_form` (present continuous)
- Exactly ONE todo should be `in_progress` at any time
- Mark todos complete IMMEDIATELY after finishing
- Remove irrelevant todos from the list entirely
- The full todo list is provided on each call (replacement, not incremental)

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
- `project-summary` - Project analysis and documentation (for @ayo)

## System Prompt Assembly

Messages are built in order:
1. Environment context (platform, date, git status)
2. Guardrails (if enabled)
3. User prefix (optional `~/.config/ayo/prompts/system-prefix.md`)
4. Agent system prompt
5. User suffix (optional `~/.config/ayo/prompts/system-suffix.md`)
6. Tools prompt (bash instructions)
7. Skills prompt (available skills XML)
8. User message

## Guardrails

Guardrails are safety constraints automatically applied to agent system prompts. They enforce rules like:
- No malicious code creation
- No credential exposure
- Confirmation before destructive actions
- Scope limitation to current project

### Configuration

Guardrails are enabled by default. To disable (dangerous):

```json
{
  "guardrails": false
}
```

**Note:** Agents in the `@ayo` namespace always have guardrails enabled regardless of this setting.

### CLI Flag

When creating agents via CLI:

```bash
# Disable guardrails (not recommended)
ayo agents create @dangerous-agent --no-guardrails -n
```

### Custom Prompts

Users can add custom prefix/suffix prompts that layer on top of guardrails:

- `~/.config/ayo/prompts/system-prefix.md` - Added after guardrails, before agent prompt
- `~/.config/ayo/prompts/system-suffix.md` - Added after agent prompt

These are optional user customizations, not replacements for guardrails.

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

The `ayo` namespace is reserved - users cannot create agents with the `@ayo` handle or `@ayo.` prefix.

To create or manage agents, ask `@ayo` directly:
```bash
ayo "help me create an agent for code review"
```

### Available via Plugins

- `@research` - Research assistant with web search (install: `ayo plugins install .../ayo-plugins-research`)
- `@crush` - Coding agent powered by Crush (install: `ayo plugins install .../ayo-plugins-crush`)

## Plugin System

Ayo supports plugins distributed via git repositories. Plugins can provide agents, skills, and tools.

### Repository Naming Convention

Plugin repositories must be named `ayo-plugins-<name>`:
- `ayo-plugins-crush` for the "crush" plugin
- `ayo-plugins-research` for a "research" plugin

### Plugin Structure

```
ayo-plugins-<name>/
├── manifest.json           # Required: plugin metadata
├── agents/                  # Optional: agent definitions
│   └── @agent-name/
│       ├── config.json
│       └── system.md
├── skills/                  # Optional: shared skills
│   └── skill-name/
│       └── SKILL.md
└── tools/                   # Optional: external tools
    └── tool-name/
        └── tool.json
```

### manifest.json

```json
{
  "name": "crush",
  "version": "1.0.0",
  "description": "Crush coding agent for ayo",
  "author": "alexcabrera",
  "repository": "https://github.com/alexcabrera/ayo-plugins-crush",
  "agents": ["@crush"],
  "skills": ["crush-coding"],
  "tools": ["crush"],
  "delegates": {
    "coding": "@crush"
  },
  "default_tools": {
    "search": "searxng"
  },
  "dependencies": {
    "binaries": ["crush"]
  },
  "ayo_version": ">=0.2.0"
}
```

| Field | Description |
|-------|-------------|
| `delegates` | Task types this plugin's agents handle (prompts user to set as global) |
| `default_tools` | Tool aliases this plugin provides (prompts user to set as default) |

### External Tools (tool.json)

External tools map CLI commands to Fantasy tool definitions:

```json
{
  "name": "my-tool",
  "description": "What this tool does",
  "command": "my-binary",
  "args": ["--flag"],
  "parameters": [
    {
      "name": "input",
      "description": "Input text",
      "type": "string",
      "required": true
    }
  ],
  "timeout": 60,
  "working_dir": "param",
  "depends_on": ["required-binary"],
  "spinner_style": "default"
}
```

| Field | Description |
|-------|-------------|
| `name` | Tool identifier used in agent configs |
| `description` | Brief description for the LLM |
| `command` | Executable to run (binary name or path) |
| `args` | Default arguments with `{{param}}` placeholders |
| `parameters` | Input schema for the tool |
| `timeout` | Timeout in seconds (0 = no timeout) |
| `working_dir` | `inherit` (default), `plugin`, or `param` |
| `depends_on` | Required binaries that must be in PATH |
| `spinner_style` | `default` (dots), `crush` (fancy), or `none` |

### CLI Commands

```bash
# Install from git (full URL required)
ayo plugins install https://github.com/owner/ayo-plugins-name.git
ayo plugins install git@gitlab.com:org/ayo-plugins-tools.git
ayo plugins install --local ./my-plugin  # For development

# Management
ayo plugins list           # List installed plugins
ayo plugins show <name>    # Show plugin details
ayo plugins update         # Update all plugins
ayo plugins update <name>  # Update specific plugin
ayo plugins remove <name>  # Uninstall plugin
```

### Installation Locations

- Plugins: `~/.local/share/ayo/plugins/<name>/`
- Registry: `~/.local/share/ayo/packages.json`

### Conflict Resolution

When installing a plugin that conflicts with existing agents/skills:
- User is prompted to choose: skip, replace, or rename
- Renames are tracked in the registry for resolution

## Delegation System

Agents can delegate specific task types to other agents. Delegation is configured at three levels (highest priority first):

### 1. Directory Config (`.ayo.json`)

Project-level configuration file placed in your project root or any parent directory:

```json
{
  "delegates": {
    "coding": "@crush",
    "research": "@research"
  },
  "model": "gpt-5.2",
  "agent": "@ayo"
}
```

| Field | Description |
|-------|-------------|
| `delegates` | Task type to agent handle mappings |
| `model` | Override the default model for this directory |
| `agent` | Default agent for this directory |

### 2. Agent Config (`config.json`)

User-defined agents can specify delegates in their `config.json`:

```json
{
  "delegates": {
    "coding": "@crush"
  }
}
```

**Note:** Built-in agents do not support the `delegates` field. To configure delegation for built-in agents, use directory config or global config.

### 3. Global Config (`~/.config/ayo/ayo.json`)

```json
{
  "delegates": {
    "coding": "@crush"
  }
}
```

### Task Types

| Type | Description |
|------|-------------|
| `coding` | Source code creation/modification |
| `research` | Web research and information gathering |
| `debug` | Debugging and troubleshooting |
| `test` | Test creation and execution |
| `docs` | Documentation generation |

### Plugin-Provided Delegates

Plugins can declare delegates in their `manifest.json`. When installed, users are prompted to set these as global defaults:

```json
{
  "name": "crush",
  "delegates": {
    "coding": "@crush"
  }
}
```

This allows plugins to automatically configure delegation for the task types they handle.

## Tool Aliases

Tool aliases allow agents to use generic tool types (like `search`) that resolve to user-configured concrete tools (like `searxng`). This enables swappable implementations for common tool categories.

### Configuration

Configure default tools in `~/.config/ayo/ayo.json`:

```json
{
  "default_tools": {
    "search": "searxng"
  }
}
```

### How It Works

1. Agent config specifies `allowed_tools: ["search"]`
2. At runtime, `search` resolves to `searxng` (or whatever is configured)
3. If no search provider is configured, the tool is not available to the agent
4. `@ayo` includes `search` by default - just install a search provider to enable

### Behavior with Delegates

When both a tool alias and a delegate are available for the same capability:

- **Research delegate configured**: `@ayo` delegates research tasks to `@research` for thorough, citation-based research
- **No research delegate, but search available**: `@ayo` uses search tool directly for quick lookups
- **Neither available**: `@ayo` informs user that web search is not configured
3. User can swap implementations without modifying agent configs

### Tool Types

| Type | Description |
|------|-------------|
| `search` | Web search (e.g., searxng, duckduckgo) |

### Plugin-Provided Tools

Plugins can declare `default_tools` in their `manifest.json`. When installed, users are prompted to set these as defaults:

```json
{
  "name": "searxng",
  "tools": ["searxng"],
  "default_tools": {
    "search": "searxng"
  }
}
```

## Crush Integration (via Plugin)

For complex coding tasks, install the crush plugin:

```bash
ayo plugins install https://github.com/alexcabrera/ayo-plugins-crush
```

### Prerequisites

Crush must be installed and available in your PATH:
```bash
go install github.com/charmbracelet/crush@latest
```

### Usage

Direct invocation:
```bash
ayo @crush "Add comprehensive error handling to the database layer"
```

Via delegation (configure in `.ayo.json` or agent config):
```bash
ayo "Refactor the authentication module to use JWT tokens"
# @ayo will delegate this to @crush via the coding skill
```

### Configuration

Add to `.ayo.json` in your project:
```json
{
  "delegates": {
    "coding": "@crush"
  }
}
```

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
# Chain two agents: first agent's output pipes to second agent's input
ayo @code-reviewer '{"repo":".", "files":["main.go"]}' | ayo @issue-reporter
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
ayo chain inspect @myagent

# Find agents that can receive this agent's output
ayo chain from @code-reviewer

# Find agents whose output this agent can receive
ayo chain to @issue-reporter

# Validate JSON against agent's input schema
ayo chain validate @myagent '{"environment": "staging", "service": "api"}'
echo '{"environment": "staging", "service": "api"}' | ayo chain validate @myagent

# Generate example input for an agent
ayo chain example @myagent
```

### Chain Context

When agents are chained, context is passed via environment variable:
- `AYO_CHAIN_CONTEXT` contains JSON with `depth`, `source`, and `source_description`
- Freeform agents receive a preamble describing the chain context

### Example: Creating Chainable Agents

To create agents that can chain together:

```bash
# Create an agent with input/output schemas
ayo agents create @code-reviewer

# Add input.jsonschema to define expected input
# Add output.jsonschema to define structured output
```

Then you can chain them:

```bash
ayo @code-reviewer '{"repo":".", "files":["main.go"]}' | ayo @issue-reporter
```
