// Package hitl provides human-in-the-loop functionality for agent input requests.
// This file implements email-based input handling.
package hitl

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EmailConfig configures the email input handler.
type EmailConfig struct {
	// SMTP configuration for sending emails
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromAddress  string
	FromName     string

	// IMAP configuration for receiving replies
	IMAPHost     string
	IMAPPort     int
	IMAPUsername string
	IMAPPassword string
	IMAPFolder   string

	// Security settings
	AllowedRecipients []string // Allowlist of email addresses that can receive requests
	PollInterval      time.Duration
}

// EmailHandler handles human input via email.
type EmailHandler struct {
	config   EmailConfig
	pending  map[string]*pendingEmailRequest
	keywords map[string]string
}

type pendingEmailRequest struct {
	Request   *InputRequest
	MessageID string
	SentAt    time.Time
}

// NewEmailHandler creates a new email input handler.
func NewEmailHandler(config EmailConfig) *EmailHandler {
	if config.PollInterval == 0 {
		config.PollInterval = 30 * time.Second
	}
	if config.IMAPFolder == "" {
		config.IMAPFolder = "INBOX"
	}

	return &EmailHandler{
		config:  config,
		pending: make(map[string]*pendingEmailRequest),
		keywords: map[string]string{
			"APPROVE":  "approve",
			"APPROVED": "approve",
			"YES":      "true",
			"REJECT":   "reject",
			"REJECTED": "reject",
			"NO":       "false",
			"DENY":     "reject",
			"DENIED":   "reject",
			"HOLD":     "hold",
			"PENDING":  "hold",
			"OK":       "true",
			"CONFIRM":  "true",
			"CANCEL":   "false",
		},
	}
}

// Send sends an input request via email and returns the message ID.
func (h *EmailHandler) Send(ctx context.Context, req *InputRequest) (string, error) {
	// Validate recipient
	if req.Recipient.Type != RecipientEmail {
		return "", fmt.Errorf("invalid recipient type: expected email, got %s", req.Recipient.Type)
	}

	if req.Recipient.Address == "" {
		return "", fmt.Errorf("recipient email address is required")
	}

	// Check allowlist
	if len(h.config.AllowedRecipients) > 0 {
		allowed := false
		for _, addr := range h.config.AllowedRecipients {
			if strings.EqualFold(addr, req.Recipient.Address) {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", fmt.Errorf("recipient %s not in allowed list", req.Recipient.Address)
		}
	}

	// Generate message ID
	messageID := fmt.Sprintf("<%s@ayo>", uuid.New().String())

	// Generate email content
	subject, body := h.generateEmail(req)

	// In a real implementation, this would send via SMTP
	// For now, we store the pending request
	h.pending[req.ID] = &pendingEmailRequest{
		Request:   req,
		MessageID: messageID,
		SentAt:    time.Now(),
	}

	// Log the email that would be sent (for debugging/testing)
	_ = subject
	_ = body

	return messageID, nil
}

// generateEmail creates the email subject and body from an input request.
func (h *EmailHandler) generateEmail(req *InputRequest) (subject, body string) {
	// Build subject
	if req.Persona != nil && req.Persona.Name != "" {
		subject = fmt.Sprintf("Input Required from %s: %s", req.Persona.Name, req.Context)
	} else {
		subject = fmt.Sprintf("Input Required: %s", req.Context)
	}
	if len(subject) > 78 {
		subject = subject[:75] + "..."
	}

	// Build body
	var sb strings.Builder

	// Greeting
	sb.WriteString("Hi,\n\n")

	// Context
	sb.WriteString(req.Context)
	sb.WriteString("\n\n")

	// Separator
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Fields
	for _, field := range req.Fields {
		sb.WriteString(fmt.Sprintf("**%s**", field.Label))
		if field.Required {
			sb.WriteString(" (required)")
		}
		sb.WriteString("\n")

		if field.Description != "" {
			sb.WriteString(field.Description)
			sb.WriteString("\n")
		}

		// Options for select fields
		if len(field.Options) > 0 {
			sb.WriteString("\nPlease reply with one of:\n")
			for _, opt := range field.Options {
				keyword := strings.ToUpper(opt.Value)
				sb.WriteString(fmt.Sprintf("• %s - %s\n", keyword, opt.Label))
			}
		}
		sb.WriteString("\n")
	}

	// Separator
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Instructions
	sb.WriteString("Please reply to this email with your response.\n\n")

	// Signature
	if req.Persona != nil && req.Persona.Signature != "" {
		sb.WriteString(req.Persona.Signature)
	} else if req.Persona != nil && req.Persona.Name != "" {
		sb.WriteString(fmt.Sprintf("Best regards,\n%s", req.Persona.Name))
	} else {
		sb.WriteString("Best regards,\nAyo Agent")
	}

	return subject, sb.String()
}

// WaitForResponse waits for a reply to the given request.
func (h *EmailHandler) WaitForResponse(ctx context.Context, requestID string) (*InputResponse, error) {
	pending, ok := h.pending[requestID]
	if !ok {
		return nil, fmt.Errorf("no pending request with ID %s", requestID)
	}

	req := pending.Request
	timeout := req.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	deadline := pending.SentAt.Add(timeout)

	// Poll for response
	ticker := time.NewTicker(h.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Check timeout
			if time.Now().After(deadline) {
				return nil, &TimeoutError{
					RequestID: requestID,
					Timeout:   timeout,
				}
			}

			// In a real implementation, this would poll IMAP for replies
			// that reference the original message ID
			response, found, err := h.checkForReply(ctx, pending.MessageID, req)
			if err != nil {
				return nil, err
			}
			if found {
				delete(h.pending, requestID)
				return response, nil
			}
		}
	}
}

// checkForReply checks for email replies to the given message.
// This is a stub that would be implemented with real IMAP polling.
func (h *EmailHandler) checkForReply(ctx context.Context, messageID string, req *InputRequest) (*InputResponse, bool, error) {
	// In a real implementation:
	// 1. Connect to IMAP server
	// 2. Search for emails with In-Reply-To header matching messageID
	// 3. Parse the reply body
	// 4. Extract values based on field types

	// This stub always returns not found
	return nil, false, nil
}

// ParseReply parses an email reply body into field values.
func (h *EmailHandler) ParseReply(body string, fields []Field) map[string]any {
	values := make(map[string]any)

	// Clean up body - remove quoted text (lines starting with >)
	lines := strings.Split(body, "\n")
	var cleanLines []string
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), ">") {
			cleanLines = append(cleanLines, line)
		}
	}
	cleanBody := strings.Join(cleanLines, "\n")
	cleanBody = strings.TrimSpace(cleanBody)

	// For single field requests, use the whole body
	if len(fields) == 1 {
		field := fields[0]
		value := h.parseFieldValue(cleanBody, field)
		if value != nil {
			values[field.Name] = value
		}
		return values
	}

	// For multiple fields, try to match keywords or labeled responses
	for _, field := range fields {
		value := h.extractFieldValue(cleanBody, field)
		if value != nil {
			values[field.Name] = value
		}
	}

	return values
}

// parseFieldValue extracts a value for a single field from the body.
func (h *EmailHandler) parseFieldValue(body string, field Field) any {
	body = strings.TrimSpace(body)

	switch field.Type {
	case FieldTypeSelect:
		// Look for keyword match
		bodyUpper := strings.ToUpper(body)
		for _, opt := range field.Options {
			keyword := strings.ToUpper(opt.Value)
			if strings.Contains(bodyUpper, keyword) {
				return opt.Value
			}
		}
		// Check standard keywords
		for keyword, value := range h.keywords {
			if strings.Contains(bodyUpper, keyword) {
				return value
			}
		}
		// Use first word as fallback
		if words := strings.Fields(body); len(words) > 0 {
			return strings.ToLower(words[0])
		}

	case FieldTypeConfirm:
		bodyUpper := strings.ToUpper(body)
		if matched, _ := regexp.MatchString(`\b(YES|TRUE|CONFIRM|OK|APPROVE|ACCEPTED?|Y)\b`, bodyUpper); matched {
			return true
		}
		if matched, _ := regexp.MatchString(`\b(NO|FALSE|CANCEL|REJECT|DENIED?|N)\b`, bodyUpper); matched {
			return false
		}
		return nil

	case FieldTypeNumber:
		// Extract first number
		re := regexp.MustCompile(`-?\d+(?:\.\d+)?`)
		if match := re.FindString(body); match != "" {
			return match
		}

	case FieldTypeText, FieldTypeTextarea:
		return body

	default:
		return body
	}

	return nil
}

// extractFieldValue extracts a specific field's value from multi-field response.
func (h *EmailHandler) extractFieldValue(body string, field Field) any {
	// Try to find the field by label pattern: "Label: value"
	patterns := []string{
		fmt.Sprintf(`(?i)%s:\s*(.+?)(?:\n|$)`, regexp.QuoteMeta(field.Label)),
		fmt.Sprintf(`(?i)%s:\s*(.+?)(?:\n|$)`, regexp.QuoteMeta(field.Name)),
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(body); len(matches) > 1 {
			return h.parseFieldValue(strings.TrimSpace(matches[1]), field)
		}
	}

	return nil
}

// Cancel cancels a pending email request.
func (h *EmailHandler) Cancel(requestID string) error {
	if _, ok := h.pending[requestID]; !ok {
		return fmt.Errorf("no pending request with ID %s", requestID)
	}
	delete(h.pending, requestID)
	return nil
}

// PendingCount returns the number of pending email requests.
func (h *EmailHandler) PendingCount() int {
	return len(h.pending)
}
