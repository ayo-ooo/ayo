// Package routing provides intelligent task routing for @ayo.
package routing

import (
	"context"
	"regexp"
	"strings"
	"unicode"

	"github.com/alexcabrera/ayo/internal/capabilities"
)

// TargetType represents the type of routing target.
type TargetType int

const (
	// Self means @ayo handles the task directly.
	Self TargetType = iota
	// Agent means route to a specialist agent.
	Agent
	// Squad means route to a multi-agent squad.
	Squad
)

// String returns a string representation of the target type.
func (t TargetType) String() string {
	switch t {
	case Self:
		return "self"
	case Agent:
		return "agent"
	case Squad:
		return "squad"
	default:
		return "unknown"
	}
}

// Thresholds for routing decisions.
const (
	// SquadThreshold is the minimum similarity score to route to a squad.
	// Higher than agent threshold because squads involve coordination overhead.
	SquadThreshold = 0.7

	// AgentThreshold is the minimum similarity score to route to an agent.
	AgentThreshold = 0.6

	// TrivialMaxLength is the maximum prompt length considered trivial.
	TrivialMaxLength = 100

	// TrivialMaxWords is the maximum word count for trivial prompts.
	TrivialMaxWords = 15
)

// RoutingDecision represents where to route a task.
type RoutingDecision struct {
	// Target is the handle of the target entity.
	// Empty for Self, "@agent" for Agent, "#squad" for Squad.
	Target string

	// TargetType indicates the category of target.
	TargetType TargetType

	// Confidence is the similarity score (0 to 1) for semantic matches.
	// 1.0 for explicit targeting or trivial prompts.
	Confidence float64

	// Reason explains why this routing decision was made.
	Reason string
}

// Router makes intelligent routing decisions for @ayo.
type Router struct {
	searcher *capabilities.UnifiedSearcher
}

// NewRouter creates a new router with the given searcher.
// If searcher is nil, all prompts will be handled by @ayo directly.
func NewRouter(searcher *capabilities.UnifiedSearcher) *Router {
	return &Router{
		searcher: searcher,
	}
}

// Decide determines where to route a prompt.
// The decision flow is:
// 1. Explicit targeting (@agent or #squad) → route to target
// 2. Trivial input → handle directly
// 3. Matching squad (>0.7 similarity) → route to squad
// 4. Matching agent (>0.6 similarity) → route to agent
// 5. Fallback → handle directly
func (r *Router) Decide(ctx context.Context, prompt string) (*RoutingDecision, error) {
	prompt = strings.TrimSpace(prompt)

	// 1. Check for explicit targeting (overrides everything)
	if target := parseExplicitTarget(prompt); target != "" {
		targetType := Agent
		if strings.HasPrefix(target, "#") {
			targetType = Squad
		}
		return &RoutingDecision{
			Target:     target,
			TargetType: targetType,
			Confidence: 1.0,
			Reason:     "explicit target",
		}, nil
	}

	// 2. Check for trivial input
	if isTrivial(prompt) {
		return &RoutingDecision{
			TargetType: Self,
			Confidence: 1.0,
			Reason:     "trivial input",
		}, nil
	}

	// 3-4. Semantic search for squads and agents
	if r.searcher == nil {
		return &RoutingDecision{
			TargetType: Self,
			Confidence: 1.0,
			Reason:     "no searcher available",
		}, nil
	}

	// 3. Search for matching squad first (squads are for coordination)
	squadResults, err := r.searcher.SearchSquadsOnly(ctx, prompt, 1)
	if err == nil && len(squadResults) > 0 && squadResults[0].Score >= SquadThreshold {
		return &RoutingDecision{
			Target:     squadResults[0].Handle,
			TargetType: Squad,
			Confidence: squadResults[0].Score,
			Reason:     "matched squad mission: " + squadResults[0].Description,
		}, nil
	}

	// 4. Search for matching agent
	agentResults, err := r.searcher.SearchAgentsOnly(ctx, prompt, 1)
	if err == nil && len(agentResults) > 0 && agentResults[0].Score >= AgentThreshold {
		return &RoutingDecision{
			Target:     agentResults[0].Handle,
			TargetType: Agent,
			Confidence: agentResults[0].Score,
			Reason:     "matched agent capability: " + agentResults[0].Description,
		}, nil
	}

	// 5. Fallback to direct handling
	return &RoutingDecision{
		TargetType: Self,
		Confidence: 1.0,
		Reason:     "no better match found",
	}, nil
}

// isTrivial returns true if the prompt is trivial enough to handle directly.
func isTrivial(prompt string) bool {
	if prompt == "" {
		return true
	}

	// Very short prompts are trivial
	if len(prompt) <= TrivialMaxLength && countWords(prompt) <= TrivialMaxWords {
		return true
	}

	return false
}

// Regex patterns for explicit targeting.
var (
	// Matches @handle at the start of the prompt
	agentPattern = regexp.MustCompile(`^@([a-zA-Z][a-zA-Z0-9._-]*)`)
	// Matches #squad at the start of the prompt
	squadPattern = regexp.MustCompile(`^#([a-zA-Z][a-zA-Z0-9._-]*)`)
)

// parseExplicitTarget extracts an explicit @agent or #squad target from the prompt.
// Returns empty string if no explicit target is found.
func parseExplicitTarget(prompt string) string {
	prompt = strings.TrimSpace(prompt)

	// Check for @agent
	if match := agentPattern.FindStringSubmatch(prompt); len(match) > 1 {
		return "@" + match[1]
	}

	// Check for #squad
	if match := squadPattern.FindStringSubmatch(prompt); len(match) > 1 {
		return "#" + match[1]
	}

	return ""
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
