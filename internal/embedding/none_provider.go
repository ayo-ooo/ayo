package embedding

import (
	"context"
	"errors"

	"github.com/alexcabrera/ayo/internal/providers"
)

// ErrNoEmbeddingProvider is returned when semantic search is requested without an embedder.
var ErrNoEmbeddingProvider = errors.New("no embedding provider available - semantic search disabled")

// NoneProvider is a fallback embedding provider that disables semantic search.
// It returns empty embeddings and relies on FTS-only search.
type NoneProvider struct {
	name string
}

// NewNoneProvider creates a new none embedding provider.
func NewNoneProvider() *NoneProvider {
	return &NoneProvider{
		name: "none",
	}
}

// Name returns the provider name.
func (p *NoneProvider) Name() string {
	return p.name
}

// Type returns the provider type.
func (p *NoneProvider) Type() providers.ProviderType {
	return providers.ProviderTypeEmbedding
}

// Init initializes the provider (no-op for none provider).
func (p *NoneProvider) Init(ctx context.Context, config map[string]any) error {
	return nil
}

// Close releases any resources (no-op for none provider).
func (p *NoneProvider) Close() error {
	return nil
}

// Embed returns an error indicating semantic search is disabled.
func (p *NoneProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	return nil, ErrNoEmbeddingProvider
}

// EmbedBatch returns an error indicating semantic search is disabled.
func (p *NoneProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, ErrNoEmbeddingProvider
}

// Dimensions returns 0 indicating no embeddings are generated.
func (p *NoneProvider) Dimensions() int {
	return 0
}

// Model returns "none" indicating no model is used.
func (p *NoneProvider) Model() string {
	return "none"
}

// Ensure NoneProvider implements providers.EmbeddingProvider.
var _ providers.EmbeddingProvider = (*NoneProvider)(nil)
