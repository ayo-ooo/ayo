package runtime

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"charm.land/fantasy"
	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/openai"
	"github.com/alexcabrera/ayo/internal/build/types"
	"github.com/pelletier/go-toml/v2"
)

// Execute runs the built agent with embedded resources.
// This function is called from the generated main() function.
func Execute(configToml, systemPrompt []byte, skillsFS, toolsFS embed.FS) error {
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

	// Create language model
	model, err := createLanguageModel(config.Agent.Model)
	if err != nil {
		return fmt.Errorf("create language model: %w", err)
	}

	// Create Fantasy agent with enhanced system prompt
	agent := fantasy.NewAgent(
		model,
		fantasy.WithSystemPrompt(fullSystemPrompt),
	)

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
	// Detect provider from model ID
	if strings.HasPrefix(modelID, "gpt-") {
		// OpenAI model
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
		}
		provider, err := openai.New(openai.WithAPIKey(apiKey))
		if err != nil {
			return nil, fmt.Errorf("create OpenAI provider: %w", err)
		}
		return provider.LanguageModel(context.Background(), modelID)
	}

	if strings.HasPrefix(modelID, "claude-") {
		// Anthropic model
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
		}
		provider, err := anthropic.New(anthropic.WithAPIKey(apiKey))
		if err != nil {
			return nil, fmt.Errorf("create Anthropic provider: %w", err)
		}
		return provider.LanguageModel(context.Background(), modelID)
	}

	// Default: try to infer from model ID
	// Try OpenAI as default
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		provider, err := openai.New(openai.WithAPIKey(apiKey))
		if err != nil {
			return nil, fmt.Errorf("create OpenAI provider: %w", err)
		}
		return provider.LanguageModel(context.Background(), modelID)
	}

	return nil, fmt.Errorf("unable to determine provider for model %s: set OPENAI_API_KEY or ANTHROPIC_API_KEY", modelID)
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

	// For now, just join arguments with spaces
	// TODO: Implement structured input parsing based on config.CLI.Mode and flags
	switch config.CLI.Mode {
	case "structured":
		// TODO: Implement structured flag parsing
		return strings.Join(args, " "), nil
	case "freeform":
		return strings.Join(args, " "), nil
	case "hybrid":
		// TODO: Implement hybrid mode (structured + freeform)
		return strings.Join(args, " "), nil
	default:
		return strings.Join(args, " "), nil
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

		// Skip directories
		if d.IsDir() {
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

