-- Reshape the soft-delete indexes and add the missing foreign-key indexes.
--
-- Two problems with the original set:
--
-- 1. `ON <table> (deleted_at) WHERE deleted_at IS NULL` indexes a column whose
--    value is, by the index's own predicate, always NULL inside the index. Every
--    live row shares one key, so the index can't narrow a search — it only ever
--    substitutes for a sequential scan when the whole table is being read. The
--    useful shape puts the *lookup* key in the index and keeps `deleted_at IS
--    NULL` as the predicate, which both narrows the search and shrinks the index
--    to live rows only.
--
-- 2. Postgres does not index the referencing side of a foreign key. Every list
--    endpoint is scoped by owner (`WHERE user_id = $1`), and none of those
--    columns had an index, so each one was a sequential scan. Unindexed FK
--    columns also make deletes/updates on the parent row scan the child table.
--
-- Plain CREATE INDEX, not CONCURRENTLY: golang-migrate runs each migration in a
-- transaction, and CONCURRENTLY cannot run inside one. Safe here because these
-- tables are empty. Once there's production data, index changes need to move to
-- CONCURRENTLY in a migration marked as non-transactional.
--
-- Note on TALLY-132: its acceptance criteria asked that soft-delete tables have
-- `deleted_at` "indexed". That intent is preserved — `deleted_at` still governs
-- every index below, as a partial predicate rather than as the key.

-- --- accounts ---------------------------------------------------------------

DROP INDEX idx_accounts_deleted_at;

-- Serves `WHERE user_id = $1 AND deleted_at IS NULL`, the account list query.
CREATE INDEX idx_accounts_user_id_active ON accounts (user_id) WHERE deleted_at IS NULL;

-- --- credit_cards -----------------------------------------------------------

DROP INDEX idx_credit_cards_deleted_at;

CREATE INDEX idx_credit_cards_user_id_active ON credit_cards (user_id) WHERE deleted_at IS NULL;

-- --- categories -------------------------------------------------------------

DROP INDEX idx_categories_deleted_at;

CREATE INDEX idx_categories_user_id_active ON categories (user_id) WHERE deleted_at IS NULL;

-- Walking the category tree, and the FK target for parent deletes.
CREATE INDEX idx_categories_parent_id_active ON categories (parent_category_id)
    WHERE deleted_at IS NULL AND parent_category_id IS NOT NULL;

-- --- transactions -----------------------------------------------------------

DROP INDEX idx_transactions_deleted_at;
DROP INDEX idx_transactions_source_id;
DROP INDEX idx_transactions_transaction_date;

-- The dominant access pattern: one account's or card's transactions, newest
-- first. transaction_date is part of the key so the ORDER BY is satisfied by the
-- index rather than a sort. source_type leads with source_id because queries
-- always know both (§5.1 — a transaction belongs to exactly one account or card).
CREATE INDEX idx_transactions_source_active
    ON transactions (source_type, source_id, transaction_date DESC)
    WHERE deleted_at IS NULL;

-- Date-range reports that span every source a user owns.
CREATE INDEX idx_transactions_date_active ON transactions (transaction_date)
    WHERE deleted_at IS NULL;

-- Recalculating a statement total, and locating the rows a StatementAdjustment
-- refers to (§5.4).
CREATE INDEX idx_transactions_statement_active ON transactions (statement_id)
    WHERE deleted_at IS NULL AND statement_id IS NOT NULL;

-- Spending-by-category reports, and the FK target for category deletes.
CREATE INDEX idx_transactions_category_active ON transactions (category_id)
    WHERE deleted_at IS NULL AND category_id IS NOT NULL;

-- --- transfers --------------------------------------------------------------

DROP INDEX idx_transfers_deleted_at;

-- Balance computation reads both sides of a transfer separately (§5.2:
-- + transfers in, - transfers out), so each direction needs its own index.
CREATE INDEX idx_transfers_from_active ON transfers (from_type, from_id, date DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_transfers_to_active ON transfers (to_type, to_id, date DESC)
    WHERE deleted_at IS NULL;

-- --- household_members ------------------------------------------------------

-- The primary key is (household_id, user_id), so household_id lookups already
-- use it as a leading-column prefix. The reverse direction — "which households
-- does this user belong to" — has no index at all.
CREATE INDEX idx_household_members_user_id ON household_members (user_id);

-- --- households -------------------------------------------------------------

CREATE INDEX idx_households_admin_user_id ON households (admin_user_id);

-- --- statements -------------------------------------------------------------

-- Listing a card's statements, newest cycle first.
CREATE INDEX idx_statements_credit_card_id_cycle
    ON statements (credit_card_id, cycle_start_date DESC);

-- --- statement_adjustments --------------------------------------------------

-- Reading one statement's append-only audit trail (§5.4).
CREATE INDEX idx_statement_adjustments_statement_id
    ON statement_adjustments (statement_id);

CREATE INDEX idx_statement_adjustments_transaction_id
    ON statement_adjustments (transaction_id);
