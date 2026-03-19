# Requirements Clarification

## Q1: Target Coverage Percentage

**Question:** What target test coverage percentage should we aim for?

**Options:**
- 70% (moderate - covers critical paths)
- 80% (high - industry standard for production code)
- 90% (very high - mission critical systems)

**Answer:** Target 80-90% coverage, with ideal goal of 100%. Focus on meaningful coverage of critical paths rather than just hitting numbers.

---

## Q2: Test Type Prioritization

**Question:** Should we prioritize unit tests, integration tests, or a mix?

**Options:**
- **Unit tests only** - Fast, isolated, test individual functions
- **Integration tests only** - Test full workflows, slower but more realistic
- **Mixed approach** - Unit tests for logic, integration tests for critical paths

**Answer:** Mixed approach - unit tests for individual functions/logic, integration tests for critical end-to-end paths (fresh → checkit → build workflow).

---

## Q3: Package Prioritization

**Question:** Which packages should we tackle first? Current state:

| Package | Coverage | Description |
|---------|----------|-------------|
| `internal/schema` | 0% | JSON schema parsing, flag generation |
| `internal/project` | 0% | Project parsing, config, validation |
| `internal/build` | 0% | Build manager, asset copying |
| `internal/cmd` | 0% | CLI commands (fresh, checkit, build) |
| `internal/generate` | 46.4% | Code generation (needs more) |
| `internal/model` | 59.1% | Provider scanning, TUI |

Should we prioritize by:
- **Dependency order** (schema → project → generate → build → cmd)
- **Risk order** (most complex/bug-prone first)
- **Coverage impact** (packages with most code first)

**Answer:** Dependency order - start from lowest level:
1. `schema` (no dependencies)
2. `project` (depends on schema)
3. `generate` (depends on schema, project)
4. `build` (depends on generate)
5. `cmd` (depends on all above)

---

## Q4: Test Patterns

**Question:** Should we follow specific testing patterns?

Looking at existing tests in `internal/generate/*_test.go` and `internal/model/*_test.go`:
- Table-driven tests for multiple scenarios
- Test helpers for common setup
- Golden file comparisons for generated code

Should we:
- **Follow existing patterns** - Match style in current tests
- **Establish new conventions** - Define standard patterns for all new tests
- **Mix** - Follow existing where applicable, add new patterns where needed

**Answer:** Establish new conventions - design testing patterns from scratch with full understanding of the application. This allows us to create consistent, well-designed test patterns rather than inheriting potentially inconsistent approaches.

---

## Q5: CI Integration

**Question:** Should tests include CI/coverage enforcement?

Options:
- **Coverage gates** - Block PRs that drop coverage below threshold
- **Coverage reports only** - Generate reports but don't block
- **Skip CI for now** - Focus on writing tests first

**Answer:** Skip CI for now - focus on writing quality tests first. Can add coverage gates later once test suite is established.

---

## Summary

| Question | Decision |
|----------|----------|
| Target coverage | 80-90% (ideal 100%) |
| Test types | Mixed: unit + integration |
| Package order | Dependency order (schema → project → generate → build → cmd) |
| Test patterns | Establish new conventions from scratch |
| CI integration | Skip for now |

**Requirements Complete?** _Pending confirmation_
