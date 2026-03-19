# Test Coverage Improvement - Design Document

## Overview

This document defines the testing architecture, conventions, and patterns for improving test coverage in the Ayo CLI project from 37.7% to 80-90%.

## Goals

- Achieve 80-90% coverage (ideal: 100% of critical paths)
- Establish consistent testing patterns across all packages
- Mix unit tests (logic) with integration tests (workflows)
- Tests in dependency order: schema → project → generate → build → cmd

---

## Package Analysis

### Coverage Targets

| Package | Current | Target | Lines | Priority |
|---------|---------|--------|-------|----------|
| `internal/schema` | 0% | 90%+ | 78 | 1 |
| `internal/project` | 0% | 90%+ | 211 | 2 |
| `internal/generate` | 46.4% | 90%+ | ~800 | 3 |
| `internal/build` | 0% | 80%+ | 260 | 4 |
| `internal/cmd` | 0% | 80%+ | ~200 | 5 |
| `internal/model` | 59.1% | 80%+ | ~200 | 6 |
| `internal/template` | 91.1% | 90%+ | ~100 | 7 |

---

## Testing Conventions

### File Organization

```
internal/
  schema/
    parser.go
    parser_test.go      # Unit tests
    testdata/           # Test fixtures
      valid_schema.json
      invalid_schema.json
  project/
    parser.go
    parser_test.go
    types_test.go
    config_test.go
    testdata/
      valid_project/
        config.toml
        system.md
```

### Test File Naming

- `<name>_test.go` - Unit tests for `<name>.go`
- `testdata/` - Test fixtures (JSON schemas, TOML configs, etc.)

### Test Function Naming

```go
func Test<FunctionName>_<Scenario>(t *testing.T) {}

// Examples:
func TestParseSchema_ValidInput(t *testing.T) {}
func TestParseSchema_InvalidJSON(t *testing.T) {}
func TestParseSchema_EmptyInput(t *testing.T) {}
func TestGenerateFlags_NilSchema(t *testing.T) {}
```

---

## Test Patterns

### Pattern 1: Table-Driven Tests

Use for functions with multiple scenarios:

```go
func TestParseSchema_Scenarios(t *testing.T) {
    tests := []struct {
        name    string
        input   []byte
        want    *ParsedSchema
        wantErr bool
    }{
        {
            name:  "valid object schema",
            input: []byte(`{"type": "object", "properties": {"name": {"type": "string"}}}`),
            want: &ParsedSchema{
                Type: "object",
                Properties: map[string]Property{
                    "name": {Type: "string"},
                },
            },
            wantErr: false,
        },
        {
            name:    "invalid JSON",
            input:   []byte(`{invalid}`),
            want:    nil,
            wantErr: true,
        },
        {
            name:    "empty input",
            input:   []byte{},
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseSchema(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseSchema() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseSchema() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Pattern 2: Code Generation Verification

For generated code, verify key structural elements:

```go
func TestGenerateAgent_ContainsExpectedFunctions(t *testing.T) {
    proj := &project.Project{
        Config: project.AgentConfig{Name: "test"},
    }

    code, err := GenerateAgent(proj, "main")
    if err != nil {
        t.Fatalf("GenerateAgent() error = %v", err)
    }

    // Verify expected functions exist
    expectedFuncs := []string{
        "func loadConfig()",
        "func saveConfig(",
        "func ensureConfig()",
        "func getAPIKeyEnv(",
    }

    for _, fn := range expectedFuncs {
        if !strings.Contains(code, fn) {
            t.Errorf("Generated code missing function: %s", fn)
        }
    }
}
```

### Pattern 3: Error Path Testing

Always test error conditions:

```go
func TestParseConfig_MissingFile(t *testing.T) {
    _, err := ParseConfig("/nonexistent/path/config.toml")
    if err == nil {
        t.Error("ParseConfig() expected error for missing file")
    }
}

func TestParseConfig_InvalidTOML(t *testing.T) {
    tmpFile := filepath.Join(t.TempDir(), "config.toml")
    os.WriteFile(tmpFile, []byte(`invalid [[toml`), 0644)

    _, err := ParseConfig(tmpFile)
    if err == nil {
        t.Error("ParseConfig() expected error for invalid TOML")
    }
}
```

### Pattern 4: Test Helpers

Create helpers for common setup:

```go
// testutil/helpers.go (or inline in test file)
func createTempProject(t *testing.T, files map[string]string) string {
    t.Helper()
    dir := t.TempDir()
    for name, content := range files {
        path := filepath.Join(dir, name)
        if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
            t.Fatalf("creating dir: %v", err)
        }
        if err := os.WriteFile(path, []byte(content), 0644); err != nil {
            t.Fatalf("writing file: %v", err)
        }
    }
    return dir
}

func parseTestSchema(t *testing.T, jsonStr string) *schema.ParsedSchema {
    t.Helper()
    s, err := schema.ParseSchema([]byte(jsonStr))
    if err != nil {
        t.Fatalf("parsing test schema: %v", err)
    }
    return s
}
```

### Pattern 5: Integration Tests

For end-to-end workflows, use build tags:

```go
//go:build integration
// +build integration

package cmd_test

import (
    "os/exec"
    "testing"
)

func TestFreshCheckitBuildWorkflow(t *testing.T) {
    // Create temp directory
    // Run ayo fresh test-agent
    // Run ayo checkit test-agent
    // Run ayo build test-agent
    // Verify binary exists and runs
}
```

---

## Package-Specific Test Plans

### 1. internal/schema

**Functions to test:**
- `ParseSchema(data []byte) (*ParsedSchema, error)`
- `GenerateFlags(schema *ParsedSchema) []Flag`

**Test cases:**
- Valid object schema
- Valid array schema
- Invalid JSON
- Empty input
- Schema with no properties
- Schema with nested objects
- Flag generation from schema
- Flag generation with nil schema

### 2. internal/project

**Functions to test:**
- `ParseProject(path string) (*Project, error)`
- `ValidateProject(p *Project) []ValidationError`
- `ParseConfig(path string) (*AgentConfig, error)`
- `Schema.UnmarshalJSON(data []byte) error`
- `ValidationError.Error() string`

**Test cases:**
- Valid minimal project (config + system.md)
- Valid full project (config + system + input schema + output schema + skills + hooks)
- Missing config.toml
- Missing system.md
- Invalid TOML config
- Invalid JSON schema
- Invalid hook type
- Validation errors formatting

### 3. internal/generate

**Functions needing tests:**
- `GenerateTypes(inputSchema, outputSchema, pkgName) (string, error)` - no test
- `GenerateTUI(proj, pkgName) (string, error)` - no test
- `toPascalCase(s string) string` - no test
- `jsonTypeToGo(jsonType string) string` - no test
- `toSafeIdentifier(s string) string` - no test
- `GenerateGoMod(proj) string` - no test

**Test cases:**
- Type generation for various JSON types
- Type generation with nil schemas
- Go.mod generation with various project names
- Safe identifier conversion

### 4. internal/build

**Functions to test:**
- `NewManager() *Manager`
- `Manager.Build(proj, outputPath) error`
- `Manager.generateFiles(proj) error`
- `Manager.copyAssets(proj) error`
- `Manager.compile(outputPath) error`
- `Manager.Cleanup()`
- `copyDirectory(src, dst) error`
- `findFantasyLibrary() (string, error)`
- `getSchema(s *project.Schema) *schema.ParsedSchema`

**Test cases:**
- Build with minimal project
- Build with skills and hooks
- Build output path resolution
- Cleanup removes temp directory
- Fantasy library discovery

**Note:** Build tests require integration setup (mock exec.Command or use build tags)

### 5. internal/cmd

**Functions to test:**
- `createProject(name string) error`
- `validateProject(path string) error`
- `buildProject(path string) error`

**Test cases:**
- Fresh command creates all files
- Fresh command fails on existing directory
- Checkit validates valid project
- Checkit reports validation errors
- Build command integration

---

## Test Utilities

### Create `internal/testutil` package

```go
// internal/testutil/fixtures.go
package testutil

import (
    "os"
    "path/filepath"
    "testing"
)

// CreateProject creates a minimal valid project structure for testing
func CreateProject(t *testing.T, name string) string {
    t.Helper()
    dir := t.TempDir()
    projectDir := filepath.Join(dir, name)
    os.MkdirAll(projectDir, 0755)

    config := `[agent]
name = "` + name + `"
version = "1.0.0"
description = "Test agent"`

    os.WriteFile(filepath.Join(projectDir, "config.toml"), []byte(config), 0644)
    os.WriteFile(filepath.Join(projectDir, "system.md"), []byte("# System"), 0644)

    return projectDir
}

// ValidInputSchema returns a valid input JSON schema for testing
func ValidInputSchema() string {
    return `{
        "type": "object",
        "properties": {
            "query": {"type": "string", "description": "Search query"}
        },
        "required": ["query"]
    }`
}

// ValidOutputSchema returns a valid output JSON schema for testing
func ValidOutputSchema() string {
    return `{
        "type": "object",
        "properties": {
            "result": {"type": "string"}
        }
    }`
}
```

---

## Implementation Order

### Phase 1: Foundation (schema + project)
1. `internal/schema/parser_test.go`
2. `internal/project/parser_test.go`
3. `internal/project/types_test.go`
4. `internal/project/config_test.go`

### Phase 2: Code Generation (generate)
5. `internal/generate/types_test.go`
6. `internal/generate/embed_test.go` (add more cases)
7. `internal/generate/hooks_test.go` (add more cases)

### Phase 3: Build & Commands
8. `internal/build/manager_test.go`
9. `internal/cmd/fresh_test.go`
10. `internal/cmd/checkit_test.go`
11. `internal/cmd/build_test.go`

### Phase 4: Integration
12. Integration tests for `fresh → checkit → build` workflow

---

## Acceptance Criteria

Each package must meet:

- [ ] 80%+ line coverage
- [ ] All exported functions have tests
- [ ] All error paths tested
- [ ] Edge cases documented and tested
- [ ] Table-driven tests for multi-scenario functions
- [ ] Test helpers created for common patterns
- [ ] No test failures in `go test ./...`

---

## Commands

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/schema/...

# Run integration tests
go test -tags=integration ./...
```
