// Package daemon provides flow execution RPC handlers.
package daemon

import (
	"context"
	"encoding/json"
	"time"

	"github.com/alexcabrera/ayo/internal/flows"
	"github.com/alexcabrera/ayo/internal/paths"
)

// Application-specific error codes for flows
const (
	ErrCodeFlowNotFound = -2001
	ErrCodeFlowInvalid  = -2002
)

// handleFlowRun handles the flow.run RPC method.
func (s *Server) handleFlowRun(ctx context.Context, req *Request) *Response {
	var params FlowRunParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "invalid params"), req.ID)
	}

	if params.FlowName == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "flow_name is required"), req.ID)
	}

	// Discover flows
	dirs := paths.FlowsDirs()

	// First try YAML flows
	yamlFlows, err := flows.DiscoverYAMLFlows(dirs[0])
	if err == nil && len(yamlFlows) > 0 {
		for _, yf := range yamlFlows {
			if yf.Name == params.FlowName {
				return s.runYAMLFlow(ctx, req, yf, params)
			}
		}
	}

	// Try all directories for YAML flows
	for _, dir := range dirs[1:] {
		yamlFlows, err := flows.DiscoverYAMLFlows(dir)
		if err != nil {
			continue
		}
		for _, yf := range yamlFlows {
			if yf.Name == params.FlowName {
				return s.runYAMLFlow(ctx, req, yf, params)
			}
		}
	}

	// Try shell script flows
	discovered, err := flows.Discover(dirs)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, "discover flows: "+err.Error()), req.ID)
	}

	for _, f := range discovered {
		if f.Name == params.FlowName {
			return s.runShellFlow(ctx, req, &f, params)
		}
	}

	return NewErrorResponse(NewError(ErrCodeFlowNotFound, "flow not found: "+params.FlowName), req.ID)
}

// runYAMLFlow executes a YAML step-based flow.
func (s *Server) runYAMLFlow(ctx context.Context, req *Request, flow *flows.YAMLFlow, params FlowRunParams) *Response {
	// Set timeout
	timeout := 5 * time.Minute
	if params.Timeout > 0 {
		timeout = time.Duration(params.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create executor
	executor := flows.NewYAMLExecutor()
	// TODO: Wire up agent invoker when runner is available

	// Execute
	result, err := executor.Execute(ctx, flow, params.Params)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, "execute: "+err.Error()), req.ID)
	}

	// Convert result
	rpcResult := FlowRunResult{
		RunID:     result.RunID,
		FlowName:  result.FlowName,
		Status:    string(result.Status),
		StartTime: result.StartTime.Unix(),
		EndTime:   result.EndTime.Unix(),
		Duration:  result.Duration.Milliseconds(),
		Error:     result.Error,
	}

	if len(result.Steps) > 0 {
		rpcResult.Steps = make(map[string]*FlowStepResult)
		for id, sr := range result.Steps {
			rpcResult.Steps[id] = &FlowStepResult{
				ID:       sr.ID,
				Status:   string(sr.Status),
				Stdout:   sr.Stdout,
				Stderr:   sr.Stderr,
				Output:   sr.Output,
				ExitCode: sr.ExitCode,
				Error:    sr.Error,
				Skipped:  sr.Skipped,
				Duration: sr.Duration.Milliseconds(),
			}
		}
	}

	resp, _ := NewResponse(rpcResult, req.ID)
	return resp
}

// runShellFlow executes a shell script flow.
func (s *Server) runShellFlow(ctx context.Context, req *Request, flow *flows.Flow, params FlowRunParams) *Response {
	// Set timeout
	timeout := 5 * time.Minute
	if params.Timeout > 0 {
		timeout = time.Duration(params.Timeout) * time.Second
	}

	// Build options
	opts := flows.RunOptions{
		Timeout: timeout,
	}

	// Convert params to JSON input
	if params.Params != nil {
		inputBytes, err := json.Marshal(params.Params)
		if err != nil {
			return NewErrorResponse(NewError(ErrCodeInvalidParams, "invalid params"), req.ID)
		}
		opts.Input = string(inputBytes)
	}

	// Execute
	result, err := flows.Run(ctx, flow, opts)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, "execute: "+err.Error()), req.ID)
	}

	// Convert result
	rpcResult := FlowRunResult{
		RunID:     result.RunID,
		FlowName:  flow.Name,
		Status:    string(result.Status),
		StartTime: result.StartTime.Unix(),
		EndTime:   result.EndTime.Unix(),
		Duration:  result.Duration.Milliseconds(),
	}

	if result.Error != nil {
		rpcResult.Error = result.Error.Error()
	}

	// For shell flows, stdout/stderr are at the flow level
	// Put them in a synthetic step
	rpcResult.Steps = map[string]*FlowStepResult{
		"main": {
			ID:       "main",
			Status:   string(result.Status),
			Stdout:   result.Stdout,
			Stderr:   result.Stderr,
			ExitCode: result.ExitCode,
			Duration: result.Duration.Milliseconds(),
		},
	}

	resp, _ := NewResponse(rpcResult, req.ID)
	return resp
}

// handleFlowList handles the flow.list RPC method.
func (s *Server) handleFlowList(req *Request) *Response {
	var params FlowListParams
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}

	dirs := paths.FlowsDirs()

	var result FlowListResult

	// Discover shell script flows
	discovered, err := flows.Discover(dirs)
	if err == nil {
		for _, f := range discovered {
			if params.Source != "" && string(f.Source) != params.Source {
				continue
			}
			result.Flows = append(result.Flows, FlowInfo{
				Name:        f.Name,
				Description: f.Description,
				Source:      string(f.Source),
				Path:        f.Path,
				IsYAML:      false,
			})
		}
	}

	// Discover YAML flows
	for i, dir := range dirs {
		yamlFlows, err := flows.DiscoverYAMLFlows(dir)
		if err != nil {
			continue
		}

		source := "user"
		if i == 0 {
			source = "project"
		} else if i == len(dirs)-1 {
			source = "built-in"
		}

		for _, yf := range yamlFlows {
			if params.Source != "" && source != params.Source {
				continue
			}
			result.Flows = append(result.Flows, FlowInfo{
				Name:        yf.Name,
				Description: yf.Description,
				Source:      source,
				Version:     yf.Version,
				IsYAML:      true,
				StepCount:   len(yf.Steps),
			})
		}
	}

	resp, _ := NewResponse(result, req.ID)
	return resp
}

// handleFlowGet handles the flow.get RPC method.
func (s *Server) handleFlowGet(req *Request) *Response {
	var params FlowGetParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "invalid params"), req.ID)
	}

	if params.Name == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "name is required"), req.ID)
	}

	dirs := paths.FlowsDirs()

	// Try YAML flows first
	for i, dir := range dirs {
		yamlFlows, err := flows.DiscoverYAMLFlows(dir)
		if err != nil {
			continue
		}

		source := "user"
		if i == 0 {
			source = "project"
		} else if i == len(dirs)-1 {
			source = "built-in"
		}

		for _, yf := range yamlFlows {
			if yf.Name == params.Name {
				result := FlowGetResult{
					Flow: FlowInfo{
						Name:        yf.Name,
						Description: yf.Description,
						Source:      source,
						Version:     yf.Version,
						IsYAML:      true,
						StepCount:   len(yf.Steps),
					},
				}
				resp, _ := NewResponse(result, req.ID)
				return resp
			}
		}
	}

	// Try shell script flows
	discovered, err := flows.Discover(dirs)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, "discover flows: "+err.Error()), req.ID)
	}

	for _, f := range discovered {
		if f.Name == params.Name {
			result := FlowGetResult{
				Flow: FlowInfo{
					Name:        f.Name,
					Description: f.Description,
					Source:      string(f.Source),
					Path:        f.Path,
					IsYAML:      false,
				},
			}
			resp, _ := NewResponse(result, req.ID)
			return resp
		}
	}

	return NewErrorResponse(NewError(ErrCodeFlowNotFound, "flow not found: "+params.Name), req.ID)
}

// handleFlowHistory handles the flow.history RPC method.
func (s *Server) handleFlowHistory(req *Request) *Response {
	var params FlowHistoryParams
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}

	// TODO: Implement when history service is wired to daemon
	result := FlowHistoryResult{
		Runs: []FlowRunSummary{},
	}

	resp, _ := NewResponse(result, req.ID)
	return resp
}
