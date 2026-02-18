package squads

import (
	"context"
	"testing"

	"charm.land/fantasy/schema"
)

func TestSquad_ValidateInput(t *testing.T) {
	t.Run("no schema allows any input", func(t *testing.T) {
		squad := &Squad{Name: "test", Schemas: nil}

		err := squad.ValidateInput(DispatchInput{
			Prompt: "hello",
			Data:   map[string]any{"anything": "goes"},
		})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("empty schema allows any input", func(t *testing.T) {
		squad := &Squad{Name: "test", Schemas: &SquadSchemas{}}

		err := squad.ValidateInput(DispatchInput{
			Prompt: "hello",
			Data:   map[string]any{"anything": "goes"},
		})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("prompt-only input is valid even with schema", func(t *testing.T) {
		squad := &Squad{
			Name: "test",
			Schemas: &SquadSchemas{
				Input: &schema.Schema{
					Type: "object",
					Properties: map[string]*schema.Schema{
						"code": {Type: "string"},
					},
					Required: []string{"code"},
				},
			},
		}

		err := squad.ValidateInput(DispatchInput{
			Prompt: "analyze something",
			Data:   nil,
		})
		if err != nil {
			t.Errorf("expected nil error for prompt-only input, got %v", err)
		}
	})

	t.Run("valid data passes validation", func(t *testing.T) {
		squad := &Squad{
			Name: "test",
			Schemas: &SquadSchemas{
				Input: &schema.Schema{
					Type: "object",
					Properties: map[string]*schema.Schema{
						"name": {Type: "string"},
						"count": {Type: "integer"},
					},
					Required: []string{"name"},
				},
			},
		}

		err := squad.ValidateInput(DispatchInput{
			Data: map[string]any{"name": "test", "count": 5},
		})
		if err != nil {
			t.Errorf("expected nil error for valid input, got %v", err)
		}
	})

	t.Run("missing required field fails validation", func(t *testing.T) {
		squad := &Squad{
			Name: "test",
			Schemas: &SquadSchemas{
				Input: &schema.Schema{
					Type: "object",
					Properties: map[string]*schema.Schema{
						"name": {Type: "string"},
					},
					Required: []string{"name"},
				},
			},
		}

		err := squad.ValidateInput(DispatchInput{
			Data: map[string]any{"other": "value"},
		})
		if err == nil {
			t.Error("expected validation error for missing required field")
		}

		// Check it's a ValidationError
		var valErr *ValidationError
		if !isValidationError(err, &valErr) {
			t.Errorf("expected ValidationError, got %T", err)
		} else if valErr.Direction != "input" {
			t.Errorf("expected direction 'input', got %q", valErr.Direction)
		}
	})
}

func TestSquad_ValidateOutput(t *testing.T) {
	t.Run("no schema allows any output", func(t *testing.T) {
		squad := &Squad{Name: "test", Schemas: nil}

		err := squad.ValidateOutput(&DispatchResult{
			Output: map[string]any{"anything": "goes"},
		})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("raw-only output is valid even with schema", func(t *testing.T) {
		squad := &Squad{
			Name: "test",
			Schemas: &SquadSchemas{
				Output: &schema.Schema{
					Type: "object",
					Properties: map[string]*schema.Schema{
						"result": {Type: "string"},
					},
					Required: []string{"result"},
				},
			},
		}

		err := squad.ValidateOutput(&DispatchResult{
			Raw:    "some raw output",
			Output: nil,
		})
		if err != nil {
			t.Errorf("expected nil error for raw-only output, got %v", err)
		}
	})

	t.Run("valid output passes validation", func(t *testing.T) {
		squad := &Squad{
			Name: "test",
			Schemas: &SquadSchemas{
				Output: &schema.Schema{
					Type: "object",
					Properties: map[string]*schema.Schema{
						"status": {Type: "string"},
					},
					Required: []string{"status"},
				},
			},
		}

		err := squad.ValidateOutput(&DispatchResult{
			Output: map[string]any{"status": "complete"},
		})
		if err != nil {
			t.Errorf("expected nil error for valid output, got %v", err)
		}
	})

	t.Run("invalid output fails validation", func(t *testing.T) {
		squad := &Squad{
			Name: "test",
			Schemas: &SquadSchemas{
				Output: &schema.Schema{
					Type: "object",
					Properties: map[string]*schema.Schema{
						"status": {Type: "string"},
					},
					Required: []string{"status"},
				},
			},
		}

		err := squad.ValidateOutput(&DispatchResult{
			Output: map[string]any{"other": "value"},
		})
		if err == nil {
			t.Error("expected validation error for invalid output")
		}

		var valErr *ValidationError
		if !isValidationError(err, &valErr) {
			t.Errorf("expected ValidationError, got %T", err)
		} else if valErr.Direction != "output" {
			t.Errorf("expected direction 'output', got %q", valErr.Direction)
		}
	})
}

func TestSquad_Dispatch(t *testing.T) {
	t.Run("dispatch fails if squad not ready", func(t *testing.T) {
		squad := &Squad{
			Name:      "test",
			Status:    SquadStatusStopped,
			LeadReady: false,
		}

		_, err := squad.Dispatch(context.Background(), DispatchInput{
			Prompt: "hello",
		})
		if err == nil {
			t.Error("expected error for squad not ready")
		}
	})

	t.Run("dispatch fails if validation fails", func(t *testing.T) {
		squad := &Squad{
			Name:      "test",
			Status:    SquadStatusRunning,
			LeadReady: true,
			Schemas: &SquadSchemas{
				Input: &schema.Schema{
					Type:     "object",
					Required: []string{"required_field"},
				},
			},
		}

		_, err := squad.Dispatch(context.Background(), DispatchInput{
			Data: map[string]any{"other": "value"},
		})
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("dispatch succeeds if squad ready and valid input", func(t *testing.T) {
		squad := &Squad{
			Name:      "test",
			Status:    SquadStatusRunning,
			LeadReady: true,
		}

		result, err := squad.Dispatch(context.Background(), DispatchInput{
			Prompt: "hello",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil result")
		}
		// Default routing to @ayo
		if result.RoutedTo != "@ayo" {
			t.Errorf("expected RoutedTo @ayo, got %q", result.RoutedTo)
		}
	})

	t.Run("dispatch routes to explicit target agent", func(t *testing.T) {
		squad := &Squad{
			Name:      "test",
			Status:    SquadStatusRunning,
			LeadReady: true,
		}

		result, err := squad.Dispatch(context.Background(), DispatchInput{
			Prompt:      "hello",
			TargetAgent: "@backend",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.RoutedTo != "@backend" {
			t.Errorf("expected RoutedTo @backend, got %q", result.RoutedTo)
		}
	})

	t.Run("dispatch routes to input_accepts from constitution", func(t *testing.T) {
		squad := &Squad{
			Name:      "test",
			Status:    SquadStatusRunning,
			LeadReady: true,
			Constitution: &Constitution{
				Frontmatter: ConstitutionFrontmatter{
					InputAccepts: "@frontend",
				},
			},
		}

		result, err := squad.Dispatch(context.Background(), DispatchInput{
			Prompt: "hello",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.RoutedTo != "@frontend" {
			t.Errorf("expected RoutedTo @frontend, got %q", result.RoutedTo)
		}
	})

	t.Run("explicit target overrides input_accepts", func(t *testing.T) {
		squad := &Squad{
			Name:      "test",
			Status:    SquadStatusRunning,
			LeadReady: true,
			Constitution: &Constitution{
				Frontmatter: ConstitutionFrontmatter{
					InputAccepts: "@frontend",
				},
			},
		}

		result, err := squad.Dispatch(context.Background(), DispatchInput{
			Prompt:      "hello",
			TargetAgent: "@qa",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Explicit target should override constitution
		if result.RoutedTo != "@qa" {
			t.Errorf("expected RoutedTo @qa, got %q", result.RoutedTo)
		}
	})
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Direction: "input",
		Err:       context.Canceled,
	}

	if err.Error() != "input validation failed: context canceled" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if err.Unwrap() != context.Canceled {
		t.Error("Unwrap should return wrapped error")
	}
}

// isValidationError checks if err is a *ValidationError and assigns it to target.
func isValidationError(err error, target **ValidationError) bool {
	if ve, ok := err.(*ValidationError); ok {
		*target = ve
		return true
	}
	return false
}

func TestSquad_GetTargetAgent(t *testing.T) {
	t.Run("returns default @ayo when no constitution", func(t *testing.T) {
		squad := &Squad{Name: "test"}

		target := squad.GetTargetAgent(DispatchInput{Prompt: "hello"})
		if target != "@ayo" {
			t.Errorf("expected @ayo, got %q", target)
		}
	})

	t.Run("returns input_accepts from constitution", func(t *testing.T) {
		squad := &Squad{
			Name: "test",
			Constitution: &Constitution{
				Frontmatter: ConstitutionFrontmatter{
					InputAccepts: "@backend",
				},
			},
		}

		target := squad.GetTargetAgent(DispatchInput{Prompt: "hello"})
		if target != "@backend" {
			t.Errorf("expected @backend, got %q", target)
		}
	})

	t.Run("returns lead when input_accepts not set", func(t *testing.T) {
		squad := &Squad{
			Name: "test",
			Constitution: &Constitution{
				Frontmatter: ConstitutionFrontmatter{
					Lead: "@architect",
				},
			},
		}

		target := squad.GetTargetAgent(DispatchInput{Prompt: "hello"})
		if target != "@architect" {
			t.Errorf("expected @architect (lead), got %q", target)
		}
	})

	t.Run("explicit target overrides constitution", func(t *testing.T) {
		squad := &Squad{
			Name: "test",
			Constitution: &Constitution{
				Frontmatter: ConstitutionFrontmatter{
					InputAccepts: "@backend",
				},
			},
		}

		target := squad.GetTargetAgent(DispatchInput{
			Prompt:      "hello",
			TargetAgent: "@qa",
		})
		if target != "@qa" {
			t.Errorf("expected @qa (explicit), got %q", target)
		}
	})

	t.Run("adds @ prefix to explicit target if missing", func(t *testing.T) {
		squad := &Squad{Name: "test"}

		target := squad.GetTargetAgent(DispatchInput{
			Prompt:      "hello",
			TargetAgent: "frontend",
		})
		if target != "@frontend" {
			t.Errorf("expected @frontend, got %q", target)
		}
	})
}
