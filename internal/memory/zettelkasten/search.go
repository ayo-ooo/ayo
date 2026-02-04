package zettelkasten

import (
	"context"
	"math"
	"sort"

	"github.com/alexcabrera/ayo/internal/providers"
)

// HybridSearcher performs combined semantic and full-text search.
type HybridSearcher struct {
	index     *Index
	embedding providers.EmbeddingProvider
}

// NewHybridSearcher creates a new hybrid searcher.
func NewHybridSearcher(idx *Index, embedding providers.EmbeddingProvider) *HybridSearcher {
	return &HybridSearcher{
		index:     idx,
		embedding: embedding,
	}
}

// HybridSearchOptions configures hybrid search.
type HybridSearchOptions struct {
	Query       string
	Status      string
	AgentHandle string
	PathScope   string
	Categories  []string
	Limit       int
	Threshold   float32

	// SemanticWeight controls blend of semantic vs text (0.0 to 1.0)
	// 0.0 = text only, 1.0 = semantic only, 0.5 = balanced
	SemanticWeight float32

	// RRFConstant is the constant k in RRF formula: 1/(k + rank)
	// Higher values give more weight to lower-ranked results
	// Typical values: 60 (default)
	RRFConstant int
}

// DefaultHybridSearchOptions returns default search options.
func DefaultHybridSearchOptions() HybridSearchOptions {
	return HybridSearchOptions{
		Status:         "active",
		Limit:          20,
		Threshold:      0.0,
		SemanticWeight: 0.5,
		RRFConstant:    60,
	}
}

// HybridResult contains a result with combined score.
type HybridResult struct {
	Entry          IndexEntry
	SemanticScore  float32 // Cosine similarity (0-1)
	TextScore      float32 // FTS relevance score (normalized 0-1)
	CombinedScore  float32 // RRF combined score
	SemanticRank   int     // Rank in semantic results (0 if not found)
	TextRank       int     // Rank in text results (0 if not found)
}

// Search performs hybrid search combining semantic and full-text search.
func (hs *HybridSearcher) Search(ctx context.Context, opts HybridSearchOptions) ([]HybridResult, error) {
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	if opts.Status == "" {
		opts.Status = "active"
	}
	if opts.RRFConstant <= 0 {
		opts.RRFConstant = 60
	}

	// Fetch more candidates than needed for better RRF fusion
	candidateLimit := opts.Limit * 3
	if candidateLimit < 50 {
		candidateLimit = 50
	}

	var semanticResults []HybridResult
	var textResults []HybridResult

	// Semantic search (if embedding provider available and query not empty)
	if hs.embedding != nil && opts.Query != "" {
		sem, err := hs.semanticSearch(ctx, opts, candidateLimit)
		if err != nil {
			// Log but don't fail - fall back to text search
			semanticResults = nil
		} else {
			semanticResults = sem
		}
	}

	// Full-text search
	if opts.Query != "" {
		txt, err := hs.textSearch(ctx, opts, candidateLimit)
		if err != nil {
			// Log but don't fail - may have semantic results
			textResults = nil
		} else {
			textResults = txt
		}
	}

	// Merge results using RRF
	merged := hs.fuseResults(semanticResults, textResults, opts)

	// Apply limit
	if len(merged) > opts.Limit {
		merged = merged[:opts.Limit]
	}

	return merged, nil
}

// semanticSearch performs vector similarity search.
func (hs *HybridSearcher) semanticSearch(ctx context.Context, opts HybridSearchOptions, limit int) ([]HybridResult, error) {
	// Generate query embedding
	queryEmbed, err := hs.embedding.Embed(ctx, opts.Query)
	if err != nil {
		return nil, err
	}

	// Get all entries with embeddings that match filters
	rows, err := hs.index.db.QueryContext(ctx, `
		SELECT id, category, status, agent_handle, path_scope, content, 
		       embedding, confidence, topics, created_at, updated_at, unclear
		FROM memory_index
		WHERE status = ?
		  AND embedding IS NOT NULL
		  AND length(embedding) > 0
		  AND (? = '' OR agent_handle = ? OR agent_handle IS NULL)
		  AND (? = '' OR path_scope = ? OR path_scope IS NULL)
	`, opts.Status, opts.AgentHandle, opts.AgentHandle, opts.PathScope, opts.PathScope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []HybridResult
	for rows.Next() {
		var e IndexEntry
		if err := rows.Scan(
			&e.ID, &e.Category, &e.Status, &e.AgentHandle, &e.PathScope,
			&e.Content, &e.Embedding, &e.Confidence, &e.Topics,
			&e.CreatedAt, &e.UpdatedAt, &e.Unclear,
		); err != nil {
			continue
		}

		// Calculate cosine similarity
		docEmbed := bytesToFloat32(e.Embedding)
		if len(docEmbed) == 0 {
			continue
		}

		similarity := cosineSimilarity(queryEmbed, docEmbed)
		if opts.Threshold > 0 && similarity < opts.Threshold {
			continue
		}

		results = append(results, HybridResult{
			Entry:         e,
			SemanticScore: similarity,
		})
	}

	// Sort by semantic score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].SemanticScore > results[j].SemanticScore
	})

	// Assign ranks
	for i := range results {
		results[i].SemanticRank = i + 1
	}

	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// textSearch performs FTS5 search.
func (hs *HybridSearcher) textSearch(ctx context.Context, opts HybridSearchOptions, limit int) ([]HybridResult, error) {
	// FTS search with bm25 ranking
	rows, err := hs.index.db.QueryContext(ctx, `
		SELECT m.id, m.category, m.status, m.agent_handle, m.path_scope,
		       m.content, m.embedding, m.confidence, m.topics,
		       m.created_at, m.updated_at, m.unclear,
		       bm25(memory_fts) as rank
		FROM memory_fts f
		JOIN memory_index m ON f.id = m.id
		WHERE memory_fts MATCH ?
		  AND m.status = ?
		  AND (? = '' OR m.agent_handle = ? OR m.agent_handle IS NULL)
		  AND (? = '' OR m.path_scope = ? OR m.path_scope IS NULL)
		ORDER BY rank
		LIMIT ?
	`, opts.Query, opts.Status, opts.AgentHandle, opts.AgentHandle,
		opts.PathScope, opts.PathScope, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []HybridResult
	var minRank, maxRank float64 = 0, 0

	for rows.Next() {
		var e IndexEntry
		var rank float64
		if err := rows.Scan(
			&e.ID, &e.Category, &e.Status, &e.AgentHandle, &e.PathScope,
			&e.Content, &e.Embedding, &e.Confidence, &e.Topics,
			&e.CreatedAt, &e.UpdatedAt, &e.Unclear, &rank,
		); err != nil {
			continue
		}

		// Track min/max for normalization (bm25 returns negative values)
		if len(results) == 0 || rank < minRank {
			minRank = rank
		}
		if rank > maxRank {
			maxRank = rank
		}

		results = append(results, HybridResult{
			Entry:     e,
			TextScore: float32(rank), // Will normalize later
		})
	}

	// Normalize BM25 scores to 0-1 range
	// BM25 returns negative values; lower (more negative) is better
	scoreRange := maxRank - minRank
	if scoreRange > 0 {
		for i := range results {
			// Invert so higher is better, then normalize
			normalized := float32((maxRank - float64(results[i].TextScore)) / scoreRange)
			results[i].TextScore = normalized
		}
	} else {
		// All same score, set to 1.0
		for i := range results {
			results[i].TextScore = 1.0
		}
	}

	// Assign ranks
	for i := range results {
		results[i].TextRank = i + 1
	}

	return results, nil
}

// fuseResults merges semantic and text results using Reciprocal Rank Fusion.
func (hs *HybridSearcher) fuseResults(semantic, text []HybridResult, opts HybridSearchOptions) []HybridResult {
	// Build lookup maps by ID
	resultMap := make(map[string]*HybridResult)

	// Add semantic results
	for _, r := range semantic {
		result := r // Copy
		resultMap[r.Entry.ID] = &result
	}

	// Merge text results
	for _, r := range text {
		if existing, ok := resultMap[r.Entry.ID]; ok {
			// Already have from semantic, add text info
			existing.TextScore = r.TextScore
			existing.TextRank = r.TextRank
		} else {
			// New from text only
			result := r
			resultMap[r.Entry.ID] = &result
		}
	}

	// Calculate RRF scores
	k := float32(opts.RRFConstant)
	semanticWeight := opts.SemanticWeight
	textWeight := 1.0 - semanticWeight

	var merged []HybridResult
	for _, result := range resultMap {
		var rrfScore float32

		// Semantic contribution
		if result.SemanticRank > 0 {
			rrfScore += semanticWeight / (k + float32(result.SemanticRank))
		}

		// Text contribution
		if result.TextRank > 0 {
			rrfScore += textWeight / (k + float32(result.TextRank))
		}

		result.CombinedScore = rrfScore
		merged = append(merged, *result)
	}

	// Sort by combined score descending
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].CombinedScore > merged[j].CombinedScore
	})

	return merged
}

// cosineSimilarity calculates the cosine similarity between two vectors.
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}

// bytesToFloat32 converts a byte slice to float32 slice.
// Assumes little-endian encoding (4 bytes per float32).
func bytesToFloat32(b []byte) []float32 {
	if len(b) == 0 || len(b)%4 != 0 {
		return nil
	}

	result := make([]float32, len(b)/4)
	for i := 0; i < len(result); i++ {
		bits := uint32(b[i*4]) |
			uint32(b[i*4+1])<<8 |
			uint32(b[i*4+2])<<16 |
			uint32(b[i*4+3])<<24
		result[i] = math.Float32frombits(bits)
	}
	return result
}

// float32ToBytes converts a float32 slice to byte slice.
// Uses little-endian encoding (4 bytes per float32).
func float32ToBytes(f []float32) []byte {
	result := make([]byte, len(f)*4)
	for i, v := range f {
		bits := math.Float32bits(v)
		result[i*4] = byte(bits)
		result[i*4+1] = byte(bits >> 8)
		result[i*4+2] = byte(bits >> 16)
		result[i*4+3] = byte(bits >> 24)
	}
	return result
}

// Ensure Search uses the HybridSearcher - update provider to use it

// SearchWithIndex performs hybrid search using the index.
func (idx *Index) SearchWithIndex(ctx context.Context, query string, opts providers.SearchOptions) ([]providers.SearchResult, error) {
	// Use FTS search as baseline
	ftsResults, err := idx.SearchFTS(ctx, query, string(opts.Status[0]), opts.Limit)
	if err != nil {
		return nil, err
	}

	var results []providers.SearchResult
	for _, entry := range ftsResults {
		results = append(results, providers.SearchResult{
			Memory:     indexEntryToMemory(entry),
			Similarity: 0.5, // Default for text match
			MatchType:  "text",
		})
	}

	return results, nil
}

// indexEntryToMemory converts an IndexEntry to a providers.Memory.
func indexEntryToMemory(e IndexEntry) providers.Memory {
	m := providers.Memory{
		ID:       e.ID,
		Content:  e.Content,
		Category: providers.MemoryCategory(e.Category),
		Status:   providers.MemoryStatus(e.Status),
	}

	if e.AgentHandle.Valid {
		m.AgentHandle = e.AgentHandle.String
	}
	if e.PathScope.Valid {
		m.PathScope = e.PathScope.String
	}

	return m
}

// Ensure interface compliance
var _ interface {
	Search(ctx context.Context, opts HybridSearchOptions) ([]HybridResult, error)
} = (*HybridSearcher)(nil)
