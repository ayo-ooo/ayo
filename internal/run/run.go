package run

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
	uipkg "github.com/alexcabrera/ayo/internal/ui"
)

// Runner executes agents using Fantasy's Agent abstraction.
type Runner struct {
	config   config.Config
	debug    bool
	depth    int // 0 = top-level, 1+ = sub-agent calls
	sessions map[string]*Session
}

// Session maintains conversation state for interactive chat.
type Session struct {
	Agent    agent.Agent
	Messages []fantasy.Message
}

const maxToolIterations = 8
const maxOutputCastRetries = 3

// NewRunnerFromConfig creates a new runner from the given configuration.
func NewRunnerFromConfig(cfg config.Config, debug bool) (*Runner, error) {
	return &Runner{
		config:   cfg,
		debug:    debug,
		sessions: make(map[string]*Session),
	}, nil
}



// Chat sends a message in an interactive session, maintaining conversation history.
func (r *Runner) Chat(ctx context.Context, ag agent.Agent, input string) (string, error) {
	session, ok := r.sessions[ag.Handle]
	if !ok {
		// Initialize new session with system messages
		var msgs []fantasy.Message
		if strings.TrimSpace(ag.CombinedSystem) != "" {
			msgs = append(msgs, fantasy.NewSystemMessage(ag.CombinedSystem))
		}
		if strings.TrimSpace(ag.ToolsPrompt) != "" {
			msgs = append(msgs, fantasy.NewSystemMessage(ag.ToolsPrompt))
		}
		if strings.TrimSpace(ag.SkillsPrompt) != "" {
			msgs = append(msgs, fantasy.NewSystemMessage(ag.SkillsPrompt))
		}
		session = &Session{Agent: ag, Messages: msgs}
		r.sessions[ag.Handle] = session
	}

	// Add user message
	session.Messages = append(session.Messages, fantasy.NewUserMessage(input))

	// Run the chat and get response
	resp, newMsgs, err := r.runChatWithHistory(ctx, ag, session.Messages)
	if err != nil {
		// Remove the failed user message
		session.Messages = session.Messages[:len(session.Messages)-1]
		return "", err
	}

	// Update session with full message history
	session.Messages = newMsgs
	return resp, nil
}

// Text runs a single prompt without maintaining history.
func (r *Runner) Text(ctx context.Context, ag agent.Agent, prompt string, attachments []string) (string, error) {
	msgs := r.buildMessagesWithAttachments(ag, prompt, attachments)
	return r.runChat(ctx, ag, msgs)
}

func (r *Runner) buildMessages(ag agent.Agent, prompt string) []fantasy.Message {
	return r.buildMessagesWithAttachments(ag, prompt, nil)
}

func (r *Runner) buildMessagesWithAttachments(ag agent.Agent, prompt string, attachments []string) []fantasy.Message {
	var msgs []fantasy.Message
	if strings.TrimSpace(ag.CombinedSystem) != "" {
		msgs = append(msgs, fantasy.NewSystemMessage(ag.CombinedSystem))
	}
	if strings.TrimSpace(ag.ToolsPrompt) != "" {
		msgs = append(msgs, fantasy.NewSystemMessage(ag.ToolsPrompt))
	}
	if strings.TrimSpace(ag.SkillsPrompt) != "" {
		msgs = append(msgs, fantasy.NewSystemMessage(ag.SkillsPrompt))
	}

	// Build file parts from attachments
	// Text files are inlined into the prompt; binary files use FilePart
	var fileParts []fantasy.FilePart
	var textAttachments []string

	for _, path := range attachments {
		data, err := os.ReadFile(path)
		if err != nil {
			// Skip files that can't be read, but include error in prompt
			prompt = fmt.Sprintf("%s\n\n[Error reading %s: %v]", prompt, path, err)
			continue
		}

		// Determine media type from extension
		ext := filepath.Ext(path)
		mediaType := mime.TypeByExtension(ext)
		if mediaType == "" {
			mediaType = "text/plain" // Default for unknown types
		}

		// Text files: inline into prompt (providers don't handle text FileParts well)
		// Binary files (images, PDFs, audio): use FilePart
		if isTextMediaType(mediaType) {
			textAttachments = append(textAttachments, fmt.Sprintf("<file path=%q>\n%s\n</file>", filepath.Base(path), string(data)))
		} else {
			fileParts = append(fileParts, fantasy.FilePart{
				Filename:  filepath.Base(path),
				Data:      data,
				MediaType: mediaType,
			})
		}
	}

	// Prepend text attachments to prompt
	if len(textAttachments) > 0 {
		prompt = strings.Join(textAttachments, "\n\n") + "\n\n" + prompt
	}

	msgs = append(msgs, fantasy.NewUserMessage(prompt, fileParts...))
	return msgs
}

// isTextMediaType returns true if the media type represents text content
// that should be inlined rather than sent as a binary file part.
func isTextMediaType(mediaType string) bool {
	// Check for explicit text types
	if strings.HasPrefix(mediaType, "text/") {
		return true
	}
	// Common text-based types that don't start with text/
	textTypes := []string{
		"application/json",
		"application/xml",
		"application/javascript",
		"application/typescript",
		"application/x-yaml",
		"application/yaml",
		"application/toml",
		"application/x-sh",
		"application/x-shellscript",
	}
	// Strip parameters (e.g., "text/plain; charset=utf-8" -> "text/plain")
	baseType := strings.Split(mediaType, ";")[0]
	for _, t := range textTypes {
		if baseType == t {
			return true
		}
	}
	return false
}

func (r *Runner) runChat(ctx context.Context, ag agent.Agent, msgs []fantasy.Message) (string, error) {
	resp, _, err := r.runChatWithHistory(ctx, ag, msgs)
	return resp, err
}

func (r *Runner) runChatWithHistory(ctx context.Context, ag agent.Agent, msgs []fantasy.Message) (string, []fantasy.Message, error) {
	if strings.TrimSpace(ag.Model) == "" {
		return "", nil, fmt.Errorf("model is required")
	}

	// Extract the last user message as the prompt for Fantasy
	// Fantasy requires a non-empty Prompt field
	var prompt string
	var historyMsgs []fantasy.Message
	if len(msgs) > 0 && msgs[len(msgs)-1].Role == fantasy.MessageRoleUser {
		// Get the text content from the last user message
		for _, part := range msgs[len(msgs)-1].Content {
			if tp, ok := part.(fantasy.TextPart); ok {
				prompt = tp.Text
				break
			}
		}
		// History is everything except the last user message
		historyMsgs = msgs[:len(msgs)-1]
	} else {
		historyMsgs = msgs
	}

	// Create language model from config
	model, err := NewLanguageModel(ctx, r.config.Provider, ag.Model)
	if err != nil {
		return "", nil, fmt.Errorf("create language model: %w", err)
	}

	// Build tool set
	tools := NewFantasyToolSet(ag.Config.AllowedTools)

	// Add agent_call if explicitly allowed in config (for any agent)
	// or if it's a non-builtin agent (user agents get it by default)
	if tools.HasTool("agent_call") || !ag.BuiltIn {
		tools.AddAgentCallTool(r.agentCallExecutor())
	}

	// Create Fantasy agent
	fantasyAgent := fantasy.NewAgent(
		model,
		fantasy.WithSystemPrompt(""), // System prompt already in messages
		fantasy.WithTools(tools.Tools()...),
	)

	ui := uipkg.NewWithDepth(r.debug, r.depth)
	var content strings.Builder
	var reasoningStarted bool
	var textStarted bool

	// Start spinner - shows "thinking..." until activity starts
	spinner := uipkg.NewSpinnerWithDepth("thinking...", r.depth)
	spinner.Start()
	spinnerActive := true

	// Track current tool call
	var currentTool uipkg.ToolCallInfo
	var toolStartTime time.Time

	// Stream the response with all callbacks
	result, err := fantasyAgent.Stream(ctx, fantasy.AgentStreamCall{
		Prompt:   prompt,
		Messages: historyMsgs,

		// Reasoning streams (for models like Claude that expose thinking)
		OnReasoningDelta: func(id, text string) error {
			if spinnerActive {
				spinner.Stop()
				spinnerActive = false
			}
			if !reasoningStarted {
				reasoningStarted = true
				ui.PrintReasoningStart()
			}
			ui.PrintReasoningDelta(text)
			return nil
		},

		OnReasoningEnd: func(id string, reasoning fantasy.ReasoningContent) error {
			if reasoningStarted {
				ui.PrintReasoningEnd()
				reasoningStarted = false
			}
			return nil
		},

		// Tool call complete - show the command we're about to run
		OnToolCall: func(tc fantasy.ToolCallContent) error {
			if spinnerActive {
				spinner.Stop()
				spinnerActive = false
			}
			toolStartTime = time.Now()
			currentTool = uipkg.ToolCallInfo{
				Name:  tc.ToolName,
				Input: tc.Input,
			}
			// Extract command and description for bash
			if tc.ToolName == "bash" {
				var params struct {
					Command     string `json:"command"`
					Description string `json:"description"`
				}
				if err := json.Unmarshal([]byte(tc.Input), &params); err == nil {
					currentTool.Command = params.Command
					currentTool.Description = params.Description
				}
			}
			ui.PrintToolCallStart(currentTool)
			return nil
		},

		// Tool result - show the output
		OnToolResult: func(result fantasy.ToolResultContent) error {
			currentTool.Duration = formatElapsed(time.Since(toolStartTime))
			currentTool.Output = formatToolResultContent(result)
			if result.Result.GetType() == fantasy.ToolResultContentTypeError {
				currentTool.Error = currentTool.Output
			}
			ui.PrintToolCallResult(currentTool)
			currentTool = uipkg.ToolCallInfo{}
			return nil
		},

		// Text response streams
		OnTextDelta: func(id, text string) error {
			if spinnerActive {
				spinner.Stop()
				spinnerActive = false
			}
			if !textStarted {
				textStarted = true
				ui.PrintAgentResponseHeader(ag.Handle)
			}
			content.WriteString(text)
			ui.PrintTextDelta(text)
			return nil
		},
	})

	// Ensure spinner is stopped
	if spinnerActive {
		if err != nil {
			spinner.StopWithError("Failed")
		} else {
			spinner.Stop()
		}
	}

	// End text streaming with a newline
	if textStarted {
		ui.PrintTextEnd()
	}

	if err != nil {
		return "", nil, err
	}

	// Get final content
	finalContent := content.String()
	if finalContent == "" && len(result.Steps) > 0 {
		// Try to get content from the last step
		lastStep := result.Steps[len(result.Steps)-1]
		for _, part := range lastStep.Content {
			if textPart, ok := part.(fantasy.TextPart); ok {
				finalContent = textPart.Text
				break
			}
		}
	}

	// Cast to structured output if agent has output schema
	if ag.HasOutputSchema() {
		structuredOutput, err := r.castToStructuredOutput(ctx, model, ag, finalContent, ui)
		if err != nil {
			return "", nil, fmt.Errorf("structured output: %w", err)
		}
		finalContent = structuredOutput

		// Append assistant message to history
		msgs = append(msgs, fantasy.Message{
			Role:    fantasy.MessageRoleAssistant,
			Content: []fantasy.MessagePart{fantasy.TextPart{Text: finalContent}},
		})

		// When piped, return raw JSON for downstream consumption
		if ui.IsPiped() {
			return strings.TrimSpace(finalContent), msgs, nil
		}

		// Render JSON with syntax highlighting for terminal
		rendered := ui.RenderJSON(finalContent)
		return strings.TrimSpace(rendered), msgs, nil
	}

	// Append assistant message to history
	msgs = append(msgs, fantasy.Message{
		Role:    fantasy.MessageRoleAssistant,
		Content: []fantasy.MessagePart{fantasy.TextPart{Text: finalContent}},
	})

	// When piped with no output schema, return raw content
	if ui.IsPiped() {
		return strings.TrimSpace(finalContent), msgs, nil
	}

	// Text was already streamed to output, return empty to avoid duplicate
	return "", msgs, nil
}

func (r *Runner) agentCallExecutor() func(ctx context.Context, params AgentCallParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	return func(ctx context.Context, params AgentCallParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
		// Normalize handle
		agentHandle := agent.NormalizeHandle(params.Agent)

		// Only allow calling builtin agents
		if !agent.IsReservedNamespace(agentHandle) {
			return fantasy.NewTextErrorResponse("agent_call can only invoke builtin agents (prefixed with 'ayo.')"), nil
		}

		// Load the target agent
		targetAgent, err := agent.Load(r.config, agentHandle)
		if err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to load agent %s: %v", agentHandle, err)), nil
		}

		// Configure timeout
		timeout := 120 * time.Second
		if params.TimeoutSeconds > 0 {
			timeout = time.Duration(params.TimeoutSeconds) * time.Second
		}
		if timeout > 300*time.Second {
			timeout = 300 * time.Second
		}

		execCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Show sub-agent start
		ui := uipkg.NewWithDepth(r.debug, r.depth)
		startTime := time.Now()
		ui.PrintSubAgentStart(agentHandle, params.Prompt)

		// Create sub-runner at increased depth
		subRunner := &Runner{
			config:   r.config,
			debug:    r.debug,
			depth:    r.depth + 1,
			sessions: make(map[string]*Session),
		}

		// Run the agent
		response, err := subRunner.Text(execCtx, targetAgent, params.Prompt, nil)

		// Show sub-agent completion
		duration := formatElapsed(time.Since(startTime))
		hasError := err != nil
		if execCtx.Err() == context.DeadlineExceeded {
			hasError = true
		}
		ui.PrintSubAgentEnd(agentHandle, duration, hasError)

		if err != nil {
			if execCtx.Err() == context.DeadlineExceeded {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("agent %s timed out after %v", agentHandle, timeout)), nil
			}
			return fantasy.NewTextErrorResponse(fmt.Sprintf("agent %s error: %v", agentHandle, err)), nil
		}

		// Truncate if too long
		const maxOutput = 128 * 1024
		if len(response) > maxOutput {
			response = response[:maxOutput]
		}

		return fantasy.NewTextResponse(strings.TrimSpace(response)), nil
	}
}

// formatToolResultContent converts a Fantasy tool result to a string for display.
func formatToolResultContent(result fantasy.ToolResultContent) string {
	switch result.Result.GetType() {
	case fantasy.ToolResultContentTypeText:
		if r, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](result.Result); ok {
			return r.Text
		}
	case fantasy.ToolResultContentTypeError:
		if r, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentError](result.Result); ok {
			return r.Error.Error()
		}
	case fantasy.ToolResultContentTypeMedia:
		if r, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentMedia](result.Result); ok {
			if r.Text != "" {
				return r.Text
			}
			return fmt.Sprintf("[media: %s]", r.MediaType)
		}
	}
	return ""
}

// formatElapsed formats a duration for display.
func formatElapsed(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// castToStructuredOutput takes the agent's response and casts it to the required output schema.
// It uses GenerateObject to produce structured output, then validates against the schema.
// If validation fails, it retries by providing error feedback to the model.
func (r *Runner) castToStructuredOutput(ctx context.Context, model fantasy.LanguageModel, ag agent.Agent, agentOutput string, ui *uipkg.UI) (string, error) {
	if ag.OutputSchema == nil {
		return agentOutput, nil
	}

	var lastError error
	for attempt := 0; attempt < maxOutputCastRetries; attempt++ {
		// Build prompt for structured output casting
		var prompt fantasy.Prompt
		if attempt == 0 {
			prompt = fantasy.Prompt{
				fantasy.NewSystemMessage("You are a data extraction assistant. Extract and format the information from the provided content into the required JSON structure. Output only valid JSON matching the schema."),
				fantasy.NewUserMessage(fmt.Sprintf("Extract and format the following content into the required JSON structure:\n\n%s", agentOutput)),
			}
		} else {
			// Retry with error feedback
			prompt = fantasy.Prompt{
				fantasy.NewSystemMessage("You are a data extraction assistant. Extract and format the information from the provided content into the required JSON structure. Output only valid JSON matching the schema."),
				fantasy.NewUserMessage(fmt.Sprintf("Extract and format the following content into the required JSON structure:\n\n%s\n\nPrevious attempt failed validation with error: %v\n\nPlease fix the output to match the schema requirements.", agentOutput, lastError)),
			}
		}

		// Show spinner for casting
		var spinner *uipkg.Spinner
		if attempt == 0 {
			spinner = uipkg.NewSpinnerWithDepth("formatting output...", r.depth)
		} else {
			spinner = uipkg.NewSpinnerWithDepth(fmt.Sprintf("reformatting output (attempt %d/%d)...", attempt+1, maxOutputCastRetries), r.depth)
		}
		spinner.Start()

		response, err := model.GenerateObject(ctx, fantasy.ObjectCall{
			Prompt:            prompt,
			Schema:            *ag.OutputSchema,
			SchemaName:        "Output",
			SchemaDescription: "Required output format for the agent response",
		})

		spinner.Stop()

		if err != nil {
			lastError = err
			continue
		}

		// Convert object back to JSON string with pretty formatting
		jsonBytes, err := json.MarshalIndent(response.Object, "", "  ")
		if err != nil {
			lastError = fmt.Errorf("failed to marshal structured output: %w", err)
			continue
		}
		jsonOutput := string(jsonBytes)

		// Validate against schema
		if err := ag.ValidateOutput(jsonOutput); err != nil {
			lastError = err
			continue
		}

		// Success - return the structured output
		return jsonOutput, nil
	}

	return "", fmt.Errorf("failed to produce valid structured output after %d attempts: %w", maxOutputCastRetries, lastError)
}
