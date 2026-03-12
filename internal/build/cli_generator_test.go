package build

import (
	"testing"

	"github.com/alexcabrera/ayo/internal/build/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetStringDefault tests the getStringDefault helper function
func TestGetStringDefault(t *testing.T) {
	tests := []struct {
		name     string
		def      any
		expected string
	}{
		{
			name:     "nil returns empty string",
			def:      nil,
			expected: "",
		},
		{
			name:     "string value returned",
			def:      "hello",
			expected: "hello",
		},
		{
			name:     "empty string",
			def:      "",
			expected: "",
		},
		{
			name:     "int returns empty string",
			def:      42,
			expected: "",
		},
		{
			name:     "bool returns empty string",
			def:      true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringDefault(tt.def)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetIntDefault tests the getIntDefault helper function
func TestGetIntDefault(t *testing.T) {
	tests := []struct {
		name     string
		def      any
		expected int
	}{
		{
			name:     "nil returns 0",
			def:      nil,
			expected: 0,
		},
		{
			name:     "int value returned",
			def:      42,
			expected: 42,
		},
		{
			name:     "float64 converted to int",
			def:      3.14,
			expected: 3,
		},
		{
			name:     "string returns 0",
			def:      "42",
			expected: 0,
		},
		{
			name:     "bool returns 0",
			def:      true,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIntDefault(tt.def)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetFloatDefault tests the getFloatDefault helper function
func TestGetFloatDefault(t *testing.T) {
	tests := []struct {
		name     string
		def      any
		expected float64
	}{
		{
			name:     "nil returns 0",
			def:      nil,
			expected: 0,
		},
		{
			name:     "float64 value returned",
			def:      3.14,
			expected: 3.14,
		},
		{
			name:     "int converted to float64",
			def:      42,
			expected: 42.0,
		},
		{
			name:     "string returns 0",
			def:      "3.14",
			expected: 0,
		},
		{
			name:     "bool returns 0",
			def:      true,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFloatDefault(tt.def)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetBoolDefault tests the getBoolDefault helper function
func TestGetBoolDefault(t *testing.T) {
	tests := []struct {
		name     string
		def      any
		expected bool
	}{
		{
			name:     "nil returns false",
			def:      nil,
			expected: false,
		},
		{
			name:     "true returned",
			def:      true,
			expected: true,
		},
		{
			name:     "false returned",
			def:      false,
			expected: false,
		},
		{
			name:     "int returns false",
			def:      1,
			expected: false,
		},
		{
			name:     "string returns false",
			def:      "true",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBoolDefault(tt.def)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetArrayDefault tests the getArrayDefault helper function
func TestGetArrayDefault(t *testing.T) {
	tests := []struct {
		name     string
		def      any
		expected []string
	}{
		{
			name:     "nil returns nil",
			def:      nil,
			expected: nil,
		},
		{
			name:     "[]string returned",
			def:      []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty []string",
			def:      []string{},
			expected: []string{},
		},
		{
			name:     "[]interface{} with strings converted",
			def:      []interface{}{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "[]interface{} with mixed types",
			def:      []interface{}{"a", 1, true},
			expected: []string{"a", "", ""},
		},
		{
			name:     "int returns nil",
			def:      42,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getArrayDefault(tt.def)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGenerateCLIBasic tests basic CLI generation
func TestGenerateCLIBasic(t *testing.T) {
	config := &types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
		CLI: types.CLIConfig{
			Mode: "freeform",
		},
	}

	cmd, err := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, cmd)
	// Use string includes "[prompt...]" for freeform mode
	assert.Contains(t, cmd.Use, "test-agent")
	assert.Equal(t, "Test agent", cmd.Short)
}

// TestGenerateCLIFlags tests CLI generation with flags
func TestGenerateCLIFlags(t *testing.T) {
	config := &types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
		CLI: types.CLIConfig{
			Mode: "freeform",
			Flags: map[string]types.CLIFlag{
				"string-flag": {
					Name:        "string-flag",
					Type:        "string",
					Description: "String flag",
					Default:     "default-value",
				},
				"int-flag": {
					Name:    "int-flag",
					Type:    "int",
					Default: 42,
				},
				"float-flag": {
					Name:    "float-flag",
					Type:    "float",
					Default: 3.14,
				},
				"bool-flag": {
					Name:    "bool-flag",
					Type:    "bool",
					Default: true,
				},
				"array-flag": {
					Name:    "array-flag",
					Type:    "array",
					Default: []string{"a", "b"},
				},
			},
		},
	}

	cmd, err := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, cmd)

	// Verify flags are defined
	assert.True(t, cmd.Flags().Lookup("string-flag") != nil)
	assert.True(t, cmd.Flags().Lookup("int-flag") != nil)
	assert.True(t, cmd.Flags().Lookup("float-flag") != nil)
	assert.True(t, cmd.Flags().Lookup("bool-flag") != nil)
	assert.True(t, cmd.Flags().Lookup("array-flag") != nil)
}

// TestGenerateCLIPositionalArgs tests CLI with positional arguments
func TestGenerateCLIPositionalArgs(t *testing.T) {
	config := &types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
		CLI: types.CLIConfig{
			Mode: "structured",
			Flags: map[string]types.CLIFlag{
				"input": {
					Name:     "input",
					Type:     "string",
					Position: 0,
				},
				"output": {
					Name:     "output",
					Type:     "string",
					Position: 1,
				},
			},
		},
	}

	cmd, err := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, cmd)

	// Check that Use string includes positional args
	assert.Contains(t, cmd.Use, "<input>")
	assert.Contains(t, cmd.Use, "<output>")
}

// TestGenerateCLIRequiredFlags tests CLI with required flags
func TestGenerateCLIRequiredFlags(t *testing.T) {
	config := &types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
		CLI: types.CLIConfig{
			Mode: "freeform",
			Flags: map[string]types.CLIFlag{
				"required-flag": {
					Name:     "required-flag",
					Type:     "string",
					Required: true,
				},
			},
		},
	}

	cmd, err := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, cmd)

	// Note: We can't easily test that MarkFlagRequired was called
	// but we can verify the flag exists
	assert.True(t, cmd.Flags().Lookup("required-flag") != nil)
}

// TestGenerateCLIInvalidFlagType tests error handling for invalid flag types
func TestGenerateCLIInvalidFlagType(t *testing.T) {
	config := &types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
		CLI: types.CLIConfig{
			Mode: "freeform",
			Flags: map[string]types.CLIFlag{
				"invalid-flag": {
					Name: "invalid-flag",
					Type: "invalid-type",
				},
			},
		},
	}

	cmd, err := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })
	assert.Error(t, err)
	assert.Nil(t, cmd)
	assert.Contains(t, err.Error(), "unsupported flag type")
}

// TestGenerateCLIModeStructured tests structured mode
func TestGenerateCLIModeStructured(t *testing.T) {
	config := &types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
		CLI: types.CLIConfig{
			Mode: "structured",
			Flags: map[string]types.CLIFlag{
				"input": {
					Name:     "input",
					Type:     "string",
					Position: 0,
				},
			},
		},
	}

	cmd, err := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "<input>")
}

// TestGenerateCLIModeFreeform tests freeform mode
func TestGenerateCLIModeFreeform(t *testing.T) {
	config := &types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
		CLI: types.CLIConfig{
			Mode: "freeform",
		},
	}

	cmd, err := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "[prompt...]")
}

// TestGenerateCLIModeHybrid tests hybrid mode
func TestGenerateCLIModeHybrid(t *testing.T) {
	config := &types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
		CLI: types.CLIConfig{
			Mode: "hybrid",
			Flags: map[string]types.CLIFlag{
				"input": {
					Name:     "input",
					Type:     "string",
					Position: 0,
				},
			},
		},
	}

	cmd, err := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "<input>")
	assert.Contains(t, cmd.Use, "[prompt...]")
}

// TestGenerateCLIDefaultMode tests that hybrid is the default mode
func TestGenerateCLIDefaultMode(t *testing.T) {
	config := &types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
		CLI: types.CLIConfig{
			// No mode specified, should default to hybrid
		},
	}

	cmd, err := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })
	require.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "[prompt...]")
}

// TestParseArgsPositionalString tests parsing positional string arguments
func TestParseArgsPositionalString(t *testing.T) {
	config := &types.Config{
		CLI: types.CLIConfig{
			Mode: "structured",
			Flags: map[string]types.CLIFlag{
				"input": {
					Name:     "input",
					Type:     "string",
					Position: 0,
				},
			},
		},
	}

	cmd, _ := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })

	result, _, err := ParseArgs(cmd, []string{"test-input"}, config)
	require.NoError(t, err)
	assert.Equal(t, "test-input", result["input"])
}

// TestParseArgsPositionalInt tests parsing positional int arguments
func TestParseArgsPositionalInt(t *testing.T) {
	config := &types.Config{
		CLI: types.CLIConfig{
			Mode: "structured",
			Flags: map[string]types.CLIFlag{
				"count": {
					Name:     "count",
					Type:     "int",
					Position: 0,
				},
			},
		},
	}

	cmd, _ := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })

	result, _, err := ParseArgs(cmd, []string{"42"}, config)
	require.NoError(t, err)
	assert.Equal(t, 42, result["count"])
}

// TestParseArgsInvalidPositionalInt tests error handling for invalid int
func TestParseArgsInvalidPositionalInt(t *testing.T) {
	config := &types.Config{
		CLI: types.CLIConfig{
			Mode: "structured",
			Flags: map[string]types.CLIFlag{
				"count": {
					Name:     "count",
					Type:     "int",
					Position: 0,
				},
			},
		},
	}

	cmd, _ := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })

	_, _, err := ParseArgs(cmd, []string{"not-a-number"}, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid integer")
}

// TestParseArgsInvalidPositionalFloat tests error handling for invalid float
func TestParseArgsInvalidPositionalFloat(t *testing.T) {
	config := &types.Config{
		CLI: types.CLIConfig{
			Mode: "structured",
			Flags: map[string]types.CLIFlag{
				"ratio": {
					Name:     "ratio",
					Type:     "float",
					Position: 0,
				},
			},
		},
	}

	cmd, _ := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })

	_, _, err := ParseArgs(cmd, []string{"not-a-float"}, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid float")
}

// TestParseArgsPositionalBool tests parsing positional bool arguments
func TestParseArgsPositionalBool(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true value", "true", true},
		{"TRUE value", "TRUE", true},
		{"1 value", "1", true},
		{"false value", "false", false},
		{"FALSE value", "FALSE", false},
		{"0 value", "0", false},
		{"other value", "yes", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.Config{
				CLI: types.CLIConfig{
					Mode: "structured",
					Flags: map[string]types.CLIFlag{
						"flag": {
							Name:     "flag",
							Type:     "bool",
							Position: 0,
						},
					},
				},
			}

			cmd, _ := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })

			result, _, err := ParseArgs(cmd, []string{tt.input}, config)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result["flag"])
		})
	}
}

// TestParseArgsPositionalArray tests parsing positional array arguments
func TestParseArgsPositionalArray(t *testing.T) {
	config := &types.Config{
		CLI: types.CLIConfig{
			Mode: "structured",
			Flags: map[string]types.CLIFlag{
				"items": {
					Name:     "items",
					Type:     "array",
					Position: 0,
				},
			},
		},
	}

	cmd, _ := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })

	result, _, err := ParseArgs(cmd, []string{"a,b,c"}, config)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, result["items"])
}

// TestParseArgsFreeformArgs tests freeform arguments in freeform mode
func TestParseArgsFreeformArgs(t *testing.T) {
	config := &types.Config{
		CLI: types.CLIConfig{
			Mode: "freeform",
		},
	}

	cmd, _ := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })

	_, positionalArgs, err := ParseArgs(cmd, []string{"arg1", "arg2", "arg3"}, config)
	require.NoError(t, err)
	assert.Equal(t, []string{"arg1", "arg2", "arg3"}, positionalArgs)
}

// TestParseArgsHybridArgs tests hybrid mode with positional and freeform args
func TestParseArgsHybridArgs(t *testing.T) {
	config := &types.Config{
		CLI: types.CLIConfig{
			Mode: "hybrid",
			Flags: map[string]types.CLIFlag{
				"input": {
					Name:     "input",
					Type:     "string",
					Position: 0,
				},
			},
		},
	}

	cmd, _ := GenerateCLI(config, func(cmd *cobra.Command, args []string) error { return nil })

	result, positionalArgs, err := ParseArgs(cmd, []string{"test-input", "arg1", "arg2"}, config)
	require.NoError(t, err)
	assert.Equal(t, "test-input", result["input"])
	assert.Equal(t, []string{"arg1", "arg2"}, positionalArgs)
}
