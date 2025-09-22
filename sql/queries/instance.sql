-- name: InstanceGetById :one
SELECT * FROM instances WHERE id = $1;

-- name: InstanceGetMany :many
SELECT id, user_id, node_id, created_at, updated_at, last_launched, name, description, version, version_distro, maintenance FROM instances
WHERE id < sqlc.arg(last_seen)
ORDER BY id DESC
LIMIT sqlc.arg(lim);

-- name: InstanceGetByUser :many
SELECT id, user_id, node_id, created_at, updated_at, last_launched, name, description, version, version_distro, maintenance FROM instances
WHERE user_id = sqlc.arg(user_id)
AND id < sqlc.arg(last_seen)
ORDER BY id DESC
LIMIT sqlc.arg(lim);

-- name: InstanceGetByNode :many
SELECT id, user_id, node_id, created_at, updated_at, last_launched, name, description, version, version_distro, maintenance FROM instances
WHERE node_id = sqlc.arg(node_id)
AND id < sqlc.arg(last_seen)
ORDER BY id DESC
LIMIT sqlc.arg(lim);

-- name: InstanceCreate :one
INSERT INTO instances (
    id,
    user_id,
    node_id,
    name,
    description,
    version,
    version_distro,
    config,
    limits
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;

-- name: InstanceUpdate :one
UPDATE instances SET
    updated_at = now(),
    name = sqlc.arg(name),
    description = sqlc.arg(description),
    version = sqlc.arg(version),
    version_distro = sqlc.arg(version_distro),
    maintenance = sqlc.arg(maintenance)
WHERE id = $1
RETURNING *;

-- name: InstanceUpdateConfig :one
UPDATE instances SET
    updated_at = now(),
    config = sqlc.arg(config)
WHERE id = $1
RETURNING *;

-- name: InstanceUpdateLimits :one
UPDATE instances SET
    updated_at = now(),
    limits = sqlc.arg(limits)
WHERE id = $1
RETURNING *;

-- name: InstanceUpdateLastLaunched :exec
UPDATE instances SET last_launched = now() WHERE id = $1;

-- name: InstanceDelete :one
DELETE FROM instances WHERE id = $1 RETURNING *;
