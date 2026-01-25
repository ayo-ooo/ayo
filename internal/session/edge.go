package session

import (
	"context"

	"github.com/alexcabrera/ayo/internal/db"
)

// EdgeType represents the type of relationship between sessions.
type EdgeType string

const (
	EdgeTypeAgentCall    EdgeType = "agent_call"
	EdgeTypeChain        EdgeType = "chain"
	EdgeTypeContinuation EdgeType = "continuation"
	EdgeTypeTitleGen     EdgeType = "title_gen"
)

// Edge represents a relationship between two sessions.
type Edge struct {
	ParentID         string
	ChildID          string
	EdgeType         EdgeType
	TriggerMessageID string
	CreatedAt        int64
}

// EdgeService provides operations on session edges.
type EdgeService struct {
	q *db.Queries
}

// NewEdgeService creates a new edge service.
func NewEdgeService(q *db.Queries) *EdgeService {
	return &EdgeService{q: q}
}

// Create creates a new edge between sessions.
func (s *EdgeService) Create(ctx context.Context, parentID, childID string, edgeType EdgeType, triggerMessageID string) error {
	return s.q.CreateEdge(ctx, db.CreateEdgeParams{
		ParentID:         parentID,
		ChildID:          childID,
		EdgeType:         string(edgeType),
		TriggerMessageID: toNullString(triggerMessageID),
	})
}

// GetParents returns all parent edges for a session.
func (s *EdgeService) GetParents(ctx context.Context, childID string) ([]Edge, error) {
	dbEdges, err := s.q.GetParentEdges(ctx, childID)
	if err != nil {
		return nil, err
	}
	return edgesFromDB(dbEdges), nil
}

// GetChildren returns all child edges for a session.
func (s *EdgeService) GetChildren(ctx context.Context, parentID string) ([]Edge, error) {
	dbEdges, err := s.q.GetChildEdges(ctx, parentID)
	if err != nil {
		return nil, err
	}
	return edgesFromDB(dbEdges), nil
}

// Delete deletes an edge.
func (s *EdgeService) Delete(ctx context.Context, parentID, childID string) error {
	return s.q.DeleteEdge(ctx, db.DeleteEdgeParams{
		ParentID: parentID,
		ChildID:  childID,
	})
}

// DeleteBySession deletes all edges involving a session.
func (s *EdgeService) DeleteBySession(ctx context.Context, sessionID string) error {
	return s.q.DeleteEdgesBySession(ctx, sessionID)
}

func edgeFromDB(d db.SessionEdge) Edge {
	return Edge{
		ParentID:         d.ParentID,
		ChildID:          d.ChildID,
		EdgeType:         EdgeType(d.EdgeType),
		TriggerMessageID: d.TriggerMessageID.String,
		CreatedAt:        d.CreatedAt,
	}
}

func edgesFromDB(ds []db.SessionEdge) []Edge {
	edges := make([]Edge, len(ds))
	for i, d := range ds {
		edges[i] = edgeFromDB(d)
	}
	return edges
}
