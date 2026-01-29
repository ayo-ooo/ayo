package run

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/google"
	"charm.land/fantasy/providers/openai"
	"charm.land/fantasy/providers/openaicompat"
	"charm.land/fantasy/providers/openrouter"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// NewFantasyProvider creates a Fantasy provider from Catwalk configuration.
func NewFantasyProvider(p catwalk.Provider) (fantasy.Provider, error) {
	apiKey := getProviderAPIKey(p)

	switch p.Type {
	case catwalk.TypeOpenAI:
		opts := []openai.Option{
			openai.WithAPIKey(apiKey),
			openai.WithUseResponsesAPI(),
		}
		if p.APIEndpoint != "" {
			opts = append(opts, openai.WithBaseURL(p.APIEndpoint))
		}
		if len(p.DefaultHeaders) > 0 {
			opts = append(opts, openai.WithHeaders(p.DefaultHeaders))
		}
		return openai.New(opts...)

	case catwalk.TypeOpenAICompat:
		opts := []openaicompat.Option{
			openaicompat.WithAPIKey(apiKey),
		}
		if p.APIEndpoint != "" {
			opts = append(opts, openaicompat.WithBaseURL(p.APIEndpoint))
		}
		if len(p.DefaultHeaders) > 0 {
			opts = append(opts, openaicompat.WithHeaders(p.DefaultHeaders))
		}
		return openaicompat.New(opts...)

	case catwalk.TypeAnthropic:
		opts := []anthropic.Option{anthropic.WithAPIKey(apiKey)}
		if p.APIEndpoint != "" {
			opts = append(opts, anthropic.WithBaseURL(p.APIEndpoint))
		}
		return anthropic.New(opts...)

	case catwalk.TypeGoogle:
		opts := []google.Option{google.WithGeminiAPIKey(apiKey)}
		return google.New(opts...)

	case catwalk.TypeOpenRouter:
		opts := []openrouter.Option{openrouter.WithAPIKey(apiKey)}
		return openrouter.New(opts...)

	default:
		// For unknown types, try OpenAI-compatible
		if p.APIEndpoint != "" {
			opts := []openaicompat.Option{
				openaicompat.WithAPIKey(apiKey),
				openaicompat.WithBaseURL(p.APIEndpoint),
			}
			return openaicompat.New(opts...)
		}
		return nil, fmt.Errorf("unsupported provider type: %s", p.Type)
	}
}

// getProviderAPIKey retrieves the API key from config or environment.
func getProviderAPIKey(p catwalk.Provider) string {
	if p.APIKey != "" {
		return p.APIKey
	}
	envKey := strings.ToUpper(string(p.ID)) + "_API_KEY"
	return os.Getenv(envKey)
}

// NewLanguageModel creates a Fantasy language model from provider and model ID.
func NewLanguageModel(ctx context.Context, p catwalk.Provider, modelID string) (fantasy.LanguageModel, error) {
	provider, err := NewFantasyProvider(p)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}
	return provider.LanguageModel(ctx, modelID)
}
