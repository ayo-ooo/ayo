package capabilities

import (
	"context"
	"testing"
)

func TestUnifiedSearcher_Search(t *testing.T) {
	// Create embedder that returns predictable embeddings
	embedder := &testEmbedder{embedding: []float32{1.0, 0.0, 0.0}}
	idx := NewLazyEntityIndex(embedder)

	// Add agents with embeddings
	idx.UpsertAgent(IndexedAgent{
		Handle:      "@code",
		Description: "Coding assistant",
		Embedding:   []float32{1.0, 0.0, 0.0}, // Perfect match with query
	})
	idx.UpsertAgent(IndexedAgent{
		Handle:      "@writer",
		Description: "Writing assistant",
		Embedding:   []float32{0.5, 0.5, 0.0}, // Partial match
	})

	// Add squads with embeddings
	idx.UpsertSquad(IndexedSquad{
		Name:      "dev-team",
		Mission:   "Build software",
		Embedding: []float32{0.8, 0.2, 0.0}, // Good match
	})

	searcher := NewUnifiedSearcher(idx, embedder)

	// Search
	results, err := searcher.Search(context.Background(), "coding task", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// First result should be the agent with perfect match
	if len(results) > 0 && results[0].Handle != "@code" {
		t.Errorf("Expected @code first, got %s", results[0].Handle)
	}
}

func TestUnifiedSearcher_FindBest(t *testing.T) {
	embedder := &testEmbedder{embedding: []float32{1.0, 0.0, 0.0}}
	idx := NewLazyEntityIndex(embedder)

	idx.UpsertAgent(IndexedAgent{
		Handle:      "@code",
		Description: "Coding assistant",
		Embedding:   []float32{1.0, 0.0, 0.0},
	})
	idx.UpsertSquad(IndexedSquad{
		Name:      "dev-team",
		Mission:   "Build software",
		Embedding: []float32{0.5, 0.5, 0.0},
	})

	searcher := NewUnifiedSearcher(idx, embedder)

	result, err := searcher.FindBest(context.Background(), "coding task")
	if err != nil {
		t.Fatalf("FindBest failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Handle != "@code" {
		t.Errorf("Expected @code, got %s", result.Handle)
	}
	if result.Type != EntityTypeAgentStr {
		t.Errorf("Expected type agent, got %s", result.Type)
	}
}

func TestUnifiedSearcher_EmptyIndex(t *testing.T) {
	embedder := &testEmbedder{embedding: []float32{1.0, 0.0, 0.0}}
	idx := NewLazyEntityIndex(embedder)
	searcher := NewUnifiedSearcher(idx, embedder)

	// Search on empty index
	results, err := searcher.Search(context.Background(), "query", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}

	// FindBest on empty index
	result, err := searcher.FindBest(context.Background(), "query")
	if err != nil {
		t.Fatalf("FindBest failed: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestUnifiedSearcher_SearchAgentsOnly(t *testing.T) {
	embedder := &testEmbedder{embedding: []float32{1.0, 0.0, 0.0}}
	idx := NewLazyEntityIndex(embedder)

	idx.UpsertAgent(IndexedAgent{
		Handle:      "@code",
		Description: "Coding assistant",
		Embedding:   []float32{1.0, 0.0, 0.0},
	})
	idx.UpsertSquad(IndexedSquad{
		Name:      "dev-team",
		Mission:   "Build software",
		Embedding: []float32{1.0, 0.0, 0.0}, // Same score as agent
	})

	searcher := NewUnifiedSearcher(idx, embedder)

	results, err := searcher.SearchAgentsOnly(context.Background(), "query", 10)
	if err != nil {
		t.Fatalf("SearchAgentsOnly failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0].Type != EntityTypeAgentStr {
		t.Errorf("Expected agent type, got %s", results[0].Type)
	}
}

func TestUnifiedSearcher_SearchSquadsOnly(t *testing.T) {
	embedder := &testEmbedder{embedding: []float32{1.0, 0.0, 0.0}}
	idx := NewLazyEntityIndex(embedder)

	idx.UpsertAgent(IndexedAgent{
		Handle:      "@code",
		Description: "Coding assistant",
		Embedding:   []float32{1.0, 0.0, 0.0},
	})
	idx.UpsertSquad(IndexedSquad{
		Name:      "dev-team",
		Mission:   "Build software",
		Embedding: []float32{1.0, 0.0, 0.0},
	})

	searcher := NewUnifiedSearcher(idx, embedder)

	results, err := searcher.SearchSquadsOnly(context.Background(), "query", 10)
	if err != nil {
		t.Fatalf("SearchSquadsOnly failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0].Type != EntityTypeSquadStr {
		t.Errorf("Expected squad type, got %s", results[0].Type)
	}
}

func TestUnifiedSearcher_Counts(t *testing.T) {
	embedder := &testEmbedder{embedding: []float32{1.0, 0.0, 0.0}}
	idx := NewLazyEntityIndex(embedder)

	idx.UpsertAgent(IndexedAgent{Handle: "@a1"})
	idx.UpsertAgent(IndexedAgent{Handle: "@a2"})
	idx.UpsertSquad(IndexedSquad{Name: "s1"})

	searcher := NewUnifiedSearcher(idx, embedder)

	if searcher.AgentCount() != 2 {
		t.Errorf("AgentCount = %d, want 2", searcher.AgentCount())
	}
	if searcher.SquadCount() != 1 {
		t.Errorf("SquadCount = %d, want 1", searcher.SquadCount())
	}
	if searcher.IndexedEntityCount() != 3 {
		t.Errorf("IndexedEntityCount = %d, want 3", searcher.IndexedEntityCount())
	}
}
