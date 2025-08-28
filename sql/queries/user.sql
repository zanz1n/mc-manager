-- name: UserGetById :one
SELECT * FROM users WHERE id = $1;

-- name: UserGetByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UserGetByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: UserGetMany :many
SELECT * FROM users
WHERE id < sqlc.arg(last_seen)
ORDER BY id DESC
LIMIT sqlc.arg(lim);

-- name: UserCreate :one
INSERT INTO users (
    id,
    username,
    first_name,
    last_name,
    minecraft_user,
    email,
    admin,
    two_fa,
    password
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;

-- name: UserDelete :one
DELETE FROM users WHERE id = $1 RETURNING *;
