package capabilities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// InferencePrompt is the prompt used to infer capabilities from an agent definition.
const InferencePrompt = `Analyze the following agent definition and infer its capabilities.

SYSTEM PROMPT:
{{ .SystemPrompt }}

{{ if .SkillNames }}
INSTALLED SKILLS:
{{ range $i, $name := .SkillNames }}
- {{ $name }}{{ if $.SkillContents }}{{ with index $.SkillContents $i }}
  {{ . }}
{{ end }}{{ end }}
{{ end }}
{{ end }}

{{ if .SchemaJSON }}
SCHEMA:
{{ .SchemaJSON }}
{{ end }}

Analyze this agent and return a JSON array of capabilities. Each capability should have:
- "name": A short kebab-case identifier (e.g., "code-review", "summarization")
- "description": A brief explanation of what this capability entails
- "confidence": A number between 0.0 and 1.0 indicating how confident you are
- "source": Where you inferred this from ("system_prompt", "skill", or "schema")

Guidelines:
1. Primary purpose should have highest confidence (0.9-1.0)
2. Secondary abilities should have medium confidence (0.6-0.8)
3. Implied abilities from skills depend on skill relevance (0.4-0.7)
4. Do NOT infer capabilities the agent explicitly denies or restricts
5. Be specific - "python-debugging" is better than "debugging"
6. Limit to 5-7 most relevant capabilities

Return ONLY the JSON array, no other text:
`

// Inferrer infers capabilities from agent definitions.
type Inferrer struct {
	// InvokeFn is the function used to invoke the LLM.
	// It takes a prompt and returns the response.
	InvokeFn func(ctx context.Context, prompt string) (string, error)

	// Model is the model name to use (for recording).
	Model string
}

// NewInferrer creates a new capability inferrer.
func NewInferrer(invokeFn func(ctx context.Context, prompt string) (string, error), model string) *Inferrer {
	return &Inferrer{
		InvokeFn: invokeFn,
		Model:    model,
	}
}

// Infer analyzes an agent definition and returns inferred capabilities.
func (i *Inferrer) Infer(ctx context.Context, input InferenceInput) (*InferenceResult, error) {
	// Build the prompt
	prompt := buildPrompt(input)

	// Invoke the LLM
	response, err := i.InvokeFn(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("invoke LLM: %w", err)
	}

	// Parse the response
	capabilities, err := parseCapabilities(response)
	if err != nil {
		return nil, fmt.Errorf("parse capabilities: %w", err)
	}

	return &InferenceResult{
		Capabilities: capabilities,
		InputHash:    input.Hash(),
		ModelUsed:    i.Model,
	}, nil
}

// buildPrompt builds the inference prompt from input.
func buildPrompt(input InferenceInput) string {
	var sb strings.Builder

	sb.WriteString("Analyze the following agent definition and infer its capabilities.\n\n")

	sb.WriteString("SYSTEM PROMPT:\n")
	sb.WriteString(input.SystemPrompt)
	sb.WriteString("\n\n")

	if len(input.SkillNames) > 0 {
		sb.WriteString("INSTALLED SKILLS:\n")
		for j, name := range input.SkillNames {
			sb.WriteString(fmt.Sprintf("- %s", name))
			if j < len(input.SkillContents) && input.SkillContents[j] != "" {
				// Include first 200 chars of skill content
				content := input.SkillContents[j]
				if len(content) > 200 {
					content = content[:200] + "..."
				}
				sb.WriteString(fmt.Sprintf("\n  %s", content))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if input.SchemaJSON != "" {
		sb.WriteString("SCHEMA:\n")
		sb.WriteString(input.SchemaJSON)
		sb.WriteString("\n\n")
	}

	sb.WriteString(`Analyze this agent and return a JSON array of capabilities. Each capability should have:
- "name": A short kebab-case identifier (e.g., "code-review", "summarization")
- "description": A brief explanation of what this capability entails
- "confidence": A number between 0.0 and 1.0 indicating how confident you are
- "source": Where you inferred this from ("system_prompt", "skill", or "schema")

Guidelines:
1. Primary purpose should have highest confidence (0.9-1.0)
2. Secondary abilities should have medium confidence (0.6-0.8)
3. Implied abilities from skills depend on skill relevance (0.4-0.7)
4. Do NOT infer capabilities the agent explicitly denies or restricts
5. Be specific - "python-debugging" is better than "debugging"
6. Limit to 5-7 most relevant capabilities

Return ONLY the JSON array, no other text:
`)

	return sb.String()
}

// parseCapabilities parses the LLM response into structured capabilities.
func parseCapabilities(response string) ([]Capability, error) {
	// Clean the response - extract JSON array if wrapped in markdown
	cleaned := strings.TrimSpace(response)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		if idx := strings.Index(cleaned, "```"); idx > 0 {
			cleaned = cleaned[:idx]
		}
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
		if idx := strings.Index(cleaned, "```"); idx > 0 {
			cleaned = cleaned[:idx]
		}
	}
	cleaned = strings.TrimSpace(cleaned)

	// Parse JSON
	var capabilities []struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Confidence  float64 `json:"confidence"`
		Source      string  `json:"source"`
	}

	if err := json.Unmarshal([]byte(cleaned), &capabilities); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Convert to Capability type
	result := make([]Capability, len(capabilities))
	for j, c := range capabilities {
		// Normalize confidence
		conf := c.Confidence
		if conf < 0 {
			conf = 0
		}
		if conf > 1 {
			conf = 1
		}

		result[j] = Capability{
			Name:        c.Name,
			Description: c.Description,
			Confidence:  conf,
		}
	}

	return result, nil
}

// InferWithoutLLM provides a simple heuristic-based inference for testing
// or when LLM is not available.
func InferWithoutLLM(input InferenceInput) *InferenceResult {
	var capabilities []Capability

	// Analyze system prompt for keywords
	prompt := strings.ToLower(input.SystemPrompt)

	keywordCapabilities := map[string][]string{
		"code":       {"code-analysis", "programming"},
		"review":     {"code-review", "feedback"},
		"debug":      {"debugging", "error-analysis"},
		"test":       {"testing", "test-generation"},
		"document":   {"documentation", "technical-writing"},
		"summarize":  {"summarization", "content-extraction"},
		"research":   {"research", "information-gathering"},
		"translate":  {"translation", "language-processing"},
		"write":      {"writing", "content-creation"},
		"analyze":    {"analysis", "evaluation"},
		"security":   {"security-analysis", "vulnerability-detection"},
		"data":       {"data-analysis", "data-processing"},
		"api":        {"api-development", "integration"},
		"sql":        {"database", "sql-queries"},
		"python":     {"python-development"},
		"javascript": {"javascript-development"},
		"go":         {"go-development"},
	}

	seen := make(map[string]bool)
	for keyword, caps := range keywordCapabilities {
		if strings.Contains(prompt, keyword) {
			for _, cap := range caps {
				if !seen[cap] {
					capabilities = append(capabilities, Capability{
						Name:        cap,
						Description: fmt.Sprintf("Inferred from keyword '%s' in system prompt", keyword),
						Confidence:  0.6,
					})
					seen[cap] = true
				}
			}
		}
	}

	// Add capabilities from skills
	for _, skill := range input.SkillNames {
		skillLower := strings.ToLower(skill)
		if !seen[skillLower] {
			capabilities = append(capabilities, Capability{
				Name:        skillLower,
				Description: fmt.Sprintf("Installed skill: %s", skill),
				Confidence:  0.7,
			})
			seen[skillLower] = true
		}
	}

	// Default capability if nothing else found
	if len(capabilities) == 0 {
		capabilities = append(capabilities, Capability{
			Name:        "general-assistance",
			Description: "General purpose assistant",
			Confidence:  0.5,
		})
	}

	return &InferenceResult{
		Capabilities: capabilities,
		InputHash:    input.Hash(),
		ModelUsed:    "heuristic",
	}
}
