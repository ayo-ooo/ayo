package ui

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"

	"github.com/alexcabrera/ayo/internal/skills"
)

// AgentCreateResult contains all the data collected from the agent creation wizard.
type AgentCreateResult struct {
	// Identity
	Handle      string
	Description string
	Model       string

	// Tools
	AllowedTools []string

	// Skills
	Skills              []string
	ExcludeSkills       []string
	IgnoreBuiltinSkills bool
	IgnoreSharedSkills  bool

	// System
	SystemMessage string
	SystemFile    string
	// Guardrails controls whether safety guardrails are applied.
	// When nil or true, guardrails are enabled (default).
	// Set to false to disable guardrails (dangerous - use with caution).
	Guardrails *bool

	// Chaining
	InputSchemaFile  string
	OutputSchemaFile string
}

// ToolInfo contains tool name and description for the wizard.
type ToolInfo struct {
	Name        string
	Description string
}

// AgentCreateFormOptions provides configuration for the agent creation form.
type AgentCreateFormOptions struct {
	// Models is the list of available models to select from.
	Models []string
	// AvailableSkills is the list of discovered skills.
	AvailableSkills []skills.Metadata
	// AvailableTools is the list of available tools with descriptions.
	AvailableTools []ToolInfo
	// PrefilledHandle is an optional pre-filled handle from CLI args.
	PrefilledHandle string
	// ExistingHandles is a set of existing agent handles (normalized, lowercase, no @ prefix) for conflict detection.
	ExistingHandles map[string]struct{}
}

// AgentCreateForm is a multi-step wizard for creating agents.
type AgentCreateForm struct {
	opts AgentCreateFormOptions
}

// NewAgentCreateForm creates a new agent creation form.
func NewAgentCreateForm(opts AgentCreateFormOptions) *AgentCreateForm {
	return &AgentCreateForm{opts: opts}
}

// Run executes the multi-step form wizard using huh.Form with groups.
func (f *AgentCreateForm) Run(ctx context.Context) (AgentCreateResult, error) {
	if len(f.opts.Models) == 0 {
		return AgentCreateResult{}, fmt.Errorf("no models configured; update config provider models")
	}

	var res AgentCreateResult

	// Handle without @ prefix
	var handleName string
	if strings.HasPrefix(f.opts.PrefilledHandle, "@") {
		handleName = strings.TrimPrefix(f.opts.PrefilledHandle, "@")
	} else {
		handleName = f.opts.PrefilledHandle
	}

	// Use pre-built existing handles set
	existingHandles := f.opts.ExistingHandles
	if existingHandles == nil {
		existingHandles = make(map[string]struct{})
	}

	// Default tools selection
	res.AllowedTools = []string{"bash"}

	// Tracking variables for conditional groups
	var systemSource string = "inline"

	// Build model options
	modelOpts := make([]huh.Option[string], 0, len(f.opts.Models))
	for _, m := range f.opts.Models {
		modelOpts = append(modelOpts, huh.NewOption(m, m))
	}

	// Build tool options with descriptions
	toolOpts := make([]huh.Option[string], 0, len(f.opts.AvailableTools))
	for _, t := range f.opts.AvailableTools {
		label := t.Name
		if t.Description != "" {
			// Format: "name - description" with description dimmed
			label = fmt.Sprintf("%s - %s", t.Name, t.Description)
		}
		toolOpts = append(toolOpts, huh.NewOption(label, t.Name))
	}

	hasSkills := len(f.opts.AvailableSkills) > 0

	// Get editor name for help text
	editorName := os.Getenv("EDITOR")
	if editorName == "" {
		editorName = "vim"
	}

	// Calculate total steps for progress indicator
	totalSteps := 5 // Identity, Tools, System, Chaining, Confirm
	if hasSkills {
		totalSteps = 6 // Add Skills step
	}

	// Step number tracking
	stepIdentity := 1
	stepTools := 2
	stepSkills := 3
	stepSystem := 4
	stepChaining := 5
	stepConfirm := 6
	if !hasSkills {
		stepSystem = 3
		stepChaining = 4
		stepConfirm = 5
	}

	// Helper to format step title
	stepTitle := func(step int, name string) string {
		return fmt.Sprintf("Step %d of %d: %s", step, totalSteps, name)
	}

	// Build all groups
	var groups []*huh.Group

	// Step 1: Identity
	groups = append(groups, huh.NewGroup(
		huh.NewInput().
			Title("Handle").
			Prompt("@ ").
			Placeholder("myagent").
			Value(&handleName).
			Validate(func(v string) error {
				if v == "" {
					return fmt.Errorf("required")
				}
				name := strings.TrimPrefix(v, "@")
				if name == "" {
					return fmt.Errorf("required")
				}
				if strings.HasPrefix(name, "ayo.") || name == "ayo" {
					return fmt.Errorf("cannot use reserved 'ayo' namespace")
				}
				if strings.ContainsAny(name, " \t\n") {
					return fmt.Errorf("handle cannot contain spaces")
				}
				// Check for conflicts with existing agents
				normalized := strings.ToLower(name)
				if _, exists := existingHandles[normalized]; exists {
					return fmt.Errorf("agent @%s already exists", name)
				}
				return nil
			}),
		huh.NewInput().
			Title("Description").
			Placeholder("A helpful assistant that...").
			Value(&res.Description),
		huh.NewSelect[string]().
			Title("Model").
			Options(modelOpts...).
			Value(&res.Model),
	).Title(stepTitle(stepIdentity, "Identity")).
		Description("Define your agent's identity. The handle is how you'll invoke it (e.g., ayo @myagent).\n\nThe description helps you remember what this agent does, and the model determines its capabilities."))

	// Step 2: Tools
	groups = append(groups, huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("Allowed Tools").
			Options(toolOpts...).
			Value(&res.AllowedTools),
	).Title(stepTitle(stepTools, "Tools")).
		Description("Tools let your agent interact with the outside world. The bash tool allows running shell commands.\n\nSelect which tools this agent should have access to."))

	// Step 3: Skills (only if skills exist)
	if hasSkills {
		// Helper to get required skills based on selected tools
		getRequiredSkills := func() []string {
			return skills.GetRequiredSkillsForTools(res.AllowedTools)
		}

		// Helper to format required skills note
		formatRequiredSkillsNote := func() string {
			reqs := skills.GetToolRequirementsForTools(res.AllowedTools)
			if len(reqs) == 0 {
				return ""
			}
			var lines []string
			lines = append(lines, "Skills Required by Tools:")
			for _, req := range reqs {
				// Format skill names as comma-separated inline code
				var skillCodes []string
				for _, s := range req.RequiredSkills {
					skillCodes = append(skillCodes, "`"+s+"`")
				}
				lines = append(lines, fmt.Sprintf("- `%s` requires: %s", req.ToolName, strings.Join(skillCodes, ", ")))
			}
			return strings.Join(lines, "\n")
		}

		// Build skill name set for filtering
		skillNameSet := make(map[string]struct{})
		for _, s := range f.opts.AvailableSkills {
			skillNameSet[s.Name] = struct{}{}
		}

		// Dynamic options that filter out required skills
		optionalSkillsOptionsFunc := func() []huh.Option[string] {
			requiredSet := make(map[string]struct{})
			for _, s := range getRequiredSkills() {
				requiredSet[s] = struct{}{}
			}

			var opts []huh.Option[string]
			for _, s := range f.opts.AvailableSkills {
				if _, isRequired := requiredSet[s.Name]; isRequired {
					continue // Skip required skills
				}
				label := s.Name
				if s.Description != "" {
					wrapped := wordWrap(s.Description, 50)
					label = fmt.Sprintf("%s\n    %s", s.Name, wrapped)
				}
				opts = append(opts, huh.NewOption(label, s.Name))
			}
			return opts
		}

		// Build the skills group - MultiSelect is the only field so it can fill available space
		// Required skills info is shown in the MultiSelect's description

		// Dynamic description for the MultiSelect
		skillsFieldDescriptionFunc := func() string {
			reqNote := formatRequiredSkillsNote()
			if reqNote != "" {
				return reqNote
			}
			return ""
		}

		// Add optional skills multi-select
		// Use a large height that will be clamped by the Group to available space
		groups = append(groups, huh.NewGroup(
			huh.NewMultiSelect[string]().
				TitleFunc(func() string {
					if len(getRequiredSkills()) > 0 {
						return "Optional Skills"
					}
					return "Available Skills"
				}, &res.AllowedTools).
				DescriptionFunc(skillsFieldDescriptionFunc, &res.AllowedTools).
				OptionsFunc(optionalSkillsOptionsFunc, &res.AllowedTools).
				Value(&res.Skills).
				Filterable(true).
				Height(100),
		).Title(stepTitle(stepSkills, "Skills")).
			Description("Skills are reusable instruction sets that teach your agent specialized tasks."))
	}

	// Step 4: System Prompt Source
	groups = append(groups, huh.NewGroup(
		huh.NewSelect[string]().
			Title("Source").
			Options(
				huh.NewOption("Write inline", "inline"),
				huh.NewOption("Load from file", "file"),
			).
			Value(&systemSource),
	).Title(stepTitle(stepSystem, "System Prompt")).
		Description("The system prompt defines your agent's personality, knowledge, and behavior.\n\nYou can write it directly or load from an existing markdown file."))

	// Step 4a: Inline system prompt
	groups = append(groups, huh.NewGroup(
		huh.NewText().
			Title("System Message").
			Description(fmt.Sprintf("ctrl+o to open in %s", editorName)).
			Placeholder("You are a helpful assistant...").
			Value(&res.SystemMessage).
			CharLimit(0).
			Lines(8).
			Editor(editorName),
	).Title(stepTitle(stepSystem, "System Prompt")).
		Description("Enter the system message that defines how your agent should behave.").
		WithHideFunc(func() bool {
			return systemSource != "inline"
		}))

	// Step 4b: File picker for system prompt
	groups = append(groups, huh.NewGroup(
		huh.NewFilePicker().
			Title("Select File").
			Description("Choose a .md or .txt file").
			CurrentDirectory(currentDir()).
			AllowedTypes([]string{".md", ".txt"}).
			Value(&res.SystemFile).
			Picking(true).
			Height(12),
	).Title(stepTitle(stepSystem, "System Prompt")).
		Description("Browse and select a markdown or text file containing your system prompt.").
		WithHideFunc(func() bool {
			return systemSource != "file"
		}))

	// Step 4b-preview: Confirm system prompt file with preview
	var confirmSystemFile bool = true // Preselect Yes
	groups = append(groups, huh.NewGroup(
		NewFilePreviewField().
			FilePath(&res.SystemFile).
			Title("File Contents").
			Height(15),
		huh.NewConfirm().
			TitleFunc(func() string {
				return fmt.Sprintf("Use %s?", shortenPath(res.SystemFile))
			}, &res.SystemFile).
			Affirmative("Yes").
			Negative("No, go back").
			Value(&confirmSystemFile),
	).Title(stepTitle(stepSystem, "System Prompt")).
		Description("Review the file contents below.").
		WithHideFunc(func() bool {
			return systemSource != "file" || res.SystemFile == ""
		}))

	// Step 4c: System Wrapper Option (shown after system prompt is defined)
	// Note: useGuardrails is the inverse of NoSystemWrapper
	// true = keep guardrails (NoSystemWrapper=false), false = disable (NoSystemWrapper=true)
	var useGuardrails bool = true // Default to keeping guardrails
	groups = append(groups, huh.NewGroup(
		huh.NewConfirm().
			Title("Use System Guardrails?").
			Description("Guardrails provide essential tool instructions, safety guidelines,\nand consistent behavior. Disabling may cause unexpected results.").
			Affirmative("Yes, keep guardrails").
			Negative("No, accept risks").
			Value(&useGuardrails),
	).Title(stepTitle(stepSystem, "System Prompt")).
		Description("The system guardrails are strongly recommended. They provide critical\ninstructions for tool usage, error handling, and agent behavior."))

	// Step 5: Structured I/O - Input Schema
	var useInputSchema bool
	groups = append(groups, huh.NewGroup(
		huh.NewConfirm().
			Title("Define Input Schema?").
			Description("An input schema validates JSON data passed to this agent via stdin.").
			Affirmative("Yes").
			Negative("No (freeform input)").
			Value(&useInputSchema),
	).Title(stepTitle(stepChaining, "Structured I/O")).
		Description("Schemas enable agent chaining via Unix pipes.\n\nDefine an input schema if this agent should accept structured JSON input."))

	// Step 5a: Input schema file picker
	groups = append(groups, huh.NewGroup(
		huh.NewFilePicker().
			Title("Select Input Schema").
			Description("Choose a .json or .jsonschema file").
			CurrentDirectory(currentDir()).
			AllowedTypes([]string{".json", ".jsonschema"}).
			Value(&res.InputSchemaFile).
			Picking(true).
			Height(10),
	).Title(stepTitle(stepChaining, "Input Schema")).
		Description("Select the JSON schema file that defines valid input for this agent.").
		WithHideFunc(func() bool {
			return !useInputSchema
		}))

	// Step 5a-preview: Confirm input schema with preview
	var confirmInputSchema bool = true // Preselect Yes
	groups = append(groups, huh.NewGroup(
		NewFilePreviewField().
			FilePath(&res.InputSchemaFile).
			Title("Schema Contents").
			Height(12),
		huh.NewConfirm().
			TitleFunc(func() string {
				return fmt.Sprintf("Use %s?", shortenPath(res.InputSchemaFile))
			}, &res.InputSchemaFile).
			Affirmative("Yes").
			Negative("No, go back").
			Value(&confirmInputSchema),
	).Title(stepTitle(stepChaining, "Input Schema")).
		Description("Review the schema below.").
		WithHideFunc(func() bool {
			return !useInputSchema || res.InputSchemaFile == ""
		}))

	// Step 5b: Structured I/O - Output Schema
	var useOutputSchema bool
	groups = append(groups, huh.NewGroup(
		huh.NewConfirm().
			Title("Define Output Schema?").
			Description("An output schema structures JSON data this agent produces.").
			Affirmative("Yes").
			Negative("No (freeform output)").
			Value(&useOutputSchema),
	).Title(stepTitle(stepChaining, "Structured I/O")).
		Description("Define an output schema if this agent should produce structured JSON output."))

	// Step 5b: Output schema file picker
	groups = append(groups, huh.NewGroup(
		huh.NewFilePicker().
			Title("Select Output Schema").
			Description("Choose a .json or .jsonschema file").
			CurrentDirectory(currentDir()).
			AllowedTypes([]string{".json", ".jsonschema"}).
			Value(&res.OutputSchemaFile).
			Picking(true).
			Height(10),
	).Title(stepTitle(stepChaining, "Output Schema")).
		Description("Select the JSON schema file that defines this agent's output format.").
		WithHideFunc(func() bool {
			return !useOutputSchema
		}))

	// Step 5b-preview: Confirm output schema with preview
	var confirmOutputSchema bool = true // Preselect Yes
	groups = append(groups, huh.NewGroup(
		NewFilePreviewField().
			FilePath(&res.OutputSchemaFile).
			Title("Schema Contents").
			Height(12),
		huh.NewConfirm().
			TitleFunc(func() string {
				return fmt.Sprintf("Use %s?", shortenPath(res.OutputSchemaFile))
			}, &res.OutputSchemaFile).
			Affirmative("Yes").
			Negative("No, go back").
			Value(&confirmOutputSchema),
	).Title(stepTitle(stepChaining, "Output Schema")).
		Description("Review the schema below.").
		WithHideFunc(func() bool {
			return !useOutputSchema || res.OutputSchemaFile == ""
		}))

	// Step 6: Review and Confirm
	var confirmed bool

	// Build review data provider that reads current form state
	reviewDataProvider := func() []ReviewSection {
		sections := []ReviewSection{}

		// Handle
		handle := "@" + strings.TrimPrefix(handleName, "@")
		sections = append(sections, ReviewSection{Label: "Handle", Value: handle})

		// Description
		if res.Description != "" {
			desc := res.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			sections = append(sections, ReviewSection{Label: "Description", Value: desc})
		}

		// Model
		sections = append(sections, ReviewSection{Label: "Model", Value: res.Model})

		// Tools
		if len(res.AllowedTools) > 0 {
			sections = append(sections, ReviewSection{Label: "Tools", Value: strings.Join(res.AllowedTools, ", ")})
		} else {
			sections = append(sections, ReviewSection{Label: "Tools", Value: "(none)"})
		}

		// Skills - show required and optional separately
		requiredSkills := skills.GetRequiredSkillsForTools(res.AllowedTools)
		if len(requiredSkills) > 0 {
			skillList := strings.Join(requiredSkills, ", ")
			if len(skillList) > 50 {
				skillList = skillList[:47] + "..."
			}
			sections = append(sections, ReviewSection{Label: "Required Skills", Value: skillList})
		}
		if len(res.Skills) > 0 {
			skillList := strings.Join(res.Skills, ", ")
			if len(skillList) > 50 {
				skillList = skillList[:47] + "..."
			}
			label := "Skills"
			if len(requiredSkills) > 0 {
				label = "Optional Skills"
			}
			sections = append(sections, ReviewSection{Label: label, Value: skillList})
		} else if len(requiredSkills) == 0 {
			sections = append(sections, ReviewSection{Label: "Skills", Value: "(none)"})
		}

		// System prompt - expandable
		if systemSource == "file" && res.SystemFile != "" {
			sections = append(sections, ReviewSection{
				Label:      "System",
				Value:      shortenPath(res.SystemFile),
				Expandable: true,
				Content:    readFileContent(res.SystemFile),
			})
		} else if res.SystemMessage != "" {
			preview := res.SystemMessage
			if len(preview) > 40 {
				preview = preview[:37] + "..."
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			sections = append(sections, ReviewSection{
				Label:      "System",
				Value:      preview,
				Expandable: true,
				Content:    res.SystemMessage,
			})
		} else {
			sections = append(sections, ReviewSection{Label: "System", Value: "(default)"})
		}

		// Guardrails status
		if res.Guardrails != nil && !*res.Guardrails {
			sections = append(sections, ReviewSection{Label: "Guardrails", Value: "disabled (dangerous)"})
		}

		// Input schema - expandable (independent of output schema)
		if res.InputSchemaFile != "" {
			sections = append(sections, ReviewSection{
				Label:      "Input Schema",
				Value:      shortenPath(res.InputSchemaFile),
				Expandable: true,
				Content:    readFileContent(res.InputSchemaFile),
			})
		}

		// Output schema - expandable (independent of input schema)
		if res.OutputSchemaFile != "" {
			sections = append(sections, ReviewSection{
				Label:      "Output Schema",
				Value:      shortenPath(res.OutputSchemaFile),
				Expandable: true,
				Content:    readFileContent(res.OutputSchemaFile),
			})
		}

		return sections
	}

	groups = append(groups, huh.NewGroup(
		NewReviewField().
			DataProvider(reviewDataProvider).
			Height(18),
		huh.NewConfirm().
			Title("Create this agent?").
			Affirmative("Create").
			Negative("Go Back").
			Value(&confirmed),
	).Title(stepTitle(stepConfirm, "Review & Confirm")).
		Description("Review your configuration. Use ↑/↓ to navigate, Enter to expand/collapse."))

	// Create custom keymap with ctrl+o for editor instead of ctrl+e
	customKeymap := huh.NewDefaultKeyMap()
	customKeymap.Text.Editor = key.NewBinding(key.WithKeys("ctrl+o"), key.WithHelp("ctrl+o", "open editor"))

	// Create and run the form with full-screen layout
	form := huh.NewForm(groups...).
		WithTheme(wizardTheme()).
		WithKeyMap(customKeymap).
		WithShowHelp(false) // We handle help in the wrapper

	wrapper := newFullScreenForm(form)
	if err := wrapper.Run(); err != nil {
		if err.Error() == "user aborted" {
			return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
		}
		return AgentCreateResult{}, err
	}

	// Check if form was aborted
	if wrapper.State() == huh.StateAborted {
		return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
	}

	// Check if user cancelled
	if !confirmed {
		return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
	}

	// Normalize handle
	handleName = strings.TrimPrefix(handleName, "@")
	res.Handle = "@" + handleName

	// Convert useGuardrails (true = keep guardrails) to Guardrails pointer
	if !useGuardrails {
		f := false
		res.Guardrails = &f
	}

	// Merge required skills into the skills list
	requiredSkills := skills.GetRequiredSkillsForTools(res.AllowedTools)
	if len(requiredSkills) > 0 {
		// Build set of already-selected skills
		skillSet := make(map[string]struct{})
		for _, s := range res.Skills {
			skillSet[s] = struct{}{}
		}
		// Add required skills that aren't already selected
		for _, s := range requiredSkills {
			if _, exists := skillSet[s]; !exists {
				res.Skills = append(res.Skills, s)
			}
		}
	}

	// If file source was selected, read the file content
	if systemSource == "file" && res.SystemFile != "" {
		expanded := expandPath(res.SystemFile)
		data, err := os.ReadFile(expanded)
		if err != nil {
			return AgentCreateResult{}, fmt.Errorf("read system file: %w", err)
		}
		res.SystemMessage = string(data)
	}

	return res, nil
}

// userHomeDir returns the user's home directory or empty string if unavailable.
func userHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// currentDir returns the current working directory or home directory as fallback.
func currentDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return userHomeDir()
}

// shortenPath shortens a path by replacing home directory with ~.
func shortenPath(path string) string {
	home := userHomeDir()
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return strings.Replace(path, "~", home, 1)
	}
	return path
}

// wordWrap wraps text to the specified width, joining lines with newline and indent.
func wordWrap(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var lines []string
	words := strings.Fields(text)
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len() == 0 {
			currentLine.WriteString(word)
		} else if currentLine.Len()+1+len(word) <= width {
			currentLine.WriteString(" ")
			currentLine.WriteString(word)
		} else {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	// Join with newline and indent for continuation lines
	if len(lines) == 0 {
		return text
	}
	if len(lines) == 1 {
		return lines[0]
	}

	var result strings.Builder
	result.WriteString(lines[0])
	for i := 1; i < len(lines); i++ {
		result.WriteString("\n    ")
		result.WriteString(lines[i])
	}
	return result.String()
}
