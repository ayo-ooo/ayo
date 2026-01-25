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
	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/plugins"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/smallmodel"
	uipkg "github.com/alexcabrera/ayo/internal/ui"
)

// Runner executes agents using Fantasy's Agent abstraction.
type Runner struct {
	config           config.Config
	debug            bool
	depth            int // 0 = top-level, 1+ = sub-agent calls
	sessions         map[string]*ChatSession
	services         *session.Services        // nil = no persistence
	memoryService    *memory.Service          // nil = no memory
	formationService *memory.FormationService // nil = no async formation
	smallModel       *smallmodel.Service      // nil = no small model for memory extraction
	onAsyncStatus    func(uipkg.AsyncStatusMsg) // nil = no async status callback
	memoryQueue      *memory.Queue            // nil = sync memory operations
}

// ChatSession maintains conversation state for interactive chat.
type ChatSession struct {
	Agent          agent.Agent
	Messages       []fantasy.Message
	SessionID      string // Database session ID (empty if no persistence)
	TitleGenerated bool   // Whether title generation has been triggered
}

const maxToolIterations = 8
const maxOutputCastRetries = 3

// NewRunnerFromConfig creates a new runner from the given configuration.
func NewRunnerFromConfig(cfg config.Config, debug bool) (*Runner, error) {
	return &Runner{
		config:   cfg,
		debug:    debug,
		sessions: make(map[string]*ChatSession),
	}, nil
}

// NewRunnerWithServices creates a runner with session persistence.
func NewRunnerWithServices(cfg config.Config, debug bool, services *session.Services) (*Runner, error) {
	return &Runner{
		config:   cfg,
		debug:    debug,
		sessions: make(map[string]*ChatSession),
		services: services,
	}, nil
}

// RunnerOptions provides optional configuration for the runner.
type RunnerOptions struct {
	Services         *session.Services
	MemoryService    *memory.Service
	FormationService *memory.FormationService
	SmallModel       *smallmodel.Service
	OnAsyncStatus    func(uipkg.AsyncStatusMsg) // Callback for async operation status updates
	MemoryQueue      *memory.Queue              // Queue for async memory operations
}

// NewRunner creates a runner with all options.
func NewRunner(cfg config.Config, debug bool, opts RunnerOptions) (*Runner, error) {
	return &Runner{
		config:           cfg,
		debug:            debug,
		sessions:         make(map[string]*ChatSession),
		services:         opts.Services,
		memoryService:    opts.MemoryService,
		formationService: opts.FormationService,
		smallModel:       opts.SmallModel,
		onAsyncStatus:    opts.OnAsyncStatus,
		memoryQueue:      opts.MemoryQueue,
	}, nil
}

// WaitForFormations waits for any pending memory formations to complete.
func (r *Runner) WaitForFormations(timeout time.Duration) {
	if r.formationService != nil {
		r.formationService.Wait(timeout)
	}
}

// WaitForMemoryQueue waits for pending memory queue operations to complete.
func (r *Runner) WaitForMemoryQueue(timeout time.Duration) {
	if r.memoryQueue != nil {
		r.memoryQueue.Stop(timeout)
	}
}

// StartMemoryQueue starts the memory queue worker if configured.
func (r *Runner) StartMemoryQueue() {
	if r.memoryQueue != nil {
		r.memoryQueue.Start()
	}
}


// Chat sends a message in an interactive session, maintaining conversation history.
func (r *Runner) Chat(ctx context.Context, ag agent.Agent, input string) (string, error) {
	chatSession, ok := r.sessions[ag.Handle]
	if !ok {
		// Initialize new session with system messages
		var msgs []fantasy.Message
		
		// Build combined system prompt with memory context
		systemPrompt := ag.CombinedSystem
		if r.memoryService != nil && ag.Config.Memory.Enabled {
			memCtx, err := agent.BuildMemoryContext(ctx, r.memoryService, ag.Handle, "", input, ag.Config.Memory)
			if err == nil && memCtx != nil {
				systemPrompt = agent.InjectMemoryContext(systemPrompt, memCtx)
			}
		}
		
		if strings.TrimSpace(systemPrompt) != "" {
			msgs = append(msgs, fantasy.NewSystemMessage(systemPrompt))
		}
		if strings.TrimSpace(ag.ToolsPrompt) != "" {
			msgs = append(msgs, fantasy.NewSystemMessage(ag.ToolsPrompt))
		}
		if strings.TrimSpace(ag.SkillsPrompt) != "" {
			msgs = append(msgs, fantasy.NewSystemMessage(ag.SkillsPrompt))
		}
		if strings.TrimSpace(ag.DelegateContext) != "" {
			msgs = append(msgs, fantasy.NewSystemMessage(ag.DelegateContext))
		}
		chatSession = &ChatSession{Agent: ag, Messages: msgs}
		r.sessions[ag.Handle] = chatSession

		// Create database session if services available
		if r.services != nil {
			dbSession, err := r.services.Sessions.Create(ctx, session.CreateParams{
				AgentHandle: ag.Handle,
				Title:       generateSessionTitle(input),
			})
			if err == nil {
				chatSession.SessionID = dbSession.ID
			}
		}
	}

	// Add user message
	chatSession.Messages = append(chatSession.Messages, fantasy.NewUserMessage(input))

	// Persist user message
	if r.services != nil && chatSession.SessionID != "" {
		r.services.Messages.Create(ctx, session.CreateMessageParams{
			SessionID: chatSession.SessionID,
			Role:      session.RoleUser,
			Parts:     []session.ContentPart{session.TextContent{Text: input}},
			Model:     ag.Model,
		})
	}

	// Inject session context for tools
	toolCtx := ctx
	if chatSession.SessionID != "" && r.services != nil {
		toolCtx = WithSessionID(toolCtx, chatSession.SessionID)
		toolCtx = WithServices(toolCtx, r.services)
	}

	// Run the chat and get response
	resp, newMsgs, err := r.runChatWithHistory(toolCtx, ag, chatSession.Messages)
	if err != nil {
		// Remove the failed user message
		chatSession.Messages = chatSession.Messages[:len(chatSession.Messages)-1]
		return "", err
	}

	// Update session with full message history
	chatSession.Messages = newMsgs

	// Persist assistant response
	if r.services != nil && chatSession.SessionID != "" {
		// Get the last assistant message from newMsgs
		for i := len(newMsgs) - 1; i >= 0; i-- {
			if newMsgs[i].Role == fantasy.MessageRoleAssistant {
				parts := r.fantasyPartsToSessionParts(newMsgs[i].Content)
				r.services.Messages.Create(ctx, session.CreateMessageParams{
					SessionID: chatSession.SessionID,
					Role:      session.RoleAssistant,
					Parts:     parts,
					Model:     ag.Model,
				})
				break
			}
		}

		// Generate title async after first exchange
		if !chatSession.TitleGenerated {
			chatSession.TitleGenerated = true
			go r.generateTitleAsync(ag.Model, chatSession.SessionID, input, resp)
		}
	}

	// Async memory formation: detect triggers and queue formation
	if r.formationService != nil && ag.Config.Memory.Enabled {
		r.maybeFormMemory(ctx, ag, input, chatSession.SessionID)
	}

	return resp, nil
}

// GetSessionID returns the current session ID for an agent (empty if no session).
func (r *Runner) GetSessionID(agentHandle string) string {
	if s, ok := r.sessions[agentHandle]; ok {
		return s.SessionID
	}
	return ""
}

// ResumeSession restores a chat session from persisted messages.
// This allows continuing a previous conversation.
func (r *Runner) ResumeSession(ctx context.Context, ag agent.Agent, sessionID string, messages []session.Message) error {
	// Build system prompt with memory context
	systemPrompt := ag.CombinedSystem
	if r.memoryService != nil && ag.Config.Memory.Enabled && ag.Config.Memory.Retrieval.AutoInject {
		// Use last user message as query for memory retrieval
		var query string
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == session.RoleUser {
				for _, part := range messages[i].Parts {
					if textPart, ok := part.(session.TextContent); ok {
						query = textPart.Text
						break
					}
				}
				break
			}
		}
		if query != "" {
			memCtx, err := agent.BuildMemoryContext(ctx, r.memoryService, ag.Handle, "", query, ag.Config.Memory)
			if err == nil && memCtx != nil {
				systemPrompt = agent.InjectMemoryContext(systemPrompt, memCtx)
			}
		}
	}

	// Build system messages
	var msgs []fantasy.Message
	if strings.TrimSpace(systemPrompt) != "" {
		msgs = append(msgs, fantasy.NewSystemMessage(systemPrompt))
	}
	if strings.TrimSpace(ag.ToolsPrompt) != "" {
		msgs = append(msgs, fantasy.NewSystemMessage(ag.ToolsPrompt))
	}
	if strings.TrimSpace(ag.SkillsPrompt) != "" {
		msgs = append(msgs, fantasy.NewSystemMessage(ag.SkillsPrompt))
	}

	// Convert persisted messages to Fantasy messages
	for _, msg := range messages {
		// Skip system messages as we already added them
		if msg.Role == session.RoleSystem {
			continue
		}
		msgs = append(msgs, msg.ToFantasyMessage())
	}

	// Create the chat session
	chatSession := &ChatSession{
		Agent:     ag,
		Messages:  msgs,
		SessionID: sessionID,
	}
	r.sessions[ag.Handle] = chatSession

	return nil
}

// GetSessionMessages retrieves messages for the current session from the database.
// Returns nil if no session exists or no services are configured.
func (r *Runner) GetSessionMessages(ctx context.Context, agentHandle string) ([]session.Message, error) {
	if r.services == nil {
		return nil, nil
	}

	chatSession, ok := r.sessions[agentHandle]
	if !ok || chatSession.SessionID == "" {
		return nil, nil
	}

	return r.services.Messages.List(ctx, chatSession.SessionID)
}

// TextResult contains the response and session ID from a Text call.
type TextResult struct {
	Response  string
	SessionID string
}

// Text runs a single prompt without maintaining history.
func (r *Runner) Text(ctx context.Context, ag agent.Agent, prompt string, attachments []string) (string, error) {
	result, err := r.TextWithSession(ctx, ag, prompt, attachments)
	if err != nil {
		return "", err
	}
	return result.Response, nil
}

// TextWithSession runs a single prompt and returns the session ID.
func (r *Runner) TextWithSession(ctx context.Context, ag agent.Agent, prompt string, attachments []string) (TextResult, error) {
	msgs := r.buildMessagesWithAttachments(ctx, ag, prompt, attachments)

	var sessionID string

	// Create database session if services available
	if r.services != nil {
		dbSession, err := r.services.Sessions.Create(ctx, session.CreateParams{
			AgentHandle: ag.Handle,
			Title:       generateSessionTitle(prompt),
		})
		if err == nil {
			sessionID = dbSession.ID

			// Persist user message
			r.services.Messages.Create(ctx, session.CreateMessageParams{
				SessionID: sessionID,
				Role:      session.RoleUser,
				Parts:     []session.ContentPart{session.TextContent{Text: prompt}},
				Model:     ag.Model,
			})
		}
	}

	// Inject session context for tools
	toolCtx := ctx
	if sessionID != "" && r.services != nil {
		toolCtx = WithSessionID(toolCtx, sessionID)
		toolCtx = WithServices(toolCtx, r.services)
	}

	resp, err := r.runChat(toolCtx, ag, msgs)
	if err != nil {
		return TextResult{}, err
	}

	// Persist assistant response and generate title
	if r.services != nil && sessionID != "" {
		r.services.Messages.Create(ctx, session.CreateMessageParams{
			SessionID: sessionID,
			Role:      session.RoleAssistant,
			Parts:     []session.ContentPart{session.TextContent{Text: resp}},
			Model:     ag.Model,
		})

		// Generate title async
		go r.generateTitleAsync(ag.Model, sessionID, prompt, resp)
	}

	// Async memory formation: detect triggers and queue formation
	if r.formationService != nil && ag.Config.Memory.Enabled {
		r.maybeFormMemory(ctx, ag, prompt, sessionID)
	}

	return TextResult{Response: resp, SessionID: sessionID}, nil
}

func (r *Runner) buildMessages(ctx context.Context, ag agent.Agent, prompt string) []fantasy.Message {
	return r.buildMessagesWithAttachments(ctx, ag, prompt, nil)
}

func (r *Runner) buildMessagesWithAttachments(ctx context.Context, ag agent.Agent, prompt string, attachments []string) []fantasy.Message {
	var msgs []fantasy.Message
	
	// Build combined system prompt with memory context
	systemPrompt := ag.CombinedSystem
	if r.memoryService != nil && ag.Config.Memory.Enabled {
		memCtx, err := agent.BuildMemoryContext(ctx, r.memoryService, ag.Handle, "", prompt, ag.Config.Memory)
		if err == nil && memCtx != nil {
			systemPrompt = agent.InjectMemoryContext(systemPrompt, memCtx)
		}
	}
	
	if strings.TrimSpace(systemPrompt) != "" {
		msgs = append(msgs, fantasy.NewSystemMessage(systemPrompt))
	}
	if strings.TrimSpace(ag.ToolsPrompt) != "" {
		msgs = append(msgs, fantasy.NewSystemMessage(ag.ToolsPrompt))
	}
	if strings.TrimSpace(ag.SkillsPrompt) != "" {
		msgs = append(msgs, fantasy.NewSystemMessage(ag.SkillsPrompt))
	}

	// Add model context for sub-agents that need to pass the model through
	if ag.Model != "" {
		modelContext := fmt.Sprintf("<model_context>\nYou are running with model: %s\nWhen delegating to external tools that accept a model parameter (like crush run --model), use this model.\n</model_context>", ag.Model)
		msgs = append(msgs, fantasy.NewSystemMessage(modelContext))
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

	// Build tool set with memory queue and depth for proper UI nesting
	baseDir, _ := os.Getwd()
	tools := NewFantasyToolSetWithOptions(ag.Config.AllowedTools, baseDir, r.memoryQueue, r.depth)

	// Add agent_call if explicitly allowed in config (for any agent)
	// or if it's a non-builtin agent (user agents get it by default)
	if tools.HasTool("agent_call") || !ag.BuiltIn {
		tools.AddAgentCallTool(r.agentCallExecutor(ag.Handle))
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
			currentTool.Metadata = result.ClientMetadata
			if result.Result.GetType() == fantasy.ToolResultContentTypeError {
				currentTool.Error = currentTool.Output
			}
			// Skip normal display for agent_call - sub-agent already shows its result
			if currentTool.Name != "agent_call" {
				ui.PrintToolCallResult(currentTool)
			}
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

func (r *Runner) agentCallExecutor(currentAgentHandle string) func(ctx context.Context, params AgentCallParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	return func(ctx context.Context, params AgentCallParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
		// Normalize handle
		agentHandle := agent.NormalizeHandle(params.Agent)

		// Prevent self-delegation loops
		if agentHandle == currentAgentHandle {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("cannot delegate to self (%s) - use bash or other tools directly", agentHandle)), nil
		}

		// Only allow calling builtin agents or plugin agents
		if !agent.IsReservedNamespace(agentHandle) && !plugins.IsPluginAgent(agentHandle) {
			return fantasy.NewTextErrorResponse("agent_call can only invoke builtin or plugin agents"), nil
		}

		// Load the target agent
		targetAgent, err := agent.Load(r.config, agentHandle)
		if err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to load agent %s: %v", agentHandle, err)), nil
		}

		// Override model if specified in params
		if params.Model != "" {
			targetAgent.Model = params.Model
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
			sessions: make(map[string]*ChatSession),
			services: r.services, // Pass services through for persistence
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

// generateSessionTitle creates a simple fallback title from the prompt.
// Used when LLM title generation fails or services are unavailable.
func generateSessionTitle(prompt string) string {
	const maxLen = 60

	// Trim whitespace and collapse multiple spaces/newlines
	title := strings.TrimSpace(prompt)
	title = strings.Join(strings.Fields(title), " ")

	if len(title) <= maxLen {
		return title
	}

	// Truncate and add ellipsis
	return title[:maxLen-1] + "…"
}

// generateTitleAsync uses an LLM to generate a concise title for the session.
// Runs in a goroutine so it doesn't block the conversation.
func (r *Runner) generateTitleAsync(modelID, sessionID, userMessage, assistantResponse string) {
	if r.services == nil || sessionID == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a model for title generation
	model, err := NewLanguageModel(ctx, r.config.Provider, modelID)
	if err != nil {
		return // Silent fail - title stays as default
	}

	// Truncate messages to avoid excessive token usage
	userMsg := truncateForTitle(userMessage, 500)
	assistantMsg := truncateForTitle(assistantResponse, 500)

	titlePrompt := fmt.Sprintf("Generate a short, descriptive title (max 50 chars) for this conversation. The title should capture the main topic or intent. Return ONLY the title, no quotes or explanation.\n\nUser: %s\n\nAssistant: %s", userMsg, assistantMsg)

	agent := fantasy.NewAgent(model)
	result, err := agent.Generate(ctx, fantasy.AgentCall{
		Prompt: titlePrompt,
	})
	if err != nil {
		return // Silent fail
	}

	// Extract title from response
	title := strings.TrimSpace(result.Response.Content.Text())
	if title == "" {
		return
	}

	// Truncate if too long
	if len(title) > 60 {
		title = title[:59] + "…"
	}

	// Update the session title
	r.services.Sessions.UpdateTitle(ctx, sessionID, title)
}

// truncateForTitle truncates a string to maxLen for title generation prompts.
func truncateForTitle(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

// maybeFormMemory uses small model to extract memorable content from user messages.
func (r *Runner) maybeFormMemory(ctx context.Context, ag agent.Agent, userMessage, sessionID string) {
	// Need memory service with embedder for deduplication and embedding generation
	if r.memoryService == nil || !r.memoryService.HasEmbedder() {
		return
	}

	// Need small model for intelligent extraction
	if r.smallModel == nil {
		return
	}

	// Check if memory is enabled for this agent
	if !ag.Config.Memory.Enabled {
		return
	}

	// Check if explicit_only mode - skip automatic extraction
	cfg := ag.Config.Memory.FormationTriggers
	if cfg.ExplicitOnly {
		return
	}

	// Use small model to extract memorable content
	extraction, err := r.smallModel.ExtractMemory(ctx, userMessage)
	if err != nil {
		if r.debug {
			fmt.Fprintf(os.Stderr, "DEBUG: memory extraction failed: %v\n", err)
		}
		return
	}

	// Nothing to remember
	if !extraction.ShouldRemember || extraction.Content == "" {
		return
	}

	// Filter based on agent config
	switch extraction.Category {
	case "correction":
		if !cfg.OnCorrection {
			return
		}
	case "preference":
		if !cfg.OnPreference {
			return
		}
	case "fact":
		if !cfg.OnProjectFact {
			return
		}
	}

	// Map category string to memory.Category
	category := categoryFromString(extraction.Category)

	// Check for duplicates using small model if we have similar memories
	existing, err := r.memoryService.Search(ctx, extraction.Content, memory.SearchOptions{
		AgentHandle: ag.Handle,
		Threshold:   memory.SupersedeThreshold,
		Limit:       5,
	})
	if err != nil {
		if r.debug {
			fmt.Fprintf(os.Stderr, "DEBUG: memory search failed: %v\n", err)
		}
		// Continue with creation anyway
		existing = nil
	}

	// If we have similar memories, use small model to decide what to do
	if len(existing) > 0 {
		existingList := make([]smallmodel.ExistingMemory, len(existing))
		for i, m := range existing {
			existingList[i] = smallmodel.ExistingMemory{
				ID:      m.Memory.ID,
				Content: m.Memory.Content,
			}
		}

		decision, err := r.smallModel.CheckDuplicate(ctx, extraction.Content, existingList)
		if err != nil {
			if r.debug {
				fmt.Fprintf(os.Stderr, "DEBUG: dedup check failed: %v\n", err)
			}
			// Continue with creation anyway
		} else {
			switch decision.Action {
			case "duplicate":
				// Already have this memory, skip
				if r.formationService != nil {
					r.formationService.NotifySkipped(extraction.Content, existingList[0].ID)
				}
				return
			case "supersede":
				// Find the memory to supersede
				targetID := decision.TargetID
				if targetID == "" && len(existingList) > 0 {
					targetID = existingList[0].ID
				}
				if targetID != "" {
					mem, err := r.memoryService.Supersede(ctx, targetID, memory.Memory{
						Content:         extraction.Content,
						Category:        category,
						AgentHandle:     ag.Handle,
						SourceSessionID: sessionID,
					}, decision.Reason)
					if err != nil {
						if r.debug {
							fmt.Fprintf(os.Stderr, "DEBUG: memory supersede failed: %v\n", err)
						}
						if r.formationService != nil {
							r.formationService.NotifyFailed(extraction.Content, err)
						}
					} else {
						if r.formationService != nil {
							r.formationService.NotifySuperseded(mem, targetID)
						}
					}
					return
				}
			}
			// action == "new", fall through to create
		}
	}

	// Create the memory
	mem, err := r.memoryService.Create(ctx, memory.Memory{
		Content:         extraction.Content,
		Category:        category,
		AgentHandle:     ag.Handle,
		SourceSessionID: sessionID,
	})
	if err != nil {
		if r.debug {
			fmt.Fprintf(os.Stderr, "DEBUG: memory creation failed: %v\n", err)
		}
		if r.formationService != nil {
			r.formationService.NotifyFailed(extraction.Content, err)
		}
	} else {
		if r.formationService != nil {
			r.formationService.NotifyCreated(mem)
		}
	}
}

// categoryFromString converts a category string to memory.Category.
func categoryFromString(s string) memory.Category {
	switch s {
	case "preference":
		return memory.CategoryPreference
	case "fact":
		return memory.CategoryFact
	case "correction":
		return memory.CategoryCorrection
	case "pattern":
		return memory.CategoryPattern
	default:
		return memory.CategoryFact
	}
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

// fantasyPartsToSessionParts converts Fantasy message parts to session content parts.
func (r *Runner) fantasyPartsToSessionParts(parts []fantasy.MessagePart) []session.ContentPart {
	var result []session.ContentPart
	for _, p := range parts {
		switch part := p.(type) {
		case fantasy.TextPart:
			result = append(result, session.TextContent{Text: part.Text})
		case fantasy.ReasoningPart:
			result = append(result, session.ReasoningContent{Text: part.Text})
		case fantasy.FilePart:
			result = append(result, session.FileContent{
				Filename:  part.Filename,
				Data:      part.Data,
				MediaType: part.MediaType,
			})
		case fantasy.ToolCallPart:
			result = append(result, session.ToolCall{
				ID:               part.ToolCallID,
				Name:             part.ToolName,
				Input:            part.Input,
				ProviderExecuted: part.ProviderExecuted,
			})
		case fantasy.ToolResultPart:
			tr := session.ToolResult{
				ToolCallID: part.ToolCallID,
			}
			switch out := part.Output.(type) {
			case fantasy.ToolResultOutputContentText:
				tr.Content = out.Text
			case fantasy.ToolResultOutputContentError:
				tr.Content = out.Error.Error()
				tr.IsError = true
			case fantasy.ToolResultOutputContentMedia:
				tr.Content = out.Text
			}
			result = append(result, tr)
		}
	}
	return result
}
