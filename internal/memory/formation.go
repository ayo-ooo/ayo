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
	Deduplicated bool   // True if an exact duplicate was found (no new memory created)
	Superseded   bool   // True if a similar memory was superseded
	SupersededID string // ID of the superseded memory (if any)
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

	// Check for duplicate/similar memories before creating
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

// TriggerType represents what triggered memory formation.
type TriggerType string

const (
	TriggerExplicit   TriggerType = "explicit"   // User explicitly asked to remember
	TriggerCorrection TriggerType = "correction" // User corrected agent behavior
	TriggerPreference TriggerType = "preference" // User expressed a preference
	TriggerFact       TriggerType = "fact"       // Agent learned a fact about project/user
)

// FormationTrigger represents a detected trigger for memory formation.
type FormationTrigger struct {
	Type        TriggerType
	Content     string
	Category    Category
	Confidence  float64
	SourceText  string
}

// DetectTriggers analyzes a message for potential memory triggers.
// This is a simple heuristic-based implementation.
func DetectTriggers(message string) []FormationTrigger {
	var triggers []FormationTrigger

	// Explicit remember requests
	explicitPhrases := []string{
		"remember that",
		"remember this",
		"keep in mind",
		"note that",
		"don't forget",
		"always remember",
	}

	lowerMsg := toLower(message)
	for _, phrase := range explicitPhrases {
		if containsPhrase(lowerMsg, phrase) {
			triggers = append(triggers, FormationTrigger{
				Type:       TriggerExplicit,
				Content:    message,
				Category:   CategoryFact,
				Confidence: 0.9,
				SourceText: message,
			})
			break
		}
	}

	// Preference expressions
	preferencePhrases := []string{
		"i prefer",
		"i like",
		"i want",
		"always use",
		"never use",
		"don't use",
		"use ... instead",
	}

	for _, phrase := range preferencePhrases {
		if containsPhrase(lowerMsg, phrase) {
			triggers = append(triggers, FormationTrigger{
				Type:       TriggerPreference,
				Content:    message,
				Category:   CategoryPreference,
				Confidence: 0.7,
				SourceText: message,
			})
			break
		}
	}

	// Correction patterns
	correctionPhrases := []string{
		"no, ",
		"actually,",
		"that's wrong",
		"that's not right",
		"you should",
		"you shouldn't",
		"don't do that",
		"instead of",
	}

	for _, phrase := range correctionPhrases {
		if containsPhrase(lowerMsg, phrase) {
			triggers = append(triggers, FormationTrigger{
				Type:       TriggerCorrection,
				Content:    message,
				Category:   CategoryCorrection,
				Confidence: 0.8,
				SourceText: message,
			})
			break
		}
	}

	return triggers
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func containsPhrase(text, phrase string) bool {
	if len(phrase) > len(text) {
		return false
	}
	for i := 0; i <= len(text)-len(phrase); i++ {
		if text[i:i+len(phrase)] == phrase {
			return true
		}
	}
	return false
}
