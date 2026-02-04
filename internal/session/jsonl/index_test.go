package jsonl

import (
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/session"
)

func TestIndex_OpenIndex(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	idx, err := OpenIndex(structure)
	if err != nil {
		t.Fatalf("OpenIndex() error = %v", err)
	}
	defer idx.Close()

	count, err := idx.Count()
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Count() = %v, want 0", count)
	}
}

func TestIndex_Rebuild(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	// Create some sessions
	for i := 0; i < 3; i++ {
		sess := session.Session{
			ID:          "session-" + string(rune('a'+i)),
			AgentHandle: "@ayo",
			Title:       "Session " + string(rune('A'+i)),
			Source:      "ayo",
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}
		w, _ := NewWriter(structure, sess)
		w.Close()
	}

	idx, err := OpenIndex(structure)
	if err != nil {
		t.Fatalf("OpenIndex() error = %v", err)
	}
	defer idx.Close()

	result, err := idx.Rebuild()
	if err != nil {
		t.Fatalf("Rebuild() error = %v", err)
	}

	if result.Indexed != 3 {
		t.Errorf("Rebuild().Indexed = %v, want 3", result.Indexed)
	}

	count, _ := idx.Count()
	if count != 3 {
		t.Errorf("Count() = %v, want 3", count)
	}
}

func TestIndex_List(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sessions := []session.Session{
		{
			ID:          "session-a",
			AgentHandle: "@ayo",
			Title:       "Ayo Session",
			Source:      "ayo",
			CreatedAt:   time.Now().Add(-2 * time.Hour).Unix(),
			UpdatedAt:   time.Now().Unix(),
		},
		{
			ID:          "session-b",
			AgentHandle: "@custom",
			Title:       "Custom Session",
			Source:      "crush",
			CreatedAt:   time.Now().Add(-1 * time.Hour).Unix(),
			UpdatedAt:   time.Now().Unix(),
		},
		{
			ID:          "session-c",
			AgentHandle: "@ayo",
			Title:       "Another Ayo",
			Source:      "ayo",
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		},
	}

	for _, sess := range sessions {
		w, _ := NewWriter(structure, sess)
		w.Close()
	}

	idx, _ := OpenIndex(structure)
	defer idx.Close()
	idx.Rebuild()

	// List all
	entries, err := idx.List("", "", 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("List() returned %d entries, want 3", len(entries))
	}

	// Filter by agent
	entries, _ = idx.List("@ayo", "", 0)
	if len(entries) != 2 {
		t.Errorf("List(@ayo) returned %d entries, want 2", len(entries))
	}

	// Filter by source
	entries, _ = idx.List("", "crush", 0)
	if len(entries) != 1 {
		t.Errorf("List(source=crush) returned %d entries, want 1", len(entries))
	}

	// With limit
	entries, _ = idx.List("", "", 2)
	if len(entries) != 2 {
		t.Errorf("List(limit=2) returned %d entries, want 2", len(entries))
	}
}

func TestIndex_Search(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sessions := []session.Session{
		{
			ID:          "session-a",
			AgentHandle: "@ayo",
			Title:       "Debug authentication issue",
			Source:      "ayo",
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		},
		{
			ID:          "session-b",
			AgentHandle: "@ayo",
			Title:       "Fix database migration",
			Source:      "ayo",
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		},
	}

	for _, sess := range sessions {
		w, _ := NewWriter(structure, sess)
		w.Close()
	}

	idx, _ := OpenIndex(structure)
	defer idx.Close()
	idx.Rebuild()

	// Search for "debug"
	entries, err := idx.Search("debug", 0)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Search('debug') returned %d entries, want 1", len(entries))
	}

	// Search for "database"
	entries, _ = idx.Search("database", 0)
	if len(entries) != 1 {
		t.Errorf("Search('database') returned %d entries, want 1", len(entries))
	}

	// Search with no matches
	entries, _ = idx.Search("nonexistent", 0)
	if len(entries) != 0 {
		t.Errorf("Search('nonexistent') returned %d entries, want 0", len(entries))
	}
}

func TestIndex_Get(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	sess := session.Session{
		ID:          "test-get-session",
		AgentHandle: "@ayo",
		Title:       "Get Test",
		Source:      "ayo",
		ChainDepth:  2,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}
	w, _ := NewWriter(structure, sess)
	w.Close()

	idx, _ := OpenIndex(structure)
	defer idx.Close()
	idx.Rebuild()

	entry, err := idx.Get("test-get-session")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if entry.Title != "Get Test" {
		t.Errorf("Title = %v, want Get Test", entry.Title)
	}
	if entry.ChainDepth != 2 {
		t.Errorf("ChainDepth = %v, want 2", entry.ChainDepth)
	}
}

func TestIndex_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	idx, _ := OpenIndex(structure)
	defer idx.Close()

	_, err := idx.Get("nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("Get() error = %v, want ErrSessionNotFound", err)
	}
}

func TestIndex_GetByPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	for i := 0; i < 3; i++ {
		sess := session.Session{
			ID:          "prefix-session-" + string(rune('a'+i)),
			AgentHandle: "@ayo",
			Title:       "Prefix Session",
			Source:      "ayo",
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}
		w, _ := NewWriter(structure, sess)
		w.Close()
	}

	// Different prefix
	other := session.Session{
		ID:          "other-session",
		AgentHandle: "@ayo",
		Title:       "Other",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}
	w, _ := NewWriter(structure, other)
	w.Close()

	idx, _ := OpenIndex(structure)
	defer idx.Close()
	idx.Rebuild()

	entries, err := idx.GetByPrefix("prefix-")
	if err != nil {
		t.Fatalf("GetByPrefix() error = %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("GetByPrefix('prefix-') returned %d entries, want 3", len(entries))
	}
}

func TestIndex_Sync(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	// Create initial sessions
	sess1 := session.Session{
		ID:          "sync-session-1",
		AgentHandle: "@ayo",
		Title:       "Sync 1",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}
	w, _ := NewWriter(structure, sess1)
	w.Close()

	idx, _ := OpenIndex(structure)
	defer idx.Close()
	idx.Rebuild()

	initialCount, _ := idx.Count()
	if initialCount != 1 {
		t.Fatalf("initial count = %d, want 1", initialCount)
	}

	// Add new session (not indexed)
	sess2 := session.Session{
		ID:          "sync-session-2",
		AgentHandle: "@ayo",
		Title:       "Sync 2",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}
	w2, _ := NewWriter(structure, sess2)
	w2.Close()

	// Sync
	result, err := idx.Sync()
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if result.Added != 1 {
		t.Errorf("Sync().Added = %d, want 1", result.Added)
	}

	afterSync, _ := idx.Count()
	if afterSync != 2 {
		t.Errorf("count after sync = %d, want 2", afterSync)
	}
}

func TestEnsureIndex(t *testing.T) {
	tmpDir := t.TempDir()
	structure := NewStructure(tmpDir)

	// Create session
	sess := session.Session{
		ID:          "ensure-session",
		AgentHandle: "@ayo",
		Title:       "Ensure Test",
		Source:      "ayo",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}
	w, _ := NewWriter(structure, sess)
	w.Close()

	// EnsureIndex should create and rebuild
	idx, err := EnsureIndex(structure)
	if err != nil {
		t.Fatalf("EnsureIndex() error = %v", err)
	}
	defer idx.Close()

	count, _ := idx.Count()
	if count != 1 {
		t.Errorf("Count() = %d, want 1", count)
	}
}
