package daemon

import (
	"encoding/json"

	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/tickets"
)

// ticketService returns or creates the ticket service.
// Lazily initialized since not all daemon users need tickets.
func (s *Server) ticketService() *tickets.Service {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tickets == nil {
		s.tickets = tickets.NewService(paths.SessionsDir())
	}
	return s.tickets
}

// squadTicketService returns the squad ticket service.
func (s *Server) squadTicketService() *tickets.SquadTicketService {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.squadTickets == nil {
		s.squadTickets = tickets.NewSquadTicketService()
	}
	return s.squadTickets
}

// getTicketService returns the appropriate service and identifier.
// If squadName is provided, returns squad ticket service with squadName as identifier.
// Otherwise returns session ticket service with sessionID as identifier.
func (s *Server) getTicketService(sessionID, squadName string) (*tickets.Service, string) {
	if squadName != "" {
		return s.squadTicketService().Service, squadName
	}
	return s.ticketService(), sessionID
}

// ticketToInfo converts a Ticket to TicketInfo for RPC responses.
func ticketToInfo(t *tickets.Ticket) TicketInfo {
	info := TicketInfo{
		ID:          t.ID,
		Status:      string(t.Status),
		Type:        string(t.Type),
		Priority:    t.Priority,
		Assignee:    t.Assignee,
		Deps:        t.Deps,
		Links:       t.Links,
		Parent:      t.Parent,
		Tags:        t.Tags,
		Title:       t.Title,
		Description: t.Description,
		Created:     t.Created.Unix(),
		Session:     t.Session,
		ExternalRef: t.ExternalRef,
		FilePath:    t.FilePath,
	}
	if t.Started != nil {
		info.Started = t.Started.Unix()
	}
	if t.Closed != nil {
		info.Closed = t.Closed.Unix()
	}
	if info.Deps == nil {
		info.Deps = []string{}
	}
	if info.Links == nil {
		info.Links = []string{}
	}
	if info.Tags == nil {
		info.Tags = []string{}
	}
	return info
}

func (s *Server) handleTicketCreate(req *Request) *Response {
	var params TicketCreateParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if params.SessionID == "" && params.SquadName == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name is required"), req.ID)
	}
	if params.Title == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "title is required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	ticket, err := svc.Create(identifier, tickets.CreateOptions{
		Title:       params.Title,
		Description: params.Description,
		Type:        tickets.Type(params.Type),
		Priority:    params.Priority,
		Assignee:    params.Assignee,
		Deps:        params.Deps,
		Parent:      params.Parent,
		Tags:        params.Tags,
		ExternalRef: params.ExternalRef,
	})
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	result := TicketCreateResult{
		ID:   ticket.ID,
		Path: ticket.FilePath,
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketGet(req *Request) *Response {
	var params TicketGetParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if (params.SessionID == "" && params.SquadName == "") || params.TicketID == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name, and ticket_id are required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	ticket, err := svc.Get(identifier, params.TicketID)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	result := TicketGetResult{
		Ticket: ticketToInfo(ticket),
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketList(req *Request) *Response {
	var params TicketListParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if params.SessionID == "" && params.SquadName == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name is required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	ticketList, err := svc.List(identifier, tickets.Filter{
		Status:   tickets.Status(params.Status),
		Assignee: params.Assignee,
		Type:     tickets.Type(params.Type),
		Tags:     params.Tags,
		Parent:   params.Parent,
	})
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	result := TicketListResult{
		Tickets: make([]TicketInfo, len(ticketList)),
	}
	for i, t := range ticketList {
		result.Tickets[i] = ticketToInfo(t)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketUpdate(req *Request) *Response {
	var params TicketUpdateParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if params.SessionID == "" || params.TicketID == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id and ticket_id are required"), req.ID)
	}

	svc := s.ticketService()
	ticket, err := svc.Get(params.SessionID, params.TicketID)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	// Apply updates
	if params.Title != nil {
		ticket.Title = *params.Title
	}
	if params.Description != nil {
		ticket.Description = *params.Description
	}
	if params.Type != nil {
		ticket.Type = tickets.Type(*params.Type)
	}
	if params.Priority != nil {
		ticket.Priority = *params.Priority
	}
	if params.Assignee != nil {
		ticket.Assignee = *params.Assignee
	}
	if params.Tags != nil {
		ticket.Tags = params.Tags
	}
	if params.ExternalRef != nil {
		ticket.ExternalRef = *params.ExternalRef
	}

	if err := svc.Update(ticket); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	result := TicketGetResult{
		Ticket: ticketToInfo(ticket),
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketDelete(req *Request) *Response {
	var params TicketDeleteParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if params.SessionID == "" || params.TicketID == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id and ticket_id are required"), req.ID)
	}

	svc := s.ticketService()
	if err := svc.Delete(params.SessionID, params.TicketID); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	resp, _ := NewResponse(struct{}{}, req.ID)
	return resp
}

func (s *Server) handleTicketStart(req *Request) *Response {
	var params TicketStatusParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if (params.SessionID == "" && params.SquadName == "") || params.TicketID == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name, and ticket_id are required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	if err := svc.Start(identifier, params.TicketID); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	// Return updated ticket
	ticket, _ := svc.Get(identifier, params.TicketID)
	result := TicketGetResult{
		Ticket: ticketToInfo(ticket),
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketClose(req *Request) *Response {
	var params TicketStatusParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if (params.SessionID == "" && params.SquadName == "") || params.TicketID == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name, and ticket_id are required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	if err := svc.Close(identifier, params.TicketID); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	// Return updated ticket
	ticket, _ := svc.Get(identifier, params.TicketID)
	result := TicketGetResult{
		Ticket: ticketToInfo(ticket),
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketReopen(req *Request) *Response {
	var params TicketStatusParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if (params.SessionID == "" && params.SquadName == "") || params.TicketID == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name, and ticket_id are required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	if err := svc.Reopen(identifier, params.TicketID); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	// Return updated ticket
	ticket, _ := svc.Get(identifier, params.TicketID)
	result := TicketGetResult{
		Ticket: ticketToInfo(ticket),
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketBlock(req *Request) *Response {
	var params TicketStatusParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if (params.SessionID == "" && params.SquadName == "") || params.TicketID == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name, and ticket_id are required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	if err := svc.Block(identifier, params.TicketID); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	// Return updated ticket
	ticket, _ := svc.Get(identifier, params.TicketID)
	result := TicketGetResult{
		Ticket: ticketToInfo(ticket),
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketAssign(req *Request) *Response {
	var params TicketAssignParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if (params.SessionID == "" && params.SquadName == "") || params.TicketID == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name, and ticket_id are required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	if err := svc.Assign(identifier, params.TicketID, params.Assignee); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	// Return updated ticket
	ticket, _ := svc.Get(identifier, params.TicketID)
	result := TicketGetResult{
		Ticket: ticketToInfo(ticket),
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketAddNote(req *Request) *Response {
	var params TicketAddNoteParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if (params.SessionID == "" && params.SquadName == "") || params.TicketID == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name, and ticket_id are required"), req.ID)
	}
	if params.Content == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "content is required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	if err := svc.AddNote(identifier, params.TicketID, params.Content); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	// Return updated ticket
	ticket, _ := svc.Get(identifier, params.TicketID)
	result := TicketGetResult{
		Ticket: ticketToInfo(ticket),
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketReady(req *Request) *Response {
	var params TicketReadyParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if params.SessionID == "" && params.SquadName == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name is required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	ticketList, err := svc.Ready(identifier, params.Assignee)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	result := TicketReadyResult{
		Tickets: make([]TicketInfo, len(ticketList)),
	}
	for i, t := range ticketList {
		result.Tickets[i] = ticketToInfo(t)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketBlocked(req *Request) *Response {
	var params TicketBlockedParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if params.SessionID == "" && params.SquadName == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name is required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	ticketList, err := svc.Blocked(identifier, params.Assignee)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	result := TicketBlockedResult{
		Tickets: make([]TicketInfo, len(ticketList)),
	}
	for i, t := range ticketList {
		result.Tickets[i] = ticketToInfo(t)
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketAddDep(req *Request) *Response {
	var params TicketDepParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if (params.SessionID == "" && params.SquadName == "") || params.TicketID == "" || params.DepID == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name, ticket_id and dep_id are required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	if err := svc.AddDep(identifier, params.TicketID, params.DepID); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	// Return updated ticket
	ticket, _ := svc.Get(identifier, params.TicketID)
	result := TicketGetResult{
		Ticket: ticketToInfo(ticket),
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleTicketRemoveDep(req *Request) *Response {
	var params TicketDepParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if (params.SessionID == "" && params.SquadName == "") || params.TicketID == "" || params.DepID == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "session_id or squad_name, ticket_id and dep_id are required"), req.ID)
	}

	svc, identifier := s.getTicketService(params.SessionID, params.SquadName)
	if err := svc.RemoveDep(identifier, params.TicketID, params.DepID); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	// Return updated ticket
	ticket, _ := svc.Get(identifier, params.TicketID)
	result := TicketGetResult{
		Ticket: ticketToInfo(ticket),
	}
	resp, _ := NewResponse(result, req.ID)
	return resp
}
