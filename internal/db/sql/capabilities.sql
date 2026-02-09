-- name: CreateCapability :exec
INSERT INTO agent_capabilities (
    id,
    agent_id,
    name,
    description,
    confidence,
    source,
    embedding,
    input_hash,
    created_at,
    updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetCapabilitiesByAgent :many
SELECT * FROM agent_capabilities
WHERE agent_id = ?
ORDER BY confidence DESC;

-- name: GetCapabilityByName :one
SELECT * FROM agent_capabilities
WHERE agent_id = ? AND name = ?;

-- name: GetCapabilitiesByHash :many
SELECT * FROM agent_capabilities
WHERE input_hash = ?;

-- name: DeleteCapabilitiesByAgent :exec
DELETE FROM agent_capabilities WHERE agent_id = ?;

-- name: ListAllCapabilities :many
SELECT * FROM agent_capabilities
ORDER BY confidence DESC;

-- name: SearchCapabilitiesByName :many
SELECT * FROM agent_capabilities
WHERE name LIKE ? OR description LIKE ?
ORDER BY confidence DESC
LIMIT ?;

-- name: UpdateCapabilityEmbedding :exec
UPDATE agent_capabilities
SET embedding = ?,
    updated_at = ?
WHERE id = ?;
