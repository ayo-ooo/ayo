---
id: ase-j5tm
status: closed
deps: [ase-lbg7]
links: []
created: 2026-02-09T03:12:06Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Unit tests for flow parser and validator

## Background

The flow system uses YAML files to define orchestration patterns. The parser reads YAML and produces a structured Flow object. The validator ensures the flow definition is correct and references valid agents/steps.

## Why This Matters

Flow files are user-authored and can contain:
- Syntax errors (invalid YAML)
- Semantic errors (referencing non-existent steps)
- Type errors (wrong parameter types)
- Circular dependencies between steps

Unit tests ensure we catch all these issues with clear error messages.

## Implementation Details

### Test Structure

```
internal/
  flows/
    parser.go
    parser_test.go
    validator.go
    validator_test.go
    testdata/
      valid/
        simple-flow.yaml
        multi-step.yaml
        with-triggers.yaml
        complex-templating.yaml
      invalid/
        bad-yaml-syntax.yaml
        missing-agent.yaml
        circular-deps.yaml
        invalid-cron.yaml
        duplicate-step-ids.yaml
        missing-required-fields.yaml
```

### Parser Test Cases

**parser_test.go:**

```go
func TestParser_ValidFlow(t *testing.T) {
    // Test each valid testdata file parses correctly
}

func TestParser_InvalidYAML(t *testing.T) {
    // Malformed YAML returns clear error
}

func TestParser_MissingRequiredFields(t *testing.T) {
    // name, version are required
}

func TestParser_StepTypes(t *testing.T) {
    // shell and agent step types parse correctly
}

func TestParser_Triggers(t *testing.T) {
    // cron and watch triggers parse correctly
}

func TestParser_Templates(t *testing.T) {
    // {{ steps.foo.stdout }} template syntax preserved
}
```

### Validator Test Cases

**validator_test.go:**

```go
func TestValidator_ValidFlow(t *testing.T) {
    // Valid flows pass validation
}

func TestValidator_CircularDependency(t *testing.T) {
    // step A depends on B, B depends on A - detected
}

func TestValidator_InvalidStepReference(t *testing.T) {
    // {{ steps.nonexistent.stdout }} - error
}

func TestValidator_InvalidAgentReference(t *testing.T) {
    // agent: "@nonexistent" - error (if strict mode)
}

func TestValidator_InvalidCronExpression(t *testing.T) {
    // schedule: "invalid cron" - error
}

func TestValidator_DuplicateStepIds(t *testing.T) {
    // Two steps with same id - error
}

func TestValidator_DuplicateTriggerIds(t *testing.T) {
    // Two triggers with same id - error
}

func TestValidator_WatchPathNotAbsolute(t *testing.T) {
    // Watch triggers need absolute paths
}

func TestValidator_ShellStepMissingRun(t *testing.T) {
    // Shell step without 'run' field
}

func TestValidator_AgentStepMissingAgent(t *testing.T) {
    // Agent step without 'agent' field
}
```

### Test Fixtures

**testdata/valid/multi-step.yaml:**
```yaml
version: 1
name: multi-step-example
steps:
  - id: gather
    type: shell
    run: find . -name "*.go"
  - id: analyze
    type: agent
    agent: "@code-analyzer"
    prompt: "Analyze these files"
    input: "{{ steps.gather.stdout }}"
  - id: report
    type: shell
    run: echo "Analysis complete"
    depends_on: [analyze]
```

**testdata/invalid/circular-deps.yaml:**
```yaml
version: 1
name: circular
steps:
  - id: a
    type: shell
    run: echo a
    depends_on: [b]
  - id: b
    type: shell
    run: echo b
    depends_on: [a]
```

## Acceptance Criteria

- [ ] Parser tests for valid YAML files
- [ ] Parser tests for malformed YAML
- [ ] Parser tests for all step types
- [ ] Parser tests for all trigger types
- [ ] Validator tests for circular dependencies
- [ ] Validator tests for invalid references
- [ ] Validator tests for duplicate IDs
- [ ] Validator tests for invalid cron expressions
- [ ] Test fixtures in testdata/ directory
- [ ] Clear error messages for all failure cases
- [ ] All tests pass with `go test ./internal/flows/...`

