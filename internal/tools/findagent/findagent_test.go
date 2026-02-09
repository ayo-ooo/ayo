package findagent

import (
	"context"
	"testing"

	"github.com/alexcabrera/ayo/internal/capabilities"
	"github.com/alexcabrera/ayo/internal/embedding"
)

// mockEmbedder implements embedding.Embedder for testing.
type mockEmbedder struct{}

func (m *mockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.5, 0.5, 0.5}, nil
}

func (m *mockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = []float32{0.5, 0.5, 0.5}
	}
	return result, nil
}

func (m *mockEmbedder) Dimension() int {
	return 3
}

func (m *mockEmbedder) Close() error {
	return nil
}

// mockRepository for testing.
type mockRepository struct {
	caps []capabilities.StoredCapability
}

func (m *mockRepository) GetAllCapabilities(ctx context.Context) ([]capabilities.StoredCapability, error) {
	return m.caps, nil
}

func (m *mockRepository) GetCapabilities(ctx context.Context, agentID string) ([]capabilities.StoredCapability, error) {
	var result []capabilities.StoredCapability
	for _, cap := range m.caps {
		if cap.AgentID == agentID {
			result = append(result, cap)
		}
	}
	return result, nil
}

func (m *mockRepository) SearchCapabilities(ctx context.Context, query string, limit int) ([]capabilities.StoredCapability, error) {
	// Return all for testing
	if limit > 0 && len(m.caps) > limit {
		return m.caps[:limit], nil
	}
	return m.caps, nil
}

func (m *mockRepository) UpdateEmbedding(ctx context.Context, capabilityID string, emb []byte) error {
	return nil
}

// mockSearcher creates a test searcher.
type mockSearcher struct {
	results []capabilities.SearchResult
}

func TestFindAgentParams(t *testing.T) {
	params := FindAgentParams{
		Task:  "review code for security issues",
		Count: 5,
	}

	if params.Task != "review code for security issues" {
		t.Error("task not set correctly")
	}
	if params.Count != 5 {
		t.Error("count not set correctly")
	}
}

func TestAgentMatch(t *testing.T) {
	match := AgentMatch{
		Name:        "@code-reviewer",
		Similarity:  0.95,
		Capability:  "code-review",
		Description: "Reviews code for issues",
	}

	if match.Name != "@code-reviewer" {
		t.Error("name not set correctly")
	}
	if match.Similarity != 0.95 {
		t.Error("similarity not set correctly")
	}
}

func TestFindAgentResultString(t *testing.T) {
	t.Run("with matches", func(t *testing.T) {
		result := FindAgentResult{
			Agents: []AgentMatch{
				{Name: "@agent1", Similarity: 0.9, Capability: "cap1", Description: "Desc 1"},
				{Name: "@agent2", Similarity: 0.8, Capability: "cap2", Description: "Desc 2"},
			},
		}

		s := result.String()
		if s == "" {
			t.Error("expected non-empty string")
		}
		if len(s) < 20 {
			t.Error("expected longer output")
		}
	})

	t.Run("no matches", func(t *testing.T) {
		result := FindAgentResult{NoMatch: true}
		s := result.String()
		if s == "" {
			t.Error("expected non-empty string")
		}
		if len(s) < 10 {
			t.Error("expected message about no matches")
		}
	})
}

func TestToolConfig(t *testing.T) {
	embed := embedding.SerializeFloat32([]float32{0.5, 0.5, 0.5})
	repo := &mockRepository{
		caps: []capabilities.StoredCapability{
			{ID: "cap-1", AgentID: "reviewer", Name: "code-review", Description: "Reviews code", Confidence: 0.9, Embedding: embed},
		},
	}

	_ = repo // Would be used to create a CapabilitySearcher

	cfg := ToolConfig{
		Searcher: nil, // Would be set in real usage
	}

	if cfg.Searcher != nil {
		t.Error("expected nil searcher in empty config")
	}
}
