package project

import "time"

type HookType string

const (
	HookAgentStart      HookType = "agent-start"
	HookAgentFinish     HookType = "agent-finish"
	HookAgentError      HookType = "agent-error"
	HookStepStart       HookType = "step-start"
	HookStepFinish      HookType = "step-finish"
	HookTextStart       HookType = "text-start"
	HookTextDelta       HookType = "text-delta"
	HookTextEnd         HookType = "text-end"
	HookReasoningStart  HookType = "reasoning-start"
	HookReasoningDelta  HookType = "reasoning-delta"
	HookReasoningEnd    HookType = "reasoning-end"
	HookToolInputStart  HookType = "tool-input-start"
	HookToolInputDelta  HookType = "tool-input-delta"
	HookToolInputEnd    HookType = "tool-input-end"
	HookToolCall        HookType = "tool-call"
	HookToolResult      HookType = "tool-result"
	HookSource          HookType = "source"
	HookStreamFinish    HookType = "stream-finish"
	HookWarnings        HookType = "warnings"
)

type Project struct {
	Path    string
	Config  AgentConfig
	System  string
	Prompt  *string
	Input   *Schema
	Output  *Schema
	Skills  []Skill
	Hooks   map[HookType]string
}

type AgentConfig struct {
	Name        string
	Version     string
	Description string
	Model       ModelRequirements
	Defaults    AgentDefaults
}

type ModelRequirements struct {
	RequiresStructuredOutput bool     `toml:"requires_structured_output"`
	RequiresTools            bool     `toml:"requires_tools"`
	RequiresVision           bool     `toml:"requires_vision"`
	Suggested                []string `toml:"suggested"`
	Default                  string   `toml:"default"`
}

type AgentDefaults struct {
	Temperature float64 `toml:"temperature"`
	MaxTokens   int     `toml:"max_tokens"`
}

type Schema struct {
	Content   []byte
	Parsed    any
}

type Skill struct {
	Name        string
	Path        string
	Description string
}

type Hook struct {
	Type     HookType
	Path     string
	Exec     []byte
}

type UserConfig struct {
	Provider string            `toml:"provider"`
	Model    string            `toml:"model"`
	APIKey   string            `toml:"api_key"`
	Hooks    map[HookType]string `toml:"hooks"`
}

type ValidationError struct {
	File    string
	Message string
	Line    int
}

func (e *ValidationError) Error() string {
	if e.Line > 0 {
		return e.File + ":" + string(rune(e.Line)) + ": " + e.Message
	}
	return e.File + ": " + e.Message
}

type HookPayload struct {
	Event     string                 `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]any `json:"data"`
}
