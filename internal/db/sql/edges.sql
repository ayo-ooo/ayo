-- name: CreateEdge :exec
INSERT INTO session_edges (
    parent_id,
    child_id,
    edge_type,
    trigger_message_id,
    created_at
) VALUES (
    @parent_id, @child_id, @edge_type, @trigger_message_id, strftime('%s', 'now')
);

-- name: GetParentEdges :many
SELECT * FROM session_edges WHERE child_id = @child_id;

-- name: GetChildEdges :many
SELECT * FROM session_edges WHERE parent_id = @parent_id;

-- name: DeleteEdge :exec
DELETE FROM session_edges WHERE parent_id = @parent_id AND child_id = @child_id;

-- name: DeleteEdgesBySession :exec
DELETE FROM session_edges WHERE parent_id = @session_id OR child_id = @session_id;
