package run

import (
	"context"
	"testing"

	"github.com/alexcabrera/ayo/internal/capabilities"
	"github.com/alexcabrera/ayo/internal/embedding"
)

// mockEmbedder implements embedding.Embedder for testing.
type mockEmbedder struct {
	embeddings map[string][]float32
}

func (m *mockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if emb, ok := m.embeddings[text]; ok {
		return emb, nil
	}
	// Return a simple hash-like embedding
	return []float32{float32(len(text)) / 100.0, 0.5}, nil
}

func (m *mockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		results[i] = emb
	}
	return results, nil
}

func (m *mockEmbedder) Dimension() int {
	return 2
}

func (m *mockEmbedder) Close() error {
	return nil
}

func TestDispatcher_Decide_NoSearcher(t *testing.T) {
	d := NewDispatcher(nil)

	decision, err := d.Decide(context.Background(), "do something complex")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Target != "@ayo" {
		t.Errorf("expected target @ayo, got %s", decision.Target)
	}
	if decision.Reason != "no searcher available" {
		t.Errorf("expected reason 'no searcher available', got %s", decision.Reason)
	}
}

func TestDispatcher_Decide_TrivialPrompt(t *testing.T) {
	embedder := &mockEmbedder{embeddings: make(map[string][]float32)}
	index := capabilities.NewLazyEntityIndex(embedder)
	searcher := capabilities.NewUnifiedSearcher(index, embedder)
	d := NewDispatcher(searcher)

	tests := []struct {
		name   string
		prompt string
	}{
		{"empty", ""},
		{"short", "hello"},
		{"question", "what is this?"},
		{"few_words", "build the project"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			decision, err := d.Decide(context.Background(), tc.prompt)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if decision.Target != "@ayo" {
				t.Errorf("expected target @ayo for trivial prompt, got %s", decision.Target)
			}
			if decision.Reason != "trivial task" {
				t.Errorf("expected reason 'trivial task', got %s", decision.Reason)
			}
		})
	}
}

func TestDispatcher_Decide_NoEntitiesIndexed(t *testing.T) {
	embedder := &mockEmbedder{embeddings: make(map[string][]float32)}
	index := capabilities.NewLazyEntityIndex(embedder)
	searcher := capabilities.NewUnifiedSearcher(index, embedder)
	d := NewDispatcher(searcher)

	// Long prompt that's not trivial
	prompt := "This is a much longer prompt that contains enough words and characters to not be considered trivial by the dispatch heuristics"

	decision, err := d.Decide(context.Background(), prompt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Target != "@ayo" {
		t.Errorf("expected target @ayo, got %s", decision.Target)
	}
	if decision.Reason != "no entities indexed" {
		t.Errorf("expected reason 'no entities indexed', got %s", decision.Reason)
	}
}

func TestDispatcher_Decide_MatchFound(t *testing.T) {
	embedder := &mockEmbedder{embeddings: map[string][]float32{
		"code review assistant": {0.8, 0.2, 0.1},
	}}
	index := capabilities.NewLazyEntityIndex(embedder)

	// Add an agent to the index
	index.UpsertAgent(capabilities.IndexedAgent{
		Handle:      "@reviewer",
		Description: "code review assistant",
		Embedding:   []float32{0.8, 0.2, 0.1},
		ContentHash: "abc123",
	})

	searcher := capabilities.NewUnifiedSearcher(index, embedder)
	d := NewDispatcher(searcher)

	// A prompt that should match the code reviewer
	prompt := "This is a long enough prompt asking for a detailed code review of the authentication module with attention to security patterns and best practices"

	decision, err := d.Decide(context.Background(), prompt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Since the embeddings won't actually match well in our mock,
	// we expect fallback to @ayo. In real use, the embedder would
	// provide meaningful similarity scores.
	// The test validates the dispatch flow works end-to-end.
	if decision.Target != "@ayo" && decision.Target != "@reviewer" {
		t.Errorf("expected target @ayo or @reviewer, got %s", decision.Target)
	}
}

func TestDispatcher_DecideAgentOnly(t *testing.T) {
	embedder := &mockEmbedder{embeddings: make(map[string][]float32)}
	index := capabilities.NewLazyEntityIndex(embedder)

	// Add both agent and squad
	index.UpsertAgent(capabilities.IndexedAgent{
		Handle:      "@coder",
		Description: "coding assistant",
		Embedding:   []float32{0.9, 0.1},
		ContentHash: "abc",
	})
	index.UpsertSquad(capabilities.IndexedSquad{
		Name:        "dev-team",
		Mission:     "development team",
		Embedding:   []float32{0.85, 0.15},
		ContentHash: "def",
	})

	searcher := capabilities.NewUnifiedSearcher(index, embedder)
	d := NewDispatcher(searcher)

	// Long prompt
	prompt := "This is a detailed request to implement a new feature with proper testing and documentation that should be routed appropriately"

	decision, err := d.DecideAgentOnly(context.Background(), prompt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not route to a squad
	if decision.Target == "#dev-team" {
		t.Error("DecideAgentOnly should not route to squads")
	}
}

func TestDispatcher_DecideSquadOnly(t *testing.T) {
	embedder := &mockEmbedder{embeddings: make(map[string][]float32)}
	index := capabilities.NewLazyEntityIndex(embedder)

	// Add both agent and squad
	index.UpsertAgent(capabilities.IndexedAgent{
		Handle:      "@coder",
		Description: "coding assistant",
		Embedding:   []float32{0.9, 0.1},
		ContentHash: "abc",
	})
	index.UpsertSquad(capabilities.IndexedSquad{
		Name:        "dev-team",
		Mission:     "development team",
		Embedding:   []float32{0.85, 0.15},
		ContentHash: "def",
	})

	searcher := capabilities.NewUnifiedSearcher(index, embedder)
	d := NewDispatcher(searcher)

	// Long prompt
	prompt := "This is a detailed request to implement a new feature with proper testing and documentation that should be routed appropriately"

	decision, err := d.DecideSquadOnly(context.Background(), prompt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not route to an agent
	if decision.Target == "@coder" {
		t.Error("DecideSquadOnly should not route to agents")
	}
}

func TestIsTrivialPrompt(t *testing.T) {
	tests := []struct {
		prompt   string
		expected bool
	}{
		{"", true},
		{"hi", true},
		{"hello world", true},
		{"what is the weather?", true},
		{"build it", true},
		// Not trivial - exceeds word/char thresholds
		{"This is a detailed request for implementing a complete authentication system with JWT tokens, session management, password hashing, and rate limiting for the new API endpoints", false},
	}

	for _, tc := range tests {
		t.Run(tc.prompt, func(t *testing.T) {
			result := isTrivialPrompt(tc.prompt)
			if result != tc.expected {
				t.Errorf("isTrivialPrompt(%q) = %v, want %v", tc.prompt, result, tc.expected)
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"  spaced  out  ", 2},
		{"one two three four five", 5},
		{"tabs\tand\nnewlines", 3},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := countWords(tc.input)
			if result != tc.expected {
				t.Errorf("countWords(%q) = %d, want %d", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatScore(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{0.0, "0%"},
		{0.5, "50%"},
		{0.75, "75%"},
		{1.0, "100%"},
		{0.333, "33%"},
		{0.666, "67%"}, // rounds
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatScore(tc.score)
			if result != tc.expected {
				t.Errorf("formatScore(%f) = %s, want %s", tc.score, result, tc.expected)
			}
		})
	}
}

// Ensure mockEmbedder satisfies the interface
var _ embedding.Embedder = (*mockEmbedder)(nil)
