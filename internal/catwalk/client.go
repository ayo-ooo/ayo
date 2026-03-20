package catwalk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"charm.land/catwalk/pkg/catwalk"
	"charm.land/catwalk/pkg/embedded"
)

const (
	defaultCacheTTL = 24 * time.Hour
	cacheFileName   = "catwalk-cache.json"
)

var (
	ErrNotModified = errors.New("not modified")
)

// Client wraps the catwalk client with local caching and API key detection.
type Client struct {
	client  *catwalk.Client
	cache   *Cache
	cacheMu sync.RWMutex
}

// Cache holds the cached provider data.
type Cache struct {
	Providers []catwalk.Provider `json:"providers"`
	ETag      string             `json:"etag"`
	CachedAt  time.Time          `json:"cached_at"`
}

// NewClient creates a new catwalk client with caching.
func NewClient() *Client {
	return &Client{
		client: catwalk.New(),
	}
}

// GetProviders returns providers, using cache if valid, fetching if needed.
// Falls back to embedded providers on error.
func (c *Client) GetProviders(ctx context.Context) ([]catwalk.Provider, error) {
	// Try to load from cache first
	if cache, err := c.loadCache(); err == nil && c.isCacheValid(cache) {
		return cache.Providers, nil
	}

	// Fetch from remote
	providers, etag, err := c.fetchProviders(ctx)
	if err != nil {
		// On any error, use embedded providers as fallback
		return embedded.GetAll(), nil
	}

	// Save to cache
	c.saveCache(&Cache{
		Providers: providers,
		ETag:      etag,
		CachedAt:  time.Now(),
	})

	return providers, nil
}

// GetAvailableProviders returns providers that have API keys configured.
func (c *Client) GetAvailableProviders(ctx context.Context) ([]catwalk.Provider, error) {
	providers, err := c.GetProviders(ctx)
	if err != nil {
		return nil, err
	}

	var available []catwalk.Provider
	for _, p := range providers {
		if c.hasAPIKey(p) {
			available = append(available, p)
		}
	}

	return available, nil
}

// hasAPIKey checks if an API key is configured for the provider.
func (c *Client) hasAPIKey(p catwalk.Provider) bool {
	envVar := GetAPIKeyEnv(string(p.ID))
	return strings.TrimSpace(os.Getenv(envVar)) != ""
}

// fetchProviders fetches providers from the remote catwalk service.
func (c *Client) fetchProviders(ctx context.Context) ([]catwalk.Provider, string, error) {
	cache, _ := c.loadCache()
	etag := ""
	if cache != nil {
		etag = cache.ETag
	}

	providers, err := c.client.GetProviders(ctx, etag)
	if err != nil {
		if errors.Is(err, catwalk.ErrNotModified) {
			return nil, "", ErrNotModified
		}
		return nil, "", err
	}

	// Generate ETag from the response
	data, err := json.Marshal(providers)
	if err != nil {
		return providers, "", nil
	}
	newETag := catwalk.Etag(data)

	return providers, newETag, nil
}

// isCacheValid checks if the cache is still valid.
func (c *Client) isCacheValid(cache *Cache) bool {
	if cache == nil {
		return false
	}
	return time.Since(cache.CachedAt) < defaultCacheTTL
}

// loadCache loads the cache from disk.
func (c *Client) loadCache() (*Cache, error) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	if c.cache != nil {
		return c.cache, nil
	}

	cachePath, err := getCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	c.cache = &cache
	return &cache, nil
}

// saveCache saves the cache to disk.
func (c *Client) saveCache(cache *Cache) error {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	c.cache = cache

	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

// getCachePath returns the path to the cache file.
func getCachePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config dir: %w", err)
	}
	return filepath.Join(configDir, "ayo", cacheFileName), nil
}

// GetAPIKeyEnv returns the environment variable name for a provider's API key.
func GetAPIKeyEnv(providerID string) string {
	switch providerID {
	case "openai":
		return "OPENAI_API_KEY"
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	case "gemini":
		return "GEMINI_API_KEY"
	case "groq":
		return "GROQ_API_KEY"
	case "openrouter":
		return "OPENROUTER_API_KEY"
	case "xai":
		return "XAI_API_KEY"
	case "zai":
		return "ZAI_API_KEY"
	case "zhipu":
		return "ZHIPU_API_KEY"
	case "zhipu-coding":
		return "ZHIPU_API_KEY"
	case "cerebras":
		return "CEREBRAS_API_KEY"
	case "venice":
		return "VENICE_API_KEY"
	case "deepseek":
		return "DEEPSEEK_API_KEY"
	case "huggingface":
		return "HUGGINGFACE_API_KEY"
	case "aihubmix":
		return "AIHUBMIX_API_KEY"
	case "kimi-coding":
		return "KIMI_API_KEY"
	case "copilot":
		return "GITHUB_TOKEN"
	case "vercel":
		return "VERCEL_API_KEY"
	case "minimax":
		return "MINIMAX_API_KEY"
	case "ionet":
		return "IONET_API_KEY"
	case "qiniucloud":
		return "QINIUCLOUD_API_KEY"
	case "avian":
		return "AVIAN_API_KEY"
	default:
		return strings.ToUpper(providerID) + "_API_KEY"
	}
}

// ProviderWithKey combines a provider with its API key status.
type ProviderWithKey struct {
	Provider   catwalk.Provider
	HasAPIKey  bool
	APIKeyEnv  string
}

// GetProvidersWithKeys returns providers with their API key status.
func (c *Client) GetProvidersWithKeys(ctx context.Context) ([]ProviderWithKey, error) {
	providers, err := c.GetProviders(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ProviderWithKey, len(providers))
	for i, p := range providers {
		envVar := GetAPIKeyEnv(string(p.ID))
		hasKey := strings.TrimSpace(os.Getenv(envVar)) != ""
		result[i] = ProviderWithKey{
			Provider:  p,
			HasAPIKey: hasKey,
			APIKeyEnv: envVar,
		}
	}

	return result, nil
}
