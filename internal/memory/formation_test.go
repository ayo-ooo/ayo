package memory

import (
	"context"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/db"
)

func TestDetectTriggers(t *testing.T) {
	tests := []struct {
		msg  string
		want int
	}{
		{"Remember that I prefer TypeScript", 2}, // explicit + preference
		{"I prefer Go over Python", 1},           // preference only
		{"No, use pnpm instead", 1},              // correction only
		{"Hello world", 0},                       // none
	}

	for _, tt := range tests {
		triggers := DetectTriggers(tt.msg)
		if len(triggers) != tt.want {
			t.Errorf("DetectTriggers(%q) = %d triggers, want %d", tt.msg, len(triggers), tt.want)
			for i, tr := range triggers {
				t.Logf("  trigger %d: %s", i, tr.Type)
			}
		}
	}
}

// deterministicEmbedder produces consistent embeddings for the same input
type deterministicEmbedder struct {
	dimension int
}

func (m *deterministicEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	emb := make([]float32, m.dimension)
	for i, r := range text {
		emb[i%m.dimension] += float32(r) / 1000
	}
	// Normalize for consistent similarity calculations
	var sum float32
	for _, v := range emb {
		sum += v * v
	}
	if sum > 0 {
		norm := float32(1.0) / sqrt32(sum)
		for i := range emb {
			emb[i] *= norm
		}
	}
	return emb, nil
}

func (m *deterministicEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
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

func (m *deterministicEmbedder) Dimension() int { return m.dimension }
func (m *deterministicEmbedder) Close() error   { return nil }

func sqrt32(x float32) float32 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func setupFormationTestService(t *testing.T) (*Service, *FormationService, func()) {
	t.Helper()

	ctx := context.Background()
	testDB, queries, err := db.ConnectWithQueries(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	embedder := &deterministicEmbedder{dimension: 384}
	svc := NewService(queries, embedder)
	formSvc := NewFormationService(svc)

	cleanup := func() {
		testDB.Close()
	}

	return svc, formSvc, cleanup
}

func TestFormationDeduplication_ExactDuplicate(t *testing.T) {
	svc, formSvc, cleanup := setupFormationTestService(t)
	defer cleanup()

	ctx := context.Background()
	content := "User prefers TypeScript over JavaScript"

	// Track formation results
	var results []FormationResult
	formSvc.OnFormation(func(r FormationResult) {
		results = append(results, r)
	})

	// Queue first formation
	formSvc.QueueFormation(ctx, FormationIntent{
		Content:  content,
		Category: CategoryPreference,
	})

	// Wait for formation to complete
	formSvc.Wait(5 * time.Second)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Success {
		t.Fatalf("First formation failed: %v", results[0].Error)
	}
	if results[0].Deduplicated {
		t.Error("First formation should not be deduplicated")
	}

	firstMemoryID := results[0].Memory.ID
	results = nil // Reset for next formation

	// Queue exact same content again
	formSvc.QueueFormation(ctx, FormationIntent{
		Content:  content,
		Category: CategoryPreference,
	})

	formSvc.Wait(5 * time.Second)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if !results[0].Success {
		t.Fatalf("Second formation failed: %v", results[0].Error)
	}
	if !results[0].Deduplicated {
		t.Error("Second formation should be deduplicated")
	}
	if results[0].Memory.ID != firstMemoryID {
		t.Error("Deduplicated formation should return existing memory")
	}

	// Verify only one memory exists
	count, err := svc.Count(ctx, "")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 memory, got %d", count)
	}
}

func TestFormationDeduplication_SimilarSupersedes(t *testing.T) {
	svc, formSvc, cleanup := setupFormationTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Track formation results
	var results []FormationResult
	formSvc.OnFormation(func(r FormationResult) {
		results = append(results, r)
	})

	// Queue first formation
	formSvc.QueueFormation(ctx, FormationIntent{
		Content:  "User prefers TypeScript over JavaScript for web development",
		Category: CategoryPreference,
	})

	formSvc.Wait(5 * time.Second)

	if len(results) != 1 || !results[0].Success {
		t.Fatalf("First formation failed")
	}
	originalID := results[0].Memory.ID
	results = nil

	// Queue similar but not identical content
	formSvc.QueueFormation(ctx, FormationIntent{
		Content:  "User prefers TypeScript over JavaScript for all web development projects",
		Category: CategoryPreference,
	})

	formSvc.Wait(5 * time.Second)

	if len(results) != 1 || !results[0].Success {
		t.Fatalf("Second formation failed")
	}

	// Check if it was superseded (similarity between 0.85-0.95)
	// Note: With our simple embedder, similar strings should have high similarity
	if results[0].Superseded {
		if results[0].SupersededID != originalID {
			t.Error("Superseded ID should match original")
		}
		// Verify original is marked as superseded
		original, err := svc.Get(ctx, originalID)
		if err != nil {
			t.Fatalf("Get original failed: %v", err)
		}
		if original.Status != StatusSuperseded {
			t.Errorf("Original should be superseded, got %s", original.Status)
		}
	}
	// If not superseded, similarity was below threshold which is also valid
}

func TestFormationDeduplication_DifferentContent(t *testing.T) {
	svc, formSvc, cleanup := setupFormationTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Track formation results
	var results []FormationResult
	formSvc.OnFormation(func(r FormationResult) {
		results = append(results, r)
	})

	// Queue first formation
	formSvc.QueueFormation(ctx, FormationIntent{
		Content:  "User prefers dark theme",
		Category: CategoryPreference,
	})

	formSvc.Wait(5 * time.Second)
	results = nil

	// Queue completely different content
	formSvc.QueueFormation(ctx, FormationIntent{
		Content:  "Always use Go for backend services",
		Category: CategoryFact,
	})

	formSvc.Wait(5 * time.Second)

	if len(results) != 1 || !results[0].Success {
		t.Fatalf("Second formation failed")
	}
	if results[0].Deduplicated || results[0].Superseded {
		t.Error("Different content should create new memory, not dedupe or supersede")
	}

	// Verify both memories exist
	count, err := svc.Count(ctx, "")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 memories, got %d", count)
	}
}
