package agent

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/db"
)

// CreateOptions contains options for creating an agent.
type CreateOptions struct {
	// Handle is the agent handle (e.g., "science-researcher").
	Handle string

	// SystemPrompt is the agent's system prompt.
	SystemPrompt string

	// Description is a brief description of what this agent does.
	Description string

	// Model is the model to use for this agent (optional).
	Model string

	// Skills is the list of skills to include (optional).
	Skills []string

	// AllowedTools is the list of tools to allow (optional).
	AllowedTools []string

	// CreatedBy is the agent or user that created this agent.
	// Common values: "@ayo", "user"
	CreatedBy string

	// CreationReason explains why the agent was created.
	CreationReason string

	// SandboxNetwork controls network access in sandbox.
	SandboxNetwork bool
}

// CreateAgentResult contains the result of creating an agent.
type CreateAgentResult struct {
	// Agent is the created agent.
	Agent Agent

	// AgentID is the database ID for ayo-created agents (empty for user-created).
	AgentID string

	// IsAyoCreated indicates if the agent was created by @ayo.
	IsAyoCreated bool
}

// CreateAgent creates a new agent with the given options.
// If CreatedBy is "@ayo", the agent is registered in the database for tracking.
func CreateAgent(ctx context.Context, cfg config.Config, q db.Querier, opts CreateOptions) (*CreateAgentResult, error) {
	handle := NormalizeHandle(opts.Handle)

	// Validate handle
	if IsReservedNamespace(handle) {
		return nil, fmt.Errorf("cannot use reserved handle %s", handle)
	}

	// Build agent config
	agentCfg := Config{
		Description:  opts.Description,
		Model:        opts.Model,
		Skills:       opts.Skills,
		AllowedTools: opts.AllowedTools,
	}

	if opts.SandboxNetwork {
		agentCfg.Sandbox.Network = boolPtr(true)
	}

	// Default system prompt
	systemPrompt := opts.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant."
	}

	// Save agent to filesystem
	agent, err := Save(cfg, handle, agentCfg, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("save agent: %w", err)
	}

	result := &CreateAgentResult{
		Agent: agent,
	}

	// If created by @ayo, register in database for tracking
	if opts.CreatedBy == "@ayo" && q != nil {
		now := time.Now().Unix()
		agentID := uuid.New().String()

		promptHash := hashPrompt(systemPrompt)

		err := q.CreateAyoAgent(ctx, db.CreateAyoAgentParams{
			AgentID:           agentID,
			AgentHandle:       "@" + handle,
			CreatedBy:         opts.CreatedBy,
			CreationReason:    sqlNullString(opts.CreationReason),
			OriginalPrompt:    systemPrompt,
			CurrentPromptHash: sqlNullString(promptHash),
			CreatedAt:         now,
			UpdatedAt:         now,
		})
		if err != nil {
			// Log but don't fail - agent is already created on filesystem
			// We can try to register it again later
		} else {
			result.AgentID = agentID
			result.IsAyoCreated = true
		}
	}

	return result, nil
}

// RecordInvocation records an agent invocation for ayo-created agents.
func RecordInvocation(ctx context.Context, q db.Querier, agentHandle string) error {
	agent, err := q.GetAyoAgentByHandle(ctx, agentHandle)
	if err != nil {
		// Not an ayo-created agent, skip
		return nil
	}

	now := time.Now().Unix()
	return q.UpdateAyoAgentInvocation(ctx, db.UpdateAyoAgentInvocationParams{
		LastUsedAt: sqlNullInt64(now),
		UpdatedAt:  now,
		AgentID:    agent.AgentID,
	})
}

// RecordSuccess records a successful invocation for ayo-created agents.
func RecordSuccess(ctx context.Context, q db.Querier, agentHandle string) error {
	agent, err := q.GetAyoAgentByHandle(ctx, agentHandle)
	if err != nil {
		return nil
	}

	now := time.Now().Unix()
	return q.UpdateAyoAgentSuccess(ctx, db.UpdateAyoAgentSuccessParams{
		UpdatedAt: now,
		AgentID:   agent.AgentID,
	})
}

// RecordFailure records a failed invocation for ayo-created agents.
func RecordFailure(ctx context.Context, q db.Querier, agentHandle string) error {
	agent, err := q.GetAyoAgentByHandle(ctx, agentHandle)
	if err != nil {
		return nil
	}

	now := time.Now().Unix()
	return q.UpdateAyoAgentFailure(ctx, db.UpdateAyoAgentFailureParams{
		UpdatedAt: now,
		AgentID:   agent.AgentID,
	})
}

// RefinementOptions contains options for refining an agent's prompt.
type RefinementOptions struct {
	AgentHandle   string
	NewPrompt     string
	Reason        string
	UpdateOnDisk  bool
}

// RefineAgent refines an ayo-created agent's prompt.
func RefineAgent(ctx context.Context, cfg config.Config, q db.Querier, opts RefinementOptions) error {
	agent, err := q.GetAyoAgentByHandle(ctx, opts.AgentHandle)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	now := time.Now().Unix()
	refinementID := uuid.New().String()

	// Get current prompt for history
	diskAgent, err := Load(cfg, opts.AgentHandle)
	if err != nil {
		return fmt.Errorf("load agent: %w", err)
	}

	// Record refinement history
	if err := q.CreateAgentRefinement(ctx, db.CreateAgentRefinementParams{
		ID:             refinementID,
		AgentID:        agent.AgentID,
		PreviousPrompt: diskAgent.System,
		NewPrompt:      opts.NewPrompt,
		Reason:         opts.Reason,
		CreatedAt:      now,
	}); err != nil {
		return fmt.Errorf("record refinement: %w", err)
	}

	// Update prompt hash in database
	newHash := hashPrompt(opts.NewPrompt)
	if err := q.UpdateAyoAgentPrompt(ctx, db.UpdateAyoAgentPromptParams{
		CurrentPromptHash: sqlNullString(newHash),
		UpdatedAt:         now,
		AgentID:           agent.AgentID,
	}); err != nil {
		return fmt.Errorf("update agent: %w", err)
	}

	// Update on disk if requested
	if opts.UpdateOnDisk {
		if _, err := Save(cfg, diskAgent.Handle, diskAgent.Config, opts.NewPrompt); err != nil {
			return fmt.Errorf("save agent: %w", err)
		}
	}

	return nil
}

// ArchiveAgent archives an ayo-created agent (hidden but not deleted).
func ArchiveAgent(ctx context.Context, q db.Querier, agentHandle string) error {
	agent, err := q.GetAyoAgentByHandle(ctx, agentHandle)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	now := time.Now().Unix()
	return q.ArchiveAyoAgent(ctx, db.ArchiveAyoAgentParams{
		UpdatedAt: now,
		AgentID:   agent.AgentID,
	})
}

// UnarchiveAgent unarchives an ayo-created agent.
func UnarchiveAgent(ctx context.Context, q db.Querier, agentHandle string) error {
	agent, err := q.GetAyoAgentByHandle(ctx, agentHandle)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	now := time.Now().Unix()
	return q.UnarchiveAyoAgent(ctx, db.UnarchiveAyoAgentParams{
		UpdatedAt: now,
		AgentID:   agent.AgentID,
	})
}

// PromoteAgent promotes an ayo-created agent to a new handle.
// The old handle continues to work but is marked as promoted.
func PromoteAgent(ctx context.Context, cfg config.Config, q db.Querier, oldHandle, newHandle string) error {
	agent, err := q.GetAyoAgentByHandle(ctx, oldHandle)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	// TODO: Copy agent directory to new handle
	// For now just record the promotion
	now := time.Now().Unix()
	return q.PromoteAyoAgent(ctx, db.PromoteAyoAgentParams{
		PromotedTo: sqlNullString("@" + NormalizeHandle(newHandle)),
		UpdatedAt:  now,
		AgentID:    agent.AgentID,
	})
}

// ListAyoCreatedAgents lists all agents created by @ayo.
func ListAyoCreatedAgents(ctx context.Context, q db.Querier, includeArchived bool) ([]db.AyoCreatedAgent, error) {
	if includeArchived {
		return q.ListArchivedAyoAgents(ctx)
	}
	return q.ListAyoAgents(ctx)
}

// IsAyoCreated returns true if the agent was created by @ayo.
func IsAyoCreated(ctx context.Context, q db.Querier, agentHandle string) bool {
	_, err := q.GetAyoAgentByHandle(ctx, agentHandle)
	return err == nil
}

// hashPrompt returns a SHA256 hash of a prompt.
func hashPrompt(prompt string) string {
	h := sha256.Sum256([]byte(prompt))
	return hex.EncodeToString(h[:])
}

func boolPtr(b bool) *bool {
	return &b
}

func sqlNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func sqlNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: true}
}
