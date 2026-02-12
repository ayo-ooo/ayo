package zettelkasten

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/util"
	"github.com/google/uuid"
)

// Observer implements providers.ObserverProvider for memory extraction.
// It processes messages and extracts memorable information using an LLM.
type Observer struct {
	mu sync.RWMutex

	memoryProvider providers.MemoryProvider
	smallModel     SmallModelProvider

	queue     chan providers.MessageEvent
	done      chan struct{}
	wg        sync.WaitGroup
	running   bool
	batchSize int
	batchWait time.Duration
}

// SmallModelProvider provides access to a small/fast LLM for extraction.
// This is typically Ollama with a small model like ministral-3:3b.
type SmallModelProvider interface {
	// Complete generates a completion for the given prompt.
	Complete(ctx context.Context, prompt string) (string, error)
}

// ObserverConfig configures the memory observer.
type ObserverConfig struct {
	// MemoryProvider is the storage backend for memories
	MemoryProvider providers.MemoryProvider

	// SmallModel is the LLM for extracting memorable content
	SmallModel SmallModelProvider

	// BatchSize is how many messages to batch before processing (default 1)
	BatchSize int

	// BatchWait is how long to wait for more messages before processing (default 100ms)
	BatchWait time.Duration

	// QueueSize is the size of the message queue (default 100)
	QueueSize int
}

// NewObserver creates a new memory observer.
func NewObserver(cfg ObserverConfig) *Observer {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 1
	}
	if cfg.BatchWait <= 0 {
		cfg.BatchWait = 100 * time.Millisecond
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 100
	}

	return &Observer{
		memoryProvider: cfg.MemoryProvider,
		smallModel:     cfg.SmallModel,
		queue:          make(chan providers.MessageEvent, cfg.QueueSize),
		batchSize:      cfg.BatchSize,
		batchWait:      cfg.BatchWait,
	}
}

// Name returns the provider name.
func (o *Observer) Name() string {
	return "memory-extractor"
}

// Type returns the provider type.
func (o *Observer) Type() providers.ProviderType {
	return providers.ProviderTypeObserver
}

// Init initializes the observer.
func (o *Observer) Init(ctx context.Context, config map[string]any) error {
	// Configuration from map if needed
	return nil
}

// Close releases resources.
func (o *Observer) Close() error {
	return o.Stop()
}

// Start begins processing messages.
func (o *Observer) Start(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.running {
		return nil
	}

	o.done = make(chan struct{})
	o.running = true

	// Start worker goroutine
	o.wg.Add(1)
	go o.worker(ctx)

	return nil
}

// Stop stops the observer gracefully.
func (o *Observer) Stop() error {
	o.mu.Lock()
	if !o.running {
		o.mu.Unlock()
		return nil
	}
	o.running = false
	close(o.done)
	o.mu.Unlock()

	// Wait for worker to finish
	o.wg.Wait()
	return nil
}

// OnMessage queues a message for processing.
func (o *Observer) OnMessage(ctx context.Context, event providers.MessageEvent) error {
	o.mu.RLock()
	running := o.running
	o.mu.RUnlock()

	if !running {
		return nil
	}

	// Non-blocking send to queue
	select {
	case o.queue <- event:
		return nil
	default:
		// Queue full, drop message
		slog.Debug("memory observer queue full, dropping message",
			"session_id", event.SessionID,
			"message_id", event.MessageID)
		return nil
	}
}

// worker processes messages from the queue.
func (o *Observer) worker(ctx context.Context) {
	defer o.wg.Done()

	var batch []providers.MessageEvent
	timer := time.NewTimer(o.batchWait)
	timer.Stop()

	for {
		select {
		case <-o.done:
			// Process remaining messages
			if len(batch) > 0 {
				o.processBatch(ctx, batch)
			}
			return

		case event := <-o.queue:
			batch = append(batch, event)

			if len(batch) >= o.batchSize {
				timer.Stop()
				o.processBatch(ctx, batch)
				batch = nil
			} else if len(batch) == 1 {
				// Start timer on first message
				timer.Reset(o.batchWait)
			}

		case <-timer.C:
			if len(batch) > 0 {
				o.processBatch(ctx, batch)
				batch = nil
			}
		}
	}
}

// processBatch processes a batch of messages.
func (o *Observer) processBatch(ctx context.Context, batch []providers.MessageEvent) {
	for _, event := range batch {
		if err := o.processMessage(ctx, event); err != nil {
			slog.Debug("failed to process message for memory",
				"session_id", event.SessionID,
				"message_id", event.MessageID,
				"error", err)
		}
	}
}

// processMessage extracts memorable content from a single message.
func (o *Observer) processMessage(ctx context.Context, event providers.MessageEvent) error {
	// Only process user and assistant messages
	if event.Role != "user" && event.Role != "assistant" {
		return nil
	}

	// Skip very short messages
	if len(event.Content) < 20 {
		return nil
	}

	// Skip if no small model available
	if o.smallModel == nil {
		return nil
	}

	// Ask LLM to extract memorable content
	extracted, err := o.extractMemories(ctx, event)
	if err != nil {
		return fmt.Errorf("extract memories: %w", err)
	}

	// Store extracted memories
	for _, mem := range extracted {
		mem.SourceSessionID = event.SessionID
		mem.SourceMessageID = event.MessageID
		mem.AgentHandle = event.AgentHandle

		if _, err := o.memoryProvider.Create(ctx, mem); err != nil {
			slog.Debug("failed to store extracted memory",
				"content", util.Truncate(mem.Content, 50),
				"error", err)
		}
	}

	return nil
}

// extractMemories uses an LLM to identify memorable content.
func (o *Observer) extractMemories(ctx context.Context, event providers.MessageEvent) ([]providers.Memory, error) {
	prompt := fmt.Sprintf(`Analyze this conversation message and extract any memorable information.
Only extract if the message contains:
- User preferences (how they like things done)
- Facts about the user or their project
- Corrections to previous behavior
- Patterns in how they work

For each memory, output one line in format:
CATEGORY|CONTENT

Categories: preference, fact, correction, pattern

If nothing memorable, output: NONE

Message (%s):
%s

Output:`, event.Role, event.Content)

	response, err := o.smallModel.Complete(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return parseExtractionResponse(response), nil
}

// parseExtractionResponse parses the LLM's extraction response.
func parseExtractionResponse(response string) []providers.Memory {
	if response == "" || response == "NONE" {
		return nil
	}

	var memories []providers.Memory
	lines := splitLines(response)

	for _, line := range lines {
		line = trimSpace(line)
		if line == "" || line == "NONE" {
			continue
		}

		parts := splitOnce(line, "|")
		if len(parts) != 2 {
			continue
		}

		category := trimSpace(parts[0])
		content := trimSpace(parts[1])

		if content == "" {
			continue
		}

		// Validate category
		switch category {
		case "preference", "fact", "correction", "pattern":
			// Valid
		default:
			category = "fact" // Default to fact
		}

		memories = append(memories, providers.Memory{
			ID:         "mem-" + uuid.New().String()[:8],
			Content:    content,
			Category:   providers.MemoryCategory(category),
			Status:     providers.MemoryStatusActive,
			Confidence: 0.7, // Lower confidence for auto-extracted
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		})
	}

	return memories
}

// Helper functions


func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitOnce(s string, sep string) []string {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return []string{s[:i], s[i+len(sep):]}
		}
	}
	return []string{s}
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}

// Ensure Observer implements ObserverProvider
var _ providers.ObserverProvider = (*Observer)(nil)
