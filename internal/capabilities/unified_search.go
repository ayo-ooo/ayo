package capabilities

import (
	"context"
	"sort"

	"github.com/alexcabrera/ayo/internal/embedding"
)

// EntityType constants for unified search results.
const (
	EntityTypeAgentStr = "agent"
	EntityTypeSquadStr = "squad"
)

// UnifiedSearchResult represents a search result that can be either an agent or squad.
type UnifiedSearchResult struct {
	// Type is either "agent" or "squad".
	Type string

	// Handle is the identifier (@agent or #squad).
	Handle string

	// Description is a brief description of the entity.
	Description string

	// Score is the cosine similarity score (0 to 1).
	Score float64

	// HasInputSchema indicates if the entity has an input schema.
	HasInputSchema bool

	// HasOutputSchema indicates if the entity has an output schema.
	HasOutputSchema bool
}

// UnifiedSearcher provides semantic search across both agents and squads.
type UnifiedSearcher struct {
	index    *LazyEntityIndex
	embedder embedding.Embedder
}

// NewUnifiedSearcher creates a new unified searcher.
func NewUnifiedSearcher(index *LazyEntityIndex, embedder embedding.Embedder) *UnifiedSearcher {
	return &UnifiedSearcher{
		index:    index,
		embedder: embedder,
	}
}

// Search finds agents and squads semantically similar to the query.
// Results are sorted by score (highest first).
func (s *UnifiedSearcher) Search(ctx context.Context, query string, limit int) ([]UnifiedSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	// Generate embedding for query
	queryEmb, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	var results []UnifiedSearchResult

	// Score all agents
	for _, agent := range s.index.Agents {
		if !agent.HasEmbedding() {
			continue
		}
		score := embedding.CosineSimilarity(queryEmb, agent.Embedding)
		results = append(results, UnifiedSearchResult{
			Type:            EntityTypeAgentStr,
			Handle:          agent.Handle,
			Description:     agent.Description,
			Score:           float64(score),
			HasInputSchema:  agent.HasInputSchema,
			HasOutputSchema: agent.HasOutputSchema,
		})
	}

	// Score all squads
	for _, squad := range s.index.Squads {
		if !squad.HasEmbedding() {
			continue
		}
		score := embedding.CosineSimilarity(queryEmb, squad.Embedding)
		results = append(results, UnifiedSearchResult{
			Type:            EntityTypeSquadStr,
			Handle:          "#" + squad.Name,
			Description:     squad.Mission,
			Score:           float64(score),
			HasInputSchema:  squad.HasInputSchema,
			HasOutputSchema: squad.HasOutputSchema,
		})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// FindBest returns the single best matching entity for the query.
// Returns nil if no entities are indexed.
func (s *UnifiedSearcher) FindBest(ctx context.Context, query string) (*UnifiedSearchResult, error) {
	results, err := s.Search(ctx, query, 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil // No match, caller should handle
	}
	return &results[0], nil
}

// SearchAgentsOnly searches only agents, not squads.
func (s *UnifiedSearcher) SearchAgentsOnly(ctx context.Context, query string, limit int) ([]UnifiedSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	queryEmb, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	var results []UnifiedSearchResult

	for _, agent := range s.index.Agents {
		if !agent.HasEmbedding() {
			continue
		}
		score := embedding.CosineSimilarity(queryEmb, agent.Embedding)
		results = append(results, UnifiedSearchResult{
			Type:            EntityTypeAgentStr,
			Handle:          agent.Handle,
			Description:     agent.Description,
			Score:           float64(score),
			HasInputSchema:  agent.HasInputSchema,
			HasOutputSchema: agent.HasOutputSchema,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// SearchSquadsOnly searches only squads, not agents.
func (s *UnifiedSearcher) SearchSquadsOnly(ctx context.Context, query string, limit int) ([]UnifiedSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	queryEmb, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	var results []UnifiedSearchResult

	for _, squad := range s.index.Squads {
		if !squad.HasEmbedding() {
			continue
		}
		score := embedding.CosineSimilarity(queryEmb, squad.Embedding)
		results = append(results, UnifiedSearchResult{
			Type:            EntityTypeSquadStr,
			Handle:          "#" + squad.Name,
			Description:     squad.Mission,
			Score:           float64(score),
			HasInputSchema:  squad.HasInputSchema,
			HasOutputSchema: squad.HasOutputSchema,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// IndexedEntityCount returns the total number of indexed entities.
func (s *UnifiedSearcher) IndexedEntityCount() int {
	return s.index.TotalCount()
}

// AgentCount returns the number of indexed agents.
func (s *UnifiedSearcher) AgentCount() int {
	return s.index.AgentCount()
}

// SquadCount returns the number of indexed squads.
func (s *UnifiedSearcher) SquadCount() int {
	return s.index.SquadCount()
}
