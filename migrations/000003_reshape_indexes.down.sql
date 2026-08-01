-- Reverses 000003: drops the reshaped/added indexes and restores the originals
-- from 000001 exactly as they were, so an up/down/up round-trip is a no-op.

-- --- statement_adjustments --------------------------------------------------

DROP INDEX idx_statement_adjustments_transaction_id;
DROP INDEX idx_statement_adjustments_statement_id;

-- --- statements -------------------------------------------------------------

DROP INDEX idx_statements_credit_card_id_cycle;

-- --- households -------------------------------------------------------------

DROP INDEX idx_households_admin_user_id;

-- --- household_members ------------------------------------------------------

DROP INDEX idx_household_members_user_id;

-- --- transfers --------------------------------------------------------------

DROP INDEX idx_transfers_to_active;
DROP INDEX idx_transfers_from_active;

CREATE INDEX idx_transfers_deleted_at ON transfers (deleted_at) WHERE deleted_at IS NULL;

-- --- transactions -----------------------------------------------------------

DROP INDEX idx_transactions_category_active;
DROP INDEX idx_transactions_statement_active;
DROP INDEX idx_transactions_date_active;
DROP INDEX idx_transactions_source_active;

CREATE INDEX idx_transactions_transaction_date ON transactions (transaction_date);
CREATE INDEX idx_transactions_source_id ON transactions (source_id);
CREATE INDEX idx_transactions_deleted_at ON transactions (deleted_at) WHERE deleted_at IS NULL;

-- --- categories -------------------------------------------------------------

DROP INDEX idx_categories_parent_id_active;
DROP INDEX idx_categories_user_id_active;

CREATE INDEX idx_categories_deleted_at ON categories (deleted_at) WHERE deleted_at IS NULL;

-- --- credit_cards -----------------------------------------------------------

DROP INDEX idx_credit_cards_user_id_active;

CREATE INDEX idx_credit_cards_deleted_at ON credit_cards (deleted_at) WHERE deleted_at IS NULL;

-- --- accounts ---------------------------------------------------------------

DROP INDEX idx_accounts_user_id_active;

CREATE INDEX idx_accounts_deleted_at ON accounts (deleted_at) WHERE deleted_at IS NULL;
