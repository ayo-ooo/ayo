// Package session provides session and message management for conversation persistence.
package session

import (
	"encoding/json"
	"fmt"
)

// ContentPart is the interface for message content parts.
type ContentPart interface {
	contentPart()
}

// partType identifies the type of content part for JSON serialization.
type partType string

const (
	partTypeText       partType = "text"
	partTypeReasoning  partType = "reasoning"
	partTypeFile       partType = "file"
	partTypeToolCall   partType = "tool_call"
	partTypeToolResult partType = "tool_result"
	partTypeFinish     partType = "finish"
)

// TextContent represents text content in a message.
type TextContent struct {
	Text string `json:"text"`
}

func (TextContent) contentPart() {}

// ReasoningContent represents reasoning/thinking content in a message.
type ReasoningContent struct {
	Text       string `json:"text"`
	Signature  string `json:"signature,omitempty"`
	StartedAt  int64  `json:"started_at,omitempty"`
	FinishedAt int64  `json:"finished_at,omitempty"`
}

func (ReasoningContent) contentPart() {}

// FileContent represents file/binary content in a message.
type FileContent struct {
	Filename  string `json:"filename"`
	Data      []byte `json:"data"`
	MediaType string `json:"media_type"`
}

func (FileContent) contentPart() {}

// ToolCall represents a tool call in a message.
type ToolCall struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Input            string `json:"input"`
	ProviderExecuted bool   `json:"provider_executed,omitempty"`
	Finished         bool   `json:"finished,omitempty"`
}

func (ToolCall) contentPart() {}

// ToolResult represents the result of a tool call.
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}

func (ToolResult) contentPart() {}

// FinishReason indicates why generation finished.
type FinishReason string

const (
	FinishReasonStop         FinishReason = "stop"
	FinishReasonLength       FinishReason = "length"
	FinishReasonToolCalls    FinishReason = "tool_calls"
	FinishReasonError        FinishReason = "error"
	FinishReasonCanceled     FinishReason = "canceled"
	FinishReasonContentFilter FinishReason = "content_filter"
)

// Finish represents the end of a message with a finish reason.
type Finish struct {
	Reason  FinishReason `json:"reason"`
	Time    int64        `json:"time"`
	Message string       `json:"message,omitempty"`
}

func (Finish) contentPart() {}

// partWrapper wraps a content part with its type for JSON serialization.
type partWrapper struct {
	Type partType    `json:"type"`
	Data ContentPart `json:"data"`
}

// MarshalParts serializes content parts to JSON.
func MarshalParts(parts []ContentPart) ([]byte, error) {
	if len(parts) == 0 {
		return []byte("[]"), nil
	}

	wrappers := make([]partWrapper, len(parts))
	for i, part := range parts {
		var typ partType
		switch part.(type) {
		case TextContent:
			typ = partTypeText
		case ReasoningContent:
			typ = partTypeReasoning
		case FileContent:
			typ = partTypeFile
		case ToolCall:
			typ = partTypeToolCall
		case ToolResult:
			typ = partTypeToolResult
		case Finish:
			typ = partTypeFinish
		default:
			return nil, fmt.Errorf("unknown content part type: %T", part)
		}
		wrappers[i] = partWrapper{Type: typ, Data: part}
	}
	return json.Marshal(wrappers)
}

// UnmarshalParts deserializes content parts from JSON.
func UnmarshalParts(data []byte) ([]ContentPart, error) {
	if len(data) == 0 || string(data) == "[]" {
		return []ContentPart{}, nil
	}

	var rawWrappers []json.RawMessage
	if err := json.Unmarshal(data, &rawWrappers); err != nil {
		return nil, err
	}

	parts := make([]ContentPart, 0, len(rawWrappers))
	for _, raw := range rawWrappers {
		var wrapper struct {
			Type partType        `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(raw, &wrapper); err != nil {
			return nil, err
		}

		var part ContentPart
		switch wrapper.Type {
		case partTypeText:
			var p TextContent
			if err := json.Unmarshal(wrapper.Data, &p); err != nil {
				return nil, err
			}
			part = p
		case partTypeReasoning:
			var p ReasoningContent
			if err := json.Unmarshal(wrapper.Data, &p); err != nil {
				return nil, err
			}
			part = p
		case partTypeFile:
			var p FileContent
			if err := json.Unmarshal(wrapper.Data, &p); err != nil {
				return nil, err
			}
			part = p
		case partTypeToolCall:
			var p ToolCall
			if err := json.Unmarshal(wrapper.Data, &p); err != nil {
				return nil, err
			}
			part = p
		case partTypeToolResult:
			var p ToolResult
			if err := json.Unmarshal(wrapper.Data, &p); err != nil {
				return nil, err
			}
			part = p
		case partTypeFinish:
			var p Finish
			if err := json.Unmarshal(wrapper.Data, &p); err != nil {
				return nil, err
			}
			part = p
		default:
			return nil, fmt.Errorf("unknown content part type: %s", wrapper.Type)
		}
		parts = append(parts, part)
	}

	return parts, nil
}
