package shared

import (
	"encoding/json"
	"fmt"
	"os"
)

// debugLogShared writes to a debug file for troubleshooting
func debugLogShared(format string, args ...any) {
	f, err := os.OpenFile("/tmp/ayo_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[shared] "+format+"\n", args...)
}

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
	debugLogShared("=== AgentCallRenderer.Render ===")
	debugLogShared("  input.Name: %s", input.Name)
	debugLogShared("  input.State: %d", input.State)
	debugLogShared("  input.RawInput: %s", input.RawInput)
	debugLogShared("  input.RawOutput (first 200): %s", truncateString(input.RawOutput, 200))
	debugLogShared("  input.RawMetadata: %s", input.RawMetadata)

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
		debugLogShared("  parsed params: agent=%s prompt=%s", params.Agent, truncateString(params.Prompt, 50))
	} else {
		debugLogShared("  failed to parse params: %v", err)
	}

	// Try to parse output JSON (like bash does)
	if input.State == ToolStateSuccess || input.State == ToolStateError {
		debugLogShared("  state is success/error, parsing output...")
		parsed, err := ParseAgentCallOutput(input.RawOutput)
		if err == nil && parsed != nil {
			debugLogShared("  parsed output: agent=%s response(first 100)=%s", parsed.Agent, truncateString(parsed.Response, 100))
			// Update label from parsed output (in case it differs)
			output.Label = parsed.Agent

			// Add the response as markdown section
			if parsed.Response != "" {
				output.Sections = append(output.Sections, RenderSection{
					Type:     SectionMarkdown,
					Content:  parsed.Response,
					MaxLines: 20,
				})
				debugLogShared("  added markdown section")
			}
		} else if input.RawOutput != "" {
			debugLogShared("  parse failed (err=%v), using raw output as fallback", err)
			// Fallback: show raw output as markdown
			output.Sections = append(output.Sections, RenderSection{
				Type:     SectionMarkdown,
				Content:  input.RawOutput,
				MaxLines: 20,
			})
		} else {
			debugLogShared("  no output to parse")
		}
	} else {
		debugLogShared("  state is not success/error, skipping output parse")
	}

	debugLogShared("  final output: Label=%s HeaderParams=%v Sections=%d", output.Label, output.HeaderParams, len(output.Sections))
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
