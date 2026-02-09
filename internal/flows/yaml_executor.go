// Package flows provides flow discovery, parsing, and execution.
// This file implements the YAML step-based flow executor.
package flows

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

// AgentInvoker is the interface for invoking agents from flow steps.
type AgentInvoker interface {
	// Invoke sends a prompt to an agent and returns the response.
	Invoke(ctx context.Context, agent, prompt string) (string, error)
}

// YAMLExecutor executes YAML-defined multi-step flows.
type YAMLExecutor struct {
	// AgentInvoker handles agent step execution.
	AgentInvoker AgentInvoker

	// WorkDir is the working directory for shell commands.
	WorkDir string

	// Env contains additional environment variables.
	Env map[string]string

	// HistoryService records execution history.
	HistoryService *HistoryService

	// DefaultShellTimeout is the timeout for shell steps (default 5m).
	DefaultShellTimeout time.Duration

	// DefaultAgentTimeout is the timeout for agent steps (default 10m).
	DefaultAgentTimeout time.Duration
}

// NewYAMLExecutor creates a new executor with default settings.
func NewYAMLExecutor() *YAMLExecutor {
	return &YAMLExecutor{
		DefaultShellTimeout: 5 * time.Minute,
		DefaultAgentTimeout: 10 * time.Minute,
	}
}

// Execute runs a YAML flow and returns the result.
func (e *YAMLExecutor) Execute(ctx context.Context, flow *YAMLFlow, params map[string]any) (*YAMLFlowResult, error) {
	result := &YAMLFlowResult{
		RunID:     generateYAMLRunID(),
		FlowName:  flow.Name,
		Status:    RunStatusRunning,
		Steps:     make(map[string]*StepResult),
		StartTime: time.Now(),
	}

	// Build dependency graph and execution order
	order, err := e.topologicalSort(flow.Steps)
	if err != nil {
		result.Status = RunStatusError
		result.Error = err.Error()
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, nil
	}

	// Initialize template context
	templateCtx := TemplateContext{
		Steps:  make(map[string]StepResult),
		Params: params,
		Env:    e.buildEnv(),
	}

	// Execute steps in dependency order
	// Steps without dependencies can run in parallel within the same "level"
	for _, level := range order {
		if err := e.executeLevel(ctx, flow, level, result, &templateCtx); err != nil {
			// Error already recorded in result
			break
		}
	}

	// Determine final status
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if result.Status == RunStatusRunning {
		// All steps completed - check for failures
		hasFailure := false
		for _, sr := range result.Steps {
			if sr.Status == StepStatusFailed || sr.Status == StepStatusTimeout {
				hasFailure = true
				break
			}
		}
		if hasFailure {
			result.Status = RunStatusFailed
		} else {
			result.Status = RunStatusSuccess
		}
	}

	return result, nil
}

// executeLevel executes a set of steps that can run in parallel.
func (e *YAMLExecutor) executeLevel(
	ctx context.Context,
	flow *YAMLFlow,
	stepIDs []string,
	result *YAMLFlowResult,
	templateCtx *TemplateContext,
) error {
	if len(stepIDs) == 0 {
		return nil
	}

	// Find step definitions
	stepMap := make(map[string]*FlowStep)
	for i := range flow.Steps {
		stepMap[flow.Steps[i].ID] = &flow.Steps[i]
	}

	// For a single step, execute directly
	if len(stepIDs) == 1 {
		step := stepMap[stepIDs[0]]
		sr := e.executeStep(ctx, step, templateCtx)
		result.Steps[step.ID] = sr
		templateCtx.Steps[step.ID] = *sr

		if sr.Status == StepStatusFailed && !step.ContinueOnError {
			result.Status = RunStatusFailed
			result.Error = fmt.Sprintf("step %q failed: %s", step.ID, sr.Error)
			return fmt.Errorf("step failed")
		}
		return nil
	}

	// Multiple steps - execute in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error

	for _, stepID := range stepIDs {
		step := stepMap[stepID]
		wg.Add(1)

		go func(s *FlowStep) {
			defer wg.Done()

			sr := e.executeStep(ctx, s, templateCtx)

			mu.Lock()
			result.Steps[s.ID] = sr
			templateCtx.Steps[s.ID] = *sr

			if sr.Status == StepStatusFailed && !s.ContinueOnError && firstError == nil {
				firstError = fmt.Errorf("step %q failed: %s", s.ID, sr.Error)
			}
			mu.Unlock()
		}(step)
	}

	wg.Wait()

	if firstError != nil {
		result.Status = RunStatusFailed
		result.Error = firstError.Error()
		return firstError
	}

	return nil
}

// executeStep executes a single step and returns the result.
func (e *YAMLExecutor) executeStep(ctx context.Context, step *FlowStep, templateCtx *TemplateContext) *StepResult {
	sr := &StepResult{
		ID:        step.ID,
		Status:    StepStatusRunning,
		StartTime: time.Now(),
	}

	// Check 'when' condition
	if step.When != "" {
		skip, err := e.evaluateCondition(step.When, templateCtx)
		if err != nil {
			sr.Status = StepStatusFailed
			sr.Error = fmt.Sprintf("evaluate condition: %v", err)
			sr.EndTime = time.Now()
			sr.Duration = sr.EndTime.Sub(sr.StartTime)
			return sr
		}
		if skip {
			sr.Status = StepStatusSkipped
			sr.Skipped = true
			sr.EndTime = time.Now()
			sr.Duration = sr.EndTime.Sub(sr.StartTime)
			return sr
		}
	}

	// Execute based on step type
	switch step.Type {
	case FlowStepTypeShell:
		e.executeShellStep(ctx, step, sr, templateCtx)
	case FlowStepTypeAgent:
		e.executeAgentStep(ctx, step, sr, templateCtx)
	default:
		sr.Status = StepStatusFailed
		sr.Error = fmt.Sprintf("unknown step type: %s", step.Type)
	}

	sr.EndTime = time.Now()
	sr.Duration = sr.EndTime.Sub(sr.StartTime)
	return sr
}

// executeShellStep runs a shell command.
func (e *YAMLExecutor) executeShellStep(ctx context.Context, step *FlowStep, sr *StepResult, templateCtx *TemplateContext) {
	// Resolve templates in the command
	command, err := ResolveTemplate(step.Run, *templateCtx)
	if err != nil {
		sr.Status = StepStatusFailed
		sr.Error = fmt.Sprintf("resolve template: %v", err)
		return
	}

	// Set timeout
	timeout := e.DefaultShellTimeout
	if step.Timeout != "" {
		if t, err := time.ParseDuration(step.Timeout); err == nil {
			timeout = t
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, "bash", "-c", command)

	// Set working directory
	if e.WorkDir != "" {
		cmd.Dir = e.WorkDir
	}

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range e.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range step.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	err = cmd.Run()
	sr.Stdout = stdout.String()
	sr.Stderr = stderr.String()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			sr.Status = StepStatusTimeout
			sr.Error = fmt.Sprintf("timeout after %v", timeout)
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			sr.ExitCode = exitErr.ExitCode()
			sr.Status = StepStatusFailed
			sr.Error = fmt.Sprintf("exit code %d", sr.ExitCode)
		} else {
			sr.Status = StepStatusFailed
			sr.Error = err.Error()
		}
		return
	}

	sr.Status = StepStatusSuccess
	sr.ExitCode = 0
}

// executeAgentStep invokes an agent.
func (e *YAMLExecutor) executeAgentStep(ctx context.Context, step *FlowStep, sr *StepResult, templateCtx *TemplateContext) {
	if e.AgentInvoker == nil {
		sr.Status = StepStatusFailed
		sr.Error = "no agent invoker configured"
		return
	}

	// Resolve templates in prompt and input
	prompt, err := ResolveTemplate(step.Prompt, *templateCtx)
	if err != nil {
		sr.Status = StepStatusFailed
		sr.Error = fmt.Sprintf("resolve prompt template: %v", err)
		return
	}

	if step.Input != "" {
		input, err := ResolveTemplate(step.Input, *templateCtx)
		if err != nil {
			sr.Status = StepStatusFailed
			sr.Error = fmt.Sprintf("resolve input template: %v", err)
			return
		}
		// Append input to prompt
		prompt = prompt + "\n\n" + input
	}

	if step.Context != "" {
		context, err := ResolveTemplate(step.Context, *templateCtx)
		if err != nil {
			sr.Status = StepStatusFailed
			sr.Error = fmt.Sprintf("resolve context template: %v", err)
			return
		}
		// Prepend context
		prompt = context + "\n\n" + prompt
	}

	// Set timeout
	timeout := e.DefaultAgentTimeout
	if step.Timeout != "" {
		if t, err := time.ParseDuration(step.Timeout); err == nil {
			timeout = t
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Invoke agent
	response, err := e.AgentInvoker.Invoke(ctx, step.Agent, prompt)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			sr.Status = StepStatusTimeout
			sr.Error = fmt.Sprintf("timeout after %v", timeout)
		} else {
			sr.Status = StepStatusFailed
			sr.Error = err.Error()
		}
		return
	}

	sr.Status = StepStatusSuccess
	sr.Output = response
}

// evaluateCondition evaluates a 'when' condition.
// Returns true if the step should be SKIPPED.
func (e *YAMLExecutor) evaluateCondition(when string, templateCtx *TemplateContext) (bool, error) {
	// Resolve templates in condition
	resolved, err := ResolveTemplate(when, *templateCtx)
	if err != nil {
		return false, err
	}

	resolved = trimWhitespace(resolved)

	// Simple evaluation: skip if condition is falsy
	// Falsy values: "false", "0", "", "no", "skip"
	switch resolved {
	case "false", "0", "", "no", "skip":
		return true, nil // Skip the step
	case "true", "1", "yes", "run":
		return false, nil // Run the step
	}

	// For more complex conditions, we'd need an expression evaluator
	// For now, any non-empty resolved value means "run the step"
	return false, nil
}

// topologicalSort orders steps by their dependencies.
// Returns a list of "levels" where each level contains steps that can run in parallel.
func (e *YAMLExecutor) topologicalSort(steps []FlowStep) ([][]string, error) {
	// Build adjacency list and in-degree map
	inDegree := make(map[string]int)
	dependents := make(map[string][]string) // step -> steps that depend on it
	stepSet := make(map[string]bool)

	for _, step := range steps {
		stepSet[step.ID] = true
		if _, ok := inDegree[step.ID]; !ok {
			inDegree[step.ID] = 0
		}
	}

	for _, step := range steps {
		for _, dep := range step.DependsOn {
			if !stepSet[dep] {
				return nil, fmt.Errorf("step %q depends on unknown step %q", step.ID, dep)
			}
			inDegree[step.ID]++
			dependents[dep] = append(dependents[dep], step.ID)
		}
	}

	// Kahn's algorithm with level grouping
	var levels [][]string
	var queue []string

	// Find all steps with no dependencies
	for _, step := range steps {
		if inDegree[step.ID] == 0 {
			queue = append(queue, step.ID)
		}
	}

	for len(queue) > 0 {
		// Current level
		level := queue
		queue = nil
		levels = append(levels, level)

		// Process this level
		for _, stepID := range level {
			for _, dependent := range dependents[stepID] {
				inDegree[dependent]--
				if inDegree[dependent] == 0 {
					queue = append(queue, dependent)
				}
			}
		}
	}

	// Check for cycles
	totalSteps := 0
	for _, level := range levels {
		totalSteps += len(level)
	}
	if totalSteps != len(steps) {
		return nil, fmt.Errorf("circular dependency detected in flow steps")
	}

	return levels, nil
}

// buildEnv builds the environment map for template resolution.
func (e *YAMLExecutor) buildEnv() map[string]string {
	env := make(map[string]string)

	// Add all OS environment variables
	for _, kv := range os.Environ() {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				env[kv[:i]] = kv[i+1:]
				break
			}
		}
	}

	// Add executor-level env (overrides OS env)
	for k, v := range e.Env {
		env[k] = v
	}

	return env
}

// generateYAMLRunID creates a unique run identifier.
func generateYAMLRunID() string {
	return "yr_" + ulid.Make().String()
}

// trimWhitespace removes leading/trailing whitespace from a string.
func trimWhitespace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
