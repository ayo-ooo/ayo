package embedding

import (
	"context"

	"github.com/alexcabrera/ayo/internal/ollama"
)

// OllamaEmbedder implements the Embedder interface using Ollama.
type OllamaEmbedder struct {
	client *ollama.Client
	model  string
}

// OllamaConfig configures the Ollama embedder.
type OllamaConfig struct {
	Host  string // Ollama host (default: http://localhost:11434)
	Model string // Embedding model (default: nomic-embed-text)
}

// NewOllamaEmbedder creates a new Ollama-based embedder.
func NewOllamaEmbedder(cfg OllamaConfig) *OllamaEmbedder {
	opts := []ollama.Option{}
	if cfg.Host != "" {
		opts = append(opts, ollama.WithHost(cfg.Host))
	}

	model := cfg.Model
	if model == "" {
		model = ollama.DefaultEmbeddingModel
	}

	return &OllamaEmbedder{
		client: ollama.NewClient(opts...),
		model:  model,
	}
}

// Embed generates an embedding for a single text.
func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	return e.client.Embed(ctx, e.model, text)
}

// EmbedBatch generates embeddings for multiple texts.
func (e *OllamaEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	return e.client.EmbedBatch(ctx, e.model, texts)
}

// Dimension returns the embedding dimension.
func (e *OllamaEmbedder) Dimension() int {
	return ollama.EmbedDimension(e.model)
}

// Close releases resources (no-op for Ollama).
func (e *OllamaEmbedder) Close() error {
	return nil
}

// IsAvailable checks if the Ollama server is available.
func (e *OllamaEmbedder) IsAvailable(ctx context.Context) bool {
	return e.client.IsAvailable(ctx)
}

// HasModel checks if the embedding model is installed.
func (e *OllamaEmbedder) HasModel(ctx context.Context) bool {
	return e.client.HasModel(ctx, e.model)
}

// Model returns the model name.
func (e *OllamaEmbedder) Model() string {
	return e.model
}
