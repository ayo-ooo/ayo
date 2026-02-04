package jsonl

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/testutil"
)

func TestGolden_ReadBasicSession(t *testing.T) {
	// Copy golden file to temp location for reading
	data, err := os.ReadFile("testdata/basic_session.jsonl")
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "basic.jsonl")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	sess, messages, err := ReadSession(path)
	if err != nil {
		t.Fatalf("parse session: %v", err)
	}

	// Verify header
	if sess.ID != "sess_test123" {
		t.Errorf("id: got %q, want %q", sess.ID, "sess_test123")
	}
	if sess.AgentHandle != "@ayo" {
		t.Errorf("agent: got %q, want %q", sess.AgentHandle, "@ayo")
	}
	if sess.Title != "Test Session" {
		t.Errorf("title: got %q, want %q", sess.Title, "Test Session")
	}
	if sess.MessageCount != 3 {
		t.Errorf("message_count: got %d, want %d", sess.MessageCount, 3)
	}

	// Verify message count
	if len(messages) != 3 {
		t.Fatalf("messages: got %d, want %d", len(messages), 3)
	}

	// Verify first message
	if messages[0].Role != session.RoleUser {
		t.Errorf("msg[0].role: got %q, want %q", messages[0].Role, session.RoleUser)
	}
}

func TestGolden_ReadSessionWithTools(t *testing.T) {
	data, err := os.ReadFile("testdata/session_with_tools.jsonl")
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "tools.jsonl")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	sess, messages, err := ReadSession(path)
	if err != nil {
		t.Fatalf("parse session: %v", err)
	}

	// Verify header
	if sess.ID != "sess_tools123" {
		t.Errorf("id: got %q, want %q", sess.ID, "sess_tools123")
	}

	// Verify message count
	if len(messages) != 4 {
		t.Fatalf("messages: got %d, want %d", len(messages), 4)
	}

	// Verify tool message
	toolMsg := messages[2]
	if toolMsg.Role != session.RoleTool {
		t.Errorf("msg[2].role: got %q, want %q", toolMsg.Role, session.RoleTool)
	}
}

func TestGolden_WriteAndReadRoundtrip(t *testing.T) {
	dir := t.TempDir()

	// Create structure
	structure := NewStructure(dir)
	if err := structure.Initialize(); err != nil {
		t.Fatalf("init structure: %v", err)
	}

	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	sess := session.Session{
		ID:          "sess_roundtrip",
		AgentHandle: "@test",
		Title:       "Roundtrip Test",
		Source:      "ayo",
		CreatedAt:   now.Unix(),
		UpdatedAt:   now.Unix(),
	}

	// Write session
	w, err := NewWriter(structure, sess)
	if err != nil {
		t.Fatalf("create writer: %v", err)
	}

	// Add a message
	msg := session.Message{
		ID:        "msg_001",
		SessionID: sess.ID,
		Role:      session.RoleUser,
		CreatedAt: now.Unix(),
		UpdatedAt: now.Unix(),
		Parts:     []session.ContentPart{session.TextContent{Text: "Hello"}},
	}
	if err := w.WriteMessage(msg); err != nil {
		t.Fatalf("write message: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	// Read session back
	path := w.Path()
	readSess, messages, err := ReadSession(path)
	if err != nil {
		t.Fatalf("parse session: %v", err)
	}

	// Verify roundtrip
	if readSess.ID != "sess_roundtrip" {
		t.Errorf("id: got %q, want %q", readSess.ID, "sess_roundtrip")
	}
	if len(messages) != 1 {
		t.Fatalf("messages: got %d, want %d", len(messages), 1)
	}
	if messages[0].ID != "msg_001" {
		t.Errorf("msg id: got %q, want %q", messages[0].ID, "msg_001")
	}
}

func TestGolden_SessionHeaderMarshaling(t *testing.T) {
	g := testutil.NewGolden(t).WithDir("testdata")

	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	header := SessionHeader{
		Type:         LineTypeSession,
		ID:           "sess_marshal",
		AgentHandle:  "@test",
		Title:        "Marshal Test",
		Source:       "ayo",
		CreatedAt:    now,
		UpdatedAt:    now,
		MessageCount: 2,
	}

	data, err := json.MarshalIndent(header, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	g.Assert("session_header", data)
}

func TestGolden_ParseTypes(t *testing.T) {
	// Test parsing various content types from golden files
	tests := []struct {
		name string
		file string
		want int // expected message count
	}{
		{"basic", "basic_session.jsonl", 3},
		{"with_tools", "session_with_tools.jsonl", 4},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tc.file))
			if err != nil {
				t.Fatalf("read file: %v", err)
			}

			dir := t.TempDir()
			path := filepath.Join(dir, "test.jsonl")
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatalf("write file: %v", err)
			}

			_, messages, err := ReadSession(path)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			if len(messages) != tc.want {
				t.Errorf("messages: got %d, want %d", len(messages), tc.want)
			}
		})
	}
}
