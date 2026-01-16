package session

import (
	"encoding/json"
	"testing"
)

func TestMarshalUnmarshalParts(t *testing.T) {
	tests := []struct {
		name  string
		parts []ContentPart
	}{
		{
			name:  "empty",
			parts: []ContentPart{},
		},
		{
			name: "text only",
			parts: []ContentPart{
				TextContent{Text: "Hello, world!"},
			},
		},
		{
			name: "reasoning",
			parts: []ContentPart{
				ReasoningContent{
					Text:       "Let me think...",
					StartedAt:  1234567890,
					FinishedAt: 1234567899,
				},
			},
		},
		{
			name: "file",
			parts: []ContentPart{
				FileContent{
					Filename:  "test.png",
					Data:      []byte{0x89, 0x50, 0x4E, 0x47},
					MediaType: "image/png",
				},
			},
		},
		{
			name: "tool call",
			parts: []ContentPart{
				ToolCall{
					ID:       "call_123",
					Name:     "bash",
					Input:    `{"command": "ls -la"}`,
					Finished: true,
				},
			},
		},
		{
			name: "tool result",
			parts: []ContentPart{
				ToolResult{
					ToolCallID: "call_123",
					Name:       "bash",
					Content:    "file1.txt\nfile2.txt",
					IsError:    false,
				},
			},
		},
		{
			name: "tool result error",
			parts: []ContentPart{
				ToolResult{
					ToolCallID: "call_456",
					Name:       "bash",
					Content:    "command not found",
					IsError:    true,
				},
			},
		},
		{
			name: "finish",
			parts: []ContentPart{
				Finish{
					Reason: FinishReasonStop,
					Time:   1234567890,
				},
			},
		},
		{
			name: "complex message",
			parts: []ContentPart{
				ReasoningContent{Text: "Thinking..."},
				TextContent{Text: "I'll run a command."},
				ToolCall{ID: "call_1", Name: "bash", Input: `{"command":"pwd"}`},
				ToolResult{ToolCallID: "call_1", Name: "bash", Content: "/home/user"},
				TextContent{Text: "The current directory is /home/user"},
				Finish{Reason: FinishReasonStop, Time: 1234567890},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := MarshalParts(tt.parts)
			if err != nil {
				t.Fatalf("MarshalParts failed: %v", err)
			}

			// Verify it's valid JSON
			var raw interface{}
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("MarshalParts produced invalid JSON: %v", err)
			}

			// Unmarshal
			result, err := UnmarshalParts(data)
			if err != nil {
				t.Fatalf("UnmarshalParts failed: %v", err)
			}

			// Compare
			if len(result) != len(tt.parts) {
				t.Fatalf("got %d parts, want %d", len(result), len(tt.parts))
			}

			for i, got := range result {
				want := tt.parts[i]
				gotJSON, _ := json.Marshal(got)
				wantJSON, _ := json.Marshal(want)
				if string(gotJSON) != string(wantJSON) {
					t.Errorf("part %d mismatch:\ngot:  %s\nwant: %s", i, gotJSON, wantJSON)
				}
			}
		})
	}
}

func TestUnmarshalPartsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty slice", []byte{}},
		{"empty array", []byte("[]")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := UnmarshalParts(tt.input)
			if err != nil {
				t.Fatalf("UnmarshalParts failed: %v", err)
			}
			if len(result) != 0 {
				t.Errorf("expected empty result, got %d parts", len(result))
			}
		})
	}
}

func TestUnmarshalPartsInvalidJSON(t *testing.T) {
	_, err := UnmarshalParts([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestUnmarshalPartsUnknownType(t *testing.T) {
	input := `[{"type":"unknown","data":{}}]`
	_, err := UnmarshalParts([]byte(input))
	if err == nil {
		t.Error("expected error for unknown type")
	}
}
