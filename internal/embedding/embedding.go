// Package embedding provides text embedding functionality for semantic search.
// It supports both local (ONNX-based) and provider-based embeddings.
package embedding

import (
	"context"
	"encoding/binary"
	"errors"
	"math"
)

// ErrModelNotLoaded is returned when attempting to embed without a loaded model.
var ErrModelNotLoaded = errors.New("embedding model not loaded")

// Dimension is the embedding vector dimension for all-MiniLM-L6-v2.
const Dimension = 384

// Embedder generates vector embeddings from text.
type Embedder interface {
	// Embed generates an embedding vector for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch generates embeddings for multiple texts.
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

	// Dimension returns the embedding dimension.
	Dimension() int

	// Close releases resources.
	Close() error
}

// Vector operations for similarity search.

// CosineSimilarity computes the cosine similarity between two vectors.
// Returns a value between -1 and 1, where 1 means identical.
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// CosineDistance computes the cosine distance (1 - similarity).
// Returns a value between 0 and 2, where 0 means identical.
func CosineDistance(a, b []float32) float32 {
	return 1 - CosineSimilarity(a, b)
}

// EuclideanDistance computes the L2 distance between two vectors.
func EuclideanDistance(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var sum float32
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return float32(math.Sqrt(float64(sum)))
}

// Normalize normalizes a vector to unit length (L2 normalization).
func Normalize(v []float32) []float32 {
	if len(v) == 0 {
		return v
	}

	var norm float32
	for _, val := range v {
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))

	if norm == 0 {
		return v
	}

	result := make([]float32, len(v))
	for i, val := range v {
		result[i] = val / norm
	}
	return result
}

// SerializeFloat32 converts a float32 slice to bytes for storage.
func SerializeFloat32(v []float32) []byte {
	buf := make([]byte, len(v)*4)
	for i, val := range v {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(val))
	}
	return buf
}

// DeserializeFloat32 converts bytes back to a float32 slice.
func DeserializeFloat32(data []byte) []float32 {
	if len(data)%4 != 0 {
		return nil
	}
	v := make([]float32, len(data)/4)
	for i := range v {
		v[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))
	}
	return v
}

// SearchResult represents a similarity search result.
type SearchResult struct {
	ID         string
	Similarity float32
	Distance   float32
}

// SearchResults is a sortable slice of search results.
type SearchResults []SearchResult

func (r SearchResults) Len() int           { return len(r) }
func (r SearchResults) Less(i, j int) bool { return r[i].Similarity > r[j].Similarity }
func (r SearchResults) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

// TopK returns the top K results by similarity.
func TopK(results SearchResults, k int) SearchResults {
	if len(results) <= k {
		return results
	}
	return results[:k]
}

// ThresholdFilter filters results by minimum similarity threshold.
func ThresholdFilter(results SearchResults, threshold float32) SearchResults {
	filtered := make(SearchResults, 0, len(results))
	for _, r := range results {
		if r.Similarity >= threshold {
			filtered = append(filtered, r)
		}
	}
	return filtered
}
