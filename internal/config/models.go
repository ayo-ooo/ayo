package config

import (
	"os"
	"strings"

	"github.com/charmbracelet/catwalk/pkg/embedded"
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
		return "gpt-4.1" // fallback when no credentials
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

	return "gpt-4.1" // ultimate fallback
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

	for _, id := range priority {
		for _, p := range providers {
			if p.ID == id {
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
