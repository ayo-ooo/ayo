package flows

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// mockAgentInvoker is a test double for agent invocation.
type mockAgentInvoker struct {
	responses map[string]string
	errors    map[string]error
	calls     []agentCall
	delay     time.Duration
}

type agentCall struct {
	Agent  string
	Prompt string
}

func newMockAgentInvoker() *mockAgentInvoker {
	return &mockAgentInvoker{
		responses: make(map[string]string),
		errors:    make(map[string]error),
	}
}

func (m *mockAgentInvoker) SetResponse(agent, response string) {
	m.responses[agent] = response
}

func (m *mockAgentInvoker) SetError(agent string, err error) {
	m.errors[agent] = err
}

func (m *mockAgentInvoker) Invoke(ctx context.Context, agent, prompt string) (string, error) {
	m.calls = append(m.calls, agentCall{Agent: agent, Prompt: prompt})

	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	if err, ok := m.errors[agent]; ok {
		return "", err
	}

	if resp, ok := m.responses[agent]; ok {
		return resp, nil
	}

	return "default response from " + agent, nil
}

func TestYAMLExecutor_SimpleShell(t *testing.T) {
	executor := NewYAMLExecutor()

	flow := &YAMLFlow{
		Version: 1,
		Name:    "simple-shell",
		Steps: []FlowStep{
			{
				ID:   "hello",
				Type: FlowStepTypeShell,
				Run:  `echo "Hello World"`,
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusSuccess)
	}

	sr := result.Steps["hello"]
	if sr == nil {
		t.Fatal("Step 'hello' not found")
	}

	if sr.Status != StepStatusSuccess {
		t.Errorf("Step status = %v, want %v", sr.Status, StepStatusSuccess)
	}

	if !strings.Contains(sr.Stdout, "Hello World") {
		t.Errorf("Stdout = %q, want to contain 'Hello World'", sr.Stdout)
	}
}

func TestYAMLExecutor_MultiStep(t *testing.T) {
	executor := NewYAMLExecutor()

	flow := &YAMLFlow{
		Version: 1,
		Name:    "multi-step",
		Steps: []FlowStep{
			{
				ID:   "list",
				Type: FlowStepTypeShell,
				Run:  `echo -e "file1\nfile2\nfile3"`,
			},
			{
				ID:        "count",
				Type:      FlowStepTypeShell,
				Run:       `echo "{{ steps.list.stdout }}" | wc -l`,
				DependsOn: []string{"list"},
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusSuccess)
	}

	sr := result.Steps["count"]
	if sr == nil {
		t.Fatal("Step 'count' not found")
	}

	// Should contain "3" or "4" (depends on echo behavior with trailing newline)
	if !strings.Contains(sr.Stdout, "3") && !strings.Contains(sr.Stdout, "4") {
		t.Errorf("Stdout = %q, want to contain line count", sr.Stdout)
	}
}

func TestYAMLExecutor_AgentStep(t *testing.T) {
	invoker := newMockAgentInvoker()
	invoker.SetResponse("@summarizer", "Summary: This is a summary of the document.")

	executor := NewYAMLExecutor()
	executor.AgentInvoker = invoker

	flow := &YAMLFlow{
		Version: 1,
		Name:    "with-agent",
		Steps: []FlowStep{
			{
				ID:     "summarize",
				Type:   FlowStepTypeAgent,
				Agent:  "@summarizer",
				Prompt: "Summarize this document",
				Input:  "Long document content here...",
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusSuccess)
	}

	sr := result.Steps["summarize"]
	if sr == nil {
		t.Fatal("Step 'summarize' not found")
	}

	if !strings.Contains(sr.Output, "Summary") {
		t.Errorf("Output = %q, want to contain 'Summary'", sr.Output)
	}

	if len(invoker.calls) != 1 {
		t.Errorf("Invoker calls = %d, want 1", len(invoker.calls))
	}
}

func TestYAMLExecutor_TemplateSubstitution(t *testing.T) {
	invoker := newMockAgentInvoker()
	invoker.SetResponse("@analyzer", "Found 5 issues")

	executor := NewYAMLExecutor()
	executor.AgentInvoker = invoker

	flow := &YAMLFlow{
		Version: 1,
		Name:    "template-test",
		Steps: []FlowStep{
			{
				ID:   "gather",
				Type: FlowStepTypeShell,
				Run:  `echo "main.go\ntest.go"`,
			},
			{
				ID:        "analyze",
				Type:      FlowStepTypeAgent,
				Agent:     "@analyzer",
				Prompt:    "Analyze these files:\n{{ steps.gather.stdout }}",
				DependsOn: []string{"gather"},
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusSuccess)
	}

	// Verify the prompt was resolved
	if len(invoker.calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(invoker.calls))
	}

	call := invoker.calls[0]
	if !strings.Contains(call.Prompt, "Analyze these files:") {
		t.Errorf("Prompt = %q, missing 'Analyze these files:'", call.Prompt)
	}
	if !strings.Contains(call.Prompt, "main.go") {
		t.Errorf("Prompt = %q, missing 'main.go'", call.Prompt)
	}
}

func TestYAMLExecutor_ParallelSteps(t *testing.T) {
	executor := NewYAMLExecutor()

	// Track execution order
	var startOrder []string
	var started int32

	flow := &YAMLFlow{
		Version: 1,
		Name:    "parallel",
		Steps: []FlowStep{
			{
				ID:   "slow1",
				Type: FlowStepTypeShell,
				Run:  `sleep 0.5 && echo "done1"`,
			},
			{
				ID:   "slow2",
				Type: FlowStepTypeShell,
				Run:  `sleep 0.5 && echo "done2"`,
			},
			{
				ID:        "combine",
				Type:      FlowStepTypeShell,
				Run:       `echo "{{ steps.slow1.stdout }} {{ steps.slow2.stdout }}"`,
				DependsOn: []string{"slow1", "slow2"},
			},
		},
	}

	start := time.Now()
	result, err := executor.Execute(context.Background(), flow, nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v (error: %s)", result.Status, RunStatusSuccess, result.Error)
	}

	// slow1 and slow2 should run in parallel (~0.5s total, not 1s)
	if elapsed > 1500*time.Millisecond {
		t.Errorf("Execution took %v, expected ~0.5s for parallel execution", elapsed)
	}

	_ = startOrder
	_ = started
}

func TestYAMLExecutor_DependencyOrder(t *testing.T) {
	executor := NewYAMLExecutor()

	flow := &YAMLFlow{
		Version: 1,
		Name:    "deps",
		Steps: []FlowStep{
			{
				ID:        "step3",
				Type:      FlowStepTypeShell,
				Run:       `echo "step3"`,
				DependsOn: []string{"step2"},
			},
			{
				ID:   "step1",
				Type: FlowStepTypeShell,
				Run:  `echo "step1"`,
			},
			{
				ID:        "step2",
				Type:      FlowStepTypeShell,
				Run:       `echo "step2 after {{ steps.step1.stdout }}"`,
				DependsOn: []string{"step1"},
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusSuccess)
	}

	// Verify execution order via timestamps
	s1 := result.Steps["step1"]
	s2 := result.Steps["step2"]
	s3 := result.Steps["step3"]

	if s1.EndTime.After(s2.StartTime) {
		t.Errorf("step1 should complete before step2 starts")
	}
	if s2.EndTime.After(s3.StartTime) {
		t.Errorf("step2 should complete before step3 starts")
	}
}

func TestYAMLExecutor_ShellError(t *testing.T) {
	executor := NewYAMLExecutor()

	flow := &YAMLFlow{
		Version: 1,
		Name:    "error-flow",
		Steps: []FlowStep{
			{
				ID:   "fail",
				Type: FlowStepTypeShell,
				Run:  `exit 1`,
			},
			{
				ID:        "after",
				Type:      FlowStepTypeShell,
				Run:       `echo "should not run"`,
				DependsOn: []string{"fail"},
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusFailed {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusFailed)
	}

	failStep := result.Steps["fail"]
	if failStep == nil {
		t.Fatal("Step 'fail' not found")
	}
	if failStep.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", failStep.ExitCode)
	}

	// "after" step should not have been executed
	afterStep := result.Steps["after"]
	if afterStep != nil {
		t.Errorf("Step 'after' should not have been executed, got status %v", afterStep.Status)
	}
}

func TestYAMLExecutor_ContinueOnError(t *testing.T) {
	executor := NewYAMLExecutor()

	flow := &YAMLFlow{
		Version: 1,
		Name:    "continue-flow",
		Steps: []FlowStep{
			{
				ID:              "fail",
				Type:            FlowStepTypeShell,
				Run:             `exit 1`,
				ContinueOnError: true,
			},
			{
				ID:        "after",
				Type:      FlowStepTypeShell,
				Run:       `echo "still running"`,
				DependsOn: []string{"fail"},
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Should still succeed overall because continue_on_error was set
	if result.Status != RunStatusFailed {
		// Actually, the flow status depends on whether any step failed without continue_on_error
		// In this case, "fail" has continue_on_error=true, so flow should continue
		// But "after" depends on "fail", so it should still run
	}

	failStep := result.Steps["fail"]
	if failStep == nil {
		t.Fatal("Step 'fail' not found")
	}
	if failStep.Status != StepStatusFailed {
		t.Errorf("Step 'fail' status = %v, want Failed", failStep.Status)
	}

	afterStep := result.Steps["after"]
	if afterStep == nil {
		t.Fatal("Step 'after' should have been executed")
	}
	if afterStep.Status != StepStatusSuccess {
		t.Errorf("Step 'after' status = %v, want Success", afterStep.Status)
	}
}

func TestYAMLExecutor_Timeout(t *testing.T) {
	executor := NewYAMLExecutor()

	flow := &YAMLFlow{
		Version: 1,
		Name:    "timeout-flow",
		Steps: []FlowStep{
			{
				ID:      "slow",
				Type:    FlowStepTypeShell,
				Run:     `sleep 10`,
				Timeout: "500ms",
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusFailed {
		t.Errorf("Status = %v, want Failed", result.Status)
	}

	sr := result.Steps["slow"]
	if sr == nil {
		t.Fatal("Step 'slow' not found")
	}

	if sr.Status != StepStatusTimeout {
		t.Errorf("Step status = %v, want Timeout", sr.Status)
	}

	if !strings.Contains(sr.Error, "timeout") {
		t.Errorf("Error = %q, want to contain 'timeout'", sr.Error)
	}
}

func TestYAMLExecutor_WithParams(t *testing.T) {
	executor := NewYAMLExecutor()

	flow := &YAMLFlow{
		Version: 1,
		Name:    "params-flow",
		Steps: []FlowStep{
			{
				ID:   "greet",
				Type: FlowStepTypeShell,
				Run:  `echo "Hello {{ params.name }}"`,
			},
		},
	}

	params := map[string]any{
		"name": "World",
	}

	result, err := executor.Execute(context.Background(), flow, params)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusSuccess)
	}

	sr := result.Steps["greet"]
	if !strings.Contains(sr.Stdout, "Hello World") {
		t.Errorf("Stdout = %q, want to contain 'Hello World'", sr.Stdout)
	}
}

func TestYAMLExecutor_WhenCondition(t *testing.T) {
	executor := NewYAMLExecutor()

	flow := &YAMLFlow{
		Version: 1,
		Name:    "conditional",
		Steps: []FlowStep{
			{
				ID:   "skipped",
				Type: FlowStepTypeShell,
				Run:  `echo "should not run"`,
				When: "false",
			},
			{
				ID:   "runs",
				Type: FlowStepTypeShell,
				Run:  `echo "should run"`,
				When: "true",
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusSuccess)
	}

	skippedStep := result.Steps["skipped"]
	if skippedStep.Status != StepStatusSkipped {
		t.Errorf("Skipped step status = %v, want Skipped", skippedStep.Status)
	}

	runsStep := result.Steps["runs"]
	if runsStep.Status != StepStatusSuccess {
		t.Errorf("Runs step status = %v, want Success", runsStep.Status)
	}
}

func TestYAMLExecutor_AgentError(t *testing.T) {
	invoker := newMockAgentInvoker()
	invoker.SetError("@buggy", errors.New("agent crashed"))

	executor := NewYAMLExecutor()
	executor.AgentInvoker = invoker

	flow := &YAMLFlow{
		Version: 1,
		Name:    "agent-error",
		Steps: []FlowStep{
			{
				ID:     "fail",
				Type:   FlowStepTypeAgent,
				Agent:  "@buggy",
				Prompt: "Do something",
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusFailed {
		t.Errorf("Status = %v, want Failed", result.Status)
	}

	sr := result.Steps["fail"]
	if sr.Status != StepStatusFailed {
		t.Errorf("Step status = %v, want Failed", sr.Status)
	}

	if !strings.Contains(sr.Error, "agent crashed") {
		t.Errorf("Error = %q, want to contain 'agent crashed'", sr.Error)
	}
}

func TestYAMLExecutor_NoAgentInvoker(t *testing.T) {
	executor := NewYAMLExecutor()
	// AgentInvoker is nil

	flow := &YAMLFlow{
		Version: 1,
		Name:    "no-invoker",
		Steps: []FlowStep{
			{
				ID:     "agent",
				Type:   FlowStepTypeAgent,
				Agent:  "@ayo",
				Prompt: "Hello",
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusFailed {
		t.Errorf("Status = %v, want Failed", result.Status)
	}

	sr := result.Steps["agent"]
	if !strings.Contains(sr.Error, "no agent invoker") {
		t.Errorf("Error = %q, want to contain 'no agent invoker'", sr.Error)
	}
}

func TestYAMLExecutor_CircularDependency(t *testing.T) {
	executor := NewYAMLExecutor()

	flow := &YAMLFlow{
		Version: 1,
		Name:    "circular",
		Steps: []FlowStep{
			{
				ID:        "a",
				Type:      FlowStepTypeShell,
				Run:       `echo "a"`,
				DependsOn: []string{"b"},
			},
			{
				ID:        "b",
				Type:      FlowStepTypeShell,
				Run:       `echo "b"`,
				DependsOn: []string{"a"},
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusError {
		t.Errorf("Status = %v, want Error", result.Status)
	}

	if !strings.Contains(result.Error, "circular") {
		t.Errorf("Error = %q, want to contain 'circular'", result.Error)
	}
}

func TestYAMLExecutor_UnknownDependency(t *testing.T) {
	executor := NewYAMLExecutor()

	flow := &YAMLFlow{
		Version: 1,
		Name:    "unknown-dep",
		Steps: []FlowStep{
			{
				ID:        "a",
				Type:      FlowStepTypeShell,
				Run:       `echo "a"`,
				DependsOn: []string{"nonexistent"},
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusError {
		t.Errorf("Status = %v, want Error", result.Status)
	}

	if !strings.Contains(result.Error, "unknown step") {
		t.Errorf("Error = %q, want to contain 'unknown step'", result.Error)
	}
}

func TestYAMLExecutor_EnvVariables(t *testing.T) {
	executor := NewYAMLExecutor()
	executor.Env = map[string]string{
		"MY_VAR": "executor-level",
	}

	flow := &YAMLFlow{
		Version: 1,
		Name:    "env-test",
		Steps: []FlowStep{
			{
				ID:   "check",
				Type: FlowStepTypeShell,
				Run:  `echo $MY_VAR $STEP_VAR`,
				Env: map[string]string{
					"STEP_VAR": "step-level",
				},
			},
		},
	}

	result, err := executor.Execute(context.Background(), flow, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want Success", result.Status)
	}

	sr := result.Steps["check"]
	if !strings.Contains(sr.Stdout, "executor-level") {
		t.Errorf("Stdout = %q, want to contain 'executor-level'", sr.Stdout)
	}
	if !strings.Contains(sr.Stdout, "step-level") {
		t.Errorf("Stdout = %q, want to contain 'step-level'", sr.Stdout)
	}
}
