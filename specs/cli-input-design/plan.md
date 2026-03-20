# Implementation Plan

## Progress
- [ ] Step 1: Add JSON payload parsing support
- [ ] Step 2: Update schema parser for new properties
- [ ] Step 3: Update CLI generator (remove position, add JSON input)
- [ ] Step 4: Update existing examples
- [ ] Step 5: Update documentation

---

## Step 1: Add JSON Payload Parsing Support

**Objective:** Support JSON input as first positional argument or stdin

**Implementation:**
- Add `parseJSONInput()` function in CLI generator
- Detect JSON input: starts with `{` or `-` for stdin
- Parse and validate against schema
- Merge with flag values (flags override JSON)

**Tests Required:**
- Valid JSON payload parsed correctly
- Invalid JSON returns clear error
- Stdin (`-`) reads from os.Stdin
- Flag overrides JSON values
- Schema validation catches mismatches

**Integration:**
- Generated `cli.go` includes JSON parsing logic
- Works alongside existing flag parsing

**Demo:**
```bash
./agent '{"text": "hello"}'
echo '{"text": "hello"}' | ./agent -
```

---

## Step 2: Update Schema Parser for New Properties

**Objective:** Support `flag` and `file` properties, deprecate `x-cli-*`

**Implementation:**
- Add `Flag` and `File` fields to `Property` struct
- Keep `x-cli-*` fields for backwards compatibility
- Add deprecation warning when `x-cli-*` detected

**Tests Required:**
- New `flag` property maps to flag name
- New `file` property marks file loading
- Old `x-cli-*` still works with warning
- Mix of old and new properties handled

**Integration:**
- `internal/schema/parser.go` updated
- Generator uses new properties

**Demo:**
```bash
ayo build ./examples/translate  # No warnings
ayo build ./old-agent           # Warning about x-cli-*
```

---

## Step 3: Update CLI Generator

**Objective:** Remove positional arguments, add JSON input flag

**Implementation:**
- Remove `x-cli-position` handling
- Generate flags for all top-level primitives
- No short flags generated
- Add `[json-input]` to usage string
- Add stdin support (`-`)

**Tests Required:**
- No positional args in generated code
- All primitive properties get flags
- Complex types (object, array) don't get flags
- Usage shows `[json-input]` placeholder

**Integration:**
- `internal/generate/cli.go` refactored
- Generated binaries accept JSON

**Demo:**
```
Usage:
  translate [json-input] [flags]

Flags:
      --from string
      --to string
      --text string
```

---

## Step 4: Update Existing Examples

**Objective:** Migrate all examples to new schema format

**Implementation:**
- Remove `x-cli-position` from all `input.jsonschema` files
- Remove `x-cli-short` from all examples
- Keep minimal schemas with just types and defaults

**Files to Update:**
- `examples/translate/input.jsonschema`
- `examples/code-review/input.jsonschema`
- `examples/task-runner/input.jsonschema`
- `examples/summarize/input.jsonschema`
- `examples/notifier/input.jsonschema`
- `examples/research/input.jsonschema`
- `examples/data-pipeline/input.jsonschema`

**Tests Required:**
- All examples build successfully
- Generated CLIs work with JSON input
- Generated CLIs work with flags

**Demo:**
```bash
ayo build ./examples/translate
./translate '{"text": "hello", "to": "spanish"}'
```

---

## Step 5: Update Documentation

**Objective:** Document new input schema approach

**Implementation:**
- Update `docs/reference/input-schema.md`
- Update `docs/getting-started/quickstart.md`
- Update `docs/guides/building-agents.md`
- Update `README.md` examples

**Content:**
- Remove `x-cli-position`, `x-cli-short` references
- Add JSON payload examples
- Add stdin examples
- Document flag override behavior

**Tests Required:**
- Docs build without errors
- Examples in docs work as written

**Demo:**
User can follow quickstart and it works with new approach
