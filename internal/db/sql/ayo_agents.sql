-- name: CreateAyoAgent :exec
INSERT INTO ayo_created_agents (
    agent_id,
    agent_handle,
    created_by,
    creation_reason,
    original_prompt,
    current_prompt_hash,
    created_at,
    updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAyoAgent :one
SELECT * FROM ayo_created_agents WHERE agent_id = ?;

-- name: GetAyoAgentByHandle :one
SELECT * FROM ayo_created_agents WHERE agent_handle = ?;

-- name: ListAyoAgents :many
SELECT * FROM ayo_created_agents
WHERE is_archived = 0
ORDER BY last_used_at DESC NULLS LAST, created_at DESC;

-- name: ListArchivedAyoAgents :many
SELECT * FROM ayo_created_agents
WHERE is_archived = 1
ORDER BY updated_at DESC;

-- name: UpdateAyoAgentInvocation :exec
UPDATE ayo_created_agents
SET invocation_count = invocation_count + 1,
    last_used_at = ?,
    updated_at = ?
WHERE agent_id = ?;

-- name: UpdateAyoAgentSuccess :exec
UPDATE ayo_created_agents
SET success_count = success_count + 1,
    confidence = MIN(1.0, confidence + 0.05),
    updated_at = ?
WHERE agent_id = ?;

-- name: UpdateAyoAgentFailure :exec
UPDATE ayo_created_agents
SET failure_count = failure_count + 1,
    confidence = MAX(0.0, confidence - 0.1),
    updated_at = ?
WHERE agent_id = ?;

-- name: UpdateAyoAgentPrompt :exec
UPDATE ayo_created_agents
SET current_prompt_hash = ?,
    refinement_count = refinement_count + 1,
    updated_at = ?
WHERE agent_id = ?;

-- name: ArchiveAyoAgent :exec
UPDATE ayo_created_agents
SET is_archived = 1,
    updated_at = ?
WHERE agent_id = ?;

-- name: UnarchiveAyoAgent :exec
UPDATE ayo_created_agents
SET is_archived = 0,
    updated_at = ?
WHERE agent_id = ?;

-- name: PromoteAyoAgent :exec
UPDATE ayo_created_agents
SET promoted_to = ?,
    updated_at = ?
WHERE agent_id = ?;

-- name: DeleteAyoAgent :exec
DELETE FROM ayo_created_agents WHERE agent_id = ?;

-- name: CreateAgentRefinement :exec
INSERT INTO agent_refinements (
    id,
    agent_id,
    previous_prompt,
    new_prompt,
    reason,
    created_at
) VALUES (?, ?, ?, ?, ?, ?);

-- name: ListAgentRefinements :many
SELECT * FROM agent_refinements
WHERE agent_id = ?
ORDER BY created_at DESC;

-- name: GetLatestRefinement :one
SELECT * FROM agent_refinements
WHERE agent_id = ?
ORDER BY created_at DESC
LIMIT 1;
