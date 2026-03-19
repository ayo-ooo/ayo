package generate

import (
	"strings"
	"testing"

	"github.com/charmbracelet/ayo/internal/project"
)

func TestGenerateAgent_APIKeyEnvMapping(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
	}

	code, err := GenerateAgent(proj, "main")
	if err != nil {
		t.Fatalf("GenerateAgent() error = %v", err)
	}

	providerEnvMap := map[string]string{
		"anthropic":  "ANTHROPIC_API_KEY",
		"openai":     "OPENAI_API_KEY",
		"gemini":     "GEMINI_API_KEY",
		"groq":       "GROQ_API_KEY",
		"openrouter": "OPENROUTER_API_KEY",
		"zai":        "ZAI_API_KEY",
	}

	for provider, expectedEnv := range providerEnvMap {
		// Check the case statement for each provider
		if !strings.Contains(code, `case "`+provider+`":`) {
			t.Errorf("Missing case for provider %q in getAPIKeyEnv", provider)
		}
		if !strings.Contains(code, `return "`+expectedEnv+`"`) {
			t.Errorf("Missing return for %q -> %q mapping", provider, expectedEnv)
		}
	}
}

func TestGenerateAgent_SelectModelStub(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
	}

	code, err := GenerateAgent(proj, "main")
	if err != nil {
		t.Fatalf("GenerateAgent() error = %v", err)
	}

	if !strings.Contains(code, "func selectModel() error {") {
		t.Error("Missing selectModel function")
	}

	if !strings.Contains(code, "runModelSelectionTUI()") {
		t.Error("selectModel should launch the TUI")
	}
}

func TestGenerateAgent_EnsureConfigLogic(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
	}

	code, err := GenerateAgent(proj, "main")
	if err != nil {
		t.Fatalf("GenerateAgent() error = %v", err)
	}

	if !strings.Contains(code, "func ensureConfig() error {") {
		t.Error("Missing ensureConfig function")
	}

	if !strings.Contains(code, "cfg != nil && cfg.Provider != \"\" && cfg.Model != \"\"") {
		t.Error("Missing check for existing valid config")
	}

	if !strings.Contains(code, "provider != \"\" && model != \"\"") {
		t.Error("Missing check for flag overrides")
	}

	if !strings.Contains(code, "selectModel()") {
		t.Error("Missing call to selectModel when no config/flags")
	}
}

func TestGenerateAgent_LoadConfig(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
	}

	code, err := GenerateAgent(proj, "main")
	if err != nil {
		t.Fatalf("GenerateAgent() error = %v", err)
	}

	if !strings.Contains(code, "func loadConfig() (*AgentConfig, error) {") {
		t.Error("Missing loadConfig function")
	}

	if !strings.Contains(code, "os.UserConfigDir()") {
		t.Error("loadConfig should use UserConfigDir")
	}

	if !strings.Contains(code, `filepath.Join(configDir, "agents"`) {
		t.Error("loadConfig should create path in agents directory")
	}

	if !strings.Contains(code, "toml.Decode") {
		t.Error("loadConfig should decode TOML")
	}
}

func TestGenerateAgent_SaveConfig(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
	}

	code, err := GenerateAgent(proj, "main")
	if err != nil {
		t.Fatalf("GenerateAgent() error = %v", err)
	}

	if !strings.Contains(code, "func saveConfig(cfg *AgentConfig) error {") {
		t.Error("Missing saveConfig function")
	}

	if !strings.Contains(code, "os.MkdirAll") {
		t.Error("saveConfig should create config directory")
	}

	if !strings.Contains(code, "os.WriteFile") {
		t.Error("saveConfig should write file")
	}

	if !strings.Contains(code, `provider = %q`) {
		t.Error("saveConfig should format TOML with provider")
	}

	if !strings.Contains(code, `model = %q`) {
		t.Error("saveConfig should format TOML with model")
	}

	if !strings.Contains(code, `api_key = %q`) {
		t.Error("saveConfig should format TOML with api_key")
	}
}

func TestGenerateAgent_ConfigPathUsesAgentName(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name:        "my-custom-agent",
			Description: "Custom agent",
		},
	}

	code, err := GenerateAgent(proj, "main")
	if err != nil {
		t.Fatalf("GenerateAgent() error = %v", err)
	}

	if !strings.Contains(code, `"my-custom-agent.toml"`) {
		t.Error("Config path should include agent name")
	}
}

func TestGenerateAgent_AgentConfigStruct(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
	}

	code, err := GenerateAgent(proj, "main")
	if err != nil {
		t.Fatalf("GenerateAgent() error = %v", err)
	}

	if !strings.Contains(code, "type AgentConfig struct {") {
		t.Error("Missing AgentConfig struct definition")
	}

	if !strings.Contains(code, "Provider string") {
		t.Error("AgentConfig should have Provider field")
	}

	if !strings.Contains(code, "Model    string") {
		t.Error("AgentConfig should have Model field")
	}

	if !strings.Contains(code, "APIKey   string") {
		t.Error("AgentConfig should have APIKey field")
	}
}

func TestGenerateAgent_FlagToEnvKeyFlow(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
	}

	code, err := GenerateAgent(proj, "main")
	if err != nil {
		t.Fatalf("GenerateAgent() error = %v", err)
	}

	if !strings.Contains(code, "os.Getenv(getAPIKeyEnv(provider))") {
		t.Error("Should call getAPIKeyEnv with provider flag to get API key from env")
	}
}

func TestGenerateAgent_DefaultEnvVarPattern(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
	}

	code, err := GenerateAgent(proj, "main")
	if err != nil {
		t.Fatalf("GenerateAgent() error = %v", err)
	}

	if !strings.Contains(code, `strings.ToUpper(p) + "_API_KEY"`) {
		t.Error("Should have default pattern for unknown providers")
	}
}
