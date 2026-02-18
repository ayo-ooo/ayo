// Package run provides agent execution and task dispatch.
package run

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/alexcabrera/ayo/internal/capabilities"
)

// DispatchDecision represents the routing decision for a prompt.
// It indicates where @ayo should route the work.
type DispatchDecision struct {
	// Target is the handle of the target entity:
	// - "@ayo" means handle it directly
	// - "@agent" means delegate to a specific agent
	// - "#squad" means delegate to a squad
	Target string

	// Confidence is the embedding similarity score (0 to 1).
	// Higher values indicate stronger matches.
	Confidence float64

	// Reason explains why this routing decision was made.
	Reason string
}

// DispatchThreshold is the minimum similarity score required
// to route work to an agent or squad instead of handling directly.
const DispatchThreshold = 0.5

// TrivialPromptMaxLength is the maximum prompt length considered trivial.
// Very short prompts are likely simple questions or commands that
// @ayo can handle directly without delegation.
const TrivialPromptMaxLength = 100

// TrivialPromptMaxWords is the maximum word count for trivial prompts.
const TrivialPromptMaxWords = 15

// Dispatcher makes routing decisions using semantic search.
type Dispatcher struct {
	searcher *capabilities.UnifiedSearcher
}

// NewDispatcher creates a new dispatcher with the given searcher.
// If searcher is nil, all prompts will be handled by @ayo.
func NewDispatcher(searcher *capabilities.UnifiedSearcher) *Dispatcher {
	return &Dispatcher{
		searcher: searcher,
	}
}

// Decide determines where to route a prompt.
// Returns a DispatchDecision indicating the target and confidence.
func (d *Dispatcher) Decide(ctx context.Context, prompt string) (*DispatchDecision, error) {
	// No searcher means always handle directly
	if d.searcher == nil {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 1.0,
			Reason:     "no searcher available",
		}, nil
	}

	// Check if trivial (short prompt, likely a simple question)
	if isTrivialPrompt(prompt) {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 1.0,
			Reason:     "trivial task",
		}, nil
	}

	// Use semantic search to find best match
	result, err := d.searcher.FindBest(ctx, prompt)
	if err != nil {
		// On error, fall back to direct handling
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 0.0,
			Reason:     "search error: " + err.Error(),
		}, nil
	}

	// No matches in index
	if result == nil {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 1.0,
			Reason:     "no entities indexed",
		}, nil
	}

	// Check if match meets threshold
	if result.Score < DispatchThreshold {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: result.Score,
			Reason:     "no good match (best score: " + formatScore(result.Score) + ")",
		}, nil
	}

	// Good match found - delegate to the matched entity
	return &DispatchDecision{
		Target:     result.Handle,
		Confidence: result.Score,
		Reason:     "semantic match: " + result.Description,
	}, nil
}

// DecideAgentOnly determines where to route a prompt, considering only agents.
// Useful when you want to delegate to an agent but not to a squad.
func (d *Dispatcher) DecideAgentOnly(ctx context.Context, prompt string) (*DispatchDecision, error) {
	if d.searcher == nil {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 1.0,
			Reason:     "no searcher available",
		}, nil
	}

	if isTrivialPrompt(prompt) {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 1.0,
			Reason:     "trivial task",
		}, nil
	}

	results, err := d.searcher.SearchAgentsOnly(ctx, prompt, 1)
	if err != nil {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 0.0,
			Reason:     "search error: " + err.Error(),
		}, nil
	}

	if len(results) == 0 {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 1.0,
			Reason:     "no agents indexed",
		}, nil
	}

	result := results[0]
	if result.Score < DispatchThreshold {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: result.Score,
			Reason:     "no good agent match",
		}, nil
	}

	return &DispatchDecision{
		Target:     result.Handle,
		Confidence: result.Score,
		Reason:     "semantic match: " + result.Description,
	}, nil
}

// DecideSquadOnly determines where to route a prompt, considering only squads.
// Useful when you want to delegate to a squad but not to an individual agent.
func (d *Dispatcher) DecideSquadOnly(ctx context.Context, prompt string) (*DispatchDecision, error) {
	if d.searcher == nil {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 1.0,
			Reason:     "no searcher available",
		}, nil
	}

	if isTrivialPrompt(prompt) {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 1.0,
			Reason:     "trivial task",
		}, nil
	}

	results, err := d.searcher.SearchSquadsOnly(ctx, prompt, 1)
	if err != nil {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 0.0,
			Reason:     "search error: " + err.Error(),
		}, nil
	}

	if len(results) == 0 {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: 1.0,
			Reason:     "no squads indexed",
		}, nil
	}

	result := results[0]
	if result.Score < DispatchThreshold {
		return &DispatchDecision{
			Target:     "@ayo",
			Confidence: result.Score,
			Reason:     "no good squad match",
		}, nil
	}

	return &DispatchDecision{
		Target:     result.Handle,
		Confidence: result.Score,
		Reason:     "semantic match: " + result.Description,
	}, nil
}

// isTrivialPrompt returns true if the prompt is likely trivial.
// Trivial prompts are short, don't contain domain-specific keywords,
// and can likely be handled directly by @ayo.
func isTrivialPrompt(prompt string) bool {
	// Normalize whitespace
	prompt = strings.TrimSpace(prompt)

	// Empty prompts are trivial
	if prompt == "" {
		return true
	}

	// Very short prompts are trivial
	if len(prompt) <= TrivialPromptMaxLength {
		words := countWords(prompt)
		if words <= TrivialPromptMaxWords {
			return true
		}
	}

	return false
}

// countWords counts the number of words in a string.
func countWords(s string) int {
	inWord := false
	count := 0
	for _, r := range s {
		if unicode.IsSpace(r) {
			if inWord {
				inWord = false
			}
		} else {
			if !inWord {
				inWord = true
				count++
			}
		}
	}
	return count
}

// formatScore formats a float64 score as a percentage string.
func formatScore(score float64) string {
	pct := int(score*100 + 0.5)
	return fmt.Sprintf("%d%%", pct)
}
