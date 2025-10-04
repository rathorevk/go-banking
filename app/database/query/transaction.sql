-- name: CreateTransaction :one
INSERT INTO transactions (
  id,
  account_id,
  amount,
  source,
  type
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;


-- name: ListTransactionsByAccount :many
SELECT * FROM transactions
WHERE account_id = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: GetTransaction :one
SELECT * FROM transactions
WHERE id = $1 LIMIT 1;

-- name: ListTransactions :many
SELECT * FROM transactions
ORDER BY id
LIMIT $1
OFFSET $2;
