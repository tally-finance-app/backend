# Requirements & Domain Model (MVP)

> Canonical spec for domain entities, business rules, and MVP scope. Migrated from Notion on
> 2026-08-03 — see the Changelog at the bottom for what's changed since, and CLAUDE.md for how
> this file fits into day-to-day workflow (it's the condensed, day-to-day duplicate of this one).

## 1. Purpose

A personal project to deepen system design and data structures/algorithms knowledge by building a
real, non-trivial application: a multi-user, multi-currency personal finance tracker with
household sharing and credit card statement management.

**Primary goal:** learning system design and DS&A deeply. Building a genuinely usable tool is a
secondary (nice-to-have) outcome, not the driver of scope decisions.

## 2. Stack

- **Language:** Go (backend)
- **Architecture:** Backend-first (API + DB), no frontend until backend is solid. **Light DDD**
  within a single Go module — tactical patterns (entities with enforced invariants, repository
  interfaces owned by the domain, a service/use-case layer separate from HTTP handlers) without
  full DDD ceremony (no formal aggregates, domain events, CQRS, or bounded-context mapping). The
  domain has enough real complexity (statement lifecycle, FX snapshotting, idempotency, household
  visibility) to deserve separation from HTTP handlers, but this is a solo project — full DDD's
  organizational overhead mainly pays off coordinating teams, not solo learning. Package layout:
  one package per domain concept under `internal/` (`account`, `creditcard`, `transaction`,
  `transfer`, `category`, `household`, `statement`, `fx`, `user`), plus `internal/platform/postgres`
  (repository implementations) and `internal/transport/http` (handlers/routing), with `cmd/api`
  and `cmd/jobs` as entrypoints.
- **Database access:** Raw SQL via `sqlc` (typed Go generated from hand-written SQL) + `pgx` driver
  + `golang-migrate` for schema migrations — deliberately not an ORM, so that index usage, join
  cost, and query shape stay visible rather than hidden behind an abstraction layer. Migrations are
  applied via `cmd/migrate` (`make migrate-up` / `make migrate-down` / `make migrate-verify`), a
  thin wrapper around the `golang-migrate` library rather than a dependency on its own `migrate`
  CLI binary — this way the module only pulls in the Postgres (`pgx/v5`) driver it actually uses
  instead of every database `golang-migrate` supports.
- **Frontend (future phase):** Angular + TypeScript, consuming the REST API once it's stable.
- **Deployment target:** Cloud (containerized, e.g. Docker + Fly.io/Render/VPS)
- **Logging:** Structured logging from the start (`log/slog`), with a request/job correlation ID
  on every log line — cheap to build in now, expensive to retrofit later. Full observability
  (metrics, tracing) is deferred to Phase 7 of the roadmap.
- **Localization:** the frontend will be localized, so the API stays locale-agnostic overall — ISO
  dates, minor-unit money amounts. Category names are the one piece of user-facing text that
  originates server-side; they're translated once, at seed time (per user's `locale`), when a new
  user's default categories are created — not translated at runtime on every request.

## 3. Engineering Workflow

**Task tracking:** Linear, project "Go Backend" (team: Tally Finance App). Work is organized as
epics (one per feature area, matching this doc's domain areas) with user stories underneath.

**Source control:** GitHub. Chosen mainly because Epic 0's CI story already specifies GitHub
Actions — a different host would mean rewriting that ticket or bolting on a mismatched CI system.
Also the best ecosystem fit for Go tooling (`golangci-lint` actions, Dependabot for Go module
updates, most Go library docs assume GitHub).

**Task format:** every story follows a fixed structure so tickets read like a PM would actually
write them, not just a one-line description:

- **Background** — why this ticket exists, what it connects to
- **User story** — As a [role], I want [feature], so that [benefit]
- **Scope** — the concrete endpoints/components this ticket covers
- **Acceptance criteria** — specific, testable behaviors, including edge cases (not just the happy
  path)
- **Out of scope** — what's deliberately excluded, to prevent scope creep
- **Dependencies** — what this needs first, and what it blocks
- **Definition of done** — the specific test/verification that proves it works

**Epic 0 — Codebase Setup & Foundation:** the prerequisite epic before any feature epic can be
implemented. Covers: Go module + package layout, Docker Compose (local Postgres), migrations
(`golang-migrate`) building the schema from the ER Diagram, `sqlc` code generation, HTTP server
skeleton (routing + middleware), typed config loading, CI pipeline (build/vet/lint/test), and one
fully-wired reference vertical slice (Account, end-to-end through every architectural layer) that
every subsequent epic's implementation copies the shape of. Epic 10 (Cross-Cutting Foundation —
structured logging, RFC 9457 errors, idempotency keys, pagination) is built alongside Epic 0,
since the HTTP server skeleton is where that middleware actually lives.

## 4. MVP Scope

**In scope:**

- User registration/login (email + password)
- Households: invite by email, admin approval to join
- Accounts (checking/savings/cash — `investment` type deferred, see Roadmap below)
- Credit Cards (separate entity from Account)
- Transactions (single category, optional household-share flag)
- Transfers (between any two sources: account↔account, account↔card, card↔card)
- Categories (hierarchical, income/expense typed; seeded per-user at registration from an internal
  default template, fully user-editable — name, color, icon, parent)
- Multi-currency support with daily cached FX rates (MVP supports `CAD`, `USD`, `BRL` only — see
  `internal/shared/currency`; more currencies can be added later without a schema change, since
  `currency` is stored as a plain string)
- Credit card statements (auto-generated on close day, immutable snapshot)
- Basic reports: spending by category, net worth, household shared spending, statement history

**Explicitly out of scope for MVP (candidate future phases):**

- Transaction splitting across multiple categories
- Credit limit enforcement
- Cached/denormalized balances (phase 2 refactor, deliberately deferred to learn cache invalidation
  as its own exercise)
- Recurring/scheduled transactions
- Budgets
- Bank sync (Plaid-like integrations)
- Report precomputation/materialized views
- Household Settlement — members set a split type (equal/percentage/income-based), and at the end
  of a period (since household creation or since the last settlement) the app calculates all
  shared transactions and determines who owes whom, ideally reduced to the minimum number of
  payments (a debt-simplification/graph problem — a strong phase 2/3 candidate once the MVP core
  is working end-to-end)
- Investment accounts (`Account.type = investment`) — appears in the domain model as a fourth
  account type, but deferred: real investment tracking (holdings, valuations, cost basis) is
  materially more complex than a simple balance and isn't worth the added scope for the MVP or the
  reference vertical slice (TALLY-137)

## 5. Domain Entities

### User

- `id`
- `email` (unique)
- `password_hash`
- `display_name` — shown in household contexts (e.g. shared transaction lists) instead of email
- `avatar_url` (nullable) — actual upload/storage pipeline deferred until the frontend is built
- `locale` (e.g. `en`, `pt-BR`) — default locale for the frontend; also used once, at registration,
  to translate the user's seeded default categories into their language (see Category below)
- `reporting_currency` — currency used for aggregate reports (e.g. net worth)
- `created_at`, `updated_at`

### Household

- `id`
- `name`
- `admin_user_id` — the user who created it / approves join requests
- `created_at`

### HouseholdMember

- `household_id`
- `user_id`
- `status` — `pending` | `approved`
- `joined_at`

Joining flow: a user is invited by email; the household admin must approve the join request before
the invitee becomes an approved member.

### Account

- `id`
- `user_id` — owner; accounts are never jointly owned
- `name`
- `type` — `checking` | `savings` | `cash` (`investment` is deferred — see Roadmap below)
- `currency`
- `initial_balance_minor_units`
- `color` (hex string, required)
- `icon` (string identifier, required — always derived server-side from `type`, not independently
  client-settable, e.g. `lucide:landmark` for `checking`)
- `created_at`, `updated_at`, `deleted_at` (soft delete)

### CreditCard

- `id`
- `user_id` — owner; never jointly owned
- `name`
- `currency`
- `credit_limit_minor_units` — reference only, not enforced in MVP
- `close_day` — day of month the billing cycle closes
- `due_day` — day of month payment is due
- `color` (hex string, required)
- `icon` (string identifier, required, client-supplied — e.g. `lucide:credit-card`; unlike
  Account, CreditCard has no `type` enum to derive it from)
- `created_at`, `updated_at`, `deleted_at` (soft delete)

### Statement

- `id`
- `credit_card_id`
- `cycle_start_date`, `cycle_end_date`
- `due_date`
- `total_amount_minor_units` — current, live total; mutable after close via logged adjustments
  (see StatementAdjustment below)
- `paid_amount_minor_units` (nullable) — snapshot of what was actually paid, captured the moment
  status flips to `paid`
- `status` — `open` | `closed` | `paid`
- `created_at`

Statements are auto-generated by a scheduled job when a card's close day is reached, and the
initial total is computed as a baseline. Transactions are assigned to a statement's cycle based on
**transaction date**, not posting date.

Because transaction entry is manual, a transaction dated inside an already-closed (or even
already-paid) cycle can still be entered later. Rather than blocking this or silently rewriting
history, the transaction attaches to that statement, the total is recalculated, and the change is
recorded in `StatementAdjustment` — so the statement's total can move after close, but every change
is traceable to the transaction that caused it. For a `paid` statement, comparing
`total_amount_minor_units` to `paid_amount_minor_units` tells the user at a glance whether they now
owe more or overpaid, rather than the app pretending the numbers still match or blocking the edit
outright.

### StatementAdjustment

- `id`
- `statement_id` (FK)
- `transaction_id` (FK) — the late transaction that triggered the recalculation
- `adjustment_type` — `late_addition` | `late_removal` | `correction`
- `amount_delta_minor_units` (signed)
- `created_at`

An append-only audit log of every change made to a statement's total after it was first generated.
Never edited or deleted, only appended to.

### Transaction

- `id`
- `source_type` — `account` | `credit_card`
- `source_id`
- `category_id`
- `amount_minor_units` (signed: positive = income/deposit, negative = expense/withdrawal)
- `currency`
- `fx_rate_to_reporting` — snapshot of the rate at time of recording
- `converted_amount_minor_units` — snapshot of the amount in the user's reporting currency
- `transaction_date`
- `description`
- `shared_to_household` (bool) — if true, visible to other approved household members
- `statement_id` (nullable) — set once a credit card statement closes over this transaction
- `created_at`, `updated_at`, `deleted_at` (soft delete)

A transaction always belongs to exactly one category (no splitting in MVP) and exactly one source
(an Account or a CreditCard, never both).

### Transfer

- `id`
- `from_type` / `from_id` — source of funds (account or credit card)
- `to_type` / `to_id` — destination (account or credit card)
- `amount_minor_units`
- `from_currency` / `to_currency`
- `fx_rate` (if cross-currency)
- `date`
- `description`
- `created_at`, `deleted_at` (soft delete)

Transfers model any two-sided money movement: account→account (e.g. checking→savings),
account→credit card (a card payment), or card→card (balance transfer). Transactions remain
strictly one-sided (spend/income only).

### Category

- `id`
- `key` — stable, non-translated identifier (e.g. `food.groceries`); the frontend maps this to its
  own localized label rather than displaying `name` directly, since the app is localized
- `name` — English reference/fallback label, not the source of truth for display
- `parent_category_id` (nullable) — enables hierarchy (e.g. Food > Groceries, Food > Dining Out)
- `type` — `income` | `expense`

System-provided defaults only in MVP; not user-customizable. Per-user color/icon customization
(via a separate `CategoryPreference` overlay, since Category rows are shared/global) is deferred
to a later phase — see Roadmap.

### FxRate

- `id`
- `currency_pair` (e.g. `USD_CAD`)
- `rate` — stored as fixed-point/decimal (not float), e.g. `numeric(20,10)`
- `date`

Populated daily via a scheduled job hitting a free FX API. Looked up by currency pair + date to
provide historical, reproducible conversions (a transaction's converted amount never shifts after
the fact, even if today's rate differs).

## 6. Key Business Rules

1. **Isolation & sharing:** Data is isolated per user by default. A transaction is visible to
   other household members only if it is explicitly flagged `shared_to_household`, and only to
   members with `approved` status in the same household. Accounts and credit cards themselves are
   never jointly owned — only individual transactions can be shared.
2. **Balances:** Computed on the fly for MVP — `balance = initial_balance + sum(transactions on
   this source) + sum(transfers in) - sum(transfers out)`. Cached/denormalized balances are a
   deliberate phase 2 exercise, not part of MVP.
3. **Money storage:** All amounts stored as integers in the currency's minor unit (e.g. cents),
   never floats. Currencies do not all share the same minor-unit exponent (JPY/KRW = 0, most = 2,
   BHD = 3) — exponent is looked up per currency, not assumed. Conversions apply a documented,
   consistent rounding rule (round-half-up to nearest minor unit).
4. **Supported currencies (MVP):** only `CAD`, `USD`, and `BRL` are accepted anywhere a currency
   is set (Account, CreditCard, Transaction, Transfer, User's `reporting_currency`) — enforced by
   the closed `currency.Code` enum in `internal/shared/currency`. All three use a 2-decimal minor
   unit, so this doesn't yet exercise the non-2-decimal case from Business Rule 3 above; adding a
   0- or 3-decimal currency (e.g. JPY, BHD) is the natural next test of that rule.
5. **Statements:** Auto-generated by a scheduled job on the card's close day, computing an initial
   baseline total. A late-entered transaction dated inside a closed or paid statement's cycle
   still attaches to that statement — the total recalculates and the change is logged in
   `StatementAdjustment` rather than being blocked or silently rewritten. Transaction date (not
   posting date) determines which cycle a transaction belongs to.
6. **Soft delete:** Account, CreditCard, Transaction, Transfer, and Category use soft delete
   (`deleted_at`) to preserve referential integrity for closed statements and historical reports.
   Hard delete is acceptable only for records with no downstream reference (e.g. an expired
   household invite).
7. **Reports:** Computed live from current data for MVP (no precomputation/materialized views) —
   spending by category (with parent/child rollup), net worth (all balances converted to the
   user's reporting currency), household shared spending, and per-card statement history.
8. **Categories:** Owned per-user, not shared/global. At registration, a default set is cloned
   from an internal seed template into the new user's own rows, translated into their `locale` at
   that moment. From then on the user can freely rename, recolor, re-icon, reparent, or delete any
   category, and create new ones — there is no separate "system category" to protect, and no
   override layer is needed. A transaction embedding a shared household member's category shows
   that member's own category name/color/icon directly, since the viewer may not own that category
   themselves.

## 7. Learning Objectives Mapped to Features

| Feature | System design / DS&A concept exercised |
| --- | --- |
| Household sharing + visibility rules | Access control modeling, authorization logic |
| Statement generation | Scheduled jobs, idempotency, immutability/snapshotting |
| Multi-currency + FX cache | Caching strategy, external API integration, historical data integrity |
| Transfers vs Transactions | Data modeling / normalization tradeoffs |
| Category hierarchy | Tree structures, recursive queries |
| Soft delete | Data lifecycle, referential integrity |
| Computed → cached balance (phase 2) | Cache invalidation, consistency, concurrency |
| Statement adjustments | Append-only audit logging, mutable-with-history vs. pure immutability |
| Reports | Query design, aggregation, (later) precomputation tradeoffs |

## Changelog

- **2026-08-03** — Migrated this doc from Notion ("Requirements & Domain Model (MVP)") into the
  repo, so it can't silently drift from CLAUDE.md or the code. The Notion page now points here.
- **2026-08-03** — `Account.color`/`Account.icon` changed from nullable to required; `icon` is now
  derived server-side from `type` rather than independently client-settable. *Reasoning:* an
  account with a missing icon/color is a display bug waiting to happen, and deriving icon from
  type keeps the two from ever drifting (TALLY-137 review).
- **2026-08-03** — `investment` removed from `Account.type` for MVP, added to §4 out-of-scope.
  *Reasoning:* real investment tracking (holdings, valuations, cost basis) is materially more
  complex than a simple balance and isn't worth the scope for TALLY-137's reference vertical slice
  or the MVP.
- **2026-08-03** — `CreditCard.color`/`CreditCard.icon` changed from nullable to required, mirroring
  the Account decision. *Reasoning:* consistency across the two visually-presented money-source
  entities; unlike Account, CreditCard has no `type` enum to derive an icon from, so both fields
  stay client-supplied but become mandatory input.
- **2026-08-03** — Documented `cmd/migrate` in §2 Stack. *Reasoning:* the binary already existed in
  code and is what `make migrate-up`/`migrate-down`/`migrate-verify` actually run, but it wasn't
  named anywhere in the spec — worth recording why it's a hand-written wrapper instead of
  `golang-migrate`'s own CLI (keeps the module's driver imports to just `pgx/v5`).
- **2026-08-03** — Documented that MVP only supports `CAD`, `USD`, `BRL` as currencies (§4, new
  Business Rule 4 in §6). *Reasoning:* `internal/shared/currency.Code` already enforced this as a
  closed enum, but no doc said so — anyone reading the domain model would have assumed the general
  "multi-currency" framing meant an open set.
