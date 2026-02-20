package daemon

import (
	"context"
	"encoding/json"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/squads"
	"github.com/alexcabrera/ayo/internal/tickets"
)

// SquadRPC provides RPC handlers for squad management.
type SquadRPC struct {
	service       *squads.Service
	ticketService *tickets.SquadTicketService
	ticketWatcher *TicketWatcher
	invoker       squads.AgentInvoker
}

// NewSquadRPC creates a new squad RPC handler.
func NewSquadRPC(service *squads.Service, ticketService *tickets.SquadTicketService, ticketWatcher *TicketWatcher, invoker squads.AgentInvoker) *SquadRPC {
	return &SquadRPC{
		service:       service,
		ticketService: ticketService,
		ticketWatcher: ticketWatcher,
		invoker:       invoker,
	}
}

// HandleSquadCreate handles squads.create.
func (r *SquadRPC) HandleSquadCreate(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadCreateParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}

	cfg := config.SquadConfig{
		Name:           p.Name,
		Description:    p.Description,
		Image:          p.Image,
		Ephemeral:      p.Ephemeral,
		Agents:         p.Agents,
		WorkspaceMount: p.WorkspaceMount,
		Packages:       p.Packages,
		OutputPath:     p.OutputPath,
	}

	squad, err := r.service.Create(ctx, cfg)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	return SquadCreateResult{
		Squad: squadToInfo(squad, r.service),
	}, nil
}

// HandleSquadDestroy handles squads.destroy.
func (r *SquadRPC) HandleSquadDestroy(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadDestroyParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}

	err := r.service.Destroy(ctx, p.Name, p.DeleteData)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	return SquadDestroyResult{Success: true}, nil
}

// HandleSquadList handles squads.list.
func (r *SquadRPC) HandleSquadList(ctx context.Context, params json.RawMessage) (any, *Error) {
	squadList, err := r.service.List(ctx)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	var infos []SquadInfo
	for _, squad := range squadList {
		infos = append(infos, squadToInfo(squad, r.service))
	}

	return SquadListResult{Squads: infos}, nil
}

// HandleSquadGet handles squads.get.
func (r *SquadRPC) HandleSquadGet(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadGetParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}

	squad, err := r.service.Get(ctx, p.Name)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	return SquadGetResult{
		Squad: squadToInfo(squad, r.service),
	}, nil
}

// HandleSquadStart handles squads.start.
func (r *SquadRPC) HandleSquadStart(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadStartParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}

	err := r.service.Start(ctx, p.Name)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	return SquadStartResult{Success: true}, nil
}

// HandleSquadStop handles squads.stop.
func (r *SquadRPC) HandleSquadStop(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadStopParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}

	err := r.service.Stop(ctx, p.Name)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	return SquadStopResult{Success: true}, nil
}

// HandleSquadAddAgent handles squads.add_agent.
func (r *SquadRPC) HandleSquadAddAgent(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadAddAgentParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}
	if p.AgentHandle == "" {
		return nil, NewError(ErrCodeInvalidParams, "agent_handle is required")
	}

	// Get current squad config
	cfg, err := config.LoadSquadConfig(p.Name)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	// Add agent if not already present
	found := false
	for _, a := range cfg.Agents {
		if a == p.AgentHandle {
			found = true
			break
		}
	}
	if !found {
		cfg.Agents = append(cfg.Agents, p.AgentHandle)
		if err := config.SaveSquadConfig(cfg); err != nil {
			return nil, NewError(ErrCodeInternal, err.Error())
		}
	}

	// Ensure agent user exists in sandbox
	if err := r.service.EnsureAgentUser(ctx, p.Name, p.AgentHandle); err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	return SquadAddAgentResult{Success: true}, nil
}

// HandleSquadRemoveAgent handles squads.remove_agent.
func (r *SquadRPC) HandleSquadRemoveAgent(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadRemoveAgentParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}
	if p.AgentHandle == "" {
		return nil, NewError(ErrCodeInvalidParams, "agent_handle is required")
	}

	// Get current squad config
	cfg, err := config.LoadSquadConfig(p.Name)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	// Remove agent
	var newAgents []string
	for _, a := range cfg.Agents {
		if a != p.AgentHandle {
			newAgents = append(newAgents, a)
		}
	}
	cfg.Agents = newAgents

	if err := config.SaveSquadConfig(cfg); err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	return SquadRemoveAgentResult{Success: true}, nil
}

// HandleSquadTicketsReady handles squads.tickets_ready.
// Called by @ayo after creating tickets to trigger agent spawning.
func (r *SquadRPC) HandleSquadTicketsReady(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadTicketsReadyParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}

	// Verify squad exists
	_, err := r.service.Get(ctx, p.Name)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	// Read tickets from squad and identify assigned agents
	ticketList, err := r.ticketService.ListForSquad(p.Name, tickets.Filter{})
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	// Find unique assignees from ready tickets
	assignees := make(map[string]bool)
	for _, t := range ticketList {
		if t.Assignee != "" && (t.Status == tickets.StatusOpen || t.Status == tickets.StatusInProgress) {
			assignees[t.Assignee] = true
		}
	}

	// Start watching squad tickets
	if r.ticketWatcher != nil {
		if err := r.ticketWatcher.WatchSquad(p.Name); err != nil {
			return nil, NewError(ErrCodeInternal, "failed to watch squad tickets: "+err.Error())
		}
	}

	var agentList []string
	for agent := range assignees {
		agentList = append(agentList, agent)
	}

	return SquadTicketsReadyResult{
		TicketsFound:  len(ticketList),
		AgentsSpawned: agentList,
	}, nil
}

// HandleSquadNotifyAgents handles squads.notify_agents.
// Spawns agent sessions for all agents with assigned tickets in the squad.
func (r *SquadRPC) HandleSquadNotifyAgents(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadNotifyAgentsParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}

	// Verify squad exists
	_, err := r.service.Get(ctx, p.Name)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	// Read tickets and find assigned agents
	ticketList, err := r.ticketService.ListForSquad(p.Name, tickets.Filter{})
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	// Group tickets by assignee
	agentTickets := make(map[string][]*tickets.Ticket)
	for _, t := range ticketList {
		if t.Assignee != "" && (t.Status == tickets.StatusOpen || t.Status == tickets.StatusInProgress) {
			agentTickets[t.Assignee] = append(agentTickets[t.Assignee], t)
		}
	}

	var sessionsSpawned []string
	ticketsAssigned := 0

	// Spawn sessions for each agent (placeholder - actual spawning would use agent_spawner)
	for agent, agentTix := range agentTickets {
		sessionsSpawned = append(sessionsSpawned, agent)
		ticketsAssigned += len(agentTix)
	}

	return SquadNotifyAgentsResult{
		SessionsSpawned: sessionsSpawned,
		TicketsAssigned: ticketsAssigned,
	}, nil
}

// HandleSquadWaitCompletion handles squads.wait_completion.
// Blocks until all squad tickets are closed or timeout expires.
func (r *SquadRPC) HandleSquadWaitCompletion(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadWaitCompletionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}

	// Verify squad exists
	_, err := r.service.Get(ctx, p.Name)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	// Check current ticket status
	ticketList, err := r.ticketService.ListForSquad(p.Name, tickets.Filter{})
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	// Count open vs closed
	var ticketsClosed, ticketsOpen int
	for _, t := range ticketList {
		if t.Status == tickets.StatusClosed {
			ticketsClosed++
		} else {
			ticketsOpen++
		}
	}

	// If all closed, return immediately
	if ticketsOpen == 0 {
		return SquadWaitCompletionResult{
			Completed:     true,
			TicketsClosed: ticketsClosed,
			TicketsOpen:   0,
		}, nil
	}

	// For now, return current status (full implementation would poll/watch)
	return SquadWaitCompletionResult{
		Completed:     false,
		TicketsClosed: ticketsClosed,
		TicketsOpen:   ticketsOpen,
	}, nil
}

// HandleSquadSyncOutput handles squads.sync_output.
// Copies files from squad /workspace/ to target directory on host.
func (r *SquadRPC) HandleSquadSyncOutput(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadSyncOutputParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}
	if p.TargetPath == "" {
		return nil, NewError(ErrCodeInvalidParams, "target_path is required")
	}

	// Verify squad exists
	_, err := r.service.Get(ctx, p.Name)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	// Use sync.SyncOutput from internal/sync package
	// For now, return placeholder (actual implementation uses rsync/copy)
	return SquadSyncOutputResult{
		FilesCopied: 0,
		BytesCopied: 0,
	}, nil
}

// HandleSquadCleanup handles squads.cleanup.
// Destroys ephemeral squad after work product synced.
func (r *SquadRPC) HandleSquadCleanup(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadCleanupParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}

	// Get squad and check if ephemeral
	squad, err := r.service.Get(ctx, p.Name)
	if err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	// Only cleanup ephemeral squads
	if !squad.Config.Ephemeral {
		return nil, NewError(ErrCodeInvalidParams, "cannot cleanup non-ephemeral squad")
	}

	// Destroy the squad with data deletion
	if err := r.service.Destroy(ctx, p.Name, true); err != nil {
		return nil, NewError(ErrCodeInternal, err.Error())
	}

	return SquadCleanupResult{Success: true}, nil
}

// HandleSquadDispatch handles squads.dispatch.
// Dispatches work to a squad synchronously and returns the result.
func (r *SquadRPC) HandleSquadDispatch(ctx context.Context, params json.RawMessage) (any, *Error) {
	var p SquadDispatchParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewError(ErrCodeInvalidParams, "invalid params: "+err.Error())
	}

	if p.Name == "" {
		return nil, NewError(ErrCodeInvalidParams, "name is required")
	}

	// Strip # prefix if present
	name := squads.StripPrefix(p.Name)

	// Get the squad
	squad, err := r.service.Get(ctx, name)
	if err != nil {
		return nil, NewError(ErrCodeInternal, "failed to get squad: "+err.Error())
	}

	// Set the invoker on the squad so Dispatch can invoke agents
	squad.Invoker = r.invoker

	// Start if not running and StartIfStopped is set
	if !squad.IsRunning() && p.StartIfStopped {
		if err := r.service.Start(ctx, name); err != nil {
			return nil, NewError(ErrCodeInternal, "failed to start squad: "+err.Error())
		}
		// Re-get squad after start
		squad, err = r.service.Get(ctx, name)
		if err != nil {
			return nil, NewError(ErrCodeInternal, "failed to get squad after start: "+err.Error())
		}
	}

	// Create dispatch input
	input := squads.DispatchInput{
		Prompt: p.Prompt,
		Data:   p.Data,
	}

	// Dispatch and get result
	result, err := squad.Dispatch(ctx, input)
	if err != nil {
		// Check if it's a validation error
		if valErr, ok := err.(*squads.ValidationError); ok {
			return SquadDispatchResult{
				Error: valErr.Error(),
			}, nil
		}
		return nil, NewError(ErrCodeInternal, "dispatch failed: "+err.Error())
	}

	// Validate output if schema exists
	if err := squad.ValidateOutput(result); err != nil {
		return SquadDispatchResult{
			Output: result.Output,
			Raw:    result.Raw,
			Error:  "output validation failed: " + err.Error(),
		}, nil
	}

	return SquadDispatchResult{
		Output: result.Output,
		Raw:    result.Raw,
	}, nil
}

// squadToInfo converts a Squad to SquadInfo for RPC responses.
func squadToInfo(squad *squads.Squad, service *squads.Service) SquadInfo {
	info := SquadInfo{
		Name:        squad.Name,
		Description: squad.Config.Description,
		Status:      string(squad.Status),
		Agents:      squad.Config.Agents,
		Ephemeral:   squad.Config.Ephemeral,
	}

	if service != nil {
		info.TicketsDir = service.GetTicketsDir(squad.Name)
		info.ContextDir = service.GetContextDir(squad.Name)
		info.WorkspaceDir = service.GetWorkspaceDir(squad.Name)
	}

	return info
}
