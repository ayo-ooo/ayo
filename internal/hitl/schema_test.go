package hitl

import (
	"encoding/json"
	"testing"
	"time"
)

func TestInputRequestJSONMarshal(t *testing.T) {
	req := &InputRequest{
		ID:      "req-123",
		Timeout: 5 * time.Minute,
		Recipient: Recipient{
			Type:    RecipientEmail,
			Address: "user@example.com",
		},
		Context: "Approval needed for invoice",
		Fields: []Field{
			{
				Name:     "decision",
				Type:     FieldTypeSelect,
				Label:    "Decision",
				Required: true,
				Options: []Option{
					{Value: "approve", Label: "Approve"},
					{Value: "reject", Label: "Reject"},
				},
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed InputRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.ID != req.ID {
		t.Errorf("ID mismatch: got %q, want %q", parsed.ID, req.ID)
	}
	if parsed.Timeout != req.Timeout {
		t.Errorf("Timeout mismatch: got %v, want %v", parsed.Timeout, req.Timeout)
	}
	if parsed.Recipient.Address != req.Recipient.Address {
		t.Errorf("Recipient.Address mismatch: got %q, want %q", parsed.Recipient.Address, req.Recipient.Address)
	}
	if len(parsed.Fields) != 1 {
		t.Fatalf("Fields count mismatch: got %d, want 1", len(parsed.Fields))
	}
	if len(parsed.Fields[0].Options) != 2 {
		t.Errorf("Options count mismatch: got %d, want 2", len(parsed.Fields[0].Options))
	}
}

func TestFieldTypes(t *testing.T) {
	tests := []struct {
		fieldType FieldType
		valid     bool
	}{
		{FieldTypeText, true},
		{FieldTypeTextarea, true},
		{FieldTypeSelect, true},
		{FieldTypeMultiselect, true},
		{FieldTypeConfirm, true},
		{FieldTypeNumber, true},
		{FieldTypeDate, true},
		{FieldTypeFile, true},
		{FieldType("invalid"), false},
		{FieldType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.fieldType), func(t *testing.T) {
			got := isValidFieldType(tt.fieldType)
			if got != tt.valid {
				t.Errorf("isValidFieldType(%q) = %v, want %v", tt.fieldType, got, tt.valid)
			}
		})
	}
}

func TestInputResponseValues(t *testing.T) {
	resp := InputResponse{
		RequestID: "req-123",
		Values: map[string]any{
			"text_field":   "hello",
			"number_field": 42,
			"bool_field":   true,
			"array_field":  []any{"a", "b"},
		},
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var parsed InputResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if parsed.RequestID != resp.RequestID {
		t.Errorf("RequestID mismatch: got %q, want %q", parsed.RequestID, resp.RequestID)
	}
	if len(parsed.Values) != 4 {
		t.Errorf("Values count mismatch: got %d, want 4", len(parsed.Values))
	}
}

func TestPersonaOptional(t *testing.T) {
	req := &InputRequest{
		ID:      "req-123",
		Timeout: time.Minute,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "q", Type: FieldTypeText, Label: "Question", Required: true},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Persona should be omitted from JSON
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	if _, ok := m["persona"]; ok {
		t.Error("persona should be omitted when nil")
	}
}

func TestInputRequestWithPersona(t *testing.T) {
	req := &InputRequest{
		ID:      "req-123",
		Timeout: time.Minute,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "q", Type: FieldTypeText, Label: "Question"},
		},
		Persona: &Persona{
			Name:      "Finance Assistant",
			Signature: "Best regards, FA",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed InputRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Persona == nil {
		t.Fatal("Persona should not be nil")
	}
	if parsed.Persona.Name != "Finance Assistant" {
		t.Errorf("Persona.Name mismatch: got %q", parsed.Persona.Name)
	}
}
