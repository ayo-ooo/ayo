---
id: am-vl6l
status: closed
deps: []
links: []
created: 2026-02-18T03:13:23Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Define PlannerPlugin interface

Create the core interface that all planner plugins must implement. This interface will be used by both near-term and long-term planners.

## Context
- Location: internal/planners/interface.go (new file)
- Planners expose tools to agents and instructions for system prompts
- Planners manage their own state within a sandbox directory

## Interface Design
```go
type PlannerPlugin interface {
    // Metadata
    Name() string
    Type() PlannerType // NearTerm or LongTerm
    
    // Lifecycle
    Init(ctx PlannerContext) error
    Close() error
    
    // For LLM integration
    Tools() []tools.Definition
    Instructions() string
    
    // State (sandbox-scoped)
    StateDir() string
}

type PlannerContext struct {
    SandboxName string
    SandboxDir  string
    StateDir    string  // Where plugin stores its state
    Config      map[string]any
}

type PlannerType string
const (
    NearTerm PlannerType = "near"
    LongTerm PlannerType = "long"
)
```

## Files to Create
- internal/planners/interface.go
- internal/planners/types.go
- internal/planners/doc.go

## Acceptance
- Interface compiles
- Types exported
- Documentation comments on all exported types

