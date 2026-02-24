package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// ValidationError represents a skill validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// Validate checks a skill directory for compliance with the agentskills spec.
// Returns a list of validation errors (empty if valid).
func Validate(skillDir string) []ValidationError {
	var errors []ValidationError

	// Check directory exists
	info, err := os.Stat(skillDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ValidationError{{Message: fmt.Sprintf("path does not exist: %s", skillDir)}}
		}
		return []ValidationError{{Message: fmt.Sprintf("cannot access path: %s", err)}}
	}
	if !info.IsDir() {
		return []ValidationError{{Message: fmt.Sprintf("not a directory: %s", skillDir)}}
	}

	// Check SKILL.md exists
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	content, err := os.ReadFile(skillMDPath)
	if err != nil {
		// Try lowercase
		skillMDPath = filepath.Join(skillDir, "skill.md")
		content, err = os.ReadFile(skillMDPath)
		if err != nil {
			return []ValidationError{{Message: "missing required file: SKILL.md"}}
		}
	}

	// Parse frontmatter
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 || parts[0] != "" {
		return []ValidationError{{Message: "SKILL.md must start with YAML frontmatter (---)"}}
	}

	yamlPart := strings.TrimSpace(parts[1])
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(yamlPart), &raw); err != nil {
		return []ValidationError{{Field: "frontmatter", Message: fmt.Sprintf("invalid YAML: %v", err)}}
	}

	dirName := filepath.Base(skillDir)

	// Validate name
	errors = append(errors, validateName(raw, dirName)...)

	// Validate description
	errors = append(errors, validateDescription(raw)...)

	// Validate optional fields
	errors = append(errors, validateCompatibility(raw)...)
	errors = append(errors, validateMetadataField(raw)...)
	errors = append(errors, validateAllowedFields(raw)...)

	return errors
}

func validateName(raw map[string]any, dirName string) []ValidationError {
	var errors []ValidationError

	nameAny, ok := raw["name"]
	if !ok {
		return []ValidationError{{Field: "name", Message: "missing required field"}}
	}

	name, ok := nameAny.(string)
	if !ok || strings.TrimSpace(name) == "" {
		return []ValidationError{{Field: "name", Message: "must be a non-empty string"}}
	}

	name = strings.TrimSpace(name)

	// Length check
	if len(name) > 64 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: fmt.Sprintf("exceeds 64 character limit (%d chars)", len(name)),
		})
	}

	// Lowercase check
	if name != strings.ToLower(name) {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "must be lowercase",
		})
	}

	// Leading/trailing hyphen check
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "cannot start or end with a hyphen",
		})
	}

	// Consecutive hyphen check
	if strings.Contains(name, "--") {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "cannot contain consecutive hyphens",
		})
	}

	// Character validation
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' {
			errors = append(errors, ValidationError{
				Field:   "name",
				Message: "can only contain letters, digits, and hyphens",
			})
			break
		}
	}

	// Directory match check
	if name != dirName {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: fmt.Sprintf("must match directory name '%s'", dirName),
		})
	}

	return errors
}

func validateDescription(raw map[string]any) []ValidationError {
	descAny, ok := raw["description"]
	if !ok {
		return []ValidationError{{Field: "description", Message: "missing required field"}}
	}

	desc, ok := descAny.(string)
	if !ok || strings.TrimSpace(desc) == "" {
		return []ValidationError{{Field: "description", Message: "must be a non-empty string"}}
	}

	desc = strings.TrimSpace(desc)
	if len(desc) > 1024 {
		return []ValidationError{{
			Field:   "description",
			Message: fmt.Sprintf("exceeds 1024 character limit (%d chars)", len(desc)),
		}}
	}

	return nil
}

func validateCompatibility(raw map[string]any) []ValidationError {
	compatAny, ok := raw["compatibility"]
	if !ok {
		return nil
	}

	compat, ok := compatAny.(string)
	if !ok {
		return []ValidationError{{Field: "compatibility", Message: "must be a string"}}
	}

	if len(compat) > 500 {
		return []ValidationError{{
			Field:   "compatibility",
			Message: fmt.Sprintf("exceeds 500 character limit (%d chars)", len(compat)),
		}}
	}

	return nil
}

func validateMetadataField(raw map[string]any) []ValidationError {
	metaAny, ok := raw["metadata"]
	if !ok {
		return nil
	}

	_, ok = metaAny.(map[string]any)
	if !ok {
		return []ValidationError{{Field: "metadata", Message: "must be a key-value mapping"}}
	}

	return nil
}

var allowedFields = map[string]bool{
	"name":          true,
	"description":   true,
	"license":       true,
	"compatibility": true,
	"metadata":      true,
	"allowed-tools": true,
}

func validateAllowedFields(raw map[string]any) []ValidationError {
	var errors []ValidationError

	for field := range raw {
		if !allowedFields[field] {
			errors = append(errors, ValidationError{
				Field:   field,
				Message: "unexpected field in frontmatter",
			})
		}
	}

	return errors
}
