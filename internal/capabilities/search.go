package capabilities

import (
	"context"
	"sort"

	"github.com/alexcabrera/ayo/internal/embedding"
)

// SearchResult represents a capability search result.
type SearchResult struct {
	// AgentID is the ID of the agent with this capability.
	AgentID string

	// AgentHandle is the handle of the agent (e.g., "@code-reviewer").
	AgentHandle string

	// Capability is the matched capability.
	Capability StoredCapability

	// Similarity is the cosine similarity score (0 to 1).
	Similarity float32
}

// CapabilitySearcher provides semantic search for agent capabilities.
type CapabilitySearcher struct {
	repo     *Repository
	embedder embedding.Embedder
}

// NewCapabilitySearcher creates a new capability searcher.
func NewCapabilitySearcher(repo *Repository, embedder embedding.Embedder) *CapabilitySearcher {
	return &CapabilitySearcher{
		repo:     repo,
		embedder: embedder,
	}
}

// Search finds capabilities semantically similar to the query.
// Results are sorted by similarity (highest first).
func (s *CapabilitySearcher) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Get all capabilities with embeddings
	allCaps, err := s.repo.GetAllCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	if len(allCaps) == 0 {
		return nil, nil
	}

	// Generate embedding for query
	queryEmbed, err := s.embedder.Embed(ctx, query)
	if err != nil {
		// Fall back to keyword matching if embedding fails
		return s.keywordSearch(ctx, query, limit)
	}

	// Calculate similarity for each capability
	type scoredResult struct {
		cap        StoredCapability
		similarity float32
	}

	var results []scoredResult
	for _, cap := range allCaps {
		if len(cap.Embedding) == 0 {
			continue // Skip capabilities without embeddings
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

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].similarity > results[j].similarity
	})

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	// Convert to SearchResult
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

// keywordSearch is a fallback when embedding fails.
func (s *CapabilitySearcher) keywordSearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	results, err := s.repo.SearchCapabilities(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	searchResults := make([]SearchResult, len(results))
	for i, cap := range results {
		searchResults[i] = SearchResult{
			AgentID:    cap.AgentID,
			Capability: cap,
			Similarity: float32(cap.Confidence), // Use confidence as fallback score
		}
	}

	return searchResults, nil
}

// EnsureEmbeddings generates and stores embeddings for capabilities that don't have them.
func (s *CapabilitySearcher) EnsureEmbeddings(ctx context.Context, agentID string) error {
	caps, err := s.repo.GetCapabilities(ctx, agentID)
	if err != nil {
		return err
	}

	for _, cap := range caps {
		if len(cap.Embedding) > 0 {
			continue // Already has embedding
		}

		// Generate embedding for capability description
		text := cap.Name + ": " + cap.Description
		embed, err := s.embedder.Embed(ctx, text)
		if err != nil {
			continue // Skip on error
		}

		// Store embedding
		embedBytes := embedding.SerializeFloat32(embed)

		if err := s.repo.UpdateEmbedding(ctx, cap.ID, embedBytes); err != nil {
			continue // Skip on error
		}
	}

	return nil
}

// FindBestAgent finds the best agent for a task based on capability matching.
func (s *CapabilitySearcher) FindBestAgent(ctx context.Context, taskDescription string) (*SearchResult, error) {
	results, err := s.Search(ctx, taskDescription, 1)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

// FindAgentsForCapability finds agents that have a specific capability.
func (s *CapabilitySearcher) FindAgentsForCapability(ctx context.Context, capabilityName string, limit int) ([]SearchResult, error) {
	// Search for the capability name
	caps, err := s.repo.SearchCapabilities(ctx, capabilityName, limit)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(caps))
	for i, cap := range caps {
		results[i] = SearchResult{
			AgentID:    cap.AgentID,
			Capability: cap,
			Similarity: float32(cap.Confidence),
		}
	}

	return results, nil
}
