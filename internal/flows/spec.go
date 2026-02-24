// Package flows provides flow discovery, parsing, and execution.
// This file defines the YAML flow specification for multi-step orchestrated flows.

package flows

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// YAMLFlow represents a multi-step flow defined in YAML.
// This is distinct from the shell-script based Flow type.
type YAMLFlow struct {
	// Version of the flow specification. Currently only 1 is supported.
	Version int `yaml:"version" json:"version"`

	// Name is the flow identifier. Must be unique within a directory.
	Name string `yaml:"name" json:"name"`

	// Description explains what the flow does.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// CreatedBy is the agent or user that created this flow.
	// Common values: "@ayo", "user"
	CreatedBy string `yaml:"created_by,omitempty" json:"created_by,omitempty"`

	// CreatedAt is when the flow was created.
	CreatedAt time.Time `yaml:"created_at,omitempty" json:"created_at,omitempty"`

	// Input schema for flow parameters (optional JSON Schema).
	Input *FlowSchema `yaml:"input,omitempty" json:"input,omitempty"`

	// Output schema for flow result (optional JSON Schema).
	Output *FlowSchema `yaml:"output,omitempty" json:"output,omitempty"`

	// Steps are the execution steps in order.
	Steps []FlowStep `yaml:"steps" json:"steps"`

	// Triggers define automatic execution conditions (optional).
	Triggers []FlowTrigger `yaml:"triggers,omitempty" json:"triggers,omitempty"`
}

// FlowSchema is a simplified JSON Schema representation.
type FlowSchema struct {
	Type       string                 `yaml:"type" json:"type"`
	Properties map[string]*FlowSchema `yaml:"properties,omitempty" json:"properties,omitempty"`
	Required   []string               `yaml:"required,omitempty" json:"required,omitempty"`
	Default    any            `yaml:"default,omitempty" json:"default,omitempty"`

	// Additional schema fields
	Items       *FlowSchema `yaml:"items,omitempty" json:"items,omitempty"` // For arrays
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Enum        []string    `yaml:"enum,omitempty" json:"enum,omitempty"`
}

// FlowStep represents a single step in a flow.
type FlowStep struct {
	// ID is the unique identifier for this step within the flow.
	// Used in template references like {{ steps.ID.stdout }}
	ID string `yaml:"id" json:"id"`

	// Type is the step type: "shell" or "agent"
	Type FlowStepType `yaml:"type" json:"type"`

	// === Shell step fields ===

	// Run is the shell command to execute (for type: shell).
	Run string `yaml:"run,omitempty" json:"run,omitempty"`

	// === Agent step fields ===

	// Agent is the agent handle to invoke (for type: agent).
	// Example: "@summarizer"
	Agent string `yaml:"agent,omitempty" json:"agent,omitempty"`

	// Squad is the squad to run the agent in (for type: agent).
	// If specified, the agent runs in the squad's sandbox with squad context.
	// If empty, the agent runs in the @ayo sandbox.
	// Example: "#frontend-team"
	Squad string `yaml:"squad,omitempty" json:"squad,omitempty"`

	// Prompt is the prompt to send to the agent.
	Prompt string `yaml:"prompt,omitempty" json:"prompt,omitempty"`

	// Context is optional additional context for the agent.
	// Prepended to the prompt as invocation context.
	Context string `yaml:"context,omitempty" json:"context,omitempty"`

	// Input is the data to pass to the agent (supports templates).
	Input string `yaml:"input,omitempty" json:"input,omitempty"`

	// === Common fields ===

	// DependsOn lists step IDs that must complete before this step runs.
	// If empty, the step can run immediately or in parallel.
	DependsOn []string `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`

	// When is a conditional expression. If it evaluates to false, the step is skipped.
	// Supports templates: {{ params.language != "english" }}
	When string `yaml:"when,omitempty" json:"when,omitempty"`

	// Timeout is the maximum duration for this step.
	// Default: 5m for shell, 10m for agent.
	Timeout string `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// ContinueOnError if true, continues flow even if this step fails.
	ContinueOnError bool `yaml:"continue_on_error,omitempty" json:"continue_on_error,omitempty"`

	// Env sets additional environment variables (shell steps only).
	Env map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

// FlowStepType is the type of a flow step.
type FlowStepType string

const (
	// FlowStepTypeShell executes a shell command.
	FlowStepTypeShell FlowStepType = "shell"

	// FlowStepTypeAgent invokes an agent.
	FlowStepTypeAgent FlowStepType = "agent"
)

// IsValid returns true if the step type is recognized.
func (t FlowStepType) IsValid() bool {
	return t == FlowStepTypeShell || t == FlowStepTypeAgent
}

// FlowTrigger defines an automatic trigger for a flow.
type FlowTrigger struct {
	// ID is the unique identifier for this trigger.
	ID string `yaml:"id" json:"id"`

	// Type is the trigger type: "cron" or "watch"
	Type FlowTriggerType `yaml:"type" json:"type"`

	// === Cron trigger fields ===

	// Schedule is the cron expression (for type: cron).
	// Supports 5-field (minute hour day month weekday) or 6-field with seconds.
	Schedule string `yaml:"schedule,omitempty" json:"schedule,omitempty"`

	// === Watch trigger fields ===

	// Path is the filesystem path to watch (for type: watch).
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// Patterns are glob patterns for files to watch (e.g., "*.md").
	Patterns []string `yaml:"patterns,omitempty" json:"patterns,omitempty"`

	// Recursive watches subdirectories.
	Recursive bool `yaml:"recursive,omitempty" json:"recursive,omitempty"`

	// Events specifies which filesystem events trigger the flow.
	// Values: "create", "modify", "delete". Default: ["create", "modify"]
	Events []string `yaml:"events,omitempty" json:"events,omitempty"`

	// === Common fields ===

	// Params are default parameters passed to the flow.
	Params map[string]any `yaml:"params,omitempty" json:"params,omitempty"`

	// RunsBeforePermanent is how many successful runs before the trigger
	// becomes permanent. If nil or 0, the trigger is already permanent.
	RunsBeforePermanent *int `yaml:"runs_before_permanent,omitempty" json:"runs_before_permanent,omitempty"`

	// Enabled controls whether the trigger is active. Default: true.
	Enabled *bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// FlowTriggerType is the type of a flow trigger.
type FlowTriggerType string

const (
	// FlowTriggerTypeCron triggers on a schedule.
	FlowTriggerTypeCron FlowTriggerType = "cron"

	// FlowTriggerTypeWatch triggers on filesystem changes.
	FlowTriggerTypeWatch FlowTriggerType = "watch"
)

// IsValid returns true if the trigger type is recognized.
func (t FlowTriggerType) IsValid() bool {
	return t == FlowTriggerTypeCron || t == FlowTriggerTypeWatch
}

// IsEnabled returns true if the trigger is enabled.
func (t FlowTrigger) IsEnabled() bool {
	if t.Enabled == nil {
		return true // Default enabled
	}
	return *t.Enabled
}

// IsPermanent returns true if the trigger is permanent (no trial period).
func (t FlowTrigger) IsPermanent() bool {
	return t.RunsBeforePermanent == nil || *t.RunsBeforePermanent <= 0
}

// StepResult contains the result of executing a single step.
type StepResult struct {
	// ID is the step identifier.
	ID string `json:"id"`

	// Status is the step execution status.
	Status StepStatus `json:"status"`

	// === Shell step results ===

	// Stdout is the captured standard output (shell steps).
	Stdout string `json:"stdout,omitempty"`

	// Stderr is the captured standard error (shell steps).
	Stderr string `json:"stderr,omitempty"`

	// ExitCode is the process exit code (shell steps).
	ExitCode int `json:"exit_code,omitempty"`

	// === Agent step results ===

	// Output is the agent response (agent steps).
	Output string `json:"output,omitempty"`

	// === Common ===

	// StartTime is when the step started.
	StartTime time.Time `json:"start_time"`

	// EndTime is when the step completed.
	EndTime time.Time `json:"end_time"`

	// Duration is how long the step took.
	Duration time.Duration `json:"duration_ns"`

	// Skipped is true if the step was skipped due to 'when' condition.
	Skipped bool `json:"skipped,omitempty"`

	// Error contains any error message.
	Error string `json:"error,omitempty"`
}

// StepStatus is the status of a step execution.
type StepStatus string

const (
	StepStatusPending  StepStatus = "pending"
	StepStatusRunning  StepStatus = "running"
	StepStatusSuccess  StepStatus = "success"
	StepStatusFailed   StepStatus = "failed"
	StepStatusSkipped  StepStatus = "skipped"
	StepStatusTimeout  StepStatus = "timeout"
)

// YAMLFlowResult contains the result of executing a YAML flow.
type YAMLFlowResult struct {
	// RunID is the unique run identifier.
	RunID string `json:"run_id"`

	// Flow is the flow that was executed.
	FlowName string `json:"flow_name"`

	// Status is the overall flow status.
	Status RunStatus `json:"status"`

	// Steps contains results for each step.
	Steps map[string]*StepResult `json:"steps"`

	// StartTime is when the flow started.
	StartTime time.Time `json:"start_time"`

	// EndTime is when the flow completed.
	EndTime time.Time `json:"end_time"`

	// Duration is the total flow duration.
	Duration time.Duration `json:"duration_ns"`

	// Error contains any top-level error message.
	Error string `json:"error,omitempty"`
}

// Validate validates a YAMLFlow for correctness.
func (f *YAMLFlow) Validate() error {
	var errs []string

	// Version check
	if f.Version != 1 {
		errs = append(errs, fmt.Sprintf("unsupported version %d, only version 1 is supported", f.Version))
	}

	// Name is required
	if f.Name == "" {
		errs = append(errs, "name is required")
	}

	// At least one step is required
	if len(f.Steps) == 0 {
		errs = append(errs, "at least one step is required")
	}

	// Validate steps
	stepIDs := make(map[string]int) // step ID -> index
	for i, step := range f.Steps {
		stepErrs := f.validateStep(step, i, stepIDs)
		errs = append(errs, stepErrs...)
		stepIDs[step.ID] = i
	}

	// Validate triggers
	triggerIDs := make(map[string]bool)
	for i, trigger := range f.Triggers {
		triggerErrs := f.validateTrigger(trigger, i, triggerIDs)
		errs = append(errs, triggerErrs...)
		triggerIDs[trigger.ID] = true
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}

	return nil
}

func (f *YAMLFlow) validateStep(step FlowStep, index int, existingSteps map[string]int) []string {
	var errs []string
	prefix := fmt.Sprintf("steps[%d]", index)

	// ID is required and must be unique
	if step.ID == "" {
		errs = append(errs, fmt.Sprintf("%s: id is required", prefix))
	} else if _, exists := existingSteps[step.ID]; exists {
		errs = append(errs, fmt.Sprintf("%s: duplicate id %q", prefix, step.ID))
	}

	// Type must be valid
	if !step.Type.IsValid() {
		errs = append(errs, fmt.Sprintf("%s: invalid type %q, must be 'shell' or 'agent'", prefix, step.Type))
	}

	// Type-specific validation
	switch step.Type {
	case FlowStepTypeShell:
		if step.Run == "" {
			errs = append(errs, fmt.Sprintf("%s: run is required for shell steps", prefix))
		}
		// Agent-specific fields should not be set
		if step.Agent != "" {
			errs = append(errs, fmt.Sprintf("%s: agent should not be set for shell steps", prefix))
		}
		if step.Prompt != "" {
			errs = append(errs, fmt.Sprintf("%s: prompt should not be set for shell steps", prefix))
		}

	case FlowStepTypeAgent:
		if step.Agent == "" {
			errs = append(errs, fmt.Sprintf("%s: agent is required for agent steps", prefix))
		}
		if step.Prompt == "" {
			errs = append(errs, fmt.Sprintf("%s: prompt is required for agent steps", prefix))
		}
		// Shell-specific fields should not be set
		if step.Run != "" {
			errs = append(errs, fmt.Sprintf("%s: run should not be set for agent steps", prefix))
		}
	}

	// Validate depends_on references
	for _, depID := range step.DependsOn {
		if _, exists := existingSteps[depID]; !exists {
			errs = append(errs, fmt.Sprintf("%s: depends_on references unknown step %q", prefix, depID))
		}
	}

	// Validate timeout format
	if step.Timeout != "" {
		if _, err := time.ParseDuration(step.Timeout); err != nil {
			errs = append(errs, fmt.Sprintf("%s: invalid timeout %q: %v", prefix, step.Timeout, err))
		}
	}

	// Validate template expressions in relevant fields
	for _, templateField := range []struct {
		name  string
		value string
	}{
		{"prompt", step.Prompt},
		{"context", step.Context},
		{"input", step.Input},
		{"run", step.Run},
		{"when", step.When},
	} {
		if templateField.value != "" {
			if templateErrs := validateTemplateReferences(templateField.value, existingSteps, prefix+"."+templateField.name); len(templateErrs) > 0 {
				errs = append(errs, templateErrs...)
			}
		}
	}

	return errs
}

func (f *YAMLFlow) validateTrigger(trigger FlowTrigger, index int, existingTriggers map[string]bool) []string {
	var errs []string
	prefix := fmt.Sprintf("triggers[%d]", index)

	// ID is required and must be unique
	if trigger.ID == "" {
		errs = append(errs, fmt.Sprintf("%s: id is required", prefix))
	} else if existingTriggers[trigger.ID] {
		errs = append(errs, fmt.Sprintf("%s: duplicate id %q", prefix, trigger.ID))
	}

	// Type must be valid
	if !trigger.Type.IsValid() {
		errs = append(errs, fmt.Sprintf("%s: invalid type %q, must be 'cron' or 'watch'", prefix, trigger.Type))
	}

	// Type-specific validation
	switch trigger.Type {
	case FlowTriggerTypeCron:
		if trigger.Schedule == "" {
			errs = append(errs, fmt.Sprintf("%s: schedule is required for cron triggers", prefix))
		} else {
			// Validate cron expression
			parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
			if _, err := parser.Parse(trigger.Schedule); err != nil {
				errs = append(errs, fmt.Sprintf("%s: invalid cron schedule %q: %v", prefix, trigger.Schedule, err))
			}
		}
		// Watch-specific fields should not be set
		if trigger.Path != "" {
			errs = append(errs, fmt.Sprintf("%s: path should not be set for cron triggers", prefix))
		}

	case FlowTriggerTypeWatch:
		if trigger.Path == "" {
			errs = append(errs, fmt.Sprintf("%s: path is required for watch triggers", prefix))
		}
		// Validate events if specified
		validEvents := map[string]bool{"create": true, "modify": true, "delete": true}
		for _, event := range trigger.Events {
			if !validEvents[event] {
				errs = append(errs, fmt.Sprintf("%s: invalid event %q, must be 'create', 'modify', or 'delete'", prefix, event))
			}
		}
		// Cron-specific fields should not be set
		if trigger.Schedule != "" {
			errs = append(errs, fmt.Sprintf("%s: schedule should not be set for watch triggers", prefix))
		}
	}

	return errs
}

// templateRefRegex matches template expressions like {{ steps.ID.stdout }}
var templateRefRegex = regexp.MustCompile(`\{\{\s*steps\.([a-zA-Z0-9_-]+)\.[a-zA-Z0-9_]+\s*\}\}`)

// validateTemplateReferences checks that step references in templates are valid.
func validateTemplateReferences(template string, existingSteps map[string]int, fieldName string) []string {
	var errs []string

	matches := templateRefRegex.FindAllStringSubmatch(template, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			stepID := match[1]
			if _, exists := existingSteps[stepID]; !exists {
				errs = append(errs, fmt.Sprintf("%s: references unknown step %q", fieldName, stepID))
			}
		}
	}

	return errs
}

// ValidationError contains multiple validation errors.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0]
	}
	return fmt.Sprintf("%d validation errors:\n- %s", len(e.Errors), strings.Join(e.Errors, "\n- "))
}
