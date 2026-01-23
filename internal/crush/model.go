package crush

import (
	"strings"

	"github.com/alexcabrera/ayo/internal/config"
)

// ModelConfig represents model configuration for Crush execution.
type ModelConfig struct {
	// Model is the model identifier in Crush format (e.g., "openai/gpt-4" or just "gpt-4")
	Model string

	// SmallModel is the small model for auxiliary tasks
	SmallModel string
}

// ModelConfigFromAyo converts ayo's config to Crush-compatible model configuration.
// Crush accepts models in format "provider/model" or just "model" (searches all providers).
func ModelConfigFromAyo(cfg config.Config, modelOverride, smallModelOverride string) ModelConfig {
	result := ModelConfig{}

	// Determine the model to use
	model := modelOverride
	if model == "" {
		model = cfg.DefaultModel
	}

	if model != "" {
		result.Model = formatModelForCrush(cfg, model)
	}

	// Handle small model override
	if smallModelOverride != "" {
		result.SmallModel = formatModelForCrush(cfg, smallModelOverride)
	}

	return result
}

// formatModelForCrush formats a model identifier for Crush's --model flag.
// If the model already contains a provider prefix (e.g., "openai/gpt-4"), it's returned as-is.
// Otherwise, the provider ID from config is prepended if available.
func formatModelForCrush(cfg config.Config, model string) string {
	// If already has provider prefix, return as-is
	if strings.Contains(model, "/") {
		return model
	}

	// Get provider ID from config
	providerID := string(cfg.Provider.ID)
	if providerID == "" {
		// No provider configured, return model as-is (Crush will search all providers)
		return model
	}

	// Combine provider/model
	return providerID + "/" + model
}

// ParseModelString parses a model string that may be in "provider/model" or "model" format.
// Returns (provider, model). Provider is empty if not specified.
func ParseModelString(s string) (provider, model string) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", s
}

// ToRunOptions converts ModelConfig to RunOptions model fields.
func (mc ModelConfig) ToRunOptions() RunOptions {
	return RunOptions{
		Model:      mc.Model,
		SmallModel: mc.SmallModel,
	}
}

// ApplyToRunOptions applies the model configuration to existing RunOptions.
func (mc ModelConfig) ApplyToRunOptions(opts *RunOptions) {
	if mc.Model != "" {
		opts.Model = mc.Model
	}
	if mc.SmallModel != "" {
		opts.SmallModel = mc.SmallModel
	}
}
