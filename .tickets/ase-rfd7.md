---
id: ase-rfd7
status: closed
deps: [ase-3iuw]
links: []
created: 2026-02-09T03:12:29Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Unit tests for template resolution

## Background

Flow steps can reference outputs from previous steps using template syntax like `{{ steps.gather.stdout }}`. The template resolver substitutes these references with actual values at runtime.

## Why This Matters

Template resolution is critical for flow execution:
- Incorrect substitution = wrong data passed to agents
- Missing error handling = confusing failures
- Edge cases in template syntax = security vulnerabilities

## Implementation Details

### Test Structure

```
internal/
  flows/
    template/
      resolver.go
      resolver_test.go
      funcs.go
      funcs_test.go
```

### Template Syntax

The template system supports:

```yaml
# Basic step output reference
input: "{{ steps.gather.stdout }}"
input: "{{ steps.gather.stderr }}"
input: "{{ steps.gather.exit_code }}"

# JSON path for structured output
input: "{{ steps.gather.json.results[0].name }}"

# Built-in functions
input: "{{ env \"HOME\" }}"
input: "{{ now | date \"2006-01-02\" }}"
input: "{{ steps.gather.stdout | trim }}"
input: "{{ steps.gather.stdout | lines | first }}"

# Conditionals
prompt: |
  {{ if steps.gather.exit_code == 0 }}
  Success: {{ steps.gather.stdout }}
  {{ else }}
  Error: {{ steps.gather.stderr }}
  {{ end }}
```

### Resolver Test Cases

**resolver_test.go:**

```go
func TestResolver_StepStdout(t *testing.T) {
    ctx := &ExecutionContext{
        Steps: map[string]*StepResult{
            "gather": {Stdout: "file1.go\nfile2.go"},
        },
    }
    result, err := Resolve("{{ steps.gather.stdout }}", ctx)
    assert.NoError(t, err)
    assert.Equal(t, "file1.go\nfile2.go", result)
}

func TestResolver_StepStderr(t *testing.T) {
    // Similar for stderr
}

func TestResolver_StepExitCode(t *testing.T) {
    // Integer exit code
}

func TestResolver_StepNotFound(t *testing.T) {
    // Error with helpful message
}

func TestResolver_NestedJSON(t *testing.T) {
    ctx := &ExecutionContext{
        Steps: map[string]*StepResult{
            "api": {Stdout: `{"users":[{"name":"alice"}]}`},
        },
    }
    result, err := Resolve("{{ steps.api.json.users[0].name }}", ctx)
    assert.Equal(t, "alice", result)
}

func TestResolver_InvalidJSON(t *testing.T) {
    // Graceful error when stdout isn't valid JSON
}

func TestResolver_MultipleReferences(t *testing.T) {
    // Template with multiple step references
}

func TestResolver_EnvFunction(t *testing.T) {
    os.Setenv("TEST_VAR", "test_value")
    result, err := Resolve("{{ env \"TEST_VAR\" }}", nil)
    assert.Equal(t, "test_value", result)
}

func TestResolver_PipelineFunctions(t *testing.T) {
    // Test | trim, | lines, | first, etc.
}

func TestResolver_Conditional(t *testing.T) {
    // if/else logic
}

func TestResolver_InvalidSyntax(t *testing.T) {
    // {{ unterminated - clear error
}

func TestResolver_SecurityEscaping(t *testing.T) {
    // Ensure template injection doesn't evaluate arbitrary code
    ctx := &ExecutionContext{
        Steps: map[string]*StepResult{
            "user_input": {Stdout: "{{ env \"SECRET\" }}"},
        },
    }
    result, err := Resolve("User said: {{ steps.user_input.stdout }}", ctx)
    // Should be literal "{{ env \"SECRET\" }}", not evaluated
    assert.Contains(t, result, "{{ env")
}
```

### Template Functions Test Cases

**funcs_test.go:**

```go
func TestTrim(t *testing.T) {
    assert.Equal(t, "hello", trim("  hello  "))
}

func TestLines(t *testing.T) {
    assert.Equal(t, []string{"a", "b", "c"}, lines("a\nb\nc"))
}

func TestFirst(t *testing.T) {
    assert.Equal(t, "a", first([]string{"a", "b", "c"}))
}

func TestLast(t *testing.T) {
    assert.Equal(t, "c", last([]string{"a", "b", "c"}))
}

func TestJoin(t *testing.T) {
    assert.Equal(t, "a,b,c", join([]string{"a", "b", "c"}, ","))
}

func TestBase64Encode(t *testing.T) {
    assert.Equal(t, "aGVsbG8=", base64Encode("hello"))
}

func TestBase64Decode(t *testing.T) {
    result, err := base64Decode("aGVsbG8=")
    assert.NoError(t, err)
    assert.Equal(t, "hello", result)
}
```

## Acceptance Criteria

- [ ] Basic step reference tests (stdout, stderr, exit_code)
- [ ] JSON path extraction tests
- [ ] Environment variable function tests
- [ ] Pipeline function tests (trim, lines, first, last, join)
- [ ] Conditional logic tests
- [ ] Error handling for missing steps
- [ ] Error handling for invalid JSON
- [ ] Error handling for invalid syntax
- [ ] Security test for template injection
- [ ] All functions in funcs.go have corresponding tests
- [ ] All tests pass with `go test ./internal/flows/template/...`

