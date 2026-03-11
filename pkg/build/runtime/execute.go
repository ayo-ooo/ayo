package runtime

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

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

		// Skip directories and the placeholder file
		if d.IsDir() || strings.HasSuffix(filePath, "placeholder") {
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

