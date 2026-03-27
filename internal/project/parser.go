package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ayo-ooo/ayo/internal/schema"
)

func ParseProject(path string) (*Project, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("accessing path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	config, err := ParseConfig(filepath.Join(absPath, "config.toml"))
	if err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	system, err := os.ReadFile(filepath.Join(absPath, "system.md"))
	if err != nil {
		return nil, fmt.Errorf("reading system.md: %w", err)
	}

	project := &Project{
		Path:   absPath,
		Config: *config,
		System: string(system),
		Hooks:  make(map[HookType]string),
	}

	promptPath := filepath.Join(absPath, "prompt.tmpl")
	if data, err := os.ReadFile(promptPath); err == nil {
		p := string(data)
		project.Prompt = &p
	}

	inputPath := filepath.Join(absPath, "input.jsonschema")
	if data, err := os.ReadFile(inputPath); err == nil {
		parsed, err := schema.ParseSchema(data)
		if err != nil {
			return nil, fmt.Errorf("parsing input.jsonschema: %w", err)
		}
		project.Input = &Schema{
			Content: data,
			Parsed:  parsed,
		}
	}

	outputPath := filepath.Join(absPath, "output.jsonschema")
	if data, err := os.ReadFile(outputPath); err == nil {
		parsed, err := schema.ParseSchema(data)
		if err != nil {
			return nil, fmt.Errorf("parsing output.jsonschema: %w", err)
		}
		project.Output = &Schema{
			Content: data,
			Parsed:  parsed,
		}
	}

	skillsDir := filepath.Join(absPath, "skills")
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				skillPath := filepath.Join(skillsDir, entry.Name())
				skill, err := parseSkill(skillPath)
				if err != nil {
					return nil, fmt.Errorf("parsing skill %s: %w", entry.Name(), err)
				}
				project.Skills = append(project.Skills, *skill)
			}
		}
	}

	hooksDir := filepath.Join(absPath, "hooks")
	if entries, err := os.ReadDir(hooksDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				hookName := entry.Name()
				if isValidHookType(hookName) {
					project.Hooks[HookType(hookName)] = filepath.Join(hooksDir, hookName)
				}
			}
		}
	}

	return project, nil
}

func parseSkill(skillPath string) (*Skill, error) {
	skill := &Skill{
		Name: filepath.Base(skillPath),
		Path: skillPath,
	}

	skillMdPath := filepath.Join(skillPath, "SKILL.md")
	if data, err := os.ReadFile(skillMdPath); err == nil {
		desc := extractDescription(string(data))
		skill.Description = desc
	}

	return skill, nil
}

func extractDescription(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

func isValidHookType(name string) bool {
	hooks := []HookType{
		HookAgentStart, HookAgentFinish, HookAgentError,
		HookStepStart, HookStepFinish,
		HookTextStart, HookTextDelta, HookTextEnd,
		HookReasoningStart, HookReasoningDelta, HookReasoningEnd,
		HookToolInputStart, HookToolInputDelta, HookToolInputEnd,
		HookToolCall, HookToolResult,
		HookSource, HookStreamFinish, HookWarnings,
	}
	for _, h := range hooks {
		if string(h) == name {
			return true
		}
	}
	return false
}

func ValidateProject(p *Project) []*ValidationError {
	var errors []*ValidationError

	if p.Config.Name == "" {
		errors = append(errors, &ValidationError{
			File:    "config.toml",
			Message: "agent name is required",
		})
	}

	if p.Config.Version == "" {
		errors = append(errors, &ValidationError{
			File:    "config.toml",
			Message: "agent version is required",
		})
	}

	if p.System == "" {
		errors = append(errors, &ValidationError{
			File:    "system.md",
			Message: "system message is empty",
		})
	}

	if p.Input != nil {
		if _, ok := p.Input.Parsed.(*schema.ParsedSchema); !ok {
			errors = append(errors, &ValidationError{
				File:    "input.jsonschema",
				Message: "invalid JSON schema format",
			})
		}
	}

	if p.Output != nil {
		if _, ok := p.Output.Parsed.(*schema.ParsedSchema); !ok {
			errors = append(errors, &ValidationError{
				File:    "output.jsonschema",
				Message: "invalid JSON schema format",
			})
		}
	}

	// Validate input_order references
	if p.Input != nil && len(p.Config.InputOrder) > 0 {
		inputSchema := p.Input.Parsed.(*schema.ParsedSchema)
		schemaProps := make(map[string]bool)
		for name := range inputSchema.Properties {
			schemaProps[name] = true
		}

		for _, name := range p.Config.InputOrder {
			if !schemaProps[name] {
				errors = append(errors, &ValidationError{
					File:    "config.toml",
					Message: fmt.Sprintf("input_order references %q not found in input.jsonschema properties", name),
				})
			}
		}
	}

	for _, skill := range p.Skills {
		skillMdPath := filepath.Join(skill.Path, "SKILL.md")
		if _, err := os.Stat(skillMdPath); os.IsNotExist(err) {
			errors = append(errors, &ValidationError{
				File:    filepath.Join(skill.Path, "SKILL.md"),
				Message: "skill missing SKILL.md file",
			})
		}
	}

	return errors
}

func (s *Schema) UnmarshalJSON(data []byte) error {
	s.Content = data
	var parsed schema.ParsedSchema
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	s.Parsed = &parsed
	return nil
}
