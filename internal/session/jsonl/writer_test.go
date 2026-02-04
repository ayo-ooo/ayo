package jsonl

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/session"
)

func TestWriter_NewWriter(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sess := session.Session{
		ID:          "test-session-123",
		AgentHandle: "@ayo",
		Title:       "Test Session",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	w, err := NewWriter(structure, sess)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	defer w.Close()

	// Check file exists
	if _, err := os.Stat(w.Path()); os.IsNotExist(err) {
		t.Error("session file not created")
	}

	// Check path format
	expected := filepath.Join(tmpDir, "ayo", time.Now().UTC().Format("2006-01"), "test-session-123.jsonl")
	if w.Path() != expected {
		t.Errorf("Path() = %v, want %v", w.Path(), expected)
	}
}

func TestWriter_WriteMessage(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sess := session.Session{
		ID:          "test-session-456",
		AgentHandle: "@ayo",
		Title:       "Test Session",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	w, err := NewWriter(structure, sess)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	// Write a message
	msg := session.Message{
		ID:        "msg-001",
		SessionID: sess.ID,
		Role:      session.RoleUser,
		Parts: []session.ContentPart{
			session.TextContent{Text: "Hello, world!"},
		},
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	if err := w.WriteMessage(msg); err != nil {
		t.Fatalf("WriteMessage() error = %v", err)
	}

	if w.MessageCount() != 1 {
		t.Errorf("MessageCount() = %v, want 1", w.MessageCount())
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Read back and verify
	reader, err := NewReader(w.Path())
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}
	defer reader.Close()

	readMsg, err := reader.NextMessage()
	if err != nil {
		t.Fatalf("NextMessage() error = %v", err)
	}
	if readMsg == nil {
		t.Fatal("NextMessage() returned nil")
	}

	if readMsg.ID != "msg-001" {
		t.Errorf("message ID = %v, want msg-001", readMsg.ID)
	}
	if readMsg.Role != session.RoleUser {
		t.Errorf("message Role = %v, want user", readMsg.Role)
	}
	if readMsg.TextContent() != "Hello, world!" {
		t.Errorf("message TextContent() = %v, want 'Hello, world!'", readMsg.TextContent())
	}
}

func TestWriter_MultipleMessages(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sess := session.Session{
		ID:          "test-session-multi",
		AgentHandle: "@ayo",
		Title:       "Multi Message Session",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	w, err := NewWriter(structure, sess)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	messages := []session.Message{
		{
			ID:        "msg-001",
			SessionID: sess.ID,
			Role:      session.RoleUser,
			Parts:     []session.ContentPart{session.TextContent{Text: "Hello"}},
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		},
		{
			ID:        "msg-002",
			SessionID: sess.ID,
			Role:      session.RoleAssistant,
			Parts:     []session.ContentPart{session.TextContent{Text: "Hi there!"}},
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		},
		{
			ID:        "msg-003",
			SessionID: sess.ID,
			Role:      session.RoleUser,
			Parts:     []session.ContentPart{session.TextContent{Text: "How are you?"}},
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		},
	}

	for _, msg := range messages {
		if err := w.WriteMessage(msg); err != nil {
			t.Fatalf("WriteMessage() error = %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Read all messages
	_, readMsgs, err := ReadSession(w.Path())
	if err != nil {
		t.Fatalf("ReadSession() error = %v", err)
	}

	if len(readMsgs) != 3 {
		t.Errorf("got %v messages, want 3", len(readMsgs))
	}
}

func TestWriter_ToolCallParts(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sess := session.Session{
		ID:          "test-session-tools",
		AgentHandle: "@ayo",
		Title:       "Tool Call Session",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	w, err := NewWriter(structure, sess)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	msg := session.Message{
		ID:        "msg-001",
		SessionID: sess.ID,
		Role:      session.RoleAssistant,
		Parts: []session.ContentPart{
			session.ToolCall{
				ID:    "call-001",
				Name:  "bash",
				Input: `{"command": "ls -la", "description": "List files"}`,
			},
		},
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	if err := w.WriteMessage(msg); err != nil {
		t.Fatalf("WriteMessage() error = %v", err)
	}

	resultMsg := session.Message{
		ID:        "msg-002",
		SessionID: sess.ID,
		Role:      session.RoleTool,
		Parts: []session.ContentPart{
			session.ToolResult{
				ToolCallID: "call-001",
				Name:       "bash",
				Content:    "file1.txt\nfile2.txt",
				IsError:    false,
			},
		},
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	if err := w.WriteMessage(resultMsg); err != nil {
		t.Fatalf("WriteMessage() error = %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Read and verify
	_, readMsgs, err := ReadSession(w.Path())
	if err != nil {
		t.Fatalf("ReadSession() error = %v", err)
	}

	if len(readMsgs) != 2 {
		t.Fatalf("got %v messages, want 2", len(readMsgs))
	}

	// Check tool call
	calls := readMsgs[0].ToolCalls()
	if len(calls) != 1 {
		t.Fatalf("got %v tool calls, want 1", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("tool call name = %v, want bash", calls[0].Name)
	}

	// Check tool result
	results := readMsgs[1].ToolResults()
	if len(results) != 1 {
		t.Fatalf("got %v tool results, want 1", len(results))
	}
	if results[0].ToolCallID != "call-001" {
		t.Errorf("tool result call_id = %v, want call-001", results[0].ToolCallID)
	}
}

func TestWriter_Finish(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sess := session.Session{
		ID:          "test-session-finish",
		AgentHandle: "@ayo",
		Title:       "Finish Test",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	w, err := NewWriter(structure, sess)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	if err := w.Finish(`{"result": "success"}`); err != nil {
		t.Fatalf("Finish() error = %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Read header and verify
	header, err := ReadSessionHeader(w.Path())
	if err != nil {
		t.Fatalf("ReadSessionHeader() error = %v", err)
	}

	if header.FinishedAt == nil {
		t.Error("FinishedAt should be set")
	}
	if header.StructuredOutput == nil || *header.StructuredOutput != `{"result": "success"}` {
		t.Error("StructuredOutput not set correctly")
	}
}

func TestReader_NewReader(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sess := session.Session{
		ID:          "test-reader-session",
		AgentHandle: "@ayo",
		Title:       "Reader Test",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	w, err := NewWriter(structure, sess)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	w.Close()

	reader, err := NewReader(w.Path())
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}
	defer reader.Close()

	if reader.Header().ID != sess.ID {
		t.Errorf("Header().ID = %v, want %v", reader.Header().ID, sess.ID)
	}

	readSess := reader.Session()
	if readSess.Title != "Reader Test" {
		t.Errorf("Session().Title = %v, want Reader Test", readSess.Title)
	}
}

func TestReader_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.jsonl")

	// Create empty file
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := NewReader(path)
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestReadSessionHeader(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sess := session.Session{
		ID:          "test-header-read",
		AgentHandle: "@custom",
		Title:       "Header Read Test",
		Source:      "ayo",
		ChainDepth:  2,
		ChainSource: "@parent",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	w, err := NewWriter(structure, sess)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	w.Close()

	header, err := ReadSessionHeader(w.Path())
	if err != nil {
		t.Fatalf("ReadSessionHeader() error = %v", err)
	}

	if header.AgentHandle != "@custom" {
		t.Errorf("AgentHandle = %v, want @custom", header.AgentHandle)
	}
	if header.ChainDepth != 2 {
		t.Errorf("ChainDepth = %v, want 2", header.ChainDepth)
	}
	if header.ChainSource != "@parent" {
		t.Errorf("ChainSource = %v, want @parent", header.ChainSource)
	}
}

func TestListSessionHeaders(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sessions := []session.Session{
		{
			ID:          "session-1",
			AgentHandle: "@ayo",
			Title:       "Session 1",
			Source:      "ayo",
			CreatedAt:   time.Now().Add(-2 * time.Hour).Unix(),
			UpdatedAt:   time.Now().Unix(),
		},
		{
			ID:          "session-2",
			AgentHandle: "@ayo",
			Title:       "Session 2",
			Source:      "ayo",
			CreatedAt:   time.Now().Add(-1 * time.Hour).Unix(),
			UpdatedAt:   time.Now().Unix(),
		},
		{
			ID:          "session-3",
			AgentHandle: "@custom",
			Title:       "Session 3",
			Source:      "ayo",
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		},
	}

	for _, sess := range sessions {
		w, err := NewWriter(structure, sess)
		if err != nil {
			t.Fatalf("NewWriter() error = %v", err)
		}
		w.Close()
	}

	headers, err := ListSessionHeaders(structure)
	if err != nil {
		t.Fatalf("ListSessionHeaders() error = %v", err)
	}

	if len(headers) != 3 {
		t.Errorf("got %v headers, want 3", len(headers))
	}

	// Should be sorted by created desc
	if headers[0].ID != "session-3" {
		t.Errorf("first header ID = %v, want session-3 (newest)", headers[0].ID)
	}
}

func TestFindSessionByID(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sess := session.Session{
		ID:          "find-me-session",
		AgentHandle: "@ayo",
		Title:       "Find Me",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	w, err := NewWriter(structure, sess)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	w.Close()

	header, path, err := FindSessionByID(structure, "find-me-session")
	if err != nil {
		t.Fatalf("FindSessionByID() error = %v", err)
	}

	if header.Title != "Find Me" {
		t.Errorf("Title = %v, want Find Me", header.Title)
	}
	if path == "" {
		t.Error("path should not be empty")
	}
}

func TestFindSessionByID_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)
	structure.Initialize()

	_, _, err := FindSessionByID(structure, "nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("error = %v, want ErrSessionNotFound", err)
	}
}

func TestDeleteSession(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sess := session.Session{
		ID:          "delete-me",
		AgentHandle: "@ayo",
		Title:       "Delete Me",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	w, err := NewWriter(structure, sess)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	path := w.Path()
	w.Close()

	// Verify exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file should exist before delete")
	}

	// Delete
	if err := DeleteSession(structure, "delete-me"); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}

	// Verify deleted
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should not exist after delete")
	}
}

func TestSessionCount(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	// Create 5 sessions
	for i := 0; i < 5; i++ {
		sess := session.Session{
			ID:          "count-session-" + string(rune('a'+i)),
			AgentHandle: "@ayo",
			Title:       "Count Session",
			Source:      "ayo",
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}
		w, _ := NewWriter(structure, sess)
		w.Close()
	}

	count, err := SessionCount(structure)
	if err != nil {
		t.Fatalf("SessionCount() error = %v", err)
	}

	if count != 5 {
		t.Errorf("SessionCount() = %v, want 5", count)
	}
}

func TestOpenWriter_Append(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sess := session.Session{
		ID:          "append-session",
		AgentHandle: "@ayo",
		Title:       "Append Test",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	// Create and write initial message
	w, err := NewWriter(structure, sess)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	msg1 := session.Message{
		ID:        "msg-001",
		SessionID: sess.ID,
		Role:      session.RoleUser,
		Parts:     []session.ContentPart{session.TextContent{Text: "First"}},
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	w.WriteMessage(msg1)
	path := w.Path()
	w.Close()

	// Reopen and append
	w2, err := OpenWriter(path)
	if err != nil {
		t.Fatalf("OpenWriter() error = %v", err)
	}

	if w2.MessageCount() != 1 {
		t.Errorf("MessageCount() = %v, want 1", w2.MessageCount())
	}

	msg2 := session.Message{
		ID:        "msg-002",
		SessionID: sess.ID,
		Role:      session.RoleAssistant,
		Parts:     []session.ContentPart{session.TextContent{Text: "Second"}},
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	w2.WriteMessage(msg2)
	w2.Close()

	// Verify both messages
	_, msgs, err := ReadSession(path)
	if err != nil {
		t.Fatalf("ReadSession() error = %v", err)
	}

	if len(msgs) != 2 {
		t.Fatalf("got %v messages, want 2", len(msgs))
	}
	if msgs[0].TextContent() != "First" {
		t.Errorf("first message = %v, want First", msgs[0].TextContent())
	}
	if msgs[1].TextContent() != "Second" {
		t.Errorf("second message = %v, want Second", msgs[1].TextContent())
	}
}
