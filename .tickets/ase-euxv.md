---
id: ase-euxv
status: open
deps: [ase-3iuw, ase-mwdy]
links: []
created: 2026-02-09T03:07:32Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-k48b
---
# Implement flow executor

Execute flow steps sequentially, handling shell commands and agent invocations.

## Background

The flow executor runs a parsed flow:
1. Validate input against schema
2. Execute steps in order (respecting 'when' conditions)
3. Pass outputs between steps via template resolution
4. Record execution in SQLite
5. Return final output

## Implementation

1. Create executor:
   ```go
   type FlowExecutor struct {
       flow       *Flow
       params     map[string]any
       stepOutputs map[string]StepOutput
       db         *Repository
       matrixBroker *MatrixBroker  // For agent invocation
   }
   
   type StepOutput struct {
       Stdout string  // For shell steps
       Stderr string
       Output any     // For agent steps (parsed JSON or string)
       ExitCode int
   }
   
   func (e *FlowExecutor) Execute(ctx context.Context) (*FlowResult, error)
   ```

2. Template resolution:
   ```go
   func (e *FlowExecutor) resolveTemplate(template string) (string, error)
   ```
   
   Handle:
   - `{{ params.X }}`
   - `{{ steps.ID.stdout }}`
   - `{{ steps.ID.output }}`
   - `{{ steps.A.output // steps.B.output }}` (fallback)
   - `{{ env.VAR }}`

3. Condition evaluation:
   ```go
   func (e *FlowExecutor) evaluateCondition(when string) (bool, error)
   ```

4. Shell step execution:
   - Run in sandbox (or host if no sandbox)
   - Capture stdout, stderr, exit code
   - Fail flow on non-zero exit (unless step has ignore_errors)

5. Agent step execution:
   - Create session room if not exists
   - Send message to agent via Matrix
   - Wait for response (with timeout)
   - Parse output

6. Recording:
   - Create flow_runs record at start
   - Update on completion/failure

## Files to create/modify

- internal/flows/executor.go (new)
- internal/flows/template.go (extend for resolution)
- internal/flows/shell.go (shell execution)
- internal/flows/agent_step.go (agent invocation)

## Acceptance Criteria

- Shell steps execute and capture output
- Agent steps invoke via Matrix and collect response
- Templates resolve correctly
- Conditions skip steps appropriately
- Execution recorded in SQLite
- Errors handled gracefully with clear messages

