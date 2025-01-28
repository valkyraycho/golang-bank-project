-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: ListAccount :many
SELECT * FROM accounts
WHERE owner_id = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: CreateAccount :one
INSERT INTO accounts (
  owner_id,
  balance,
  currency
) VALUES (
  $1, $2, $3
)RETURNING *;

-- name: UpdateAccount :one
UPDATE accounts
SET
  balance = $2
WHERE id = $1
RETURNING *;

-- name: AddAccountBalance :one
UPDATE accounts
SET
  balance = balance + sqlc.arg(amount)
WHERE id = sqlc.arg(id)
RETURNING *;


-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;