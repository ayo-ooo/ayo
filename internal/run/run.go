package run

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/share"
	"github.com/alexcabrera/ayo/internal/smallmodel"
	"github.com/alexcabrera/ayo/internal/squads"
	uipkg "github.com/alexcabrera/ayo/internal/ui"
	"github.com/alexcabrera/ayo/internal/util"
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
	streamWriter     StreamWriter             // nil = use default PrintWriter
	sandboxProvider  providers.SandboxProvider // nil = no sandbox, run locally
	squadName        string                   // squad name for constitution injection (empty = no squad)
	dispatcher       *Dispatcher              // nil = no semantic dispatch
	shareService     *share.Service           // nil = no share service for request_access
}

// ChatSession maintains conversation state for interactive chat.
type ChatSession struct {
	Agent          agent.Agent
	Messages       []fantasy.Message
	SessionID      string // Database session ID (empty if no persistence)
	TitleGenerated bool   // Whether title generation has been triggered
}

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
	StreamWriter     StreamWriter               // Unified stream writer interface
	SandboxProvider  providers.SandboxProvider  // Sandbox provider for isolated execution
	SquadName        string                     // Squad name for squad context injection (empty = no squad)
	Dispatcher       *Dispatcher                // Dispatcher for semantic routing decisions
	ShareService     *share.Service             // Share service for request_access tool
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
		streamWriter:     opts.StreamWriter,
		sandboxProvider:  opts.SandboxProvider,
		squadName:        opts.SquadName,
		dispatcher:       opts.Dispatcher,
		shareService:     opts.ShareService,
	}, nil
}

// SetStreamWriter sets a custom stream writer for streaming output.
func (r *Runner) SetStreamWriter(w StreamWriter) {
	r.streamWriter = w
}

// MemoryService returns the memory service, or nil if not configured.
func (r *Runner) MemoryService() *memory.Service {
	return r.memoryService
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

// Dispatcher returns the dispatcher for semantic routing decisions, or nil if not configured.
func (r *Runner) Dispatcher() *Dispatcher {
	return r.dispatcher
}

// DecideDispatch uses semantic search to determine where to route a prompt.
// Returns the best target (@ayo, @agent, or #squad) based on embedding similarity.
// If no dispatcher is configured, returns @ayo as the default target.
func (r *Runner) DecideDispatch(ctx context.Context, prompt string) (*DispatchDecision, error) {
	if r.dispatcher == nil {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 1.0,
			Reason:     "no dispatcher configured",
		}, nil
	}
	return r.dispatcher.Decide(ctx, prompt)
}

// Chat sends a message in an interactive session, maintaining conversation history.
func (r *Runner) Chat(ctx context.Context, ag agent.Agent, input string) (string, error) {
	chatSession, ok := r.sessions[ag.Handle]
	if !ok {
		// Initialize new session with system messages
		var msgs []fantasy.Message
		
		// Build combined system prompt with memory context
		systemPrompt := ag.CombinedSystem
		
		// Inject squad constitution if running in a squad context
		if r.squadName != "" {
			constitution, err := squads.LoadConstitution(r.squadName)
			if err == nil && constitution != nil {
				systemPrompt = squads.InjectConstitution(systemPrompt, constitution)
			}
		}
		
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

// ContinueSessionWithPrompt continues an existing session with a new prompt.
// Unlike ResumeSession (for interactive mode), this runs a single prompt and returns.
func (r *Runner) ContinueSessionWithPrompt(ctx context.Context, ag agent.Agent, existingSessionID string, prompt string, attachments []string) (TextResult, error) {
	if r.services == nil {
		return TextResult{}, fmt.Errorf("session continuation requires database services")
	}

	// Load existing messages from the session
	existingMsgs, err := r.services.Messages.List(ctx, existingSessionID)
	if err != nil {
		return TextResult{}, fmt.Errorf("failed to load session messages: %w", err)
	}

	// Build system messages (same as buildMessagesWithAttachments)
	var msgs []fantasy.Message

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

	// Convert persisted messages to Fantasy messages
	for _, msg := range existingMsgs {
		if msg.Role == session.RoleSystem {
			continue
		}
		msgs = append(msgs, msg.ToFantasyMessage())
	}

	// Handle attachments in the new prompt
	newMsgs := r.buildMessagesWithAttachments(ctx, ag, prompt, attachments)
	// Find the user message from newMsgs (last message) and add it
	for _, m := range newMsgs {
		if m.Role == fantasy.MessageRoleUser {
			msgs = append(msgs, m)
			break
		}
	}

	// Persist the new user message
	r.services.Messages.Create(ctx, session.CreateMessageParams{
		SessionID: existingSessionID,
		Role:      session.RoleUser,
		Parts:     []session.ContentPart{session.TextContent{Text: prompt}},
		Model:     ag.Model,
	})

	// Inject session context for tools
	toolCtx := WithSessionID(ctx, existingSessionID)
	toolCtx = WithServices(toolCtx, r.services)

	resp, err := r.runChat(toolCtx, ag, msgs)
	if err != nil {
		return TextResult{}, err
	}

	// Persist assistant response
	r.services.Messages.Create(ctx, session.CreateMessageParams{
		SessionID: existingSessionID,
		Role:      session.RoleAssistant,
		Parts:     []session.ContentPart{session.TextContent{Text: resp}},
		Model:     ag.Model,
	})

	// Trigger memory formation
	if r.formationService != nil && ag.Config.Memory.Enabled {
		r.maybeFormMemory(ctx, ag, prompt, existingSessionID)
	}

	return TextResult{Response: resp, SessionID: existingSessionID}, nil
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
	return slices.Contains(textTypes, baseType)
}

func (r *Runner) runChat(ctx context.Context, ag agent.Agent, msgs []fantasy.Message) (string, error) {
	resp, _, err := r.runChatWithHistory(ctx, ag, msgs)
	return resp, err
}

func (r *Runner) runChatWithHistory(ctx context.Context, ag agent.Agent, msgs []fantasy.Message) (string, []fantasy.Message, error) {
	if strings.TrimSpace(ag.Model) == "" {
		return "", nil, fmt.Errorf("model is required")
	}

	// Extract prompt and history from messages
	prompt, historyMsgs := extractPromptAndHistory(msgs)

	// Build Fantasy agent with tools
	fantasyAgent, model, sandboxID, err := r.buildFantasyAgent(ctx, ag)
	if err != nil {
		return "", nil, err
	}
	
	// Ensure sandbox cleanup if one was created
	if sandboxID != "" && r.sandboxProvider != nil {
		defer func() {
			_ = r.sandboxProvider.Delete(context.Background(), sandboxID, true)
		}()
	}

	// Use custom stream writer if provided, otherwise use default print writer
	var handler *FantasyAdapter
	if r.streamWriter != nil {
		handler = NewFantasyAdapter(r.streamWriter)
	} else {
		// Default: create PrintWriter which implements StreamWriter
		handler = NewFantasyAdapter(NewPrintWriter(ag.Handle, r.debug, r.depth))
	}

	var content strings.Builder
	var reasoningStartTime time.Time
	var toolStartTime time.Time

	// Stream the response with all callbacks
	result, err := fantasyAgent.Stream(ctx, fantasy.AgentStreamCall{
		Prompt:   prompt,
		Messages: historyMsgs,

		// Reasoning streams (for models like Claude that expose thinking)
		OnReasoningDelta: func(id, text string) error {
			if reasoningStartTime.IsZero() {
				reasoningStartTime = time.Now()
				handler.OnReasoningStart(id)
			}
			return handler.OnReasoningDelta(id, text)
		},

		OnReasoningEnd: func(id string, reasoning fantasy.ReasoningContent) error {
			duration := time.Since(reasoningStartTime)
			reasoningStartTime = time.Time{}
			return handler.OnReasoningEnd(id, duration)
		},

		// Tool call complete - show the command we're about to run
		OnToolCall: func(tc fantasy.ToolCallContent) error {
			toolStartTime = time.Now()
			return handler.OnToolCall(tc)
		},

		// Tool result - show the output
		OnToolResult: func(result fantasy.ToolResultContent) error {
			duration := time.Since(toolStartTime)
			toolStartTime = time.Time{}
			return handler.OnToolResult(result, duration)
		},

		// Text response streams
		OnTextDelta: func(id, text string) error {
			content.WriteString(text)
			return handler.OnTextDelta(id, text)
		},
	})

	// Notify handler of text completion
	if content.Len() > 0 {
		handler.OnTextEnd("")
	}

	// Handle errors
	if err != nil {
		handler.OnError(err)
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
		ui := uipkg.NewWithDepth(r.debug, r.depth)
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
	ui := uipkg.NewWithDepth(r.debug, r.depth)
	if ui.IsPiped() {
		return strings.TrimSpace(finalContent), msgs, nil
	}

	// Text was already streamed to output, return empty to avoid duplicate
	return "", msgs, nil
}

// extractPromptAndHistory extracts the last user message as prompt and returns remaining history.
// Fantasy requires a non-empty Prompt field.
func extractPromptAndHistory(msgs []fantasy.Message) (string, []fantasy.Message) {
	if len(msgs) == 0 {
		return "", nil
	}
	last := msgs[len(msgs)-1]
	if last.Role != fantasy.MessageRoleUser {
		return "", msgs
	}
	// Get text content from last user message
	var prompt string
	for _, part := range last.Content {
		if tp, ok := part.(fantasy.TextPart); ok {
			prompt = tp.Text
			break
		}
	}
	return prompt, msgs[:len(msgs)-1]
}

// buildFantasyAgent creates a Fantasy agent with tools for the given agent configuration.
// Returns the agent, model, sandbox ID (if created), and any error.
// The caller is responsible for cleaning up the sandbox if sandboxID is non-empty.
func (r *Runner) buildFantasyAgent(ctx context.Context, ag agent.Agent) (fantasy.Agent, fantasy.LanguageModel, string, error) {
	model, err := NewLanguageModel(ctx, r.config.Provider, ag.Model)
	if err != nil {
		return nil, nil, "", fmt.Errorf("create language model: %w", err)
	}

	baseDir, _ := os.Getwd()
	
	// Check if agent has sandbox enabled and we have a provider
	var sandboxExecutor *sandbox.Executor
	var sandboxID string
	if ag.Config.SandboxEnabled() && r.sandboxProvider != nil {
		// Check if this is the @ayo orchestrator agent - use dedicated sandbox
		if isAyoAgent(ag.Handle) {
			sb, ayoErr := r.ensureAyoSandbox(ctx, baseDir)
			if ayoErr != nil {
				return nil, nil, "", fmt.Errorf("ensure ayo sandbox: %w", ayoErr)
			}
			sandboxID = sb.ID
			sandboxExecutor = sandbox.NewExecutor(r.sandboxProvider, sandboxID, baseDir, "ayo")
			
			if r.debug {
				fmt.Fprintf(os.Stderr, "[sandbox] Using dedicated @ayo sandbox %s\n", sandboxID)
			}
		} else {
			// Create sandbox for this agent
			// Sanitize handle for container name (remove @ and other invalid chars)
			safeName := strings.TrimPrefix(ag.Handle, "@")
			safeName = strings.ReplaceAll(safeName, ".", "-")
			// Use SandboxUser() to get effective user (explicit or derived from handle)
			sandboxUser := ag.Config.Sandbox.SandboxUser(ag.Handle)

			// Build mount list
			mounts := []providers.Mount{{
				Source:      baseDir,
				Destination: baseDir,
				Mode:        providers.MountModeVirtioFS,
				ReadOnly:    false,
			}}

			// Add persistent home directory mount if enabled
			if ag.Config.Sandbox.PersistHomeEnabled() && sandboxUser != "" {
				hostHomeDir, homeErr := paths.EnsureAgentHomeDir(ag.Handle)
				if homeErr != nil {
					return nil, nil, "", fmt.Errorf("create agent home directory: %w", homeErr)
				}
				containerHomeDir := fmt.Sprintf("/home/%s", sandboxUser)
				mounts = append(mounts, providers.Mount{
					Source:      hostHomeDir,
					Destination: containerHomeDir,
					Mode:        providers.MountModeVirtioFS,
					ReadOnly:    false,
				})
				if r.debug {
					fmt.Fprintf(os.Stderr, "[sandbox] Mounting persistent home %s -> %s\n", hostHomeDir, containerHomeDir)
				}
			}

			// Add Matrix/daemon runtime directory mount for inter-agent communication
			mounts = append(mounts, providers.Mount{
				Source:      paths.RuntimeDir(),
				Destination: "/run/ayo",
				Mode:        providers.MountModeVirtioFS,
				ReadOnly:    false,
			})
			if r.debug {
				fmt.Fprintf(os.Stderr, "[sandbox] Mounting daemon socket directory %s -> /run/ayo\n", paths.RuntimeDir())
			}

			sb, createErr := r.sandboxProvider.Create(ctx, providers.SandboxCreateOptions{
				Name:   fmt.Sprintf("ayo-%s-%d", safeName, time.Now().UnixNano()),
				Image:  ag.Config.Sandbox.SandboxImage(),
				User:   sandboxUser,
				Mounts: mounts,
			})
			if createErr != nil {
				return nil, nil, "", fmt.Errorf("create sandbox: %w", createErr)
			}
			sandboxID = sb.ID

			// Ensure agent user exists in sandbox before any tool execution
			// Build dotfiles path from agent directory if it exists
			var dotfilesPath string
			if ag.Dir != "" {
				candidatePath := filepath.Join(ag.Dir, "sandbox", "dotfiles")
				if info, err := os.Stat(candidatePath); err == nil && info.IsDir() {
					dotfilesPath = candidatePath
				}
			}
			if err := r.sandboxProvider.EnsureAgentUser(ctx, sandboxID, sandboxUser, dotfilesPath); err != nil {
				// Clean up sandbox on failure
				_ = r.sandboxProvider.Delete(context.Background(), sandboxID, true)
				return nil, nil, "", fmt.Errorf("ensure agent user: %w", err)
			}
			if r.debug && dotfilesPath != "" {
				fmt.Fprintf(os.Stderr, "[sandbox] Copied dotfiles from %s\n", dotfilesPath)
			}

			sandboxExecutor = sandbox.NewExecutor(r.sandboxProvider, sandboxID, baseDir, sandboxUser)
			
			// Set up session workspace with environment variables
			sessionID := GetSessionIDFromContext(ctx)
			sandboxExecutor.SetSession(sessionID, ag.Handle)
			
			// Create the workspace directory structure if session ID is available
			if sessionID != "" {
				if err := sandboxExecutor.CreateSessionWorkspace(ctx); err != nil {
					// Log warning but don't fail - workspace is optional
					if r.debug {
						fmt.Fprintf(os.Stderr, "[sandbox] Warning: failed to create workspace: %v\n", err)
					}
				} else if r.debug {
					fmt.Fprintf(os.Stderr, "[sandbox] Created workspace at %s\n", sandboxExecutor.WorkspaceDir())
				}
			}
			
			if r.debug {
				fmt.Fprintf(os.Stderr, "[sandbox] Created sandbox %s for %s (user=%s)\n", sandboxID, ag.Handle, sandboxUser)
			}
		}
	} else if ag.Config.SandboxEnabled() && r.sandboxProvider == nil {
		if r.debug {
			fmt.Fprintf(os.Stderr, "[sandbox] Agent %s has sandbox enabled but no provider available\n", ag.Handle)
		}
	}
	
	tools := NewFantasyToolSet(ToolSetOptions{
		AllowedTools:    ag.Config.AllowedTools,
		BaseDir:         baseDir,
		MemoryQueue:     r.memoryQueue,
		Depth:           r.depth,
		DisableTodo:     ag.Config.DisableTodo,
		SandboxExecutor: sandboxExecutor,
		ShareService:    r.shareService,
		SessionID:       sandboxID, // Use sandbox ID for session-scoped shares
	})

	fantasyAgent := fantasy.NewAgent(
		model,
		fantasy.WithSystemPrompt(""), // System prompt already in messages
		fantasy.WithTools(tools.Tools()...),
	)

	return fantasyAgent, model, sandboxID, nil
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
	userMsg := util.TruncateTitle(userMessage, 500)
	assistantMsg := util.TruncateTitle(assistantResponse, 500)

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

	// Store the memory (handles dedup, supersede, and create)
	r.storeMemory(ctx, memoryInput{
		content:     extraction.Content,
		category:    category,
		agentHandle: ag.Handle,
		sessionID:   sessionID,
	}, existing)
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

// memoryInput bundles parameters for memory storage.
type memoryInput struct {
	content     string
	category    memory.Category
	agentHandle string
	sessionID   string
}

// storeMemory handles deduplication and storage of a memory.
// It checks for duplicates/supersedes using the small model, then creates or updates.
func (r *Runner) storeMemory(ctx context.Context, input memoryInput, existing []memory.SearchResult) {
	// If we have similar memories, use small model to decide what to do
	if len(existing) > 0 {
		existingList := make([]smallmodel.ExistingMemory, len(existing))
		for i, m := range existing {
			existingList[i] = smallmodel.ExistingMemory{
				ID:      m.Memory.ID,
				Content: m.Memory.Content,
			}
		}

		decision, err := r.smallModel.CheckDuplicate(ctx, input.content, existingList)
		if err != nil {
			if r.debug {
				fmt.Fprintf(os.Stderr, "DEBUG: dedup check failed: %v\n", err)
			}
			// Continue with creation anyway
		} else {
			switch decision.Action {
			case "duplicate":
				if r.formationService != nil {
					r.formationService.NotifySkipped(input.content, existingList[0].ID)
				}
				return
			case "supersede":
				r.supersedeMemory(ctx, input, decision, existingList)
				return
			}
			// action == "new", fall through to create
		}
	}

	r.createMemory(ctx, input)
}

// supersedeMemory replaces an existing memory with new content.
func (r *Runner) supersedeMemory(ctx context.Context, input memoryInput, decision *smallmodel.DedupDecision, existingList []smallmodel.ExistingMemory) {
	targetID := decision.TargetID
	if targetID == "" && len(existingList) > 0 {
		targetID = existingList[0].ID
	}
	if targetID == "" {
		return
	}

	mem, err := r.memoryService.Supersede(ctx, targetID, memory.Memory{
		Content:         input.content,
		Category:        input.category,
		AgentHandle:     input.agentHandle,
		SourceSessionID: input.sessionID,
	}, decision.Reason)
	if err != nil {
		if r.debug {
			fmt.Fprintf(os.Stderr, "DEBUG: memory supersede failed: %v\n", err)
		}
		if r.formationService != nil {
			r.formationService.NotifyFailed(input.content, err)
		}
		return
	}
	if r.formationService != nil {
		r.formationService.NotifySuperseded(mem, targetID)
	}
}

// createMemory creates a new memory.
func (r *Runner) createMemory(ctx context.Context, input memoryInput) {
	mem, err := r.memoryService.Create(ctx, memory.Memory{
		Content:         input.content,
		Category:        input.category,
		AgentHandle:     input.agentHandle,
		SourceSessionID: input.sessionID,
	})
	if err != nil {
		if r.debug {
			fmt.Fprintf(os.Stderr, "DEBUG: memory creation failed: %v\n", err)
		}
		if r.formationService != nil {
			r.formationService.NotifyFailed(input.content, err)
		}
		return
	}
	if r.formationService != nil {
		r.formationService.NotifyCreated(mem)
	}
}

// castToStructuredOutput takes the agent's response and casts it to the required output schema.
// It uses GenerateObject to produce structured output, then validates against the schema.
// If validation fails, it retries by providing error feedback to the model.
func (r *Runner) castToStructuredOutput(ctx context.Context, model fantasy.LanguageModel, ag agent.Agent, agentOutput string, _ *uipkg.UI) (string, error) {
	if ag.OutputSchema == nil {
		return agentOutput, nil
	}

	var lastError error
	for attempt := range maxOutputCastRetries {
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

// isAyoAgent checks if the given agent handle is the @ayo orchestrator.
func isAyoAgent(handle string) bool {
	return handle == "@ayo" || handle == "ayo"
}

// ensureAyoSandbox ensures the dedicated @ayo sandbox exists and is running.
// Returns the sandbox info for use with the sandbox executor.
func (r *Runner) ensureAyoSandbox(ctx context.Context, baseDir string) (providers.Sandbox, error) {
	// Use the sandbox package's EnsureAyoSandbox function
	appleProvider, ok := r.sandboxProvider.(*sandbox.AppleProvider)
	if !ok {
		return providers.Sandbox{}, fmt.Errorf("ayo sandbox requires Apple Container provider")
	}
	
	sb, err := sandbox.EnsureAyoSandbox(ctx, appleProvider)
	if err != nil {
		return providers.Sandbox{}, err
	}
	
	// Add the current working directory as a mount
	// This allows @ayo to access the user's project
	// Note: The mount is configured in EnsureAyoSandbox, but we may want to
	// add dynamic mounts here in the future
	
	return sb, nil
}
