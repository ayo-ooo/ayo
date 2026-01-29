package ollama

import (
	"context"
	"os/exec"
)

// Status represents the current state of Ollama.
type Status int

const (
	// StatusNotInstalled means the ollama binary is not in PATH.
	StatusNotInstalled Status = iota
	// StatusInstalled means the binary exists but server is not running.
	StatusInstalled
	// StatusRunning means the server is running and responding.
	StatusRunning
)

// String returns a human-readable status description.
func (s Status) String() string {
	switch s {
	case StatusNotInstalled:
		return "not installed"
	case StatusInstalled:
		return "installed (not running)"
	case StatusRunning:
		return "running"
	default:
		return "unknown"
	}
}

// IsBinaryInstalled checks if the ollama binary is in PATH.
func IsBinaryInstalled() bool {
	_, err := exec.LookPath("ollama")
	return err == nil
}

// GetStatus returns the current Ollama status.
func GetStatus(ctx context.Context) Status {
	if !IsBinaryInstalled() {
		return StatusNotInstalled
	}

	client := NewClient()
	if client.IsAvailable(ctx) {
		return StatusRunning
	}

	return StatusInstalled
}

// CapableModels is a list of known-good models for chat.
// These models are verified to work well for agentic tasks.
var CapableModels = []string{
	"mistral:7b",
	"llama3:8b",
	"llama3.2:3b",
	"llama3.1:8b",
	"mixtral:8x7b",
	"codellama:7b",
	"deepseek-coder:6.7b",
	"phi3:mini",
	"gemma2:9b",
	"qwen2.5:7b",
	"ministral:3b",
}

// RecommendedModel is the suggested model for first-time users.
const RecommendedModel = "mistral:7b"

// FastModel is a smaller model that's faster to download.
const FastModel = "ministral:3b"

// IsCapableForChat returns true if the model is known to work well for chat.
func IsCapableForChat(modelName string) bool {
	name := normalizeModelName(modelName)
	for _, m := range CapableModels {
		if normalizeModelName(m) == name {
			return true
		}
		// Also check base name without size suffix
		// e.g., "llama3:8b-instruct" should match "llama3:8b"
		if getBaseName(name) == getBaseName(m) {
			return true
		}
	}
	return false
}

// getBaseName extracts the base model name (before variant suffixes).
func getBaseName(name string) string {
	// Handle names like "llama3:8b-instruct" -> "llama3"
	// and "mistral:7b" -> "mistral"
	for i, c := range name {
		if c == ':' {
			return name[:i]
		}
	}
	return name
}

// ListCapableModels returns installed models that are suitable for chat.
// This is a convenience function that creates a new client.
func ListCapableModels(ctx context.Context) ([]Model, error) {
	client := NewClient()
	return client.ListCapableModels(ctx)
}

// ListCapableModels returns installed models that are suitable for chat.
func (c *Client) ListCapableModels(ctx context.Context) ([]Model, error) {
	models, err := c.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	var capable []Model
	for _, m := range models {
		if IsCapableForChat(m.Name) {
			capable = append(capable, m)
		}
	}
	return capable, nil
}

// ListAllModelsWithCapability returns all models with a capability flag.
type ModelWithCapability struct {
	Model
	Capable bool
}

// ListAllModelsWithCapability returns all installed models with capability info.
func (c *Client) ListAllModelsWithCapability(ctx context.Context) ([]ModelWithCapability, error) {
	models, err := c.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ModelWithCapability, len(models))
	for i, m := range models {
		result[i] = ModelWithCapability{
			Model:   m,
			Capable: IsCapableForChat(m.Name),
		}
	}
	return result, nil
}

// SuggestedModels returns models that can be installed for first-time users.
type SuggestedModel struct {
	Name        string
	Description string
	SizeGB      float64
}

// GetSuggestedModels returns models recommended for installation.
func GetSuggestedModels() []SuggestedModel {
	return []SuggestedModel{
		{Name: "mistral:7b", Description: "Recommended - good balance of speed and quality", SizeGB: 4.1},
		{Name: "llama3.2:3b", Description: "Fast and lightweight", SizeGB: 2.0},
		{Name: "llama3:8b", Description: "High quality responses", SizeGB: 4.7},
		{Name: "phi3:mini", Description: "Microsoft's efficient model", SizeGB: 2.3},
		{Name: "ministral:3b", Description: "Smallest capable model", SizeGB: 1.8},
	}
}
