CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email               varchar NOT NULL UNIQUE,
    password_hash       varchar NOT NULL,
    display_name        varchar NOT NULL,
    avatar_url          varchar,
    locale              varchar NOT NULL,
    reporting_currency  varchar NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE households (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name           varchar NOT NULL,
    admin_user_id  uuid NOT NULL REFERENCES users (id),
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE household_members (
    household_id  uuid NOT NULL REFERENCES households (id),
    user_id       uuid NOT NULL REFERENCES users (id),
    status        varchar NOT NULL,
    joined_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (household_id, user_id)
);

CREATE TABLE categories (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              uuid NOT NULL REFERENCES users (id),
    key                  varchar NOT NULL,
    name                 varchar NOT NULL,
    parent_category_id   uuid REFERENCES categories (id),
    type                 varchar NOT NULL,
    color                varchar,
    icon                 varchar,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz
);

CREATE INDEX idx_categories_deleted_at ON categories (deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE accounts (
    id                            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                       uuid NOT NULL REFERENCES users (id),
    name                          varchar NOT NULL,
    type                          varchar NOT NULL,
    currency                      varchar NOT NULL,
    initial_balance_minor_units   bigint NOT NULL,
    color                         varchar,
    icon                          varchar,
    created_at                    timestamptz NOT NULL DEFAULT now(),
    updated_at                    timestamptz NOT NULL DEFAULT now(),
    deleted_at                    timestamptz
);

CREATE INDEX idx_accounts_deleted_at ON accounts (deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE credit_cards (
    id                        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                   uuid NOT NULL REFERENCES users (id),
    name                      varchar NOT NULL,
    currency                  varchar NOT NULL,
    credit_limit_minor_units  bigint NOT NULL,
    close_day                 int NOT NULL,
    due_day                   int NOT NULL,
    color                     varchar,
    icon                      varchar,
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now(),
    deleted_at                timestamptz
);

CREATE INDEX idx_credit_cards_deleted_at ON credit_cards (deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE statements (
    id                          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_card_id              uuid NOT NULL REFERENCES credit_cards (id),
    cycle_start_date            date NOT NULL,
    cycle_end_date              date NOT NULL,
    due_date                    date NOT NULL,
    total_amount_minor_units    bigint NOT NULL,
    paid_amount_minor_units     bigint NOT NULL,
    status                      varchar NOT NULL,
    created_at                  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE transactions (
    id                              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
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
    created_at                      timestamptz NOT NULL DEFAULT now(),
    updated_at                      timestamptz NOT NULL DEFAULT now(),
    deleted_at                      timestamptz
);

CREATE INDEX idx_transactions_transaction_date ON transactions (transaction_date);
CREATE INDEX idx_transactions_source_id ON transactions (source_id);
CREATE INDEX idx_transactions_deleted_at ON transactions (deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE statement_adjustments (
    id                          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    statement_id                uuid NOT NULL REFERENCES statements (id),
    transaction_id              uuid NOT NULL REFERENCES transactions (id),
    adjustment_type             varchar NOT NULL,
    amount_delta_minor_units    bigint NOT NULL,
    created_at                  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE transfers (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
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
    created_at             timestamptz NOT NULL DEFAULT now(),
    deleted_at             timestamptz
);

CREATE INDEX idx_transfers_deleted_at ON transfers (deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE fx_rates (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    currency_pair   varchar NOT NULL,
    rate            numeric(20,10) NOT NULL,
    date            date NOT NULL
);

CREATE UNIQUE INDEX idx_fx_rates_currency_pair_date ON fx_rates (currency_pair, date);
