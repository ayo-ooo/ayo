package hitl

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// mockRenderer implements FormRenderer for testing.
type mockRenderer struct {
	response *InputResponse
	err      error
	delay    time.Duration
}

func (m *mockRenderer) Render(ctx context.Context, req *InputRequest) (*InputResponse, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.response, m.err
}

func TestTimeoutHandler_Run_Success(t *testing.T) {
	h := NewTimeoutHandler(time.Minute)

	req := &InputRequest{
		ID:      "req-123",
		Timeout: time.Second,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "choice", Type: FieldTypeSelect, Label: "Choose"},
		},
	}

	renderer := &mockRenderer{
		response: &InputResponse{
			RequestID: "req-123",
			Values:    map[string]any{"choice": "a"},
		},
	}

	resp, err := h.Run(context.Background(), req, renderer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Values["choice"] != "a" {
		t.Errorf("expected choice=a, got %v", resp.Values["choice"])
	}
}

func TestTimeoutHandler_Run_TimeoutError(t *testing.T) {
	h := NewTimeoutHandler(time.Minute)

	req := &InputRequest{
		ID:      "req-123",
		Timeout: 50 * time.Millisecond,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "choice", Type: FieldTypeSelect, Label: "Choose"},
		},
	}

	renderer := &mockRenderer{
		delay: 200 * time.Millisecond, // Will timeout
	}

	_, err := h.Run(context.Background(), req, renderer)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	var timeoutErr *TimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("expected TimeoutError, got %T", err)
	}
	if timeoutErr.RequestID != "req-123" {
		t.Errorf("expected request ID req-123, got %s", timeoutErr.RequestID)
	}
}

func TestTimeoutHandler_Run_FallbackDefault(t *testing.T) {
	h := NewTimeoutHandler(time.Minute)

	req := &InputRequest{
		ID:      "req-123",
		Timeout: 50 * time.Millisecond,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "choice", Type: FieldTypeSelect, Label: "Choose", Default: "default_choice"},
		},
		Fallback: &Fallback{
			Action: FallbackDefault,
		},
	}

	renderer := &mockRenderer{
		delay: 200 * time.Millisecond,
	}

	resp, err := h.Run(context.Background(), req, renderer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Values["choice"] != "default_choice" {
		t.Errorf("expected default_choice, got %v", resp.Values["choice"])
	}
}

func TestTimeoutHandler_Run_FallbackDefaultValues(t *testing.T) {
	h := NewTimeoutHandler(time.Minute)

	req := &InputRequest{
		ID:      "req-123",
		Timeout: 50 * time.Millisecond,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "choice", Type: FieldTypeSelect, Label: "Choose", Default: "field_default"},
			{Name: "other", Type: FieldTypeText, Label: "Other"},
		},
		Fallback: &Fallback{
			Action: FallbackDefault,
			DefaultValues: map[string]any{
				"choice": "fallback_override",
				"other":  "fallback_value",
			},
		},
	}

	renderer := &mockRenderer{
		delay: 200 * time.Millisecond,
	}

	resp, err := h.Run(context.Background(), req, renderer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Fallback default values take precedence
	if resp.Values["choice"] != "fallback_override" {
		t.Errorf("expected fallback_override, got %v", resp.Values["choice"])
	}
	if resp.Values["other"] != "fallback_value" {
		t.Errorf("expected fallback_value, got %v", resp.Values["other"])
	}
}

func TestTimeoutHandler_Run_FallbackDefaultMissingRequired(t *testing.T) {
	h := NewTimeoutHandler(time.Minute)

	req := &InputRequest{
		ID:      "req-123",
		Timeout: 50 * time.Millisecond,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "required_field", Type: FieldTypeText, Label: "Required", Required: true},
		},
		Fallback: &Fallback{
			Action: FallbackDefault,
		},
	}

	renderer := &mockRenderer{
		delay: 200 * time.Millisecond,
	}

	_, err := h.Run(context.Background(), req, renderer)
	if err == nil {
		t.Fatal("expected error for missing required field default")
	}
}

func TestTimeoutHandler_Run_FallbackSkip(t *testing.T) {
	h := NewTimeoutHandler(time.Minute)

	req := &InputRequest{
		ID:      "req-123",
		Timeout: 50 * time.Millisecond,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "choice", Type: FieldTypeSelect, Label: "Choose"},
		},
		Fallback: &Fallback{
			Action: FallbackSkip,
		},
	}

	renderer := &mockRenderer{
		delay: 200 * time.Millisecond,
	}

	resp, err := h.Run(context.Background(), req, renderer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Skipped {
		t.Error("expected response to be marked as skipped")
	}
}

func TestTimeoutHandler_Run_FallbackRetry(t *testing.T) {
	h := NewTimeoutHandler(time.Minute)

	req := &InputRequest{
		ID:      "req-123",
		Timeout: 50 * time.Millisecond,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "choice", Type: FieldTypeSelect, Label: "Choose"},
		},
		Fallback: &Fallback{
			Action:     FallbackRetry,
			MaxRetries: 3,
		},
	}

	renderer := &mockRenderer{
		delay: 200 * time.Millisecond,
	}

	_, err := h.Run(context.Background(), req, renderer)
	if err == nil {
		t.Fatal("expected retry error")
	}

	var retryErr *RetryError
	if !errors.As(err, &retryErr) {
		t.Fatalf("expected RetryError, got %T", err)
	}
	if retryErr.MaxRetries != 3 {
		t.Errorf("expected max retries 3, got %d", retryErr.MaxRetries)
	}
}

func TestTimeoutHandler_Run_FallbackEscalate(t *testing.T) {
	h := NewTimeoutHandler(time.Minute)

	var escalationReceived *EscalationRequest
	h.SetEscalateFunc(func(req *EscalationRequest) error {
		escalationReceived = req
		return nil
	})

	escalationRecipient := &Recipient{
		Type:    RecipientEmail,
		Address: "manager@example.com",
	}

	req := &InputRequest{
		ID:      "req-123",
		Timeout: 50 * time.Millisecond,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "choice", Type: FieldTypeSelect, Label: "Choose"},
		},
		Fallback: &Fallback{
			Action:              FallbackEscalate,
			EscalationRecipient: escalationRecipient,
		},
	}

	renderer := &mockRenderer{
		delay: 200 * time.Millisecond,
	}

	resp, err := h.Run(context.Background(), req, renderer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Error("expected nil response for escalation")
	}
	if escalationReceived == nil {
		t.Fatal("escalation function not called")
	}
	if escalationReceived.NewRecipient.Address != "manager@example.com" {
		t.Errorf("wrong escalation recipient: %v", escalationReceived.NewRecipient)
	}
}

func TestTimeoutHandler_Reminders(t *testing.T) {
	h := NewTimeoutHandler(time.Minute)

	var reminderCount int32
	h.SetReminderFunc(func(req *InputRequest, reminderType string) error {
		atomic.AddInt32(&reminderCount, 1)
		return nil
	})

	req := &InputRequest{
		ID:      "req-123",
		Timeout: 100 * time.Millisecond, // Short timeout for test but >= 1h in code check
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "choice", Type: FieldTypeSelect, Label: "Choose"},
		},
	}

	// Override the timeout temporarily to trigger reminder logic
	req.Timeout = 2 * time.Hour

	renderer := &mockRenderer{
		response: &InputResponse{
			RequestID: "req-123",
			Values:    map[string]any{"choice": "a"},
		},
		delay: 10 * time.Millisecond, // Quick response
	}

	resp, err := h.Run(context.Background(), req, renderer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response")
	}

	// Reminders should have been scheduled but not triggered (fast response)
	// Give a tiny bit of time for potential timer race
	time.Sleep(5 * time.Millisecond)
	if count := atomic.LoadInt32(&reminderCount); count != 0 {
		t.Errorf("expected 0 reminders triggered, got %d", count)
	}
}

func TestTimeoutHandler_RunWithRetry(t *testing.T) {
	h := NewTimeoutHandler(time.Minute)

	attemptCount := 0
	renderer := &mockRenderer{}

	// First 2 attempts timeout, third succeeds
	originalRender := renderer.Render
	_ = originalRender

	req := &InputRequest{
		ID:      "req-123",
		Timeout: 50 * time.Millisecond,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "choice", Type: FieldTypeSelect, Label: "Choose"},
		},
		Fallback: &Fallback{
			Action:     FallbackRetry,
			MaxRetries: 3,
		},
	}

	// Create a dynamic renderer
	dynamicRenderer := &dynamicMockRenderer{
		renderFn: func(ctx context.Context, r *InputRequest) (*InputResponse, error) {
			attemptCount++
			if attemptCount < 3 {
				// Timeout
				select {
				case <-time.After(200 * time.Millisecond):
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			return &InputResponse{
				RequestID: r.ID,
				Values:    map[string]any{"choice": "success"},
			}, nil
		},
	}

	resp, err := h.RunWithRetry(context.Background(), req, dynamicRenderer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Values["choice"] != "success" {
		t.Errorf("expected choice=success, got %v", resp.Values["choice"])
	}
	if attemptCount != 3 {
		t.Errorf("expected 3 attempts, got %d", attemptCount)
	}
}

func TestTimeoutHandler_RunWithRetry_ExceedsMax(t *testing.T) {
	h := NewTimeoutHandler(time.Minute)

	req := &InputRequest{
		ID:      "req-123",
		Timeout: 50 * time.Millisecond,
		Recipient: Recipient{
			Type: RecipientOwner,
		},
		Fields: []Field{
			{Name: "choice", Type: FieldTypeSelect, Label: "Choose"},
		},
		Fallback: &Fallback{
			Action:     FallbackRetry,
			MaxRetries: 2,
		},
	}

	renderer := &mockRenderer{
		delay: 200 * time.Millisecond, // Always timeout
	}

	_, err := h.RunWithRetry(context.Background(), req, renderer)
	if err == nil {
		t.Fatal("expected max retries exceeded error")
	}
}

func TestTimeoutHandler_DefaultTimeout(t *testing.T) {
	h := NewTimeoutHandler(0) // Should use DefaultTimeout

	if h.defaultTimeout != DefaultTimeout {
		t.Errorf("expected default timeout %v, got %v", DefaultTimeout, h.defaultTimeout)
	}
}

type dynamicMockRenderer struct {
	renderFn func(ctx context.Context, req *InputRequest) (*InputResponse, error)
}

func (d *dynamicMockRenderer) Render(ctx context.Context, req *InputRequest) (*InputResponse, error) {
	return d.renderFn(ctx, req)
}
