// Package memory provides agent memory storage and retrieval.
// Memories are persistent facts, preferences, and patterns learned about users
// that help agents provide more personalized and contextual responses.
package memory

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/alexcabrera/ayo/internal/db"
	"github.com/alexcabrera/ayo/internal/embedding"
	"github.com/google/uuid"
)

// Category represents the type of memory.
type Category string

const (
	CategoryPreference  Category = "preference"  // User preferences
	CategoryFact        Category = "fact"        // Facts about user or project
	CategoryCorrection  Category = "correction"  // User corrections to agent behavior
	CategoryPattern     Category = "pattern"     // Observed behavioral patterns
)

// Status represents the lifecycle state of a memory.
type Status string

const (
	StatusActive     Status = "active"     // Currently active and retrievable
	StatusSuperseded Status = "superseded" // Replaced by a newer memory
	StatusArchived   Status = "archived"   // Manually archived
	StatusForgotten  Status = "forgotten"  // Soft deleted
)

// Memory represents a stored memory.
type Memory struct {
	ID                 string
	AgentHandle        string    // Empty for global memories
	PathScope          string    // Empty for non-path-scoped memories
	Content            string
	Category           Category
	Embedding          []float32
	SourceSessionID    string
	SourceMessageID    string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Confidence         float64
	LastAccessedAt     time.Time
	AccessCount        int64
	SupersedesID       string
	SupersededByID     string
	SupersessionReason string
	Status             Status
}

// SearchResult represents a memory search result with similarity score.
type SearchResult struct {
	Memory     Memory
	Similarity float32
	Distance   float32
}

// SearchOptions configures memory search.
type SearchOptions struct {
	AgentHandle string  // Filter by agent (empty = include global)
	PathScope   string  // Filter by path scope (empty = include global)
	Threshold   float32 // Minimum similarity threshold (0-1)
	Limit       int     // Maximum results
	Categories  []Category // Filter by categories (empty = all)
}

// Service provides memory operations.
type Service struct {
	queries  *db.Queries
	embedder embedding.Embedder
}

// NewService creates a new memory service.
func NewService(queries *db.Queries, embedder embedding.Embedder) *Service {
	return &Service{
		queries:  queries,
		embedder: embedder,
	}
}

// Create stores a new memory with automatic embedding generation.
func (s *Service) Create(ctx context.Context, m Memory) (Memory, error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if m.Status == "" {
		m.Status = StatusActive
	}
	if m.Confidence == 0 {
		m.Confidence = 1.0
	}

	// Generate embedding if embedder is available
	if s.embedder != nil && len(m.Embedding) == 0 {
		emb, err := s.embedder.Embed(ctx, m.Content)
		if err != nil {
			// Log but don't fail - memory can still be stored without embedding
			// The embedding can be generated later
		} else {
			m.Embedding = emb
		}
	}

	err := s.queries.CreateMemory(ctx, db.CreateMemoryParams{
		ID:              m.ID,
		AgentHandle:     toNullString(m.AgentHandle),
		PathScope:       toNullString(m.PathScope),
		Content:         m.Content,
		Category:        string(m.Category),
		Embedding:       embedding.SerializeFloat32(m.Embedding),
		SourceSessionID: toNullString(m.SourceSessionID),
		SourceMessageID: toNullString(m.SourceMessageID),
		CreatedAt:       m.CreatedAt.Unix(),
		UpdatedAt:       m.UpdatedAt.Unix(),
		Confidence:      sql.NullFloat64{Float64: m.Confidence, Valid: true},
		Status:          toNullString(string(m.Status)),
	})
	if err != nil {
		return Memory{}, err
	}

	return m, nil
}

// Get retrieves a memory by ID.
func (s *Service) Get(ctx context.Context, id string) (Memory, error) {
	dbMem, err := s.queries.GetMemory(ctx, id)
	if err != nil {
		return Memory{}, err
	}
	return fromDBMemory(dbMem), nil
}

// GetByPrefix retrieves a memory by ID prefix match.
// If multiple memories match, returns an error.
func (s *Service) GetByPrefix(ctx context.Context, prefix string) (Memory, error) {
	// First try exact match
	mem, err := s.Get(ctx, prefix)
	if err == nil {
		return mem, nil
	}

	// Try prefix match
	memories, err := s.queries.ListMemories(ctx, db.ListMemoriesParams{
		Status: sql.NullString{String: string(StatusActive), Valid: true},
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		return Memory{}, err
	}

	var matches []Memory
	for _, m := range memories {
		if len(m.ID) >= len(prefix) && m.ID[:len(prefix)] == prefix {
			matches = append(matches, fromDBMemory(m))
		}
	}

	if len(matches) == 0 {
		return Memory{}, sql.ErrNoRows
	}
	if len(matches) > 1 {
		return Memory{}, fmt.Errorf("ambiguous prefix: %d memories match", len(matches))
	}
	return matches[0], nil
}

// Update modifies an existing memory.
func (s *Service) Update(ctx context.Context, m Memory) error {
	m.UpdatedAt = time.Now()

	// Regenerate embedding if content changed
	if s.embedder != nil {
		emb, err := s.embedder.Embed(ctx, m.Content)
		if err == nil {
			m.Embedding = emb
		}
	}

	return s.queries.UpdateMemory(ctx, db.UpdateMemoryParams{
		Content:    m.Content,
		Category:   string(m.Category),
		Embedding:  embedding.SerializeFloat32(m.Embedding),
		Confidence: sql.NullFloat64{Float64: m.Confidence, Valid: true},
		UpdatedAt:  m.UpdatedAt.Unix(),
		ID:         m.ID,
	})
}

// Supersede replaces an old memory with a new one, maintaining the chain.
func (s *Service) Supersede(ctx context.Context, oldID string, newMemory Memory, reason string) (Memory, error) {
	newMemory.SupersedesID = oldID

	created, err := s.Create(ctx, newMemory)
	if err != nil {
		return Memory{}, err
	}

	// Mark old memory as superseded
	err = s.queries.SupersedeMemory(ctx, db.SupersedeMemoryParams{
		SupersededByID: toNullString(created.ID),
		UpdatedAt:      time.Now().Unix(),
		ID:             oldID,
	})
	if err != nil {
		return Memory{}, err
	}

	return created, nil
}

// Forget soft-deletes a memory.
func (s *Service) Forget(ctx context.Context, id string) error {
	return s.queries.ForgetMemory(ctx, db.ForgetMemoryParams{
		UpdatedAt: time.Now().Unix(),
		ID:        id,
	})
}

// Delete permanently removes a memory.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.queries.DeleteMemory(ctx, id)
}

// Search performs semantic search over memories.
// If no embedder is configured, returns empty results.
func (s *Service) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	// Without an embedder, we can't do semantic search
	if s.embedder == nil {
		return nil, nil
	}
	
	if opts.Limit == 0 {
		opts.Limit = 10
	}
	if opts.Threshold == 0 {
		opts.Threshold = 0.5
	}

	// Generate query embedding
	queryEmb, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	// Get all candidate memories
	candidates, err := s.queries.GetMemoriesForSearch(ctx, db.GetMemoriesForSearchParams{
		AgentHandle: toNullString(opts.AgentHandle),
		PathScope:   toNullString(opts.PathScope),
	})
	if err != nil {
		return nil, err
	}

	// Calculate similarity for each candidate
	var results []SearchResult
	for _, c := range candidates {
		memEmb := embedding.DeserializeFloat32(c.Embedding)
		if len(memEmb) == 0 {
			continue
		}

		similarity := embedding.CosineSimilarity(queryEmb, memEmb)
		if similarity < opts.Threshold {
			continue
		}

		// Filter by category if specified
		if len(opts.Categories) > 0 {
			found := false
			for _, cat := range opts.Categories {
				if string(cat) == c.Category {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		results = append(results, SearchResult{
			Memory: Memory{
				ID:             c.ID,
				AgentHandle:    fromNullString(c.AgentHandle),
				PathScope:      fromNullString(c.PathScope),
				Content:        c.Content,
				Category:       Category(c.Category),
				Embedding:      memEmb,
				Confidence:     c.Confidence.Float64,
				LastAccessedAt: time.Unix(c.LastAccessedAt.Int64, 0),
				AccessCount:    c.AccessCount.Int64,
				CreatedAt:      time.Unix(c.CreatedAt, 0),
			},
			Similarity: similarity,
			Distance:   embedding.CosineDistance(queryEmb, memEmb),
		})
	}

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Apply limit
	if len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	// Update access timestamps for returned results
	now := time.Now().Unix()
	for _, r := range results {
		_ = s.queries.UpdateMemoryAccess(ctx, db.UpdateMemoryAccessParams{
			LastAccessedAt: sql.NullInt64{Int64: now, Valid: true},
			ID:             r.Memory.ID,
		})
	}

	return results, nil
}

// List returns memories with optional filtering.
func (s *Service) List(ctx context.Context, agentHandle string, limit, offset int64) ([]Memory, error) {
	var dbMems []db.Memory
	var err error

	if agentHandle != "" {
		dbMems, err = s.queries.ListMemoriesByAgent(ctx, db.ListMemoriesByAgentParams{
			AgentHandle: toNullString(agentHandle),
			Status:      sql.NullString{}, // defaults to 'active'
			Limit:       limit,
			Offset:      offset,
		})
	} else {
		dbMems, err = s.queries.ListMemories(ctx, db.ListMemoriesParams{
			Status: sql.NullString{}, // defaults to 'active'
			Limit:  limit,
			Offset: offset,
		})
	}

	if err != nil {
		return nil, err
	}

	memories := make([]Memory, len(dbMems))
	for i, m := range dbMems {
		memories[i] = fromDBMemory(m)
	}
	return memories, nil
}

// Count returns the total number of active memories.
func (s *Service) Count(ctx context.Context, agentHandle string) (int64, error) {
	if agentHandle != "" {
		return s.queries.CountMemoriesByAgent(ctx, db.CountMemoriesByAgentParams{
			AgentHandle: toNullString(agentHandle),
			Status:      sql.NullString{},
		})
	}
	return s.queries.CountMemories(ctx, sql.NullString{})
}

// Clear removes all memories for an agent (or all if agentHandle is empty).
func (s *Service) Clear(ctx context.Context, agentHandle string) error {
	now := time.Now().Unix()
	if agentHandle != "" {
		return s.queries.ClearMemoriesByAgent(ctx, db.ClearMemoriesByAgentParams{
			UpdatedAt:   now,
			AgentHandle: toNullString(agentHandle),
		})
	}
	return s.queries.ClearAllMemories(ctx, now)
}

// Helper functions

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func fromNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func fromDBMemory(m db.Memory) Memory {
	return Memory{
		ID:                 m.ID,
		AgentHandle:        fromNullString(m.AgentHandle),
		PathScope:          fromNullString(m.PathScope),
		Content:            m.Content,
		Category:           Category(m.Category),
		Embedding:          embedding.DeserializeFloat32(m.Embedding),
		SourceSessionID:    fromNullString(m.SourceSessionID),
		SourceMessageID:    fromNullString(m.SourceMessageID),
		CreatedAt:          time.Unix(m.CreatedAt, 0),
		UpdatedAt:          time.Unix(m.UpdatedAt, 0),
		Confidence:         m.Confidence.Float64,
		LastAccessedAt:     time.Unix(m.LastAccessedAt.Int64, 0),
		AccessCount:        m.AccessCount.Int64,
		SupersedesID:       fromNullString(m.SupersedesID),
		SupersededByID:     fromNullString(m.SupersededByID),
		SupersessionReason: fromNullString(m.SupersessionReason),
		Status:             Status(fromNullString(m.Status)),
	}
}
