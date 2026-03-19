# Test Coverage Improvement - Implementation Plan

## Overview

This plan breaks down test implementation into discrete, verifiable steps. Each step produces working, testable code and increases coverage incrementally.

**Current Coverage:** 37.7%
**Target Coverage:** 80-90%

---

## Phase 1: Foundation (internal/schema + internal/project)

### Step 1.1: Create test utilities package
**Package:** `internal/testutil`
**Files:** `fixtures.go`

**Tasks:**
- [ ] Create `internal/testutil/fixtures.go`
- [ ] Implement `CreateProject(t, name) string` helper
- [ ] Implement `ValidInputSchema() string` helper
- [ ] Implement `ValidOutputSchema() string` helper

**Verification:**
```bash
go test ./internal/testutil/...
```

**Expected coverage:** N/A (utility package)

---

### Step 1.2: Test schema parsing
**Package:** `internal/schema`
**Files:** `parser_test.go`, `testdata/`

**Tasks:**
- [ ] Create `internal/schema/testdata/` directory
- [ ] Add test fixtures (valid_schema.json, invalid_schema.json)
- [ ] Create `parser_test.go`
- [ ] Test `ParseSchema()` with valid object schema
- [ ] Test `ParseSchema()` with valid array schema
- [ ] Test `ParseSchema()` with invalid JSON
- [ ] Test `ParseSchema()` with empty input
- [ ] Test `ParseSchema()` with nested objects
- [ ] Test `GenerateFlags()` with valid schema
- [ ] Test `GenerateFlags()` with nil schema
- [ ] Test `GenerateFlags()` flag name sanitization

**Verification:**
```bash
go test -cover ./internal/schema/
```

**Expected coverage:** 90%+

---

### Step 1.3: Test project parsing
**Package:** `internal/project`
**Files:** `parser_test.go`

**Tasks:**
- [ ] Create `parser_test.go`
- [ ] Test `ParseProject()` with minimal valid project
- [ ] Test `ParseProject()` with full project (all optional files)
- [ ] Test `ParseProject()` with missing config.toml
- [ ] Test `ParseProject()` with missing system.md
- [ ] Test `ParseProject()` with input.jsonschema
- [ ] Test `ParseProject()` with output.jsonschema
- [ ] Test `ParseProject()` with skills directory
- [ ] Test `ParseProject()` with hooks directory
- [ ] Test `parseSkill()` with valid skill
- [ ] Test `isValidHookType()` for all valid types
- [ ] Test `isValidHookType()` for invalid types

**Verification:**
```bash
go test -cover ./internal/project/
```

**Expected coverage:** 85%+ (parser.go)

---

### Step 1.4: Test project types
**Package:** `internal/project`
**Files:** `types_test.go`

**Tasks:**
- [ ] Create `types_test.go`
- [ ] Test `ValidationError.Error()` formatting
- [ ] Test `Schema.UnmarshalJSON()` with valid JSON
- [ ] Test `Schema.UnmarshalJSON()` with invalid JSON
- [ ] Test `Schema` parsed field assignment

**Verification:**
```bash
go test -cover ./internal/project/
```

**Expected coverage:** 90%+ (types.go)

---

### Step 1.5: Test config parsing
**Package:** `internal/project`
**Files:** `config_test.go`

**Tasks:**
- [ ] Create `config_test.go`
- [ ] Test `ParseConfig()` with valid TOML
- [ ] Test `ParseConfig()` with missing file
- [ ] Test `ParseConfig()` with invalid TOML
- [ ] Test config field parsing (name, version, description, model defaults)

**Verification:**
```bash
go test -cover ./internal/project/
```

**Expected coverage:** 90%+ (config.go)

---

### Step 1.6: Test project validation
**Package:** `internal/project`
**Files:** `validation_test.go` (if exists) or add to `parser_test.go`

**Tasks:**
- [ ] Test `ValidateProject()` with valid project
- [ ] Test `ValidateProject()` with missing name
- [ ] Test `ValidateProject()` with empty system.md
- [ ] Test `ValidateProject()` with invalid input schema
- [ ] Test `ValidateProject()` with invalid output schema
- [ ] Test `ValidateProject()` with invalid hook type

**Verification:**
```bash
go test -cover ./internal/project/
```

**Expected coverage:** 90%+ (overall project package)

---

## Phase 2: Code Generation (internal/generate)

### Step 2.1: Test type generation
**Package:** `internal/generate`
**Files:** `types_test.go`

**Tasks:**
- [ ] Create `types_test.go`
- [ ] Test `GenerateTypes()` with input schema only
- [ ] Test `GenerateTypes()` with output schema only
- [ ] Test `GenerateTypes()` with both schemas
- [ ] Test `GenerateTypes()` with nil schemas
- [ ] Test `jsonTypeToGo()` for all JSON types
- [ ] Test `toPascalCase()` with various inputs

**Verification:**
```bash
go test -cover ./internal/generate/
```

**Expected coverage:** 50%+ (types.go fully covered)

---

### Step 2.2: Enhance embed tests
**Package:** `internal/generate`
**Files:** `embed_test.go`

**Tasks:**
- [ ] Add test for `GenerateEmbeds()` with skills
- [ ] Add test for `GenerateEmbeds()` with hooks
- [ ] Add test for `GenerateEmbeds()` with prompt template
- [ ] Add test for `toSafeIdentifier()` edge cases
- [ ] Add test for `GenerateGoMod()` with various names

**Verification:**
```bash
go test -cover ./internal/generate/
```

**Expected coverage:** 60%+ (embed.go fully covered)

---

### Step 2.3: Enhance hooks tests
**Package:** `internal/generate`
**Files:** `hooks_test.go`

**Tasks:**
- [ ] Add test for `GenerateHooks()` with multiple hook types
- [ ] Add test for generated HookRunner struct
- [ ] Add test for hook type constants
- [ ] Add test for serializeHookPayload generation

**Verification:**
```bash
go test -cover ./internal/generate/
```

**Expected coverage:** 65%+

---

### Step 2.4: Add TUI generation tests
**Package:** `internal/generate`
**Files:** `tui_test.go`

**Tasks:**
- [ ] Create `tui_test.go`
- [ ] Test `GenerateTUI()` returns valid Go code
- [ ] Test provider models map generation
- [ ] Test TUI state constants
- [ ] Test style definitions

**Verification:**
```bash
go test -cover ./internal/generate/
```

**Expected coverage:** 75%+

---

## Phase 3: Build & Commands

### Step 3.1: Test build manager (unit tests)
**Package:** `internal/build`
**Files:** `manager_test.go`

**Tasks:**
- [ ] Create `manager_test.go`
- [ ] Test `NewManager()` returns valid manager
- [ ] Test `getSchema()` with nil schema
- [ ] Test `getSchema()` with valid schema
- [ ] Test `copyDirectory()` with nested directories
- [ ] Test `hasDependency()` (if keeping this function)

**Verification:**
```bash
go test -cover ./internal/build/
```

**Expected coverage:** 30%+ (isolated functions)

---

### Step 3.2: Test build manager (integration)
**Package:** `internal/build`
**Files:** `manager_test.go` with build tag

**Tasks:**
- [ ] Add integration test for `Manager.Build()` (requires `//go:build integration`)
- [ ] Test build with minimal project
- [ ] Test `Manager.Cleanup()` removes temp dir
- [ ] Test output path resolution

**Verification:**
```bash
go test -tags=integration -cover ./internal/build/
```

**Expected coverage:** 70%+ (with integration tests)

---

### Step 3.3: Test fresh command
**Package:** `internal/cmd`
**Files:** `fresh_test.go`

**Tasks:**
- [ ] Create `fresh_test.go`
- [ ] Test `createProject()` creates all files
- [ ] Test `createProject()` file contents are correct
- [ ] Test `createProject()` fails on existing directory
- [ ] Test `createProject()` with various names

**Verification:**
```bash
go test -cover ./internal/cmd/
```

**Expected coverage:** 50%+ (fresh.go)

---

### Step 3.4: Test checkit command
**Package:** `internal/cmd`
**Files:** `checkit_test.go`

**Tasks:**
- [ ] Create `checkit_test.go`
- [ ] Test `validateProject()` with valid project
- [ ] Test `validateProject()` with invalid path
- [ ] Test `validateProject()` with file path (not directory)
- [ ] Test error output formatting

**Verification:**
```bash
go test -cover ./internal/cmd/
```

**Expected coverage:** 70%+ (checkit.go + fresh.go)

---

### Step 3.5: Test build command
**Package:** `internal/cmd`
**Files:** `build_test.go`

**Tasks:**
- [ ] Create `build_test.go`
- [ ] Test `buildProject()` validates project first
- [ ] Test `buildProject()` returns error on invalid project
- [ ] Test output path flag handling

**Verification:**
```bash
go test -cover ./internal/cmd/
```

**Expected coverage:** 80%+

---

## Phase 4: Integration Tests

### Step 4.1: Create integration test suite
**Package:** `tests/integration`
**Files:** `workflow_test.go`

**Tasks:**
- [ ] Create `tests/integration/` directory
- [ ] Create `workflow_test.go` with `//go:build integration` tag
- [ ] Test `ayo fresh <name>` command
- [ ] Test `ayo checkit <name>` command
- [ ] Test `ayo build <name>` command
- [ ] Test built binary runs and responds to --help
- [ ] Test full `fresh → checkit → build` workflow

**Verification:**
```bash
go test -tags=integration ./tests/integration/...
```

**Expected coverage:** End-to-end workflow verified

---

## Summary

### Files to Create
| File | Purpose |
|------|---------|
| `internal/testutil/fixtures.go` | Test helper utilities |
| `internal/schema/parser_test.go` | Schema parsing tests |
| `internal/schema/testdata/*.json` | Test fixtures |
| `internal/project/parser_test.go` | Project parsing tests |
| `internal/project/types_test.go` | Type tests |
| `internal/project/config_test.go` | Config parsing tests |
| `internal/generate/types_test.go` | Type generation tests |
| `internal/generate/tui_test.go` | TUI generation tests |
| `internal/build/manager_test.go` | Build manager tests |
| `internal/cmd/fresh_test.go` | Fresh command tests |
| `internal/cmd/checkit_test.go` | Checkit command tests |
| `internal/cmd/build_test.go` | Build command tests |
| `tests/integration/workflow_test.go` | Integration tests |

### Coverage Milestones
| Phase | After Step | Expected Coverage |
|-------|------------|-------------------|
| Phase 1 | Step 1.6 | 60%+ overall |
| Phase 2 | Step 2.4 | 70%+ overall |
| Phase 3 | Step 3.5 | 80%+ overall |
| Phase 4 | Step 4.1 | 85%+ overall |

---

## Verification Commands

```bash
# After each step
go test ./...

# Check coverage for specific package
go test -cover ./internal/schema/

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run only unit tests (skip integration)
go test ./...

# Run integration tests
go test -tags=integration ./...
```
