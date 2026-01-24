package ui

import (
	"strings"
	"testing"

	"github.com/alexcabrera/ayo/internal/session"
)

func TestRenderHistory_Empty(t *testing.T) {
	result := RenderHistory(nil, "@test")
	if !strings.Contains(result, "No conversation history") {
		t.Errorf("empty history should show placeholder, got: %q", result)
	}

	result = RenderHistory([]session.Message{}, "@test")
	if !strings.Contains(result, "No conversation history") {
		t.Errorf("empty slice should show placeholder, got: %q", result)
	}
}

func TestRenderHistory_SkipsSystemMessages(t *testing.T) {
	messages := []session.Message{
		{
			Role: session.RoleSystem,
			Parts: []session.ContentPart{
				session.TextContent{Text: "You are a helpful assistant"},
			},
		},
		{
			Role: session.RoleUser,
			Parts: []session.ContentPart{
				session.TextContent{Text: "Hello"},
			},
		},
	}

	result := RenderHistory(messages, "@test")

	// Should not contain system message
	if strings.Contains(result, "helpful assistant") {
		t.Error("system message should be hidden")
	}

	// Should contain user message
	if !strings.Contains(result, "Hello") {
		t.Error("user message should be visible")
	}
}

func TestRenderHistory_UserMessage(t *testing.T) {
	messages := []session.Message{
		{
			Role: session.RoleUser,
			Parts: []session.ContentPart{
				session.TextContent{Text: "What is Go?"},
			},
		},
	}

	result := RenderHistory(messages, "@ayo")

	// Should start with prompt symbol
	if !strings.Contains(result, ">") {
		t.Error("user message should have prompt symbol")
	}

	// Should contain the text
	if !strings.Contains(result, "What is Go?") {
		t.Error("user message text should be visible")
	}
}

func TestRenderHistory_AssistantMessage(t *testing.T) {
	messages := []session.Message{
		{
			Role: session.RoleAssistant,
			Parts: []session.ContentPart{
				session.TextContent{Text: "Go is a programming language."},
			},
		},
	}

	result := RenderHistory(messages, "@ayo")

	// Should contain agent handle
	if !strings.Contains(result, "@ayo") {
		t.Error("assistant message should show agent handle")
	}

	// Should contain the text
	if !strings.Contains(result, "Go is a programming language") {
		t.Error("assistant message text should be visible")
	}
}

func TestRenderHistory_WithReasoning(t *testing.T) {
	messages := []session.Message{
		{
			Role: session.RoleAssistant,
			Parts: []session.ContentPart{
				session.ReasoningContent{Text: "Let me think about this..."},
				session.TextContent{Text: "Here is my answer."},
			},
		},
	}

	result := RenderHistory(messages, "@ayo")

	// Should contain thinking indicator
	if !strings.Contains(result, "Thinking") {
		t.Error("reasoning should show thinking label")
	}

	// Should contain reasoning text
	if !strings.Contains(result, "Let me think") {
		t.Error("reasoning text should be visible")
	}
}

func TestRenderHistory_WithToolCall(t *testing.T) {
	messages := []session.Message{
		{
			Role: session.RoleAssistant,
			Parts: []session.ContentPart{
				session.ToolCall{
					ID:    "call_123",
					Name:  "bash",
					Input: `{"command":"ls -la","description":"List files"}`,
				},
			},
		},
	}

	result := RenderHistory(messages, "@ayo")

	// Should contain bash indicator
	if !strings.Contains(result, "bash") {
		t.Error("tool call should show tool name")
	}

	// Should contain description
	if !strings.Contains(result, "List files") {
		t.Error("tool call should show description")
	}

	// Should contain command
	if !strings.Contains(result, "ls -la") {
		t.Error("tool call should show command")
	}
}

func TestRenderHistory_WithToolResult(t *testing.T) {
	messages := []session.Message{
		{
			Role: session.RoleTool,
			Parts: []session.ContentPart{
				session.ToolResult{
					ToolCallID: "call_123",
					Content:    "file1.txt\nfile2.txt",
					IsError:    false,
				},
			},
		},
	}

	result := RenderHistory(messages, "@ayo")

	// Should contain success indicator
	if !strings.Contains(result, IconSuccess) {
		t.Error("successful tool result should show success icon")
	}

	// Should contain output
	if !strings.Contains(result, "file1.txt") {
		t.Error("tool result should show output")
	}
}

func TestRenderHistory_ErrorToolResult(t *testing.T) {
	messages := []session.Message{
		{
			Role: session.RoleTool,
			Parts: []session.ContentPart{
				session.ToolResult{
					ToolCallID: "call_123",
					Content:    "command not found",
					IsError:    true,
				},
			},
		},
	}

	result := RenderHistory(messages, "@ayo")

	// Should contain error indicator
	if !strings.Contains(result, IconError) {
		t.Error("error tool result should show error icon")
	}
}

func TestRenderHistory_FullConversation(t *testing.T) {
	messages := []session.Message{
		{
			Role: session.RoleSystem,
			Parts: []session.ContentPart{
				session.TextContent{Text: "You are a helpful assistant."},
			},
		},
		{
			Role: session.RoleUser,
			Parts: []session.ContentPart{
				session.TextContent{Text: "List files"},
			},
		},
		{
			Role: session.RoleAssistant,
			Parts: []session.ContentPart{
				session.TextContent{Text: "I'll list the files for you."},
				session.ToolCall{
					ID:    "call_1",
					Name:  "bash",
					Input: `{"command":"ls","description":"List files"}`,
				},
			},
		},
		{
			Role: session.RoleTool,
			Parts: []session.ContentPart{
				session.ToolResult{
					ToolCallID: "call_1",
					Content:    "README.md\ngo.mod",
				},
			},
		},
		{
			Role: session.RoleAssistant,
			Parts: []session.ContentPart{
				session.TextContent{Text: "Found 2 files."},
			},
		},
	}

	result := RenderHistory(messages, "@ayo")

	// Verify all expected content is present
	checks := []string{
		"List files",       // User message
		"@ayo",             // Agent handle
		"I'll list",        // First assistant response
		"bash",             // Tool name
		"ls",               // Command
		"README.md",        // Tool output
		"Found 2 files",    // Second assistant response
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("full conversation should contain %q", check)
		}
	}

	// System message should NOT be present
	if strings.Contains(result, "helpful assistant") {
		t.Error("system message should be hidden")
	}
}

func TestRenderHistory_LongReasoningTruncated(t *testing.T) {
	// Create long reasoning with many lines
	longReasoning := strings.Repeat("Line of reasoning.\n", 20)

	messages := []session.Message{
		{
			Role: session.RoleAssistant,
			Parts: []session.ContentPart{
				session.ReasoningContent{Text: longReasoning},
			},
		},
	}

	result := RenderHistory(messages, "@ayo")

	// Should indicate truncation
	if !strings.Contains(result, "more lines") {
		t.Error("long reasoning should be truncated with indicator")
	}
}

func TestRenderHistory_LongToolOutputTruncated(t *testing.T) {
	// Create long tool output with many lines
	longOutput := strings.Repeat("output line\n", 20)

	messages := []session.Message{
		{
			Role: session.RoleTool,
			Parts: []session.ContentPart{
				session.ToolResult{
					ToolCallID: "call_1",
					Content:    longOutput,
				},
			},
		},
	}

	result := RenderHistory(messages, "@ayo")

	// Should indicate truncation
	if !strings.Contains(result, "more lines") {
		t.Error("long tool output should be truncated with indicator")
	}
}
