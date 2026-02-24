// Package hitl provides human-in-the-loop functionality for agent input requests.
package hitl

import (
	"encoding/json"
	"time"
)

// FieldType represents the type of input field.
type FieldType string

const (
	FieldTypeText        FieldType = "text"
	FieldTypeTextarea    FieldType = "textarea"
	FieldTypeSelect      FieldType = "select"
	FieldTypeMultiselect FieldType = "multiselect"
	FieldTypeConfirm     FieldType = "confirm"
	FieldTypeNumber      FieldType = "number"
	FieldTypeDate        FieldType = "date"
	FieldTypeFile        FieldType = "file"
)

// RecipientType represents who should receive the input request.
type RecipientType string

const (
	RecipientOwner RecipientType = "owner"
	RecipientEmail RecipientType = "email"
	RecipientChat  RecipientType = "chat"
)

// DefaultMaxRetries is the default number of validation retries.
const DefaultMaxRetries = 3

// DefaultTimeout is the default timeout for input requests.
const DefaultTimeout = 24 * time.Hour

// FallbackAction specifies what to do when an input request times out.
type FallbackAction string

const (
	FallbackError    FallbackAction = "error"
	FallbackDefault  FallbackAction = "default"
	FallbackRetry    FallbackAction = "retry"
	FallbackEscalate FallbackAction = "escalate"
	FallbackSkip     FallbackAction = "skip"
)

// Fallback specifies timeout behavior for input requests.
type Fallback struct {
	Action               FallbackAction `json:"action"`
	DefaultValues        map[string]any `json:"default_values,omitempty"`
	EscalationRecipient  *Recipient     `json:"escalation_recipient,omitempty"`
	MaxRetries           int            `json:"max_retries,omitempty"`
}

// InputRequest represents a request for human input from an agent.
type InputRequest struct {
	ID         string        `json:"id"`
	Timeout    time.Duration `json:"timeout"`
	Recipient  Recipient     `json:"recipient"`
	Context    string        `json:"context"`
	Fields     []Field       `json:"fields"`
	Persona    *Persona      `json:"persona,omitempty"`
	MaxRetries int           `json:"max_retries,omitempty"`
	Fallback   *Fallback     `json:"fallback,omitempty"`
}

// Field represents a single input field in the request.
type Field struct {
	Name        string      `json:"name"`
	Type        FieldType   `json:"type"`
	Label       string      `json:"label"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required"`
	Default     any         `json:"default,omitempty"`
	Options     []Option    `json:"options,omitempty"`
	Validation  *Validation `json:"validation,omitempty"`
}

// Option represents a choice for select/multiselect fields.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Validation contains validation rules for a field.
type Validation struct {
	MinLength *int    `json:"minLength,omitempty"`
	MaxLength *int    `json:"maxLength,omitempty"`
	Min       *int    `json:"min,omitempty"`
	Max       *int    `json:"max,omitempty"`
	Pattern   *string `json:"pattern,omitempty"`
	Message   string  `json:"message,omitempty"` // Custom error message
}

// Recipient specifies who should receive the input request.
type Recipient struct {
	Type    RecipientType `json:"type"`
	Address string        `json:"address,omitempty"`
}

// Persona allows agents to customize how they present themselves.
type Persona struct {
	Name      string `json:"name"`
	Signature string `json:"signature,omitempty"`
}

// InputResponse represents the human's response to an input request.
type InputResponse struct {
	RequestID string         `json:"request_id"`
	Values    map[string]any `json:"values"`
	Timestamp time.Time      `json:"timestamp"`
	Skipped   bool           `json:"skipped,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for InputRequest to handle Duration.
func (r *InputRequest) MarshalJSON() ([]byte, error) {
	type Alias InputRequest
	return json.Marshal(&struct {
		Timeout string `json:"timeout"`
		*Alias
	}{
		Timeout: r.Timeout.String(),
		Alias:   (*Alias)(r),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for InputRequest to handle Duration.
func (r *InputRequest) UnmarshalJSON(data []byte) error {
	type Alias InputRequest
	aux := &struct {
		Timeout string `json:"timeout"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Timeout != "" {
		d, err := time.ParseDuration(aux.Timeout)
		if err != nil {
			return err
		}
		r.Timeout = d
	}
	return nil
}
