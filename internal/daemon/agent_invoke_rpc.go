// Package daemon provides the background daemon service.
package daemon

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/run"
)

// handleAgentInvoke handles the agent.invoke RPC method.
func (s *Server) handleAgentInvoke(ctx context.Context, req *Request) *Response {
	var params AgentInvokeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "invalid params: "+err.Error()), req.ID)
	}

	// Validate required fields
	if params.Agent == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "agent is required"), req.ID)
	}
	if params.Prompt == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "prompt is required"), req.ID)
	}

	// Normalize agent handle
	agentHandle := params.Agent
	if len(agentHandle) > 0 && agentHandle[0] != '@' {
		agentHandle = "@" + agentHandle
	}

	// Load the agent
	ag, err := agent.Load(s.config, agentHandle)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, fmt.Sprintf("load agent %s: %v", agentHandle, err)), req.ID)
	}

	// Add any requested skills to the agent's config before execution
	if len(params.Skills) > 0 {
		ag.Config.Skills = append(ag.Config.Skills, params.Skills...)
	}

	// Create runner
	runner, err := run.NewRunnerFromConfig(s.config, false)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, fmt.Sprintf("create runner: %v", err)), req.ID)
	}

	// Execute the agent
	result, err := runner.TextWithSession(ctx, ag, params.Prompt, nil)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, fmt.Sprintf("invoke agent: %v", err)), req.ID)
	}

	// Use result session ID or provided session ID
	sessionID := result.SessionID
	if sessionID == "" && params.SessionID != "" {
		sessionID = params.SessionID
	}

	invokeResult := AgentInvokeResult{
		SessionID: sessionID,
		Response:  result.Response,
	}

	resp, _ := NewResponse(invokeResult, req.ID)
	return resp
}
