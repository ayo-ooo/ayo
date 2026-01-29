// Package config provides configuration loading and credential management.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
)

// DetectedProvider represents a provider with its credential status.
type DetectedProvider struct {
	ID      string // catwalk ID: "anthropic", "openai", etc.
	Name    string // Display name: "Anthropic (Claude)"
	EnvVar  string // Environment variable name
	HasKey  bool   // Whether env var is set
}

// knownProviders lists all providers we can detect.
// Based on catwalk's embedded provider configurations.
var knownProviders = []DetectedProvider{
	{ID: "anthropic", Name: "Anthropic (Claude)", EnvVar: "ANTHROPIC_API_KEY"},
	{ID: "openai", Name: "OpenAI (GPT-4)", EnvVar: "OPENAI_API_KEY"},
	{ID: "google", Name: "Google (Gemini)", EnvVar: "GEMINI_API_KEY"},
	{ID: "openrouter", Name: "OpenRouter", EnvVar: "OPENROUTER_API_KEY"},
	{ID: "azure", Name: "Azure OpenAI", EnvVar: "AZURE_OPENAI_API_KEY"},
	{ID: "groq", Name: "Groq", EnvVar: "GROQ_API_KEY"},
	{ID: "deepseek", Name: "DeepSeek", EnvVar: "DEEPSEEK_API_KEY"},
	{ID: "cerebras", Name: "Cerebras", EnvVar: "CEREBRAS_API_KEY"},
	{ID: "xai", Name: "xAI (Grok)", EnvVar: "XAI_API_KEY"},
	{ID: "together", Name: "Together.ai", EnvVar: "TOGETHER_API_KEY"},
}

// DetectProviders checks which providers have API keys available.
// It checks both environment variables and stored credentials.
func DetectProviders() []DetectedProvider {
	result := make([]DetectedProvider, len(knownProviders))
	for i, p := range knownProviders {
		result[i] = DetectedProvider{
			ID:     p.ID,
			Name:   p.Name,
			EnvVar: p.EnvVar,
			HasKey: os.Getenv(p.EnvVar) != "",
		}
	}
	return result
}

// HasAnyProvider returns true if at least one provider has credentials.
func HasAnyProvider() bool {
	for _, p := range DetectProviders() {
		if p.HasKey {
			return true
		}
	}
	return false
}

// GetProvidersWithCredentials returns only providers that have API keys.
func GetProvidersWithCredentials() []DetectedProvider {
	var result []DetectedProvider
	for _, p := range DetectProviders() {
		if p.HasKey {
			result = append(result, p)
		}
	}
	return result
}

// ProviderCredential stores a single provider's API key.
type ProviderCredential struct {
	APIKey  string    `json:"api_key"`
	AddedAt time.Time `json:"added_at"`
}

// StoredCredentials holds all stored credentials.
type StoredCredentials struct {
	Version     int                           `json:"version"`
	Credentials map[string]ProviderCredential `json:"credentials"`
}

var (
	credentialsMu    sync.RWMutex
	credentialsCache *StoredCredentials
)

// credentialsPath returns the path to the credentials file.
func credentialsPath() string {
	return filepath.Join(paths.ConfigDir(), "credentials.json")
}

// LoadStoredCredentials reads credentials from disk.
func LoadStoredCredentials() (*StoredCredentials, error) {
	credentialsMu.RLock()
	if credentialsCache != nil {
		defer credentialsMu.RUnlock()
		return credentialsCache, nil
	}
	credentialsMu.RUnlock()

	credentialsMu.Lock()
	defer credentialsMu.Unlock()

	// Double-check after acquiring write lock
	if credentialsCache != nil {
		return credentialsCache, nil
	}

	path := credentialsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			credentialsCache = &StoredCredentials{
				Version:     1,
				Credentials: make(map[string]ProviderCredential),
			}
			return credentialsCache, nil
		}
		return nil, err
	}

	var creds StoredCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	if creds.Credentials == nil {
		creds.Credentials = make(map[string]ProviderCredential)
	}

	credentialsCache = &creds
	return credentialsCache, nil
}

// SaveCredentials writes credentials to disk with secure permissions.
func SaveCredentials(creds *StoredCredentials) error {
	credentialsMu.Lock()
	defer credentialsMu.Unlock()

	path := credentialsPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	// Write with secure permissions (0600 = owner read/write only)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}

	credentialsCache = creds
	return nil
}

// StoreCredential adds a new credential for a provider.
// Only stores if the env var is not already set.
func StoreCredential(providerID, apiKey string) error {
	creds, err := LoadStoredCredentials()
	if err != nil {
		return err
	}

	creds.Credentials[providerID] = ProviderCredential{
		APIKey:  apiKey,
		AddedAt: time.Now(),
	}

	return SaveCredentials(creds)
}

// InjectCredentials loads stored credentials and sets them as environment variables.
// Does NOT overwrite existing environment variables.
// Call this early in main before any provider initialization.
func InjectCredentials() error {
	creds, err := LoadStoredCredentials()
	if err != nil {
		return err
	}

	for providerID, cred := range creds.Credentials {
		// Find the env var name for this provider
		var envVar string
		for _, p := range knownProviders {
			if p.ID == providerID {
				envVar = p.EnvVar
				break
			}
		}

		if envVar == "" {
			continue // Unknown provider, skip
		}

		// Only set if not already in environment
		if os.Getenv(envVar) == "" {
			os.Setenv(envVar, cred.APIKey)
		}
	}

	return nil
}

// ClearCredentialCache clears the in-memory credential cache.
// Useful for testing.
func ClearCredentialCache() {
	credentialsMu.Lock()
	defer credentialsMu.Unlock()
	credentialsCache = nil
}
