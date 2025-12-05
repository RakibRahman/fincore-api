-- name: CreateTransaction :one
INSERT INTO transactions (
  account_id,
  type,
  amount_cents,
  balance_after_cents
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetTransaction :one
SELECT * FROM transactions
WHERE id = $1 LIMIT 1;

-- name: ListTransactions :many
SELECT * FROM transactions
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;
