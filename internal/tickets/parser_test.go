package tickets

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseAndSerialize(t *testing.T) {
	content := `---
id: test-1234
status: open
deps: []
links: []
created: 2026-02-12T10:30:00Z
type: task
priority: 2
assignee: "@coder"
parent: test-0001
tags: [backend, auth]
---
# Implement JWT authentication

Add JWT-based authentication to the API server.

## Description

- Generate tokens on login
- Validate tokens on protected routes

## Notes

### 2026-02-12T11:00:00Z
Started implementation, creating auth middleware first.

### 2026-02-12T11:30:00Z
Middleware complete, working on token generation.
`

	ticket, err := ParseBytes([]byte(content), "/tmp/test-1234.md")
	if err != nil {
		t.Fatalf("ParseBytes failed: %v", err)
	}

	// Check frontmatter fields
	if ticket.ID != "test-1234" {
		t.Errorf("ID = %q, want %q", ticket.ID, "test-1234")
	}
	if ticket.Status != StatusOpen {
		t.Errorf("Status = %q, want %q", ticket.Status, StatusOpen)
	}
	if ticket.Type != TypeTask {
		t.Errorf("Type = %q, want %q", ticket.Type, TypeTask)
	}
	if ticket.Priority != 2 {
		t.Errorf("Priority = %d, want %d", ticket.Priority, 2)
	}
	if ticket.Assignee != "@coder" {
		t.Errorf("Assignee = %q, want %q", ticket.Assignee, "@coder")
	}
	if ticket.Parent != "test-0001" {
		t.Errorf("Parent = %q, want %q", ticket.Parent, "test-0001")
	}
	if len(ticket.Tags) != 2 || ticket.Tags[0] != "backend" {
		t.Errorf("Tags = %v, want [backend auth]", ticket.Tags)
	}

	// Check parsed body
	if ticket.Title != "Implement JWT authentication" {
		t.Errorf("Title = %q, want %q", ticket.Title, "Implement JWT authentication")
	}
	if len(ticket.Notes) != 2 {
		t.Errorf("Notes count = %d, want %d", len(ticket.Notes), 2)
	}

	// Test round-trip serialization
	serialized, err := Serialize(ticket)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	ticket2, err := ParseBytes(serialized, "/tmp/test-1234.md")
	if err != nil {
		t.Fatalf("ParseBytes after Serialize failed: %v", err)
	}

	if ticket2.ID != ticket.ID {
		t.Errorf("Round-trip ID = %q, want %q", ticket2.ID, ticket.ID)
	}
	if ticket2.Title != ticket.Title {
		t.Errorf("Round-trip Title = %q, want %q", ticket2.Title, ticket.Title)
	}
	if len(ticket2.Notes) != len(ticket.Notes) {
		t.Errorf("Round-trip Notes count = %d, want %d", len(ticket2.Notes), len(ticket.Notes))
	}
}

func TestParseInvalidFormat(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"no frontmatter", "# Just a title\n\nSome content"},
		{"single delimiter", "---\nid: test\nstatus: open"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseBytes([]byte(tt.content), "/tmp/test.md")
			if err == nil {
				t.Error("Expected error for invalid format")
			}
		})
	}
}

func TestStatusValid(t *testing.T) {
	validStatuses := []Status{StatusOpen, StatusInProgress, StatusBlocked, StatusClosed}
	for _, s := range validStatuses {
		if !s.Valid() {
			t.Errorf("Status %q should be valid", s)
		}
	}

	invalid := Status("invalid")
	if invalid.Valid() {
		t.Error("Invalid status should not be valid")
	}
}

func TestTypeValid(t *testing.T) {
	validTypes := []Type{TypeEpic, TypeFeature, TypeTask, TypeBug, TypeChore}
	for _, typ := range validTypes {
		if !typ.Valid() {
			t.Errorf("Type %q should be valid", typ)
		}
	}

	invalid := Type("invalid")
	if invalid.Valid() {
		t.Error("Invalid type should not be valid")
	}
}

func TestValidatePriority(t *testing.T) {
	for i := 0; i <= 4; i++ {
		if !ValidatePriority(i) {
			t.Errorf("Priority %d should be valid", i)
		}
	}

	invalidPriorities := []int{-1, 5, 100}
	for _, p := range invalidPriorities {
		if ValidatePriority(p) {
			t.Errorf("Priority %d should not be valid", p)
		}
	}
}

func TestGenerateID(t *testing.T) {
	tmpDir := t.TempDir()
	ticketsDir := filepath.Join(tmpDir, "test-session", ".tickets")
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	id, err := GenerateID(ticketsDir)
	if err != nil {
		t.Fatalf("GenerateID failed: %v", err)
	}

	// Should have format prefix-xxxx
	if len(id) < 6 { // at least "xx-xxxx"
		t.Errorf("ID too short: %q", id)
	}
}

func TestGenerateUniqueID(t *testing.T) {
	tmpDir := t.TempDir()
	ticketsDir := filepath.Join(tmpDir, "test-session", ".tickets")
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id, err := GenerateUniqueID(ticketsDir)
		if err != nil {
			t.Fatalf("GenerateUniqueID failed: %v", err)
		}
		if ids[id] {
			t.Errorf("Duplicate ID generated: %q", id)
		}
		ids[id] = true

		// Create a file so next iteration has to avoid it
		f, _ := os.Create(filepath.Join(ticketsDir, id+".md"))
		f.Close()
	}
}

func TestDerivePrefix(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/home/user/.ayo/sessions/ses_abc123/.tickets", "abc1"},
		{"/home/user/.ayo/sessions/myproject/.tickets", "mypr"},
		{"/tmp/x/.tickets", "tk"}, // Too short, falls back
	}

	for _, tt := range tests {
		got := derivePrefix(tt.path)
		if len(got) < 2 {
			t.Errorf("derivePrefix(%q) = %q, too short", tt.path, got)
		}
	}
}

func TestParseNotesTimestamps(t *testing.T) {
	content := `---
id: test-1234
status: open
deps: []
links: []
created: 2026-02-12T10:30:00Z
type: task
priority: 2
---
# Test ticket

## Notes

### 2026-02-12T11:00:00Z
First note content.

### 2026-02-12T12:00:00+05:00
Second note with timezone.
`

	ticket, err := ParseBytes([]byte(content), "/tmp/test.md")
	if err != nil {
		t.Fatalf("ParseBytes failed: %v", err)
	}

	if len(ticket.Notes) != 2 {
		t.Fatalf("Expected 2 notes, got %d", len(ticket.Notes))
	}

	// Check first note timestamp
	expected1 := time.Date(2026, 2, 12, 11, 0, 0, 0, time.UTC)
	if !ticket.Notes[0].Timestamp.Equal(expected1) {
		t.Errorf("Note 1 timestamp = %v, want %v", ticket.Notes[0].Timestamp, expected1)
	}

	if ticket.Notes[0].Content != "First note content." {
		t.Errorf("Note 1 content = %q", ticket.Notes[0].Content)
	}
}
