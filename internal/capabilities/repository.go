package capabilities

import (
	"context"
	"fmt"
)

// Repository provides capability storage operations.
// In the build system, capabilities are determined at build time
// and embedded in the compiled executable.
type Repository struct {
	// Build system does not use database for capabilities
}

// NewRepository creates a new capability repository.
// Returns error since capabilities are not supported in build system.
func NewRepository() (*Repository, error) {
	return nil, fmt.Errorf("capabilities repository is not supported in the build system. Capabilities are now determined at build time and embedded in executables")
}

// StoreCapabilities is not supported in build system.
func (r *Repository) StoreCapabilities(ctx context.Context, agentID string, result *InferenceResult) error {
	return fmt.Errorf("capabilities storage is not supported in the build system")
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

// GetAllCapabilities is not supported in build system.
func (r *Repository) GetAllCapabilities(ctx context.Context) ([]StoredCapability, error) {
	return nil, fmt.Errorf("capabilities retrieval is not supported in the build system")
}

// NeedsRefresh is not supported in build system.
func (r *Repository) NeedsRefresh(ctx context.Context, agentID string, currentHash string) (bool, error) {
	return false, fmt.Errorf("capabilities refresh check is not supported in the build system")
}

// SearchCapabilities is not supported in build system.
func (r *Repository) SearchCapabilities(ctx context.Context, query string, limit int) ([]StoredCapability, error) {
	return nil, fmt.Errorf("capabilities search is not supported in the build system")
}

// UpdateEmbedding is not supported in build system.
func (r *Repository) UpdateEmbedding(ctx context.Context, capabilityID string, embedding []byte) error {
	return fmt.Errorf("capabilities embedding update is not supported in the build system")
}

// GetCapabilities is not supported in build system.
func (r *Repository) GetCapabilities(ctx context.Context, agentID string) ([]StoredCapability, error) {
	return nil, fmt.Errorf("capabilities retrieval is not supported in the build system")
}
