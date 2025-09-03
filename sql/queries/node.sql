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

-- name: NodeDelete :one
DELETE FROM nodes WHERE id = $1 RETURNING *;
