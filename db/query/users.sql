-- name: CreateUser :one
INSERT INTO users (
  first_name, last_name, email, password_hash
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT id, first_name, last_name, email FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserById :one
SELECT id, first_name, last_name, email FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT id, first_name, last_name, email
FROM users
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET
    first_name = COALESCE(sqlc.narg(first_name), first_name),
    last_name = COALESCE(sqlc.narg(last_name), last_name),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id, email, first_name, last_name, created_at, updated_at;

