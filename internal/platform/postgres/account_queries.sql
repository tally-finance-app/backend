-- name: CreateAccount :one
INSERT INTO accounts (
    id, user_id, name, type, currency, initial_balance_minor_units, color, icon, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetAccountByID :one
SELECT *
FROM accounts
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListAccountsByFilters :many
-- sort_by/sort_dir are a closed, validated enum (see account.SortBy /
-- account.SortDirection) — never a raw client string — so this stays fully
-- parameterized with no dynamic SQL construction. Only created_at is backed by
-- an index (idx_accounts_user_id_active); name/type/currency sort over the
-- already user_id-filtered rows in memory, which is fine at this table's
-- per-user cardinality (see CLAUDE.md §6).
SELECT * FROM accounts
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (sqlc.narg('type')::text IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('currency')::text IS NULL OR currency = sqlc.narg('currency'))
ORDER BY
  CASE WHEN sqlc.arg('sort_by')::text = 'created_at' AND sqlc.arg('sort_dir')::text = 'asc'  THEN created_at END ASC,
  CASE WHEN sqlc.arg('sort_by')::text = 'created_at' AND sqlc.arg('sort_dir')::text = 'desc' THEN created_at END DESC,
  CASE WHEN sqlc.arg('sort_by')::text = 'name'       AND sqlc.arg('sort_dir')::text = 'asc'  THEN name END ASC,
  CASE WHEN sqlc.arg('sort_by')::text = 'name'       AND sqlc.arg('sort_dir')::text = 'desc' THEN name END DESC,
  CASE WHEN sqlc.arg('sort_by')::text = 'type'       AND sqlc.arg('sort_dir')::text = 'asc'  THEN type END ASC,
  CASE WHEN sqlc.arg('sort_by')::text = 'type'       AND sqlc.arg('sort_dir')::text = 'desc' THEN type END DESC,
  CASE WHEN sqlc.arg('sort_by')::text = 'currency'   AND sqlc.arg('sort_dir')::text = 'asc'  THEN currency END ASC,
  CASE WHEN sqlc.arg('sort_by')::text = 'currency'   AND sqlc.arg('sort_dir')::text = 'desc' THEN currency END DESC,
  id
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountAccountsByFilters :one
SELECT count(*) FROM accounts
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (sqlc.narg('type')::text IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('currency')::text IS NULL OR currency = sqlc.narg('currency'));

-- name: UpdateAccountByID :one
UPDATE accounts
SET name = $2, type = $3, icon = $4, color = $5, updated_at = $6
WHERE id = $1 AND user_id = $7 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteAccount :one
-- account_exists is scoped to this user too: a row that exists but belongs to
-- someone else must look identical to a row that doesn't exist at all, so the
-- response never confirms another user's account ID is valid.
WITH updated AS (
    UPDATE accounts
    SET deleted_at = $3
    WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
    RETURNING id
)
SELECT
    EXISTS (SELECT 1 FROM accounts WHERE accounts.id = $1 AND accounts.user_id = $2) AS account_exists,
    EXISTS (SELECT 1 FROM updated WHERE updated.id = $1) AS was_deleted;
