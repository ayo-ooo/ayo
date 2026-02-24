package hitl

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/huh"
)

// CLIFormRenderer renders InputRequest schemas as interactive TUI forms.
type CLIFormRenderer struct {
	accessible bool
}

// NewCLIFormRenderer creates a new CLI form renderer.
func NewCLIFormRenderer() *CLIFormRenderer {
	return &CLIFormRenderer{}
}

// SetAccessible enables accessible mode (simpler prompts).
func (r *CLIFormRenderer) SetAccessible(accessible bool) {
	r.accessible = accessible
}

// Render displays the form and blocks until complete or cancelled.
func (r *CLIFormRenderer) Render(ctx context.Context, req *InputRequest) (*InputResponse, error) {
	values := make(map[string]any)
	var groups []*huh.Group

	// Create form fields
	var fields []huh.Field
	for _, field := range req.Fields {
		f, valuePtr := r.createField(field)
		if f != nil {
			fields = append(fields, f)
			values[field.Name] = valuePtr
		}
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("no valid fields in request")
	}

	groups = append(groups, huh.NewGroup(fields...))

	form := huh.NewForm(groups...).
		WithAccessible(r.accessible)

	// Run form with context cancellation
	errCh := make(chan error, 1)
	go func() {
		errCh <- form.Run()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("form cancelled or failed: %w", err)
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Extract values from pointers
	result := make(map[string]any)
	for _, field := range req.Fields {
		ptr, ok := values[field.Name]
		if !ok {
			continue
		}
		result[field.Name] = extractValue(ptr)
	}

	return &InputResponse{
		RequestID: req.ID,
		Values:    result,
		Timestamp: time.Now(),
	}, nil
}

// createField creates a huh field from a schema field.
func (r *CLIFormRenderer) createField(field Field) (huh.Field, any) {
	switch field.Type {
	case FieldTypeText:
		var value string
		if field.Default != nil {
			if s, ok := field.Default.(string); ok {
				value = s
			}
		}
		f := huh.NewInput().
			Title(field.Label).
			Description(field.Description).
			Value(&value)
		if field.Validation != nil {
			f = f.Validate(r.textValidator(field))
		}
		return f, &value

	case FieldTypeTextarea:
		var value string
		if field.Default != nil {
			if s, ok := field.Default.(string); ok {
				value = s
			}
		}
		f := huh.NewText().
			Title(field.Label).
			Description(field.Description).
			Value(&value)
		if field.Validation != nil {
			f = f.Validate(r.textValidator(field))
		}
		return f, &value

	case FieldTypeNumber:
		var value string
		if field.Default != nil {
			value = fmt.Sprintf("%v", field.Default)
		}
		f := huh.NewInput().
			Title(field.Label).
			Description(field.Description).
			Value(&value).
			Validate(r.numberValidator(field))
		return f, &value

	case FieldTypeSelect:
		var value string
		if field.Default != nil {
			if s, ok := field.Default.(string); ok {
				value = s
			}
		}
		options := make([]huh.Option[string], len(field.Options))
		for i, opt := range field.Options {
			options[i] = huh.NewOption(opt.Label, opt.Value)
		}
		f := huh.NewSelect[string]().
			Title(field.Label).
			Description(field.Description).
			Options(options...).
			Value(&value)
		return f, &value

	case FieldTypeMultiselect:
		var value []string
		if field.Default != nil {
			if arr, ok := field.Default.([]any); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						value = append(value, s)
					}
				}
			}
		}
		options := make([]huh.Option[string], len(field.Options))
		for i, opt := range field.Options {
			options[i] = huh.NewOption(opt.Label, opt.Value)
		}
		f := huh.NewMultiSelect[string]().
			Title(field.Label).
			Description(field.Description).
			Options(options...).
			Value(&value)
		return f, &value

	case FieldTypeConfirm:
		var value bool
		if field.Default != nil {
			if b, ok := field.Default.(bool); ok {
				value = b
			}
		}
		f := huh.NewConfirm().
			Title(field.Label).
			Description(field.Description).
			Value(&value)
		return f, &value

	case FieldTypeDate:
		var value string
		if field.Default != nil {
			if s, ok := field.Default.(string); ok {
				value = s
			}
		}
		f := huh.NewInput().
			Title(field.Label).
			Description(field.Description + " (YYYY-MM-DD)").
			Placeholder("2006-01-02").
			Value(&value).
			Validate(r.dateValidator(field))
		return f, &value

	case FieldTypeFile:
		var value string
		f := huh.NewFilePicker().
			Title(field.Label).
			Description(field.Description).
			Picking(true).
			Value(&value)
		return f, &value

	default:
		// Fallback to text input
		var value string
		f := huh.NewInput().
			Title(field.Label).
			Description(field.Description).
			Value(&value)
		return f, &value
	}
}

// textValidator creates a text validation function.
func (r *CLIFormRenderer) textValidator(field Field) func(string) error {
	return func(value string) error {
		if field.Required && value == "" {
			return fmt.Errorf("%s is required", field.Label)
		}
		if field.Validation == nil || value == "" {
			return nil
		}
		if field.Validation.MinLength != nil && len(value) < *field.Validation.MinLength {
			return fmt.Errorf("minimum length is %d", *field.Validation.MinLength)
		}
		if field.Validation.MaxLength != nil && len(value) > *field.Validation.MaxLength {
			return fmt.Errorf("maximum length is %d", *field.Validation.MaxLength)
		}
		return nil
	}
}

// numberValidator creates a number validation function.
func (r *CLIFormRenderer) numberValidator(field Field) func(string) error {
	return func(value string) error {
		if field.Required && value == "" {
			return fmt.Errorf("%s is required", field.Label)
		}
		if value == "" {
			return nil
		}
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("please enter a valid number")
		}
		if field.Validation != nil {
			if field.Validation.Min != nil && n < float64(*field.Validation.Min) {
				return fmt.Errorf("minimum value is %d", *field.Validation.Min)
			}
			if field.Validation.Max != nil && n > float64(*field.Validation.Max) {
				return fmt.Errorf("maximum value is %d", *field.Validation.Max)
			}
		}
		return nil
	}
}

// dateValidator creates a date validation function.
func (r *CLIFormRenderer) dateValidator(field Field) func(string) error {
	return func(value string) error {
		if field.Required && value == "" {
			return fmt.Errorf("%s is required", field.Label)
		}
		if value == "" {
			return nil
		}
		_, err := time.Parse("2006-01-02", value)
		if err != nil {
			return fmt.Errorf("please enter a valid date (YYYY-MM-DD)")
		}
		return nil
	}
}

// extractValue extracts the actual value from a pointer.
func extractValue(ptr any) any {
	switch v := ptr.(type) {
	case *string:
		return *v
	case *bool:
		return *v
	case *[]string:
		// Convert to []any for consistency
		result := make([]any, len(*v))
		for i, s := range *v {
			result[i] = s
		}
		return result
	default:
		return ptr
	}
}
