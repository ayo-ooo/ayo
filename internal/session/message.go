package session

import (
	"charm.land/fantasy"
)

// MessageRole represents the role of a message sender.
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// Message represents a message in a conversation session.
type Message struct {
	ID         string
	SessionID  string
	Role       MessageRole
	Parts      []ContentPart
	Model      string
	Provider   string
	CreatedAt  int64
	UpdatedAt  int64
	FinishedAt int64
}

// TextContent returns the first text content from the message, or empty string.
func (m *Message) TextContent() string {
	for _, part := range m.Parts {
		if tc, ok := part.(TextContent); ok {
			return tc.Text
		}
	}
	return ""
}

// ReasoningText returns the first reasoning content from the message, or empty string.
func (m *Message) ReasoningText() string {
	for _, part := range m.Parts {
		if rc, ok := part.(ReasoningContent); ok {
			return rc.Text
		}
	}
	return ""
}

// ToolCalls returns all tool calls from the message.
func (m *Message) ToolCalls() []ToolCall {
	var calls []ToolCall
	for _, part := range m.Parts {
		if tc, ok := part.(ToolCall); ok {
			calls = append(calls, tc)
		}
	}
	return calls
}

// ToolResults returns all tool results from the message.
func (m *Message) ToolResults() []ToolResult {
	var results []ToolResult
	for _, part := range m.Parts {
		if tr, ok := part.(ToolResult); ok {
			results = append(results, tr)
		}
	}
	return results
}

// IsFinished returns true if the message contains a finish part.
func (m *Message) IsFinished() bool {
	for _, part := range m.Parts {
		if _, ok := part.(Finish); ok {
			return true
		}
	}
	return false
}

// FinishPart returns the finish part if present, or nil.
func (m *Message) FinishPart() *Finish {
	for _, part := range m.Parts {
		if f, ok := part.(Finish); ok {
			return &f
		}
	}
	return nil
}

// ToFantasyMessage converts a session Message to a Fantasy Message.
func (m *Message) ToFantasyMessage() fantasy.Message {
	var fantasyParts []fantasy.MessagePart

	for _, part := range m.Parts {
		switch p := part.(type) {
		case TextContent:
			fantasyParts = append(fantasyParts, fantasy.TextPart{Text: p.Text})
		case ReasoningContent:
			fantasyParts = append(fantasyParts, fantasy.ReasoningPart{Text: p.Text})
		case FileContent:
			fantasyParts = append(fantasyParts, fantasy.FilePart{
				Filename:  p.Filename,
				Data:      p.Data,
				MediaType: p.MediaType,
			})
		case ToolCall:
			fantasyParts = append(fantasyParts, fantasy.ToolCallPart{
				ToolCallID:       p.ID,
				ToolName:         p.Name,
				Input:            p.Input,
				ProviderExecuted: p.ProviderExecuted,
			})
		case ToolResult:
			var output fantasy.ToolResultOutputContent
			if p.IsError {
				output = fantasy.ToolResultOutputContentError{
					Error: errorString(p.Content),
				}
			} else {
				output = fantasy.ToolResultOutputContentText{Text: p.Content}
			}
			fantasyParts = append(fantasyParts, fantasy.ToolResultPart{
				ToolCallID: p.ToolCallID,
				Output:     output,
			})
		case Finish:
			// Finish is not a Fantasy message part, skip
		}
	}

	var role fantasy.MessageRole
	switch m.Role {
	case RoleSystem:
		role = fantasy.MessageRoleSystem
	case RoleUser:
		role = fantasy.MessageRoleUser
	case RoleAssistant:
		role = fantasy.MessageRoleAssistant
	case RoleTool:
		role = fantasy.MessageRoleTool
	}

	return fantasy.Message{
		Role:    role,
		Content: fantasyParts,
	}
}

// FromFantasyMessage creates session Messages from a Fantasy Message.
// Note: A single Fantasy message may produce multiple session messages
// if it contains tool results (which use the tool role).
func FromFantasyMessage(fm fantasy.Message, sessionID, model, provider string) []Message {
	role := roleFromFantasy(fm.Role)
	parts := partsFromFantasy(fm.Content)

	// If this is an assistant message with tool calls followed by tool results,
	// we need to split them as tool results get their own message with role=tool
	if role == RoleAssistant {
		var assistantParts []ContentPart
		var toolResultParts []ContentPart

		for _, part := range parts {
			if _, ok := part.(ToolResult); ok {
				toolResultParts = append(toolResultParts, part)
			} else {
				assistantParts = append(assistantParts, part)
			}
		}

		if len(toolResultParts) > 0 {
			var msgs []Message
			if len(assistantParts) > 0 {
				msgs = append(msgs, Message{
					SessionID: sessionID,
					Role:      RoleAssistant,
					Parts:     assistantParts,
					Model:     model,
					Provider:  provider,
				})
			}
			msgs = append(msgs, Message{
				SessionID: sessionID,
				Role:      RoleTool,
				Parts:     toolResultParts,
				Model:     model,
				Provider:  provider,
			})
			return msgs
		}
	}

	return []Message{{
		SessionID: sessionID,
		Role:      role,
		Parts:     parts,
		Model:     model,
		Provider:  provider,
	}}
}

func roleFromFantasy(fr fantasy.MessageRole) MessageRole {
	switch fr {
	case fantasy.MessageRoleSystem:
		return RoleSystem
	case fantasy.MessageRoleUser:
		return RoleUser
	case fantasy.MessageRoleAssistant:
		return RoleAssistant
	case fantasy.MessageRoleTool:
		return RoleTool
	default:
		return RoleUser
	}
}

func partsFromFantasy(fantasyParts []fantasy.MessagePart) []ContentPart {
	var parts []ContentPart

	for _, fp := range fantasyParts {
		switch p := fp.(type) {
		case fantasy.TextPart:
			parts = append(parts, TextContent{Text: p.Text})
		case fantasy.ReasoningPart:
			parts = append(parts, ReasoningContent{Text: p.Text})
		case fantasy.FilePart:
			parts = append(parts, FileContent{
				Filename:  p.Filename,
				Data:      p.Data,
				MediaType: p.MediaType,
			})
		case fantasy.ToolCallPart:
			parts = append(parts, ToolCall{
				ID:               p.ToolCallID,
				Name:             p.ToolName,
				Input:            p.Input,
				ProviderExecuted: p.ProviderExecuted,
			})
		case fantasy.ToolResultPart:
			tr := ToolResult{
				ToolCallID: p.ToolCallID,
			}
			switch out := p.Output.(type) {
			case fantasy.ToolResultOutputContentText:
				tr.Content = out.Text
			case fantasy.ToolResultOutputContentError:
				tr.Content = out.Error.Error()
				tr.IsError = true
			case fantasy.ToolResultOutputContentMedia:
				tr.Content = out.Text
			}
			parts = append(parts, tr)
		}
	}

	return parts
}

// errorString implements error interface for string.
type errorString string

func (e errorString) Error() string { return string(e) }
