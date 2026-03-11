package types

import (
	"testing"
)

// TestConfig_Validate tests the Config.Validate method.
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Agent: AgentConfig{
					Name:        "test-agent",
					Description: "Test agent",
					Model:       "gpt-4",
				},
				CLI: CLIConfig{
					Mode:        "freeform",
					Description: "Test CLI",
				},
			},
			wantErr: false,
		},
		{
			name: "missing agent name",
			config: Config{
				Agent: AgentConfig{
					Name:        "",
					Description: "Test agent",
					Model:       "gpt-4",
				},
				CLI: CLIConfig{
					Mode:        "freeform",
					Description: "Test CLI",
				},
			},
			wantErr: true,
		},
		{
			name: "missing agent description",
			config: Config{
				Agent: AgentConfig{
					Name:        "test-agent",
					Description: "",
					Model:       "gpt-4",
				},
				CLI: CLIConfig{
					Mode:        "freeform",
					Description: "Test CLI",
				},
			},
			wantErr: true,
		},
		{
			name: "missing model",
			config: Config{
				Agent: AgentConfig{
					Name:        "test-agent",
					Description: "Test agent",
					Model:       "",
				},
				CLI: CLIConfig{
					Mode:        "freeform",
					Description: "Test CLI",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid CLI mode",
			config: Config{
				Agent: AgentConfig{
					Name:        "test-agent",
					Description: "Test agent",
					Model:       "gpt-4",
				},
				CLI: CLIConfig{
					Mode:        "invalid",
					Description: "Test CLI",
				},
			},
			wantErr: true,
		},
		{
			name: "temperature out of range",
			config: Config{
				Agent: AgentConfig{
					Name:        "test-agent",
					Description: "Test agent",
					Model:       "gpt-4",
					Temperature: func() *float64 { v := 3.0; return &v }(),
				},
				CLI: CLIConfig{
					Mode:        "freeform",
					Description: "Test CLI",
				},
			},
			wantErr: true,
		},
		{
			name: "valid temperature",
			config: Config{
				Agent: AgentConfig{
					Name:        "test-agent",
					Description: "Test agent",
					Model:       "gpt-4",
					Temperature: func() *float64 { v := 0.7; return &v }(),
				},
				CLI: CLIConfig{
					Mode:        "freeform",
					Description: "Test CLI",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAgentConfig_Validate tests the AgentConfig.Validate method.
func TestAgentConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  AgentConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: AgentConfig{
				Name:        "test",
				Description: "desc",
				Model:       "gpt-4",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: AgentConfig{
				Name:        "",
				Description: "desc",
				Model:       "gpt-4",
			},
			wantErr: true,
		},
		{
			name: "memory enabled with invalid scope",
			config: AgentConfig{
				Name:        "test",
				Description: "desc",
				Model:       "gpt-4",
				Memory: AgentMemory{
					Enabled: true,
					Scope:   "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "memory enabled with valid scope",
			config: AgentConfig{
				Name:        "test",
				Description: "desc",
				Model:       "gpt-4",
				Memory: AgentMemory{
					Enabled: true,
					Scope:   "agent",
				},
			},
			wantErr: false,
		},
		{
			name: "negative max tokens",
			config: AgentConfig{
				Name:        "test",
				Description: "desc",
				Model:       "gpt-4",
				MaxTokens:   -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AgentConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCLIConfig_Validate tests the CLIConfig.Validate method.
func TestCLIConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  CLIConfig
		wantErr bool
	}{
		{
			name: "valid freeform",
			config: CLIConfig{
				Mode:        "freeform",
				Description: "desc",
			},
			wantErr: false,
		},
		{
			name: "valid structured",
			config: CLIConfig{
				Mode:        "structured",
				Description: "desc",
			},
			wantErr: false,
		},
		{
			name: "valid hybrid",
			config: CLIConfig{
				Mode:        "hybrid",
				Description: "desc",
			},
			wantErr: false,
		},
		{
			name: "missing mode",
			config: CLIConfig{
				Mode:        "",
				Description: "desc",
			},
			wantErr: true,
		},
		{
			name: "missing description",
			config: CLIConfig{
				Mode:        "freeform",
				Description: "",
			},
			wantErr: true,
		},
		{
			name: "invalid flag type",
			config: CLIConfig{
				Mode:        "structured",
				Description: "desc",
				Flags: map[string]CLIFlag{
					"test": {
						Name:        "test",
						Type:        "invalid",
						Description: "test flag",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "long short flag",
			config: CLIConfig{
				Mode:        "structured",
				Description: "desc",
				Flags: map[string]CLIFlag{
					"test": {
						Name:        "test",
						Type:        "string",
						Short:       "ab",
						Description: "test flag",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative position",
			config: CLIConfig{
				Mode:        "structured",
				Description: "desc",
				Flags: map[string]CLIFlag{
					"test": {
						Name:        "test",
						Type:        "string",
						Position:    -1,
						Description: "test flag",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CLIConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIsValidModel tests the IsValidModel function.
func TestIsValidModel(t *testing.T) {
	tests := []struct {
		model string
		valid bool
	}{
		{"gpt-4", true},
		{"gpt-3.5-turbo", true},
		{"claude-3-opus", true},
		{"claude-3-sonnet", true},
		{"o1-preview", true},
		{"o1-mini", true},
		{"gemini-pro", true},
		{"invalid-model", false},
		{"test", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := IsValidModel(tt.model)
			if result != tt.valid {
				t.Errorf("IsValidModel(%q) = %v, want %v", tt.model, result, tt.valid)
			}
		})
	}
}
