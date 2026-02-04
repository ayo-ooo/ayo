package embedding

import (
	"context"
	"errors"
	"testing"

	"github.com/alexcabrera/ayo/internal/providers"
)

func TestOllamaProviderInterface(t *testing.T) {
	// Verify that OllamaProvider implements providers.EmbeddingProvider
	var _ providers.EmbeddingProvider = (*OllamaProvider)(nil)
}

func TestOllamaProviderName(t *testing.T) {
	p := NewOllamaProvider()
	if p.Name() != "ollama" {
		t.Errorf("expected name 'ollama', got %q", p.Name())
	}
}

func TestOllamaProviderType(t *testing.T) {
	p := NewOllamaProvider()
	if p.Type() != providers.ProviderTypeEmbedding {
		t.Errorf("expected type 'embedding', got %q", p.Type())
	}
}

func TestOllamaProviderDimensions(t *testing.T) {
	p := NewOllamaProvider()
	// Without init, should return default dimensions
	dims := p.Dimensions()
	if dims != 768 { // nomic-embed-text default
		t.Errorf("expected 768 dimensions for default model, got %d", dims)
	}
}

func TestOllamaProviderModel(t *testing.T) {
	p := NewOllamaProvider()
	// Without init, should return default model
	model := p.Model()
	if model != "nomic-embed-text" {
		t.Errorf("expected 'nomic-embed-text', got %q", model)
	}
}

func TestOllamaProviderEmbedWithoutInit(t *testing.T) {
	p := NewOllamaProvider()
	_, err := p.Embed(context.Background(), "test")
	if err == nil {
		t.Error("expected error for embed without init")
	}
}

func TestOllamaProviderEmbedBatchWithoutInit(t *testing.T) {
	p := NewOllamaProvider()
	_, err := p.EmbedBatch(context.Background(), []string{"test1", "test2"})
	if err == nil {
		t.Error("expected error for embed batch without init")
	}
}

func TestOllamaProviderClose(t *testing.T) {
	p := NewOllamaProvider()
	// Close without init should not error
	if err := p.Close(); err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}

func TestNoneProviderInterface(t *testing.T) {
	// Verify that NoneProvider implements providers.EmbeddingProvider
	var _ providers.EmbeddingProvider = (*NoneProvider)(nil)
}

func TestNoneProviderName(t *testing.T) {
	p := NewNoneProvider()
	if p.Name() != "none" {
		t.Errorf("expected name 'none', got %q", p.Name())
	}
}

func TestNoneProviderType(t *testing.T) {
	p := NewNoneProvider()
	if p.Type() != providers.ProviderTypeEmbedding {
		t.Errorf("expected type 'embedding', got %q", p.Type())
	}
}

func TestNoneProviderInit(t *testing.T) {
	p := NewNoneProvider()
	if err := p.Init(context.Background(), nil); err != nil {
		t.Errorf("unexpected error on init: %v", err)
	}
}

func TestNoneProviderClose(t *testing.T) {
	p := NewNoneProvider()
	if err := p.Close(); err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}

func TestNoneProviderEmbed(t *testing.T) {
	p := NewNoneProvider()
	_ = p.Init(context.Background(), nil)

	result, err := p.Embed(context.Background(), "test")
	if err == nil {
		t.Error("expected error from none provider embed")
	}
	if !errors.Is(err, ErrNoEmbeddingProvider) {
		t.Errorf("expected ErrNoEmbeddingProvider, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestNoneProviderEmbedBatch(t *testing.T) {
	p := NewNoneProvider()
	_ = p.Init(context.Background(), nil)

	result, err := p.EmbedBatch(context.Background(), []string{"test1", "test2"})
	if err == nil {
		t.Error("expected error from none provider embed batch")
	}
	if !errors.Is(err, ErrNoEmbeddingProvider) {
		t.Errorf("expected ErrNoEmbeddingProvider, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestNoneProviderDimensions(t *testing.T) {
	p := NewNoneProvider()
	if p.Dimensions() != 0 {
		t.Errorf("expected 0 dimensions, got %d", p.Dimensions())
	}
}

func TestNoneProviderModel(t *testing.T) {
	p := NewNoneProvider()
	if p.Model() != "none" {
		t.Errorf("expected model 'none', got %q", p.Model())
	}
}
