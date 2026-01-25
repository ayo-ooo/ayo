// Package ollama provides a client for the Ollama API.
package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultHost is the default Ollama API host.
	DefaultHost = "http://localhost:11434"

	// DefaultTimeout is the default HTTP timeout for non-streaming requests.
	DefaultTimeout = 30 * time.Second

	// DefaultModel is the default small model for internal operations.
	DefaultModel = "ministral-3:3b"

	// DefaultEmbeddingModel is the default embedding model.
	DefaultEmbeddingModel = "nomic-embed-text"
)

// Client is an HTTP client for the Ollama API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures the Ollama client.
type Option func(*Client)

// WithHost sets the Ollama API host.
func WithHost(host string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(host, "/")
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new Ollama client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL: DefaultHost,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// IsAvailable checks if the Ollama server is running and responding.
func (c *Client) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Model represents an Ollama model.
type Model struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
}

// ListModelsResponse is the response from /api/tags.
type ListModelsResponse struct {
	Models []Model `json:"models"`
}

// ListModels returns all installed models.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list models: status %d: %s", resp.StatusCode, string(body))
	}

	var result ListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Models, nil
}

// HasModel checks if a model is installed.
func (c *Client) HasModel(ctx context.Context, name string) bool {
	models, err := c.ListModels(ctx)
	if err != nil {
		return false
	}

	// Normalize name for comparison (remove :latest suffix)
	name = normalizeModelName(name)

	for _, m := range models {
		if normalizeModelName(m.Name) == name {
			return true
		}
	}
	return false
}

// normalizeModelName removes the :latest suffix if present.
func normalizeModelName(name string) string {
	name = strings.TrimSuffix(name, ":latest")
	return name
}

// VersionResponse is the response from /api/version.
type VersionResponse struct {
	Version string `json:"version"`
}

// GetVersion returns the Ollama version.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/version", nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("get version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get version: status %d: %s", resp.StatusCode, string(body))
	}

	var result VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Version, nil
}

// Host returns the configured base URL.
func (c *Client) Host() string {
	return c.baseURL
}
