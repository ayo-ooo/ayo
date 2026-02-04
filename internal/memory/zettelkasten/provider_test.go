package zettelkasten

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexcabrera/ayo/internal/providers"
)

func TestProvider_Interface(t *testing.T) {
	// Verify Provider implements MemoryProvider
	var _ providers.MemoryProvider = (*Provider)(nil)
}

func TestProvider_InitAndClose(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvider()

	ctx := context.Background()
	config := map[string]any{"root": tmpDir}

	if err := p.Init(ctx, config); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Verify directories created
	s := NewStructure(tmpDir)
	if !s.Exists() {
		t.Error("structure should exist after init")
	}

	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestProvider_Create(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvider()
	ctx := context.Background()

	if err := p.Init(ctx, map[string]any{"root": tmpDir}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer p.Close()

	m := providers.Memory{
		Content:  "User prefers dark mode",
		Category: providers.MemoryCategoryPreference,
		Topics:   []string{"ui", "settings"},
	}

	created, err := p.Create(ctx, m)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if created.ID == "" {
		t.Error("created memory should have an ID")
	}
	if created.Content != m.Content {
		t.Errorf("Content = %q, want %q", created.Content, m.Content)
	}
	if created.Category != m.Category {
		t.Errorf("Category = %q, want %q", created.Category, m.Category)
	}
	if created.Status != providers.MemoryStatusActive {
		t.Errorf("Status = %q, want %q", created.Status, providers.MemoryStatusActive)
	}
	if created.Confidence != 1.0 {
		t.Errorf("Confidence = %f, want 1.0", created.Confidence)
	}
	if created.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	// Verify file exists
	path := filepath.Join(tmpDir, "preferences", created.ID+".md")
	if !fileExists(path) {
		t.Errorf("memory file should exist at %s", path)
	}
}

func TestProvider_Get(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvider()
	ctx := context.Background()

	if err := p.Init(ctx, map[string]any{"root": tmpDir}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer p.Close()

	// Create a memory
	m := providers.Memory{
		Content:  "Test fact",
		Category: providers.MemoryCategoryFact,
	}
	created, _ := p.Create(ctx, m)

	// Get it back
	got, err := p.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
	if got.Content != m.Content {
		t.Errorf("Content = %q, want %q", got.Content, m.Content)
	}

	// Get non-existent
	_, err = p.Get(ctx, "non-existent")
	if err == nil {
		t.Error("expected error for non-existent memory")
	}
}

func TestProvider_List(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvider()
	ctx := context.Background()

	if err := p.Init(ctx, map[string]any{"root": tmpDir}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer p.Close()

	// Create some memories
	memories := []providers.Memory{
		{Content: "Fact 1", Category: providers.MemoryCategoryFact, Topics: []string{"go"}},
		{Content: "Fact 2", Category: providers.MemoryCategoryFact, Topics: []string{"python"}},
		{Content: "Pref 1", Category: providers.MemoryCategoryPreference},
	}

	for _, m := range memories {
		if _, err := p.Create(ctx, m); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	// List all active
	all, err := p.List(ctx, providers.ListOptions{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("List all = %d, want 3", len(all))
	}

	// List by category
	facts, err := p.List(ctx, providers.ListOptions{
		Categories: []providers.MemoryCategory{providers.MemoryCategoryFact},
	})
	if err != nil {
		t.Fatalf("List facts: %v", err)
	}
	if len(facts) != 2 {
		t.Errorf("List facts = %d, want 2", len(facts))
	}

	// List by topic
	goFacts, err := p.List(ctx, providers.ListOptions{
		Topics: []string{"go"},
	})
	if err != nil {
		t.Fatalf("List by topic: %v", err)
	}
	if len(goFacts) != 1 {
		t.Errorf("List by topic = %d, want 1", len(goFacts))
	}

	// List with limit
	limited, err := p.List(ctx, providers.ListOptions{Limit: 1})
	if err != nil {
		t.Fatalf("List limited: %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("List limited = %d, want 1", len(limited))
	}
}

func TestProvider_Search(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvider()
	ctx := context.Background()

	if err := p.Init(ctx, map[string]any{"root": tmpDir}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer p.Close()

	// Create some memories
	p.Create(ctx, providers.Memory{Content: "User prefers dark mode for coding", Category: providers.MemoryCategoryPreference})
	p.Create(ctx, providers.Memory{Content: "Project uses Go version 1.22", Category: providers.MemoryCategoryFact})
	p.Create(ctx, providers.Memory{Content: "Light mode preferred for reading docs", Category: providers.MemoryCategoryPreference})

	// Search for "dark mode"
	results, err := p.Search(ctx, "dark mode", providers.SearchOptions{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result for 'dark mode'")
	}

	// Verify match type
	if len(results) > 0 && results[0].MatchType != "text" {
		t.Errorf("MatchType = %q, want 'text'", results[0].MatchType)
	}

	// Search with category filter
	results, err = p.Search(ctx, "mode", providers.SearchOptions{
		Categories: []providers.MemoryCategory{providers.MemoryCategoryFact},
	})
	if err != nil {
		t.Fatalf("Search with filter: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results for 'mode' in facts, got %d", len(results))
	}
}

func TestProvider_Update(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvider()
	ctx := context.Background()

	if err := p.Init(ctx, map[string]any{"root": tmpDir}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer p.Close()

	// Create a memory
	created, _ := p.Create(ctx, providers.Memory{
		Content:  "Original content",
		Category: providers.MemoryCategoryFact,
	})

	// Update it
	created.Content = "Updated content"
	created.Topics = []string{"updated"}
	if err := p.Update(ctx, created); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// Verify update
	got, _ := p.Get(ctx, created.ID)
	if got.Content != "Updated content" {
		t.Errorf("Content = %q, want 'Updated content'", got.Content)
	}
	if len(got.Topics) != 1 || got.Topics[0] != "updated" {
		t.Errorf("Topics = %v, want [updated]", got.Topics)
	}
}

func TestProvider_Forget(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvider()
	ctx := context.Background()

	if err := p.Init(ctx, map[string]any{"root": tmpDir}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer p.Close()

	// Create a memory
	created, _ := p.Create(ctx, providers.Memory{
		Content:  "To be forgotten",
		Category: providers.MemoryCategoryFact,
	})

	// Forget it
	if err := p.Forget(ctx, created.ID); err != nil {
		t.Fatalf("Forget: %v", err)
	}

	// Verify status changed
	got, _ := p.Get(ctx, created.ID)
	if got.Status != providers.MemoryStatusForgotten {
		t.Errorf("Status = %q, want %q", got.Status, providers.MemoryStatusForgotten)
	}

	// Should not appear in active list
	all, _ := p.List(ctx, providers.ListOptions{})
	for _, m := range all {
		if m.ID == created.ID {
			t.Error("forgotten memory should not appear in active list")
		}
	}
}

func TestProvider_Supersede(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvider()
	ctx := context.Background()

	if err := p.Init(ctx, map[string]any{"root": tmpDir}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer p.Close()

	// Create old memory
	old, _ := p.Create(ctx, providers.Memory{
		Content:  "Old preference",
		Category: providers.MemoryCategoryPreference,
	})

	// Supersede it
	newMem := providers.Memory{
		Content:  "New preference",
		Category: providers.MemoryCategoryPreference,
	}
	created, err := p.Supersede(ctx, old.ID, newMem, "User changed preference")
	if err != nil {
		t.Fatalf("Supersede: %v", err)
	}

	if created.SupersedesID != old.ID {
		t.Errorf("SupersedesID = %q, want %q", created.SupersedesID, old.ID)
	}
	if created.SupersessionReason != "User changed preference" {
		t.Errorf("SupersessionReason = %q, want 'User changed preference'", created.SupersessionReason)
	}

	// Old memory should be superseded
	oldUpdated, _ := p.Get(ctx, old.ID)
	if oldUpdated.Status != providers.MemoryStatusSuperseded {
		t.Errorf("old Status = %q, want %q", oldUpdated.Status, providers.MemoryStatusSuperseded)
	}
	if oldUpdated.SupersededByID != created.ID {
		t.Errorf("old SupersededByID = %q, want %q", oldUpdated.SupersededByID, created.ID)
	}
}

func TestProvider_Topics(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvider()
	ctx := context.Background()

	if err := p.Init(ctx, map[string]any{"root": tmpDir}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer p.Close()

	// Create memories with topics
	p.Create(ctx, providers.Memory{Content: "Go stuff", Category: providers.MemoryCategoryFact, Topics: []string{"go", "programming"}})
	p.Create(ctx, providers.Memory{Content: "Python stuff", Category: providers.MemoryCategoryFact, Topics: []string{"python", "programming"}})

	topics, err := p.Topics(ctx)
	if err != nil {
		t.Fatalf("Topics: %v", err)
	}

	// Should have 3 unique topics: go, python, programming
	if len(topics) != 3 {
		t.Errorf("Topics count = %d, want 3 (got: %v)", len(topics), topics)
	}
}

func TestProvider_LinkAndUnlink(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvider()
	ctx := context.Background()

	if err := p.Init(ctx, map[string]any{"root": tmpDir}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer p.Close()

	// Create two memories
	m1, _ := p.Create(ctx, providers.Memory{Content: "Memory 1", Category: providers.MemoryCategoryFact})
	m2, _ := p.Create(ctx, providers.Memory{Content: "Memory 2", Category: providers.MemoryCategoryFact})

	// Link them
	if err := p.Link(ctx, m1.ID, m2.ID); err != nil {
		t.Fatalf("Link: %v", err)
	}

	// Verify bidirectional link
	got1, _ := p.Get(ctx, m1.ID)
	got2, _ := p.Get(ctx, m2.ID)

	if len(got1.LinkedIDs) != 1 || got1.LinkedIDs[0] != m2.ID {
		t.Errorf("m1 LinkedIDs = %v, want [%s]", got1.LinkedIDs, m2.ID)
	}
	if len(got2.LinkedIDs) != 1 || got2.LinkedIDs[0] != m1.ID {
		t.Errorf("m2 LinkedIDs = %v, want [%s]", got2.LinkedIDs, m1.ID)
	}

	// Unlink them
	if err := p.Unlink(ctx, m1.ID, m2.ID); err != nil {
		t.Fatalf("Unlink: %v", err)
	}

	// Verify link removed
	got1, _ = p.Get(ctx, m1.ID)
	got2, _ = p.Get(ctx, m2.ID)

	if len(got1.LinkedIDs) != 0 {
		t.Errorf("m1 LinkedIDs after unlink = %v, want []", got1.LinkedIDs)
	}
	if len(got2.LinkedIDs) != 0 {
		t.Errorf("m2 LinkedIDs after unlink = %v, want []", got2.LinkedIDs)
	}
}

func TestProvider_Reindex(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewProvider()
	ctx := context.Background()

	if err := p.Init(ctx, map[string]any{"root": tmpDir}); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Create a memory
	created, _ := p.Create(ctx, providers.Memory{
		Content:  "Test memory",
		Category: providers.MemoryCategoryFact,
	})

	// Close and create new provider
	p.Close()

	p2 := NewProvider()
	if err := p2.Init(ctx, map[string]any{"root": tmpDir}); err != nil {
		t.Fatalf("Init p2: %v", err)
	}
	defer p2.Close()

	// Should find the memory (loaded from cache during init)
	got, err := p2.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get after reload: %v", err)
	}
	if got.Content != "Test memory" {
		t.Errorf("Content = %q, want 'Test memory'", got.Content)
	}

	// Reindex should work
	if err := p2.Reindex(ctx); err != nil {
		t.Fatalf("Reindex: %v", err)
	}

	// Should still find the memory
	got, err = p2.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get after reindex: %v", err)
	}
	if got.Content != "Test memory" {
		t.Errorf("Content after reindex = %q, want 'Test memory'", got.Content)
	}
}

func TestProvider_NotInitialized(t *testing.T) {
	p := NewProvider()
	ctx := context.Background()

	// All operations should fail when not initialized
	_, err := p.Create(ctx, providers.Memory{})
	if err == nil {
		t.Error("Create should fail when not initialized")
	}

	_, err = p.Get(ctx, "test")
	if err == nil {
		t.Error("Get should fail when not initialized")
	}

	_, err = p.Search(ctx, "test", providers.SearchOptions{})
	if err == nil {
		t.Error("Search should fail when not initialized")
	}

	_, err = p.List(ctx, providers.ListOptions{})
	if err == nil {
		t.Error("List should fail when not initialized")
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
