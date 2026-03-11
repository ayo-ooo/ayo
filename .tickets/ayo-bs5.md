---
id: ayo-bs5
status: open
deps: [ayo-bs2]
links: []
created: 2026-03-11T18:00:00Z
type: epic
priority: 5
assignee: Alex Cabrera
tags: [build-system, json-schema, cli-generation]
---
# Phase 5: Schema to CLI Conversion

Convert input.jsonschema to command-line arguments. This is a critical feature for user-friendly agent interactions.

## Context

input.jsonschema defines the structure of agent input. We must:

1. Parse JSON Schema format
2. Generate CLI flags from schema properties
3. Support all JSON Schema types and features
4. Generate help text from schema descriptions

## Tasks

### 5.1 JSON Schema Parser
- [ ] Parse full JSON Schema draft 7 subset
- [ ] Support type system (string, number, integer, boolean, array, object)
- [ ] Handle nested schemas
- [ ] Support $refs and schema composition
- [ ] Validate schema structure

### 5.2 CLI Flag Generation
- [ ] Generate long flags (--name)
- [ ] Generate short flags (-n) for common fields
- [ ] Handle required fields
- [ ] Apply default values
- [ ] Support multiple values (arrays)
- [ ] Handle flag types (string, int, bool, float)

### 5.3 Schema Features Support
- [ ] Enum constraints
- [ ] Format patterns (email, uri, etc.)
- [ ] Minimum/maximum for numbers
- [ ] MinLength/maxLength for strings
- [ ] Required fields
- [ ] Description for help text

### 5.4 Help Text Generation
- [ ] Generate usage string
- [ ] Generate flag descriptions
- [ ] Show default values
- [ ] Indicate required fields
- [ ] Format for terminal output

### 5.5 CLI Integration
- [ ] Integrate with cobra flag parsing
- [ ] Validate parsed values against schema
- [ ] Provide helpful error messages
- [ ] Support --help flag
- [ ] Support JSON input via stdin (alternative to flags)

## Technical Details

### Input Schema Example

```json
{
  "type": "object",
  "properties": {
    "file": {
      "type": "string",
      "description": "File to process"
    },
    "task": {
      "type": "string",
      "description": "Task to perform",
      "enum": ["review", "fix", "test"]
    },
    "verbose": {
      "type": "boolean",
      "description": "Verbose output",
      "default": false
    }
  },
  "required": ["file", "task"]
}
```

### Generated CLI Interface

```bash
Usage: ./my-agent [options]

Options:
  --file string        File to process (required)
  --task string        Task to perform: review|fix|test (required)
  --verbose            Verbose output (default: false)
  --help, -h           Show help

Alternative: echo '{"file": "...", "task": "..."}' | ./my-agent
```

### Implementation Strategy

1. Load input.jsonschema at build time
2. Generate CLI flag definitions
3. Embed flag definitions in executable
4. At runtime, parse flags using cobra
5. Validate values against schema
6. Also accept JSON via stdin as fallback

## Deliverables

- [ ] Full JSON Schema parser
- [ ] CLI flag generator
- [ ] Help text generator
- [ ] Schema validation at runtime
- [ ] JSON stdin support
- [ ] Test coverage > 85%
- [ ] Documentation on schema format

## Acceptance Criteria

1. All JSON Schema types convert to CLI flags
2. Required fields are enforced
3. Default values work correctly
4. Help text is clear and accurate
5. JSON stdin works as alternative
6. Validation errors are helpful

## Dependencies

- **ayo-bs2**: Build system core (to embed schema)

## Out of Scope

- Interactive prompts for missing required fields
- Schema composition beyond basic $refs
- Full JSON Schema draft 2020-12 (subset only)

## Risks

- **Complexity**: Full JSON Schema is very complex
  - **Mitigation**: Implement commonly used subset first, expand incrementally

## Notes

Use existing JSON Schema validation libraries where possible. Focus on features agents actually need.
