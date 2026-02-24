// Package humaninput provides the human_input tool for requesting structured input from humans.
package humaninput

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/hitl"
)

// FieldParams defines a single field for human input.
type FieldParams struct {
	Name        string   `json:"name" jsonschema:"required,description=Unique identifier for this field"`
	Type        string   `json:"type" jsonschema:"required,enum=text,enum=select,enum=multiselect,enum=confirm,enum=number,enum=textarea,description=Field type: text, select, multiselect, confirm, number, textarea"`
	Label       string   `json:"label" jsonschema:"required,description=Human-readable label shown to user"`
	Description string   `json:"description,omitempty" jsonschema:"description=Additional help text for the field"`
	Required    bool     `json:"required,omitempty" jsonschema:"description=Whether this field is required"`
	Options     []string `json:"options,omitempty" jsonschema:"description=Options for select/multiselect fields"`
	Default     string   `json:"default,omitempty" jsonschema:"description=Default value for the field"`
}

// HumanInputParams defines parameters for the human_input tool.
type HumanInputParams struct {
	Context   string        `json:"context" jsonschema:"required,description=Brief explanation of why you need this input"`
	Fields    []FieldParams `json:"fields" jsonschema:"required,description=Fields to collect from the human"`
	Recipient string        `json:"recipient,omitempty" jsonschema:"description=Who to ask: 'owner' (default) or email address"`
	Timeout   string        `json:"timeout,omitempty" jsonschema:"description=How long to wait e.g. '5m', '1h', '24h'. Default is 1h."`
}

// FormRenderer is called to present the form to the user.
// This is a type alias for hitl.FormRenderer to maintain API compatibility.
type FormRenderer = hitl.FormRenderer

// ToolConfig configures the human_input tool.
type ToolConfig struct {
	// Renderer presents forms to users. If nil, returns error.
	Renderer FormRenderer
	// DefaultTimeout is the timeout when none specified.
	DefaultTimeout time.Duration
}

// NewHumanInputTool creates a new human_input tool for requesting human input.
func NewHumanInputTool(cfg ToolConfig) fantasy.AgentTool {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = time.Hour
	}

	return fantasy.NewAgentTool(
		"human_input",
		"Request structured input from a human. Use when you need approval, clarification, or information only a human can provide.",
		func(ctx context.Context, params HumanInputParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if cfg.Renderer == nil {
				return fantasy.NewTextErrorResponse("human_input tool not available in this context"), nil
			}

			// Build InputRequest from params
			req, err := buildInputRequest(params, cfg.DefaultTimeout)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid input request: %v", err)), nil
			}

			// Validate the request
			if err := hitl.ValidateRequest(req); err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid input request: %v", err)), nil
			}

			// Render the form and block until response
			resp, err := cfg.Renderer.Render(ctx, req)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to get human input: %v", err)), nil
			}

			if resp.Skipped {
				return fantasy.NewTextResponse("The user skipped this input request."), nil
			}

			// Return values as JSON-formatted response
			jsonData, err := json.Marshal(resp.Values)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to serialize response: %v", err)), nil
			}
			return fantasy.NewTextResponse(string(jsonData)), nil
		},
	)
}

// buildInputRequest converts tool params to an InputRequest.
func buildInputRequest(params HumanInputParams, defaultTimeout time.Duration) (*hitl.InputRequest, error) {
	// Parse timeout
	timeout := defaultTimeout
	if params.Timeout != "" {
		var err error
		timeout, err = time.ParseDuration(params.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout: %w", err)
		}
	}

	// Parse recipient
	recipientType := hitl.RecipientOwner
	recipientAddr := ""
	if params.Recipient != "" && params.Recipient != "owner" {
		recipientType = hitl.RecipientEmail
		recipientAddr = params.Recipient
	}

	// Build fields
	fields := make([]hitl.Field, len(params.Fields))
	for i, f := range params.Fields {
		fieldType := hitl.FieldType(f.Type)
		
		// Build options for select/multiselect
		var options []hitl.Option
		if len(f.Options) > 0 {
			options = make([]hitl.Option, len(f.Options))
			for j, opt := range f.Options {
				options[j] = hitl.Option{Value: opt, Label: opt}
			}
		}

		// Parse default value
		var defaultVal any
		if f.Default != "" {
			defaultVal = f.Default
		}

		fields[i] = hitl.Field{
			Name:        f.Name,
			Type:        fieldType,
			Label:       f.Label,
			Description: f.Description,
			Required:    f.Required,
			Default:     defaultVal,
			Options:     options,
		}
	}

	return &hitl.InputRequest{
		ID:      fmt.Sprintf("input-%d", time.Now().UnixNano()),
		Timeout: timeout,
		Recipient: hitl.Recipient{
			Type:    recipientType,
			Address: recipientAddr,
		},
		Context: params.Context,
		Fields:  fields,
	}, nil
}
