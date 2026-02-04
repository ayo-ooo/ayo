package zettelkasten

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestIndex_OpenClose(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStructure(tmpDir)
	if err := s.Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	idx := NewIndex(s)
	ctx := context.Background()

	if err := idx.Open(ctx); err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := idx.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestIndex_InsertAndSearch(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStructure(tmpDir)
	s.Initialize()

	idx := NewIndex(s)
	ctx := context.Background()
	idx.Open(ctx)
	defer idx.Close()

	// Insert a memory
	entry := IndexEntry{
		ID:         "mem-test1",
		Category:   "fact",
		Status:     "active",
		Content:    "The user prefers dark mode in VS Code",
		Confidence: 0.9,
		Topics:     sql.NullString{String: `["vscode","preferences"]`, Valid: true},
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}

	if err := idx.Insert(ctx, entry); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	// Verify count
	count, err := idx.Count(ctx)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 1 {
		t.Errorf("Count = %d, want 1", count)
	}

	// Search using FTS
	results, err := idx.SearchFTS(ctx, "dark mode", "active", 10)
	if err != nil {
		t.Fatalf("SearchFTS: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("SearchFTS results = %d, want 1", len(results))
	}
	if len(results) > 0 && results[0].ID != "mem-test1" {
		t.Errorf("SearchFTS ID = %q, want %q", results[0].ID, "mem-test1")
	}
}

func TestIndex_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStructure(tmpDir)
	s.Initialize()

	idx := NewIndex(s)
	ctx := context.Background()
	idx.Open(ctx)
	defer idx.Close()

	// Insert and delete
	entry := IndexEntry{
		ID:        "mem-delete",
		Category:  "fact",
		Status:    "active",
		Content:   "To be deleted",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	idx.Insert(ctx, entry)

	if err := idx.Delete(ctx, "mem-delete"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	count, _ := idx.Count(ctx)
	if count != 0 {
		t.Errorf("Count after delete = %d, want 0", count)
	}
}

func TestIndex_Rebuild(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStructure(tmpDir)
	s.Initialize()

	// Create some memory files first
	mf1 := NewMemoryFile("mem-rb1", "fact", "First memory for rebuild")
	mf1.WriteFile(s.MemoryPath("mem-rb1", "fact"))

	mf2 := NewMemoryFile("mem-rb2", "preference", "Second memory for rebuild")
	mf2.WriteFile(s.MemoryPath("mem-rb2", "preference"))

	// Open index and rebuild
	idx := NewIndex(s)
	ctx := context.Background()
	idx.Open(ctx)
	defer idx.Close()

	if err := idx.Rebuild(ctx); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	// Should have 2 entries
	count, _ := idx.Count(ctx)
	if count != 2 {
		t.Errorf("Count after rebuild = %d, want 2", count)
	}

	// Search should work
	results, _ := idx.SearchFTS(ctx, "rebuild", "active", 10)
	if len(results) != 2 {
		t.Errorf("SearchFTS after rebuild = %d, want 2", len(results))
	}
}

func TestIndex_CountByStatus(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStructure(tmpDir)
	s.Initialize()

	idx := NewIndex(s)
	ctx := context.Background()
	idx.Open(ctx)
	defer idx.Close()

	now := time.Now().Unix()

	// Insert active and forgotten memories
	idx.Insert(ctx, IndexEntry{
		ID: "active1", Category: "fact", Status: "active",
		Content: "Active", CreatedAt: now, UpdatedAt: now,
	})
	idx.Insert(ctx, IndexEntry{
		ID: "active2", Category: "fact", Status: "active",
		Content: "Active 2", CreatedAt: now, UpdatedAt: now,
	})
	idx.Insert(ctx, IndexEntry{
		ID: "forgotten1", Category: "fact", Status: "forgotten",
		Content: "Forgotten", CreatedAt: now, UpdatedAt: now,
	})

	activeCount, _ := idx.CountByStatus(ctx, "active")
	if activeCount != 2 {
		t.Errorf("Active count = %d, want 2", activeCount)
	}

	forgottenCount, _ := idx.CountByStatus(ctx, "forgotten")
	if forgottenCount != 1 {
		t.Errorf("Forgotten count = %d, want 1", forgottenCount)
	}
}

func TestIndex_UpdateEmbedding(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStructure(tmpDir)
	s.Initialize()

	idx := NewIndex(s)
	ctx := context.Background()
	idx.Open(ctx)
	defer idx.Close()

	now := time.Now().Unix()
	idx.Insert(ctx, IndexEntry{
		ID: "mem-embed", Category: "fact", Status: "active",
		Content: "Test", CreatedAt: now, UpdatedAt: now,
	})

	// Initially should have no embedding
	missing, err := idx.GetMissingEmbeddings(ctx, 10)
	if err != nil {
		t.Fatalf("GetMissingEmbeddings: %v", err)
	}
	if len(missing) != 1 {
		t.Errorf("Missing embeddings = %d, want 1", len(missing))
	}

	// Update embedding
	embedding := []byte{0x01, 0x02, 0x03, 0x04}
	if err := idx.UpdateEmbedding(ctx, "mem-embed", embedding); err != nil {
		t.Fatalf("UpdateEmbedding: %v", err)
	}

	// Should no longer be missing
	missing, _ = idx.GetMissingEmbeddings(ctx, 10)
	if len(missing) != 0 {
		t.Errorf("Missing after update = %d, want 0", len(missing))
	}
}

func TestIndex_FTSSearchFilters(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStructure(tmpDir)
	s.Initialize()

	idx := NewIndex(s)
	ctx := context.Background()
	idx.Open(ctx)
	defer idx.Close()

	now := time.Now().Unix()

	// Insert various memories
	idx.Insert(ctx, IndexEntry{
		ID: "active-go", Category: "fact", Status: "active",
		Content: "Go programming language", CreatedAt: now, UpdatedAt: now,
	})
	idx.Insert(ctx, IndexEntry{
		ID: "forgotten-go", Category: "fact", Status: "forgotten",
		Content: "Go concurrency patterns", CreatedAt: now, UpdatedAt: now,
	})

	// Search active only (default)
	results, _ := idx.SearchFTS(ctx, "Go", "active", 10)
	if len(results) != 1 {
		t.Errorf("Active search = %d, want 1", len(results))
	}

	// Search forgotten
	results, _ = idx.SearchFTS(ctx, "Go", "forgotten", 10)
	if len(results) != 1 {
		t.Errorf("Forgotten search = %d, want 1", len(results))
	}
}

func TestIndexFromMemoryFile(t *testing.T) {
	mf := NewMemoryFile("mem-convert", "preference", "User prefers vim")
	mf.Frontmatter.Topics = []string{"editor", "vim"}
	mf.Frontmatter.Scope.Agent = "@ayo"
	mf.Frontmatter.Scope.Path = "/home/user/project"
	mf.Frontmatter.Confidence = 0.85
	mf.Frontmatter.Supersession.Supersedes = "mem-old"

	entry := IndexFromMemoryFile(mf)

	if entry.ID != "mem-convert" {
		t.Errorf("ID = %q, want %q", entry.ID, "mem-convert")
	}
	if entry.Category != "preference" {
		t.Errorf("Category = %q, want %q", entry.Category, "preference")
	}
	if !entry.AgentHandle.Valid || entry.AgentHandle.String != "@ayo" {
		t.Errorf("AgentHandle = %v, want @ayo", entry.AgentHandle)
	}
	if !entry.PathScope.Valid || entry.PathScope.String != "/home/user/project" {
		t.Errorf("PathScope = %v, want /home/user/project", entry.PathScope)
	}
	if !entry.Topics.Valid {
		t.Error("Topics should be valid")
	}
	if !entry.SupersedesID.Valid || entry.SupersedesID.String != "mem-old" {
		t.Errorf("SupersedesID = %v, want mem-old", entry.SupersedesID)
	}
}

func TestTopicsToJSON(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{nil, "[]"},
		{[]string{}, "[]"},
		{[]string{"go"}, `["go"]`},
		{[]string{"go", "testing"}, `["go","testing"]`},
	}

	for _, tt := range tests {
		result := topicsToJSON(tt.input)
		if result != tt.expected {
			t.Errorf("topicsToJSON(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
