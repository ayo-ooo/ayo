package chat

import (
	"context"
	"encoding/json"
	"time"

	"charm.land/fantasy"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/ui/chat/panels"
	"github.com/alexcabrera/ayo/internal/ui/pubsub"
)

// TUIStreamHandler implements run.StreamHandler by sending tea.Msg
// to the Bubble Tea program and publishing events to brokers.
type TUIStreamHandler struct {
	program *tea.Program

	// Memory service for retrieving initial memories
	memoryService *memory.Service

	// Memory query context (set when SendInitialMemories is called)
	memoryAgentHandle string
	memoryThreshold   float32
	memoryLimit       int

	// Brokers for different event types
	messageBroker   *pubsub.Broker[pubsub.MessageEvent]
	toolBroker      *pubsub.Broker[pubsub.ToolEvent]
	memoryBroker    *pubsub.Broker[pubsub.MemoryEvent]
	textBroker      *pubsub.Broker[pubsub.TextDeltaEvent]
	reasoningBroker *pubsub.Broker[pubsub.ReasoningEvent]
}

// TUIStreamHandlerOption is a functional option for configuring TUIStreamHandler.
type TUIStreamHandlerOption func(*TUIStreamHandler)

// WithMemoryService sets the memory service for the handler.
func WithMemoryService(svc *memory.Service) TUIStreamHandlerOption {
	return func(h *TUIStreamHandler) {
		h.memoryService = svc
	}
}

// NewTUIStreamHandler creates a handler that sends messages to a tea.Program.
func NewTUIStreamHandler(p *tea.Program, opts ...TUIStreamHandlerOption) *TUIStreamHandler {
	h := &TUIStreamHandler{
		program:         p,
		messageBroker:   pubsub.NewBroker[pubsub.MessageEvent](32),
		toolBroker:      pubsub.NewBroker[pubsub.ToolEvent](32),
		memoryBroker:    pubsub.NewBroker[pubsub.MemoryEvent](16),
		textBroker:      pubsub.NewBroker[pubsub.TextDeltaEvent](64),
		reasoningBroker: pubsub.NewBroker[pubsub.ReasoningEvent](32),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// MessageBroker returns the message event broker for subscription.
func (h *TUIStreamHandler) MessageBroker() *pubsub.Broker[pubsub.MessageEvent] {
	return h.messageBroker
}

// ToolBroker returns the tool event broker for subscription.
func (h *TUIStreamHandler) ToolBroker() *pubsub.Broker[pubsub.ToolEvent] {
	return h.toolBroker
}

// MemoryBroker returns the memory event broker for subscription.
func (h *TUIStreamHandler) MemoryBroker() *pubsub.Broker[pubsub.MemoryEvent] {
	return h.memoryBroker
}

// TextBroker returns the text delta broker for subscription.
func (h *TUIStreamHandler) TextBroker() *pubsub.Broker[pubsub.TextDeltaEvent] {
	return h.textBroker
}

// ReasoningBroker returns the reasoning event broker for subscription.
func (h *TUIStreamHandler) ReasoningBroker() *pubsub.Broker[pubsub.ReasoningEvent] {
	return h.reasoningBroker
}

// Subscribe returns a channel of events for a given broker.
func (h *TUIStreamHandler) SubscribeMessages(ctx context.Context) <-chan pubsub.Event[pubsub.MessageEvent] {
	return h.messageBroker.Subscribe(ctx)
}

func (h *TUIStreamHandler) SubscribeTools(ctx context.Context) <-chan pubsub.Event[pubsub.ToolEvent] {
	return h.toolBroker.Subscribe(ctx)
}

func (h *TUIStreamHandler) SubscribeMemory(ctx context.Context) <-chan pubsub.Event[pubsub.MemoryEvent] {
	return h.memoryBroker.Subscribe(ctx)
}

func (h *TUIStreamHandler) SubscribeText(ctx context.Context) <-chan pubsub.Event[pubsub.TextDeltaEvent] {
	return h.textBroker.Subscribe(ctx)
}

func (h *TUIStreamHandler) SubscribeReasoning(ctx context.Context) <-chan pubsub.Event[pubsub.ReasoningEvent] {
	return h.reasoningBroker.Subscribe(ctx)
}

// StreamHandler implementation

func (h *TUIStreamHandler) OnTextDelta(id, text string) error {
	// Send directly to TUI for immediate rendering
	if h.program != nil {
		h.program.Send(TextDeltaMsg{Delta: text})
	}

	// Publish to broker for subscribers
	h.textBroker.Publish(pubsub.Event[pubsub.TextDeltaEvent]{
		Type: pubsub.UpdatedEvent,
		Payload: pubsub.TextDeltaEvent{
			ID:    id,
			Delta: text,
		},
	})

	return nil
}

func (h *TUIStreamHandler) OnTextEnd(id string) error {
	if h.program != nil {
		h.program.Send(TextEndMsg{})
	}

	h.textBroker.Publish(pubsub.Event[pubsub.TextDeltaEvent]{
		Type: pubsub.CompletedEvent,
		Payload: pubsub.TextDeltaEvent{
			ID:    id,
			Final: true,
		},
	})

	return nil
}

func (h *TUIStreamHandler) OnReasoningStart(id string) error {
	if h.program != nil {
		h.program.Send(ReasoningStartMsg{})
	}

	h.reasoningBroker.Publish(pubsub.Event[pubsub.ReasoningEvent]{
		Type: pubsub.StartedEvent,
		Payload: pubsub.ReasoningEvent{
			ID: id,
		},
	})

	return nil
}

func (h *TUIStreamHandler) OnReasoningDelta(id, text string) error {
	if h.program != nil {
		h.program.Send(ReasoningDeltaMsg{Delta: text})
	}

	h.reasoningBroker.Publish(pubsub.Event[pubsub.ReasoningEvent]{
		Type: pubsub.UpdatedEvent,
		Payload: pubsub.ReasoningEvent{
			ID:      id,
			Content: text,
		},
	})

	return nil
}

func (h *TUIStreamHandler) OnReasoningEnd(id string, duration time.Duration) error {
	if h.program != nil {
		h.program.Send(ReasoningEndMsg{Duration: duration.String()})
	}

	h.reasoningBroker.Publish(pubsub.Event[pubsub.ReasoningEvent]{
		Type: pubsub.CompletedEvent,
		Payload: pubsub.ReasoningEvent{
			ID:       id,
			Duration: duration,
		},
	})

	return nil
}

func (h *TUIStreamHandler) OnToolCall(tc fantasy.ToolCallContent) error {
	if h.program != nil {
		h.program.Send(ToolCallStartMsg{
			Name:        tc.ToolName,
			Description: "", // Will be extracted from input
			Command:     "", // Will be extracted from input for bash
		})
	}

	h.toolBroker.Publish(pubsub.Event[pubsub.ToolEvent]{
		Type: pubsub.StartedEvent,
		Payload: pubsub.ToolEvent{
			ToolCallID: tc.ToolCallID,
			Name:       tc.ToolName,
			Input:      tc.Input,
		},
	})

	return nil
}

func (h *TUIStreamHandler) OnToolResult(result fantasy.ToolResultContent, duration time.Duration) error {
	output := ""
	isError := result.Result != nil && result.Result.GetType() == fantasy.ToolResultContentTypeError
	if result.Result != nil {
		if text, ok := result.Result.(*fantasy.ToolResultOutputContentText); ok {
			output = text.Text
		} else if text, ok := result.Result.(fantasy.ToolResultOutputContentText); ok {
			output = text.Text
		}
	}

	if h.program != nil {
		errStr := ""
		if isError {
			errStr = output
		}
		h.program.Send(ToolCallResultMsg{
			Name:     result.ToolName,
			Output:   output,
			Error:    errStr,
			Duration: duration.String(),
		})

		// Handle todo tool metadata to update the planning panel
		if result.ToolName == "todo" && result.ClientMetadata != "" {
			var metadata struct {
				Todos []struct {
					Content    string `json:"content"`
					Status     string `json:"status"`
					ActiveForm string `json:"active_form"`
				} `json:"todos"`
			}
			if err := json.Unmarshal([]byte(result.ClientMetadata), &metadata); err == nil {
				todos := make([]panels.TodoItem, len(metadata.Todos))
				for i, t := range metadata.Todos {
					todos[i] = panels.TodoItem{
						Content:    t.Content,
						Status:     t.Status,
						ActiveForm: t.ActiveForm,
					}
				}
				h.program.Send(panels.TodosUpdateMsg{Todos: todos})
			}
		}
	}

	h.toolBroker.Publish(pubsub.Event[pubsub.ToolEvent]{
		Type: pubsub.CompletedEvent,
		Payload: pubsub.ToolEvent{
			ToolCallID: result.ToolCallID,
			Name:       result.ToolName,
			Output:     output,
			IsError:    isError,
			Metadata:   result.ClientMetadata,
			Duration:   duration,
		},
	})

	return nil
}

func (h *TUIStreamHandler) OnAgentStart(handle, prompt string) error {
	if h.program != nil {
		h.program.Send(SubAgentStartMsg{
			Handle: handle,
			Prompt: prompt,
		})
	}

	return nil
}

func (h *TUIStreamHandler) OnAgentEnd(handle string, duration time.Duration, err error) error {
	if h.program != nil {
		h.program.Send(SubAgentEndMsg{
			Handle:   handle,
			Duration: duration.String(),
			Error:    err != nil,
		})
	}

	return nil
}

func (h *TUIStreamHandler) OnMemoryEvent(event string, count int) error {
	if h.program != nil {
		h.program.Send(MemoryEventMsg{Type: event})

		// Re-query memories when a new one is created to update the panel
		if event == "created" && h.memoryService != nil && h.memoryAgentHandle != "" {
			// Use background context since we're in a callback
			ctx := context.Background()
			results, err := h.memoryService.Search(ctx, "session context", memory.SearchOptions{
				AgentHandle: h.memoryAgentHandle,
				Threshold:   h.memoryThreshold,
				Limit:       h.memoryLimit,
			})
			if err == nil && len(results) > 0 {
				items := convertSearchResultsToMemoryItems(results)
				h.program.Send(panels.MemoriesUpdateMsg{Memories: items})
			}
		}
	}

	h.memoryBroker.Publish(pubsub.Event[pubsub.MemoryEvent]{
		Type: pubsub.CreatedEvent,
		Payload: pubsub.MemoryEvent{
			Operation: event,
			Count:     count,
		},
	})

	return nil
}

func (h *TUIStreamHandler) OnError(err error) error {
	if h.program != nil {
		h.program.Send(ErrorMsg{Error: err})
	}

	return nil
}

// convertSearchResultToMemoryItem converts a memory search result to a panel item.
func convertSearchResultToMemoryItem(result memory.SearchResult) panels.MemoryItem {
	scope := result.Memory.AgentHandle
	if scope == "" {
		scope = result.Memory.PathScope
	}
	if scope == "" {
		scope = "global"
	}

	return panels.MemoryItem{
		ID:       result.Memory.ID,
		Content:  result.Memory.Content,
		Category: string(result.Memory.Category),
		Scope:    scope,
	}
}

// convertSearchResultsToMemoryItems converts multiple search results to panel items.
func convertSearchResultsToMemoryItems(results []memory.SearchResult) []panels.MemoryItem {
	items := make([]panels.MemoryItem, len(results))
	for i, r := range results {
		items[i] = convertSearchResultToMemoryItem(r)
	}
	return items
}

// SendInitialMemories retrieves and sends relevant memories to the TUI.
// This should be called after creating the program but before Run().
func (h *TUIStreamHandler) SendInitialMemories(ctx context.Context, query string, agentHandle string, threshold float32, limit int) error {
	if h.memoryService == nil || h.program == nil {
		return nil
	}

	// Use defaults if not specified
	if threshold <= 0 {
		threshold = 0.3
	}
	if limit <= 0 {
		limit = 10
	}

	// Save context for re-querying on memory events
	h.memoryAgentHandle = agentHandle
	h.memoryThreshold = threshold
	h.memoryLimit = limit

	results, err := h.memoryService.Search(ctx, query, memory.SearchOptions{
		AgentHandle: agentHandle,
		Threshold:   threshold,
		Limit:       limit,
	})
	if err != nil {
		return err
	}

	if len(results) > 0 {
		items := convertSearchResultsToMemoryItems(results)
		h.program.Send(panels.MemoriesUpdateMsg{Memories: items})
	}

	return nil
}

// Verify TUIStreamHandler implements run.StreamHandler
var _ run.StreamHandler = (*TUIStreamHandler)(nil)
