package zettelkasten

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
		check   func(*testing.T, *MemoryFile)
	}{
		{
			name: "valid memory file",
			input: `+++
id = "mem-abc123"
created = 2024-01-15T10:30:00Z
category = "fact"
topics = ["golang", "testing"]
+++

User prefers tabs over spaces.
`,
			check: func(t *testing.T, mf *MemoryFile) {
				if mf.Frontmatter.ID != "mem-abc123" {
					t.Errorf("ID = %q, want %q", mf.Frontmatter.ID, "mem-abc123")
				}
				if mf.Frontmatter.Category != "fact" {
					t.Errorf("Category = %q, want %q", mf.Frontmatter.Category, "fact")
				}
				if mf.Frontmatter.Status != "active" {
					t.Errorf("Status = %q, want %q (default)", mf.Frontmatter.Status, "active")
				}
				if mf.Frontmatter.Confidence != 1.0 {
					t.Errorf("Confidence = %f, want %f (default)", mf.Frontmatter.Confidence, 1.0)
				}
				if len(mf.Frontmatter.Topics) != 2 {
					t.Errorf("Topics len = %d, want 2", len(mf.Frontmatter.Topics))
				}
				if mf.Content != "User prefers tabs over spaces." {
					t.Errorf("Content = %q, want %q", mf.Content, "User prefers tabs over spaces.")
				}
			},
		},
		{
			name: "full frontmatter",
			input: `+++
id = "mem-full"
created = 2024-01-15T10:30:00Z
updated = 2024-01-16T12:00:00Z
category = "preference"
status = "active"
topics = ["editor"]
confidence = 0.85

[source]
session_id = "sess-123"
message_id = "msg-456"

[scope]
agent = "@ayo"
path = "/home/user/project"

[access]
last_accessed = 2024-01-17T08:00:00Z
access_count = 5

[supersession]
supersedes = "mem-old"
reason = "Updated preference"

[links]
related = ["mem-abc", "mem-def"]

[unclear]
flagged = true
reason = "Needs clarification"
+++

User prefers VS Code for Go development.
`,
			check: func(t *testing.T, mf *MemoryFile) {
				fm := mf.Frontmatter
				if fm.ID != "mem-full" {
					t.Errorf("ID = %q, want %q", fm.ID, "mem-full")
				}
				if fm.Confidence != 0.85 {
					t.Errorf("Confidence = %f, want %f", fm.Confidence, 0.85)
				}
				if fm.Source.SessionID != "sess-123" {
					t.Errorf("Source.SessionID = %q, want %q", fm.Source.SessionID, "sess-123")
				}
				if fm.Scope.Agent != "@ayo" {
					t.Errorf("Scope.Agent = %q, want %q", fm.Scope.Agent, "@ayo")
				}
				if fm.Access.AccessCount != 5 {
					t.Errorf("Access.AccessCount = %d, want 5", fm.Access.AccessCount)
				}
				if fm.Supersession.Supersedes != "mem-old" {
					t.Errorf("Supersession.Supersedes = %q, want %q", fm.Supersession.Supersedes, "mem-old")
				}
				if len(fm.Links.Related) != 2 {
					t.Errorf("Links.Related len = %d, want 2", len(fm.Links.Related))
				}
				if !fm.Unclear.Flagged {
					t.Error("Unclear.Flagged = false, want true")
				}
			},
		},
		{
			name:    "missing frontmatter delimiter",
			input:   "Just some text without frontmatter",
			wantErr: ErrNoFrontmatter,
		},
		{
			name: "missing closing delimiter",
			input: `+++
id = "mem-test"
created = 2024-01-15T10:30:00Z
category = "fact"

Content without closing delimiter.
`,
			wantErr: ErrMissingDelimiter,
		},
		{
			name: "missing ID",
			input: `+++
created = 2024-01-15T10:30:00Z
category = "fact"
+++

Content
`,
			wantErr: ErrMissingID,
		},
		{
			name: "missing created",
			input: `+++
id = "mem-test"
category = "fact"
+++

Content
`,
			wantErr: ErrMissingCreated,
		},
		{
			name: "missing category",
			input: `+++
id = "mem-test"
created = 2024-01-15T10:30:00Z
+++

Content
`,
			wantErr: ErrMissingCategory,
		},
		{
			name: "invalid category",
			input: `+++
id = "mem-test"
created = 2024-01-15T10:30:00Z
category = "invalid"
+++

Content
`,
			wantErr: ErrInvalidCategory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf, err := Parse([]byte(tt.input))

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errorContains(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.check != nil {
				tt.check(t, mf)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	tmpDir := t.TempDir()

	content := `+++
id = "mem-file"
created = 2024-01-15T10:30:00Z
category = "fact"
+++

File content here.
`
	path := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	mf, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	if mf.Frontmatter.ID != "mem-file" {
		t.Errorf("ID = %q, want %q", mf.Frontmatter.ID, "mem-file")
	}
	if mf.Content != "File content here." {
		t.Errorf("Content = %q, want %q", mf.Content, "File content here.")
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path.md")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestMemoryFile_Marshal(t *testing.T) {
	created := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mf := &MemoryFile{
		Frontmatter: Frontmatter{
			ID:         "mem-test",
			Created:    created,
			Updated:    created,
			Category:   "preference",
			Status:     "active",
			Topics:     []string{"testing"},
			Confidence: 0.9,
		},
		Content: "Test content.",
	}

	data, err := mf.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Parse it back
	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse marshaled data: %v", err)
	}

	if parsed.Frontmatter.ID != mf.Frontmatter.ID {
		t.Errorf("ID = %q, want %q", parsed.Frontmatter.ID, mf.Frontmatter.ID)
	}
	if parsed.Frontmatter.Category != mf.Frontmatter.Category {
		t.Errorf("Category = %q, want %q", parsed.Frontmatter.Category, mf.Frontmatter.Category)
	}
	if parsed.Content != mf.Content {
		t.Errorf("Content = %q, want %q", parsed.Content, mf.Content)
	}
}

func TestMemoryFile_WriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "output.md")

	mf := NewMemoryFile("mem-write", "fact", "Written content.")
	if err := mf.WriteFile(path); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Read it back
	parsed, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	if parsed.Frontmatter.ID != "mem-write" {
		t.Errorf("ID = %q, want %q", parsed.Frontmatter.ID, "mem-write")
	}
	if parsed.Content != "Written content." {
		t.Errorf("Content = %q, want %q", parsed.Content, "Written content.")
	}
}

func TestNewMemoryFile(t *testing.T) {
	mf := NewMemoryFile("mem-new", "correction", "Test content")

	if mf.Frontmatter.ID != "mem-new" {
		t.Errorf("ID = %q, want %q", mf.Frontmatter.ID, "mem-new")
	}
	if mf.Frontmatter.Category != "correction" {
		t.Errorf("Category = %q, want %q", mf.Frontmatter.Category, "correction")
	}
	if mf.Frontmatter.Status != "active" {
		t.Errorf("Status = %q, want %q", mf.Frontmatter.Status, "active")
	}
	if mf.Frontmatter.Confidence != 1.0 {
		t.Errorf("Confidence = %f, want %f", mf.Frontmatter.Confidence, 1.0)
	}
	if mf.Frontmatter.Created.IsZero() {
		t.Error("Created should not be zero")
	}
	if mf.Frontmatter.Updated.IsZero() {
		t.Error("Updated should not be zero")
	}
	if mf.Content != "Test content" {
		t.Errorf("Content = %q, want %q", mf.Content, "Test content")
	}
}

func TestFrontmatter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		fm      Frontmatter
		wantErr error
	}{
		{
			name: "valid",
			fm: Frontmatter{
				ID:       "mem-test",
				Created:  time.Now(),
				Category: "fact",
			},
			wantErr: nil,
		},
		{
			name: "missing ID",
			fm: Frontmatter{
				Created:  time.Now(),
				Category: "fact",
			},
			wantErr: ErrMissingID,
		},
		{
			name: "missing created",
			fm: Frontmatter{
				ID:       "mem-test",
				Category: "fact",
			},
			wantErr: ErrMissingCreated,
		},
		{
			name: "missing category",
			fm: Frontmatter{
				ID:      "mem-test",
				Created: time.Now(),
			},
			wantErr: ErrMissingCategory,
		},
		{
			name: "invalid category",
			fm: Frontmatter{
				ID:       "mem-test",
				Created:  time.Now(),
				Category: "invalid",
			},
			wantErr: ErrInvalidCategory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fm.Validate()
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errorContains(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidCategories(t *testing.T) {
	categories := ValidCategories()
	expected := []string{"fact", "preference", "correction", "pattern"}

	if len(categories) != len(expected) {
		t.Fatalf("len = %d, want %d", len(categories), len(expected))
	}

	for i, c := range expected {
		if categories[i] != c {
			t.Errorf("categories[%d] = %q, want %q", i, categories[i], c)
		}
	}
}

func TestIsValidCategory(t *testing.T) {
	validCases := []string{"fact", "preference", "correction", "pattern"}
	for _, c := range validCases {
		if !IsValidCategory(c) {
			t.Errorf("IsValidCategory(%q) = false, want true", c)
		}
	}

	invalidCases := []string{"", "invalid", "FACT", "facts"}
	for _, c := range invalidCases {
		if IsValidCategory(c) {
			t.Errorf("IsValidCategory(%q) = true, want false", c)
		}
	}
}

// errorContains checks if err contains or wraps target.
func errorContains(err, target error) bool {
	if err == nil {
		return target == nil
	}
	if target == nil {
		return false
	}
	// Check for direct match or wrapped error
	return err == target || err.Error() == target.Error() ||
		(len(err.Error()) > len(target.Error()) &&
			err.Error()[:len(target.Error())] == target.Error()[:len(target.Error())])
}
