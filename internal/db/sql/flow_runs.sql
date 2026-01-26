-- name: CreateFlowRun :one
INSERT INTO flow_runs (
    id,
    flow_name,
    flow_path,
    flow_source,
    status,
    input_json,
    input_validated,
    started_at,
    parent_run_id,
    session_id
) VALUES (
    @id,
    @flow_name,
    @flow_path,
    @flow_source,
    'running',
    @input_json,
    @input_validated,
    @started_at,
    @parent_run_id,
    @session_id
) RETURNING *;

-- name: CompleteFlowRun :one
UPDATE flow_runs SET
    status = @status,
    exit_code = @exit_code,
    error_message = @error_message,
    output_json = @output_json,
    stderr_log = @stderr_log,
    output_validated = @output_validated,
    finished_at = @finished_at,
    duration_ms = @duration_ms
WHERE id = @id
RETURNING *;

-- name: GetFlowRun :one
SELECT * FROM flow_runs WHERE id = @id LIMIT 1;

-- name: GetFlowRunByPrefix :many
SELECT * FROM flow_runs WHERE id LIKE @prefix || '%' ORDER BY started_at DESC LIMIT 10;

-- name: ListFlowRuns :many
SELECT * FROM flow_runs ORDER BY started_at DESC LIMIT @limit;

-- name: ListFlowRunsByName :many
SELECT * FROM flow_runs WHERE flow_name = @flow_name ORDER BY started_at DESC LIMIT @limit;

-- name: ListFlowRunsByStatus :many
SELECT * FROM flow_runs WHERE status = @status ORDER BY started_at DESC LIMIT @limit;

-- name: ListFlowRunsBySession :many
SELECT * FROM flow_runs WHERE session_id = @session_id ORDER BY started_at DESC;

-- name: GetLastFlowRun :one
SELECT * FROM flow_runs WHERE flow_name = @flow_name ORDER BY started_at DESC LIMIT 1;

-- name: DeleteFlowRun :exec
DELETE FROM flow_runs WHERE id = @id;

-- name: PruneFlowRunsByAge :exec
DELETE FROM flow_runs WHERE started_at < @cutoff_timestamp;

-- name: PruneFlowRunsByCount :exec
DELETE FROM flow_runs WHERE id NOT IN (
    SELECT id FROM flow_runs ORDER BY started_at DESC LIMIT @keep_count
);

-- name: CountFlowRuns :one
SELECT COUNT(*) FROM flow_runs;

-- name: CountFlowRunsByName :one
SELECT COUNT(*) FROM flow_runs WHERE flow_name = @flow_name;

-- name: CountFlowRunsByStatus :one
SELECT COUNT(*) FROM flow_runs WHERE status = @status;
