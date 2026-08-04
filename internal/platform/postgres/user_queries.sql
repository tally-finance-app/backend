-- name: CreateUser :one
INSERT INTO users (
    id, email, password_hash, display_name, locale, reporting_currency, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ExistsUserByEmail :one
SELECT EXISTS (
    SELECT 1 FROM users WHERE email = $1
);
