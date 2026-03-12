package runtime

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"charm.land/catwalk/pkg/embedded"
	"charm.land/fantasy"
	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/google"
	"charm.land/fantasy/providers/openai"
	"charm.land/fantasy/providers/openaicompat"
	"charm.land/fantasy/providers/openrouter"
	"github.com/alexcabrera/ayo/internal/build/types"
	"github.com/pelletier/go-toml/v2"
)

// Execute runs the built agent with embedded resources.
// This function is called from the generated main() function.
func Execute(configToml, systemPrompt []byte, skillsFS, toolsFS embed.FS) error {
	// Check for --help/-h before doing anything else
	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			return printHelp()
		}
	}

	// Parse embedded config
	var config types.Config
	if err := toml.Unmarshal(configToml, &config); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	// Load skills from embedded filesystem
	skills, err := loadSkills(skillsFS)
	if err != nil {
		return fmt.Errorf("load skills: %w", err)
	}

	// Combine system prompt with skills
	fullSystemPrompt := combineSystemPromptAndSkills(string(systemPrompt), skills)

	// Load tools from embedded filesystem
	tools, err := loadTools(toolsFS, config.Agent.Tools.Allowed)
	if err != nil {
		return fmt.Errorf("load tools: %w", err)
	}

	// Create language model
	model, err := createLanguageModel(config.Agent.Model)
	if err != nil {
		return fmt.Errorf("create language model: %w", err)
	}

	// Create Fantasy agent with enhanced system prompt and tools
	agentOpts := []fantasy.AgentOption{
		fantasy.WithSystemPrompt(fullSystemPrompt),
	}
	if len(tools) > 0 {
		agentOpts = append(agentOpts, fantasy.WithTools(tools...))
	}
	agent := fantasy.NewAgent(model, agentOpts...)

	// Get user input from CLI args or stdin
	input, err := getUserInput(&config)
	if err != nil {
		return fmt.Errorf("get input: %w", err)
	}

	// Execute agent
	ctx := context.Background()
	response, err := agent.Generate(ctx, fantasy.AgentCall{
		Prompt: input,
	})
	if err != nil {
		return fmt.Errorf("execute agent: %w", err)
	}

	// Format and output response
	fmt.Println(formatResponse(response.Response.Content.Text()))

	return nil
}

// createLanguageModel creates a Fantasy language model from the model ID.
// Supports common model providers: OpenAI, Anthropic, etc.
func createLanguageModel(modelID string) (fantasy.LanguageModel, error) {
	// Get API key with fallback to saved config
	apiKey, provider, modelID, err := getAPIKey(modelID)
	if err != nil {
		return nil, err
	}

	// Create provider based on detected provider type
	switch provider {
	case "openai":
		prov, err := openai.New(openai.WithAPIKey(apiKey))
		if err != nil {
			return nil, fmt.Errorf("create OpenAI provider: %w", err)
		}
		return prov.LanguageModel(context.Background(), modelID)
	case "anthropic":
		prov, err := anthropic.New(anthropic.WithAPIKey(apiKey))
		if err != nil {
			return nil, fmt.Errorf("create Anthropic provider: %w", err)
		}
		return prov.LanguageModel(context.Background(), modelID)
	case "google":
		prov, err := google.New(google.WithGeminiAPIKey(apiKey))
		if err != nil {
			return nil, fmt.Errorf("create Google provider: %w", err)
		}
		return prov.LanguageModel(context.Background(), modelID)
	case "openrouter":
		prov, err := openrouter.New(openrouter.WithAPIKey(apiKey))
		if err != nil {
			return nil, fmt.Errorf("create OpenRouter provider: %w", err)
		}
		return prov.LanguageModel(context.Background(), modelID)
	case "xai":
		// xAI uses OpenAI-compatible API
		prov, err := openaicompat.New(
			openaicompat.WithAPIKey(apiKey),
			openaicompat.WithBaseURL("https://api.x.ai/v1"),
		)
		if err != nil {
			return nil, fmt.Errorf("create xAI provider: %w", err)
		}
		return prov.LanguageModel(context.Background(), modelID)
	case "groq":
		// Groq uses OpenAI-compatible API
		prov, err := openaicompat.New(
			openaicompat.WithAPIKey(apiKey),
			openaicompat.WithBaseURL("https://api.groq.com/openai/v1"),
		)
		if err != nil {
			return nil, fmt.Errorf("create Groq provider: %w", err)
		}
		return prov.LanguageModel(context.Background(), modelID)
	case "deepseek":
		// DeepSeek uses OpenAI-compatible API
		prov, err := openaicompat.New(
			openaicompat.WithAPIKey(apiKey),
			openaicompat.WithBaseURL("https://api.deepseek.com"),
		)
		if err != nil {
			return nil, fmt.Errorf("create DeepSeek provider: %w", err)
		}
		return prov.LanguageModel(context.Background(), modelID)
	case "cerebras":
		// Cerebras uses OpenAI-compatible API
		prov, err := openaicompat.New(
			openaicompat.WithAPIKey(apiKey),
			openaicompat.WithBaseURL("https://api.cerebras.ai/v1"),
		)
		if err != nil {
			return nil, fmt.Errorf("create Cerebras provider: %w", err)
		}
		return prov.LanguageModel(context.Background(), modelID)
	case "together":
		// Together.ai uses OpenAI-compatible API
		prov, err := openaicompat.New(
			openaicompat.WithAPIKey(apiKey),
			openaicompat.WithBaseURL("https://api.together.xyz/v1"),
		)
		if err != nil {
			return nil, fmt.Errorf("create Together.ai provider: %w", err)
		}
		return prov.LanguageModel(context.Background(), modelID)
	case "azure":
		// Azure OpenAI requires different handling (endpoint in env var)
		endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
		if endpoint == "" {
			return nil, fmt.Errorf("AZURE_OPENAI_ENDPOINT not set")
		}
		prov, err := openaicompat.New(
			openaicompat.WithAPIKey(apiKey),
			openaicompat.WithBaseURL(endpoint),
		)
		if err != nil {
			return nil, fmt.Errorf("create Azure OpenAI provider: %w", err)
		}
		return prov.LanguageModel(context.Background(), modelID)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

// RuntimeConfig stores user preferences for the agent
type RuntimeConfig struct {
	Provider string `toml:"provider"`
	Model    string `toml:"model"` // Optional: can be empty for auto-detection
}

// getAPIKey gets the API key, provider, and model ID from environment or saved config.
// If modelID is empty, auto-detects the provider and uses its default model.
// Returns: apiKey, providerID, modelID, error
func getAPIKey(modelID string) (string, string, string, error) {
	availableKeys := detectAvailableKeys()

	if len(availableKeys) == 0 {
		return "", "", "", fmt.Errorf("no API keys found in environment\n\nSet one of these environment variables to get started:\n  - ANTHROPIC_API_KEY (for Claude models)\n  - OPENAI_API_KEY (for GPT models)\n  - GEMINI_API_KEY (for Gemini models)\n  - OPENROUTER_API_KEY (for multi-provider access)\n  - XAI_API_KEY (for Grok models)\n  - GROQ_API_KEY (for fast inference)\n  - DEEPSEEK_API_KEY (for DeepSeek models)\n  - CEREBRAS_API_KEY (for Cerebras models)\n  - TOGETHER_API_KEY (for Together AI)\n  - AZURE_OPENAI_API_KEY (for Azure OpenAI)")
	}

	// If model ID is specified, use the legacy path
	if modelID != "" {
		return getAPIKeyForModel(modelID, availableKeys)
	}

	// Auto-detect: no model specified
	return getAPIKeyAutoDetect(availableKeys)
}

// getAPIKeyForModel gets the API key for a specified model ID.
// Returns: apiKey, providerID, modelID, error
func getAPIKeyForModel(modelID string, availableKeys map[string]string) (string, string, string, error) {
	detectedProvider := detectProviderFromModel(modelID)

	// Try to load saved config
	config, err := loadRuntimeConfig()
	if err == nil && config.Provider != "" {
		// Check if saved provider's key is available
		if _, ok := availableKeys[config.Provider]; ok {
			return availableKeys[config.Provider], config.Provider, modelID, nil
		}
	}

	// If we detected a provider from model ID and that key is available
	if detectedProvider != "" {
		if key, ok := availableKeys[detectedProvider]; ok {
			// Save this choice for future use
			saveRuntimeConfig(&RuntimeConfig{Provider: detectedProvider, Model: modelID})
			return key, detectedProvider, modelID, nil
		}
	}

	// If only one key is available, use it
	if len(availableKeys) == 1 {
		for provider, key := range availableKeys {
			// Save this choice for future use
			saveRuntimeConfig(&RuntimeConfig{Provider: provider, Model: modelID})
			return key, provider, modelID, nil
		}
	}

	// Multiple keys available, no preference saved - prompt user
	provider := promptForProvider(availableKeys)
	if provider == "" {
		return "", "", "", fmt.Errorf("no provider selected")
	}

	// Save user's choice
	saveRuntimeConfig(&RuntimeConfig{Provider: provider, Model: modelID})
	return availableKeys[provider], provider, modelID, nil
}

// getAPIKeyAutoDetect auto-detects provider and model when none is specified.
// Returns: apiKey, providerID, modelID, error
func getAPIKeyAutoDetect(availableKeys map[string]string) (string, string, string, error) {
	// Try to load saved config
	config, err := loadRuntimeConfig()
	if err == nil && config.Provider != "" {
		// Check if saved provider's key is available
		if key, ok := availableKeys[config.Provider]; ok {
			// If saved model exists and is valid, use it
			modelID := config.Model
			if modelID == "" {
				modelID = getProviderDefaultModel(config.Provider)
			}
			if modelID != "" {
				// Save the model to config if it wasn't there before
				if config.Model == "" {
					saveRuntimeConfig(&RuntimeConfig{Provider: config.Provider, Model: modelID})
				}
				return key, config.Provider, modelID, nil
			}
		}
	}

	// Priority order: openai > anthropic > google > openrouter > first available
	priority := []string{"openai", "anthropic", "google", "openrouter"}
	for _, provider := range priority {
		if key, ok := availableKeys[provider]; ok {
			modelID := getProviderDefaultModel(provider)
			if modelID != "" {
				saveRuntimeConfig(&RuntimeConfig{Provider: provider, Model: modelID})
				return key, provider, modelID, nil
			}
		}
	}

	// If only one key is available, use it
	if len(availableKeys) == 1 {
		for provider, key := range availableKeys {
			modelID := getProviderDefaultModel(provider)
			if modelID != "" {
				saveRuntimeConfig(&RuntimeConfig{Provider: provider, Model: modelID})
				return key, provider, modelID, nil
			}
		}
	}

	// Multiple keys available, no priority matched - prompt user
	provider := promptForProvider(availableKeys)
	if provider == "" {
		return "", "", "", fmt.Errorf("no provider selected")
	}

	modelID := getProviderDefaultModel(provider)
	if modelID == "" {
		return "", "", "", fmt.Errorf("unable to determine default model for provider '%s'. This is unexpected - please report this issue.", provider)
	}

	// Save user's choice
	saveRuntimeConfig(&RuntimeConfig{Provider: provider, Model: modelID})
	return availableKeys[provider], provider, modelID, nil
}

// detectProviderFromModel detects the provider from the model ID
func detectProviderFromModel(modelID string) string {
	switch {
	case strings.HasPrefix(modelID, "gpt-") || strings.HasPrefix(modelID, "o1-") || strings.HasPrefix(modelID, "o3-"):
		return "openai"
	case strings.HasPrefix(modelID, "claude-"):
		return "anthropic"
	case strings.HasPrefix(modelID, "gemini-"):
		return "google"
	case strings.HasPrefix(modelID, "grok-"):
		return "xai"
	case strings.HasPrefix(modelID, "deepseek-"):
		return "deepseek"
	case strings.HasPrefix(modelID, "llama-") || strings.HasPrefix(modelID, "mixtral-"):
		// Could be groq, together, or openrouter - ambiguous
		return ""
	default:
		return ""
	}
}

// providerEnvVars maps provider IDs to their environment variable names
var providerEnvVars = map[string]string{
	"anthropic":   "ANTHROPIC_API_KEY",
	"openai":      "OPENAI_API_KEY",
	"google":      "GEMINI_API_KEY",
	"openrouter":  "OPENROUTER_API_KEY",
	"azure":       "AZURE_OPENAI_API_KEY",
	"groq":        "GROQ_API_KEY",
	"deepseek":    "DEEPSEEK_API_KEY",
	"cerebras":    "CEREBRAS_API_KEY",
	"xai":         "XAI_API_KEY",
	"together":    "TOGETHER_API_KEY",
}

// detectAvailableKeys checks which API keys are available in the environment
func detectAvailableKeys() map[string]string {
	keys := make(map[string]string)

	for provider, envVar := range providerEnvVars {
		if key := os.Getenv(envVar); key != "" {
			keys[provider] = key
		}
	}

	return keys
}

// getProviderDefaultModel returns the default model ID for a provider
// Uses catwalk's embedded provider configurations.
func getProviderDefaultModel(providerID string) string {
	for _, p := range embedded.GetAll() {
		if string(p.ID) == providerID {
			return p.DefaultLargeModelID
		}
	}
	return ""
}

// getConfigPath returns the path to the runtime config file
func getConfigPath() (string, error) {
	// Get binary name
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	binaryName := filepath.Base(execPath)

	// Use ~/.config/{binary-name}/config.toml
	configDir := filepath.Join(os.Getenv("HOME"), ".config", binaryName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.toml"), nil
}

// loadRuntimeConfig loads the runtime config from disk
func loadRuntimeConfig() (*RuntimeConfig, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config RuntimeConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// saveRuntimeConfig saves the runtime config to disk
func saveRuntimeConfig(config *RuntimeConfig) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// promptForProvider prompts the user to choose a provider
func promptForProvider(availableKeys map[string]string) string {
	fmt.Println("\nMultiple API keys detected. Please choose a provider:")
	fmt.Println()

	providers := make([]string, 0, len(availableKeys))
	i := 1
	for provider := range availableKeys {
		providers = append(providers, provider)
		fmt.Printf("  %d. %s\n", i, provider)
		i++
	}

	fmt.Println()
	fmt.Printf("Enter choice [1-%d]: ", len(providers))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(providers) {
		fmt.Println("Invalid choice")
		return ""
	}

	selected := providers[choice-1]
	fmt.Printf("\nUsing %s. This choice will be saved for future runs.\n", selected)

	return selected
}

// printHelp displays usage information
func printHelp() error {
	// Get binary name
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	binaryName := filepath.Base(execPath)

	fmt.Printf("Usage: %s [prompt]\n\n", binaryName)
	fmt.Printf("An AI agent that responds to your prompts.\n\n")
	fmt.Printf("Provider Setup:\n")
	fmt.Printf("  Set an API key environment variable and we'll auto-detect the provider and model.\n")
	fmt.Printf("  Supported providers: ANTHROPIC_API_KEY, OPENAI_API_KEY, GEMINI_API_KEY,\n")
	fmt.Printf("  XAI_API_KEY, GROQ_API_KEY, DEEPSEEK_API_KEY, CEREBRAS_API_KEY,\n")
	fmt.Printf("  TOGETHER_API_KEY, OPENROUTER_API_KEY, or AZURE_OPENAI_API_KEY\n\n")
	fmt.Printf("Options:\n")
	fmt.Printf("  -h, --help    Show this help message\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  %s \"What is the capital of France?\"\n", binaryName)
	fmt.Printf("  echo \"Explain quantum computing\" | %s\n", binaryName)
	fmt.Printf("  export OPENAI_API_KEY=sk-xxx && %s \"Hello world\"\n", binaryName)
	return nil
}

// getUserInput gets user input based on CLI configuration.
func getUserInput(config *types.Config) (string, error) {
	// Check for input from stdin (pipe or redirect)
	if isPiped() {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(data), nil
	}

	// Get input from command-line arguments
	args := os.Args[1:] // Skip program name

	switch config.CLI.Mode {
	case "structured":
		return parseStructuredInput(args, config)
	case "freeform":
		return strings.Join(args, " "), nil
	case "hybrid":
		// Hybrid mode: try structured first, fall back to freeform
		input, err := parseStructuredInput(args, config)
		if err != nil {
			// If structured parsing fails, treat as freeform
			return strings.Join(args, " "), nil
		}
		return input, nil
	default:
		return strings.Join(args, " "), nil
	}
}

// parseStructuredInput parses command-line arguments according to the config's flags.
func parseStructuredInput(args []string, config *types.Config) (string, error) {
	// If no flags are defined, return error to trigger fallback in hybrid mode
	if len(config.CLI.Flags) == 0 {
		return "", fmt.Errorf("no flags defined")
	}

	// Create a map to collect flag values
	values := make(map[string]any)
	positionals := make([]string, 0)

	// Find maximum position for positional arguments
	maxPosition := -1
	for _, flag := range config.CLI.Flags {
		if flag.Position > maxPosition {
			maxPosition = flag.Position
		}
	}
	// Initialize positional array with empty strings
	for i := 0; i <= maxPosition; i++ {
		positionals = append(positionals, "")
	}

	// Parse arguments
	i := 0
	for i < len(args) {
		arg := args[i]

		// Check if it's a flag (--flag or -s)
		if strings.HasPrefix(arg, "--") {
			flagName := strings.TrimPrefix(arg, "--")
			// Find the flag definition
			flagDef, exists := config.CLI.Flags[flagName]
			if !exists {
				// Unknown flag, might be freeform input
				return "", fmt.Errorf("unknown flag: %s", flagName)
			}

			// Parse the flag value
			if flagDef.Type == "bool" {
				// Boolean flags don't need a value
				values[flagName] = true
				i++
			} else if flagDef.Multiple {
				// Collect multiple values
				var vals []string
				i++
				for i < len(args) && !strings.HasPrefix(args[i], "-") {
					vals = append(vals, args[i])
					i++
				}
				values[flagName] = vals
			} else {
				// Single value
				if i+1 >= len(args) {
					return "", fmt.Errorf("flag %s requires a value", flagName)
				}
				value, err := parseFlagValue(args[i+1], flagDef.Type)
				if err != nil {
					return "", fmt.Errorf("invalid value for flag %s: %w", flagName, err)
				}
				values[flagName] = value
				i += 2
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) == 2 {
			// Short flag
			shortFlag := strings.TrimPrefix(arg, "-")
			// Find the flag with this short name
			var flagName string
			var flagDef types.CLIFlag
			for name, flag := range config.CLI.Flags {
				if flag.Short == shortFlag {
					flagName = name
					flagDef = flag
					break
				}
			}
			if flagName == "" {
				return "", fmt.Errorf("unknown short flag: -%s", shortFlag)
			}

			// Parse the flag value (same logic as long flags)
			if flagDef.Type == "bool" {
				values[flagName] = true
				i++
			} else {
				if i+1 >= len(args) {
					return "", fmt.Errorf("flag -%s requires a value", shortFlag)
				}
				value, err := parseFlagValue(args[i+1], flagDef.Type)
				if err != nil {
					return "", fmt.Errorf("invalid value for flag -%s: %w", shortFlag, err)
				}
				values[flagName] = value
				i += 2
			}
		} else {
			// Positional argument
			// Find which flag this belongs to
			var flagName string
			for name, flag := range config.CLI.Flags {
				if flag.Position == len(positionals)-1 {
					flagName = name
					break
				}
			}
			if flagName != "" {
				// This is a positional argument for a flag
				flagDef := config.CLI.Flags[flagName]
				value, err := parseFlagValue(arg, flagDef.Type)
				if err != nil {
					return "", fmt.Errorf("invalid value for positional argument %s: %w", flagName, err)
				}
				values[flagName] = value
			}
			positionals[len(positionals)-1] = arg
			i++
		}
	}

	// Set defaults for missing flags
	for name, flag := range config.CLI.Flags {
		if _, exists := values[name]; !exists && flag.Default != nil {
			values[name] = flag.Default
		}
		// Check required flags
		if flag.Required && values[name] == nil {
			return "", fmt.Errorf("required flag %s is missing", name)
		}
	}

	// Convert to JSON for structured input
	jsonData, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("marshal input: %w", err)
	}

	return string(jsonData), nil
}

// parseFlagValue parses a string value according to the flag type.
func parseFlagValue(value, flagType string) (any, error) {
	switch flagType {
	case "string":
		return value, nil
	case "int":
		return strconv.Atoi(value)
	case "float":
		return strconv.ParseFloat(value, 64)
	case "bool":
		return strconv.ParseBool(value)
	case "array":
		return []string{value}, nil
	default:
		return nil, fmt.Errorf("unknown flag type: %s", flagType)
	}
}

// isPiped checks if input is coming from a pipe or redirect.
func isPiped() bool {
	info, _ := os.Stdin.Stat()
	return (info.Mode() & os.ModeCharDevice) == 0
}

// formatResponse formats the agent's response for output.
func formatResponse(response string) string {
	// Use lipgloss for nice formatting
	style := lipgloss.NewStyle().
		Width(80).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("62"))
	return style.Render(response)
}

// loadSkills reads all skill files from the embedded filesystem.
// Returns a map of skill name to content.
func loadSkills(skillsFS embed.FS) (map[string]string, error) {
	skills := make(map[string]string)

	// Walk the skills directory
	if err := fs.WalkDir(skillsFS, "skills", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and placeholder files
		if d.IsDir() || strings.HasSuffix(filePath, "placeholder") || strings.HasSuffix(filePath, ".gitkeep") {
			return nil
		}

		// Read file content
		content, err := skillsFS.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read skill file %s: %w", filePath, err)
		}

		// Extract skill name from path
		// filePath is like "skills/coding.md" or "skills/writing.md"
		skillName := path.Base(filePath)

		skills[skillName] = string(content)

		return nil
	}); err != nil && err != fs.ErrNotExist {
		// ErrNotExist is OK - it means no skills directory
		return nil, fmt.Errorf("walk skills directory: %w", err)
	}

	return skills, nil
}

// combineSystemPromptAndSkills combines the system prompt with loaded skills.
// Skills are appended to the system prompt in a structured format.
func combineSystemPromptAndSkills(systemPrompt string, skills map[string]string) string {
	if len(skills) == 0 {
		return systemPrompt
	}

	var builder strings.Builder

	// Add original system prompt
	if systemPrompt != "" {
		builder.WriteString(systemPrompt)
		builder.WriteString("\n\n")
	}

	// Add skills section
	builder.WriteString("## Skills\n\n")
	builder.WriteString("You have the following skills available:\n\n")

	for name, content := range skills {
		// Remove .md extension for cleaner display
		displayName := strings.TrimSuffix(name, ".md")
		builder.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", displayName, content))
	}

	return builder.String()
}

// CustomToolParams defines parameters for custom tool execution.
type CustomToolParams struct {
	Args string `json:"args,omitempty" description:"Arguments to pass to the tool (space-separated)"`
}

// loadTools reads all tool scripts from the embedded filesystem.
// Returns a list of Fantasy tools.
func loadTools(toolsFS embed.FS, allowedTools []string) ([]fantasy.AgentTool, error) {
	var tools []fantasy.AgentTool

	// If no allowed tools specified, return empty list
	if len(allowedTools) == 0 {
		return tools, nil
	}

	// Create a set of allowed tool names for quick lookup
	allowed := make(map[string]bool)
	for _, t := range allowedTools {
		allowed[t] = true
	}

	// Walk the tools directory
	if err := fs.WalkDir(toolsFS, "tools", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and placeholder files
		if d.IsDir() || strings.HasSuffix(filePath, "placeholder") || strings.HasSuffix(filePath, ".gitkeep") {
			return nil
		}

		// Extract tool name from path
		// filePath is like "tools/mytool.sh" or "tools/mytool"
		toolName := path.Base(filePath)
		toolName = strings.TrimSuffix(toolName, path.Ext(toolName))

		// Check if tool is in allowed list
		if !allowed[toolName] {
			return nil
		}

		// Read tool script content
		scriptContent, err := toolsFS.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read tool file %s: %w", filePath, err)
		}

		// Extract description from first comment line
		description := extractToolDescription(string(scriptContent), toolName)

		// Create a Fantasy tool that executes the script
		tool := fantasy.NewAgentTool(
			toolName,
			description,
			func(ctx context.Context, params CustomToolParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				return executeToolScript(ctx, filePath, params.Args, toolsFS, scriptContent)
			},
		)

		tools = append(tools, tool)

		return nil
	}); err != nil && err != fs.ErrNotExist {
		// ErrNotExist is OK - it means no tools directory
		return nil, fmt.Errorf("walk tools directory: %w", err)
	}

	// Always include bash tool if allowed
	if allowed["bash"] {
		tools = append(tools, createBashTool())
	}

	return tools, nil
}

// extractToolDescription extracts a description from the tool script.
// Looks for the first comment line that describes the tool.
func extractToolDescription(script, toolName string) string {
	lines := strings.Split(script, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// Remove the # and trim
			desc := strings.TrimSpace(line[1:])
			if desc != "" {
				return desc
			}
		}
	}
	return fmt.Sprintf("Execute the %s tool", toolName)
}

// createBashTool creates the standard bash tool for shell command execution.
func createBashTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"bash",
		"Execute a shell command and return stdout/stderr",
		func(ctx context.Context, params BashParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if strings.TrimSpace(params.Command) == "" {
				return fantasy.NewTextErrorResponse("command is required; provide a string like {\"command\":\"echo hello world\"}"), nil
			}

			timeout := 30 * time.Second
			if params.TimeoutSeconds > 0 {
				timeout = time.Duration(params.TimeoutSeconds) * time.Second
			}

			execCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// Execute the command
			cmd := exec.CommandContext(execCtx, "/bin/sh", "-c", params.Command)
			if params.WorkingDir != "" {
				cmd.Dir = params.WorkingDir
			}

			output, err := cmd.CombinedOutput()

			if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
				return fantasy.NewTextErrorResponse("bash command timed out"), nil
			}

			if err != nil {
				return fantasy.NewTextResponse(string(output)), nil
			}

			return fantasy.NewTextResponse(string(output)), nil
		},
	)
}

// BashParams defines parameters for bash command execution.
type BashParams struct {
	Command        string `json:"command"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
	WorkingDir     string `json:"working_dir,omitempty"`
}

// executeToolScript executes a custom tool script with optional arguments.
func executeToolScript(ctx context.Context, scriptPath string, args string, toolsFS embed.FS, scriptContent []byte) (fantasy.ToolResponse, error) {
	// Create a temp file for the script
	tmpFile, err := os.CreateTemp("", "ayo-tool-*.sh")
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("create temp file: %v", err)), nil
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write script content to temp file
	if _, err := tmpFile.Write(scriptContent); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("write script: %v", err)), nil
	}

	// Make script executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("make executable: %v", err)), nil
	}

	// Build command with arguments
	cmdArgs := []string{tmpFile.Name()}
	if strings.TrimSpace(args) != "" {
		cmdArgs = append(cmdArgs, strings.Fields(args)...)
	}

	// Execute the script
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fantasy.NewTextResponse(fmt.Sprintf("Tool exited with error:\n%s", string(output))), nil
	}

	return fantasy.NewTextResponse(string(output)), nil
}

