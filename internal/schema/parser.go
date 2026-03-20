package schema

import (
	"encoding/json"
	"fmt"
)

type ParsedSchema struct {
	Type       string                     `json:"type"`
	Properties map[string]Property        `json:"properties"`
	Required   []string                   `json:"required"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Default     any      `json:"default"`
	Enum        []string `json:"enum"`

	// New CLI hints
	Flag string `json:"flag"` // Custom flag name
	File bool   `json:"file"` // Load file contents

	// Deprecated CLI extensions (supported during migration)
	CLIPosition int    `json:"x-cli-position"`
	CLIFlag     string `json:"x-cli-flag"`
	CLIShort    string `json:"x-cli-short"`
	CLIFile     bool   `json:"x-cli-file"`
}

type FlagDef struct {
	Name         string // CLI flag name (e.g., "source-language")
	PropertyName string // Original property name for variable (e.g., "from")
	ShortName    string
	Type         string
	DefaultValue any
	Description  string
	Position     int
	IsFile       bool
	Required     bool
}

func ParseSchema(data []byte) (*ParsedSchema, error) {
	var schema ParsedSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parsing JSON schema: %w", err)
	}
	return &schema, nil
}

func GenerateFlags(schema *ParsedSchema) []FlagDef {
	var flags []FlagDef
	requiredSet := make(map[string]bool)
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	for name, prop := range schema.Properties {
		// Only generate flags for primitive types
		if !isPrimitiveType(prop.Type) {
			continue
		}

		flag := FlagDef{
			Name:         name,
			PropertyName: name,
			Type:         prop.Type,
			DefaultValue: prop.Default,
			Description:  prop.Description,
			IsFile:       prop.File || prop.CLIFile, // Prefer new File field
			Required:     requiredSet[name],
		}

		// Prefer new Flag field over deprecated CLIFlag
		if prop.Flag != "" {
			flag.Name = prop.Flag
		} else if prop.CLIFlag != "" {
			flag.Name = prop.CLIFlag
		} else {
			flag.Name = toKebab(name)
		}

		flags = append(flags, flag)
	}

	return flags
}

func isPrimitiveType(t string) bool {
	switch t {
	case "string", "integer", "number", "boolean":
		return true
	default:
		return false
	}
}

func toKebab(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-', r+32) // Convert to lowercase
		} else if r >= 'A' && r <= 'Z' {
			result = append(result, r+32)
		} else if r == '_' {
			result = append(result, '-')
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
