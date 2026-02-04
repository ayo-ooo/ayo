package zettelkasten

import (
	"context"
	"math"
	"os"
	"testing"
	"time"
	"unsafe"

	"github.com/alexcabrera/ayo/internal/providers"
)

func TestMergerFindCandidates(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "merge-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize provider
	provider := NewProvider()
	if err := provider.Init(context.Background(), map[string]any{"root": tmpDir}); err != nil {
		t.Fatal(err)
	}
	defer provider.Close()

	// Create two similar memories
	mem1 := providers.Memory{
		ID:       "mem-test-001",
		Category: "fact",
		Content:  "The user prefers Go for backend development.",
		Topics:   []string{"golang"},
	}
	mem2 := providers.Memory{
		ID:       "mem-test-002",
		Category: "fact",
		Content:  "The user likes Go for backend services.",
		Topics:   []string{"golang"},
	}

	// Store memories
	if _, err := provider.Create(context.Background(), mem1); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Create(context.Background(), mem2); err != nil {
		t.Fatal(err)
	}

	// Create index with embeddings
	structure := NewStructure(tmpDir)
	idx := NewIndex(structure)
	if err := idx.Open(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	// Add entries with mock embeddings (similar vectors)
	idx.Insert(context.Background(), IndexEntry{
		ID:        "mem-test-001",
		Category:  "fact",
		Status:    "active",
		Content:   mem1.Content,
		Embedding: testFloat32ToBytes([]float32{0.9, 0.1, 0.1, 0.1}),
	})
	idx.Insert(context.Background(), IndexEntry{
		ID:        "mem-test-002",
		Category:  "fact",
		Status:    "active",
		Content:   mem2.Content,
		Embedding: testFloat32ToBytes([]float32{0.85, 0.15, 0.1, 0.1}),
	})

	// Create merger
	config := DefaultMergeConfig()
	config.SimilarityThreshold = 0.99 // High threshold so these don't auto-merge
	config.UnclearThreshold = 0.80    // These should be flagged as unclear

	merger := NewMerger(provider, idx, nil, config)

	// Find candidates
	candidates, err := merger.FindCandidates(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Should find one candidate pair
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}

	c := candidates[0]
	if c.Similarity < 0.8 {
		t.Errorf("expected similarity > 0.8, got %f", c.Similarity)
	}
}

func TestMergerClassifyPair(t *testing.T) {
	merger := &Merger{
		config: DefaultMergeConfig(),
	}

	tests := []struct {
		name       string
		similarity float32
		wantAction MergeAction
	}{
		{"high similarity merges", 0.95, ActionMerge},
		{"medium similarity unclear", 0.82, ActionFlagUnclear},
		{"low similarity links", 0.76, ActionFlagUnclear}, // Still in unclear range
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memA := &MemoryFile{
				Frontmatter: Frontmatter{
					ID:       "a",
					Category: "fact",
				},
			}
			memB := &MemoryFile{
				Frontmatter: Frontmatter{
					ID:       "b",
					Category: "fact",
				},
			}

			action, _ := merger.classifyPair(memA, memB, tt.similarity)
			if action != tt.wantAction {
				t.Errorf("got action %s, want %s", action, tt.wantAction)
			}
		})
	}
}

func TestMergerSkipsAlreadyLinked(t *testing.T) {
	merger := &Merger{
		config: DefaultMergeConfig(),
	}

	memA := &MemoryFile{
		Frontmatter: Frontmatter{
			ID:       "a",
			Category: "fact",
			Links: LinksSection{
				Related: []string{"b"},
			},
		},
	}
	memB := &MemoryFile{
		Frontmatter: Frontmatter{
			ID:       "b",
			Category: "fact",
		},
	}

	action, reason := merger.classifyPair(memA, memB, 0.95)
	if action != ActionSkip {
		t.Errorf("expected ActionSkip for already linked, got %s", action)
	}
	if reason != "already linked" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

func TestMergerSkipsSuperseded(t *testing.T) {
	merger := &Merger{
		config: DefaultMergeConfig(),
	}

	memA := &MemoryFile{
		Frontmatter: Frontmatter{
			ID:       "a",
			Category: "fact",
			Supersession: SupersessionSection{
				SupersededBy: "c",
			},
		},
	}
	memB := &MemoryFile{
		Frontmatter: Frontmatter{
			ID:       "b",
			Category: "fact",
		},
	}

	action, reason := merger.classifyPair(memA, memB, 0.95)
	if action != ActionSkip {
		t.Errorf("expected ActionSkip for superseded, got %s", action)
	}
	if reason != "already superseded" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

func TestMergerDryRun(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "merge-dryrun-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize provider
	provider := NewProvider()
	if err := provider.Init(context.Background(), map[string]any{"root": tmpDir}); err != nil {
		t.Fatal(err)
	}
	defer provider.Close()

	config := DefaultMergeConfig()
	config.DryRun = true

	merger := NewMerger(provider, nil, nil, config)

	candidates := []MergeCandidate{
		{
			MemoryA: &MemoryFile{
				Frontmatter: Frontmatter{ID: "a"},
			},
			MemoryB: &MemoryFile{
				Frontmatter: Frontmatter{ID: "b"},
			},
			Similarity: 0.95,
			Action:     ActionMerge,
		},
		{
			MemoryA: &MemoryFile{
				Frontmatter: Frontmatter{ID: "c"},
			},
			MemoryB: &MemoryFile{
				Frontmatter: Frontmatter{ID: "d"},
			},
			Similarity: 0.80,
			Action:     ActionFlagUnclear,
		},
	}

	result, err := merger.Execute(context.Background(), candidates)
	if err != nil {
		t.Fatal(err)
	}

	if result.Merged != 1 {
		t.Errorf("expected 1 merged, got %d", result.Merged)
	}
	if result.FlaggedAsUnclear != 1 {
		t.Errorf("expected 1 flagged, got %d", result.FlaggedAsUnclear)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors in dry run, got %v", result.Errors)
	}
}

func TestDefaultMergeConfig(t *testing.T) {
	config := DefaultMergeConfig()

	if config.SimilarityThreshold != 0.90 {
		t.Errorf("expected similarity threshold 0.90, got %f", config.SimilarityThreshold)
	}
	if config.UnclearThreshold != 0.75 {
		t.Errorf("expected unclear threshold 0.75, got %f", config.UnclearThreshold)
	}
	if config.RequireSameCategory != true {
		t.Error("expected RequireSameCategory true")
	}
	if config.RequireSameScope != true {
		t.Error("expected RequireSameScope true")
	}
	if config.DryRun != false {
		t.Error("expected DryRun false")
	}
}

func TestMergerExecuteMerge(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "merge-exec-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize provider
	provider := NewProvider()
	if err := provider.Init(context.Background(), map[string]any{"root": tmpDir}); err != nil {
		t.Fatal(err)
	}
	defer provider.Close()

	// Create memories
	now := time.Now().UTC()
	oldTime := now.Add(-time.Hour)

	mem1 := providers.Memory{
		ID:        "mem-older",
		Category:  "fact",
		Content:   "First memory content",
		CreatedAt: oldTime,
	}
	mem2 := providers.Memory{
		ID:        "mem-newer",
		Category:  "fact",
		Content:   "Second memory content",
		CreatedAt: now,
	}

	if _, err := provider.Create(context.Background(), mem1); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Create(context.Background(), mem2); err != nil {
		t.Fatal(err)
	}

	// Get memory files from cache
	provider.mu.RLock()
	memA := provider.cache["mem-older"]
	memB := provider.cache["mem-newer"]
	provider.mu.RUnlock()

	config := DefaultMergeConfig()
	merger := NewMerger(provider, nil, nil, config)

	candidate := MergeCandidate{
		MemoryA:    memA,
		MemoryB:    memB,
		Similarity: 0.95,
		Action:     ActionMerge,
	}

	err = merger.executeMerge(context.Background(), candidate)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the newer one is the keeper
	provider.mu.RLock()
	keeperFile := provider.cache["mem-newer"]
	mergedFile := provider.cache["mem-older"]
	provider.mu.RUnlock()

	// Keeper should have combined content
	if keeperFile.Content == "Second memory content" {
		t.Error("expected keeper to have merged content")
	}

	// Merged should be superseded
	if mergedFile.Frontmatter.Status != "superseded" {
		t.Errorf("expected status superseded, got %s", mergedFile.Frontmatter.Status)
	}
	if mergedFile.Frontmatter.Supersession.SupersededBy != "mem-newer" {
		t.Errorf("expected superseded_by mem-newer, got %s", mergedFile.Frontmatter.Supersession.SupersededBy)
	}
}

// testFloat32ToBytes converts float32 slice to bytes for testing.
func testFloat32ToBytes(floats []float32) []byte {
	bytes := make([]byte, len(floats)*4)
	for i, f := range floats {
		bits := *(*uint32)(unsafe.Pointer(&f))
		bytes[i*4] = byte(bits)
		bytes[i*4+1] = byte(bits >> 8)
		bytes[i*4+2] = byte(bits >> 16)
		bytes[i*4+3] = byte(bits >> 24)
	}
	return bytes
}

func TestCosineSimilarityMerge(t *testing.T) {
	// Test that similar vectors produce high similarity
	v1 := []float32{0.9, 0.1, 0.1, 0.1}
	v2 := []float32{0.85, 0.15, 0.1, 0.1}

	sim := cosineSimilarity(v1, v2)
	if sim < 0.9 {
		t.Errorf("expected similar vectors to have similarity > 0.9, got %f", sim)
	}

	// Test orthogonal vectors
	v3 := []float32{1, 0, 0, 0}
	v4 := []float32{0, 1, 0, 0}

	sim2 := cosineSimilarity(v3, v4)
	if math.Abs(float64(sim2)) > 0.01 {
		t.Errorf("expected orthogonal vectors to have ~0 similarity, got %f", sim2)
	}
}
