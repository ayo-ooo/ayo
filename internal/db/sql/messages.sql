-- name: CreateMessage :one
INSERT INTO messages (
    id,
    session_id,
    role,
    parts,
    model,
    provider,
    created_at,
    updated_at,
    finished_at
) VALUES (
    @id, @session_id, @role, @parts, @model, @provider, strftime('%s', 'now'), strftime('%s', 'now'), NULL
) RETURNING *;

-- name: GetMessage :one
SELECT * FROM messages WHERE id = @id LIMIT 1;

-- name: ListMessagesBySession :many
SELECT * FROM messages WHERE session_id = @session_id ORDER BY created_at ASC;

-- name: UpdateMessage :exec
UPDATE messages SET
    parts = @parts,
    finished_at = @finished_at,
    updated_at = strftime('%s', 'now')
WHERE id = @id;

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = @id;

-- name: DeleteMessagesBySession :exec
DELETE FROM messages WHERE session_id = @session_id;

-- name: CountMessagesBySession :one
SELECT COUNT(*) FROM messages WHERE session_id = @session_id;
