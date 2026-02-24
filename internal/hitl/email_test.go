package hitl

import (
	"context"
	"strings"
	"testing"
)

func TestNewEmailHandler(t *testing.T) {
	config := EmailConfig{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		FromAddress:  "test@example.com",
		PollInterval: 0, // Should default to 30s
	}

	handler := NewEmailHandler(config)

	if handler.config.PollInterval == 0 {
		t.Error("PollInterval should have a default value")
	}
	if handler.config.IMAPFolder != "INBOX" {
		t.Errorf("IMAPFolder = %q, want INBOX", handler.config.IMAPFolder)
	}
	if len(handler.keywords) == 0 {
		t.Error("keywords should be initialized")
	}
}

func TestEmailHandler_Send(t *testing.T) {
	config := EmailConfig{
		SMTPHost:    "smtp.example.com",
		FromAddress: "bot@example.com",
	}
	handler := NewEmailHandler(config)

	req := &InputRequest{
		ID:      "test-1",
		Context: "Please approve the invoice",
		Recipient: Recipient{
			Type:    RecipientEmail,
			Address: "user@example.com",
		},
		Fields: []Field{
			{
				Name:    "approval",
				Type:    FieldTypeSelect,
				Label:   "Approval Decision",
				Options: []Option{{Value: "approve", Label: "Approve"}, {Value: "reject", Label: "Reject"}},
			},
		},
	}

	messageID, err := handler.Send(context.Background(), req)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if messageID == "" {
		t.Error("messageID should not be empty")
	}
	if handler.PendingCount() != 1 {
		t.Errorf("PendingCount = %d, want 1", handler.PendingCount())
	}
}

func TestEmailHandler_Send_InvalidRecipient(t *testing.T) {
	handler := NewEmailHandler(EmailConfig{})

	// Wrong recipient type
	req := &InputRequest{
		ID: "test-1",
		Recipient: Recipient{
			Type: RecipientOwner,
		},
	}
	_, err := handler.Send(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid recipient type")
	}

	// Missing address
	req = &InputRequest{
		ID: "test-2",
		Recipient: Recipient{
			Type:    RecipientEmail,
			Address: "",
		},
	}
	_, err = handler.Send(context.Background(), req)
	if err == nil {
		t.Error("expected error for missing address")
	}
}

func TestEmailHandler_Send_Allowlist(t *testing.T) {
	config := EmailConfig{
		AllowedRecipients: []string{"allowed@example.com"},
	}
	handler := NewEmailHandler(config)

	// Allowed address
	req := &InputRequest{
		ID: "test-1",
		Recipient: Recipient{
			Type:    RecipientEmail,
			Address: "allowed@example.com",
		},
		Fields: []Field{{Name: "test", Type: FieldTypeText}},
	}
	_, err := handler.Send(context.Background(), req)
	if err != nil {
		t.Errorf("allowed address should succeed: %v", err)
	}

	// Disallowed address
	req = &InputRequest{
		ID: "test-2",
		Recipient: Recipient{
			Type:    RecipientEmail,
			Address: "blocked@example.com",
		},
	}
	_, err = handler.Send(context.Background(), req)
	if err == nil {
		t.Error("expected error for non-allowed address")
	}
}

func TestEmailHandler_GenerateEmail(t *testing.T) {
	handler := NewEmailHandler(EmailConfig{})

	req := &InputRequest{
		ID:      "test-1",
		Context: "Please approve this invoice for $1,000",
		Persona: &Persona{
			Name:      "Finance Bot",
			Signature: "Best,\nFinance Department",
		},
		Fields: []Field{
			{
				Name:        "decision",
				Type:        FieldTypeSelect,
				Label:       "Decision",
				Description: "Choose your action",
				Required:    true,
				Options: []Option{
					{Value: "approve", Label: "Approve the payment"},
					{Value: "reject", Label: "Reject the payment"},
					{Value: "hold", Label: "Put on hold"},
				},
			},
		},
	}

	subject, body := handler.generateEmail(req)

	if subject == "" {
		t.Error("subject should not be empty")
	}
	if !strings.Contains(subject, "Finance Bot") {
		t.Errorf("subject should contain persona name: %s", subject)
	}

	if body == "" {
		t.Error("body should not be empty")
	}
	if !strings.Contains(body, "APPROVE") {
		t.Error("body should contain keyword options")
	}
	if !strings.Contains(body, "Finance Department") {
		t.Error("body should contain signature")
	}
}

func TestEmailHandler_ParseReply(t *testing.T) {
	handler := NewEmailHandler(EmailConfig{})

	tests := []struct {
		name     string
		body     string
		fields   []Field
		expected map[string]any
	}{
		{
			name: "single select field - keyword",
			body: "APPROVE",
			fields: []Field{
				{Name: "decision", Type: FieldTypeSelect, Options: []Option{
					{Value: "approve", Label: "Approve"},
					{Value: "reject", Label: "Reject"},
				}},
			},
			expected: map[string]any{"decision": "approve"},
		},
		{
			name: "confirm field - yes",
			body: "Yes, please proceed",
			fields: []Field{
				{Name: "confirm", Type: FieldTypeConfirm},
			},
			expected: map[string]any{"confirm": true},
		},
		{
			name: "confirm field - no",
			body: "No, cancel it",
			fields: []Field{
				{Name: "confirm", Type: FieldTypeConfirm},
			},
			expected: map[string]any{"confirm": false},
		},
		{
			name: "text field",
			body: "Please use the blue version instead",
			fields: []Field{
				{Name: "feedback", Type: FieldTypeText},
			},
			expected: map[string]any{"feedback": "Please use the blue version instead"},
		},
		{
			name: "body with quoted text",
			body: "> Original message here\n\nAPPROVE",
			fields: []Field{
				{Name: "decision", Type: FieldTypeSelect, Options: []Option{
					{Value: "approve", Label: "Approve"},
				}},
			},
			expected: map[string]any{"decision": "approve"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := handler.ParseReply(tt.body, tt.fields)

			for key, expected := range tt.expected {
				if values[key] != expected {
					t.Errorf("values[%s] = %v, want %v", key, values[key], expected)
				}
			}
		})
	}
}

func TestEmailHandler_Cancel(t *testing.T) {
	handler := NewEmailHandler(EmailConfig{})

	req := &InputRequest{
		ID: "test-1",
		Recipient: Recipient{
			Type:    RecipientEmail,
			Address: "user@example.com",
		},
	}
	handler.Send(context.Background(), req)

	// Cancel existing request
	err := handler.Cancel("test-1")
	if err != nil {
		t.Errorf("Cancel failed: %v", err)
	}
	if handler.PendingCount() != 0 {
		t.Error("pending request should be removed")
	}

	// Cancel non-existent request
	err = handler.Cancel("non-existent")
	if err == nil {
		t.Error("expected error for non-existent request")
	}
}
