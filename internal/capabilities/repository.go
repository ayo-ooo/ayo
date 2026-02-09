package capabilities

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/alexcabrera/ayo/internal/db"
)

// Repository provides capability storage operations.
type Repository struct {
	q db.Querier
}

// NewRepository creates a new capability repository.
func NewRepository(q db.Querier) *Repository {
	return &Repository{q: q}
}

// StoreCapabilities stores inferred capabilities for an agent.
// This replaces any existing capabilities for the agent.
func (r *Repository) StoreCapabilities(ctx context.Context, agentID string, result *InferenceResult) error {
	// Delete existing capabilities
	if err := r.q.DeleteCapabilitiesByAgent(ctx, agentID); err != nil {
		return err
	}

	now := time.Now().Unix()

	// Store each capability
	for _, cap := range result.Capabilities {
		capID := uuid.New().String()

		if err := r.q.CreateCapability(ctx, db.CreateCapabilityParams{
			ID:          capID,
			AgentID:     agentID,
			Name:        cap.Name,
			Description: cap.Description,
			Confidence:  cap.Confidence,
			Source:      "inference", // default source
			InputHash:   result.InputHash,
			CreatedAt:   now,
			UpdatedAt:   now,
		}); err != nil {
			return err
		}
	}

	return nil
}

// GetCapabilities returns all capabilities for an agent.
func (r *Repository) GetCapabilities(ctx context.Context, agentID string) ([]StoredCapability, error) {
	rows, err := r.q.GetCapabilitiesByAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}

	result := make([]StoredCapability, len(rows))
	for i, row := range rows {
		result[i] = StoredCapability{
			ID:          row.ID,
			AgentID:     row.AgentID,
			Name:        row.Name,
			Description: row.Description,
			Confidence:  row.Confidence,
			Source:      row.Source,
			Embedding:   row.Embedding,
			InputHash:   row.InputHash,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}
	}

	return result, nil
}

// GetAllCapabilities returns all capabilities across all agents.
func (r *Repository) GetAllCapabilities(ctx context.Context) ([]StoredCapability, error) {
	rows, err := r.q.ListAllCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]StoredCapability, len(rows))
	for i, row := range rows {
		result[i] = StoredCapability{
			ID:          row.ID,
			AgentID:     row.AgentID,
			Name:        row.Name,
			Description: row.Description,
			Confidence:  row.Confidence,
			Source:      row.Source,
			Embedding:   row.Embedding,
			InputHash:   row.InputHash,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}
	}

	return result, nil
}

// NeedsRefresh checks if capabilities need to be re-inferred based on input hash.
func (r *Repository) NeedsRefresh(ctx context.Context, agentID string, currentHash string) (bool, error) {
	caps, err := r.q.GetCapabilitiesByAgent(ctx, agentID)
	if err != nil {
		return true, nil // If error, assume needs refresh
	}

	if len(caps) == 0 {
		return true, nil // No capabilities, needs inference
	}

	// Check if hash matches
	if caps[0].InputHash != currentHash {
		return true, nil // Hash changed, needs refresh
	}

	return false, nil
}

// SearchCapabilities searches for capabilities by name or description.
func (r *Repository) SearchCapabilities(ctx context.Context, query string, limit int) ([]StoredCapability, error) {
	searchPattern := "%" + query + "%"
	rows, err := r.q.SearchCapabilitiesByName(ctx, db.SearchCapabilitiesByNameParams{
		Name:        searchPattern,
		Description: searchPattern,
		Limit:       int64(limit),
	})
	if err != nil {
		return nil, err
	}

	result := make([]StoredCapability, len(rows))
	for i, row := range rows {
		result[i] = StoredCapability{
			ID:          row.ID,
			AgentID:     row.AgentID,
			Name:        row.Name,
			Description: row.Description,
			Confidence:  row.Confidence,
			Source:      row.Source,
			Embedding:   row.Embedding,
			InputHash:   row.InputHash,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}
	}

	return result, nil
}

// UpdateEmbedding updates the embedding for a capability.
func (r *Repository) UpdateEmbedding(ctx context.Context, capabilityID string, embedding []byte) error {
	now := time.Now().Unix()
	return r.q.UpdateCapabilityEmbedding(ctx, db.UpdateCapabilityEmbeddingParams{
		Embedding: embedding,
		UpdatedAt: now,
		ID:        capabilityID,
	})
}
