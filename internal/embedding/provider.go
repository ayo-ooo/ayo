package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ProviderEmbedder generates embeddings using a cloud provider API.
type ProviderEmbedder struct {
	endpoint   string
	apiKey     string
	model      string
	dimension  int
	httpClient *http.Client
}

// ProviderConfig configures a provider-based embedder.
type ProviderConfig struct {
	// Provider is the embedding provider (e.g., "openai", "voyage", "ollama").
	Provider string

	// APIKey is the API key for the provider.
	APIKey string

	// Model is the embedding model to use.
	Model string

	// Endpoint overrides the default API endpoint.
	Endpoint string

	// Dimension is the expected embedding dimension.
	Dimension int
}

// Known provider configurations.
var providerDefaults = map[string]struct {
	endpoint  string
	model     string
	dimension int
	keyEnv    string
}{
	"openai": {
		endpoint:  "https://api.openai.com/v1/embeddings",
		model:     "text-embedding-3-small",
		dimension: 1536,
		keyEnv:    "OPENAI_API_KEY",
	},
	"voyage": {
		endpoint:  "https://api.voyageai.com/v1/embeddings",
		model:     "voyage-2",
		dimension: 1024,
		keyEnv:    "VOYAGE_API_KEY",
	},
	"ollama": {
		endpoint:  "http://localhost:11434/api/embeddings",
		model:     "nomic-embed-text",
		dimension: 768,
		keyEnv:    "",
	},
}

// NewProviderEmbedder creates a new provider-based embedder.
func NewProviderEmbedder(cfg ProviderConfig) (*ProviderEmbedder, error) {
	defaults, ok := providerDefaults[cfg.Provider]
	if !ok && cfg.Endpoint == "" {
		return nil, fmt.Errorf("unknown provider %q and no endpoint specified", cfg.Provider)
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = defaults.endpoint
	}

	model := cfg.Model
	if model == "" && ok {
		model = defaults.model
	}

	dimension := cfg.Dimension
	if dimension == 0 && ok {
		dimension = defaults.dimension
	}

	apiKey := cfg.APIKey
	if apiKey == "" && ok && defaults.keyEnv != "" {
		apiKey = os.Getenv(defaults.keyEnv)
	}

	return &ProviderEmbedder{
		endpoint:  endpoint,
		apiKey:    apiKey,
		model:     model,
		dimension: dimension,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Embed generates an embedding for the given text.
func (e *ProviderEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	results, err := e.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return results[0], nil
}

// EmbedBatch generates embeddings for multiple texts.
func (e *ProviderEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Build request based on provider type
	var body []byte
	var err error

	if e.endpoint == "http://localhost:11434/api/embeddings" {
		// Ollama uses a different API format
		return e.embedOllama(ctx, texts)
	}

	// OpenAI-compatible API
	reqBody := map[string]any{
		"model": e.model,
		"input": texts,
	}
	body, err = json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	embeddings := make([][]float32, len(result.Data))
	for i, d := range result.Data {
		embeddings[i] = Normalize(d.Embedding)
	}

	return embeddings, nil
}

// embedOllama handles Ollama's embedding API.
func (e *ProviderEmbedder) embedOllama(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, 0, len(texts))

	for _, text := range texts {
		reqBody := map[string]any{
			"model":  e.model,
			"prompt": text,
		}
		body, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", e.endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := e.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		var result struct {
			Embedding []float32 `json:"embedding"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		results = append(results, Normalize(result.Embedding))
	}

	return results, nil
}

// Dimension returns the embedding dimension.
func (e *ProviderEmbedder) Dimension() int {
	return e.dimension
}

// Close releases resources.
func (e *ProviderEmbedder) Close() error {
	return nil
}
