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
