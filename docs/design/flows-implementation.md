# Flows Implementation Plan

> **Status**: Implemented  
> **Last Updated**: January 2025  
> **Prerequisite**: [flows.md](./flows.md)  
> **Detailed Breakdown**: [flows-stories.md](./flows-stories.md)

## Executive Summary

This document provides an exhaustive, hierarchical, milestone-based implementation plan for the Flows system. The plan is organized into **6 milestones** across **4 phases**, with each milestone containing discrete work items that can be independently tested and deployed.

### Implementation Status

| Phase | Status | Notes |
|-------|--------|-------|
| Phase 1: Foundation | Complete | Core types, discovery, `list`, `show` |
| Phase 2: Execution | Complete | Run engine, validation, history, `run`, `history`, `replay` |
| Phase 3: Authoring | Complete | `new`, flows skill for @ayo |
| Phase 4: Integration | Partial | CLI skill updated, docs updated; chain/session integration deferred |

### Key Design Decisions (from planning session)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Run history** | Store in ayo database | Enables debugging, replay, analytics without requiring orchestrator persistence |
| **Streaming** | Buffer stdout, stream stderr | Human-in-the-loop interactions require real-time feedback |
| **Flow-to-flow** | Naturally supported | Flows are shell scripts; `ayo flows run` works inside flows |
| **Built-in examples** | Rich cross-domain set | Ship many examples, expect user to discard most |

---

## Architecture Overview

### New Packages

```
internal/
├── flows/
│   ├── flow.go           # Core Flow type and loading
│   ├── discover.go       # Discovery from filesystem
│   ├── frontmatter.go    # Frontmatter parsing
│   ├── validate.go       # Validation logic
│   ├── execute.go        # Execution engine
│   └── history.go        # Run history service
```

### Database Changes

New migration: `002_flows.sql`
- `flow_runs` table for run history
- `flow_run_outputs` table for captured I/O

### CLI Commands

```
cmd/ayo/
├── flows.go              # Parent command
├── flows_list.go         # ayo flows list
├── flows_show.go         # ayo flows show <name>
├── flows_run.go          # ayo flows run <name> [input]
├── flows_new.go          # ayo flows new <name>
├── flows_validate.go     # ayo flows validate <path>
├── flows_history.go      # ayo flows history [name]
├── flows_replay.go       # ayo flows replay <run-id>
```

### Example Flows

Examples are **documentation only** (not embedded in binary). Three examples in `docs/guides/flows.md` demonstrate key patterns:

1. **Sequential pipeline**: A → B chain
2. **Conditional logic**: If/else based on agent output  
3. **External integration**: Flow calling external APIs/tools

---

## Phase 1: Foundation

**Goal**: Basic flow discovery, loading, and listing.

### Milestone 1.1: Core Types and Discovery

**Duration**: 1-2 days  
**Dependencies**: None

#### 1.1.1 Define Flow Type

Create `internal/flows/flow.go`:

```go
package flows

// Flow represents a discovered flow
type Flow struct {
    Name        string            // From frontmatter or filename
    Description string            // From frontmatter
    Path        string            // Absolute path to flow.sh or name.sh
    Dir         string            // Parent directory
    Source      FlowSource        // builtin, user, project
    
    // Optional schemas (nil if not present)
    InputSchema  *jsonschema.Schema
    OutputSchema *jsonschema.Schema
    
    // Metadata
    Version string
    Author  string
    
    // Parsed but not validated
    Raw FlowRaw
}

type FlowSource string

const (
    FlowSourceBuiltin FlowSource = "built-in"
    FlowSourceUser    FlowSource = "user"
    FlowSourceProject FlowSource = "project"
)

type FlowRaw struct {
    Frontmatter map[string]string
    Script      string // Everything after frontmatter
}
```

**Acceptance Criteria**:
- [ ] Flow struct defined with all fields
- [ ] FlowSource enum defined
- [ ] Unit tests for type construction

#### 1.1.2 Implement Frontmatter Parser

Create `internal/flows/frontmatter.go`:

```go
// ParseFrontmatter extracts metadata from a flow file
func ParseFrontmatter(content []byte) (FlowRaw, error)

// ValidateFrontmatter checks required fields
func ValidateFrontmatter(fm map[string]string) error
```

Grammar (from flows.md):
```
flow_file     = shebang frontmatter script
shebang       = "#!/usr/bin/env bash" newline
frontmatter   = marker (metadata)*
marker        = "# ayo:flow" newline
metadata      = "# " key ":" value newline
```

**Acceptance Criteria**:
- [ ] Parses shebang line
- [ ] Detects `# ayo:flow` marker
- [ ] Extracts key-value metadata
- [ ] Stops parsing at first non-comment line
- [ ] Returns script content (everything after frontmatter)
- [ ] Error on missing marker
- [ ] Error on missing required fields (name, description)
- [ ] Unit tests with various edge cases

#### 1.1.3 Implement Flow Discovery

Create `internal/flows/discover.go`:

```go
// Discover finds all flows in the given directories
func Discover(dirs []string) ([]Flow, error)

// DiscoverOne loads a single flow by path
func DiscoverOne(path string) (*Flow, error)
```

Discovery rules:
1. Files ending in `.sh` directly in `flows/` are simple flows
2. Directories containing `flow.sh` are flow packages
3. Skip files without `# ayo:flow` marker
4. Load schemas from sibling files if present

**Acceptance Criteria**:
- [ ] Discovers simple flows (*.sh files)
- [ ] Discovers packaged flows (dirs with flow.sh)
- [ ] Skips non-flow shell scripts (no marker)
- [ ] Loads input.jsonschema if present
- [ ] Loads output.jsonschema if present
- [ ] Deduplicates by name (first found wins)
- [ ] Handles missing directories gracefully
- [ ] Unit tests with test fixtures

#### 1.1.4 Add Path Resolution

Update `internal/paths/paths.go`:

```go
func UserFlowsDir() string      // ~/.config/ayo/flows
func BuiltinFlowsDir() string   // ~/.local/share/ayo/flows
func ProjectFlowsDir() string   // ./.ayo/flows (if exists)
```

**Acceptance Criteria**:
- [ ] Functions return correct paths per platform
- [ ] Consistent with existing path patterns
- [ ] Unit tests

### Milestone 1.2: CLI Discovery Commands

**Duration**: 1 day  
**Dependencies**: Milestone 1.1

#### 1.2.1 Implement `ayo flows` Parent Command

Create `cmd/ayo/flows.go`:

```go
func newFlowsCmd(cfgPath *string) *cobra.Command
```

**Acceptance Criteria**:
- [ ] Parent command with description
- [ ] Registers subcommands
- [ ] Shows help when run without subcommand

#### 1.2.2 Implement `ayo flows list`

Create `cmd/ayo/flows_list.go`:

```bash
$ ayo flows list
NAME            SOURCE     DESCRIPTION                              INPUT   OUTPUT
code-review     built-in   Review code and create GitHub issues     yes     yes
daily-standup   user       Generate daily standup from git log      no      yes
research        project    Research a topic and summarize           yes     no

$ ayo flows list --source=user
$ ayo flows list --json
```

**Acceptance Criteria**:
- [ ] Lists all discovered flows
- [ ] Shows source (built-in, user, project)
- [ ] Shows schema presence (yes/no)
- [ ] `--source` flag to filter
- [ ] `--json` flag for machine output
- [ ] Styled table output
- [ ] Empty state message

#### 1.2.3 Implement `ayo flows show`

Create `cmd/ayo/flows_show.go`:

```bash
$ ayo flows show code-review
Name:        code-review
Description: Review code and create GitHub issues
Source:      built-in
Path:        ~/.local/share/ayo/flows/examples/code-review/flow.sh
Version:     1.0.0
Author:      ayo

Input Schema:
  repo: string (required) - Repository path
  files: array of strings - Specific files to review

Output Schema:
  status: string - "clean" or "issues_found"
  findings: array of objects
    - file: string
    - line: number
    - severity: string
    - message: string

Script Preview:
  #!/usr/bin/env bash
  # ayo:flow
  ...
```

**Acceptance Criteria**:
- [ ] Shows all flow metadata
- [ ] Pretty-prints JSON schemas
- [ ] Shows script preview (first N lines)
- [ ] Error if flow not found
- [ ] `--json` flag for machine output
- [ ] `--script` flag to show full script

---

## Phase 2: Execution Engine

**Goal**: Execute flows with proper I/O handling, environment setup, and run history.

### Milestone 2.1: Basic Execution

**Duration**: 2-3 days  
**Dependencies**: Phase 1

#### 2.1.1 Implement Execution Engine

Create `internal/flows/execute.go`:

```go
type RunOptions struct {
    Input       string        // JSON input (from arg or stdin)
    InputFile   string        // Path to input file
    Timeout     time.Duration // Execution timeout
    WorkingDir  string        // Override working directory
    Validate    bool          // Validate only, don't run
    Environment map[string]string // Additional env vars
}

type RunResult struct {
    RunID      string        // Unique run identifier
    Flow       *Flow         // Flow that was run
    Status     RunStatus     // success, error, timeout, validation_failed
    ExitCode   int           // Shell exit code
    Stdout     string        // Captured stdout (JSON)
    Stderr     string        // Captured stderr (logs)
    StartTime  time.Time
    EndTime    time.Time
    Duration   time.Duration
    InputUsed  string        // Actual input JSON
    Error      error         // Error if any
}

type RunStatus string

const (
    RunStatusSuccess          RunStatus = "success"
    RunStatusError            RunStatus = "error"
    RunStatusTimeout          RunStatus = "timeout"
    RunStatusValidationFailed RunStatus = "validation_failed"
)

// Run executes a flow and returns the result
func Run(ctx context.Context, flow *Flow, opts RunOptions) (*RunResult, error)
```

Execution steps:
1. Resolve input (arg > stdin > file)
2. Validate input against schema (if present)
3. Set up environment variables
4. Execute script with timeout
5. Capture stdout/stderr separately
6. Parse and validate output (if schema present)
7. Return structured result

**Acceptance Criteria**:
- [ ] Executes shell script
- [ ] Passes input as first argument
- [ ] Captures stdout separately from stderr
- [ ] Respects timeout
- [ ] Returns structured result
- [ ] Propagates exit codes correctly
- [ ] Unit tests with mock scripts

#### 2.1.2 Environment Variable Injection

Set up execution environment:

```bash
AYO_FLOW_NAME=code-review
AYO_FLOW_RUN_ID=run_abc123
AYO_FLOW_INPUT_FILE=/tmp/ayo-flow-input.json
AYO_FLOW_DIR=/path/to/flow/directory
AYO_VERSION=0.2.0
```

**Acceptance Criteria**:
- [ ] All documented env vars set
- [ ] Run ID is unique (ULID or UUID)
- [ ] Input file created if large input
- [ ] Cleanup temp files on completion
- [ ] Unit tests verify env vars

#### 2.1.3 Input Resolution

Implement input resolution order:

```go
func resolveInput(opts RunOptions) (string, error) {
    // 1. Explicit argument
    if opts.Input != "" {
        return opts.Input, nil
    }
    
    // 2. Input file
    if opts.InputFile != "" {
        data, err := os.ReadFile(opts.InputFile)
        return string(data), err
    }
    
    // 3. Stdin (if piped)
    if !terminal.IsTerminal(os.Stdin.Fd()) {
        data, err := io.ReadAll(os.Stdin)
        return string(data), err
    }
    
    // 4. Empty input
    return "{}", nil
}
```

**Acceptance Criteria**:
- [ ] Argument takes precedence
- [ ] File input works
- [ ] Stdin piping works
- [ ] Empty default for no input
- [ ] Error on invalid JSON

#### 2.1.4 Implement `ayo flows run`

Create `cmd/ayo/flows_run.go`:

```bash
# Basic usage
$ ayo flows run code-review '{"repo": "."}'

# With stdin
$ echo '{"repo": "."}' | ayo flows run code-review

# With file
$ ayo flows run code-review --input request.json

# Validation only
$ ayo flows run code-review --validate '{"repo": "."}'

# With timeout
$ ayo flows run code-review --timeout 300 '{"repo": "."}'
```

**Acceptance Criteria**:
- [ ] Runs flow with argument input
- [ ] Reads from stdin when piped
- [ ] `--input` flag for file input
- [ ] `--validate` flag for dry run
- [ ] `--timeout` flag (default 5 minutes)
- [ ] Outputs JSON to stdout
- [ ] Streams logs to stderr
- [ ] Correct exit codes (0, 1, 2, 3, 124)
- [ ] Error messages to stderr

### Milestone 2.2: Schema Validation

**Duration**: 1-2 days  
**Dependencies**: Milestone 2.1

#### 2.2.1 Input Schema Validation

```go
func validateInput(input string, schema *jsonschema.Schema) error
```

**Acceptance Criteria**:
- [ ] Validates JSON against schema
- [ ] Clear error messages for violations
- [ ] Returns validation errors (not panic)
- [ ] Skips validation if no schema
- [ ] Exit code 2 on validation failure

#### 2.2.2 Output Schema Validation (Advisory)

```go
func validateOutput(output string, schema *jsonschema.Schema) ([]string, error)
```

**Acceptance Criteria**:
- [ ] Validates output JSON against schema
- [ ] Returns warnings, not errors
- [ ] Logs warnings to stderr
- [ ] Still returns output even if invalid
- [ ] Graceful handling of non-JSON output

#### 2.2.3 Implement `ayo flows validate`

```bash
$ ayo flows validate ./my-flow.sh
✓ Flow file is valid
  Name: my-flow
  Description: My custom flow
  Input schema: not defined
  Output schema: not defined

$ ayo flows validate ./my-flow/
✓ Flow package is valid
  Name: my-flow
  Description: My custom flow
  Input schema: valid (5 properties)
  Output schema: valid (3 properties)

$ ayo flows validate ./broken-flow.sh
✗ Flow validation failed
  - Missing required field: description
  - Frontmatter parse error on line 3
```

**Acceptance Criteria**:
- [ ] Validates single file flows
- [ ] Validates packaged flows
- [ ] Checks frontmatter
- [ ] Validates JSON schemas (if present)
- [ ] Clear success/failure output
- [ ] Exit code 0 on success, 1 on failure

### Milestone 2.3: Run History

**Duration**: 2 days  
**Dependencies**: Milestone 2.1

#### 2.3.1 Database Schema

Create `internal/db/migrations/002_flows.sql`:

```sql
-- +goose Up

-- Flow runs (execution history)
CREATE TABLE flow_runs (
    id TEXT PRIMARY KEY,              -- ULID for ordering
    flow_name TEXT NOT NULL,          -- e.g., "code-review"
    flow_path TEXT NOT NULL,          -- Absolute path at run time
    flow_source TEXT NOT NULL,        -- "built-in", "user", "project"
    
    -- Execution context
    working_dir TEXT,                 -- Working directory
    triggered_by TEXT,                -- "cli", "flow", "session"
    parent_run_id TEXT,               -- Parent flow run (if nested)
    session_id TEXT,                  -- Associated session (if any)
    
    -- Input/Output
    input_json TEXT,                  -- Full input JSON
    output_json TEXT,                 -- Captured stdout
    stderr_log TEXT,                  -- Captured stderr
    
    -- Status
    status TEXT NOT NULL,             -- "running", "success", "error", "timeout", "validation_failed"
    exit_code INTEGER,
    error_message TEXT,
    
    -- Timing
    started_at INTEGER NOT NULL,      -- Unix timestamp
    finished_at INTEGER,
    duration_ms INTEGER,
    
    -- Environment snapshot
    environment_json TEXT,            -- Env vars at execution time
    
    FOREIGN KEY (parent_run_id) REFERENCES flow_runs(id) ON DELETE SET NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL
);

CREATE INDEX idx_flow_runs_name ON flow_runs(flow_name);
CREATE INDEX idx_flow_runs_status ON flow_runs(status);
CREATE INDEX idx_flow_runs_started ON flow_runs(started_at DESC);
CREATE INDEX idx_flow_runs_parent ON flow_runs(parent_run_id);
CREATE INDEX idx_flow_runs_session ON flow_runs(session_id);

-- +goose Down
DROP INDEX IF EXISTS idx_flow_runs_session;
DROP INDEX IF EXISTS idx_flow_runs_parent;
DROP INDEX IF EXISTS idx_flow_runs_started;
DROP INDEX IF EXISTS idx_flow_runs_status;
DROP INDEX IF EXISTS idx_flow_runs_name;
DROP TABLE IF EXISTS flow_runs;
```

**Acceptance Criteria**:
- [ ] Migration runs successfully
- [ ] Indexes for common queries
- [ ] Foreign keys for relationships
- [ ] SQLC queries generated

#### 2.3.2 History Service

Create `internal/flows/history.go`:

```go
type HistoryService struct {
    queries *db.Queries
}

func NewHistoryService(queries *db.Queries) *HistoryService

// Recording
func (s *HistoryService) RecordStart(ctx context.Context, flow *Flow, opts RunOptions) (string, error)
func (s *HistoryService) RecordComplete(ctx context.Context, runID string, result *RunResult) error

// Querying
func (s *HistoryService) GetRun(ctx context.Context, runID string) (*FlowRun, error)
func (s *HistoryService) ListRuns(ctx context.Context, opts ListRunsOptions) ([]FlowRun, error)
func (s *HistoryService) ListRunsByFlow(ctx context.Context, flowName string, limit int) ([]FlowRun, error)
func (s *HistoryService) GetLastRun(ctx context.Context, flowName string) (*FlowRun, error)

// Cleanup
func (s *HistoryService) DeleteRun(ctx context.Context, runID string) error
func (s *HistoryService) PruneOldRuns(ctx context.Context, olderThan time.Time) (int, error)
```

**Acceptance Criteria**:
- [ ] Records run start immediately
- [ ] Updates with result on completion
- [ ] Handles incomplete runs (crashes)
- [ ] Query methods work correctly
- [ ] Prune removes old entries
- [ ] Unit tests with mock DB

#### 2.3.3 Integrate History into Execution

Update `Run()` to record history:

```go
func Run(ctx context.Context, flow *Flow, opts RunOptions, history *HistoryService) (*RunResult, error) {
    // Record start
    runID, err := history.RecordStart(ctx, flow, opts)
    
    // Execute...
    result := executeFlow(...)
    result.RunID = runID
    
    // Record completion
    history.RecordComplete(ctx, runID, result)
    
    return result, nil
}
```

**Acceptance Criteria**:
- [ ] Every run is recorded
- [ ] Incomplete runs marked as "error"
- [ ] RunID returned in result
- [ ] History optional (nil history skips recording)

#### 2.3.4 Implement `ayo flows history`

```bash
$ ayo flows history
RUN ID          FLOW            STATUS    DURATION  STARTED
run_01HQ...     code-review     success   12.3s     2 minutes ago
run_01HP...     daily-standup   error     0.8s      1 hour ago
run_01HP...     code-review     success   45.2s     3 hours ago

$ ayo flows history code-review
(shows history for specific flow)

$ ayo flows history --limit 50
$ ayo flows history --status error
$ ayo flows history --since "1 week ago"
$ ayo flows history --json
```

**Acceptance Criteria**:
- [ ] Lists recent runs
- [ ] Filter by flow name
- [ ] `--limit` flag
- [ ] `--status` filter
- [ ] `--since` filter
- [ ] `--json` for machine output
- [ ] Relative timestamps

#### 2.3.5 Implement `ayo flows replay`

```bash
$ ayo flows replay run_01HQ...
Replaying code-review with original input...
(executes flow with same input)

$ ayo flows replay run_01HQ... --dry-run
Would replay code-review with input:
{"repo": ".", "files": ["main.go"]}
```

**Acceptance Criteria**:
- [ ] Retrieves run by ID
- [ ] Extracts original input
- [ ] Re-executes flow
- [ ] `--dry-run` shows input only
- [ ] Error if run not found
- [ ] Links new run to original (parent_run_id)

---

## Phase 3: Authoring and Built-ins

**Goal**: Scaffolding for new flows, rich set of example flows.

### Milestone 3.1: Flow Scaffolding

**Duration**: 1 day  
**Dependencies**: Phase 1

#### 3.1.1 Implement `ayo flows new`

```bash
$ ayo flows new my-flow
Created: ~/.config/ayo/flows/my-flow.sh

$ ayo flows new my-flow --with-schemas
Created: ~/.config/ayo/flows/my-flow/
  - flow.sh
  - input.jsonschema
  - output.jsonschema

$ ayo flows new my-flow --project
Created: ./.ayo/flows/my-flow.sh
```

Template for simple flow:
```bash
#!/usr/bin/env bash
# ayo:flow
# name: my-flow
# description: TODO: Describe what this flow does

set -euo pipefail

INPUT="${1:-$(cat)}"

# TODO: Implement your flow
echo "$INPUT" | ayo @ayo "Process this input and return JSON"
```

**Acceptance Criteria**:
- [ ] Creates flow in user directory by default
- [ ] `--project` creates in project directory
- [ ] `--with-schemas` creates package structure
- [ ] Generates valid frontmatter
- [ ] Makes file executable
- [ ] Error if flow already exists
- [ ] `--force` to overwrite

### Milestone 3.2: Flows Skill for @ayo

**Duration**: 1 day  
**Dependencies**: Milestone 3.1

#### 3.2.1 Create Flows Skill

Create `internal/builtin/skills/flows/SKILL.md`:

```markdown
---
name: flows
description: Discover, compose, and execute ayo flows for multi-step agent pipelines.
---

# Flows Skill

You can help users work with ayo flows - composable agent pipelines.

## Discovery Commands

- `ayo flows list` - List all available flows
- `ayo flows show <name>` - Show flow details and schemas
- `ayo flows history` - Show recent flow runs

## Execution

- `ayo flows run <name> '<json>'` - Execute a flow
- `ayo flows run <name> --validate '<json>'` - Validate input only

## Authoring

When users want to create flows:
1. Use `ayo flows new <name>` to scaffold
2. Edit the generated script
3. Use `ayo flows validate <path>` to check

## Flow Composition Patterns

### Sequential Pipeline
```bash
echo "$INPUT" | ayo @agent1 | ayo @agent2
```

### Conditional
```bash
RESULT=$(echo "$INPUT" | ayo @agent1)
if echo "$RESULT" | jq -e '.condition'; then
  echo "$RESULT" | ayo @agent2
fi
```

### Parallel (fan-out)
```bash
RESULT1=$(echo "$INPUT" | ayo @agent1 &)
RESULT2=$(echo "$INPUT" | ayo @agent2 &)
wait
jq -n --argjson r1 "$RESULT1" --argjson r2 "$RESULT2" '{a: $r1, b: $r2}'
```
```

**Acceptance Criteria**:
- [ ] Skill documented
- [ ] Added to @ayo's skills
- [ ] @ayo can help with flow discovery
- [ ] @ayo can compose flows

#### 3.2.2 Add Example Flows to Documentation

Create 3 example flows in `docs/guides/flows.md` (documentation only, not embedded):

1. **Sequential Pipeline**: Support ticket → classification → response draft
2. **Conditional Logic**: Code review → create issues only if problems found
3. **External Integration**: Research topic using external API + agent synthesis

**Acceptance Criteria**:
- [ ] 3 examples in docs/guides/flows.md
- [ ] Each demonstrates a distinct pattern
- [ ] Examples are copy-pasteable and work with @ayo

---

## Phase 4: Integration and Polish

**Goal**: Integrate with existing systems, documentation, testing.

### Milestone 4.1: Chain Integration

**Duration**: 1 day  
**Dependencies**: Phase 2

#### 4.1.1 Update Chain Commands

Flows should participate in chain discovery:

```bash
$ ayo chain from @code-reviewer
AGENT/FLOW           COMPATIBILITY
@issue-reporter      exact
code-review          structural (flow)

$ ayo chain ls --include-flows
(shows both agents and flows with schemas)
```

**Acceptance Criteria**:
- [ ] `ayo chain from` includes flows
- [ ] `ayo chain to` includes flows
- [ ] `ayo chain ls` has `--include-flows` flag
- [ ] Compatibility checking works for flows

### Milestone 4.2: Session Integration

**Duration**: 1 day  
**Dependencies**: Phase 2

#### 4.2.1 Link Flow Runs to Sessions

When a flow triggers agent sessions, link them:

```go
// In run/run.go, when creating session
session.FlowRunID = opts.FlowRunID // New field

// In flow execution
result := run.Agent(ctx, agent, prompt, run.Options{
    FlowRunID: currentFlowRunID,
})
```

**Acceptance Criteria**:
- [ ] Sessions record associated flow run
- [ ] `ayo flows show <run-id>` shows linked sessions
- [ ] `ayo sessions list --flow <run-id>` filters

### Milestone 4.3: Documentation

**Duration**: 1 day  
**Dependencies**: All previous

#### 4.3.1 Update AGENTS.md

Add flows section to AGENTS.md with:
- CLI commands reference
- Flow file format
- Example flows
- Orchestrator integration examples

#### 4.3.2 Update README.md

Add flows overview to main README

#### 4.3.3 Create Flows Guide

Create `docs/guides/flows.md` with:
- Getting started tutorial
- Authoring guide
- Best practices
- Troubleshooting

**Acceptance Criteria**:
- [ ] AGENTS.md updated
- [ ] README.md updated  
- [ ] Flows guide written
- [ ] All examples tested

### Milestone 4.4: Testing

**Duration**: 2 days  
**Dependencies**: All previous

#### 4.4.1 Unit Tests

Comprehensive unit tests for:
- Frontmatter parsing edge cases
- Discovery with various directory structures
- Validation logic
- Schema validation
- History service

#### 4.4.2 Integration Tests

End-to-end tests:
- Run simple flow
- Run flow with schemas
- Run flow with piped input
- History recording
- Replay functionality

**Acceptance Criteria**:
- [ ] Unit test coverage >80%
- [ ] All integration tests pass
- [ ] CI runs all tests

---

## Implementation Schedule

| Week | Milestone | Deliverables |
|------|-----------|--------------|
| 1 | 1.1, 1.2 | Core types, discovery, `flows list`, `flows show` |
| 2 | 2.1, 2.2 | Execution engine, schema validation, `flows run` |
| 2 | 2.3 | Run history, `flows history`, `flows replay` |
| 3 | 3.1, 3.2 | `flows new`, flows skill, documentation examples |
| 3 | 4.1-4.4 | Chain integration, session integration, docs, tests |

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Shell portability | Document bash requirement, test on Linux/macOS |
| Large outputs | Truncate in history, full output available in logs |
| Nested flow recursion | Track depth in env, limit to 10 levels |
| Stdin detection edge cases | Use robust terminal detection, fallback to empty |
| Schema validation perf | Compile schemas once, cache |

---

## Success Metrics

| Metric | Target |
|--------|--------|
| Discovery speed | <100ms for 100 flows |
| Execution overhead | <50ms added to flow runtime |
| History query speed | <10ms for recent 100 runs |
| Documentation examples | 3 flows demonstrating key patterns |
| Test coverage | >80% for flows package |

---

## Appendix: File Checklist

### New Files to Create

```
internal/flows/
├── flow.go
├── flow_test.go
├── discover.go
├── discover_test.go
├── frontmatter.go
├── frontmatter_test.go
├── execute.go
├── execute_test.go
├── validate.go
├── validate_test.go
├── history.go
├── history_test.go

cmd/ayo/
├── flows.go
├── flows_list.go
├── flows_show.go
├── flows_run.go
├── flows_new.go
├── flows_validate.go
├── flows_history.go
├── flows_replay.go

internal/db/migrations/
├── 002_flows.sql

internal/db/sql/
├── flow_runs.sql

internal/builtin/skills/flows/
├── SKILL.md

docs/guides/
├── flows.md
```

### Files to Modify

```
internal/paths/paths.go          # Add flow directories
cmd/ayo/root.go                  # Add flows command
AGENTS.md                        # Document flows
README.md                        # Add flows overview
docs/design/flows.md             # Mark as implemented
```
