# ER Diagram

> Migrated from Notion on 2026-08-03. Reflects the schema built by `migrations/000001_initial_schema.up.sql`.
> Note: Mermaid's `erDiagram` syntax shown below doesn't annotate nullability, so it won't visibly
> show which columns are `NOT NULL` — check the migration or `.docs/requirements-and-domain-model.md`
> for that.

```mermaid
erDiagram
    USER ||--o{ ACCOUNT : owns
    USER ||--o{ CREDIT_CARD : owns
    USER ||--o{ HOUSEHOLD_MEMBER : "is member"
    USER ||--o| HOUSEHOLD : administers
    USER ||--o{ CATEGORY : owns

    HOUSEHOLD ||--o{ HOUSEHOLD_MEMBER : has

    ACCOUNT ||--o{ TRANSACTION : "source of"
    CREDIT_CARD ||--o{ TRANSACTION : "source of"
    CREDIT_CARD ||--o{ STATEMENT : generates

    STATEMENT ||--o{ TRANSACTION : includes
    STATEMENT ||--o{ STATEMENT_ADJUSTMENT : logs
    TRANSACTION ||--o{ STATEMENT_ADJUSTMENT : triggers

    CATEGORY ||--o{ TRANSACTION : classifies
    CATEGORY ||--o{ CATEGORY : "parent of"

    ACCOUNT ||--o{ TRANSFER : "from/to"
    CREDIT_CARD ||--o{ TRANSFER : "from/to"

    USER {
        uuid id PK
        string email
        string password_hash
        string display_name
        string avatar_url
        string locale
        string reporting_currency
        timestamp created_at
        timestamp updated_at
    }

    HOUSEHOLD {
        uuid id PK
        string name
        uuid admin_user_id FK
        timestamp created_at
    }

    HOUSEHOLD_MEMBER {
        uuid household_id FK
        uuid user_id FK
        string status
        timestamp joined_at
    }

    ACCOUNT {
        uuid id PK
        uuid user_id FK
        string name
        string type
        string currency
        bigint initial_balance_minor_units
        string color
        string icon
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    CREDIT_CARD {
        uuid id PK
        uuid user_id FK
        string name
        string currency
        bigint credit_limit_minor_units
        int close_day
        int due_day
        string color
        string icon
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    STATEMENT {
        uuid id PK
        uuid credit_card_id FK
        date cycle_start_date
        date cycle_end_date
        date due_date
        bigint total_amount_minor_units
        bigint paid_amount_minor_units
        string status
        timestamp created_at
    }

    STATEMENT_ADJUSTMENT {
        uuid id PK
        uuid statement_id FK
        uuid transaction_id FK
        string adjustment_type
        bigint amount_delta_minor_units
        timestamp created_at
    }

    TRANSACTION {
        uuid id PK
        string source_type
        uuid source_id
        uuid category_id FK
        bigint amount_minor_units
        string currency
        numeric fx_rate_to_reporting
        bigint converted_amount_minor_units
        date transaction_date
        string description
        bool shared_to_household
        uuid statement_id FK
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    TRANSFER {
        uuid id PK
        string from_type
        uuid from_id
        string to_type
        uuid to_id
        bigint amount_minor_units
        string from_currency
        string to_currency
        numeric fx_rate
        date date
        string description
        timestamp created_at
        timestamp deleted_at
    }

    CATEGORY {
        uuid id PK
        uuid user_id FK
        string key
        string name
        uuid parent_category_id FK
        string type
        string color
        string icon
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    FX_RATE {
        uuid id PK
        string currency_pair
        numeric rate
        date date
    }
```

## Changelog

- **2026-08-03** — Migrated this doc from Notion ("ER Diagram") into the repo; no content changes.
  The Notion page now points here.
