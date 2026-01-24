// Package smallmodel provides services for using small LLMs for internal operations
// like memory extraction, deduplication, and title generation.
package smallmodel

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alexcabrera/ayo/internal/ollama"
)

// Service provides small model operations for ayo internals.
type Service struct {
	client *ollama.Client
	model  string
}

// Config configures the small model service.
type Config struct {
	Host  string // Ollama host (default: http://localhost:11434)
	Model string // Model to use (default: ministral-3:3b)
}

// NewService creates a new small model service.
func NewService(cfg Config) *Service {
	opts := []ollama.Option{}
	if cfg.Host != "" {
		opts = append(opts, ollama.WithHost(cfg.Host))
	}

	model := cfg.Model
	if model == "" {
		model = ollama.DefaultModel
	}

	return &Service{
		client: ollama.NewClient(opts...),
		model:  model,
	}
}

// IsAvailable checks if the small model service is available.
func (s *Service) IsAvailable(ctx context.Context) bool {
	return s.client.IsAvailable(ctx)
}

// MemoryExtraction represents the result of extracting memorable content.
type MemoryExtraction struct {
	ShouldRemember bool   `json:"should_remember"`
	Content        string `json:"content,omitempty"`
	Category       string `json:"category,omitempty"` // preference, fact, correction
	Confidence     float64 `json:"confidence,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

const memoryExtractionPrompt = `Analyze this user message and determine if it contains information worth remembering for future conversations.

Look for:
- Explicit requests to remember ("remember that...", "keep in mind...", "note that...")
- Preferences ("I prefer...", "I like...", "always use...", "never use...", "don't use...")
- Facts about the user ("I work at...", "my project uses...", "I'm building...")
- Corrections ("actually...", "no, I meant...", "that's wrong...", "I said...")

If memorable, extract the CORE information in THIRD PERSON (e.g., "User prefers TypeScript" not "I prefer TypeScript").

Respond with valid JSON only:
{"should_remember": true/false, "content": "distilled memory in third person", "category": "preference|fact|correction", "confidence": 0.0-1.0, "reason": "why this is memorable"}

User message: %s`

// ExtractMemory analyzes a user message for memorable content.
func (s *Service) ExtractMemory(ctx context.Context, userMessage string) (*MemoryExtraction, error) {
	prompt := fmt.Sprintf(memoryExtractionPrompt, userMessage)

	result, err := s.client.ChatJSON(ctx, s.model, []ollama.Message{
		{Role: "user", Content: prompt},
	}, &ollama.Options{
		Temperature: 0.1, // Low temperature for consistent extraction
	})
	if err != nil {
		return nil, fmt.Errorf("extract memory: %w", err)
	}

	var extraction MemoryExtraction
	if err := json.Unmarshal(result, &extraction); err != nil {
		return nil, fmt.Errorf("parse extraction: %w", err)
	}

	return &extraction, nil
}

// DedupDecision represents the result of comparing memories.
type DedupDecision struct {
	Action   string `json:"action"`   // "new", "duplicate", "supersede"
	Reason   string `json:"reason"`
	TargetID string `json:"target_id,omitempty"` // ID of memory to supersede (if action=supersede)
}

const dedupPrompt = `Compare a new memory against existing memories and decide what to do.

New memory: %s

Existing memories:
%s

Decide:
- "new": The new memory is genuinely new information
- "duplicate": The new memory is essentially the same as an existing one (skip it)  
- "supersede": The new memory updates/replaces an existing one (mark old as superseded)

Respond with valid JSON only:
{"action": "new|duplicate|supersede", "reason": "explanation", "target_id": "id of memory to supersede if action=supersede"}`

// ExistingMemory represents a memory for deduplication comparison.
type ExistingMemory struct {
	ID      string
	Content string
}

// CheckDuplicate compares a new memory against existing memories.
func (s *Service) CheckDuplicate(ctx context.Context, newContent string, existing []ExistingMemory) (*DedupDecision, error) {
	if len(existing) == 0 {
		return &DedupDecision{Action: "new", Reason: "no existing memories"}, nil
	}

	// Format existing memories
	var existingStr strings.Builder
	for _, m := range existing {
		fmt.Fprintf(&existingStr, "- [%s] %s\n", m.ID[:8], m.Content)
	}

	prompt := fmt.Sprintf(dedupPrompt, newContent, existingStr.String())

	result, err := s.client.ChatJSON(ctx, s.model, []ollama.Message{
		{Role: "user", Content: prompt},
	}, &ollama.Options{
		Temperature: 0.1,
	})
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}

	var decision DedupDecision
	if err := json.Unmarshal(result, &decision); err != nil {
		return nil, fmt.Errorf("parse decision: %w", err)
	}

	return &decision, nil
}

const titlePrompt = `Generate a short, descriptive title for this conversation based on the first message.

The title should be:
- 3-6 words maximum
- Descriptive of the topic/intent
- No quotes or punctuation at the end

First message: %s

Respond with just the title, nothing else.`

// GenerateTitle generates a session title from the first message.
func (s *Service) GenerateTitle(ctx context.Context, firstMessage string) (string, error) {
	prompt := fmt.Sprintf(titlePrompt, firstMessage)

	resp, err := s.client.Chat(ctx, s.model, []ollama.Message{
		{Role: "user", Content: prompt},
	}, &ollama.Options{
		Temperature: 0.3,
		NumPredict:  20, // Short output
	})
	if err != nil {
		return "", fmt.Errorf("generate title: %w", err)
	}

	title := strings.TrimSpace(resp.Message.Content)
	// Remove any quotes that might have been added
	title = strings.Trim(title, "\"'")
	
	return title, nil
}

// Model returns the model name being used.
func (s *Service) Model() string {
	return s.model
}

// CategoryResult represents the result of categorizing content.
type CategoryResult struct {
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
}

const categorizePrompt = `Categorize this piece of information into exactly one category.

Categories:
- "preference": User preferences, likes, dislikes, style choices (e.g., "User prefers TypeScript", "User always uses tabs")
- "fact": Facts about the user, project, or environment (e.g., "User works at Acme", "Project uses PostgreSQL")
- "correction": Corrections to previous agent behavior (e.g., "that's wrong", "don't do that again")
- "pattern": Observed patterns in user behavior (e.g., "User usually asks for tests", "User likes verbose output")

Content: %s

Respond with valid JSON only:
{"category": "preference|fact|correction|pattern", "confidence": 0.0-1.0}`

// CategorizeMemory determines the category for a piece of content.
func (s *Service) CategorizeMemory(ctx context.Context, content string) (*CategoryResult, error) {
	prompt := fmt.Sprintf(categorizePrompt, content)

	result, err := s.client.ChatJSON(ctx, s.model, []ollama.Message{
		{Role: "user", Content: prompt},
	}, &ollama.Options{
		Temperature: 0.1,
	})
	if err != nil {
		return nil, fmt.Errorf("categorize memory: %w", err)
	}

	var cat CategoryResult
	if err := json.Unmarshal(result, &cat); err != nil {
		return nil, fmt.Errorf("parse category: %w", err)
	}

	// Validate category
	switch cat.Category {
	case "preference", "fact", "correction", "pattern":
		// valid
	default:
		cat.Category = "fact" // default fallback
	}

	return &cat, nil
}
