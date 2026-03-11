---
id: ayo-bs6
status: open
deps: [ayo-bs2, ayo-bs3, ayo-bs4, ayo-bs5, ayo-bs7]
links: []
created: 2026-03-11T18:00:00Z
type: epic
priority: 6
assignee: Alex Cabrera
tags: [build-system, runtime, agent-execution]
---
# Phase 6: Runtime Execution Framework

Implement standalone agent execution in generated binaries. This is the runtime that makes built agents actually work.

## Context

Generated executables need to:

1. Parse embedded config
2. Parse CLI arguments (from Phase 5)
3. Accept JSON via stdin (alternative input)
4. Initialize Fantasy agent with system prompt + skills
5. Load and register tools (from Phase 4)
6. Execute agent and get response
7. Validate output against output.jsonschema
8. Handle errors gracefully

## Tasks

### 6.1 Config Parsing
- [ ] Parse embedded config.toml
- [ ] Load embedded system prompt
- [ ] Load embedded output.jsonschema (if present)
- [ ] Validate config structure
- [ ] Handle missing/invalid config

### 6.2 CLI Argument Parsing
- [ ] Parse CLI flags using embedded definitions
- [ ] Validate against input.jsonschema
- [ ] Handle required fields
- [ ] Apply default values
- [ ] Show help when requested

### 6.3 JSON Stdin Support
- [ ] Detect JSON input on stdin
- [ ] Parse JSON input
- [ ] Validate against input.jsonschema
- [ ] Merge with CLI args (if both provided)
- [ ] Handle parse errors

### 6.4 Fantasy Agent Initialization
- [ ] Initialize Fantasy with embedded config
- [ ] Build system prompt: base + skills
- [ ] Inject skills prompt into system prompt
- [ ] Configure model settings
- [ ] Set up memory if enabled

### 6.5 Tools Loading
- [ ] Discover embedded tools
- [ ] Register tools with Fantasy
- [ ] Load tool definitions
- [ ] Handle tool conflicts
- [ ] Set up tool execution hooks

### 6.6 Agent Execution
- [ ] Execute agent with parsed input
- [ ] Stream responses to user
- [ ] Handle tool calls
- [ ] Manage conversation context
- [ ] Support multi-turn conversations

### 6.7 Output Validation
- [ ] Parse output.jsonschema (if present)
- [ ] Validate LLM responses against schema
- [ ] Retry on validation failure (with limit)
- [ ] Provide helpful validation errors
- [ ] Return valid output to user

### 6.8 Error Handling
- [ ] Handle config errors
- [ ] Handle input validation errors
- [ ] Handle LLM API errors
- [ ] Handle tool execution errors
- [ ] Provide clear error messages
- [ ] Support exit codes

## Technical Details

### Runtime Flow

```
1. Load embedded resources (config, prompts, schemas)
2. Parse CLI args OR read JSON stdin
3. Validate input against schema
4. Initialize Fantasy agent
5. Inject skills into system prompt
6. Register tools
7. Execute agent with input
8. Validate output against schema
9. Print output
10. Exit with appropriate code
```

### Input Methods

**CLI Args** (from Phase 5):
```bash
./my-agent --file main.go --task review
```

**JSON Stdin**:
```bash
echo '{"file": "main.go", "task": "review"}' | ./my-agent
```

**Freeform** (no schema):
```bash
./my-agent "Review this code"
```

### Output Validation

If output.jsonschema exists:
```go
output := agent.Execute(input)
if err := schema.Validate(output, outputSchema); err != nil {
    // Retry or return error
}
```

## Deliverables

- [ ] Complete runtime execution framework
- [ ] Config parsing from embedded resources
- [ ] CLI argument parsing with validation
- [ ] JSON stdin support
- [ ] Fantasy agent initialization
- [ ] Tools loading and registration
- [ ] Output schema validation
- [ ] Comprehensive error handling
- [ ] Test coverage > 85%
- [ ] Runtime documentation

## Acceptance Criteria

1. Built executable runs without ayo
2. CLI args work with validation
3. JSON stdin works with validation
4. Skills inject into system prompt
5. Tools execute correctly
6. Output validates against schema
7. Errors are clear and helpful
8. Exit codes indicate success/failure

## Dependencies

- **ayo-bs2**: Build system core (embedded resources)
- **ayo-bs3**: Skills system (prompt injection)
- **ayo-bs4**: Tools system (tool loading)
- **ayo-bs5**: Schema to CLI (flag parsing)

## Out of Scope

- Interactive mode (future enhancement)
- Streaming responses (can add later)
- Multi-agent coordination (removed in Phase 1)

## Risks

- **Complexity**: Runtime must handle many edge cases
  - **Mitigation**: Start with happy path, add error handling incrementally
- **Performance**: Schema validation may be slow
  - **Mitigation**: Cache compiled schemas, use efficient validation

## Notes

This is the final piece that makes everything work together. Test extensively.
