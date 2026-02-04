package zettelkasten

import (
	"context"
	"fmt"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
)

// MergeConfig configures the auto-merge behavior.
type MergeConfig struct {
	// SimilarityThreshold is the cosine similarity above which two memories
	// are considered potential duplicates. Default: 0.90
	SimilarityThreshold float32

	// UnclearThreshold is the similarity range for marking memories as unclear.
	// Memories with similarity between UnclearThreshold and SimilarityThreshold
	// are flagged for user clarification. Default: 0.75
	UnclearThreshold float32

	// MaxMergeAge limits how old a memory can be to be merged (0 = no limit).
	// Older memories are considered more established and won't be merged.
	MaxMergeAge time.Duration

	// RequireSameCategory only merges memories with matching categories.
	RequireSameCategory bool

	// RequireSameScope only merges memories with matching agent/path scope.
	RequireSameScope bool

	// DryRun if true, simulates merge without making changes.
	DryRun bool
}

// DefaultMergeConfig returns sensible defaults.
func DefaultMergeConfig() MergeConfig {
	return MergeConfig{
		SimilarityThreshold: 0.90,
		UnclearThreshold:    0.75,
		RequireSameCategory: true,
		RequireSameScope:    true,
		DryRun:              false,
	}
}

// MergeCandidate represents a potential merge between two memories.
type MergeCandidate struct {
	MemoryA    *MemoryFile
	MemoryB    *MemoryFile
	Similarity float32
	Action     MergeAction
	Reason     string
}

// MergeAction indicates what should happen with a candidate pair.
type MergeAction string

const (
	// ActionMerge means memories are similar enough to merge automatically.
	ActionMerge MergeAction = "merge"

	// ActionFlagUnclear means memories are similar but need user clarification.
	ActionFlagUnclear MergeAction = "flag_unclear"

	// ActionLink means memories are related but distinct.
	ActionLink MergeAction = "link"

	// ActionSkip means no action needed.
	ActionSkip MergeAction = "skip"
)

// MergeResult summarizes the merge operation.
type MergeResult struct {
	Merged       int              // Number of memories merged
	FlaggedAsUnclear int          // Number flagged for clarification
	Linked       int              // Number of new links created
	Candidates   []MergeCandidate // All candidates analyzed
	Errors       []error          // Any errors encountered
}

// Merger finds and merges similar memories.
type Merger struct {
	provider  *Provider
	index     *Index
	embedding providers.EmbeddingProvider
	config    MergeConfig
}

// NewMerger creates a new memory merger.
func NewMerger(provider *Provider, index *Index, embedding providers.EmbeddingProvider, config MergeConfig) *Merger {
	if config.SimilarityThreshold == 0 {
		config.SimilarityThreshold = 0.90
	}
	if config.UnclearThreshold == 0 {
		config.UnclearThreshold = 0.75
	}
	return &Merger{
		provider:  provider,
		index:     index,
		embedding: embedding,
		config:    config,
	}
}

// FindCandidates finds all potential merge candidates.
func (m *Merger) FindCandidates(ctx context.Context) ([]MergeCandidate, error) {
	// Get all active memories
	memories, err := m.provider.listAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list memories: %w", err)
	}

	// Load embeddings from index
	embeddings, err := m.loadEmbeddings(ctx, memories)
	if err != nil {
		return nil, fmt.Errorf("load embeddings: %w", err)
	}

	var candidates []MergeCandidate

	// Compare each pair (O(n^2) - could optimize with approximate nearest neighbors)
	for i := 0; i < len(memories); i++ {
		memA := memories[i]
		embA, ok := embeddings[memA.Frontmatter.ID]
		if !ok || len(embA) == 0 {
			continue
		}

		for j := i + 1; j < len(memories); j++ {
			memB := memories[j]
			embB, ok := embeddings[memB.Frontmatter.ID]
			if !ok || len(embB) == 0 {
				continue
			}

			// Check scoping constraints
			if m.config.RequireSameCategory {
				if memA.Frontmatter.Category != memB.Frontmatter.Category {
					continue
				}
			}
			if m.config.RequireSameScope {
				if memA.Frontmatter.Scope.Agent != memB.Frontmatter.Scope.Agent ||
					memA.Frontmatter.Scope.Path != memB.Frontmatter.Scope.Path {
					continue
				}
			}

			// Calculate similarity
			sim := cosineSimilarity(embA, embB)
			if sim < m.config.UnclearThreshold {
				continue // Not similar enough to consider
			}

			// Determine action based on similarity
			action, reason := m.classifyPair(memA, memB, sim)

			candidates = append(candidates, MergeCandidate{
				MemoryA:    memA,
				MemoryB:    memB,
				Similarity: sim,
				Action:     action,
				Reason:     reason,
			})
		}
	}

	return candidates, nil
}

// classifyPair determines the appropriate action for a memory pair.
func (m *Merger) classifyPair(memA, memB *MemoryFile, similarity float32) (MergeAction, string) {
	// Already linked?
	if contains(memA.Frontmatter.Links.Related, memB.Frontmatter.ID) {
		return ActionSkip, "already linked"
	}

	// One supersedes the other?
	if memA.Frontmatter.Supersession.SupersededBy != "" ||
		memB.Frontmatter.Supersession.SupersededBy != "" {
		return ActionSkip, "already superseded"
	}

	// Already flagged as unclear?
	if memA.Frontmatter.Unclear.Flagged || memB.Frontmatter.Unclear.Flagged {
		return ActionSkip, "already flagged unclear"
	}

	// Check age constraints
	if m.config.MaxMergeAge > 0 {
		cutoff := time.Now().Add(-m.config.MaxMergeAge)
		if memA.Frontmatter.Created.Before(cutoff) && memB.Frontmatter.Created.Before(cutoff) {
			return ActionLink, "both memories too old to merge, linking instead"
		}
	}

	// High similarity = merge
	if similarity >= m.config.SimilarityThreshold {
		return ActionMerge, fmt.Sprintf("similarity %.2f >= threshold %.2f", similarity, m.config.SimilarityThreshold)
	}

	// Medium similarity = flag as unclear
	if similarity >= m.config.UnclearThreshold {
		return ActionFlagUnclear, fmt.Sprintf("similarity %.2f in unclear range (%.2f-%.2f)", 
			similarity, m.config.UnclearThreshold, m.config.SimilarityThreshold)
	}

	return ActionLink, "similar but distinct"
}

// Execute performs the merge operation based on candidates.
func (m *Merger) Execute(ctx context.Context, candidates []MergeCandidate) (*MergeResult, error) {
	result := &MergeResult{
		Candidates: candidates,
	}

	for _, c := range candidates {
		switch c.Action {
		case ActionMerge:
			if m.config.DryRun {
				result.Merged++
				continue
			}
			if err := m.executeMerge(ctx, c); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("merge %s <-> %s: %w", 
					c.MemoryA.Frontmatter.ID, c.MemoryB.Frontmatter.ID, err))
			} else {
				result.Merged++
			}

		case ActionFlagUnclear:
			if m.config.DryRun {
				result.FlaggedAsUnclear++
				continue
			}
			if err := m.flagUnclear(ctx, c); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("flag unclear %s <-> %s: %w", 
					c.MemoryA.Frontmatter.ID, c.MemoryB.Frontmatter.ID, err))
			} else {
				result.FlaggedAsUnclear++
			}

		case ActionLink:
			if m.config.DryRun {
				result.Linked++
				continue
			}
			if err := m.provider.Link(ctx, c.MemoryA.Frontmatter.ID, c.MemoryB.Frontmatter.ID); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("link %s <-> %s: %w", 
					c.MemoryA.Frontmatter.ID, c.MemoryB.Frontmatter.ID, err))
			} else {
				result.Linked++
			}
		}
	}

	return result, nil
}

// executeMerge performs the actual merge of two memories.
// The newer memory supersedes the older one. Content is combined.
func (m *Merger) executeMerge(ctx context.Context, c MergeCandidate) error {
	// Determine which is newer (keeper) and older (merged)
	keeper, merged := c.MemoryA, c.MemoryB
	if c.MemoryB.Frontmatter.Created.After(c.MemoryA.Frontmatter.Created) {
		keeper, merged = c.MemoryB, c.MemoryA
	}

	now := time.Now().UTC()

	// Update keeper with combined content
	if keeper.Content != merged.Content {
		keeper.Content = keeper.Content + "\n\n---\n(Merged from " + merged.Frontmatter.ID + "):\n" + merged.Content
	}
	keeper.Frontmatter.Updated = now

	// Merge topics
	for _, t := range merged.Frontmatter.Topics {
		if !containsString(keeper.Frontmatter.Topics, t) {
			keeper.Frontmatter.Topics = append(keeper.Frontmatter.Topics, t)
		}
	}

	// Add link to merged memory
	if !contains(keeper.Frontmatter.Links.Related, merged.Frontmatter.ID) {
		keeper.Frontmatter.Links.Related = append(keeper.Frontmatter.Links.Related, merged.Frontmatter.ID)
	}

	// Mark merged memory as superseded
	merged.Frontmatter.Supersession.SupersededBy = keeper.Frontmatter.ID
	merged.Frontmatter.Supersession.Reason = fmt.Sprintf("auto-merged due to %.0f%% similarity", c.Similarity*100)
	merged.Frontmatter.Status = "superseded"
	merged.Frontmatter.Updated = now

	// Save both files
	keeperPath := m.provider.structure.MemoryPath(keeper.Frontmatter.ID, keeper.Frontmatter.Category)
	if err := keeper.WriteFile(keeperPath); err != nil {
		return fmt.Errorf("write keeper: %w", err)
	}

	mergedPath := m.provider.structure.MemoryPath(merged.Frontmatter.ID, merged.Frontmatter.Category)
	if err := merged.WriteFile(mergedPath); err != nil {
		return fmt.Errorf("write merged: %w", err)
	}

	// Update cache
	m.provider.mu.Lock()
	m.provider.cache[keeper.Frontmatter.ID] = keeper
	m.provider.cache[merged.Frontmatter.ID] = merged
	m.provider.mu.Unlock()

	return nil
}

// flagUnclear marks both memories as needing clarification.
func (m *Merger) flagUnclear(ctx context.Context, c MergeCandidate) error {
	now := time.Now().UTC()
	reason := fmt.Sprintf("Similar memory found: %s (%.0f%% similarity). Please clarify if these should be merged.",
		c.MemoryB.Frontmatter.ID, c.Similarity*100)

	// Flag memory A
	c.MemoryA.Frontmatter.Unclear.Flagged = true
	c.MemoryA.Frontmatter.Unclear.Reason = reason
	c.MemoryA.Frontmatter.Updated = now

	pathA := m.provider.structure.MemoryPath(c.MemoryA.Frontmatter.ID, c.MemoryA.Frontmatter.Category)
	if err := c.MemoryA.WriteFile(pathA); err != nil {
		return fmt.Errorf("write memory A: %w", err)
	}

	// Flag memory B with reference to A
	reasonB := fmt.Sprintf("Similar memory found: %s (%.0f%% similarity). Please clarify if these should be merged.",
		c.MemoryA.Frontmatter.ID, c.Similarity*100)
	c.MemoryB.Frontmatter.Unclear.Flagged = true
	c.MemoryB.Frontmatter.Unclear.Reason = reasonB
	c.MemoryB.Frontmatter.Updated = now

	pathB := m.provider.structure.MemoryPath(c.MemoryB.Frontmatter.ID, c.MemoryB.Frontmatter.Category)
	if err := c.MemoryB.WriteFile(pathB); err != nil {
		return fmt.Errorf("write memory B: %w", err)
	}

	// Update cache
	m.provider.mu.Lock()
	m.provider.cache[c.MemoryA.Frontmatter.ID] = c.MemoryA
	m.provider.cache[c.MemoryB.Frontmatter.ID] = c.MemoryB
	m.provider.mu.Unlock()

	return nil
}

// loadEmbeddings loads embeddings for all given memories from the index.
func (m *Merger) loadEmbeddings(ctx context.Context, memories []*MemoryFile) (map[string][]float32, error) {
	if m.index == nil {
		return nil, fmt.Errorf("index required for merge operation")
	}

	embeddings := make(map[string][]float32)

	for _, mem := range memories {
		id := mem.Frontmatter.ID

		// Query index for embedding
		row := m.index.db.QueryRowContext(ctx, `
			SELECT embedding FROM memory_index WHERE id = ?
		`, id)

		var embBytes []byte
		if err := row.Scan(&embBytes); err != nil {
			continue // Skip memories without embeddings
		}

		if len(embBytes) > 0 {
			embeddings[id] = bytesToFloat32(embBytes)
		}
	}

	return embeddings, nil
}

// listAllActive returns all active memories from the provider.
func (p *Provider) listAllActive(ctx context.Context) ([]*MemoryFile, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*MemoryFile
	for _, mf := range p.cache {
		if mf.Frontmatter.Status == "" || mf.Frontmatter.Status == "active" {
			result = append(result, mf)
		}
	}
	return result, nil
}

// containsString checks if a slice contains a string.
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
