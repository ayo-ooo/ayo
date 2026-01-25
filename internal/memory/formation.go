package memory

import (
	"context"
	"sync"
	"time"
)

// FormationIntent represents an intent to form a memory.
type FormationIntent struct {
	ID            string
	Content       string
	Category      Category
	AgentHandle   string
	PathScope     string
	SourceSession string
	SourceMessage string
}

// FormationResult represents the result of memory formation.
type FormationResult struct {
	Intent       FormationIntent
	Memory       Memory
	Success      bool
	Error        error
	Elapsed      time.Duration
	Deduplicated bool    // True if an exact duplicate was found (no new memory created)
	Superseded   bool    // True if a similar memory was superseded
	SupersededID string  // ID of the superseded memory (if any)
	Skipped      bool    // True if skipped due to existing memory
	SkipReason   string  // Reason for skipping (e.g., "already remembered")
	Failed       bool    // True if creation failed
	FailReason   string  // Reason for failure
}

// FormationEventType describes what happened during memory formation.
type FormationEventType string

const (
	FormationEventCreated    FormationEventType = "created"    // New memory created
	FormationEventSkipped    FormationEventType = "skipped"    // Already exists, skipped
	FormationEventSuperseded FormationEventType = "superseded" // Old memory replaced
	FormationEventFailed     FormationEventType = "failed"     // Creation failed
)

// EventType returns the type of formation event.
func (r FormationResult) EventType() FormationEventType {
	if r.Failed || r.Error != nil {
		return FormationEventFailed
	}
	if r.Skipped || r.Deduplicated {
		return FormationEventSkipped
	}
	if r.Superseded {
		return FormationEventSuperseded
	}
	return FormationEventCreated
}

// FormationCallback is called when memory formation completes.
type FormationCallback func(result FormationResult)

// FormationService handles async memory formation.
type FormationService struct {
	svc       *Service
	callbacks []FormationCallback
	mu        sync.RWMutex
	pending   map[string]FormationIntent
}

// NewFormationService creates a new formation service.
func NewFormationService(svc *Service) *FormationService {
	return &FormationService{
		svc:     svc,
		pending: make(map[string]FormationIntent),
	}
}

// OnFormation registers a callback for formation events.
func (f *FormationService) OnFormation(cb FormationCallback) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callbacks = append(f.callbacks, cb)
}

// QueueFormation queues a memory for async formation.
// Returns the intent ID immediately.
func (f *FormationService) QueueFormation(ctx context.Context, intent FormationIntent) string {
	if intent.ID == "" {
		intent.ID = generateID()
	}

	f.mu.Lock()
	f.pending[intent.ID] = intent
	f.mu.Unlock()

	// Start formation in background
	go f.processFormation(ctx, intent)

	return intent.ID
}

// GetPending returns the number of pending formations.
func (f *FormationService) GetPending() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.pending)
}

// IsPending checks if a formation is still pending.
func (f *FormationService) IsPending(id string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.pending[id]
	return ok
}

// Wait blocks until all pending formations complete or timeout expires.
func (f *FormationService) Wait(timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		f.mu.RLock()
		pending := len(f.pending)
		f.mu.RUnlock()
		if pending == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// NotifyCreated notifies callbacks that a memory was created (for synchronous creation).
func (f *FormationService) NotifyCreated(mem Memory) {
	if f == nil {
		return
	}
	f.notify(FormationResult{
		Memory:  mem,
		Success: true,
	})
}

// NotifySkipped notifies callbacks that a memory was skipped (already exists).
func (f *FormationService) NotifySkipped(content string, existingID string) {
	if f == nil {
		return
	}
	f.notify(FormationResult{
		Memory:     Memory{Content: content, ID: existingID},
		Success:    true,
		Skipped:    true,
		SkipReason: "already remembered",
	})
}

// NotifySuperseded notifies callbacks that a memory was superseded.
func (f *FormationService) NotifySuperseded(mem Memory, oldID string) {
	if f == nil {
		return
	}
	f.notify(FormationResult{
		Memory:       mem,
		Success:      true,
		Superseded:   true,
		SupersededID: oldID,
	})
}

// NotifyFailed notifies callbacks that memory creation failed.
func (f *FormationService) NotifyFailed(content string, err error) {
	if f == nil {
		return
	}
	f.notify(FormationResult{
		Memory:     Memory{Content: content},
		Success:    false,
		Failed:     true,
		FailReason: err.Error(),
		Error:      err,
	})
}

func (f *FormationService) notify(result FormationResult) {
	f.mu.RLock()
	callbacks := make([]FormationCallback, len(f.callbacks))
	copy(callbacks, f.callbacks)
	f.mu.RUnlock()

	for _, cb := range callbacks {
		cb(result)
	}
}

// Similarity thresholds for deduplication
const (
	// ExactDuplicateThreshold: memories above this are considered exact duplicates (skip creation)
	ExactDuplicateThreshold float32 = 0.95
	// SupersedeThreshold: memories above this but below exact are superseded
	SupersedeThreshold float32 = 0.85
)

func (f *FormationService) processFormation(ctx context.Context, intent FormationIntent) {
	start := time.Now()

	result := FormationResult{
		Intent: intent,
	}

	// Secondary deduplication check (primary check happens before queueing in maybeFormMemory).
	// This catches race conditions and handles the supersede logic for similar-but-not-exact matches.
	existing, err := f.svc.Search(ctx, intent.Content, SearchOptions{
		AgentHandle: intent.AgentHandle,
		PathScope:   intent.PathScope,
		Threshold:   SupersedeThreshold,
		Limit:       1,
	})
	if err != nil {
		// Log but continue with creation
		existing = nil
	}

	var mem Memory
	if len(existing) > 0 && existing[0].Similarity >= ExactDuplicateThreshold {
		// Exact duplicate - skip creation, return existing memory
		mem = existing[0].Memory
		result.Deduplicated = true
	} else if len(existing) > 0 && existing[0].Similarity >= SupersedeThreshold {
		// Similar memory exists - supersede it
		mem, err = f.svc.Supersede(ctx, existing[0].Memory.ID, Memory{
			Content:         intent.Content,
			Category:        intent.Category,
			AgentHandle:     intent.AgentHandle,
			PathScope:       intent.PathScope,
			SourceSessionID: intent.SourceSession,
			SourceMessageID: intent.SourceMessage,
		}, "updated via memory formation")
		result.Superseded = true
		result.SupersededID = existing[0].Memory.ID
	} else {
		// No duplicate - create new memory
		mem, err = f.svc.Create(ctx, Memory{
			Content:         intent.Content,
			Category:        intent.Category,
			AgentHandle:     intent.AgentHandle,
			PathScope:       intent.PathScope,
			SourceSessionID: intent.SourceSession,
			SourceMessageID: intent.SourceMessage,
		})
	}

	result.Elapsed = time.Since(start)

	if err != nil {
		result.Error = err
		result.Success = false
	} else {
		result.Memory = mem
		result.Success = true
	}

	// Remove from pending
	f.mu.Lock()
	delete(f.pending, intent.ID)
	callbacks := make([]FormationCallback, len(f.callbacks))
	copy(callbacks, f.callbacks)
	f.mu.Unlock()

	// Notify callbacks
	for _, cb := range callbacks {
		cb(result)
	}
}

func generateID() string {
	return time.Now().Format("20060102150405.000000")
}
