package hitl

import (
	"testing"
	"time"
)

func TestValidateRequest_Valid(t *testing.T) {
	req := &InputRequest{
		ID:      "req-123",
		Timeout: 5 * time.Minute,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{
				Name:     "decision",
				Type:     FieldTypeSelect,
				Label:    "Choose an option",
				Required: true,
				Options: []Option{
					{Value: "a", Label: "Option A"},
					{Value: "b", Label: "Option B"},
				},
			},
		},
	}

	if err := ValidateRequest(req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateRequest_MissingID(t *testing.T) {
	req := &InputRequest{
		Timeout: time.Minute,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "q", Type: FieldTypeText, Label: "Question"},
		},
	}

	err := ValidateRequest(req)
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
	errs := err.(ValidationErrors)
	if len(errs) != 1 || errs[0].Field != "id" {
		t.Errorf("unexpected errors: %v", errs)
	}
}

func TestValidateRequest_MissingFields(t *testing.T) {
	req := &InputRequest{
		ID:      "req-123",
		Timeout: time.Minute,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{},
	}

	err := ValidateRequest(req)
	if err == nil {
		t.Fatal("expected error for missing fields")
	}
}

func TestValidateRequest_DuplicateFieldNames(t *testing.T) {
	req := &InputRequest{
		ID:      "req-123",
		Timeout: time.Minute,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "q", Type: FieldTypeText, Label: "Question 1"},
			{Name: "q", Type: FieldTypeText, Label: "Question 2"},
		},
	}

	err := ValidateRequest(req)
	if err == nil {
		t.Fatal("expected error for duplicate field names")
	}
	errs := err.(ValidationErrors)
	found := false
	for _, e := range errs {
		if e.Message == "duplicate field name" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected duplicate field name error, got: %v", errs)
	}
}

func TestValidateRequest_SelectWithoutOptions(t *testing.T) {
	req := &InputRequest{
		ID:      "req-123",
		Timeout: time.Minute,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "choice", Type: FieldTypeSelect, Label: "Choose"},
		},
	}

	err := ValidateRequest(req)
	if err == nil {
		t.Fatal("expected error for select without options")
	}
}

func TestValidateRequest_InvalidPattern(t *testing.T) {
	invalidPattern := "[invalid"
	req := &InputRequest{
		ID:      "req-123",
		Timeout: time.Minute,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{
				Name:  "email",
				Type:  FieldTypeText,
				Label: "Email",
				Validation: &Validation{
					Pattern: &invalidPattern,
				},
			},
		},
	}

	err := ValidateRequest(req)
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestValidateRequest_EmailRequiresAddress(t *testing.T) {
	req := &InputRequest{
		ID:      "req-123",
		Timeout: time.Minute,
		Recipient: Recipient{
			Type: RecipientEmail,
		},
		Fields: []Field{
			{Name: "q", Type: FieldTypeText, Label: "Question"},
		},
	}

	err := ValidateRequest(req)
	if err == nil {
		t.Fatal("expected error for email without address")
	}
}

func TestValidateResponse_Valid(t *testing.T) {
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{Name: "name", Type: FieldTypeText, Label: "Name", Required: true},
			{Name: "age", Type: FieldTypeNumber, Label: "Age"},
		},
	}

	resp := &InputResponse{
		RequestID: "req-123",
		Values: map[string]any{
			"name": "Alice",
			"age":  30,
		},
		Timestamp: time.Now(),
	}

	if err := ValidateResponse(req, resp); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateResponse_MissingRequired(t *testing.T) {
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{Name: "name", Type: FieldTypeText, Label: "Name", Required: true},
		},
	}

	resp := &InputResponse{
		RequestID: "req-123",
		Values:    map[string]any{},
		Timestamp: time.Now(),
	}

	err := ValidateResponse(req, resp)
	if err == nil {
		t.Fatal("expected error for missing required field")
	}
}

func TestValidateResponse_TextMinLength(t *testing.T) {
	minLen := 5
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "name",
				Type:  FieldTypeText,
				Label: "Name",
				Validation: &Validation{
					MinLength: &minLen,
				},
			},
		},
	}

	resp := &InputResponse{
		RequestID: "req-123",
		Values: map[string]any{
			"name": "AB",
		},
		Timestamp: time.Now(),
	}

	err := ValidateResponse(req, resp)
	if err == nil {
		t.Fatal("expected error for text too short")
	}
}

func TestValidateResponse_NumberBounds(t *testing.T) {
	min, max := 1, 100
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "score",
				Type:  FieldTypeNumber,
				Label: "Score",
				Validation: &Validation{
					Min: &min,
					Max: &max,
				},
			},
		},
	}

	tests := []struct {
		name      string
		value     any
		wantError bool
	}{
		{"valid", 50, false},
		{"at min", 1, false},
		{"at max", 100, false},
		{"below min", 0, true},
		{"above max", 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &InputResponse{
				RequestID: "req-123",
				Values:    map[string]any{"score": tt.value},
				Timestamp: time.Now(),
			}
			err := ValidateResponse(req, resp)
			if (err != nil) != tt.wantError {
				t.Errorf("got error %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateResponse_SelectInvalidOption(t *testing.T) {
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "choice",
				Type:  FieldTypeSelect,
				Label: "Choice",
				Options: []Option{
					{Value: "a", Label: "A"},
					{Value: "b", Label: "B"},
				},
			},
		},
	}

	resp := &InputResponse{
		RequestID: "req-123",
		Values: map[string]any{
			"choice": "c",
		},
		Timestamp: time.Now(),
	}

	err := ValidateResponse(req, resp)
	if err == nil {
		t.Fatal("expected error for invalid select option")
	}
}

func TestValidateResponse_Multiselect(t *testing.T) {
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "tags",
				Type:  FieldTypeMultiselect,
				Label: "Tags",
				Options: []Option{
					{Value: "a", Label: "A"},
					{Value: "b", Label: "B"},
					{Value: "c", Label: "C"},
				},
			},
		},
	}

	// Valid response
	resp := &InputResponse{
		RequestID: "req-123",
		Values: map[string]any{
			"tags": []any{"a", "c"},
		},
		Timestamp: time.Now(),
	}

	if err := ValidateResponse(req, resp); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Invalid option in array
	resp.Values["tags"] = []any{"a", "d"}
	if err := ValidateResponse(req, resp); err == nil {
		t.Fatal("expected error for invalid option in multiselect")
	}
}

func TestValidateResponse_Confirm(t *testing.T) {
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{Name: "agree", Type: FieldTypeConfirm, Label: "Agree?"},
		},
	}

	// Valid boolean
	resp := &InputResponse{
		RequestID: "req-123",
		Values:    map[string]any{"agree": true},
		Timestamp: time.Now(),
	}

	if err := ValidateResponse(req, resp); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Invalid type
	resp.Values["agree"] = "yes"
	if err := ValidateResponse(req, resp); err == nil {
		t.Fatal("expected error for non-boolean confirm")
	}
}

func TestValidateResponse_Pattern(t *testing.T) {
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "email",
				Type:  FieldTypeText,
				Label: "Email",
				Validation: &Validation{
					Pattern: &emailPattern,
				},
			},
		},
	}

	tests := []struct {
		value     string
		wantError bool
	}{
		{"user@example.com", false},
		{"invalid-email", true},
		{"", false}, // Empty is OK (not required)
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			resp := &InputResponse{
				RequestID: "req-123",
				Values:    map[string]any{"email": tt.value},
				Timestamp: time.Now(),
			}
			err := ValidateResponse(req, resp)
			if (err != nil) != tt.wantError {
				t.Errorf("value %q: got error %v, wantError %v", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestValidator_CollectWithRetry_Success(t *testing.T) {
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{Name: "name", Type: FieldTypeText, Label: "Name", Required: true},
		},
	}

	v := NewValidator(3)
	
	promptFn := func(field Field, prevErr *ValidationError) (any, error) {
		return "Alice", nil
	}
	
	resp, err := v.CollectWithRetry(req, promptFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Values["name"] != "Alice" {
		t.Errorf("expected name=Alice, got %v", resp.Values["name"])
	}
}

func TestValidator_CollectWithRetry_RetryThenSuccess(t *testing.T) {
	min := 3
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "code",
				Type:  FieldTypeText,
				Label: "Code",
				Validation: &Validation{
					MinLength: &min,
				},
			},
		},
	}

	v := NewValidator(3)
	
	attempt := 0
	promptFn := func(field Field, prevErr *ValidationError) (any, error) {
		attempt++
		if attempt == 1 {
			// First attempt - too short
			return "AB", nil
		}
		// Second attempt - valid
		return "ABCDEF", nil
	}
	
	resp, err := v.CollectWithRetry(req, promptFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Values["code"] != "ABCDEF" {
		t.Errorf("expected code=ABCDEF, got %v", resp.Values["code"])
	}
	if attempt != 2 {
		t.Errorf("expected 2 attempts, got %d", attempt)
	}
}

func TestValidator_CollectWithRetry_MaxRetriesExceeded(t *testing.T) {
	min := 10
	req := &InputRequest{
		ID:         "req-123",
		MaxRetries: 2,
		Fields: []Field{
			{
				Name:  "code",
				Type:  FieldTypeText,
				Label: "Code",
				Validation: &Validation{
					MinLength: &min,
				},
			},
		},
	}

	v := NewValidator(3) // Default is 3, but request overrides to 2
	
	promptFn := func(field Field, prevErr *ValidationError) (any, error) {
		return "short", nil // Always returns invalid value
	}
	
	_, err := v.CollectWithRetry(req, promptFn)
	if err == nil {
		t.Fatal("expected max retries error")
	}
	
	maxErr, ok := err.(*MaxRetriesError)
	if !ok {
		t.Fatalf("expected MaxRetriesError, got %T", err)
	}
	if maxErr.Attempts != 3 { // 1 initial + 2 retries
		t.Errorf("expected 3 attempts, got %d", maxErr.Attempts)
	}
}

func TestValidator_CollectWithRetry_PreviousErrorPassed(t *testing.T) {
	min := 5
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "name",
				Type:  FieldTypeText,
				Label: "Name",
				Validation: &Validation{
					MinLength: &min,
				},
			},
		},
	}

	v := NewValidator(3)
	
	var receivedErrors []*ValidationError
	attempt := 0
	promptFn := func(field Field, prevErr *ValidationError) (any, error) {
		receivedErrors = append(receivedErrors, prevErr)
		attempt++
		if attempt == 1 {
			return "AB", nil // Too short
		}
		return "ValidName", nil
	}
	
	_, err := v.CollectWithRetry(req, promptFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// First call should have no previous error
	if receivedErrors[0] != nil {
		t.Error("first call should have nil previous error")
	}
	// Second call should have the validation error
	if receivedErrors[1] == nil {
		t.Error("second call should have previous error")
	}
}

func TestValidator_CollectWithRetry_CustomMessage(t *testing.T) {
	min := 5
	customMsg := "Code must be at least 5 characters"
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "code",
				Type:  FieldTypeText,
				Label: "Code",
				Validation: &Validation{
					MinLength: &min,
					Message:   customMsg,
				},
			},
		},
	}

	v := NewValidator(2)
	
	var lastError *ValidationError
	promptFn := func(field Field, prevErr *ValidationError) (any, error) {
		lastError = prevErr
		return "AB", nil // Always invalid
	}
	
	_, err := v.CollectWithRetry(req, promptFn)
	if err == nil {
		t.Fatal("expected error")
	}
	
	// The last error passed should have the custom message
	if lastError != nil && lastError.Message != customMsg {
		t.Errorf("expected custom message %q, got %q", customMsg, lastError.Message)
	}
}

func TestValidator_CollectWithRetry_PreservesDefaults(t *testing.T) {
	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{Name: "name", Type: FieldTypeText, Label: "Name", Default: "DefaultUser"},
		},
	}

	v := NewValidator(3)
	
	promptFn := func(field Field, prevErr *ValidationError) (any, error) {
		// Return empty to use default
		return "", nil
	}
	
	resp, err := v.CollectWithRetry(req, promptFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// The response should have the value from the prompt, not the default
	// (defaults are initialization, prompt overrides)
	if resp.Values["name"] != "" {
		t.Errorf("expected empty string from prompt, got %v", resp.Values["name"])
	}
}
