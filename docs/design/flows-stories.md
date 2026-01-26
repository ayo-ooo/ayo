# Flows Implementation: Stories & Tasks

> **Parent Document**: [flows-implementation.md](flows-implementation.md)  
> **Status**: Planning  
> **Last Updated**: January 2025

This document breaks down the flows implementation into **Stories** (user-facing value), **Tasks** (developer work items), and **Atomic Units** (individually testable/committable changes).

---

## Summary

| Phase | Stories | Tasks | Atomic Units |
|-------|---------|-------|--------------|
| Phase 1: Foundation | 2 | 7 | 35 |
| Phase 2: Execution | 3 | 13 | 68 |
| Phase 3: Authoring | 2 | 3 | 17 |
| Phase 4: Polish | 3 | 6 | 23 |
| **Total** | **10** | **29** | **143** |

---

## Phase 1: Foundation

### Story 1.1: As a user, I can discover what flows are available

> "I want to see what flows exist so I can understand what's possible."

#### Task 1.1.1: Define Flow Data Model

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 1.1.1.1 | Create `internal/flows/flow.go` with `Flow` struct | Unit: struct instantiation |
| 1.1.1.2 | Define `FlowSource` enum (built-in, user, project) | Unit: enum values |
| 1.1.1.3 | Define `FlowRaw` struct for parsed frontmatter | Unit: struct fields |
| 1.1.1.4 | Add `FlowMetadata` for optional fields (version, author) | Unit: optional handling |

#### Task 1.1.2: Implement Frontmatter Parser

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 1.1.2.1 | Create `internal/flows/frontmatter.go` | File exists |
| 1.1.2.2 | Implement shebang line detection | Unit: `#!/usr/bin/env bash` required |
| 1.1.2.3 | Implement `# ayo:flow` marker detection | Unit: marker present/absent |
| 1.1.2.4 | Parse `# key: value` metadata lines | Unit: various keys |
| 1.1.2.5 | Stop parsing at first non-comment line | Unit: script content preserved |
| 1.1.2.6 | Validate required fields (name, description) | Unit: missing field errors |
| 1.1.2.7 | Handle edge cases (empty lines, trailing spaces) | Unit: edge cases |

#### Task 1.1.3: Implement Flow Discovery

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 1.1.3.1 | Create `internal/flows/discover.go` | File exists |
| 1.1.3.2 | Discover simple flows (`*.sh` files) | Unit: finds .sh files |
| 1.1.3.3 | Discover packaged flows (dirs with `flow.sh`) | Unit: finds dirs |
| 1.1.3.4 | Skip files without `# ayo:flow` marker | Unit: non-flows ignored |
| 1.1.3.5 | Load `input.jsonschema` if present | Unit: schema loaded |
| 1.1.3.6 | Load `output.jsonschema` if present | Unit: schema loaded |
| 1.1.3.7 | Deduplicate by name (first found wins) | Unit: priority order |
| 1.1.3.8 | Handle missing/empty directories gracefully | Unit: no panic |
| 1.1.3.9 | Implement `DiscoverOne(path)` for single flow | Unit: single file load |

#### Task 1.1.4: Add Path Resolution

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 1.1.4.1 | Add `UserFlowsDir()` to `internal/paths/paths.go` | Unit: correct path |
| 1.1.4.2 | Add `BuiltinFlowsDir()` | Unit: correct path |
| 1.1.4.3 | Add `ProjectFlowsDir()` (checks for `.ayo/flows`) | Unit: project detection |
| 1.1.4.4 | Update path tests | All paths tests pass |

---

### Story 1.2: As a user, I can list and inspect flows via CLI

> "I want CLI commands to browse flows without reading files."

#### Task 1.2.1: Create Flows Parent Command

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 1.2.1.1 | Create `cmd/ayo/flows.go` with `newFlowsCmd()` | Command exists |
| 1.2.1.2 | Register in `root.go` | `ayo flows --help` works |
| 1.2.1.3 | Add description and usage | Help text correct |

#### Task 1.2.2: Implement `ayo flows list`

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 1.2.2.1 | Create `cmd/ayo/flows_list.go` | File exists |
| 1.2.2.2 | List all discovered flows in table | Integration: shows flows |
| 1.2.2.3 | Show columns: NAME, SOURCE, DESCRIPTION, INPUT, OUTPUT | Visual check |
| 1.2.2.4 | Add `--source` flag to filter | Unit: filter works |
| 1.2.2.5 | Add `--json` flag for machine output | Unit: valid JSON |
| 1.2.2.6 | Style table with lipgloss | Visual check |
| 1.2.2.7 | Handle empty state gracefully | Unit: no flows message |

#### Task 1.2.3: Implement `ayo flows show`

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 1.2.3.1 | Create `cmd/ayo/flows_show.go` | File exists |
| 1.2.3.2 | Display flow metadata (name, description, path) | Integration: correct output |
| 1.2.3.3 | Pretty-print input schema if present | Unit: schema rendering |
| 1.2.3.4 | Pretty-print output schema if present | Unit: schema rendering |
| 1.2.3.5 | Show script preview (first 20 lines) | Unit: truncation |
| 1.2.3.6 | Add `--json` flag | Unit: valid JSON |
| 1.2.3.7 | Add `--script` flag to show full script | Unit: full content |
| 1.2.3.8 | Error if flow not found | Unit: error message |

---

## Phase 2: Execution Engine

### Story 2.1: As a user, I can run a flow and get JSON output

> "I want to execute a flow and capture structured results."

#### Task 2.1.1: Implement Core Execution Engine

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.1.1.1 | Create `internal/flows/execute.go` | File exists |
| 2.1.1.2 | Define `RunOptions` struct | Unit: struct fields |
| 2.1.1.3 | Define `RunResult` struct | Unit: struct fields |
| 2.1.1.4 | Define `RunStatus` enum | Unit: all statuses |
| 2.1.1.5 | Implement basic script execution | Integration: script runs |
| 2.1.1.6 | Capture stdout separately from stderr | Unit: streams separated |
| 2.1.1.7 | Implement timeout handling | Unit: timeout triggers |
| 2.1.1.8 | Propagate exit codes correctly | Unit: codes match |
| 2.1.1.9 | Generate unique run IDs (ULID) | Unit: IDs unique |

#### Task 2.1.2: Implement Input Resolution

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.1.2.1 | Resolve from explicit argument first | Unit: arg precedence |
| 2.1.2.2 | Resolve from `--input` file | Unit: file read |
| 2.1.2.3 | Resolve from stdin when piped | Unit: stdin detection |
| 2.1.2.4 | Default to `{}` when no input | Unit: empty default |
| 2.1.2.5 | Validate input is valid JSON | Unit: parse errors |

#### Task 2.1.3: Implement Environment Injection

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.1.3.1 | Set `AYO_FLOW_NAME` | Unit: env var set |
| 2.1.3.2 | Set `AYO_FLOW_RUN_ID` | Unit: env var set |
| 2.1.3.3 | Set `AYO_FLOW_DIR` | Unit: env var set |
| 2.1.3.4 | Set `AYO_VERSION` | Unit: env var set |
| 2.1.3.5 | Create temp input file for large inputs | Unit: file created |
| 2.1.3.6 | Set `AYO_FLOW_INPUT_FILE` when temp file used | Unit: env var set |
| 2.1.3.7 | Cleanup temp files on completion | Unit: files removed |

#### Task 2.1.4: Implement `ayo flows run`

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.1.4.1 | Create `cmd/ayo/flows_run.go` | File exists |
| 2.1.4.2 | Accept flow name as first argument | Integration: flow runs |
| 2.1.4.3 | Accept JSON input as second argument | Integration: input passed |
| 2.1.4.4 | Read from stdin when piped | Integration: pipe works |
| 2.1.4.5 | Add `--input` flag for file input | Unit: file read |
| 2.1.4.6 | Add `--timeout` flag (default 5 min) | Unit: timeout configurable |
| 2.1.4.7 | Output JSON to stdout only | Unit: clean stdout |
| 2.1.4.8 | Stream stderr in real-time | Integration: stderr streams |
| 2.1.4.9 | Exit with correct codes (0, 1, 2, 3, 124) | Unit: all exit codes |

---

### Story 2.2: As a user, I can validate flow input/output against schemas

> "I want to catch input errors before running and verify output format."

#### Task 2.2.1: Implement Input Validation

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.2.1.1 | Create `internal/flows/validate.go` | File exists |
| 2.2.1.2 | Add JSON Schema validation library | go.mod updated |
| 2.2.1.3 | Validate input against schema | Unit: valid input passes |
| 2.2.1.4 | Return clear error messages | Unit: error clarity |
| 2.2.1.5 | Skip validation if no schema | Unit: nil schema ok |
| 2.2.1.6 | Exit code 2 on validation failure | Unit: exit code |

#### Task 2.2.2: Implement Output Validation (Advisory)

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.2.2.1 | Validate output JSON against schema | Unit: validation runs |
| 2.2.2.2 | Log warnings to stderr, don't fail | Unit: warnings only |
| 2.2.2.3 | Return output even if invalid | Unit: output preserved |
| 2.2.2.4 | Handle non-JSON output gracefully | Unit: no panic |

#### Task 2.2.3: Add `--validate` Flag

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.2.3.1 | Add `--validate` to `ayo flows run` | Flag exists |
| 2.2.3.2 | Validate input without executing | Unit: no execution |
| 2.2.3.3 | Output validation result | Unit: success/failure output |

#### Task 2.2.4: Implement `ayo flows validate`

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.2.4.1 | Create `cmd/ayo/flows_validate.go` | File exists |
| 2.2.4.2 | Validate single file flows | Integration: file validated |
| 2.2.4.3 | Validate packaged flows | Integration: dir validated |
| 2.2.4.4 | Check frontmatter validity | Unit: frontmatter checked |
| 2.2.4.5 | Check JSON schema validity | Unit: schemas checked |
| 2.2.4.6 | Clear success/failure output | Visual check |
| 2.2.4.7 | Exit code 0 success, 1 failure | Unit: exit codes |

---

### Story 2.3: As a user, I can view history of flow runs

> "I want to see what flows ran, when, and whether they succeeded."

#### Task 2.3.1: Create Database Schema

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.3.1.1 | Create `internal/db/migrations/002_flows.sql` | File exists |
| 2.3.1.2 | Define `flow_runs` table | Migration runs |
| 2.3.1.3 | Add indexes for common queries | Indexes created |
| 2.3.1.4 | Add foreign keys (parent_run_id, session_id) | FKs work |
| 2.3.1.5 | Create `internal/db/sql/flow_runs.sql` SQLC queries | Queries generated |
| 2.3.1.6 | Regenerate SQLC | `go generate` succeeds |

#### Task 2.3.2: Implement History Service

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.3.2.1 | Create `internal/flows/history.go` | File exists |
| 2.3.2.2 | Implement `RecordStart()` | Unit: row created |
| 2.3.2.3 | Implement `RecordComplete()` | Unit: row updated |
| 2.3.2.4 | Implement `GetRun()` | Unit: retrieval works |
| 2.3.2.5 | Implement `ListRuns()` with filters | Unit: filtering works |
| 2.3.2.6 | Implement `ListRunsByFlow()` | Unit: per-flow listing |
| 2.3.2.7 | Implement `GetLastRun()` | Unit: latest returned |
| 2.3.2.8 | Implement `DeleteRun()` | Unit: deletion works |
| 2.3.2.9 | Implement `PruneOldRuns()` | Unit: pruning works |

#### Task 2.3.3: Implement Auto-Pruning

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.3.3.1 | Add `flows.history_retention_days` config (default 30) | Config parsed |
| 2.3.3.2 | Add `flows.history_max_runs` config (default 1000) | Config parsed |
| 2.3.3.3 | Prune on each new run (cheap check) | Unit: prune triggered |
| 2.3.3.4 | Delete by age OR count, whichever is less | Unit: both limits work |

#### Task 2.3.4: Integrate History into Execution

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.3.4.1 | Record start before execution | Unit: start recorded |
| 2.3.4.2 | Record completion after execution | Unit: completion recorded |
| 2.3.4.3 | Handle crashes (incomplete runs) | Unit: status correct |
| 2.3.4.4 | Return RunID in result | Unit: ID populated |
| 2.3.4.5 | Make history optional (nil skips) | Unit: nil safe |

#### Task 2.3.5: Implement `ayo flows history`

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.3.5.1 | Create `cmd/ayo/flows_history.go` | File exists |
| 2.3.5.2 | List recent runs in table | Integration: shows runs |
| 2.3.5.3 | Filter by flow name (positional arg) | Unit: filter works |
| 2.3.5.4 | Add `--limit` flag (default 20) | Unit: limit works |
| 2.3.5.5 | Add `--status` flag | Unit: status filter |
| 2.3.5.6 | Add `--since` flag (duration parsing) | Unit: since filter |
| 2.3.5.7 | Add `--json` flag | Unit: valid JSON |
| 2.3.5.8 | Show relative timestamps ("2 hours ago") | Visual check |
| 2.3.5.9 | Style with lipgloss | Visual check |

#### Task 2.3.6: Implement `ayo flows replay`

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 2.3.6.1 | Create `cmd/ayo/flows_replay.go` | File exists |
| 2.3.6.2 | Look up run by ID | Unit: lookup works |
| 2.3.6.3 | Support ID prefix matching | Unit: partial ID works |
| 2.3.6.4 | Extract original input | Unit: input retrieved |
| 2.3.6.5 | Re-execute flow with same input | Integration: replay works |
| 2.3.6.6 | Add `--dry-run` flag | Unit: shows input only |
| 2.3.6.7 | Link new run to original (parent_run_id) | Unit: link created |
| 2.3.6.8 | Error if run not found | Unit: error message |

---

## Phase 3: Authoring & Documentation

### Story 3.1: As a user, I can scaffold new flows

> "I want to create new flows without memorizing the format."

#### Task 3.1.1: Implement `ayo flows new`

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 3.1.1.1 | Create `cmd/ayo/flows_new.go` | File exists |
| 3.1.1.2 | Generate flow in user directory by default | Integration: file created |
| 3.1.1.3 | Add `--project` flag for `.ayo/flows/` | Unit: project path |
| 3.1.1.4 | Add `--with-schemas` for package structure | Unit: dir created |
| 3.1.1.5 | Generate valid frontmatter | Unit: frontmatter valid |
| 3.1.1.6 | Make file executable (chmod +x) | Unit: executable |
| 3.1.1.7 | Error if flow already exists | Unit: error message |
| 3.1.1.8 | Add `--force` to overwrite | Unit: overwrite works |
| 3.1.1.9 | Generate input.jsonschema template | Unit: schema valid |
| 3.1.1.10 | Generate output.jsonschema template | Unit: schema valid |

---

### Story 3.2: As an agent (@ayo), I can help users with flows

> "The default agent should understand flows and help compose them."

#### Task 3.2.1: Create Flows Skill

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 3.2.1.1 | Create `internal/builtin/skills/flows/SKILL.md` | File exists |
| 3.2.1.2 | Document discovery commands | Content complete |
| 3.2.1.3 | Document execution commands | Content complete |
| 3.2.1.4 | Document authoring workflow | Content complete |
| 3.2.1.5 | Document composition patterns | Content complete |
| 3.2.1.6 | Add skill to @ayo's config | Skill attached |

#### Task 3.2.2: Add Example Flows to Documentation

Examples live in docs only (not embedded/installed). Show 3 patterns:

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 3.2.2.1 | **Sequential pipeline**: Simple A → B chain | Doc complete |
| 3.2.2.2 | **Conditional logic**: If/else based on agent output | Doc complete |
| 3.2.2.3 | **External integration**: Flow that calls external APIs/tools | Doc complete |

**Example 1: Sequential Pipeline** (support ticket → response)
```bash
#!/usr/bin/env bash
# ayo:flow
# name: support-response
# description: Classify ticket and draft response

set -euo pipefail
INPUT="${1:-$(cat)}"

# Classify the ticket
CLASSIFIED=$(echo "$INPUT" | ayo @ayo "Classify this support ticket as billing/technical/general. Return JSON: {\"category\": \"...\", \"priority\": \"low|medium|high\"}")

# Draft response based on classification
echo "$CLASSIFIED" | ayo @ayo "Draft a helpful response for this classified ticket. Return JSON: {\"response\": \"...\", \"suggested_actions\": [...]}"
```

**Example 2: Conditional Logic** (code review with optional issue creation)
```bash
#!/usr/bin/env bash
# ayo:flow
# name: smart-review
# description: Review code, only create issues if problems found

set -euo pipefail
INPUT="${1:-$(cat)}"

REVIEW=$(echo "$INPUT" | ayo @ayo "Review this code diff. Return JSON: {\"findings\": [...], \"severity\": \"none|low|high\"}")

if echo "$REVIEW" | jq -e '.severity != "none"' > /dev/null; then
  echo "$REVIEW" | ayo @ayo "Create GitHub issue descriptions for these findings. Return JSON: {\"issues\": [...]}"
else
  echo '{"status": "clean", "message": "No issues found"}'
fi
```

**Example 3: External Integration** (research with web search)
```bash
#!/usr/bin/env bash
# ayo:flow
# name: research-report
# description: Research topic using web search and summarize

set -euo pipefail
TOPIC="${1:-$(cat)}"

# Use curl to search (or any external tool)
SEARCH_RESULTS=$(curl -s "https://api.example.com/search?q=$(echo "$TOPIC" | jq -r '.query')")

# Synthesize with agent
echo "$SEARCH_RESULTS" | ayo @ayo "Synthesize these search results into a research brief. Return JSON: {\"summary\": \"...\", \"key_points\": [...], \"sources\": [...]}"
```

---

## Phase 4: Integration & Polish

### Story 4.1: Flows integrate with existing ayo features

> "Flows should work seamlessly with chains, sessions, and other features."

#### Task 4.1.1: Chain Integration

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 4.1.1.1 | Include flows in `ayo chain from` output | Integration: flows shown |
| 4.1.1.2 | Include flows in `ayo chain to` output | Integration: flows shown |
| 4.1.1.3 | Add `--include-flows` to `ayo chain ls` | Unit: flag works |
| 4.1.1.4 | Schema compatibility checking for flows | Unit: compat works |

#### Task 4.1.2: Session Integration

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 4.1.2.1 | Add `flow_run_id` column to sessions table | Migration works |
| 4.1.2.2 | Pass flow run ID to agent execution | Unit: ID passed |
| 4.1.2.3 | Show linked sessions in `ayo flows show <run-id>` | Integration: shown |
| 4.1.2.4 | Add `--flow` filter to `ayo sessions list` | Unit: filter works |

---

### Story 4.2: Flows are well-documented

> "Users can learn flows from docs without reading source code."

#### Task 4.2.1: Update Documentation

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 4.2.1.1 | Add Flows section to AGENTS.md | Content complete |
| 4.2.1.2 | Update AGENTS.md CLI reference | Commands documented |
| 4.2.1.3 | Add flows overview to README.md | Content complete |
| 4.2.1.4 | Create `docs/guides/flows.md` tutorial | Guide complete |
| 4.2.1.5 | Add orchestrator integration examples | Examples complete |
| 4.2.1.6 | Update flows.md status to "Implemented" | Status updated |

---

### Story 4.3: Flows are thoroughly tested

> "The flows system is reliable and regression-free."

#### Task 4.3.1: Unit Tests

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 4.3.1.1 | Frontmatter parsing tests (all edge cases) | Coverage >90% |
| 4.3.1.2 | Discovery tests (various structures) | Coverage >90% |
| 4.3.1.3 | Execution tests (mock scripts) | Coverage >90% |
| 4.3.1.4 | Validation tests (schemas) | Coverage >90% |
| 4.3.1.5 | History service tests | Coverage >90% |

#### Task 4.3.2: Integration Tests

| Atomic Unit | Description | Test |
|-------------|-------------|------|
| 4.3.2.1 | Run simple flow end-to-end | Test passes |
| 4.3.2.2 | Run flow with schemas | Test passes |
| 4.3.2.3 | Run flow with piped input | Test passes |
| 4.3.2.4 | History recording end-to-end | Test passes |
| 4.3.2.5 | Replay functionality end-to-end | Test passes |
| 4.3.2.6 | Nested flow execution | Test passes |

---

## Appendix: History Retention Strategy

Based on standard logging patterns:

| Tool | Default Retention |
|------|-------------------|
| logrotate | 4-7 files by size/time |
| systemd journal | 10% of disk, max 4GB |
| Docker | 100MB per container, 5 files |
| CloudWatch | 30 days |

### Ayo Strategy

```json
{
  "flows": {
    "history_retention_days": 30,
    "history_max_runs": 1000
  }
}
```

- **Default**: Keep last 1000 runs OR 30 days, whichever is less
- **Auto-prune**: On each new run (cheap timestamp check)
- **Manual prune**: `ayo flows history prune --before "30 days ago"`
- **Disable**: Set `history_retention_days: 0` to disable auto-prune

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
