package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/alexcabrera/ayo/internal/ui"
)

// QueueRequest represents a memory storage request.
type QueueRequest struct {
	ID          string   // Unique request ID
	Content     string   // Memory content
	Category    Category // Memory category
	AgentHandle string   // Optional agent scope
	PathScope   string   // Optional path scope
}

// Queue manages async memory storage operations.
type Queue struct {
	service    *Service
	requests   chan QueueRequest
	onStatus   func(ui.AsyncStatusMsg)
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	bufferSize int
}

// QueueConfig configures the memory queue.
type QueueConfig struct {
	BufferSize int                       // Channel buffer size (default: 100)
	OnStatus   func(ui.AsyncStatusMsg)   // Status callback
}

// NewQueue creates a new memory queue.
func NewQueue(service *Service, cfg QueueConfig) *Queue {
	bufferSize := cfg.BufferSize
	if bufferSize <= 0 {
		bufferSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Queue{
		service:    service,
		requests:   make(chan QueueRequest, bufferSize),
		onStatus:   cfg.OnStatus,
		ctx:        ctx,
		cancel:     cancel,
		bufferSize: bufferSize,
	}
}

// Enqueue adds a memory request to the queue and returns the request ID.
// Returns immediately without blocking on the actual storage.
func (q *Queue) Enqueue(content string, category Category, agentHandle, pathScope string) string {
	id := uuid.New().String()[:8]

	req := QueueRequest{
		ID:          id,
		Content:     content,
		Category:    category,
		AgentHandle: agentHandle,
		PathScope:   pathScope,
	}

	// Send status: pending
	q.sendStatus(id, ui.AsyncStatusPending, "Memory queued")

	// Non-blocking send - if buffer is full, drop the request
	select {
	case q.requests <- req:
		// Successfully queued
	default:
		// Queue full - send failure status
		q.sendStatus(id, ui.AsyncStatusFailed, "Memory queue full")
	}

	return id
}

// Start spawns the worker goroutine that processes memory requests.
func (q *Queue) Start() {
	q.wg.Add(1)
	go q.worker()
}

// Stop gracefully stops the queue, waiting for pending items to complete.
func (q *Queue) Stop(timeout time.Duration) {
	// Signal shutdown
	q.cancel()

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Clean shutdown
	case <-time.After(timeout):
		// Timeout - some items may be lost
	}
}

// Pending returns the approximate number of pending requests.
func (q *Queue) Pending() int {
	return len(q.requests)
}

// worker processes memory requests from the queue.
func (q *Queue) worker() {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			// Drain remaining items before exit
			q.drain()
			return

		case req := <-q.requests:
			q.processRequest(req)
		}
	}
}

// drain processes any remaining requests in the buffer.
func (q *Queue) drain() {
	for {
		select {
		case req := <-q.requests:
			q.processRequest(req)
		default:
			return
		}
	}
}

// processRequest handles a single memory request.
func (q *Queue) processRequest(req QueueRequest) {
	// Guard against nil service (shouldn't happen in production)
	if q.service == nil {
		q.sendStatus(req.ID, ui.AsyncStatusFailed, "Memory service not available")
		return
	}

	// Send status: in progress
	q.sendStatus(req.ID, ui.AsyncStatusInProgress, "Storing memory...")

	// Create the memory (this includes embedding generation)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := q.service.Create(ctx, Memory{
		Content:     req.Content,
		Category:    req.Category,
		AgentHandle: req.AgentHandle,
		PathScope:   req.PathScope,
	})

	if err != nil {
		q.sendStatus(req.ID, ui.AsyncStatusFailed, "Failed: "+err.Error())
		return
	}

	q.sendStatus(req.ID, ui.AsyncStatusCompleted, "Memory stored")
}

// sendStatus sends a status update if callback is configured.
func (q *Queue) sendStatus(id string, status ui.AsyncStatus, message string) {
	if q.onStatus != nil {
		q.onStatus(ui.AsyncStatusMsg{
			ID:        id,
			Operation: ui.AsyncOpMemoryStore,
			Status:    status,
			Message:   message,
		})
	}
}
