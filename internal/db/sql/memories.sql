-- name: CreateMemory :exec
INSERT INTO memories (
    id, agent_handle, path_scope, content, category, embedding,
    source_session_id, source_message_id, created_at, updated_at,
    confidence, status
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetMemory :one
SELECT * FROM memories WHERE id = ?;

-- name: UpdateMemory :exec
UPDATE memories SET
    content = ?,
    category = ?,
    embedding = ?,
    confidence = ?,
    updated_at = ?
WHERE id = ?;

-- name: UpdateMemoryAccess :exec
UPDATE memories SET
    last_accessed_at = ?,
    access_count = access_count + 1
WHERE id = ?;

-- name: SupersedeMemory :exec
UPDATE memories SET
    status = 'superseded',
    superseded_by_id = ?,
    updated_at = ?
WHERE id = ?;

-- name: ForgetMemory :exec
UPDATE memories SET
    status = 'forgotten',
    updated_at = ?
WHERE id = ?;

-- name: DeleteMemory :exec
DELETE FROM memories WHERE id = ?;

-- name: ListMemories :many
SELECT * FROM memories
WHERE status = COALESCE(sqlc.narg(status), 'active')
ORDER BY created_at DESC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: ListMemoriesByAgent :many
SELECT * FROM memories
WHERE agent_handle = sqlc.arg(agent)
  AND status = COALESCE(sqlc.narg(status), 'active')
ORDER BY created_at DESC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: ListMemoriesByPath :many
SELECT * FROM memories
WHERE (path_scope = sqlc.arg(path) OR path_scope IS NULL)
  AND status = COALESCE(sqlc.narg(status), 'active')
ORDER BY 
    CASE WHEN path_scope IS NOT NULL THEN 0 ELSE 1 END,
    created_at DESC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: ListMemoriesByAgentAndPath :many
SELECT * FROM memories
WHERE (agent_handle = sqlc.arg(agent) OR agent_handle IS NULL)
  AND (path_scope = sqlc.arg(path) OR path_scope IS NULL)
  AND status = COALESCE(sqlc.narg(status), 'active')
ORDER BY 
    CASE WHEN agent_handle IS NOT NULL THEN 0 ELSE 1 END,
    CASE WHEN path_scope IS NOT NULL THEN 0 ELSE 1 END,
    created_at DESC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: ListMemoriesByCategory :many
SELECT * FROM memories
WHERE category = sqlc.arg(cat)
  AND status = COALESCE(sqlc.narg(status), 'active')
ORDER BY created_at DESC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: GetMemoryHistory :many
WITH RECURSIVE chain AS (
    SELECT m.id, m.content, m.category, m.status, m.supersedes_id, m.superseded_by_id, 
           m.supersession_reason, m.created_at, 0 as depth
    FROM memories m WHERE m.id = ?
    
    UNION ALL
    
    SELECT m.id, m.content, m.category, m.status, m.supersedes_id, m.superseded_by_id,
           m.supersession_reason, m.created_at, c.depth + 1
    FROM memories m
    JOIN chain c ON m.id = c.supersedes_id
    WHERE c.depth < 100
)
SELECT * FROM chain ORDER BY depth;

-- name: CountMemories :one
SELECT COUNT(*) FROM memories WHERE status = COALESCE(sqlc.narg(status), 'active');

-- name: CountMemoriesByAgent :one
SELECT COUNT(*) FROM memories 
WHERE agent_handle = ?
  AND status = COALESCE(sqlc.narg(status), 'active');

-- name: GetAllActiveMemoriesWithEmbeddings :many
SELECT id, agent_handle, path_scope, content, category, embedding, confidence,
       last_accessed_at, access_count, created_at
FROM memories
WHERE status = 'active'
  AND embedding IS NOT NULL;

-- name: GetMemoriesForSearch :many
SELECT id, agent_handle, path_scope, content, category, embedding, confidence,
       last_accessed_at, access_count, created_at
FROM memories
WHERE status = 'active'
  AND embedding IS NOT NULL
  AND (agent_handle = sqlc.narg(agent_handle) OR agent_handle IS NULL OR sqlc.narg(agent_handle) IS NULL)
  AND (path_scope = sqlc.narg(path_scope) OR path_scope IS NULL OR sqlc.narg(path_scope) IS NULL);

-- name: ClearMemoriesByAgent :exec
UPDATE memories SET
    status = 'forgotten',
    updated_at = ?
WHERE agent_handle = ?;

-- name: ClearAllMemories :exec
UPDATE memories SET
    status = 'forgotten',
    updated_at = ?;
