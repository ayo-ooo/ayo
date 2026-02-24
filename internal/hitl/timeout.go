package hitl

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// TimeoutError is returned when an input request times out without a fallback.
type TimeoutError struct {
	RequestID string
	Timeout   time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("input request %q timed out after %s", e.RequestID, e.Timeout)
}

// RetryError signals that the request should be retried.
type RetryError struct {
	RequestID    string
	AttemptsMade int
	MaxRetries   int
}

func (e *RetryError) Error() string {
	return fmt.Sprintf("input request %q needs retry (attempt %d/%d)", e.RequestID, e.AttemptsMade, e.MaxRetries)
}

// EscalationRequest signals that the request should be escalated.
type EscalationRequest struct {
	OriginalRequest *InputRequest
	NewRecipient    *Recipient
	Reason          string
}

// FormRenderer renders an input request and collects a response.
type FormRenderer interface {
	Render(ctx context.Context, req *InputRequest) (*InputResponse, error)
}

// ReminderFunc is called to send reminders before timeout.
type ReminderFunc func(req *InputRequest, reminderType string) error

// TimeoutHandler manages input request timeouts and fallbacks.
type TimeoutHandler struct {
	defaultTimeout time.Duration
	reminderFn     ReminderFunc
	onEscalate     func(*EscalationRequest) error
}

// NewTimeoutHandler creates a new TimeoutHandler with the given default timeout.
func NewTimeoutHandler(defaultTimeout time.Duration) *TimeoutHandler {
	if defaultTimeout == 0 {
		defaultTimeout = DefaultTimeout
	}
	return &TimeoutHandler{
		defaultTimeout: defaultTimeout,
	}
}

// SetReminderFunc sets the function to call for reminders.
func (h *TimeoutHandler) SetReminderFunc(fn ReminderFunc) {
	h.reminderFn = fn
}

// SetEscalateFunc sets the function to call for escalations.
func (h *TimeoutHandler) SetEscalateFunc(fn func(*EscalationRequest) error) {
	h.onEscalate = fn
}

// Run executes an input request with timeout handling.
func (h *TimeoutHandler) Run(ctx context.Context, req *InputRequest, renderer FormRenderer) (*InputResponse, error) {
	timeout := req.Timeout
	if timeout == 0 {
		timeout = h.defaultTimeout
	}

	// Schedule reminders for long timeouts
	var reminderCancel func()
	if timeout >= time.Hour && h.reminderFn != nil {
		reminderCancel = h.scheduleReminders(req, timeout)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	defer func() {
		if reminderCancel != nil {
			reminderCancel()
		}
	}()

	resp, err := renderer.Render(ctx, req)
	if errors.Is(err, context.DeadlineExceeded) {
		return h.handleTimeout(req, timeout)
	}
	return resp, err
}

// handleTimeout handles timeout based on fallback configuration.
func (h *TimeoutHandler) handleTimeout(req *InputRequest, timeout time.Duration) (*InputResponse, error) {
	if req.Fallback == nil {
		return nil, &TimeoutError{RequestID: req.ID, Timeout: timeout}
	}

	switch req.Fallback.Action {
	case FallbackDefault:
		return h.buildDefaultResponse(req)
	case FallbackEscalate:
		return h.escalate(req)
	case FallbackRetry:
		return nil, &RetryError{
			RequestID:    req.ID,
			AttemptsMade: 1,
			MaxRetries:   req.Fallback.MaxRetries,
		}
	case FallbackSkip:
		return &InputResponse{
			RequestID: req.ID,
			Values:    make(map[string]any),
			Timestamp: time.Now(),
			Skipped:   true,
		}, nil
	default:
		return nil, &TimeoutError{RequestID: req.ID, Timeout: timeout}
	}
}

// buildDefaultResponse creates a response using default values.
func (h *TimeoutHandler) buildDefaultResponse(req *InputRequest) (*InputResponse, error) {
	values := make(map[string]any)

	// First apply fallback default values
	if req.Fallback != nil && req.Fallback.DefaultValues != nil {
		for k, v := range req.Fallback.DefaultValues {
			values[k] = v
		}
	}

	// Then apply field defaults (fallback values take precedence if specified)
	for _, field := range req.Fields {
		if _, exists := values[field.Name]; !exists && field.Default != nil {
			values[field.Name] = field.Default
		}
	}

	// Check required fields have values
	for _, field := range req.Fields {
		if field.Required {
			if _, exists := values[field.Name]; !exists {
				return nil, fmt.Errorf("no default value for required field %q", field.Name)
			}
		}
	}

	return &InputResponse{
		RequestID: req.ID,
		Values:    values,
		Timestamp: time.Now(),
	}, nil
}

// escalate sends the request to a different recipient.
func (h *TimeoutHandler) escalate(req *InputRequest) (*InputResponse, error) {
	if req.Fallback == nil || req.Fallback.EscalationRecipient == nil {
		return nil, fmt.Errorf("escalation requested but no escalation recipient configured")
	}

	escalation := &EscalationRequest{
		OriginalRequest: req,
		NewRecipient:    req.Fallback.EscalationRecipient,
		Reason:          "original recipient did not respond in time",
	}

	if h.onEscalate != nil {
		if err := h.onEscalate(escalation); err != nil {
			return nil, fmt.Errorf("escalation failed: %w", err)
		}
	}

	// Return nil to indicate the request has been escalated and caller should wait
	return nil, nil
}

// scheduleReminders sets up reminder timers for long timeouts.
func (h *TimeoutHandler) scheduleReminders(req *InputRequest, timeout time.Duration) func() {
	halfTime := timeout / 2
	nearEnd := timeout * 9 / 10

	halfTimer := time.AfterFunc(halfTime, func() {
		if h.reminderFn != nil {
			_ = h.reminderFn(req, "halfway")
		}
	})

	finalTimer := time.AfterFunc(nearEnd, func() {
		if h.reminderFn != nil {
			_ = h.reminderFn(req, "final")
		}
	})

	return func() {
		halfTimer.Stop()
		finalTimer.Stop()
	}
}

// RunWithRetry executes an input request with retry support.
func (h *TimeoutHandler) RunWithRetry(ctx context.Context, req *InputRequest, renderer FormRenderer) (*InputResponse, error) {
	maxRetries := 3
	if req.Fallback != nil && req.Fallback.MaxRetries > 0 {
		maxRetries = req.Fallback.MaxRetries
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := h.Run(ctx, req, renderer)
		if err == nil {
			return resp, nil
		}

		var retryErr *RetryError
		if errors.As(err, &retryErr) {
			lastErr = err
			continue
		}

		return nil, err
	}

	return nil, fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, lastErr)
}
