package shared

import (
	"encoding/json"
)

// AgentCallOutput represents the structured output from agent_call tool.
type AgentCallOutput struct {
	Agent    string `json:"agent"`
	Prompt   string `json:"prompt"`
	Response string `json:"response"`
	Duration string `json:"duration"`
}

// ParseAgentCallOutput parses agent_call output JSON from RawOutput.
func ParseAgentCallOutput(raw string) (*AgentCallOutput, error) {
	if raw == "" {
		return nil, nil
	}
	var out AgentCallOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AgentCallRenderer renders agent_call tool output with threaded conversation.
type AgentCallRenderer struct{}

// Name returns the tool name.
func (r *AgentCallRenderer) Name() string { return "agent_call" }

// Render produces the render output for agent_call.
func (r *AgentCallRenderer) Render(input ToolRenderInput) ToolRenderOutput {
	output := ToolRenderOutput{
		Label: "Agent",
		State: input.State,
	}

	// Parse input params for header (always available)
	var params struct {
		Agent  string `json:"agent"`
		Prompt string `json:"prompt"`
	}
	if err := json.Unmarshal([]byte(input.RawInput), &params); err == nil && params.Agent != "" {
		output.Label = params.Agent
		output.HeaderParams = []string{truncateString(params.Prompt, 60)}
	}

	// Try to parse output JSON (like bash does)
	if input.State == ToolStateSuccess || input.State == ToolStateError {
		parsed, err := ParseAgentCallOutput(input.RawOutput)
		if err == nil && parsed != nil {
			// Update label from parsed output (in case it differs)
			output.Label = parsed.Agent

			// Add the response as markdown section
			if parsed.Response != "" {
				output.Sections = append(output.Sections, RenderSection{
					Type:     SectionMarkdown,
					Content:  parsed.Response,
					MaxLines: 20,
				})
			}
		} else if input.RawOutput != "" {
			// Fallback: show raw output as markdown
			output.Sections = append(output.Sections, RenderSection{
				Type:     SectionMarkdown,
				Content:  input.RawOutput,
				MaxLines: 20,
			})
		}
	}

	return output
}

// truncateString truncates a string to maxLen, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func init() {
	RegisterToolRenderer(&AgentCallRenderer{})
}
