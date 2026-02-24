---
id: ayo-c5mt
status: closed
deps: [ayo-dicu]
links: []
created: 2026-02-23T22:15:27Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-whmn
tags: [ui, approval]
---
# Implement file request approval UI

Add terminal UI for file_request approval prompts.

## UI Design

```
┌─────────────────────────────────────────────────────────┐
│ @ayo wants to update:                                   │
│   ~/Projects/app/main.go                                │
│                                                         │
│ Reason: Fixed authentication bug in login handler       │
│                                                         │
│ ─────────────────────────────────────────────────────── │
│ [Y]es  [N]o  [D]iff  [A]lways for session  [?]Help      │
└─────────────────────────────────────────────────────────┘
```

## Options

| Key | Action |
|-----|--------|
| Y | Approve this request |
| N | Deny this request |
| D | Show diff (for update actions) |
| A | Approve all future requests this session |
| ? | Show help text |

## Implementation

### Location
`internal/ui/approval.go`

### Library
Use `github.com/charmbracelet/huh` for form-based prompts, or `github.com/charmbracelet/bubbletea` for more complex UI.

### Interface

```go
type ApprovalRequest struct {
    Agent   string
    Action  string // create, update, delete
    Path    string // Host path (converted from /mnt/{user}/...)
    Content string // For diff display
    Reason  string
}

type ApprovalResponse struct {
    Approved      bool
    AlwaysApprove bool // "A" was selected
}

func PromptApproval(req ApprovalRequest) (ApprovalResponse, error)
```

### Diff Display

When user presses 'D':
- Read current file from host
- Show unified diff using `github.com/sergi/go-diff` or similar
- Return to prompt after viewing

## Files to Create/Modify

- Create `internal/ui/approval.go`
- Add diff library to go.mod
- Wire into daemon RPC handler

## Testing

- Test with mock stdin for automated testing
- Test diff display with various file sizes
- Test session caching ("A" option)
