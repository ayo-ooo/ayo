---
id: ayo-0e81
status: closed
deps: [ayo-66df, ayo-dicu]
links: []
created: 2026-02-23T22:15:38Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-whmn
tags: [tools, filesystem]
---
# Add publish tool for moving files to host

Create a `publish` tool that allows agents to move files from the `/output/` safe write zone to user-specified host locations. This provides a controlled path for agents to "deliver" completed work.

## Context

Agents can freely write to `/output/` (ayo-66df) without approval. The `publish` tool provides a way to move those files to the host filesystem with user approval (via file_request flow from ayo-dicu).

Workflow:
1. Agent writes files to `/output/` (no approval needed)
2. Agent calls `publish` to move files to host
3. User sees approval prompt with file preview
4. On approval, files are moved to host destination

## Tool Definition

```json
{
  "name": "publish",
  "description": "Move files from /output/ to host filesystem",
  "parameters": {
    "type": "object",
    "properties": {
      "files": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "source": { "type": "string", "description": "Path in /output/" },
            "destination": { "type": "string", "description": "Host path" }
          }
        },
        "description": "Files to publish"
      },
      "message": {
        "type": "string",
        "description": "Optional message explaining what's being published"
      }
    },
    "required": ["files"]
  }
}
```

## Example Usage

```json
{
  "name": "publish",
  "arguments": {
    "files": [
      { "source": "/output/report.pdf", "destination": "~/Documents/report.pdf" },
      { "source": "/output/data.csv", "destination": "~/Documents/data.csv" }
    ],
    "message": "Generated analysis report and supporting data"
  }
}
```

## Implementation

```go
// internal/tools/publish.go
func (t *PublishTool) Execute(ctx context.Context, args PublishArgs) (string, error) {
    // Validate all sources exist in /output/
    for _, f := range args.Files {
        if !strings.HasPrefix(f.Source, "/output/") {
            return "", fmt.Errorf("source must be in /output/: %s", f.Source)
        }
    }
    
    // Create file_request for batch approval
    req := FileRequest{
        Type:    "publish",
        Files:   args.Files,
        Message: args.Message,
    }
    
    // Send to host for approval
    approved, err := t.daemon.RequestApproval(ctx, req)
    if err != nil || !approved {
        return "Publish request denied", nil
    }
    
    // Move files on approval
    for _, f := range args.Files {
        if err := t.moveToHost(f.Source, f.Destination); err != nil {
            return "", err
        }
    }
    
    return fmt.Sprintf("Published %d files", len(args.Files)), nil
}
```

## Files to Create/Modify

1. **`internal/tools/publish.go`** (new) - Tool implementation
2. **`internal/tools/tools.go`** - Register publish tool
3. **`internal/ui/approval.go`** - Handle batch publish approval UI

## Approval UI for Publish

Show a special UI for publish operations:
```
┌─ Publish Request ───────────────────────────┐
│ Agent wants to publish 2 files:             │
│                                             │
│   /output/report.pdf → ~/Documents/report.pdf
│   /output/data.csv   → ~/Documents/data.csv │
│                                             │
│ Message: Generated analysis report          │
│                                             │
│ [Allow]  [Allow All]  [Deny]  [Preview]     │
└─────────────────────────────────────────────┘
```

## Acceptance Criteria

- [ ] publish tool validates sources are in /output/
- [ ] Tool creates batch file_request for approval
- [ ] User can preview files before approving
- [ ] Files are moved (not copied) on approval
- [ ] Operation is atomic (all files or none)
- [ ] Audit log records publish operations
- [ ] Works with session approval cache

## Testing

- Test source validation (must be in /output/)
- Test batch approval flow
- Test atomic move operation
- Test audit logging
- Test error handling for missing files
