// Package flows provides flow discovery, parsing, and execution.
// This file defines the YAML flow specification for multi-step orchestrated flows.

package flows

import (
	"time"
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
	Default    interface{}            `yaml:"default,omitempty" json:"default,omitempty"`

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
	Params map[string]interface{} `yaml:"params,omitempty" json:"params,omitempty"`

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
