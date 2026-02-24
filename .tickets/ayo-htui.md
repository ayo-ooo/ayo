---
id: ayo-htui
status: closed
deps: [ayo-hscm]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-hitl
tags: [human-in-the-loop, tui, forms]
---
# Task: Implement bubbletea/huh Form Renderer

## Summary

Create a form renderer that takes an InputRequest schema and renders it as an interactive TUI form using charmbracelet/huh. This provides a rich form experience in the terminal.

## Dependencies

- `github.com/charmbracelet/huh` - Form library
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling

## Implementation

### FormRenderer Interface

```go
type FormRenderer interface {
    // Render displays the form and blocks until complete or cancelled
    Render(ctx context.Context, req *InputRequest) (*InputResponse, error)
}

type CLIFormRenderer struct {
    // Uses huh to render forms
}
```

### Field Type Mapping

| Schema Type | huh Component |
|-------------|---------------|
| `text` | `huh.NewInput()` |
| `textarea` | `huh.NewText()` |
| `select` | `huh.NewSelect()` |
| `multiselect` | `huh.NewMultiSelect()` |
| `confirm` | `huh.NewConfirm()` |
| `number` | `huh.NewInput()` with validation |
| `date` | `huh.NewInput()` with date parsing |
| `file` | `huh.NewFilePicker()` |

### Styling

- Use consistent theme with ayo's existing TUI
- Support dark/light mode
- Accessible colors (WCAG compliant)

## Files to Create

- `internal/hitl/cli_form.go` - CLI form renderer
- `internal/hitl/cli_form_test.go` - Tests

## Acceptance Criteria

- [ ] All field types render correctly
- [ ] Form blocks until submitted or cancelled
- [ ] Validation errors shown inline
- [ ] Tab/arrow key navigation works
- [ ] Escape cancels form
- [ ] Enter submits form
- [ ] Styling matches ayo theme
