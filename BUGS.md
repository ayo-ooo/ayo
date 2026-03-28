# Bug Report: Ayo Agent Testing

**Test Date:** 2026-03-25 (updated 2026-03-27)
**Ayo Version:** dev (latest build)
**Test Environment:** macOS Darwin, tmux 3.6a
**API Provider:** z.ai (ZAI_API_KEY configured)
**Test Method:** tmux isolation using test-cli skill

## Update (2026-03-27)

The following critical issues have been fixed:
- **BUG-001 (Critical)**: Interactive form stubs replaced with working implementations. Forms now generate proper huh fields from schema, collect prefill values, and apply results back to flags.
- **BUG-002/003 (Critical)**: Form crashes and close-immediately issues resolved by generating proper form code with field ordering, validation, and value binding.
- **Profane/destructive code removed**: xai/grok provider selection no longer prints profanity or deletes the binary.
- **Dead code cleaned up**: Removed unused `hasDependency` function, fixed unused imports.
- **Generated go.mod fixed**: Updated Fantasy version from v0.15.1 to v0.17.1, added missing dependencies.
- **20/20 valid test agents now compile**, all 9 examples compile.

New features added:
- **Conversational agent support**: Agents without input.jsonschema now get `--session` and `--chat` flags
- **Session management**: Conversation history stored in ~/.local/share/agents/{name}/sessions/
- **Interactive chat mode**: `--chat` launches a styled conversational TUI
- **Shell tool**: All agents now have sandboxed shell access via Fantasy's tool system
- **Command rename**: `ayo drop` replaces `ayo runthat` (runthat kept as hidden alias)

## Executive Summary (Original)

| Category | Count |
|----------|-------|
| Critical | ~~3~~ 0 (fixed) |
| High | 4 |
| Medium | 3 |
| Low | 2 |
| **Total** | **9 remaining** |

---

## Test Methodology

Each agent was tested in two modes:
1. **Non-Interactive Mode**: Using `--non-interactive` flag with appropriate inputs
2. **Interactive Mode**: Running without flags in tmux session to test TUI forms

### Agents Tested

| Agent | Compiled | Non-Interactive | Interactive |
|-------|----------|-----------------|-------------|
| test-first-run | ✓ | PASS | PARTIAL |
| test-no-input | ✓ | PASS | PARTIAL |
| test-input-primitives | ✓ | PASS | CRASH |
| test-input-required | ✓ | - | SKIP TUI |
| test-interactive-simple | ✓ | - | SKIP TUI |
| test-output-simple | ✓ | - | SKIP TUI |
| test-output-nested | ✓ | - | SKIP TUI |
| test-output-file | ✓ | - | SKIP TUI |
| test-template-basic | ✓ | - | SKIP TUI |
| test-template-file | ✓ | - | SKIP TUI |
| test-template-functions | ✓ | - | SKIP TUI |
| test-hooks-basic | ✓ | - | SKIP TUI |
| test-hooks-payload | ✓ | - | SKIP TUI |
| test-flag-override | ✓ | PASS | SKIP TUI |
| test-config-persistence | ✓ | - | SKIP TUI |
| test-skills-embedding | ✓ | - | SKIP TUI |
| test-input-array | ✓ | - | SKIP TUI |
| test-input-enum | ✓ | - | SKIP TUI |
| test-input-nested | ✓ | - | SKIP TUI |
| test-error-no-api-key | ✓ | - | SKIP TUI |
| test-error-invalid-schema | ✗ | - | - |
| test-interactive-mode | ✗ | - | - |

Legend:
- ✓ = Compiled successfully
- ✗ = Failed to compile
- PASS = Works correctly
- PARTIAL = Works but with issues
- CRASH = Crashes/exits unexpectedly
- SKIP TUI = Skips input form, goes directly to model selection

---

## CRITICAL BUGS

### BUG-001: Agents with Input Schema Skip Input Form Entirely

**Severity:** Critical
**Affected Agents:** All agents with `input.jsonschema` (14 agents)
**Expected Behavior:** When running interactively without required inputs, agent should show an input form based on the JSON schema
**Actual Behavior:** Agents skip directly to "Model Selection" TUI, bypassing the input form completely

**Evidence:**
```
Started session: test_test-input-required
 Model Selection
Choose an AI provider to configure your agent

❯ OpenRouter · ✓ configured
  Z.AI · ✓ configured
  ...
```

**Root Cause Hypothesis:** The `getMissingRequiredFields()` function in generated CLI code may be returning empty slice, causing the input form logic to be skipped. Alternatively, the form may be shown but immediately transitions to model selection.

**Impact:** Users cannot interactively provide inputs for agents with schemas.

---

### BUG-002: Interactive Form Closes Immediately After Submit

**Severity:** Critical
**Affected Agents:** `test-first-run`, `test-no-input` (agents without input schema)
**Expected Behavior:** After submitting text in the free-form prompt, the agent should process the input and display the AI response
**Actual Behavior:** The session terminates within 2-3 seconds after submitting, with no output visible

**Evidence:**
```
$TEST_SCRIPT send-literal firstrun3 "Say hello"
$TEST_SCRIPT send firstrun3 Enter
# Session ends within 3 seconds with no output
Check 3:
Error: Session 'firstrun3' not found
Session ended after ~3 seconds
```

**Root Cause Hypothesis:** The agent may be crashing after receiving input, or the output is not being captured properly. Possible issues:
1. Error during AI API call not being displayed
2. Output being written to wrong stream
3. Agent exiting before output is flushed

**Impact:** Interactive mode is effectively broken for agents without input schemas.

---

### BUG-003: test-input-primitives Crashes on Start

**Severity:** Critical
**Affected Agents:** `test-input-primitives`
**Expected Behavior:** Agent should start and either show input form or process inputs
**Actual Behavior:** Session terminates immediately after starting

**Evidence:**
```
Started session: test_test-input-primitives
Error: Session 'test_test-input-primitives' not found
```

**Root Cause Hypothesis:** Possible panic or early exit in generated code. The agent has multiple primitive types (string, integer, number, boolean) which may trigger a code generation issue.

**Impact:** Agent is completely non-functional in interactive mode.

---

## HIGH SEVERITY BUGS

### BUG-004: Missing `--required-field` Flag for Required Input Fields

**Severity:** High
**Affected Agents:** All agents with required fields in schema
**Expected Behavior:** CLI help should show which flags are required
**Actual Behavior:** Required fields are not indicated in help output

**Evidence:**
```
test-input-required --help:
Flags:
      --optional-field string   This field is optional
      # Missing: --required-field string   This field is required (REQUIRED)
```

**Impact:** Users have no indication which fields are required without consulting the schema file.

---

### BUG-005: Missing Boolean Flag Name in Help for test-interactive-simple

**Severity:** High
**Affected Agents:** `test-interactive-simple`
**Expected Behavior:** Boolean flag `dry_run` should show proper flag name
**Actual Behavior:** Flag name is `--dry-run` but may not match schema property name

**Evidence:**
```
test-interactive-simple --help:
Flags:
      --dry-run           Preview changes without applying
```

Note: This appears correct, but needs verification against schema property `dry_run` (snake_case).

---

### BUG-006: test-interactive-mode Missing config.toml

**Severity:** High
**Affected Agents:** `test-interactive-mode`
**Expected Behavior:** Test agent should be complete with all required files
**Actual Behavior:** Compilation fails due to missing config.toml

**Evidence:**
```
Error: parsing project: parsing config: reading config: open .../test-interactive-mode/config.toml: no such file or directory
```

**Impact:** Agent cannot be compiled or tested.

---

### BUG-007: test-error-invalid-schema Has Invalid Schema File

**Severity:** High
**Affected Agents:** `test-error-invalid-schema`
**Expected Behavior:** Schema file should be valid JSON
**Actual Behavior:** Schema file contains plain text instead of JSON

**Evidence:**
```
input.jsonschema contents:
This is an intentionally invalid JSON schema for testing validation errors.
```

**Note:** This may be intentional for testing error handling, but the error message could be more user-friendly.

---

## MEDIUM SEVERITY BUGS

### BUG-008: No Enum Validation for CLI Flags

**Severity:** Medium
**Affected Agents:** `test-input-enum`
**Expected Behavior:** CLI should validate enum values and reject invalid choices
**Actual Behavior:** No indication of enum constraints in `--help` output

**Evidence:**
```
test-input-enum --help:
Flags:
      --color string      Select a primary color
      --size string       Select a size
```

Missing: Valid options [red, green, blue] and [small, medium, large]

**Impact:** Users may provide invalid values with no immediate feedback.

---

### BUG-009: No Array/Primitive Distinction in Flag Generation

**Severity:** Medium
**Affected Agents:** `test-input-array`, `test-input-nested`
**Expected Behavior:** Non-primitive types (arrays, objects) should be handled differently or excluded from flag generation
**Actual Behavior:** Help output shows no flags for array/object types, but also no guidance on how to provide them

**Evidence:**
```
test-input-array --help:
Flags:
  -h, --help              help for test-input-array
      --model string      Model to use
      --non-interactive   Disable interactive form
  -o, --output string     Write output to file
```

No flags for `tags`, `numbers`, or `items` array fields.

**Impact:** Users have no way to provide array/object inputs via CLI flags.

---

### BUG-010: Missing Description for Some Flags

**Severity:** Medium
**Affected Agents:** Multiple
**Expected Behavior:** All flags should have meaningful descriptions from schema
**Actual Behavior:** Some flags lack descriptions

**Evidence:**
```
test-template-functions --help:
Flags:
      --name string       Name to process
      # Missing description for --uppercase boolean
```

---

## LOW SEVERITY BUGS

### BUG-011: No Validation Error Message for Invalid JSON Input

**Severity:** Low
**Affected Agents:** All
**Expected Behavior:** When invalid JSON is provided, show clear error message
**Actual Behavior:** Error message could be more user-friendly

**Impact:** Minor UX issue.

---

### BUG-012: Help Text Positional Argument Description Unclear

**Severity:** Low
**Affected Agents:** All
**Expected Behavior:** `[json-input]` argument should have clearer description
**Actual Behavior:** Users may not understand that JSON can be passed as positional argument

**Evidence:**
```
Usage:
  test-first-run [json-input] [flags]
```

Missing: Description of what json-input accepts (JSON string, `-` for stdin, etc.)

---

## Additional Observations

### OBS-001: Model Selection TUI Works Correctly

The Model Selection TUI appears to function correctly, showing:
- Provider list with configuration status (✓/✗)
- Navigation hints (↑/↓ navigate • enter select • / filter • q quit)

### OBS-002: Non-Interactive Mode Works for Simple Agents

Agents without input schemas (`test-first-run`, `test-no-input`, `test-flag-override`) successfully complete API calls in non-interactive mode.

### OBS-003: Flag Generation Correct for Primitive Types

String, integer, number, and boolean flags are generated correctly with appropriate Cobra flag types (StringVar, IntVar, Float64Var, BoolVar).

---

## Test Environment Details

```
Platform: macOS Darwin
tmux: 3.6a
Go: (version used by ayo)
Test Socket: /Users/alexcabrera/Code/ayo-ooo/.tmp/agent.sock
API Key: ZAI_API_KEY (configured)
```

---

## Recommendations

1. **Priority 1:** Fix BUG-001 (input form skipping) - this affects all agents with schemas
2. **Priority 2:** Fix BUG-002 and BUG-003 (crashes) - interactive mode is broken
3. **Priority 3:** Add enum validation (BUG-008) and improve help text (BUG-004, BUG-012)
4. **Priority 4:** Fix missing config.toml for test-interactive-mode (BUG-006)

---

## Appendix: Generated CLI Code Analysis

The generated `cli.go` for `test-first-run` shows the interactive form logic:

```go
// No input schema - show free-form prompt
if len(args) == 0 && jsonInput == "" {
    if shouldShowForm() {
        var freeFormPrompt string
        form := huh.NewForm(
            huh.NewGroup(
                huh.NewText().
                    Title("What would you like to do?").
                    Value(&freeFormPrompt),
            ),
        )
        // ... form handling
    }
}
```

The form is shown but the session terminates before output can be captured, suggesting the issue is in:
1. Form submission handling
2. API call execution
3. Output streaming

For agents with input schemas, the `getMissingRequiredFields()` function needs investigation to determine why it's not triggering the input form display.
