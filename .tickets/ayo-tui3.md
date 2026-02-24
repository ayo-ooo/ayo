---
id: ayo-tui3
status: closed
deps: [ayo-tui1]
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-memx
tags: [tui, interactive, tools]
---
# Implement inline tool display

Implement inline display for tool calls in interactive mode.

## Design

Tools appear inline with the conversation, using chevron and indentation:

```
@ayo: I'll check the files.

  ▸ bash: ls -la src/
    └─ 12 files

  ▸ view: src/main.go:1-50
    └─ Read 50 lines

  ▸ edit: src/main.go
    └─ Added error handling

The changes have been made.
```

## Tool States

| State | Display |
|-------|---------|
| Running | `▸ bash: command...` (with spinner) |
| Success | `▸ bash: command` + `└─ result summary` |
| Error | `▸ bash: command` + `└─ ✗ error message` |

## Implementation

### ToolDisplay

```go
// internal/ui/interactive/tool.go
type ToolDisplay struct {
    name    string
    input   string
    state   ToolState
    result  string
    spinner spinner.Model
}

func (t *ToolDisplay) View() string {
    var b strings.Builder
    
    // Tool line
    if t.state == Running {
        b.WriteString(fmt.Sprintf("  %s %s: %s\n", 
            t.spinner.View(), t.name, t.input))
    } else {
        b.WriteString(fmt.Sprintf("  ▸ %s: %s\n", t.name, t.input))
    }
    
    // Result line (if complete)
    if t.state == Success {
        b.WriteString(fmt.Sprintf("    └─ %s\n", t.result))
    } else if t.state == Error {
        b.WriteString(fmt.Sprintf("    └─ ✗ %s\n", t.result))
    }
    
    return b.String()
}
```

### Result Summarization

Each tool type has a summarizer:

```go
type ToolSummarizer func(output string) string

var summarizers = map[string]ToolSummarizer{
    "bash": func(output string) string {
        lines := strings.Count(output, "\n")
        if lines > 5 {
            return fmt.Sprintf("%d lines of output", lines)
        }
        return strings.TrimSpace(output)
    },
    "view": func(output string) string {
        lines := strings.Count(output, "\n")
        return fmt.Sprintf("Read %d lines", lines)
    },
    "edit": func(output string) string {
        return "File updated"
    },
}
```

## Nested Tool Calls

For sub-agent invocations, show hierarchy:

```
  ▸ delegate: @reviewer
    ├─ view: src/main.go
    │   └─ Read 100 lines
    └─ Approved with minor suggestions
```

## Testing

- Test all tool states render correctly
- Test result summarization
- Test nested tool display
- Test long output truncation
