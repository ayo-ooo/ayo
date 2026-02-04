package embedding

import (
	"context"
	"fmt"

	"github.com/alexcabrera/ayo/internal/ollama"
	"github.com/alexcabrera/ayo/internal/providers"
)

// OllamaProvider implements providers.EmbeddingProvider using Ollama.
// It wraps the OllamaEmbedder and implements the full provider interface.
type OllamaProvider struct {
	embedder *OllamaEmbedder
	name     string
}

// NewOllamaProvider creates a new Ollama embedding provider.
func NewOllamaProvider() *OllamaProvider {
	return &OllamaProvider{
		name: "ollama",
	}
}

// Name returns the provider name.
func (p *OllamaProvider) Name() string {
	return p.name
}

// Type returns the provider type.
func (p *OllamaProvider) Type() providers.ProviderType {
	return providers.ProviderTypeEmbedding
}

// Init initializes the provider with configuration.
// Accepted config keys:
//   - host: Ollama server URL (default: http://localhost:11434)
//   - model: Embedding model name (default: nomic-embed-text)
func (p *OllamaProvider) Init(ctx context.Context, config map[string]any) error {
	cfg := OllamaConfig{}

	if host, ok := config["host"].(string); ok && host != "" {
		cfg.Host = host
	}
	if model, ok := config["model"].(string); ok && model != "" {
		cfg.Model = model
	}

	p.embedder = NewOllamaEmbedder(cfg)

	// Verify connectivity
	if !p.embedder.IsAvailable(ctx) {
		return fmt.Errorf("Ollama not available at %s", p.embedder.client.Host())
	}

	// Check if embedding model is available
	if !p.embedder.HasModel(ctx) {
		return fmt.Errorf("embedding model %s not installed (run: ollama pull %s)",
			p.embedder.model, p.embedder.model)
	}

	return nil
}

// Close releases any resources.
func (p *OllamaProvider) Close() error {
	if p.embedder != nil {
		return p.embedder.Close()
	}
	return nil
}

// Embed generates an embedding for a single text.
func (p *OllamaProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if p.embedder == nil {
		return nil, fmt.Errorf("provider not initialized")
	}
	return p.embedder.Embed(ctx, text)
}

// EmbedBatch generates embeddings for multiple texts.
func (p *OllamaProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if p.embedder == nil {
		return nil, fmt.Errorf("provider not initialized")
	}
	return p.embedder.EmbedBatch(ctx, texts)
}

// Dimensions returns the dimensionality of embeddings.
func (p *OllamaProvider) Dimensions() int {
	if p.embedder == nil {
		return ollama.EmbedDimension(ollama.DefaultEmbeddingModel)
	}
	return p.embedder.Dimension()
}

// Model returns the name of the embedding model being used.
func (p *OllamaProvider) Model() string {
	if p.embedder == nil {
		return ollama.DefaultEmbeddingModel
	}
	return p.embedder.Model()
}

// Ensure OllamaProvider implements providers.EmbeddingProvider.
var _ providers.EmbeddingProvider = (*OllamaProvider)(nil)
