package flows

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// TemplateContext provides data for template resolution.
type TemplateContext struct {
	// Steps contains results from previous steps.
	// Key is step ID, value is StepResult.
	Steps map[string]StepResult

	// Params contains flow input parameters.
	Params map[string]any

	// Env contains environment variables.
	Env map[string]string
}

// ResolveTemplate resolves template expressions in a string.
// Supports:
//   - {{ steps.ID.stdout }} - Output from a step
//   - {{ steps.ID.stderr }} - Stderr from a step
//   - {{ steps.ID.exit_code }} - Exit code from a step
//   - {{ params.NAME }} - Flow input parameter
//   - {{ env.VAR }} - Environment variable
//   - {{ A // B }} - Fallback: use B if A is empty
func ResolveTemplate(template string, ctx TemplateContext) (string, error) {
	// Match all template expressions
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	var resolveErr error
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract the expression inside {{ }}
		inner := strings.TrimSpace(match[2 : len(match)-2])

		// Handle fallback syntax: {{ A // B }}
		if parts := strings.SplitN(inner, "//", 2); len(parts) == 2 {
			primary := strings.TrimSpace(parts[0])
			fallback := strings.TrimSpace(parts[1])

			value, err := resolveExpression(primary, ctx)
			if err != nil || value == "" {
				// Use fallback
				value, err = resolveExpression(fallback, ctx)
				if err != nil {
					resolveErr = err
					return match
				}
			}
			return value
		}

		// Simple expression
		value, err := resolveExpression(inner, ctx)
		if err != nil {
			resolveErr = err
			return match
		}
		return value
	})

	return result, resolveErr
}

// resolveExpression resolves a single template expression.
func resolveExpression(expr string, ctx TemplateContext) (string, error) {
	expr = strings.TrimSpace(expr)

	// Handle quoted strings (fallback values)
	if len(expr) >= 2 && (expr[0] == '"' && expr[len(expr)-1] == '"') {
		return expr[1 : len(expr)-1], nil
	}
	if len(expr) >= 2 && (expr[0] == '\'' && expr[len(expr)-1] == '\'') {
		return expr[1 : len(expr)-1], nil
	}

	parts := strings.Split(expr, ".")
	if len(parts) == 0 {
		return "", fmt.Errorf("empty expression")
	}

	switch parts[0] {
	case "steps":
		if len(parts) < 3 {
			return "", fmt.Errorf("invalid steps reference: %s (expected steps.ID.field)", expr)
		}
		stepID := parts[1]
		field := parts[2]

		stepResult, ok := ctx.Steps[stepID]
		if !ok {
			return "", fmt.Errorf("step %q not found", stepID)
		}

		switch field {
		case "stdout":
			return stepResult.Stdout, nil
		case "stderr":
			return stepResult.Stderr, nil
		case "exit_code":
			return fmt.Sprintf("%d", stepResult.ExitCode), nil
		case "output":
			return stepResult.Output, nil
		default:
			return "", fmt.Errorf("unknown step field %q", field)
		}

	case "params":
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid params reference: %s (expected params.NAME)", expr)
		}
		paramName := parts[1]

		value, ok := ctx.Params[paramName]
		if !ok {
			return "", nil // Empty string for missing params
		}

		// Convert to string
		switch v := value.(type) {
		case string:
			return v, nil
		case bool:
			if v {
				return "true", nil
			}
			return "false", nil
		case float64:
			if v == float64(int(v)) {
				return fmt.Sprintf("%d", int(v)), nil
			}
			return fmt.Sprintf("%g", v), nil
		default:
			// JSON encode complex types
			bytes, err := json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("cannot serialize param %q: %w", paramName, err)
			}
			return string(bytes), nil
		}

	case "env":
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid env reference: %s (expected env.VAR)", expr)
		}
		varName := parts[1]

		value, ok := ctx.Env[varName]
		if !ok {
			return "", nil // Empty string for missing env vars
		}
		return value, nil

	default:
		return "", fmt.Errorf("unknown reference type %q", parts[0])
	}
}

// ValidateTemplateExpressions validates template expressions without resolving them.
// It checks that references are syntactically correct.
func ValidateTemplateExpressions(template string, availableSteps []string) []error {
	var errs []error

	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	stepSet := make(map[string]bool)
	for _, s := range availableSteps {
		stepSet[s] = true
	}

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		inner := strings.TrimSpace(match[1])

		// Handle fallback syntax
		for _, part := range strings.Split(inner, "//") {
			part = strings.TrimSpace(part)

			// Skip quoted strings
			if len(part) >= 2 && ((part[0] == '"' && part[len(part)-1] == '"') ||
				(part[0] == '\'' && part[len(part)-1] == '\'')) {
				continue
			}

			parts := strings.Split(part, ".")
			if len(parts) == 0 {
				continue
			}

			switch parts[0] {
			case "steps":
				if len(parts) >= 2 {
					stepID := parts[1]
					if !stepSet[stepID] {
						errs = append(errs, fmt.Errorf("template references unknown step %q", stepID))
					}
				}
			case "params", "env":
				// These are valid reference types, no validation needed
			default:
				errs = append(errs, fmt.Errorf("unknown reference type %q", parts[0]))
			}
		}
	}

	return errs
}
