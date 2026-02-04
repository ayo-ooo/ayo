package zettelkasten

import (
	"context"
	"testing"
	"time"
)

func TestHybridSearcher_TextOnly(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStructure(tmpDir)
	s.Initialize()

	idx := NewIndex(s)
	ctx := context.Background()
	idx.Open(ctx)
	defer idx.Close()

	// Insert test data
	now := time.Now().Unix()
	idx.Insert(ctx, IndexEntry{
		ID: "mem-1", Category: "fact", Status: "active",
		Content: "User prefers dark mode for coding", CreatedAt: now, UpdatedAt: now,
	})
	idx.Insert(ctx, IndexEntry{
		ID: "mem-2", Category: "fact", Status: "active",
		Content: "Light mode is better for reading documentation", CreatedAt: now, UpdatedAt: now,
	})
	idx.Insert(ctx, IndexEntry{
		ID: "mem-3", Category: "preference", Status: "active",
		Content: "Project uses Go 1.22", CreatedAt: now, UpdatedAt: now,
	})

	// Search without embedding provider (text only)
	hs := NewHybridSearcher(idx, nil)

	results, err := hs.Search(ctx, HybridSearchOptions{
		Query: "dark mode",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	// First result should be the dark mode one
	if results[0].Entry.ID != "mem-1" {
		t.Errorf("first result ID = %q, want %q", results[0].Entry.ID, "mem-1")
	}

	// Should have text rank but no semantic rank
	if results[0].TextRank == 0 {
		t.Error("TextRank should be set")
	}
	if results[0].SemanticRank != 0 {
		t.Errorf("SemanticRank = %d, want 0 (no embeddings)", results[0].SemanticRank)
	}
}

func TestHybridSearcher_Options(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStructure(tmpDir)
	s.Initialize()

	idx := NewIndex(s)
	ctx := context.Background()
	idx.Open(ctx)
	defer idx.Close()

	// Insert test data
	now := time.Now().Unix()
	for i := 0; i < 10; i++ {
		idx.Insert(ctx, IndexEntry{
			ID: "mem-" + string(rune('a'+i)), Category: "fact", Status: "active",
			Content: "Testing hybrid search functionality", CreatedAt: now, UpdatedAt: now,
		})
	}

	hs := NewHybridSearcher(idx, nil)

	// Test limit
	results, _ := hs.Search(ctx, HybridSearchOptions{
		Query: "testing",
		Limit: 3,
	})
	if len(results) > 3 {
		t.Errorf("limit not applied: got %d, want <= 3", len(results))
	}
}

func TestRRFFusion(t *testing.T) {
	// Test RRF formula: score = sum(1/(k + rank))
	hs := &HybridSearcher{}

	semantic := []HybridResult{
		{Entry: IndexEntry{ID: "a"}, SemanticRank: 1, SemanticScore: 0.9},
		{Entry: IndexEntry{ID: "b"}, SemanticRank: 2, SemanticScore: 0.7},
		{Entry: IndexEntry{ID: "c"}, SemanticRank: 3, SemanticScore: 0.5},
	}

	text := []HybridResult{
		{Entry: IndexEntry{ID: "b"}, TextRank: 1, TextScore: 0.95},
		{Entry: IndexEntry{ID: "a"}, TextRank: 2, TextScore: 0.8},
		{Entry: IndexEntry{ID: "d"}, TextRank: 3, TextScore: 0.6},
	}

	opts := HybridSearchOptions{
		SemanticWeight: 0.5,
		RRFConstant:    60,
	}

	merged := hs.fuseResults(semantic, text, opts)

	// Should have 4 unique results
	if len(merged) != 4 {
		t.Errorf("merged count = %d, want 4", len(merged))
	}

	// Check that 'b' is first (rank 2 semantic + rank 1 text)
	// 'a' is second (rank 1 semantic + rank 2 text)
	if len(merged) >= 2 {
		// Both 'a' and 'b' appear in both lists
		// b: semantic_rank=2, text_rank=1
		// a: semantic_rank=1, text_rank=2
		// With equal weights, they should be close
		foundA := false
		foundB := false
		for i := 0; i < 2 && i < len(merged); i++ {
			if merged[i].Entry.ID == "a" {
				foundA = true
			}
			if merged[i].Entry.ID == "b" {
				foundB = true
			}
		}
		if !foundA || !foundB {
			t.Error("expected both 'a' and 'b' in top 2")
		}
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []float32
		expected float32
	}{
		{
			name:     "identical vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{1, 0, 0},
			expected: 1.0,
		},
		{
			name:     "orthogonal vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{0, 1, 0},
			expected: 0.0,
		},
		{
			name:     "opposite vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{-1, 0, 0},
			expected: -1.0,
		},
		{
			name:     "similar vectors",
			a:        []float32{1, 1, 0},
			b:        []float32{1, 0, 0},
			expected: 0.7071, // 1/(sqrt(2)*1)
		},
		{
			name:     "empty vectors",
			a:        []float32{},
			b:        []float32{},
			expected: 0.0,
		},
		{
			name:     "different lengths",
			a:        []float32{1, 2},
			b:        []float32{1, 2, 3},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cosineSimilarity(tt.a, tt.b)
			// Allow small floating point tolerance
			if abs(result-tt.expected) > 0.001 {
				t.Errorf("cosineSimilarity = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestBytesFloat32Conversion(t *testing.T) {
	original := []float32{1.5, -2.5, 3.14159, 0, -0}

	// Convert to bytes and back
	bytes := float32ToBytes(original)
	result := bytesToFloat32(bytes)

	if len(result) != len(original) {
		t.Fatalf("length mismatch: got %d, want %d", len(result), len(original))
	}

	for i := range original {
		if result[i] != original[i] {
			t.Errorf("value[%d] = %f, want %f", i, result[i], original[i])
		}
	}
}

func TestBytesToFloat32_Invalid(t *testing.T) {
	// Empty
	if bytesToFloat32(nil) != nil {
		t.Error("nil should return nil")
	}
	if bytesToFloat32([]byte{}) != nil {
		t.Error("empty should return nil")
	}

	// Not multiple of 4
	if bytesToFloat32([]byte{1, 2, 3}) != nil {
		t.Error("odd length should return nil")
	}
}

func TestDefaultHybridSearchOptions(t *testing.T) {
	opts := DefaultHybridSearchOptions()

	if opts.Status != "active" {
		t.Errorf("Status = %q, want 'active'", opts.Status)
	}
	if opts.Limit != 20 {
		t.Errorf("Limit = %d, want 20", opts.Limit)
	}
	if opts.SemanticWeight != 0.5 {
		t.Errorf("SemanticWeight = %f, want 0.5", opts.SemanticWeight)
	}
	if opts.RRFConstant != 60 {
		t.Errorf("RRFConstant = %d, want 60", opts.RRFConstant)
	}
}

func abs(f float32) float32 {
	if f < 0 {
		return -f
	}
	return f
}
