package guardrails

import "testing"

func TestLayerString(t *testing.T) {
	tests := []struct {
		layer Layer
		want  string
	}{
		{LayerInfrastructure, "infrastructure"},
		{LayerProtocol, "protocol"},
		{LayerPrompt, "prompt"},
		{LayerBehavioral, "behavioral"},
		{Layer(99), "unknown"},
	}

	for _, tt := range tests {
		got := tt.layer.String()
		if got != tt.want {
			t.Errorf("Layer(%d).String() = %q, want %q", tt.layer, got, tt.want)
		}
	}
}

func TestActiveLayers(t *testing.T) {
	tests := []struct {
		level     Level
		wantCount int
	}{
		{LevelMinimal, 2},
		{LevelStandard, 3},
		{LevelStrict, 4},
		{"", 3}, // Default to standard
	}

	for _, tt := range tests {
		layers := ActiveLayers(tt.level)
		if len(layers) != tt.wantCount {
			t.Errorf("ActiveLayers(%q) = %d layers, want %d", tt.level, len(layers), tt.wantCount)
		}
	}
}

func TestIsLayerActive(t *testing.T) {
	tests := []struct {
		level  Level
		layer  Layer
		active bool
	}{
		// Minimal: only L1+L2
		{LevelMinimal, LayerInfrastructure, true},
		{LevelMinimal, LayerProtocol, true},
		{LevelMinimal, LayerPrompt, false},
		{LevelMinimal, LayerBehavioral, false},
		// Standard: L1+L2+L3
		{LevelStandard, LayerInfrastructure, true},
		{LevelStandard, LayerProtocol, true},
		{LevelStandard, LayerPrompt, true},
		{LevelStandard, LayerBehavioral, false},
		// Strict: all layers
		{LevelStrict, LayerInfrastructure, true},
		{LevelStrict, LayerProtocol, true},
		{LevelStrict, LayerPrompt, true},
		{LevelStrict, LayerBehavioral, true},
	}

	for _, tt := range tests {
		got := IsLayerActive(tt.level, tt.layer)
		if got != tt.active {
			t.Errorf("IsLayerActive(%q, %s) = %v, want %v", tt.level, tt.layer, got, tt.active)
		}
	}
}

func TestConfigIsEnabled(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{"nil config", nil, true},
		{"nil enabled", &Config{}, true},
		{"enabled true", &Config{Enabled: boolPtr(true)}, true},
		{"enabled false", &Config{Enabled: boolPtr(false)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsEnabled()
			if got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigGetLevel(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   Level
	}{
		{"nil config", nil, LevelStandard},
		{"empty level", &Config{}, LevelStandard},
		{"minimal", &Config{Level: LevelMinimal}, LevelMinimal},
		{"strict", &Config{Level: LevelStrict}, LevelStrict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetLevel()
			if got != tt.want {
				t.Errorf("GetLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
