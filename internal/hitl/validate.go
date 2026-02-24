package hitl

import (
	"fmt"
	"regexp"
)

// ValidationError represents a validation failure for a specific field.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("%d validation errors", len(e))
}

// ValidateRequest validates an InputRequest schema.
func ValidateRequest(req *InputRequest) error {
	var errs ValidationErrors

	if req.ID == "" {
		errs = append(errs, ValidationError{Field: "id", Message: "required"})
	}

	if len(req.Fields) == 0 {
		errs = append(errs, ValidationError{Field: "fields", Message: "at least one field required"})
	}

	fieldNames := make(map[string]bool)
	for i, field := range req.Fields {
		fieldKey := fmt.Sprintf("fields[%d]", i)

		if field.Name == "" {
			errs = append(errs, ValidationError{Field: fieldKey + ".name", Message: "required"})
		} else if fieldNames[field.Name] {
			errs = append(errs, ValidationError{Field: fieldKey + ".name", Message: "duplicate field name"})
		} else {
			fieldNames[field.Name] = true
		}

		if field.Label == "" {
			errs = append(errs, ValidationError{Field: fieldKey + ".label", Message: "required"})
		}

		if !isValidFieldType(field.Type) {
			errs = append(errs, ValidationError{Field: fieldKey + ".type", Message: "invalid field type"})
		}

		if (field.Type == FieldTypeSelect || field.Type == FieldTypeMultiselect) && len(field.Options) == 0 {
			errs = append(errs, ValidationError{Field: fieldKey + ".options", Message: "options required for select/multiselect"})
		}

		if field.Validation != nil && field.Validation.Pattern != nil {
			if _, err := regexp.Compile(*field.Validation.Pattern); err != nil {
				errs = append(errs, ValidationError{Field: fieldKey + ".validation.pattern", Message: "invalid regex pattern"})
			}
		}
	}

	switch req.Recipient.Type {
	case RecipientOwner, RecipientChat:
		// Valid, no address required
	case RecipientEmail:
		if req.Recipient.Address == "" {
			errs = append(errs, ValidationError{Field: "recipient.address", Message: "email address required"})
		}
	case "":
		errs = append(errs, ValidationError{Field: "recipient.type", Message: "required"})
	default:
		errs = append(errs, ValidationError{Field: "recipient.type", Message: "invalid recipient type"})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// ValidateResponse validates an InputResponse against its request schema.
func ValidateResponse(req *InputRequest, resp *InputResponse) error {
	var errs ValidationErrors

	if resp.RequestID != req.ID {
		errs = append(errs, ValidationError{Field: "request_id", Message: "does not match request"})
	}

	// Check required fields
	for _, field := range req.Fields {
		value, exists := resp.Values[field.Name]
		if field.Required {
			if !exists || value == nil || value == "" {
				errs = append(errs, ValidationError{Field: field.Name, Message: "required"})
				continue
			}
		}

		if !exists || value == nil {
			continue
		}

		// Type-specific validation
		if err := validateFieldValue(field, value); err != nil {
			errs = append(errs, *err)
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validateFieldValue(field Field, value any) *ValidationError {
	switch field.Type {
	case FieldTypeText, FieldTypeTextarea:
		str, ok := value.(string)
		if !ok {
			return &ValidationError{Field: field.Name, Message: "must be a string"}
		}
		if field.Validation != nil && str != "" {
			if field.Validation.MinLength != nil && len(str) < *field.Validation.MinLength {
				return &ValidationError{Field: field.Name, Message: fmt.Sprintf("minimum length is %d", *field.Validation.MinLength)}
			}
			if field.Validation.MaxLength != nil && len(str) > *field.Validation.MaxLength {
				return &ValidationError{Field: field.Name, Message: fmt.Sprintf("maximum length is %d", *field.Validation.MaxLength)}
			}
			if field.Validation.Pattern != nil {
				re, err := regexp.Compile(*field.Validation.Pattern)
				if err != nil {
					return &ValidationError{Field: field.Name, Message: "invalid pattern"}
				}
				if !re.MatchString(str) {
					return &ValidationError{Field: field.Name, Message: "does not match pattern"}
				}
			}
		}

	case FieldTypeNumber:
		var num float64
		switch v := value.(type) {
		case float64:
			num = v
		case int:
			num = float64(v)
		case int64:
			num = float64(v)
		default:
			return &ValidationError{Field: field.Name, Message: "must be a number"}
		}
		if field.Validation != nil {
			if field.Validation.Min != nil && num < float64(*field.Validation.Min) {
				return &ValidationError{Field: field.Name, Message: fmt.Sprintf("minimum value is %d", *field.Validation.Min)}
			}
			if field.Validation.Max != nil && num > float64(*field.Validation.Max) {
				return &ValidationError{Field: field.Name, Message: fmt.Sprintf("maximum value is %d", *field.Validation.Max)}
			}
		}

	case FieldTypeConfirm:
		if _, ok := value.(bool); !ok {
			return &ValidationError{Field: field.Name, Message: "must be a boolean"}
		}

	case FieldTypeSelect:
		str, ok := value.(string)
		if !ok {
			return &ValidationError{Field: field.Name, Message: "must be a string"}
		}
		if !isValidOption(field.Options, str) {
			return &ValidationError{Field: field.Name, Message: "invalid option"}
		}

	case FieldTypeMultiselect:
		arr, ok := value.([]any)
		if !ok {
			return &ValidationError{Field: field.Name, Message: "must be an array"}
		}
		for _, v := range arr {
			str, ok := v.(string)
			if !ok {
				return &ValidationError{Field: field.Name, Message: "array values must be strings"}
			}
			if !isValidOption(field.Options, str) {
				return &ValidationError{Field: field.Name, Message: fmt.Sprintf("invalid option: %s", str)}
			}
		}
	}

	return nil
}

func isValidFieldType(t FieldType) bool {
	switch t {
	case FieldTypeText, FieldTypeTextarea, FieldTypeSelect, FieldTypeMultiselect,
		FieldTypeConfirm, FieldTypeNumber, FieldTypeDate, FieldTypeFile:
		return true
	}
	return false
}

func isValidOption(options []Option, value string) bool {
	for _, opt := range options {
		if opt.Value == value {
			return true
		}
	}
	return false
}

// PromptFunc is called to get input from the user.
// It receives the field to prompt for and any previous validation error.
// Returns the value or an error if prompting failed.
type PromptFunc func(field Field, previousError *ValidationError) (any, error)

// Validator handles input validation with retry support.
type Validator struct {
	maxRetries int
}

// NewValidator creates a new validator with the given max retries.
func NewValidator(maxRetries int) *Validator {
	if maxRetries <= 0 {
		maxRetries = DefaultMaxRetries
	}
	return &Validator{maxRetries: maxRetries}
}

// MaxRetriesError is returned when validation fails after all retries are exhausted.
type MaxRetriesError struct {
	Field      string
	Attempts   int
	LastError  *ValidationError
}

func (e *MaxRetriesError) Error() string {
	return fmt.Sprintf("validation failed after %d attempts for field %q: %s", e.Attempts, e.Field, e.LastError.Message)
}

// CollectWithRetry collects input for all fields, retrying invalid fields up to maxRetries.
// It preserves valid values and only re-prompts for invalid fields.
func (v *Validator) CollectWithRetry(req *InputRequest, promptFn PromptFunc) (*InputResponse, error) {
	values := make(map[string]any)
	
	// Initialize with defaults
	for _, field := range req.Fields {
		if field.Default != nil {
			values[field.Name] = field.Default
		}
	}
	
	maxRetries := v.maxRetries
	if req.MaxRetries > 0 {
		maxRetries = req.MaxRetries
	}
	
	for _, field := range req.Fields {
		var lastError *ValidationError
		
		for attempt := 0; attempt <= maxRetries; attempt++ {
			value, err := promptFn(field, lastError)
			if err != nil {
				return nil, fmt.Errorf("prompt failed for field %q: %w", field.Name, err)
			}
			
			values[field.Name] = value
			
			// Validate the field
			if validErr := validateFieldValue(field, value); validErr != nil {
				// Use custom message if provided
				if field.Validation != nil && field.Validation.Message != "" {
					validErr.Message = field.Validation.Message
				}
				lastError = validErr
				continue
			}
			
			// Check required
			if field.Required && isEmpty(value) {
				lastError = &ValidationError{
					Field:   field.Name,
					Message: fmt.Sprintf("%s is required", field.Label),
				}
				continue
			}
			
			// Valid - move to next field
			lastError = nil
			break
		}
		
		if lastError != nil {
			return nil, &MaxRetriesError{
				Field:     field.Name,
				Attempts:  maxRetries + 1,
				LastError: lastError,
			}
		}
	}
	
	return &InputResponse{
		RequestID: req.ID,
		Values:    values,
	}, nil
}

func isEmpty(value any) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return v == ""
	case []any:
		return len(v) == 0
	}
	return false
}
