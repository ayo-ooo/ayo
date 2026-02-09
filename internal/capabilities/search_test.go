package capabilities

import (
	"context"
	"errors"
	"testing"

	"github.com/alexcabrera/ayo/internal/embedding"
)

// mockEmbedder implements embedding.Embedder for testing.
type mockEmbedder struct {
	embeddings map[string][]float32
	err        error
}

func (m *mockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	if emb, ok := m.embeddings[text]; ok {
		return emb, nil
	}
	// Return a default embedding based on text length for variety
	return []float32{float32(len(text)) / 100, 0.5, 0.5}, nil
}

func (m *mockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([][]float32, len(texts))
	for i, text := range texts {
		result[i], _ = m.Embed(ctx, text)
	}
	return result, nil
}

func (m *mockEmbedder) Dimension() int {
	return 3
}

func (m *mockEmbedder) Close() error {
	return nil
}

// mockQuerier implements db.Querier for testing.
type mockQuerier struct {
	capabilities []storedCapabilityData
	searchResult []storedCapabilityData
	embeddings   map[string][]byte
}

type storedCapabilityData struct {
	ID          string
	AgentID     string
	Name        string
	Description string
	Confidence  float64
	Source      string
	Embedding   []byte
	InputHash   string
	CreatedAt   int64
	UpdatedAt   int64
}

// mockRepository creates a mock repository for testing.
type mockRepository struct {
	capabilities []StoredCapability
}

func (m *mockRepository) GetAllCapabilities(ctx context.Context) ([]StoredCapability, error) {
	return m.capabilities, nil
}

func (m *mockRepository) GetCapabilities(ctx context.Context, agentID string) ([]StoredCapability, error) {
	var result []StoredCapability
	for _, cap := range m.capabilities {
		if cap.AgentID == agentID {
			result = append(result, cap)
		}
	}
	return result, nil
}

func (m *mockRepository) SearchCapabilities(ctx context.Context, query string, limit int) ([]StoredCapability, error) {
	return m.capabilities, nil
}

func (m *mockRepository) UpdateEmbedding(ctx context.Context, capabilityID string, embedding []byte) error {
	for i, cap := range m.capabilities {
		if cap.ID == capabilityID {
			m.capabilities[i].Embedding = embedding
			break
		}
	}
	return nil
}

// capabilitySearcherWithMock creates a searcher with a mock repository.
type capabilitySearcherWithMock struct {
	repo     *mockRepository
	embedder embedding.Embedder
}

func (s *capabilitySearcherWithMock) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	allCaps, err := s.repo.GetAllCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	if len(allCaps) == 0 {
		return nil, nil
	}

	queryEmbed, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return s.keywordSearch(ctx, query, limit)
	}

	type scoredResult struct {
		cap        StoredCapability
		similarity float32
	}

	var results []scoredResult
	for _, cap := range allCaps {
		if len(cap.Embedding) == 0 {
			continue
		}

		capEmbed := embedding.DeserializeFloat32(cap.Embedding)
		if len(capEmbed) == 0 {
			continue
		}

		similarity := embedding.CosineSimilarity(queryEmbed, capEmbed)
		results = append(results, scoredResult{
			cap:        cap,
			similarity: similarity,
		})
	}

	// Sort by similarity descending
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].similarity > results[i].similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	searchResults := make([]SearchResult, len(results))
	for i, r := range results {
		searchResults[i] = SearchResult{
			AgentID:    r.cap.AgentID,
			Capability: r.cap,
			Similarity: r.similarity,
		}
	}

	return searchResults, nil
}

func (s *capabilitySearcherWithMock) keywordSearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	results, err := s.repo.SearchCapabilities(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	searchResults := make([]SearchResult, len(results))
	for i, cap := range results {
		searchResults[i] = SearchResult{
			AgentID:    cap.AgentID,
			Capability: cap,
			Similarity: float32(cap.Confidence),
		}
	}

	return searchResults, nil
}

func TestSearchResultStruct(t *testing.T) {
	result := SearchResult{
		AgentID:     "agent-1",
		AgentHandle: "@test-agent",
		Capability: StoredCapability{
			ID:          "cap-1",
			Name:        "testing",
			Description: "Test capability",
		},
		Similarity: 0.95,
	}

	if result.AgentID != "agent-1" {
		t.Errorf("expected AgentID 'agent-1', got %q", result.AgentID)
	}
	if result.Similarity != 0.95 {
		t.Errorf("expected Similarity 0.95, got %f", result.Similarity)
	}
}

func TestSearchWithEmbeddings(t *testing.T) {
	// Create embeddings for test capabilities
	codeReviewEmbed := []float32{0.9, 0.1, 0.0}
	securityEmbed := []float32{0.1, 0.9, 0.0}
	testingEmbed := []float32{0.0, 0.1, 0.9}

	mockRepo := &mockRepository{
		capabilities: []StoredCapability{
			{
				ID:          "cap-1",
				AgentID:     "agent-1",
				Name:        "code-review",
				Description: "Review code for quality",
				Confidence:  0.9,
				Embedding:   embedding.SerializeFloat32(codeReviewEmbed),
			},
			{
				ID:          "cap-2",
				AgentID:     "agent-2",
				Name:        "security-analysis",
				Description: "Security vulnerability scanning",
				Confidence:  0.8,
				Embedding:   embedding.SerializeFloat32(securityEmbed),
			},
			{
				ID:          "cap-3",
				AgentID:     "agent-3",
				Name:        "testing",
				Description: "Generate tests",
				Confidence:  0.7,
				Embedding:   embedding.SerializeFloat32(testingEmbed),
			},
		},
	}

	mockEmb := &mockEmbedder{
		embeddings: map[string][]float32{
			"code quality": {0.85, 0.15, 0.0}, // Similar to code-review
		},
	}

	searcher := &capabilitySearcherWithMock{
		repo:     mockRepo,
		embedder: mockEmb,
	}

	results, err := searcher.Search(context.Background(), "code quality", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// First result should be most similar to query
	if results[0].Capability.Name != "code-review" {
		t.Errorf("expected first result to be 'code-review', got %q", results[0].Capability.Name)
	}
}

func TestSearchWithLimit(t *testing.T) {
	embed := embedding.SerializeFloat32([]float32{0.5, 0.5, 0.5})

	mockRepo := &mockRepository{
		capabilities: []StoredCapability{
			{ID: "cap-1", AgentID: "agent-1", Name: "cap1", Embedding: embed},
			{ID: "cap-2", AgentID: "agent-2", Name: "cap2", Embedding: embed},
			{ID: "cap-3", AgentID: "agent-3", Name: "cap3", Embedding: embed},
		},
	}

	mockEmb := &mockEmbedder{
		embeddings: map[string][]float32{
			"query": {0.5, 0.5, 0.5},
		},
	}

	searcher := &capabilitySearcherWithMock{
		repo:     mockRepo,
		embedder: mockEmb,
	}

	// Test with limit of 2
	results, err := searcher.Search(context.Background(), "query", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results (limited), got %d", len(results))
	}
}

func TestSearchFallbackToKeyword(t *testing.T) {
	mockRepo := &mockRepository{
		capabilities: []StoredCapability{
			{ID: "cap-1", AgentID: "agent-1", Name: "code-review", Confidence: 0.9},
			{ID: "cap-2", AgentID: "agent-2", Name: "security", Confidence: 0.8},
		},
	}

	// Embedder that always fails
	mockEmb := &mockEmbedder{
		err: errors.New("embedding service unavailable"),
	}

	searcher := &capabilitySearcherWithMock{
		repo:     mockRepo,
		embedder: mockEmb,
	}

	results, err := searcher.Search(context.Background(), "code", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should fall back to keyword search
	if len(results) != 2 {
		t.Errorf("expected 2 results from keyword fallback, got %d", len(results))
	}

	// Similarity should be based on confidence
	if results[0].Similarity != 0.9 {
		t.Errorf("expected similarity 0.9 from confidence, got %f", results[0].Similarity)
	}
}

func TestSearchEmptyCapabilities(t *testing.T) {
	mockRepo := &mockRepository{
		capabilities: []StoredCapability{},
	}

	mockEmb := &mockEmbedder{}

	searcher := &capabilitySearcherWithMock{
		repo:     mockRepo,
		embedder: mockEmb,
	}

	results, err := searcher.Search(context.Background(), "test", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results != nil {
		t.Errorf("expected nil results for empty capabilities, got %v", results)
	}
}

func TestSearchSkipsCapabilitiesWithoutEmbeddings(t *testing.T) {
	embed := embedding.SerializeFloat32([]float32{0.5, 0.5, 0.5})

	mockRepo := &mockRepository{
		capabilities: []StoredCapability{
			{ID: "cap-1", AgentID: "agent-1", Name: "with-embed", Embedding: embed},
			{ID: "cap-2", AgentID: "agent-2", Name: "without-embed", Embedding: nil},
			{ID: "cap-3", AgentID: "agent-3", Name: "empty-embed", Embedding: []byte{}},
		},
	}

	mockEmb := &mockEmbedder{
		embeddings: map[string][]float32{
			"query": {0.5, 0.5, 0.5},
		},
	}

	searcher := &capabilitySearcherWithMock{
		repo:     mockRepo,
		embedder: mockEmb,
	}

	results, err := searcher.Search(context.Background(), "query", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only capability with valid embedding should be returned
	if len(results) != 1 {
		t.Errorf("expected 1 result (only with embedding), got %d", len(results))
	}

	if results[0].Capability.Name != "with-embed" {
		t.Errorf("expected 'with-embed', got %q", results[0].Capability.Name)
	}
}

func TestNewCapabilitySearcher(t *testing.T) {
	mockEmb := &mockEmbedder{}

	// This test verifies the constructor doesn't panic
	// In a real test, we'd use the actual Repository with a mock db.Querier
	searcher := NewCapabilitySearcher(nil, mockEmb)
	if searcher == nil {
		t.Error("expected non-nil searcher")
	}
	if searcher.embedder != mockEmb {
		t.Error("embedder not set correctly")
	}
}
