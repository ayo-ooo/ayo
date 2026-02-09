package flows

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

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
