package hitl

import (
	"context"
	"testing"
)

func TestConversationalHandler_Run_SelectField(t *testing.T) {
	var sentMessages []string
	responseIdx := 0
	responses := []string{"2"} // Select option 2

	h := NewConversationalHandler(
		func(msg string) error {
			sentMessages = append(sentMessages, msg)
			return nil
		},
		func(ctx context.Context) (string, error) {
			if responseIdx >= len(responses) {
				return "", nil
			}
			resp := responses[responseIdx]
			responseIdx++
			return resp, nil
		},
	)

	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "choice",
				Type:  FieldTypeSelect,
				Label: "Which option",
				Options: []Option{
					{Value: "a", Label: "Option A"},
					{Value: "b", Label: "Option B"},
					{Value: "c", Label: "Option C"},
				},
			},
		},
	}

	resp, err := h.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Values["choice"] != "b" {
		t.Errorf("expected choice=b, got %v", resp.Values["choice"])
	}
	
	// Should have sent: prompt, ack, confirmation
	if len(sentMessages) != 3 {
		t.Errorf("expected 3 sent messages, got %d", len(sentMessages))
	}
}

func TestConversationalHandler_Run_MultiselectField(t *testing.T) {
	responseIdx := 0
	responses := []string{"1, 3"} // Select options 1 and 3

	h := NewConversationalHandler(
		func(msg string) error { return nil },
		func(ctx context.Context) (string, error) {
			if responseIdx >= len(responses) {
				return "", nil
			}
			resp := responses[responseIdx]
			responseIdx++
			return resp, nil
		},
	)

	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "tags",
				Type:  FieldTypeMultiselect,
				Label: "Select tags",
				Options: []Option{
					{Value: "a", Label: "Tag A"},
					{Value: "b", Label: "Tag B"},
					{Value: "c", Label: "Tag C"},
				},
			},
		},
	}

	resp, err := h.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	values, ok := resp.Values["tags"].([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", resp.Values["tags"])
	}
	if len(values) != 2 {
		t.Errorf("expected 2 values, got %d", len(values))
	}
	if values[0] != "a" || values[1] != "c" {
		t.Errorf("expected [a, c], got %v", values)
	}
}

func TestConversationalHandler_Run_ConfirmField(t *testing.T) {
	tests := []struct {
		response string
		want     bool
	}{
		{"yes", true},
		{"Yes", true},
		{"y", true},
		{"no", false},
		{"No", false},
		{"n", false},
	}

	for _, tt := range tests {
		t.Run(tt.response, func(t *testing.T) {
			responseIdx := 0
			responses := []string{tt.response}

			h := NewConversationalHandler(
				func(msg string) error { return nil },
				func(ctx context.Context) (string, error) {
					if responseIdx >= len(responses) {
						return "", nil
					}
					resp := responses[responseIdx]
					responseIdx++
					return resp, nil
				},
			)

			req := &InputRequest{
				ID: "req-123",
				Fields: []Field{
					{
						Name:  "confirm",
						Type:  FieldTypeConfirm,
						Label: "Are you sure",
					},
				},
			}

			resp, err := h.Run(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Values["confirm"] != tt.want {
				t.Errorf("expected %v, got %v", tt.want, resp.Values["confirm"])
			}
		})
	}
}

func TestConversationalHandler_Run_NumberField(t *testing.T) {
	responseIdx := 0
	responses := []string{"42"}

	h := NewConversationalHandler(
		func(msg string) error { return nil },
		func(ctx context.Context) (string, error) {
			if responseIdx >= len(responses) {
				return "", nil
			}
			resp := responses[responseIdx]
			responseIdx++
			return resp, nil
		},
	)

	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "count",
				Type:  FieldTypeNumber,
				Label: "Enter count",
			},
		},
	}

	resp, err := h.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Values["count"] != float64(42) {
		t.Errorf("expected 42, got %v", resp.Values["count"])
	}
}

func TestConversationalHandler_Run_TextField(t *testing.T) {
	responseIdx := 0
	responses := []string{"Hello World"}

	h := NewConversationalHandler(
		func(msg string) error { return nil },
		func(ctx context.Context) (string, error) {
			if responseIdx >= len(responses) {
				return "", nil
			}
			resp := responses[responseIdx]
			responseIdx++
			return resp, nil
		},
	)

	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "message",
				Type:  FieldTypeText,
				Label: "Enter message",
			},
		},
	}

	resp, err := h.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Values["message"] != "Hello World" {
		t.Errorf("expected 'Hello World', got %v", resp.Values["message"])
	}
}

func TestConversationalHandler_Run_SkipOptionalField(t *testing.T) {
	responseIdx := 0
	responses := []string{"skip"}

	h := NewConversationalHandler(
		func(msg string) error { return nil },
		func(ctx context.Context) (string, error) {
			if responseIdx >= len(responses) {
				return "", nil
			}
			resp := responses[responseIdx]
			responseIdx++
			return resp, nil
		},
	)

	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:     "notes",
				Type:     FieldTypeText,
				Label:    "Any notes?",
				Required: false,
			},
		},
	}

	resp, err := h.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, exists := resp.Values["notes"]; exists {
		t.Error("expected notes to be skipped")
	}
}

func TestConversationalHandler_Run_ValidationRetry(t *testing.T) {
	var sentMessages []string
	responseIdx := 0
	responses := []string{"invalid", "2"} // First invalid, then valid

	h := NewConversationalHandler(
		func(msg string) error {
			sentMessages = append(sentMessages, msg)
			return nil
		},
		func(ctx context.Context) (string, error) {
			if responseIdx >= len(responses) {
				return "", nil
			}
			resp := responses[responseIdx]
			responseIdx++
			return resp, nil
		},
	)

	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{
				Name:  "choice",
				Type:  FieldTypeSelect,
				Label: "Pick one",
				Options: []Option{
					{Value: "a", Label: "A"},
					{Value: "b", Label: "B"},
				},
			},
		},
	}

	resp, err := h.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Values["choice"] != "b" {
		t.Errorf("expected choice=b, got %v", resp.Values["choice"])
	}
	
	// Should have error message in retry prompt
	foundError := false
	for _, msg := range sentMessages {
		if containsHelper(msg, "❌") {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("expected error message in retry")
	}
}

func TestConversationalHandler_Run_WithContext(t *testing.T) {
	var sentMessages []string
	responseIdx := 0
	responses := []string{"yes"}

	h := NewConversationalHandler(
		func(msg string) error {
			sentMessages = append(sentMessages, msg)
			return nil
		},
		func(ctx context.Context) (string, error) {
			if responseIdx >= len(responses) {
				return "", nil
			}
			resp := responses[responseIdx]
			responseIdx++
			return resp, nil
		},
	)

	req := &InputRequest{
		ID:      "req-123",
		Context: "I need your help with something important.",
		Fields: []Field{
			{
				Name:  "confirm",
				Type:  FieldTypeConfirm,
				Label: "Continue",
			},
		},
	}

	_, err := h.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// First message should be the context
	if len(sentMessages) == 0 || sentMessages[0] != "I need your help with something important." {
		t.Error("expected context to be sent first")
	}
}

func TestConversationalHandler_Run_MultipleFields(t *testing.T) {
	responseIdx := 0
	responses := []string{"Alice", "25", "yes"}

	h := NewConversationalHandler(
		func(msg string) error { return nil },
		func(ctx context.Context) (string, error) {
			if responseIdx >= len(responses) {
				return "", nil
			}
			resp := responses[responseIdx]
			responseIdx++
			return resp, nil
		},
	)

	req := &InputRequest{
		ID: "req-123",
		Fields: []Field{
			{Name: "name", Type: FieldTypeText, Label: "Name", Required: true},
			{Name: "age", Type: FieldTypeNumber, Label: "Age"},
			{Name: "confirm", Type: FieldTypeConfirm, Label: "All correct?"},
		},
	}

	resp, err := h.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Values["name"] != "Alice" {
		t.Errorf("expected name=Alice, got %v", resp.Values["name"])
	}
	if resp.Values["age"] != float64(25) {
		t.Errorf("expected age=25, got %v", resp.Values["age"])
	}
	if resp.Values["confirm"] != true {
		t.Errorf("expected confirm=true, got %v", resp.Values["confirm"])
	}
}

func TestFormatFieldPrompt_Select(t *testing.T) {
	h := NewConversationalHandler(nil, nil)

	field := Field{
		Name:  "choice",
		Type:  FieldTypeSelect,
		Label: "Choose an option",
		Options: []Option{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
		},
	}

	prompt := h.formatFieldPrompt(field)

	if !containsHelper(prompt, "1️⃣") {
		t.Error("expected emoji numbers")
	}
	if !containsHelper(prompt, "Option A") {
		t.Error("expected option labels")
	}
	if !containsHelper(prompt, "Reply with 1-2") {
		t.Error("expected instruction")
	}
}

func TestFormatFieldPrompt_Optional(t *testing.T) {
	h := NewConversationalHandler(nil, nil)

	field := Field{
		Name:     "notes",
		Type:     FieldTypeText,
		Label:    "Notes",
		Required: false,
	}

	prompt := h.formatFieldPrompt(field)

	if !containsHelper(prompt, "skip") {
		t.Error("expected skip option for optional field")
	}
}

func TestNumberEmoji(t *testing.T) {
	if numberEmoji(1) != "1️⃣" {
		t.Error("expected 1️⃣")
	}
	if numberEmoji(9) != "9️⃣" {
		t.Error("expected 9️⃣")
	}
	if numberEmoji(10) != "10." {
		t.Error("expected fallback for 10")
	}
}

func TestIsSkipResponse(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"skip", true},
		{"Skip", true},
		{"SKIP", true},
		{"s", true},
		{"-", true},
		{"", true},
		{"actual value", false},
		{"skipping", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isSkipResponse(tt.input)
			if got != tt.want {
				t.Errorf("isSkipResponse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
