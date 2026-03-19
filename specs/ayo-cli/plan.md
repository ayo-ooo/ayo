# Ayo CLI Implementation Plan

## Overview

This plan breaks down the ayo CLI implementation into incremental, testable phases. Each phase delivers working functionality that can be verified before moving to the next.

---

## Phase 1: Project Foundation

**Goal**: Set up Go module structure and implement project parsing/validation.

### 1.1 Initialize Go Module

```bash
go mod init github.com/charmbracelet/ayo
```

**Deliverable**: Working Go module with dependencies:
- github.com/charmbracelet/bubbletea
- github.com/charmbracelet/bubbles
- github.com/charmbracelet/lipgloss
- github.com/charmbracelet/fang
- github.com/spf13/cobra
- github.com/BurntSushi/toml

### 1.2 Define Core Types

**File**: `internal/project/types.go`

**Deliverable**: structs for:
- `Project`
- `AgentConfig`
- `ModelRequirements`
- `AgentDefaults`
- `Schema`
- `HookType`
- `Skill`

### 1.3 Implement Config Parser

**File**: `internal/project/config.go`

**Function**: `ParseConfig(path string) (*AgentConfig, error)`

**Deliverable**: Parse `config.toml` into `AgentConfig` struct.

**Test**: Unit test with sample config.toml files.

### 1.4 Implement Schema Parser

**File**: `internal/schema/parser.go`

**Functions**:
- `ParseSchema(data []byte) (*ParsedSchema, error)`
- `GenerateFlags(schema *ParsedSchema) []FlagDef`

**Deliverable**: Parse JSON Schema files and extract CLI flag definitions.

**Test**: Unit tests with various JSON Schema examples including x-cli-* extensions.

### 1.5 Implement Project Parser

**File**: `internal/project/parser.go`

**Functions**:
- `ParseProject(path string) (*Project, error)`
- `ValidateProject(p *Project) []error`

**Deliverable**: Parse entire agent directory, validate structure.

**Test**: Integration tests with sample project directories.

---

## Phase 2: CLI Framework

**Goal**: Implement the ayo CLI commands with Fang styling.

### 2.1 Root Command Setup

**File**: `cmd/ayo/main.go`

**Deliverable**: Cobra root command with Fang styling.

**Test**: `ayo --help` displays styled help.

### 2.2 `ayo fresh` Command

**File**: `internal/cmd/fresh.go`

**Deliverable**: Creates new agent project directory with template files:
- config.toml (with agent name)
- system.md (placeholder)
- .gitignore

**Test**: `ayo fresh my-agent` creates valid project structure.

### 2.3 `ayo checkit` Command

**File**: `internal/cmd/checkit.go`

**Deliverable**: Validates project structure, reports all issues.

**Test**: Runs on valid and invalid projects, reports correctly.

### 2.4 `ayo build` Stub

**File**: `internal/cmd/build.go`

**Deliverable**: Placeholder that validates project (calls checkit logic).

**Test**: `ayo build ./my-agent` validates and reports "not yet implemented".

---

## Phase 3: Code Generation

**Goal**: Generate working Go code from agent definitions.

### 3.1 Type Generator

**File**: `internal/generate/types.go`

**Function**: `GenerateTypes(input, output *ParsedSchema) (string, error)`

**Deliverable**: Generate Go structs from JSON Schemas.

**Test**: Generated code compiles and matches expected output.

### 3.2 CLI Generator

**File**: `internal/generate/cli.go`

**Function**: `GenerateCLI(p *Project, flags []FlagDef) (string, error)`

**Deliverable**: Generate Cobra command definitions with flags from input schema.

**Test**: Generated CLI code compiles and parses flags correctly.

### 3.3 Hook Runner Generator

**File**: `internal/generate/hooks.go`

**Function**: `GenerateHooks(hooks map[HookType]string) (string, error)`

**Deliverable**: Generate hook runner that executes embedded and user hooks.

**Test**: Generated code executes hooks in correct order.

### 3.4 Agent Runtime Generator

**File**: `internal/generate/agent.go`

**Function**: `GenerateAgent(p *Project) (string, error)`

**Deliverable**: Generate agent setup code using Fantasy.

**Test**: Generated code initializes Fantasy agent correctly.

### 3.5 Main Generator

**File**: `internal/generate/main.go`

**Function**: `GenerateMain(p *Project) (string, error)`

**Deliverable**: Generate main.go entry point.

**Test**: Generated main.go compiles and runs.

---

## Phase 4: Template Rendering

**Goal**: Implement prompt template rendering.

### 4.1 Template Parser

**File**: `internal/template/parser.go`

**Function**: `ParseTemplate(data string) (*template.Template, error)`

**Deliverable**: Parse Go templates with custom functions.

**Test**: Parse various template strings.

### 4.2 Template Renderer

**File**: `internal/template/renderer.go`

**Function**: `Render(tmpl *template.Template, data map[string]any) (string, error)`

**Deliverable**: Render templates with input data.

**Test**: Render templates with various data structures.

### 4.3 Custom Template Functions

**File**: `internal/template/funcs.go`

**Deliverable**: Custom template functions:
- `json` - marshal to JSON
- `yaml` - marshal to YAML
- `file` - read file contents
- `env` - get environment variable

**Test**: Functions work correctly in templates.

---

## Phase 5: Build System

**Goal**: Implement complete build process.

### 5.1 Embedding Generator

**File**: `internal/generate/embed.go`

**Function**: `GenerateEmbeds(p *Project) (string, error)`

**Deliverable**: Generate embed.FS declarations for:
- system.md
- prompt.tmpl
- skills/
- hooks/

**Test**: Generated embeds compile and contain correct files.

### 5.2 Build Directory Manager

**File**: `internal/build/manager.go`

**Functions**:
- `CreateBuildDir(p *Project) (string, error)`
- `WriteGeneratedFiles(dir string, files map[string]string) error`
- `Cleanup(dir string) error`

**Deliverable**: Manage temporary build directory.

**Test**: Build directory created and cleaned up correctly.

### 5.3 Go Build Executor

**File**: `internal/build/compiler.go`

**Function**: `Compile(buildDir, outputPath string) error`

**Deliverable**: Execute `go build` with static linking flags.

**Test**: Compiled binary runs on target platform.

### 5.4 Complete `ayo build`

**File**: `internal/cmd/build.go`

**Deliverable**: Full build command implementation:
1. Parse and validate project
2. Generate code
3. Create build directory
4. Write generated files
5. Compile binary
6. Clean up

**Test**: `ayo build ./my-agent` produces working binary.

---

## Phase 6: Model Selection

**Goal**: Implement first-run model selection.

### 6.1 Environment Scanner

**File**: `internal/model/scanner.go`

**Function**: `ScanEnvironment() map[string]string`

**Deliverable**: Detect available providers by checking environment variables.

**Test**: Correctly detects set API keys.

### 6.2 Catwalk Integration

**File**: `internal/model/registry.go`

**Function**: `FilterModels(reqs ModelRequirements, availableProviders []string) []Model`

**Deliverable**: Query Catwalk registry, filter by requirements and available providers.

**Test**: Returns correct models based on requirements.

### 6.3 TUI Model Selector

**File**: `internal/model/tui.go`

**Deliverable**: Bubbletea TUI for model selection:
- List available providers
- List models per provider
- Show model capabilities
- Keyboard navigation
- Selection confirmation

**Test**: TUI displays correctly, selection works.

### 6.4 Flag Model Selector

**File**: `internal/model/flags.go`

**Deliverable**: Non-interactive model selection via `--provider` and `--model` flags.

**Test**: Flags override TUI, config saved correctly.

### 6.5 Config Manager

**File**: `internal/config/manager.go`

**Functions**:
- `Load(agentName string) (*UserConfig, error)`
- `Save(agentName string, cfg *UserConfig) error`
- `Exists(agentName string) bool`

**Deliverable**: Manage `~/.config/agents/<agent-name>.toml`.

**Test**: Config loaded and saved correctly.

---

## Phase 7: Hook System

**Goal**: Implement complete hook execution system.

### 7.1 Hook Runner

**File**: `internal/hooks/runner.go`

**Functions**:
- `NewRunner(embedded map[HookType][]byte, user map[HookType]string) *Runner`
- `Run(ctx context.Context, hookType HookType, payload any) error`

**Deliverable**: Execute hooks in order (embedded then user), blocking.

**Test**: Hooks execute in correct order, blocking works.

### 7.2 Hook Payload Serialization

**File**: `internal/hooks/payload.go`

**Function**: `Serialize(hookType HookType, data any) ([]byte, error)`

**Deliverable**: Create JSON payload for each hook type.

**Test**: Payloads match expected format.

### 7.3 Fantasy Callback Adapter

**File**: `internal/hooks/adapter.go`

**Function**: `NewCallbackAdapter(runner *Runner) *CallbackAdapter`

**Deliverable**: Convert Fantasy callbacks to hook executions.

**Test**: Fantasy events trigger correct hooks.

---

## Phase 8: Skills System

**Goal**: Implement Agent Skills embedding and discovery.

### 8.1 Skills Parser

**File**: `internal/skills/parser.go`

**Function**: `ParseSkills(skillsDir string) ([]Skill, error)`

**Deliverable**: Parse skills directory structure.

**Test**: Correctly identifies skills from directory.

### 8.2 Skills Catalog Generator

**File**: `internal/skills/catalog.go`

**Function**: `GenerateCatalog(skills []Skill) string`

**Deliverable**: Generate catalog section for system message.

**Test**: Catalog lists all skills with descriptions.

### 8.3 Skills System Message Injector

**File**: `internal/skills/injector.go`

**Function**: `InjectSkills(systemMessage string, skills []Skill) string`

**Deliverable**: Append skills catalog/instructions to system message.

**Test**: System message includes skill information.

---

## Phase 9: Output Handling

**Goal**: Implement structured output with schema validation.

### 9.1 Output Validator

**File**: `internal/output/validator.go`

**Function**: `Validate(output any, schema *ParsedSchema) error`

**Deliverable**: Validate output against JSON Schema.

**Test**: Valid outputs pass, invalid outputs fail with errors.

### 9.2 Output Writer

**File**: `internal/output/writer.go`

**Function**: `Write(output any, path string) error`

**Deliverable**: Write JSON to stdout, optionally to file via `--output` flag.

**Test**: Output written correctly to stdout and file.

### 9.3 Fantasy Integration

**File**: `internal/output/fantasy.go`

**Deliverable**: Integrate with Fantasy's `object.Generate[T]` for structured output.

**Test**: Structured output matches schema.

---

## Phase 10: Integration & Polish

**Goal**: End-to-end testing and documentation.

### 10.1 Example Agent

**Deliverable**: Complete example agent with:
- input.jsonschema
- output.jsonschema
- prompt.tmpl
- skills/
- hooks/

**Test**: Example builds and runs successfully.

### 10.2 End-to-End Tests

**File**: `test/e2e/`

**Deliverable**: Integration tests for:
- Full build pipeline
- Generated binary execution
- Model selection (mocked)
- Hook execution
- Skills discovery

### 10.3 Error Messages

**Deliverable**: Clear, actionable error messages for all failure modes.

**Test**: Error messages are helpful and accurate.

### 10.4 Documentation

**Files**:
- `README.md` - Quick start guide
- `docs/agent-definition.md` - Agent definition reference
- `docs/cli-reference.md` - CLI commands reference
- `docs/hooks.md` - Hook system documentation
- `docs/skills.md` - Skills system documentation

---

## Dependency Graph

```
Phase 1 (Foundation)
    ├── 1.1 Go Module
    ├── 1.2 Core Types
    ├── 1.3 Config Parser ──────┐
    ├── 1.4 Schema Parser ──────┤
    └── 1.5 Project Parser ─────┤
                                │
Phase 2 (CLI)                   │
    ├── 2.1 Root Command        │
    ├── 2.2 fresh ──────────────┤
    ├── 2.3 checkit ◄───────────┤
    └── 2.4 build stub          │
                                │
Phase 3 (Code Gen)              │
    ├── 3.1 Type Gen ◄──────────┤
    ├── 3.2 CLI Gen ◄───────────┤
    ├── 3.3 Hook Gen ◄──────────┤
    ├── 3.4 Agent Gen ◄─────────┤
    └── 3.5 Main Gen            │
                                │
Phase 4 (Templates)             │
    ├── 4.1 Parser              │
    ├── 4.2 Renderer            │
    └── 4.3 Functions           │
                                │
Phase 5 (Build)                 │
    ├── 5.1 Embeds              │
    ├── 5.2 Build Dir           │
    ├── 5.3 Compiler            │
    └── 5.4 Complete Build ◄────┘
                                │
Phase 6 (Model Selection)       │
    ├── 6.1 Env Scanner         │
    ├── 6.2 Catwalk             │
    ├── 6.3 TUI                 │
    ├── 6.4 Flags               │
    └── 6.5 Config              │
                                │
Phase 7 (Hooks)                 │
    ├── 7.1 Runner              │
    ├── 7.2 Payload             │
    └── 7.3 Adapter             │
                                │
Phase 8 (Skills)                │
    ├── 8.1 Parser              │
    ├── 8.2 Catalog             │
    └── 8.3 Injector            │
                                │
Phase 9 (Output)                │
    ├── 9.1 Validator           │
    ├── 9.2 Writer              │
    └── 9.3 Fantasy             │
                                │
Phase 10 (Polish)               │
    ├── 10.1 Example            │
    ├── 10.2 E2E Tests          │
    ├── 10.3 Errors             │
    └── 10.4 Docs               │
```

---

## Testing Strategy

Each phase includes:
1. **Unit tests** for individual functions
2. **Integration tests** for component interactions
3. **Example code** that demonstrates usage

### Test Commands

```bash
# Unit tests
go test ./...

# Integration tests
go test ./test/integration/...

# End-to-end tests
go test ./test/e2e/...

# Coverage
go test -cover ./...
```

---

## Success Criteria

Phase is complete when:
1. All code compiles without errors
2. All tests pass
3. Acceptance criteria from design.md are met
4. Code follows existing patterns and conventions

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Fantasy API changes | Pin version, monitor upstream |
| Catwalk registry unavailable | Cache registry data locally |
| Complex JSON Schema features | Start with subset, expand incrementally |
| Cross-platform build issues | Test on multiple platforms early |
| TUI rendering issues | Test in various terminals |

---

## Timeline Estimate

| Phase | Effort |
|-------|--------|
| Phase 1 | Foundation | 1-2 days |
| Phase 2 | CLI Framework | 1 day |
| Phase 3 | Code Generation | 2-3 days |
| Phase 4 | Templates | 1 day |
| Phase 5 | Build System | 2 days |
| Phase 6 | Model Selection | 2-3 days |
| Phase 7 | Hooks | 1-2 days |
| Phase 8 | Skills | 1 day |
| Phase 9 | Output | 1 day |
| Phase 10 | Polish | 2 days |
| **Total** | **14-18 days** |
