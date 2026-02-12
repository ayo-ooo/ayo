package daemon

import (
	"context"
	"encoding/json"
)

// Matrix RPC handlers

func (s *Server) handleMatrixStatus(req *Request) *Response {
	var conduitStatus ConduitStatus
	if s.conduit != nil {
		conduitStatus = s.conduit.Status()
	}

	var brokerStatus BrokerStatus
	if s.matrixBroker != nil {
		brokerStatus = s.matrixBroker.Status()
	}

	result := MatrixStatusResult{
		Conduit: conduitStatus,
		Broker:  brokerStatus,
	}

	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleMatrixRoomsList(req *Request) *Response {
	if s.matrixBroker == nil {
		return NewErrorResponse(NewError(ErrCodeInternal, "matrix broker not initialized"), req.ID)
	}

	var params MatrixRoomsListParams
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}

	rooms := s.matrixBroker.ListRooms()

	// Filter by session if requested
	if params.SessionID != "" {
		filtered := make([]*RoomInfo, 0)
		for _, room := range rooms {
			if room.Session == params.SessionID {
				filtered = append(filtered, room)
			}
		}
		rooms = filtered
	}

	result := MatrixRoomsListResult{Rooms: rooms}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleMatrixRoomsCreate(_ context.Context, req *Request) *Response {
	if s.matrixBroker == nil {
		return NewErrorResponse(NewError(ErrCodeInternal, "matrix broker not initialized"), req.ID)
	}

	var params MatrixRoomsCreateParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if params.Name == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "name is required"), req.ID)
	}

	roomID, err := s.matrixBroker.CreateSessionRoom(params.SessionID, params.Name)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	// Invite agents if specified
	for _, handle := range params.Invite {
		if err := s.matrixBroker.InviteAgent(roomID, handle); err != nil {
			// Log but don't fail the whole operation
		}
	}

	result := MatrixRoomsCreateResult{RoomID: roomID}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleMatrixRoomsMembers(req *Request) *Response {
	if s.matrixBroker == nil {
		return NewErrorResponse(NewError(ErrCodeInternal, "matrix broker not initialized"), req.ID)
	}

	var params MatrixRoomsMembersParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	room, err := s.matrixBroker.GetRoom(params.RoomID)
	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	members := make([]MemberInfo, 0, len(room.Members))
	for _, handle := range room.Members {
		members = append(members, MemberInfo{
			UserID:      "@" + handle + ":ayo.local",
			DisplayName: handle,
			IsAgent:     true,
			Handle:      handle,
		})
	}

	result := MatrixRoomsMembersResult{Members: members}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleMatrixRoomsInvite(req *Request) *Response {
	if s.matrixBroker == nil {
		return NewErrorResponse(NewError(ErrCodeInternal, "matrix broker not initialized"), req.ID)
	}

	var params MatrixRoomsInviteParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if err := s.matrixBroker.InviteAgent(params.RoomID, params.Handle); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	resp, _ := NewResponse(struct{}{}, req.ID)
	return resp
}

func (s *Server) handleMatrixRoomsJoin(req *Request) *Response {
	if s.matrixBroker == nil {
		return NewErrorResponse(NewError(ErrCodeInternal, "matrix broker not initialized"), req.ID)
	}

	var params MatrixRoomsJoinParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	// Ensure agent exists and invite them
	if _, err := s.matrixBroker.EnsureAgentUser(params.Handle); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	if err := s.matrixBroker.InviteAgent(params.RoomID, params.Handle); err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	resp, _ := NewResponse(struct{}{}, req.ID)
	return resp
}

func (s *Server) handleMatrixRoomsLeave(req *Request) *Response {
	if s.matrixBroker == nil {
		return NewErrorResponse(NewError(ErrCodeInternal, "matrix broker not initialized"), req.ID)
	}

	var params MatrixRoomsLeaveParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	// Leave room functionality not yet implemented
	return NewErrorResponse(NewError(ErrCodeInternal, "leave room not implemented"), req.ID)
}

func (s *Server) handleMatrixSend(req *Request) *Response {
	if s.matrixBroker == nil {
		return NewErrorResponse(NewError(ErrCodeInternal, "matrix broker not initialized"), req.ID)
	}

	var params MatrixSendParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	if params.RoomID == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "room_id is required"), req.ID)
	}
	if params.Content == "" {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, "content is required"), req.ID)
	}

	// Default to ayo agent
	handle := params.AsAgent
	if handle == "" {
		handle = "ayo"
	}

	var eventID string
	var err error

	if params.Format == "markdown" {
		eventID, err = s.matrixBroker.SendMarkdown(handle, params.RoomID, params.Content)
	} else {
		eventID, err = s.matrixBroker.SendMessage(handle, params.RoomID, params.Content)
	}

	if err != nil {
		return NewErrorResponse(NewError(ErrCodeInternal, err.Error()), req.ID)
	}

	result := MatrixSendResult{EventID: eventID}
	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleMatrixRead(req *Request) *Response {
	if s.matrixBroker == nil {
		return NewErrorResponse(NewError(ErrCodeInternal, "matrix broker not initialized"), req.ID)
	}

	var params MatrixReadParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(NewError(ErrCodeInvalidParams, err.Error()), req.ID)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	// For now, return messages from the ayo agent's queue
	// In a full implementation, this would fetch from Matrix history
	messages := s.matrixBroker.GetMessages("ayo", limit)

	result := MatrixReadResult{
		Messages: messages,
		HasMore:  false,
	}

	resp, _ := NewResponse(result, req.ID)
	return resp
}

func (s *Server) handleMatrixReadStream(_ context.Context, req *Request) *Response {
	// For streaming, we return immediately with any available messages
	// The client should long-poll
	return s.handleMatrixRead(req)
}
