---
id: ayo-bs4
status: open
deps: [ayo-bs2]
links: []
created: 2026-03-11T18:00:00Z
type: epic
priority: 4
assignee: Alex Cabrera
tags: [build-system, tools, executable-tools]
---
# Phase 4: Tools System Implementation

Comprehensive tools support with easy loading. Tools are simply executable binaries in the tools/ directory.

## Context

Tools in the build system are:

1. **Built-in Tools**: bash, file_read, file_write, git, web_search
2. **Custom Tools**: Any executable in tools/ directory
3. **No Go Code Required**: Just make a file executable

The tools system needs to:
- Discover tools from tools/ directory
- Generate descriptions for the LLM
- Execute tools as subprocesses
- Handle tool availability

## Tasks

### 4.1 Tool Discovery
- [ ] Scan tools/ for executables
- [ ] Detect file types (scripts, binaries)
- [ ] Check executability (chmod +x, file headers)
- [ ] Ignore hidden files and templates
- [ ] Support symbolic links

### 4.2 Tool Description Generation
- [ ] Parse shebang lines for scripts
- [ ] Extract help text if available
- [ ] Generate descriptions for custom tools
- [ ] Map tools to Fantasy tool format
- [ ] Handle tool conflicts (built-in vs custom)

### 4.3 Tool Execution Framework
- [ ] Implement subprocess execution
- [ ] Handle stdin/stdout/stderr
- [ ] Set appropriate environment variables
- [ ] Timeout handling
- [ ] Error propagation

### 4.4 Built-in Tools Implementation
- [ ] bash: Execute shell commands
- [ ] file_read: Read file contents
- [ ] file_write: Write to files
- [ ] git: Git operations (clone, checkout, status)
- [ ] web_search: Web search (optional, requires API key)

### 4.5 Tool Availability Checking
- [ ] Check tool dependencies
- [ ] Warn about missing tools
- [ ] Graceful degradation when tools unavailable
- [ ] Tool health checks

## Technical Details

### Tool Directory Structure

```
tools/
├── my-script          # Shell script (chmod +x)
├── python-tool.py      # Python script
├── go-tool            # Go binary
└── README.md          # Tool documentation
```

### Tool Description Format

```json
{
  "name": "my-tool",
  "description": "Description of what tool does",
  "parameters": {
    "type": "object",
    "properties": {
      "input": {"type": "string"}
    }
  }
}
```

For custom tools, infer parameters from help text or defaults.

### Tool Execution

```go
func ExecuteTool(name string, args []string) (string, error) {
    // Find tool in tools/ or built-ins
    // Execute as subprocess
    // Capture output
    // Return result or error
}
```

## Deliverables

- [ ] Tools discovery from tools/ directory
- [ ] All built-in tools implemented
- [ ] Custom tool execution works
- [ ] Tool descriptions generated for LLM
- [ ] Tool availability checking
- [ ] Test coverage > 80% for tools code
- [ ] Tools documentation

## Acceptance Criteria

1. Executables in tools/ are discovered
2. Built-in tools work correctly
3. Custom tools execute as subprocesses
4. Tool errors are handled gracefully
5. Tool descriptions appear in LLM context
6. Missing tools show helpful error messages

## Dependencies

- **ayo-bs2**: Build system core (to embed tools)
- **ayo-bs3**: Skills system (for coordination)

## Out of Scope

- Tool development environment (users write tools however they want)
- Tool sandboxing (tools run with same permissions as agent)
- Remote tool loading (only local tools/)

## Risks

- **Security**: Tools run with agent's permissions
  - **Mitigation**: Document security implications, warn about dangerous tools
- **Compatibility**: Different platforms may not have same tools
  - **Mitigation**: Platform detection, clear error messages

## Notes

Keep tool system simple - no complex tool plugins or registration. Just executables.
