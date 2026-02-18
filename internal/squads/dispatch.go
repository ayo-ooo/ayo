// Package squads provides squad management for agent team coordination.
package squads

import (
	"context"
	"fmt"

	"charm.land/fantasy/schema"
)

// DispatchInput represents input for a squad invocation.
type DispatchInput struct {
	// Prompt is a free-form text prompt for the squad.
	Prompt string `json:"prompt,omitempty"`

	// Data contains structured input data.
	// If the squad has an input schema, this data is validated against it.
	Data map[string]any `json:"data,omitempty"`

	// TargetAgent overrides the default input routing.
	// If set, input is sent to this agent instead of input_accepts/lead.
	TargetAgent string `json:"target_agent,omitempty"`
}

// DispatchResult represents the result of a squad invocation.
type DispatchResult struct {
	// Output contains structured output data.
	// If the squad has an output schema, this data conforms to it.
	Output map[string]any `json:"output,omitempty"`

	// Raw is the raw text output if not structured.
	Raw string `json:"raw,omitempty"`

	// Error contains any error message from the squad.
	Error string `json:"error,omitempty"`

	// RoutedTo indicates which agent received the input.
	RoutedTo string `json:"routed_to,omitempty"`
}

// ValidationError is returned when input or output fails schema validation.
type ValidationError struct {
	Direction string // "input" or "output"
	Err       error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s validation failed: %v", e.Direction, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// ValidateInput validates the dispatch input against the squad's input schema.
// Returns nil if no schema is defined (free-form mode) or if validation passes.
func (s *Squad) ValidateInput(input DispatchInput) error {
	if s.Schemas == nil || s.Schemas.Input == nil {
		return nil // Free-form mode - any input accepted
	}

	// If Data is empty but we have a prompt, that's valid
	if input.Data == nil && input.Prompt != "" {
		return nil
	}

	// Validate data against schema
	if input.Data != nil {
		if err := schema.ValidateAgainstSchema(input.Data, *s.Schemas.Input); err != nil {
			return &ValidationError{Direction: "input", Err: err}
		}
	}

	return nil
}

// ValidateOutput validates the dispatch result against the squad's output schema.
// Returns nil if no schema is defined or if validation passes.
func (s *Squad) ValidateOutput(result *DispatchResult) error {
	if s.Schemas == nil || s.Schemas.Output == nil {
		return nil // Free-form mode - any output accepted
	}

	// If Output is empty but we have Raw, that's valid
	if result.Output == nil && result.Raw != "" {
		return nil
	}

	// Validate output against schema
	if result.Output != nil {
		if err := schema.ValidateAgainstSchema(result.Output, *s.Schemas.Output); err != nil {
			return &ValidationError{Direction: "output", Err: err}
		}
	}

	return nil
}

// GetTargetAgent determines which agent should receive the dispatch input.
// Priority:
//  1. Explicit TargetAgent in the input (command-line override)
//  2. InputAccepts from SQUAD.md frontmatter
//  3. Lead from SQUAD.md frontmatter
//  4. Default: @ayo
func (s *Squad) GetTargetAgent(input DispatchInput) string {
	// Explicit target in input takes priority
	if input.TargetAgent != "" {
		agent := input.TargetAgent
		if len(agent) > 0 && agent[0] != '@' {
			agent = "@" + agent
		}
		return agent
	}

	// Use constitution routing if available
	if s.Constitution != nil {
		return s.Constitution.Frontmatter.GetInputAcceptsAgent()
	}

	// Default to @ayo
	return "@ayo"
}

// Dispatch dispatches work to the squad after validating input.
// Input is routed to the appropriate agent based on:
// 1. Explicit TargetAgent in input
// 2. InputAccepts from SQUAD.md
// 3. Lead from SQUAD.md (default @ayo)
func (s *Squad) Dispatch(ctx context.Context, input DispatchInput) (*DispatchResult, error) {
	// Validate input
	if err := s.ValidateInput(input); err != nil {
		return nil, err
	}

	// Check if squad can accept input
	if !s.CanAcceptInput() {
		return nil, fmt.Errorf("squad %s is not ready to accept input (status: %s, lead_ready: %v)",
			s.Name, s.Status, s.LeadReady)
	}

	// Determine target agent for input routing
	targetAgent := s.GetTargetAgent(input)

	// TODO: Actual invocation of target agent will be added by am-2wpf (AgentInvoker)
	// For now, return result indicating successful routing
	return &DispatchResult{
		Raw:      fmt.Sprintf("dispatch to %s routed to %s (prompt: %s)", s.Name, targetAgent, input.Prompt),
		RoutedTo: targetAgent,
	}, nil
}

// DispatchOptions configures a dispatch operation.
type DispatchOptions struct {
	// StartIfStopped starts the squad if it's not running.
	StartIfStopped bool

	// Timeout is the maximum time to wait for a result.
	// If zero, a default timeout is used.
	Timeout int
}

// DispatchWithOptions dispatches work to the squad with additional options.
func (s *Squad) DispatchWithOptions(ctx context.Context, input DispatchInput, opts DispatchOptions) (*DispatchResult, error) {
	// Validate input first
	if err := s.ValidateInput(input); err != nil {
		return nil, err
	}

	// Check if we need to start the squad
	if !s.IsRunning() {
		if !opts.StartIfStopped {
			return nil, fmt.Errorf("squad %s is not running and StartIfStopped=false", s.Name)
		}
		// Note: actual start logic would go through daemon
	}

	return s.Dispatch(ctx, input)
}
