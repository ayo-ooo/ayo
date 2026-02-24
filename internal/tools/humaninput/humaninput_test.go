package humaninput

import (
	"context"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/hitl"
)

// mockRenderer is a test implementation of FormRenderer.
type mockRenderer struct {
	response *hitl.InputResponse
	err      error
}

func (m *mockRenderer) Render(ctx context.Context, req *hitl.InputRequest) (*hitl.InputResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.response == nil {
		return &hitl.InputResponse{
			RequestID: req.ID,
			Values:    map[string]any{},
			Timestamp: time.Now(),
		}, nil
	}
	return m.response, nil
}

func TestBuildInputRequest(t *testing.T) {
	params := HumanInputParams{
		Context: "Need user confirmation",
		Fields: []FieldParams{
			{
				Name:     "confirm",
				Type:     "confirm",
				Label:    "Do you approve?",
				Required: true,
			},
			{
				Name:    "notes",
				Type:    "text",
				Label:   "Additional notes",
				Default: "None",
			},
		},
		Recipient: "owner",
		Timeout:   "5m",
	}

	req, err := buildInputRequest(params, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Context != "Need user confirmation" {
		t.Errorf("wrong context: %s", req.Context)
	}
	if req.Timeout != 5*time.Minute {
		t.Errorf("wrong timeout: %v", req.Timeout)
	}
	if req.Recipient.Type != hitl.RecipientOwner {
		t.Errorf("wrong recipient type: %v", req.Recipient.Type)
	}
	if len(req.Fields) != 2 {
		t.Errorf("wrong number of fields: %d", len(req.Fields))
	}
	if req.Fields[0].Type != hitl.FieldTypeConfirm {
		t.Errorf("wrong field type: %v", req.Fields[0].Type)
	}
	if !req.Fields[0].Required {
		t.Error("field should be required")
	}
	if req.Fields[1].Default != "None" {
		t.Errorf("wrong default: %v", req.Fields[1].Default)
	}
}

func TestBuildInputRequest_EmailRecipient(t *testing.T) {
	params := HumanInputParams{
		Context: "Test",
		Fields: []FieldParams{
			{Name: "test", Type: "text", Label: "Test"},
		},
		Recipient: "user@example.com",
	}

	req, err := buildInputRequest(params, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Recipient.Type != hitl.RecipientEmail {
		t.Errorf("expected email recipient, got %v", req.Recipient.Type)
	}
	if req.Recipient.Address != "user@example.com" {
		t.Errorf("wrong address: %s", req.Recipient.Address)
	}
}

func TestBuildInputRequest_SelectOptions(t *testing.T) {
	params := HumanInputParams{
		Context: "Choose",
		Fields: []FieldParams{
			{
				Name:    "choice",
				Type:    "select",
				Label:   "Pick one",
				Options: []string{"a", "b", "c"},
			},
		},
	}

	req, err := buildInputRequest(params, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(req.Fields[0].Options) != 3 {
		t.Errorf("expected 3 options, got %d", len(req.Fields[0].Options))
	}
	if req.Fields[0].Options[0].Value != "a" {
		t.Errorf("wrong option value: %s", req.Fields[0].Options[0].Value)
	}
}

func TestBuildInputRequest_InvalidTimeout(t *testing.T) {
	params := HumanInputParams{
		Context: "Test",
		Fields: []FieldParams{
			{Name: "test", Type: "text", Label: "Test"},
		},
		Timeout: "invalid",
	}

	_, err := buildInputRequest(params, time.Hour)
	if err == nil {
		t.Error("expected error for invalid timeout")
	}
}

func TestBuildInputRequest_DefaultTimeout(t *testing.T) {
	params := HumanInputParams{
		Context: "Test",
		Fields: []FieldParams{
			{Name: "test", Type: "text", Label: "Test"},
		},
	}

	req, err := buildInputRequest(params, 2*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Timeout != 2*time.Hour {
		t.Errorf("expected default timeout 2h, got %v", req.Timeout)
	}
}

func TestNewHumanInputTool_NoRenderer(t *testing.T) {
	tool := NewHumanInputTool(ToolConfig{
		Renderer: nil,
	})

	if tool == nil {
		t.Fatal("expected tool to be created")
	}
}
