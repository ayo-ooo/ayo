package config

import (
	"os"
	"strings"

	"charm.land/catwalk/pkg/embedded"
)

type ModelChoice struct {
	ID       string
	Name     string
	Provider string // Provider ID (openai, anthropic, etc.)
}

func ConfiguredModels(cfg Config) []ModelChoice {
	provider := cfg.Provider
	models := provider.Models

	if len(models) == 0 {
		for _, p := range embedded.GetAll() {
			if p.ID == provider.ID {
				models = p.Models
				break
			}
		}
	}

	envKey := strings.ToUpper(string(provider.ID)) + "_API_KEY"
	if envKey != "" && os.Getenv(envKey) == "" {
		return nil
	}

	choices := make([]ModelChoice, 0, len(models))
	for _, m := range models {
		choices = append(choices, ModelChoice{ID: m.ID, Name: m.Name, Provider: string(provider.ID)})
	}
	return choices
}

// AllConfiguredModels returns models from all providers that have credentials.
// This is used for model selection in setup.
func AllConfiguredModels() []ModelChoice {
	var choices []ModelChoice

	configuredProviders := GetProvidersWithCredentials()
	providerIDs := make(map[string]bool)
	for _, p := range configuredProviders {
		providerIDs[p.ID] = true
	}

	for _, provider := range embedded.GetAll() {
		if !providerIDs[string(provider.ID)] {
			continue
		}

		for _, m := range provider.Models {
			choices = append(choices, ModelChoice{
				ID:       m.ID,
				Name:     m.Name,
				Provider: string(provider.ID),
			})
		}
	}

	return choices
}

// GetProviderDefaultModel returns the default model ID for a provider.
// Uses catwalk's embedded provider data.
func GetProviderDefaultModel(providerID string) string {
	for _, p := range embedded.GetAll() {
		if string(p.ID) == providerID {
			return p.DefaultLargeModelID
		}
	}
	return ""
}

// GetProviderDefaultSmallModel returns the default small model ID for a provider.
// Uses catwalk's embedded provider data.
func GetProviderDefaultSmallModel(providerID string) string {
	for _, p := range embedded.GetAll() {
		if string(p.ID) == providerID {
			return p.DefaultSmallModelID
		}
	}
	return ""
}

// GetDefaultModelForConfiguredProvider returns the best default model
// based on which providers have credentials.
// Priority: openai > anthropic > google > openrouter > first available
func GetDefaultModelForConfiguredProvider() string {
	providers := GetProvidersWithCredentials()
	if len(providers) == 0 {
		return "gpt-5.2" // fallback when no credentials
	}

	// Priority order
	priority := []string{"openai", "anthropic", "google", "openrouter"}

	for _, id := range priority {
		for _, p := range providers {
			if p.ID == id {
				if defaultModel := GetProviderDefaultModel(id); defaultModel != "" {
					return defaultModel
				}
			}
		}
	}

	// Fall back to first provider's default
	if defaultModel := GetProviderDefaultModel(providers[0].ID); defaultModel != "" {
		return defaultModel
	}

	return "gpt-5.2" // ultimate fallback
}

// GetDefaultSmallModelForConfiguredProvider returns the best default small model
// based on which providers have credentials. Used for fast/cheap tasks.
// Falls back to Ollama if no cloud provider is configured.
func GetDefaultSmallModelForConfiguredProvider() string {
	providers := GetProvidersWithCredentials()
	if len(providers) == 0 {
		return "ollama/ministral-3:3b" // fallback to local
	}

	// Priority order - prefer providers with good small models
	priority := []string{"openai", "anthropic", "google", "groq", "openrouter"}

	// Override map for providers where we want a different small model than catwalk default
	overrides := map[string]string{
		"openai": "gpt-5.2", // Use gpt-5.2 instead of catwalk's default
	}

	for _, id := range priority {
		for _, p := range providers {
			if p.ID == id {
				// Check for override first
				if override, ok := overrides[id]; ok {
					return override
				}
				if defaultModel := GetProviderDefaultSmallModel(id); defaultModel != "" {
					return defaultModel
				}
			}
		}
	}

	// Fall back to first provider's small model
	if defaultModel := GetProviderDefaultSmallModel(providers[0].ID); defaultModel != "" {
		return defaultModel
	}

	return "ollama/ministral-3:3b" // ultimate fallback to local
}

// ResolveModelProvider parses a model string and returns provider and model ID.
// Supports formats: "model-name" or "provider/model-name".
// If no provider prefix is found, returns empty provider.
func ResolveModelProvider(modelStr string) (provider, model string) {
	if modelStr == "" {
		return "", ""
	}
	if idx := strings.Index(modelStr, "/"); idx > 0 {
		return modelStr[:idx], modelStr[idx+1:]
	}
	return "", modelStr
}

// GetLargeModel returns the configured large model from the config.
// Resolution order:
// 1. Config.Models[ModelTypeLarge] if set
// 2. Config.DefaultModel (legacy) converted to SelectedModel
// 3. Default from provider with credentials
func (c Config) GetLargeModel() SelectedModel {
	// Check typed Models map first
	if c.Models != nil {
		if m, ok := c.Models[ModelTypeLarge]; ok && !m.IsEmpty() {
			return m
		}
	}

	// Fall back to legacy DefaultModel field
	if c.DefaultModel != "" {
		provider, model := ResolveModelProvider(c.DefaultModel)
		return SelectedModel{
			Model:    model,
			Provider: provider,
		}
	}

	// Fall back to provider detection
	modelStr := GetDefaultModelForConfiguredProvider()
	provider, model := ResolveModelProvider(modelStr)
	return SelectedModel{
		Model:    model,
		Provider: provider,
	}
}

// GetSmallModel returns the configured small model from the config.
// Resolution order:
// 1. Config.Models[ModelTypeSmall] if set
// 2. Config.SmallModel (legacy) converted to SelectedModel
// 3. Default from provider with credentials (or Ollama fallback)
func (c Config) GetSmallModel() SelectedModel {
	// Check typed Models map first
	if c.Models != nil {
		if m, ok := c.Models[ModelTypeSmall]; ok && !m.IsEmpty() {
			return m
		}
	}

	// Fall back to legacy SmallModel field
	if c.SmallModel != "" {
		provider, model := ResolveModelProvider(c.SmallModel)
		return SelectedModel{
			Model:    model,
			Provider: provider,
		}
	}

	// Fall back to provider detection (prefers Ollama for local)
	modelStr := GetDefaultSmallModelForConfiguredProvider()
	provider, model := ResolveModelProvider(modelStr)
	return SelectedModel{
		Model:    model,
		Provider: provider,
	}
}
