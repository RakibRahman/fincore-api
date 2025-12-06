-- name: CreateUser :one
INSERT INTO users (
  first_name, last_name, email, password_hash
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
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
    first_name = $2,
    last_name = $3,
    email = $4
WHERE id = $1
RETURNING id, first_name, last_name, email;

--name 