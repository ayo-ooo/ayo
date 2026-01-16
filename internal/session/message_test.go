package session

import (
	"testing"

	"charm.land/fantasy"
)

func TestMessageTextContent(t *testing.T) {
	m := Message{
		Parts: []ContentPart{
			ReasoningContent{Text: "thinking..."},
			TextContent{Text: "Hello"},
		},
	}

	if got := m.TextContent(); got != "Hello" {
		t.Errorf("TextContent() = %q, want %q", got, "Hello")
	}
}

func TestMessageReasoningText(t *testing.T) {
	m := Message{
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
			ReasoningContent{Text: "thinking..."},
		},
	}

	if got := m.ReasoningText(); got != "thinking..." {
		t.Errorf("ReasoningText() = %q, want %q", got, "thinking...")
	}
}

func TestMessageToolCalls(t *testing.T) {
	m := Message{
		Parts: []ContentPart{
			TextContent{Text: "Let me check"},
			ToolCall{ID: "1", Name: "bash"},
			ToolCall{ID: "2", Name: "read"},
		},
	}

	calls := m.ToolCalls()
	if len(calls) != 2 {
		t.Fatalf("got %d tool calls, want 2", len(calls))
	}
	if calls[0].ID != "1" || calls[1].ID != "2" {
		t.Errorf("unexpected tool call IDs")
	}
}

func TestMessageToolResults(t *testing.T) {
	m := Message{
		Parts: []ContentPart{
			ToolResult{ToolCallID: "1", Content: "output1"},
			ToolResult{ToolCallID: "2", Content: "output2", IsError: true},
		},
	}

	results := m.ToolResults()
	if len(results) != 2 {
		t.Fatalf("got %d tool results, want 2", len(results))
	}
	if !results[1].IsError {
		t.Error("expected second result to be error")
	}
}

func TestMessageIsFinished(t *testing.T) {
	m := Message{Parts: []ContentPart{TextContent{Text: "hi"}}}
	if m.IsFinished() {
		t.Error("expected not finished")
	}

	m.Parts = append(m.Parts, Finish{Reason: FinishReasonStop})
	if !m.IsFinished() {
		t.Error("expected finished")
	}
}

func TestMessageFinishPart(t *testing.T) {
	m := Message{Parts: []ContentPart{TextContent{Text: "hi"}}}
	if m.FinishPart() != nil {
		t.Error("expected nil finish part")
	}

	m.Parts = append(m.Parts, Finish{Reason: FinishReasonStop, Time: 123})
	f := m.FinishPart()
	if f == nil {
		t.Fatal("expected finish part")
	}
	if f.Reason != FinishReasonStop {
		t.Errorf("got reason %v, want %v", f.Reason, FinishReasonStop)
	}
}

func TestToFantasyMessage(t *testing.T) {
	m := Message{
		Role: RoleAssistant,
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
			ToolCall{ID: "1", Name: "bash", Input: `{"cmd":"ls"}`},
		},
	}

	fm := m.ToFantasyMessage()

	if fm.Role != fantasy.MessageRoleAssistant {
		t.Errorf("role = %v, want %v", fm.Role, fantasy.MessageRoleAssistant)
	}

	if len(fm.Content) != 2 {
		t.Fatalf("got %d parts, want 2", len(fm.Content))
	}

	if tp, ok := fm.Content[0].(fantasy.TextPart); !ok || tp.Text != "Hello" {
		t.Errorf("first part: got %T %+v, want TextPart{Hello}", fm.Content[0], fm.Content[0])
	}

	if tc, ok := fm.Content[1].(fantasy.ToolCallPart); !ok || tc.ToolCallID != "1" {
		t.Errorf("second part: got %T %+v, want ToolCallPart", fm.Content[1], fm.Content[1])
	}
}

func TestToFantasyMessageRoles(t *testing.T) {
	tests := []struct {
		role     MessageRole
		expected fantasy.MessageRole
	}{
		{RoleSystem, fantasy.MessageRoleSystem},
		{RoleUser, fantasy.MessageRoleUser},
		{RoleAssistant, fantasy.MessageRoleAssistant},
		{RoleTool, fantasy.MessageRoleTool},
	}

	for _, tt := range tests {
		m := Message{Role: tt.role, Parts: []ContentPart{TextContent{Text: "test"}}}
		fm := m.ToFantasyMessage()
		if fm.Role != tt.expected {
			t.Errorf("role %v: got %v, want %v", tt.role, fm.Role, tt.expected)
		}
	}
}

func TestFromFantasyMessageSimple(t *testing.T) {
	fm := fantasy.Message{
		Role: fantasy.MessageRoleUser,
		Content: []fantasy.MessagePart{
			fantasy.TextPart{Text: "Hello, world!"},
		},
	}

	msgs := FromFantasyMessage(fm, "session-1", "gpt-4", "openai")

	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}

	m := msgs[0]
	if m.Role != RoleUser {
		t.Errorf("role = %v, want %v", m.Role, RoleUser)
	}
	if m.SessionID != "session-1" {
		t.Errorf("sessionID = %v, want session-1", m.SessionID)
	}
	if m.Model != "gpt-4" {
		t.Errorf("model = %v, want gpt-4", m.Model)
	}
	if len(m.Parts) != 1 {
		t.Fatalf("got %d parts, want 1", len(m.Parts))
	}
	if tc, ok := m.Parts[0].(TextContent); !ok || tc.Text != "Hello, world!" {
		t.Errorf("part = %T %+v, want TextContent{Hello, world!}", m.Parts[0], m.Parts[0])
	}
}

func TestFromFantasyMessageToolResult(t *testing.T) {
	fm := fantasy.Message{
		Role: fantasy.MessageRoleTool,
		Content: []fantasy.MessagePart{
			fantasy.ToolResultPart{
				ToolCallID: "call-1",
				Output:     fantasy.ToolResultOutputContentText{Text: "output"},
			},
		},
	}

	msgs := FromFantasyMessage(fm, "s1", "", "")

	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if msgs[0].Role != RoleTool {
		t.Errorf("role = %v, want %v", msgs[0].Role, RoleTool)
	}

	results := msgs[0].ToolResults()
	if len(results) != 1 {
		t.Fatalf("got %d tool results, want 1", len(results))
	}
	if results[0].Content != "output" {
		t.Errorf("content = %q, want %q", results[0].Content, "output")
	}
}

func TestFromFantasyMessageToolError(t *testing.T) {
	fm := fantasy.Message{
		Role: fantasy.MessageRoleTool,
		Content: []fantasy.MessagePart{
			fantasy.ToolResultPart{
				ToolCallID: "call-1",
				Output:     fantasy.ToolResultOutputContentError{Error: errorString("failed")},
			},
		},
	}

	msgs := FromFantasyMessage(fm, "s1", "", "")
	results := msgs[0].ToolResults()

	if !results[0].IsError {
		t.Error("expected error result")
	}
	if results[0].Content != "failed" {
		t.Errorf("content = %q, want %q", results[0].Content, "failed")
	}
}
