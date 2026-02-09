// Package capabilities provides capability inference for agents.
//
// Capabilities are inferred by analyzing an agent's system prompt, skills,
// and schemas using an LLM. This enables @ayo to intelligently route tasks
// to appropriate agents.
package capabilities

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Capability represents an inferred capability of an agent.
type Capability struct {
	// Name is a short kebab-case identifier (e.g., "code-review", "summarization").
	Name string `json:"name"`

	// Description explains what this capability entails.
	Description string `json:"description"`

	// Confidence is how confident we are in this capability (0.0-1.0).
	Confidence float64 `json:"confidence"`
}

// InferenceInput contains all the information needed to infer capabilities.
type InferenceInput struct {
	// SystemPrompt is the agent's system prompt.
	SystemPrompt string

	// SkillNames is the list of installed skill names.
	SkillNames []string

	// SkillContents is the SKILL.md content for each skill.
	SkillContents []string

	// SchemaJSON is the input/output JSON schema if present.
	SchemaJSON string
}

// InferenceResult contains the results of capability inference.
type InferenceResult struct {
	// Capabilities is the list of inferred capabilities.
	Capabilities []Capability `json:"capabilities"`

	// InputHash is the hash of the inference input for cache invalidation.
	InputHash string `json:"input_hash"`

	// ModelUsed is the model that performed the inference.
	ModelUsed string `json:"model_used"`
}

// Hash returns a stable hash of the inference input for cache invalidation.
func (i *InferenceInput) Hash() string {
	// Combine all inputs into a single string
	var parts []string
	parts = append(parts, i.SystemPrompt)
	parts = append(parts, strings.Join(i.SkillNames, "|"))
	parts = append(parts, strings.Join(i.SkillContents, "|"))
	parts = append(parts, i.SchemaJSON)

	combined := strings.Join(parts, "\n---\n")
	h := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(h[:])
}

// StoredCapability represents a capability stored in the database.
type StoredCapability struct {
	ID          string  `json:"id"`
	AgentID     string  `json:"agent_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Source      string  `json:"source"` // "system_prompt", "skill", "schema"
	Embedding   []byte  `json:"embedding,omitempty"`
	InputHash   string  `json:"input_hash"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`
}
