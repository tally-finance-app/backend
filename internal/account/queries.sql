-- name: CreateAccount :one
INSERT INTO accounts (
    user_id, name, type, currency, initial_balance_minor_units, color, icon
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetAccountByID :one
SELECT *
FROM accounts
WHERE id = $1 AND deleted_at IS NULL;
