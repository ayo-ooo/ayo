package flows

import (
	"strings"
	"testing"
)

func TestYAMLFlowValidate(t *testing.T) {
	t.Run("valid flow", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo hello"},
			},
		}

		if err := flow.Validate(); err != nil {
			t.Errorf("expected valid flow, got error: %v", err)
		}
	})

	t.Run("valid flow with agent step", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeAgent, Agent: "@test", Prompt: "Do something"},
			},
		}

		if err := flow.Validate(); err != nil {
			t.Errorf("expected valid flow, got error: %v", err)
		}
	})

	t.Run("missing version", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 0,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo hello"},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for missing version")
		}
		if !strings.Contains(err.Error(), "unsupported version") {
			t.Errorf("expected version error, got: %v", err)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo hello"},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for missing name")
		}
		if !strings.Contains(err.Error(), "name is required") {
			t.Errorf("expected name error, got: %v", err)
		}
	})

	t.Run("no steps", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps:   []FlowStep{},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for no steps")
		}
		if !strings.Contains(err.Error(), "at least one step is required") {
			t.Errorf("expected steps error, got: %v", err)
		}
	})

	t.Run("duplicate step IDs", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1"},
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 2"},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for duplicate step IDs")
		}
		if !strings.Contains(err.Error(), "duplicate id") {
			t.Errorf("expected duplicate id error, got: %v", err)
		}
	})

	t.Run("invalid step type", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: "invalid", Run: "echo hello"},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for invalid step type")
		}
		if !strings.Contains(err.Error(), "invalid type") {
			t.Errorf("expected type error, got: %v", err)
		}
	})

	t.Run("shell step missing run", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for shell step missing run")
		}
		if !strings.Contains(err.Error(), "run is required") {
			t.Errorf("expected run required error, got: %v", err)
		}
	})

	t.Run("agent step missing agent", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeAgent, Prompt: "Do something"},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for agent step missing agent")
		}
		if !strings.Contains(err.Error(), "agent is required") {
			t.Errorf("expected agent required error, got: %v", err)
		}
	})

	t.Run("agent step missing prompt", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeAgent, Agent: "@test"},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for agent step missing prompt")
		}
		if !strings.Contains(err.Error(), "prompt is required") {
			t.Errorf("expected prompt required error, got: %v", err)
		}
	})

	t.Run("invalid depends_on reference", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1"},
				{ID: "step2", Type: FlowStepTypeShell, Run: "echo 2", DependsOn: []string{"nonexistent"}},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for invalid depends_on reference")
		}
		if !strings.Contains(err.Error(), "unknown step") {
			t.Errorf("expected unknown step error, got: %v", err)
		}
	})

	t.Run("invalid timeout format", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1", Timeout: "invalid"},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for invalid timeout")
		}
		if !strings.Contains(err.Error(), "invalid timeout") {
			t.Errorf("expected timeout error, got: %v", err)
		}
	})

	t.Run("valid timeout", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1", Timeout: "5m30s"},
			},
		}

		if err := flow.Validate(); err != nil {
			t.Errorf("expected valid flow, got error: %v", err)
		}
	})

	t.Run("valid depends_on", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1"},
				{ID: "step2", Type: FlowStepTypeShell, Run: "echo 2", DependsOn: []string{"step1"}},
			},
		}

		if err := flow.Validate(); err != nil {
			t.Errorf("expected valid flow, got error: %v", err)
		}
	})

	t.Run("template reference to unknown step", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1"},
				{ID: "step2", Type: FlowStepTypeShell, Run: "echo {{ steps.unknown.stdout }}"},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for template reference to unknown step")
		}
		if !strings.Contains(err.Error(), "unknown step") {
			t.Errorf("expected unknown step error, got: %v", err)
		}
	})

	t.Run("valid template reference", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo hello"},
				{ID: "step2", Type: FlowStepTypeShell, Run: "echo {{ steps.step1.stdout }}"},
			},
		}

		if err := flow.Validate(); err != nil {
			t.Errorf("expected valid flow, got error: %v", err)
		}
	})
}

func TestYAMLFlowValidateTriggers(t *testing.T) {
	t.Run("valid cron trigger", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1"},
			},
			Triggers: []FlowTrigger{
				{ID: "trigger1", Type: FlowTriggerTypeCron, Schedule: "0 * * * *"},
			},
		}

		if err := flow.Validate(); err != nil {
			t.Errorf("expected valid flow, got error: %v", err)
		}
	})

	t.Run("invalid cron schedule", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1"},
			},
			Triggers: []FlowTrigger{
				{ID: "trigger1", Type: FlowTriggerTypeCron, Schedule: "invalid cron"},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for invalid cron schedule")
		}
		if !strings.Contains(err.Error(), "invalid cron schedule") {
			t.Errorf("expected cron schedule error, got: %v", err)
		}
	})

	t.Run("cron trigger missing schedule", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1"},
			},
			Triggers: []FlowTrigger{
				{ID: "trigger1", Type: FlowTriggerTypeCron},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for cron trigger missing schedule")
		}
		if !strings.Contains(err.Error(), "schedule is required") {
			t.Errorf("expected schedule required error, got: %v", err)
		}
	})

	t.Run("valid watch trigger", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1"},
			},
			Triggers: []FlowTrigger{
				{ID: "trigger1", Type: FlowTriggerTypeWatch, Path: "/path/to/watch"},
			},
		}

		if err := flow.Validate(); err != nil {
			t.Errorf("expected valid flow, got error: %v", err)
		}
	})

	t.Run("watch trigger missing path", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1"},
			},
			Triggers: []FlowTrigger{
				{ID: "trigger1", Type: FlowTriggerTypeWatch},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for watch trigger missing path")
		}
		if !strings.Contains(err.Error(), "path is required") {
			t.Errorf("expected path required error, got: %v", err)
		}
	})

	t.Run("watch trigger invalid event", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1"},
			},
			Triggers: []FlowTrigger{
				{ID: "trigger1", Type: FlowTriggerTypeWatch, Path: "/path", Events: []string{"invalid"}},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for invalid event")
		}
		if !strings.Contains(err.Error(), "invalid event") {
			t.Errorf("expected invalid event error, got: %v", err)
		}
	})

	t.Run("duplicate trigger IDs", func(t *testing.T) {
		flow := &YAMLFlow{
			Version: 1,
			Name:    "test-flow",
			Steps: []FlowStep{
				{ID: "step1", Type: FlowStepTypeShell, Run: "echo 1"},
			},
			Triggers: []FlowTrigger{
				{ID: "trigger1", Type: FlowTriggerTypeCron, Schedule: "0 * * * *"},
				{ID: "trigger1", Type: FlowTriggerTypeCron, Schedule: "30 * * * *"},
			},
		}

		err := flow.Validate()
		if err == nil {
			t.Fatal("expected error for duplicate trigger IDs")
		}
		if !strings.Contains(err.Error(), "duplicate id") {
			t.Errorf("expected duplicate id error, got: %v", err)
		}
	})
}

func TestValidationError(t *testing.T) {
	t.Run("single error", func(t *testing.T) {
		err := &ValidationError{Errors: []string{"only one error"}}
		if err.Error() != "only one error" {
			t.Errorf("expected 'only one error', got: %s", err.Error())
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		err := &ValidationError{Errors: []string{"error 1", "error 2", "error 3"}}
		msg := err.Error()
		if !strings.Contains(msg, "3 validation errors") {
			t.Errorf("expected '3 validation errors' in message, got: %s", msg)
		}
		if !strings.Contains(msg, "error 1") || !strings.Contains(msg, "error 2") || !strings.Contains(msg, "error 3") {
			t.Errorf("expected all errors in message, got: %s", msg)
		}
	})
}
