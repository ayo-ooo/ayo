package crush

import (
	"testing"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

func TestFormatModelForCrush(t *testing.T) {
	tests := []struct {
		name       string
		cfg        config.Config
		model      string
		wantResult string
	}{
		{
			name: "model with provider prefix unchanged",
			cfg: config.Config{
				Provider: catwalk.Provider{ID: catwalk.InferenceProviderOpenAI},
			},
			model:      "anthropic/claude-3.5-sonnet",
			wantResult: "anthropic/claude-3.5-sonnet",
		},
		{
			name: "model without prefix gets provider added",
			cfg: config.Config{
				Provider: catwalk.Provider{ID: catwalk.InferenceProviderOpenAI},
			},
			model:      "gpt-4",
			wantResult: "openai/gpt-4",
		},
		{
			name: "anthropic provider",
			cfg: config.Config{
				Provider: catwalk.Provider{ID: catwalk.InferenceProviderAnthropic},
			},
			model:      "claude-3.5-sonnet",
			wantResult: "anthropic/claude-3.5-sonnet",
		},
		{
			name: "openrouter provider",
			cfg: config.Config{
				Provider: catwalk.Provider{ID: catwalk.InferenceProviderOpenRouter},
			},
			model:      "claude-3.5-sonnet",
			wantResult: "openrouter/claude-3.5-sonnet",
		},
		{
			name:       "no provider configured",
			cfg:        config.Config{},
			model:      "gpt-4",
			wantResult: "gpt-4",
		},
		{
			name: "empty model",
			cfg: config.Config{
				Provider: catwalk.Provider{ID: catwalk.InferenceProviderOpenAI},
			},
			model:      "",
			wantResult: "openai/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatModelForCrush(tt.cfg, tt.model)
			if got != tt.wantResult {
				t.Errorf("formatModelForCrush() = %q, want %q", got, tt.wantResult)
			}
		})
	}
}

func TestModelConfigFromAyo(t *testing.T) {
	tests := []struct {
		name               string
		cfg                config.Config
		modelOverride      string
		smallModelOverride string
		wantModel          string
		wantSmallModel     string
	}{
		{
			name: "uses default model when no override",
			cfg: config.Config{
				DefaultModel: "gpt-4.1",
				Provider:     catwalk.Provider{ID: catwalk.InferenceProviderOpenAI},
			},
			modelOverride:  "",
			wantModel:      "openai/gpt-4.1",
			wantSmallModel: "",
		},
		{
			name: "override takes precedence",
			cfg: config.Config{
				DefaultModel: "gpt-4.1",
				Provider:     catwalk.Provider{ID: catwalk.InferenceProviderOpenAI},
			},
			modelOverride:  "claude-3.5-sonnet",
			wantModel:      "openai/claude-3.5-sonnet",
			wantSmallModel: "",
		},
		{
			name: "override with provider prefix",
			cfg: config.Config{
				DefaultModel: "gpt-4.1",
				Provider:     catwalk.Provider{ID: catwalk.InferenceProviderOpenAI},
			},
			modelOverride:  "anthropic/claude-3.5-sonnet",
			wantModel:      "anthropic/claude-3.5-sonnet",
			wantSmallModel: "",
		},
		{
			name: "small model override",
			cfg: config.Config{
				DefaultModel: "gpt-4.1",
				Provider:     catwalk.Provider{ID: catwalk.InferenceProviderOpenAI},
			},
			modelOverride:      "",
			smallModelOverride: "gpt-4o-mini",
			wantModel:          "openai/gpt-4.1",
			wantSmallModel:     "openai/gpt-4o-mini",
		},
		{
			name: "both overrides",
			cfg: config.Config{
				DefaultModel: "gpt-4.1",
				Provider:     catwalk.Provider{ID: catwalk.InferenceProviderAnthropic},
			},
			modelOverride:      "claude-3.5-sonnet",
			smallModelOverride: "claude-3-haiku",
			wantModel:          "anthropic/claude-3.5-sonnet",
			wantSmallModel:     "anthropic/claude-3-haiku",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ModelConfigFromAyo(tt.cfg, tt.modelOverride, tt.smallModelOverride)
			if got.Model != tt.wantModel {
				t.Errorf("Model = %q, want %q", got.Model, tt.wantModel)
			}
			if got.SmallModel != tt.wantSmallModel {
				t.Errorf("SmallModel = %q, want %q", got.SmallModel, tt.wantSmallModel)
			}
		})
	}
}

func TestParseModelString(t *testing.T) {
	tests := []struct {
		input        string
		wantProvider string
		wantModel    string
	}{
		{
			input:        "openai/gpt-4",
			wantProvider: "openai",
			wantModel:    "gpt-4",
		},
		{
			input:        "anthropic/claude-3.5-sonnet",
			wantProvider: "anthropic",
			wantModel:    "claude-3.5-sonnet",
		},
		{
			input:        "gpt-4",
			wantProvider: "",
			wantModel:    "gpt-4",
		},
		{
			input:        "",
			wantProvider: "",
			wantModel:    "",
		},
		{
			input:        "openrouter/anthropic/claude-3.5-sonnet",
			wantProvider: "openrouter",
			wantModel:    "anthropic/claude-3.5-sonnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			provider, model := ParseModelString(tt.input)
			if provider != tt.wantProvider {
				t.Errorf("provider = %q, want %q", provider, tt.wantProvider)
			}
			if model != tt.wantModel {
				t.Errorf("model = %q, want %q", model, tt.wantModel)
			}
		})
	}
}

func TestModelConfig_ToRunOptions(t *testing.T) {
	mc := ModelConfig{
		Model:      "openai/gpt-4",
		SmallModel: "openai/gpt-4o-mini",
	}

	opts := mc.ToRunOptions()

	if opts.Model != "openai/gpt-4" {
		t.Errorf("Model = %q, want %q", opts.Model, "openai/gpt-4")
	}
	if opts.SmallModel != "openai/gpt-4o-mini" {
		t.Errorf("SmallModel = %q, want %q", opts.SmallModel, "openai/gpt-4o-mini")
	}
}

func TestModelConfig_ApplyToRunOptions(t *testing.T) {
	mc := ModelConfig{
		Model:      "openai/gpt-4",
		SmallModel: "openai/gpt-4o-mini",
	}

	opts := RunOptions{
		WorkingDir: "/some/path",
		Quiet:      true,
	}

	mc.ApplyToRunOptions(&opts)

	if opts.Model != "openai/gpt-4" {
		t.Errorf("Model = %q, want %q", opts.Model, "openai/gpt-4")
	}
	if opts.SmallModel != "openai/gpt-4o-mini" {
		t.Errorf("SmallModel = %q, want %q", opts.SmallModel, "openai/gpt-4o-mini")
	}
	// Original fields preserved
	if opts.WorkingDir != "/some/path" {
		t.Errorf("WorkingDir = %q, want %q", opts.WorkingDir, "/some/path")
	}
	if !opts.Quiet {
		t.Error("Quiet should be true")
	}
}

func TestModelConfig_ApplyToRunOptions_Empty(t *testing.T) {
	mc := ModelConfig{} // Empty config

	opts := RunOptions{
		Model:      "existing/model",
		SmallModel: "existing/small",
	}

	mc.ApplyToRunOptions(&opts)

	// Empty model config should not overwrite existing values
	if opts.Model != "existing/model" {
		t.Errorf("Model = %q, want %q", opts.Model, "existing/model")
	}
	if opts.SmallModel != "existing/small" {
		t.Errorf("SmallModel = %q, want %q", opts.SmallModel, "existing/small")
	}
}
