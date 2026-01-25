package memory

import (
	"context"
	"testing"

	"github.com/alexcabrera/ayo/internal/db"
)

// mockEmbedder is a simple mock for testing without ONNX.
type mockEmbedder struct {
	dimension int
}

func (m *mockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// Return a simple hash-based embedding for testing
	emb := make([]float32, m.dimension)
	for i, r := range text {
		emb[i%m.dimension] += float32(r) / 1000
	}
	return emb, nil
}

func (m *mockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, t := range texts {
		emb, err := m.Embed(ctx, t)
		if err != nil {
			return nil, err
		}
		results[i] = emb
	}
	return results, nil
}

func (m *mockEmbedder) Dimension() int { return m.dimension }
func (m *mockEmbedder) Close() error   { return nil }

func setupTestService(t *testing.T) (*Service, func()) {
	t.Helper()

	ctx := context.Background()
	testDB, queries, err := db.ConnectWithQueries(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	embedder := &mockEmbedder{dimension: 384}
	svc := NewService(queries, embedder)

	cleanup := func() {
		testDB.Close()
	}

	return svc, cleanup
}

func TestCreateAndGet(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	mem := Memory{
		Content:     "User prefers tabs over spaces",
		Category:    CategoryPreference,
		AgentHandle: "@ayo",
		Confidence:  0.9,
	}

	created, err := svc.Create(ctx, mem)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if len(created.Embedding) == 0 {
		t.Error("Expected embedding to be generated")
	}
	if created.Status != StatusActive {
		t.Errorf("Expected status %s, got %s", StatusActive, created.Status)
	}

	// Get the memory
	retrieved, err := svc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Content != mem.Content {
		t.Errorf("Content mismatch: got %s, want %s", retrieved.Content, mem.Content)
	}
	if retrieved.Category != mem.Category {
		t.Errorf("Category mismatch: got %s, want %s", retrieved.Category, mem.Category)
	}
}

func TestSearch(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create some memories
	memories := []Memory{
		{Content: "User likes Go programming language", Category: CategoryFact},
		{Content: "User prefers dark theme", Category: CategoryPreference},
		{Content: "Always use gofmt for formatting", Category: CategoryCorrection},
	}

	for _, m := range memories {
		_, err := svc.Create(ctx, m)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Search for Go-related memories
	results, err := svc.Search(ctx, "Go programming", SearchOptions{
		Threshold: 0.0, // Low threshold for mock embeddings
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected search results")
	}

	// Results should be sorted by similarity
	for i := 1; i < len(results); i++ {
		if results[i].Similarity > results[i-1].Similarity {
			t.Error("Results should be sorted by similarity descending")
		}
	}
}

func TestSupersede(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create original memory
	original, err := svc.Create(ctx, Memory{
		Content:  "User prefers 2-space indentation",
		Category: CategoryPreference,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Supersede with new memory
	newMem := Memory{
		Content:  "User prefers 4-space indentation",
		Category: CategoryPreference,
	}
	superseding, err := svc.Supersede(ctx, original.ID, newMem, "User changed preference")
	if err != nil {
		t.Fatalf("Supersede failed: %v", err)
	}

	if superseding.SupersedesID != original.ID {
		t.Error("New memory should reference the superseded memory")
	}

	// Check original is marked as superseded
	retrieved, err := svc.Get(ctx, original.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Status != StatusSuperseded {
		t.Errorf("Expected status %s, got %s", StatusSuperseded, retrieved.Status)
	}
	if retrieved.SupersededByID != superseding.ID {
		t.Error("Original should reference the superseding memory")
	}
}

func TestForget(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	mem, err := svc.Create(ctx, Memory{
		Content:  "Something to forget",
		Category: CategoryFact,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = svc.Forget(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Forget failed: %v", err)
	}

	retrieved, err := svc.Get(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Status != StatusForgotten {
		t.Errorf("Expected status %s, got %s", StatusForgotten, retrieved.Status)
	}
}

func TestListAndCount(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create memories for different agents
	for i := 0; i < 3; i++ {
		_, _ = svc.Create(ctx, Memory{
			Content:     "Global memory",
			Category:    CategoryFact,
			AgentHandle: "",
		})
	}
	for i := 0; i < 2; i++ {
		_, _ = svc.Create(ctx, Memory{
			Content:     "Agent-specific memory",
			Category:    CategoryFact,
			AgentHandle: "@ayo",
		})
	}

	// Count all
	total, err := svc.Count(ctx, "")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if total != 5 {
		t.Errorf("Expected 5 total memories, got %d", total)
	}

	// Count by agent
	agentCount, err := svc.Count(ctx, "@ayo")
	if err != nil {
		t.Fatalf("Count by agent failed: %v", err)
	}
	if agentCount != 2 {
		t.Errorf("Expected 2 agent memories, got %d", agentCount)
	}

	// List all
	all, err := svc.List(ctx, "", 100, 0)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(all) != 5 {
		t.Errorf("Expected 5 memories, got %d", len(all))
	}

	// List by agent
	agentMems, err := svc.List(ctx, "@ayo", 100, 0)
	if err != nil {
		t.Fatalf("List by agent failed: %v", err)
	}
	if len(agentMems) != 2 {
		t.Errorf("Expected 2 agent memories, got %d", len(agentMems))
	}
}

func TestClear(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create some memories
	for i := 0; i < 3; i++ {
		_, _ = svc.Create(ctx, Memory{
			Content:     "To be cleared",
			Category:    CategoryFact,
			AgentHandle: "@ayo",
		})
	}

	count, _ := svc.Count(ctx, "@ayo")
	if count != 3 {
		t.Fatalf("Expected 3 memories before clear, got %d", count)
	}

	err := svc.Clear(ctx, "@ayo")
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	count, _ = svc.Count(ctx, "@ayo")
	if count != 0 {
		t.Errorf("Expected 0 memories after clear, got %d", count)
	}
}
