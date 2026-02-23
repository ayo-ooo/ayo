package capabilities

import (
	"context"
	"time"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/embedding"
	"github.com/alexcabrera/ayo/internal/squads"
)

// EntityType distinguishes between agents and squads in the index.
type EntityType string

const (
	// EntityTypeAgent is for indexed agents.
	EntityTypeAgent EntityType = "agent"
	// EntityTypeSquad is for indexed squads.
	EntityTypeSquad EntityType = "squad"
)

// IndexedAgent represents an agent in the capability index.
// It stores the agent's description, embedding, and content hash
// for lazy invalidation of embeddings.
type IndexedAgent struct {
	// Handle is the agent handle (e.g., "@crush").
	Handle string

	// Description is the agent's description from config.json.
	Description string

	// ContentHash is SHA256 of description + system.md + skill names.
	// Used for lazy invalidation - if hash changes, re-compute embedding.
	ContentHash string

	// Embedding is the vector representation of the agent's capabilities.
	// Used for semantic search when routing tasks.
	Embedding []float32

	// HasInputSchema indicates if the agent has an input.jsonschema.
	HasInputSchema bool

	// HasOutputSchema indicates if the agent has an output.jsonschema.
	HasOutputSchema bool

	// UpdatedAt is when the index entry was last updated.
	UpdatedAt time.Time
}

// IndexedSquad represents a squad in the capability index.
// It stores the squad's mission, embedding, and content hash
// for lazy invalidation of embeddings.
type IndexedSquad struct {
	// Name is the squad name.
	Name string

	// Mission is extracted from the squad's SQUAD.md constitution.
	Mission string

	// ContentHash is SHA256 of the SQUAD.md content.
	// Used for lazy invalidation - if hash changes, re-compute embedding.
	ContentHash string

	// Embedding is the vector representation of the squad's mission.
	// Used for semantic search when routing tasks.
	Embedding []float32

	// HasInputSchema indicates if the squad has an input.jsonschema.
	HasInputSchema bool

	// HasOutputSchema indicates if the squad has an output.jsonschema.
	HasOutputSchema bool

	// UpdatedAt is when the index entry was last updated.
	UpdatedAt time.Time
}

// EntityIndex provides unified access to agent and squad embeddings.
// It supports lazy invalidation through content hashing - embeddings
// are only recomputed when the underlying content changes.
type EntityIndex struct {
	// Agents is the list of indexed agents.
	Agents []IndexedAgent

	// Squads is the list of indexed squads.
	Squads []IndexedSquad
}

// NewEntityIndex creates a new empty EntityIndex.
func NewEntityIndex() *EntityIndex {
	return &EntityIndex{
		Agents: make([]IndexedAgent, 0),
		Squads: make([]IndexedSquad, 0),
	}
}

// GetAgent returns an indexed agent by handle, or nil if not found.
func (idx *EntityIndex) GetAgent(handle string) *IndexedAgent {
	for i := range idx.Agents {
		if idx.Agents[i].Handle == handle {
			return &idx.Agents[i]
		}
	}
	return nil
}

// GetSquad returns an indexed squad by name, or nil if not found.
func (idx *EntityIndex) GetSquad(name string) *IndexedSquad {
	for i := range idx.Squads {
		if idx.Squads[i].Name == name {
			return &idx.Squads[i]
		}
	}
	return nil
}

// UpsertAgent adds or updates an agent in the index.
// If an agent with the same handle exists, it's replaced.
func (idx *EntityIndex) UpsertAgent(agent IndexedAgent) {
	for i := range idx.Agents {
		if idx.Agents[i].Handle == agent.Handle {
			idx.Agents[i] = agent
			return
		}
	}
	idx.Agents = append(idx.Agents, agent)
}

// UpsertSquad adds or updates a squad in the index.
// If a squad with the same name exists, it's replaced.
func (idx *EntityIndex) UpsertSquad(squad IndexedSquad) {
	for i := range idx.Squads {
		if idx.Squads[i].Name == squad.Name {
			idx.Squads[i] = squad
			return
		}
	}
	idx.Squads = append(idx.Squads, squad)
}

// RemoveAgent removes an agent from the index by handle.
func (idx *EntityIndex) RemoveAgent(handle string) {
	for i := range idx.Agents {
		if idx.Agents[i].Handle == handle {
			idx.Agents = append(idx.Agents[:i], idx.Agents[i+1:]...)
			return
		}
	}
}

// RemoveSquad removes a squad from the index by name.
func (idx *EntityIndex) RemoveSquad(name string) {
	for i := range idx.Squads {
		if idx.Squads[i].Name == name {
			idx.Squads = append(idx.Squads[:i], idx.Squads[i+1:]...)
			return
		}
	}
}

// AgentCount returns the number of agents in the index.
func (idx *EntityIndex) AgentCount() int {
	return len(idx.Agents)
}

// SquadCount returns the number of squads in the index.
func (idx *EntityIndex) SquadCount() int {
	return len(idx.Squads)
}

// TotalCount returns the total number of entities in the index.
func (idx *EntityIndex) TotalCount() int {
	return len(idx.Agents) + len(idx.Squads)
}

// NeedsUpdate returns true if the agent's content hash differs from the stored hash.
func (a *IndexedAgent) NeedsUpdate(currentHash string) bool {
	return a.ContentHash != currentHash
}

// NeedsUpdate returns true if the squad's content hash differs from the stored hash.
func (s *IndexedSquad) NeedsUpdate(currentHash string) bool {
	return s.ContentHash != currentHash
}

// HasEmbedding returns true if the agent has a computed embedding.
func (a *IndexedAgent) HasEmbedding() bool {
	return len(a.Embedding) > 0
}

// HasEmbedding returns true if the squad has a computed embedding.
func (s *IndexedSquad) HasEmbedding() bool {
	return len(s.Embedding) > 0
}

// LazyEntityIndex wraps EntityIndex with an embedder for lazy invalidation.
// When embeddings are requested, it checks the content hash and re-embeds if stale.
type LazyEntityIndex struct {
	*EntityIndex
	embedder embedding.Embedder
}

// NewLazyEntityIndex creates a new lazy entity index with an embedder.
func NewLazyEntityIndex(embedder embedding.Embedder) *LazyEntityIndex {
	return &LazyEntityIndex{
		EntityIndex: NewEntityIndex(),
		embedder:    embedder,
	}
}

// GetAgentEmbedding returns the embedding for an agent, re-computing if stale.
// If the agent's content hash differs from the stored hash, the embedding is recomputed.
func (idx *LazyEntityIndex) GetAgentEmbedding(ctx context.Context, ag agent.Agent) ([]float32, error) {
	currentHash := ComputeAgentHash(ag)

	// Check if we have a cached entry with matching hash
	stored := idx.GetAgent(ag.Handle)
	if stored != nil && !stored.NeedsUpdate(currentHash) && stored.HasEmbedding() {
		return stored.Embedding, nil // Cache hit
	}

	// Re-embed: compute text from agent capabilities
	text := ag.Config.Description
	if ag.System != "" {
		text += "\n\n" + ag.System
	}

	// Truncate to avoid exceeding model context limits
	text = truncateForEmbedding(text)

	emb, err := idx.embedder.Embed(ctx, text)
	if err != nil {
		return nil, err
	}

	// Update index
	idx.UpsertAgent(IndexedAgent{
		Handle:      ag.Handle,
		Description: ag.Config.Description,
		ContentHash: currentHash,
		Embedding:   emb,
		UpdatedAt:   time.Now(),
	})

	return emb, nil
}

// GetSquadEmbedding returns the embedding for a squad, re-computing if stale.
// If the squad's content hash differs from the stored hash, the embedding is recomputed.
func (idx *LazyEntityIndex) GetSquadEmbedding(ctx context.Context, constitution *squads.Constitution, squadName string) ([]float32, error) {
	if constitution == nil {
		return nil, nil
	}

	currentHash := ComputeSquadHash(constitution)

	// Check if we have a cached entry with matching hash
	stored := idx.GetSquad(squadName)
	if stored != nil && !stored.NeedsUpdate(currentHash) && stored.HasEmbedding() {
		return stored.Embedding, nil // Cache hit
	}

	// Re-embed: compute text from squad constitution
	text := truncateForEmbedding(constitution.Raw)

	emb, err := idx.embedder.Embed(ctx, text)
	if err != nil {
		return nil, err
	}

	// Update index
	idx.UpsertSquad(IndexedSquad{
		Name:        squadName,
		Mission:     constitution.Raw, // Use raw content as mission description
		ContentHash: currentHash,
		Embedding:   emb,
		UpdatedAt:   time.Now(),
	})

	return emb, nil
}

// maxEmbeddingChars is the maximum characters to send for embedding.
// nomic-embed-text has an 8192 token limit. For code/markdown with many
// special characters, the ratio is ~2-3 chars/token. We use a conservative
// limit of 16000 chars (~2 chars/token) to ensure we stay within limits.
const maxEmbeddingChars = 16000

// truncateForEmbedding truncates text to fit within embedding model context limits.
func truncateForEmbedding(text string) string {
	if len(text) <= maxEmbeddingChars {
		return text
	}
	// Truncate at word boundary if possible
	truncated := text[:maxEmbeddingChars]
	if lastSpace := len(truncated) - 1; lastSpace > maxEmbeddingChars-100 {
		for i := len(truncated) - 1; i > maxEmbeddingChars-100; i-- {
			if truncated[i] == ' ' || truncated[i] == '\n' {
				truncated = truncated[:i]
				break
			}
		}
	}
	return truncated
}
