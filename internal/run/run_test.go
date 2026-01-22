package run

import (
	"context"
	"os"
	"strings"
	"testing"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/agent"
)

func TestBuildMessagesOmitsEmpty(t *testing.T) {
	r := &Runner{}
	ag := agent.Agent{CombinedSystem: "", SkillsPrompt: "", Model: "m"}
	msgs := r.buildMessages(ag, "hi")
	if len(msgs) != 1 {
		t.Fatalf("expected single user message, got %d", len(msgs))
	}
	if msgs[0].Role != fantasy.MessageRoleUser {
		t.Fatalf("expected user role, got %s", msgs[0].Role)
	}
	// Check content contains "hi"
	content := getTextContent(msgs[0])
	if content != "hi" {
		t.Fatalf("expected user message to contain 'hi': got %q", content)
	}
}

func TestBuildMessagesOrdersSystemSkillsUser(t *testing.T) {
	r := &Runner{}
	ag := agent.Agent{CombinedSystem: "SYS", SkillsPrompt: "SKILLS", Model: "m"}
	msgs := r.buildMessages(ag, "hi")
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	wantRoles := []fantasy.MessageRole{fantasy.MessageRoleSystem, fantasy.MessageRoleSystem, fantasy.MessageRoleUser}
	wantContents := []string{"SYS", "SKILLS", "hi"}
	for i := range wantRoles {
		if msgs[i].Role != wantRoles[i] {
			t.Fatalf("msg %d role mismatch: got %s, want %s", i, msgs[i].Role, wantRoles[i])
		}
		content := getTextContent(msgs[i])
		if content != wantContents[i] {
			t.Fatalf("msg %d content should be %q: got %q", i, wantContents[i], content)
		}
	}
}

func TestRunChatStopsAfterEmptyModel(t *testing.T) {
	r := &Runner{sessions: make(map[string]*ChatSession)}
	_, err := r.runChat(context.Background(), agent.Agent{Model: ""}, nil)
	if err == nil {
		t.Fatalf("expected error from empty model")
	}
}

// getTextContent extracts text content from a fantasy message.
func getTextContent(msg fantasy.Message) string {
	for _, part := range msg.Content {
		if tp, ok := part.(fantasy.TextPart); ok {
			return tp.Text
		}
	}
	return ""
}

func TestBuildMessagesWithTextAttachment(t *testing.T) {
	// Create a temp text file for testing
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("file contents"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	r := &Runner{}
	ag := agent.Agent{CombinedSystem: "SYS", Model: "m"}
	msgs := r.buildMessagesWithAttachments(ag, "summarize", []string{testFile})

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages (system + user), got %d", len(msgs))
	}

	// Check user message has text with inlined file content (text files are inlined, not FilePart)
	userMsg := msgs[1]
	if userMsg.Role != fantasy.MessageRoleUser {
		t.Fatalf("expected user role, got %s", userMsg.Role)
	}

	content := getTextContent(userMsg)
	// Text files should be inlined in XML-like format
	if !strings.Contains(content, "<file path=\"test.txt\">") {
		t.Errorf("expected inlined file tag, got %q", content)
	}
	if !strings.Contains(content, "file contents") {
		t.Errorf("expected file contents in prompt, got %q", content)
	}
	if !strings.Contains(content, "summarize") {
		t.Errorf("expected original prompt, got %q", content)
	}

	// Text files should NOT create FilePart
	for _, part := range userMsg.Content {
		if _, ok := part.(fantasy.FilePart); ok {
			t.Error("text files should be inlined, not sent as FilePart")
		}
	}
}

func TestBuildMessagesWithBinaryAttachment(t *testing.T) {
	// Create a temp PNG file for testing (binary file should use FilePart)
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test.png"
	// Minimal PNG header bytes
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if err := os.WriteFile(testFile, pngData, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	r := &Runner{}
	ag := agent.Agent{Model: "m"}
	msgs := r.buildMessagesWithAttachments(ag, "describe", []string{testFile})

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	userMsg := msgs[0]
	var hasFile bool
	for _, part := range userMsg.Content {
		if fp, ok := part.(fantasy.FilePart); ok {
			hasFile = true
			if fp.Filename != "test.png" {
				t.Errorf("expected filename 'test.png', got %q", fp.Filename)
			}
			if !strings.HasPrefix(fp.MediaType, "image/png") {
				t.Errorf("expected image/png media type, got %q", fp.MediaType)
			}
		}
	}

	if !hasFile {
		t.Error("binary files should be sent as FilePart")
	}
}

func TestBuildMessagesWithMissingAttachment(t *testing.T) {
	r := &Runner{}
	ag := agent.Agent{Model: "m"}
	msgs := r.buildMessagesWithAttachments(ag, "prompt", []string{"/nonexistent/file.txt"})

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	// Check that error message was appended to prompt
	content := getTextContent(msgs[0])
	if !strings.Contains(content, "Error reading /nonexistent/file.txt") {
		t.Errorf("expected error message in prompt, got %q", content)
	}
}

func TestGetSessionMessages_NoServices(t *testing.T) {
	r := &Runner{sessions: make(map[string]*ChatSession)}

	// No services configured - should return nil, nil
	msgs, err := r.GetSessionMessages(context.Background(), "@test")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if msgs != nil {
		t.Errorf("expected nil messages, got %v", msgs)
	}
}

func TestGetSessionMessages_NoSession(t *testing.T) {
	r := &Runner{
		sessions: make(map[string]*ChatSession),
		// services would be set but we're testing the no-session case
	}

	// No session for this agent - should return nil, nil
	msgs, err := r.GetSessionMessages(context.Background(), "@nonexistent")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if msgs != nil {
		t.Errorf("expected nil messages, got %v", msgs)
	}
}

func TestGetSessionMessages_EmptySessionID(t *testing.T) {
	r := &Runner{
		sessions: map[string]*ChatSession{
			"@test": {SessionID: ""}, // Empty session ID
		},
	}

	// Session exists but has no ID - should return nil, nil
	msgs, err := r.GetSessionMessages(context.Background(), "@test")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if msgs != nil {
		t.Errorf("expected nil messages, got %v", msgs)
	}
}
