-- name: NodeGetById :one
SELECT * FROM nodes WHERE id = $1;

-- name: NodeGetMany :many
SELECT * FROM nodes
WHERE id < sqlc.arg(last_seen)
ORDER BY id DESC
LIMIT sqlc.arg(lim);

-- name: NodeCreate :one
INSERT INTO nodes (
    id,
    name,
    description,
    token,
    endpoint,
    endpoint_tls,
    ftp_port,
    grpc_port
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: NodeUpdate :one
UPDATE nodes SET
    updated_at = now(),
    name = sqlc.arg(name),
    description = sqlc.arg(description),
    maintenance = sqlc.arg(maintenance),
    token = sqlc.arg(token),
    endpoint = sqlc.arg(endpoint),
    endpoint_tls = sqlc.arg(endpoint_tls),
    ftp_port = sqlc.arg(ftp_port),
    grpc_port = sqlc.arg(grpc_port)
WHERE id = $1
RETURNING *;

-- name: NodeDelete :one
DELETE FROM nodes WHERE id = $1 RETURNING *;
