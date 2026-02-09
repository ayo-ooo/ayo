---
id: ase-deny
status: closed
deps: [ase-0oyk]
links: []
created: 2026-02-09T03:11:13Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-cjpe
---
# Integrate capabilities with @ayo planning

## Background

@ayo is the executive agent that orchestrates other agents. When @ayo receives a complex task, it needs to break it down and delegate subtasks to appropriate agents. This ticket integrates the capability inference and semantic search systems with @ayo's planning logic.

## Why This Matters

Without this integration:
- @ayo would need to be manually told which agents exist
- @ayo couldn't automatically discover new agents
- Task delegation would be hardcoded or require explicit configuration

With integration:
- @ayo dynamically discovers agents based on capabilities
- @ayo can explain WHY it chose a particular agent
- New agents automatically become available for delegation

## Implementation Details

### @ayo's Enhanced System Prompt

Add capability awareness to @ayo's system prompt:

```
You are @ayo, the executive agent for this system.

When delegating tasks, use the `find_agent` tool to discover capable agents:

<tool name="find_agent">
  <description>Find agents capable of performing a task</description>
  <parameters>
    <task>Description of the task to delegate</task>
    <count>Number of candidates to return (default: 3)</count>
  </parameters>
</tool>

Example:
User: "Review the authentication code for security issues"
You: <find_agent task="security code review" count="3"/>
System returns: @security-auditor (0.94), @code-reviewer (0.78), @senior-dev (0.65)
You: "I'll delegate this to @security-auditor who specializes in security analysis."
```

### find_agent Tool Implementation

```go
// internal/tools/find_agent/find_agent.go
type FindAgentTool struct {
    capSearch *capabilities.CapabilitySearch
}

type FindAgentParams struct {
    Task  string `json:"task"`
    Count int    `json:"count,omitempty"`
}

type FindAgentResult struct {
    Agents []AgentMatch `json:"agents"`
}

type AgentMatch struct {
    Name        string  `json:"name"`
    Similarity  float64 `json:"similarity"`
    Capability  string  `json:"matching_capability"`
    Description string  `json:"description"`
}

func (t *FindAgentTool) Execute(params FindAgentParams) (*FindAgentResult, error) {
    if params.Count == 0 {
        params.Count = 3
    }
    
    results, err := t.capSearch.Search(params.Task, params.Count)
    if err != nil {
        return nil, err
    }
    
    matches := make([]AgentMatch, len(results))
    for i, r := range results {
        matches[i] = AgentMatch{
            Name:        r.AgentName,
            Similarity:  r.Similarity,
            Capability:  r.Capability.Name,
            Description: r.Capability.Description,
        }
    }
    
    return &FindAgentResult{Agents: matches}, nil
}
```

### Planning Flow

```
1. User gives task to @ayo
2. @ayo breaks down task into subtasks
3. For each subtask:
   a. Call find_agent tool
   b. Receive ranked agent list
   c. Select best agent (or create new one if no match)
4. Execute delegation via Matrix messages
5. Aggregate results
```

### Files to Create/Modify

1. Create `internal/tools/find_agent/find_agent.go` - Tool implementation
2. Create `internal/tools/find_agent/find_agent_test.go` - Tests
3. Modify `internal/builtin/agents/ayo/AYO.md` - Add tool to @ayo's config
4. Modify `internal/agent/invoke.go` - Register find_agent tool for @ayo

### Edge Cases

- **No matching agents**: @ayo should consider creating a new agent (links to ase-yqtq)
- **Low similarity scores**: @ayo should express uncertainty
- **Multiple equally good matches**: @ayo picks based on additional context or asks user

## Acceptance Criteria

- [ ] find_agent tool implemented and registered
- [ ] Tool returns ranked agent matches with similarity scores
- [ ] @ayo's system prompt includes tool documentation
- [ ] @ayo can explain agent selection reasoning
- [ ] Handles no-match case gracefully
- [ ] Handles low-confidence matches appropriately
- [ ] Integration test with @ayo selecting agent for task

