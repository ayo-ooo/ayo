package runtime

import (
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
	"strconv"
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

