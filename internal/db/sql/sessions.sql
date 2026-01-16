-- name: CreateSession :one
INSERT INTO sessions (
    id,
    agent_handle,
    title,
    input_schema,
    output_schema,
    structured_input,
    structured_output,
    chain_depth,
    chain_source,
    message_count,
    created_at,
    updated_at,
    finished_at
) VALUES (
    @id, @agent_handle, @title, @input_schema, @output_schema, @structured_input, @structured_output, @chain_depth, @chain_source, 0, strftime('%s', 'now'), strftime('%s', 'now'), NULL
) RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions WHERE id = @id LIMIT 1;

-- name: GetSessionByPrefix :many
SELECT * FROM sessions WHERE id LIKE @prefix || '%' ORDER BY updated_at DESC LIMIT 10;

-- name: ListSessions :many
SELECT * FROM sessions ORDER BY updated_at DESC LIMIT @limit;

-- name: ListSessionsByAgent :many
SELECT * FROM sessions WHERE agent_handle = @agent_handle ORDER BY updated_at DESC LIMIT @limit;

-- name: SearchSessionsByTitle :many
SELECT * FROM sessions WHERE title LIKE '%' || @query || '%' ORDER BY updated_at DESC LIMIT @limit;

-- name: UpdateSession :one
UPDATE sessions SET
    title = @title,
    structured_output = @structured_output,
    finished_at = @finished_at,
    updated_at = strftime('%s', 'now')
WHERE id = @id
RETURNING *;

-- name: UpdateSessionTitle :exec
UPDATE sessions SET
    title = @title,
    updated_at = strftime('%s', 'now')
WHERE id = @id;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = @id;

-- name: CountSessions :one
SELECT COUNT(*) FROM sessions;

-- name: CountSessionsByAgent :one
SELECT COUNT(*) FROM sessions WHERE agent_handle = @agent_handle;
