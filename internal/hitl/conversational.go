package hitl

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MessageSender sends a message and returns any error.
type MessageSender func(string) error

// MessageReceiver waits for and returns a user message.
type MessageReceiver func(ctx context.Context) (string, error)

// ConversationalHandler presents InputRequest schemas as conversational Q&A sequences.
type ConversationalHandler struct {
	send     MessageSender
	receive  MessageReceiver
	persona  *PersonaManager
	sanitize *Sanitizer
}

// NewConversationalHandler creates a new ConversationalHandler.
func NewConversationalHandler(send MessageSender, receive MessageReceiver) *ConversationalHandler {
	return &ConversationalHandler{
		send:     send,
		receive:  receive,
		sanitize: DefaultSanitizer(),
	}
}

// SetPersona sets the persona manager for the handler.
func (h *ConversationalHandler) SetPersona(p *PersonaManager) {
	h.persona = p
}

// Run presents the input request as a conversation and collects responses.
func (h *ConversationalHandler) Run(ctx context.Context, req *InputRequest) (*InputResponse, error) {
	values := make(map[string]any)
	
	// Send context/intro if provided
	if req.Context != "" {
		if err := h.send(req.Context); err != nil {
			return nil, fmt.Errorf("failed to send context: %w", err)
		}
	}
	
	// Ask each field
	for _, field := range req.Fields {
		value, skipped, err := h.askField(ctx, field)
		if err != nil {
			return nil, err
		}
		if !skipped {
			values[field.Name] = value
		}
	}
	
	// Send confirmation
	if err := h.send("Thanks! I have all the information I need."); err != nil {
		return nil, fmt.Errorf("failed to send confirmation: %w", err)
	}
	
	return &InputResponse{
		RequestID: req.ID,
		Values:    values,
		Timestamp: time.Now(),
	}, nil
}

// askField asks a single field and validates the response.
func (h *ConversationalHandler) askField(ctx context.Context, field Field) (any, bool, error) {
	prompt := h.formatFieldPrompt(field)
	
	maxRetries := DefaultMaxRetries
	var lastError error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Send prompt (with error message if retrying)
		msg := prompt
		if lastError != nil {
			msg = fmt.Sprintf("❌ %s\n\n%s", lastError.Error(), prompt)
		}
		
		if err := h.send(msg); err != nil {
			return nil, false, fmt.Errorf("failed to send prompt: %w", err)
		}
		
		// Receive response
		response, err := h.receive(ctx)
		if err != nil {
			return nil, false, fmt.Errorf("failed to receive response: %w", err)
		}
		
		response = strings.TrimSpace(response)
		
		// Check for skip on optional fields
		if !field.Required && isSkipResponse(response) {
			return nil, true, nil
		}
		
		// Parse and validate
		value, err := h.parseResponse(field, response)
		if err != nil {
			lastError = err
			continue
		}
		
		// Validate
		if validErr := validateFieldValue(field, value); validErr != nil {
			lastError = validErr
			continue
		}
		
		// Check required
		if field.Required && isEmpty(value) {
			lastError = fmt.Errorf("%s is required", field.Label)
			continue
		}
		
		// Send acknowledgment
		ack := h.formatAcknowledgment(field, value)
		if err := h.send(ack); err != nil {
			return nil, false, fmt.Errorf("failed to send acknowledgment: %w", err)
		}
		
		return value, false, nil
	}
	
	return nil, false, fmt.Errorf("max retries exceeded for field %q: %w", field.Name, lastError)
}

// formatFieldPrompt creates a conversational prompt for a field.
func (h *ConversationalHandler) formatFieldPrompt(field Field) string {
	var sb strings.Builder
	
	// Label as question
	label := field.Label
	if !strings.HasSuffix(label, "?") && !strings.HasSuffix(label, ":") {
		if field.Type == FieldTypeConfirm {
			label += "?"
		} else {
			label += ":"
		}
	}
	sb.WriteString(label)
	
	// Description if present
	if field.Description != "" {
		sb.WriteString("\n")
		sb.WriteString(field.Description)
	}
	
	// Type-specific formatting
	switch field.Type {
	case FieldTypeSelect:
		sb.WriteString("\n\n")
		for i, opt := range field.Options {
			sb.WriteString(fmt.Sprintf("%s %s\n", numberEmoji(i+1), opt.Label))
		}
		sb.WriteString(fmt.Sprintf("\nReply with 1-%d", len(field.Options)))
		
	case FieldTypeMultiselect:
		sb.WriteString("\n\n")
		for i, opt := range field.Options {
			sb.WriteString(fmt.Sprintf("%s %s\n", numberEmoji(i+1), opt.Label))
		}
		sb.WriteString(fmt.Sprintf("\nReply with numbers separated by commas (e.g., 1,3)"))
		
	case FieldTypeConfirm:
		sb.WriteString("\n\nReply Yes or No")
		
	case FieldTypeNumber:
		if field.Validation != nil {
			if field.Validation.Min != nil && field.Validation.Max != nil {
				sb.WriteString(fmt.Sprintf("\n\n(Enter a number between %d and %d)", *field.Validation.Min, *field.Validation.Max))
			} else if field.Validation.Min != nil {
				sb.WriteString(fmt.Sprintf("\n\n(Enter a number >= %d)", *field.Validation.Min))
			} else if field.Validation.Max != nil {
				sb.WriteString(fmt.Sprintf("\n\n(Enter a number <= %d)", *field.Validation.Max))
			}
		}
		
	case FieldTypeDate:
		sb.WriteString("\n\n(Enter a date like 'tomorrow', 'next Monday', or 'March 15')")
	}
	
	// Skip option for optional fields
	if !field.Required {
		sb.WriteString("\n\n(or reply \"skip\" to skip)")
	}
	
	return sb.String()
}

// parseResponse parses a user response based on field type.
func (h *ConversationalHandler) parseResponse(field Field, response string) (any, error) {
	switch field.Type {
	case FieldTypeText, FieldTypeTextarea:
		return response, nil
		
	case FieldTypeNumber:
		n, err := strconv.ParseFloat(response, 64)
		if err != nil {
			return nil, fmt.Errorf("please enter a valid number")
		}
		return n, nil
		
	case FieldTypeConfirm:
		lower := strings.ToLower(response)
		switch lower {
		case "yes", "y", "true", "1":
			return true, nil
		case "no", "n", "false", "0":
			return false, nil
		default:
			return nil, fmt.Errorf("please reply Yes or No")
		}
		
	case FieldTypeSelect:
		// Accept number or option value
		idx, err := strconv.Atoi(response)
		if err == nil {
			if idx < 1 || idx > len(field.Options) {
				return nil, fmt.Errorf("please enter a number between 1 and %d", len(field.Options))
			}
			return field.Options[idx-1].Value, nil
		}
		// Try matching option value
		for _, opt := range field.Options {
			if strings.EqualFold(opt.Value, response) || strings.EqualFold(opt.Label, response) {
				return opt.Value, nil
			}
		}
		return nil, fmt.Errorf("please enter a number between 1 and %d", len(field.Options))
		
	case FieldTypeMultiselect:
		parts := strings.Split(response, ",")
		var values []any
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			idx, err := strconv.Atoi(p)
			if err == nil {
				if idx < 1 || idx > len(field.Options) {
					return nil, fmt.Errorf("invalid option number: %d", idx)
				}
				values = append(values, field.Options[idx-1].Value)
			} else {
				// Try matching option value
				found := false
				for _, opt := range field.Options {
					if strings.EqualFold(opt.Value, p) || strings.EqualFold(opt.Label, p) {
						values = append(values, opt.Value)
						found = true
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("invalid option: %s", p)
				}
			}
		}
		return values, nil
		
	case FieldTypeDate:
		// For now, accept as string - could add natural language parsing
		return response, nil
		
	default:
		return response, nil
	}
}

// formatAcknowledgment creates an acknowledgment message for a value.
func (h *ConversationalHandler) formatAcknowledgment(field Field, value any) string {
	switch field.Type {
	case FieldTypeSelect:
		// Find the label for the value
		for _, opt := range field.Options {
			if opt.Value == value {
				return fmt.Sprintf("Got it - \"%s\"", opt.Label)
			}
		}
		return fmt.Sprintf("Got it - %v", value)
		
	case FieldTypeMultiselect:
		values, ok := value.([]any)
		if !ok || len(values) == 0 {
			return "Got it"
		}
		var labels []string
		for _, v := range values {
			for _, opt := range field.Options {
				if opt.Value == v {
					labels = append(labels, opt.Label)
					break
				}
			}
		}
		return fmt.Sprintf("Got it - %s", strings.Join(labels, ", "))
		
	case FieldTypeConfirm:
		if b, ok := value.(bool); ok && b {
			return "Got it - Yes"
		}
		return "Got it - No"
		
	default:
		return fmt.Sprintf("Got it - %v", value)
	}
}

// isSkipResponse checks if a response indicates the user wants to skip.
func isSkipResponse(response string) bool {
	lower := strings.ToLower(response)
	return lower == "skip" || lower == "s" || lower == "-" || lower == ""
}

// numberEmoji returns an emoji for a number (1-9).
func numberEmoji(n int) string {
	emojis := []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣"}
	if n >= 1 && n <= 9 {
		return emojis[n-1]
	}
	return fmt.Sprintf("%d.", n)
}
