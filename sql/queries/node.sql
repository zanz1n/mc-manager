-- name: NodeGetById :one
SELECT * FROM nodes WHERE id = $1;
