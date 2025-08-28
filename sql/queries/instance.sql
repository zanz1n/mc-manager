-- name: InstanceGetById :one
SELECT * FROM instances WHERE id = $1;

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

-- name: InstanceDelete :one
DELETE FROM instances WHERE id = $1 RETURNING *;
