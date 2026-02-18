---
id: am-mjx7
status: closed
deps: [am-hsum]
links: []
created: 2026-02-18T03:18:30Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-hin9
---
# Add # symbol to shell completion

Update shell completions to suggest squad names with # prefix.

## Context
- Tab completion should work for #squad
- List available squads as completion candidates

## Implementation
Update completion generators:
```go
// cmd/ayo/completion.go or similar

func squadCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    if !strings.HasPrefix(toComplete, "#") {
        return nil, cobra.ShellCompDirectiveNoFileComp
    }
    
    prefix := strings.TrimPrefix(toComplete, "#")
    squads, _ := daemon.ListSquads(context.Background())
    
    var completions []string
    for _, s := range squads {
        if strings.HasPrefix(s.Name, prefix) {
            completions = append(completions, "#"+s.Name)
        }
    }
    return completions, cobra.ShellCompDirectiveNoFileComp
}
```

## Files to Modify
- cmd/ayo/completion.go or root.go

## Dependencies
- am-hsum (CLI parsing)

## Acceptance
- ayo #front<TAB> completes to #frontend-team
- Works in bash, zsh, fish

