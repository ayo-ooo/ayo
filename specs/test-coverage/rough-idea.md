# Test Coverage Improvement

## Rough Idea

The Ayo CLI project currently has **37.7% test coverage** with significant gaps:

| Package | Coverage | Priority |
|---------|----------|----------|
| `internal/template` | 91.1% | Done |
| `internal/model` | 59.1% | Medium |
| `internal/generate` | 46.4% | Medium |
| `internal/build` | 0.0% | High |
| `internal/cmd` | 0.0% | High |
| `internal/project` | 0.0% | High |
| `internal/schema` | 0.0% | High |
| `cmd/ayo` | 0.0% | Low (entry point) |

## Goals

1. Increase overall test coverage to a target percentage
2. Ensure critical paths are tested (project parsing, schema handling, code generation, build process)
3. Create tests that are maintainable and document expected behavior
4. Identify and fix bugs discovered through testing

## Key Questions

- What target coverage percentage should we aim for?
- Should we prioritize unit tests, integration tests, or both?
- What are the critical paths that must be tested?
- Are there existing patterns in the tested modules we should follow?
