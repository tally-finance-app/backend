CREATE TABLE users (
    id                  uuid PRIMARY KEY,
    email               varchar NOT NULL UNIQUE,
    password_hash       varchar NOT NULL,
    display_name        varchar NOT NULL,
    avatar_key          varchar,
    locale              varchar NOT NULL,
    reporting_currency  varchar NOT NULL,
    created_at          timestamptz NOT NULL,
    updated_at          timestamptz NOT NULL
);

COMMENT ON TABLE users IS 'Registered users of the Tally application.';

CREATE TABLE households (
    id             uuid PRIMARY KEY,
    name           varchar NOT NULL,
    admin_user_id  uuid NOT NULL REFERENCES users (id),
    created_at     timestamptz NOT NULL
);

CREATE INDEX idx_households_admin_user_id ON households (admin_user_id);

CREATE TABLE household_members (
    household_id  uuid NOT NULL REFERENCES households (id),
    user_id       uuid NOT NULL REFERENCES users (id),
    status        varchar NOT NULL,
    joined_at     timestamptz NOT NULL,
    PRIMARY KEY (household_id, user_id)
);

-- The primary key is (household_id, user_id), so household_id lookups already
-- use it as a leading-column prefix. The reverse direction — "which households
-- does this user belong to" — needs its own index.
CREATE INDEX idx_household_members_user_id ON household_members (user_id);

CREATE TABLE categories (
    id                   uuid PRIMARY KEY,
    user_id              uuid NOT NULL REFERENCES users (id),
    name                 varchar NOT NULL,
    parent_category_id   uuid REFERENCES categories (id),
    type                 varchar NOT NULL,
    color                varchar,
    icon                 varchar,
    created_at           timestamptz NOT NULL,
    updated_at           timestamptz NOT NULL,
    deleted_at           timestamptz
);

-- Serves `WHERE user_id = $1 AND deleted_at IS NULL`, the category list query.
CREATE INDEX idx_categories_user_id_active ON categories (user_id) WHERE deleted_at IS NULL;

-- Walking the category tree, and the FK target for parent deletes.
CREATE INDEX idx_categories_parent_id_active ON categories (parent_category_id)
    WHERE deleted_at IS NULL AND parent_category_id IS NOT NULL;

CREATE TABLE accounts (
    id                            uuid PRIMARY KEY,
    user_id                       uuid NOT NULL REFERENCES users (id),
    name                          varchar NOT NULL,
    type                          varchar NOT NULL,
    currency                      varchar NOT NULL,
    initial_balance_minor_units   bigint NOT NULL,
    color                         varchar NOT NULL,
    icon                          varchar NOT NULL,
    created_at                    timestamptz NOT NULL,
    updated_at                    timestamptz NOT NULL,
    deleted_at                    timestamptz
);

-- Serves `WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at, id`,
-- the account list query — created_at/id trail the filter column so the same
-- index scan also satisfies the ORDER BY, with no separate sort step needed.
CREATE INDEX idx_accounts_user_id_active ON accounts (user_id, created_at, id) WHERE deleted_at IS NULL;

CREATE TABLE credit_cards (
    id                        uuid PRIMARY KEY,
    user_id                   uuid NOT NULL REFERENCES users (id),
    name                      varchar NOT NULL,
    currency                  varchar NOT NULL,
    credit_limit_minor_units  bigint NOT NULL,
    close_day                 int NOT NULL,
    due_day                   int NOT NULL,
    color                     varchar NOT NULL,
    icon                      varchar NOT NULL,
    created_at                timestamptz NOT NULL,
    updated_at                timestamptz NOT NULL,
    deleted_at                timestamptz
);

CREATE INDEX idx_credit_cards_user_id_active ON credit_cards (user_id) WHERE deleted_at IS NULL;

CREATE TABLE statements (
    id                          uuid PRIMARY KEY,
    credit_card_id              uuid NOT NULL REFERENCES credit_cards (id),
    cycle_start_date            date NOT NULL,
    cycle_end_date              date NOT NULL,
    due_date                    date NOT NULL,
    total_amount_minor_units    bigint NOT NULL,
    paid_amount_minor_units     bigint NOT NULL,
    status                      varchar NOT NULL,
    created_at                  timestamptz NOT NULL
);

-- Listing a card's statements, newest cycle first.
CREATE INDEX idx_statements_credit_card_id_cycle
    ON statements (credit_card_id, cycle_start_date DESC);

CREATE TABLE transactions (
    id                              uuid PRIMARY KEY,
    source_type                     varchar NOT NULL,
    source_id                       uuid NOT NULL,
    category_id                     uuid REFERENCES categories (id),
    amount_minor_units              bigint NOT NULL,
    currency                        varchar NOT NULL,
    fx_rate_to_reporting            numeric(20,10) NOT NULL,
    converted_amount_minor_units    bigint NOT NULL,
    transaction_date                date NOT NULL,
    description                     varchar,
    shared_to_household             boolean NOT NULL DEFAULT false,
    statement_id                    uuid REFERENCES statements (id),
    created_at                      timestamptz NOT NULL,
    updated_at                      timestamptz NOT NULL,
    deleted_at                      timestamptz
);

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

CREATE TABLE statement_adjustments (
    id                          uuid PRIMARY KEY,
    statement_id                uuid NOT NULL REFERENCES statements (id),
    transaction_id              uuid NOT NULL REFERENCES transactions (id),
    adjustment_type             varchar NOT NULL,
    amount_delta_minor_units    bigint NOT NULL,
    created_at                  timestamptz NOT NULL
);

-- Reading one statement's append-only audit trail (§5.4).
CREATE INDEX idx_statement_adjustments_statement_id
    ON statement_adjustments (statement_id);

CREATE INDEX idx_statement_adjustments_transaction_id
    ON statement_adjustments (transaction_id);

CREATE TABLE transfers (
    id                     uuid PRIMARY KEY,
    from_type              varchar NOT NULL,
    from_id                uuid NOT NULL,
    to_type                varchar NOT NULL,
    to_id                  uuid NOT NULL,
    amount_minor_units     bigint NOT NULL,
    from_currency          varchar NOT NULL,
    to_currency            varchar NOT NULL,
    fx_rate                numeric(20,10) NOT NULL,
    date                   date NOT NULL,
    description            varchar,
    created_at             timestamptz NOT NULL,
    deleted_at             timestamptz
);

-- Balance computation reads both sides of a transfer separately (§5.2:
-- + transfers in, - transfers out), so each direction needs its own index.
CREATE INDEX idx_transfers_from_active ON transfers (from_type, from_id, date DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_transfers_to_active ON transfers (to_type, to_id, date DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE fx_rates (
    id              uuid PRIMARY KEY,
    currency_pair   varchar NOT NULL,
    rate            numeric(20,10) NOT NULL,
    date            date NOT NULL
);

CREATE UNIQUE INDEX idx_fx_rates_currency_pair_date ON fx_rates (currency_pair, date);
