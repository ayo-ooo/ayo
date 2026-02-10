---
id: ase-625q
status: closed
deps: [ase-et8m]
links: []
created: 2026-02-10T01:36:06Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-8d04
---
# Add deprecation warnings to mount commands

Add deprecation warnings to all 'ayo mount' commands directing users to use 'ayo share' instead.

## Context
The mount command will be deprecated in favor of share. During the transition period, mount should continue to work but warn users.

## File to Modify
- cmd/ayo/mount.go

## Dependencies
- ase-et8m (share add works - so we have something to recommend)

## Implementation

Add a deprecation warning function:
```go
func printMountDeprecationWarning() {
    warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
    fmt.Fprintf(os.Stderr, "%s 'ayo mount' is deprecated. Use 'ayo share' instead.\n", 
        warnStyle.Render("⚠"))
    fmt.Fprintf(os.Stderr, "  See 'ayo share --help' for usage.\n\n")
}
```

Call this at the start of each command's RunE:

### newMountAddCmd()
```go
RunE: func(cmd *cobra.Command, args []string) error {
    printMountDeprecationWarning()
    // ... existing code ...
}
```

### newMountListCmd()
```go
RunE: func(cmd *cobra.Command, args []string) error {
    printMountDeprecationWarning()
    // ... existing code ...
}
```

### newMountRmCmd()
```go
RunE: func(cmd *cobra.Command, args []string) error {
    printMountDeprecationWarning()
    // ... existing code ...
}
```

## Important Notes
1. Warning goes to stderr (not stdout) so it doesn't break JSON output
2. Commands still work normally after warning
3. Don't show warning when --json flag is used (optional - discuss)

## Alternative: Check for JSON flag
```go
func printMountDeprecationWarning(jsonOutput bool) {
    if jsonOutput {
        return // Don't pollute JSON output
    }
    // ... warning ...
}
```

## Update Long Description
Also update the command's Long description:
```go
Long: `[DEPRECATED] Use 'ayo share' instead.

Manage persistent filesystem access for sandboxed agents.
...
```

## Acceptance Criteria

- [ ] Warning shown for 'ayo mount add'
- [ ] Warning shown for 'ayo mount list'
- [ ] Warning shown for 'ayo mount rm'
- [ ] Warning goes to stderr
- [ ] Commands still work after warning
- [ ] Warning mentions 'ayo share' alternative
- [ ] Long description mentions deprecation
- [ ] JSON output not affected

