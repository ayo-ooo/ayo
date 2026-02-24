package flows

import (
	"strings"
	"testing"
)

func TestResolveTemplate(t *testing.T) {
	ctx := TemplateContext{
		Steps: map[string]StepResult{
			"step1": {
				ID:       "step1",
				Status:   StepStatusSuccess,
				Stdout:   "hello world",
				Stderr:   "some warning",
				ExitCode: 0,
				Output:   "formatted output",
			},
		},
		Params: map[string]any{
			"name":    "test",
			"count":   42.0,
			"enabled": true,
		},
		Env: map[string]string{
			"HOME": "/home/user",
			"PATH": "/usr/bin",
		},
	}

	tests := []struct {
		name     string
		template string
		expected string
		wantErr  bool
	}{
		{
			name:     "no templates",
			template: "plain text",
			expected: "plain text",
		},
		{
			name:     "step stdout",
			template: "Output: {{ steps.step1.stdout }}",
			expected: "Output: hello world",
		},
		{
			name:     "step stderr",
			template: "Error: {{ steps.step1.stderr }}",
			expected: "Error: some warning",
		},
		{
			name:     "step exit_code",
			template: "Exit: {{ steps.step1.exit_code }}",
			expected: "Exit: 0",
		},
		{
			name:     "step output",
			template: "Result: {{ steps.step1.output }}",
			expected: "Result: formatted output",
		},
		{
			name:     "param string",
			template: "Name: {{ params.name }}",
			expected: "Name: test",
		},
		{
			name:     "param number",
			template: "Count: {{ params.count }}",
			expected: "Count: 42",
		},
		{
			name:     "param boolean",
			template: "Enabled: {{ params.enabled }}",
			expected: "Enabled: true",
		},
		{
			name:     "env var",
			template: "Home: {{ env.HOME }}",
			expected: "Home: /home/user",
		},
		{
			name:     "multiple templates",
			template: "{{ params.name }} wrote {{ steps.step1.stdout }} to {{ env.HOME }}",
			expected: "test wrote hello world to /home/user",
		},
		{
			name:     "fallback with primary",
			template: "{{ steps.step1.stdout // \"default\" }}",
			expected: "hello world",
		},
		{
			name:     "fallback without primary",
			template: "{{ steps.nonexistent.stdout // \"default value\" }}",
			expected: "default value",
		},
		{
			name:     "missing param returns empty",
			template: "Value: {{ params.missing }}",
			expected: "Value: ",
		},
		{
			name:     "missing env returns empty",
			template: "Value: {{ env.MISSING }}",
			expected: "Value: ",
		},
		{
			name:     "unknown step",
			template: "{{ steps.unknown.stdout }}",
			wantErr:  true,
		},
		{
			name:     "unknown reference type",
			template: "{{ invalid.type }}",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveTemplate(tt.template, ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestResolveExpression(t *testing.T) {
	ctx := TemplateContext{
		Steps: map[string]StepResult{
			"step1": {ID: "step1", Stdout: "output"},
		},
		Params: map[string]any{
			"list": []string{"a", "b", "c"},
			"obj":  map[string]any{"key": "value"},
		},
		Env: map[string]string{
			"VAR": "value",
		},
	}

	t.Run("complex param types are JSON encoded", func(t *testing.T) {
		result, err := resolveExpression("params.list", ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != `["a","b","c"]` {
			t.Errorf("expected JSON array, got %q", result)
		}
	})

	t.Run("object param is JSON encoded", func(t *testing.T) {
		result, err := resolveExpression("params.obj", ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != `{"key":"value"}` {
			t.Errorf("expected JSON object, got %q", result)
		}
	})

	t.Run("quoted string literals", func(t *testing.T) {
		result, err := resolveExpression(`"literal value"`, ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "literal value" {
			t.Errorf("expected 'literal value', got %q", result)
		}
	})

	t.Run("single quoted string literals", func(t *testing.T) {
		result, err := resolveExpression(`'single quoted'`, ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "single quoted" {
			t.Errorf("expected 'single quoted', got %q", result)
		}
	})

	t.Run("invalid step reference format", func(t *testing.T) {
		_, err := resolveExpression("steps.only_id", ctx)
		if err == nil {
			t.Error("expected error for invalid step reference")
		}
	})

	t.Run("invalid params reference format", func(t *testing.T) {
		_, err := resolveExpression("params", ctx)
		if err == nil {
			t.Error("expected error for invalid params reference")
		}
	})

	t.Run("invalid env reference format", func(t *testing.T) {
		_, err := resolveExpression("env", ctx)
		if err == nil {
			t.Error("expected error for invalid env reference")
		}
	})

	t.Run("unknown step field", func(t *testing.T) {
		_, err := resolveExpression("steps.step1.unknown", ctx)
		if err == nil {
			t.Error("expected error for unknown step field")
		}
	})
}

func TestValidateTemplateExpressions(t *testing.T) {
	availableSteps := []string{"step1", "step2"}

	t.Run("valid step reference", func(t *testing.T) {
		errs := ValidateTemplateExpressions("{{ steps.step1.stdout }}", availableSteps)
		if len(errs) > 0 {
			t.Errorf("expected no errors, got: %v", errs)
		}
	})

	t.Run("invalid step reference", func(t *testing.T) {
		errs := ValidateTemplateExpressions("{{ steps.unknown.stdout }}", availableSteps)
		if len(errs) == 0 {
			t.Error("expected error for unknown step")
		}
	})

	t.Run("params and env are always valid", func(t *testing.T) {
		errs := ValidateTemplateExpressions("{{ params.anything }} {{ env.ANY_VAR }}", availableSteps)
		if len(errs) > 0 {
			t.Errorf("expected no errors for params/env, got: %v", errs)
		}
	})

	t.Run("unknown reference type", func(t *testing.T) {
		errs := ValidateTemplateExpressions("{{ invalid.type }}", availableSteps)
		if len(errs) == 0 {
			t.Error("expected error for unknown reference type")
		}
	})

	t.Run("fallback with valid step", func(t *testing.T) {
		errs := ValidateTemplateExpressions("{{ steps.step1.stdout // \"fallback\" }}", availableSteps)
		if len(errs) > 0 {
			t.Errorf("expected no errors, got: %v", errs)
		}
	})

	t.Run("fallback with invalid step", func(t *testing.T) {
		errs := ValidateTemplateExpressions("{{ steps.unknown.stdout // \"fallback\" }}", availableSteps)
		if len(errs) == 0 {
			t.Error("expected error for unknown step in fallback")
		}
	})

	t.Run("multiple templates", func(t *testing.T) {
		errs := ValidateTemplateExpressions("{{ steps.step1.stdout }} and {{ steps.unknown.stdout }}", availableSteps)
		if len(errs) != 1 {
			t.Errorf("expected 1 error, got %d", len(errs))
		}
		if len(errs) > 0 && !strings.Contains(errs[0].Error(), "unknown") {
			t.Errorf("expected unknown step error, got: %v", errs[0])
		}
	})

	t.Run("no templates", func(t *testing.T) {
		errs := ValidateTemplateExpressions("plain text with no templates", availableSteps)
		if len(errs) > 0 {
			t.Errorf("expected no errors for plain text, got: %v", errs)
		}
	})
}
