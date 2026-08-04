# API Contract (MVP)

> Canonical REST API spec. Migrated from Notion on 2026-08-03 — see the Changelog at the bottom
> for what's changed since, and CLAUDE.md §6 for the condensed, day-to-day version of these
> conventions.

## Conventions

- **Base path:** `/api/v1`
- **Auth:** Bearer token (session token issued at login) in `Authorization: Bearer <token>`
  header. All endpoints except `/auth/register` and `/auth/login` require it.
- **IDs:** UUIDv4 strings.
- **Money:** every amount field is an integer in the currency's minor unit (e.g. cents). Never a
  float. Field names always end in `_minor_units`.
- **Currencies:** MVP only accepts `CAD`, `USD`, `BRL` in any `currency`/`from_currency`/
  `to_currency`/`reporting_currency` field — a closed allowlist, not open-ended multi-currency
  support yet (see Requirements & Domain Model §6.4). Anything else is a `400 validation_error`.
- **Dates:** `transaction_date`, `cycle_start_date`, etc. are `YYYY-MM-DD`. Timestamps
  (`created_at`, etc.) are ISO 8601 UTC (`2026-07-31T14:00:00Z`).
- **Pagination:** list endpoints accept `?page=1&page_size=50`; `page_size` must be one of `10,
  25, 50, 100, 200` (a fixed allowlist, not just a range check — any other value is a `400
  validation_error`), default `50`. Responds with
  `{ "data": [...], "page": 1, "page_size": 50, "total": 137 }`
- **Ordering:** every list endpoint returns a stable, deterministic order — default `created_at
  ASC` with `id` as a secondary sort key, so rows sharing a timestamp still paginate
  deterministically across pages.
- **Errors:** follow [RFC 9457 (Problem Details for HTTP APIs)](https://www.rfc-editor.org/info/rfc9457).
  Responses use `Content-Type: application/problem+json`:

```json
{
  "type": "https://docs.financetracker.dev/problems/validation_error",
  "title": "Bad Request",
  "status": 400,
  "detail": "amount_minor_units must be non-zero",
  "instance": "/api/v1/transactions",
  "errors": [ { "field": "amount_minor_units", "detail": "must be non-zero" } ]
}
```

  - `type` — stable, machine-readable identifier for the problem category (clients match on this,
    not on `title`). One per category: `validation_error`, `not_found`, `unauthorized`,
    `forbidden`, `conflict`.
  - `title` — fixed human-readable summary for that type.
  - `detail` — specific explanation for this occurrence.
  - `instance` — the request path; pairs naturally with the request correlation ID used in
    structured logging.
  - `errors` — our own extension member, used for multi-field validation errors.
- **Soft delete:** `DELETE` endpoints on soft-deletable entities set `deleted_at`; the record
  disappears from default list/get results but is not physically removed.
- **Success status codes:** `200` for `GET`/`PATCH`, `201` for a `POST` that creates a resource,
  `204 No Content` for `DELETE` and logout (no response body).
- **Idempotency:** money-moving `POST` endpoints (`POST /transactions`, `POST /transfers`) require
  an `Idempotency-Key` request header — a client-generated UUID per logical operation. The server
  stores the completed response keyed by it for 24h; a retried request with the same key returns
  the original response instead of creating a duplicate record. Missing header on these endpoints
  → `400` (`missing_idempotency_key` problem type).
- **Ownership & sharing:** unless noted, a resource is only visible to its owning user. Where a
  household share applies, this is called out explicitly per endpoint.

## Auth

### `POST /auth/register` — `201`

Creates a new user.

```json
{ "email": "alex@example.com", "password": "•••", "display_name": "Alex", "reporting_currency": "CAD", "locale": "en" }
```

### `POST /auth/login` — `200`

```json
{ "email": "alex@example.com", "password": "•••" }
```

Response: `{ "token": "opaque-session-token", "expires_at": "..." }`

### `POST /auth/logout` — `204`

Invalidates the current session token.

## Users

### `GET /users/me` — `200`

Returns the authenticated user's profile.

### `PATCH /users/me` — `200`

Updates `display_name`, `avatar_key`, `locale`, and/or `reporting_currency`.

## Households

### `POST /households` — `201`

Creates a household; caller becomes `admin_user_id` and an auto-approved member.

### `GET /households/:id` — `200`

Returns household details + member list (only visible to approved members).

### `POST /households/:id/invitations` — `201`

Admin-only. Invites a user by email. Creates a `HouseholdMember` row with `status = pending`.

### `GET /households/:id/members` — `200`

Lists members and their status.

### `PATCH /households/:id/members/:userId` — `200`

Admin-only. Updates a member's status — used to approve a pending join request.
`{ "status": "approved" }`

### `DELETE /households/:id/members/:userId` — `204`

Admin can remove a member; a member can remove themself (leave).

## Accounts

### `POST /accounts` — `201`

```json
{ "name": "Checking", "type": "checking", "currency": "CAD", "initial_balance_minor_units": 150000, "color": "#2563EB" }
```

`color` is required. `icon` is not client-settable — it's always derived server-side from `type`
(see Requirements & Domain Model).

### `GET /accounts` — `200`

Lists the caller's own accounts (never household-shared at the account level).

### `GET /accounts/:id` — `200`

### `PATCH /accounts/:id` — `200`

Updatable: `name`, `type`, `color`. `icon` is not independently settable — changing `type` updates
`icon` to match automatically. `currency` and `initial_balance_minor_units` are immutable after
creation (changing them would silently invalidate historical balance calculations).

### `DELETE /accounts/:id` — `204`

Soft delete. Fails with `409 CONFLICT` if the account has transactions/transfers referencing it —
archive, don't delete, is the intended flow for accounts with history.

### `GET /accounts/:id/balance` — `200`

```json
{ "balance_minor_units": 152340, "currency": "CAD", "as_of": "2026-07-31T14:00:00Z" }
```

Computed live per Business Rule 2 (MVP has no cached balance).

## Credit Cards

### `POST /credit-cards` — `201`

```json
{ "name": "Visa Rewards", "currency": "CAD", "credit_limit_minor_units": 500000, "close_day": 15, "due_day": 5, "color": "#7C3AED", "icon": "lucide:credit-card" }
```

`color` and `icon` are both required.

### `GET /credit-cards` — `200`

### `GET /credit-cards/:id` — `200`

### `PATCH /credit-cards/:id` — `200`

Updatable: `name`, `credit_limit_minor_units`, `close_day`, `due_day`, `color`, `icon`. Currency
is immutable.

### `DELETE /credit-cards/:id` — `204`

Soft delete, same referential-integrity guard as Accounts.

### `GET /credit-cards/:id/balance` — `200`

Same shape as the Account balance endpoint — computed live.

## Categories

Categories are owned per-user (not shared/global). At registration, a default set is cloned from
an internal seed template into the new user's own rows, translated into their `locale` at that
moment — this happens once, server-side, and isn't part of this API surface.

### `POST /categories` — `201`

```json
{ "name": "Hobbies", "type": "expense", "parent_category_id": null, "color": "#059669", "icon": "lucide:palette" }
```

### `GET /categories` — `200`

Lists the caller's own categories. Pass `?tree=true` to receive nested children inline instead of
a flat list with `parent_category_id`.

### `GET /categories/:id` — `200`

### `PATCH /categories/:id` — `200`

Any field is editable — `name`, `type`, `parent_category_id`, `color`, `icon` — since categories
are fully user-owned. `parent_category_id` must reference a category owned by the same user.

### `DELETE /categories/:id` — `204`

Soft delete. Fails with `409 CONFLICT` if the category has transactions referencing it.

## Transactions

### `POST /transactions` — `201`

Requires an `Idempotency-Key` header (see Conventions).

```json
{
  "source_type": "credit_card",
  "source_id": "uuid",
  "category_id": "uuid",
  "amount_minor_units": -4599,
  "currency": "CAD",
  "transaction_date": "2026-07-28",
  "description": "Groceries",
  "shared_to_household": false
}
```

Server computes `fx_rate_to_reporting` and `converted_amount_minor_units` from the cached
`FxRate` table (looked up by `currency` → user's `reporting_currency` for that date). If the
transaction's date falls inside an already-closed or paid Statement's cycle (credit card source
only), it's still accepted: the transaction attaches to that statement, the statement
recalculates, and a `StatementAdjustment` is written (see Statements section).

### `GET /transactions` — `200`

Query params: `source_id`, `category_id`, `date_from`, `date_to`, `household_id` (to include
household members' shared transactions alongside your own).

> When `household_id` is passed and the caller is an approved member, results include this
> household's transactions where `shared_to_household = true`, from all approved members — not
> just the caller's own.

### `GET /transactions/:id` — `200`

### `PATCH /transactions/:id` — `200`

Any field except `source_type`/`source_id` is editable. Editing `amount_minor_units` or
`transaction_date` on a transaction already attached to a statement triggers the same
adjustment-logging behavior as a late addition.

### `DELETE /transactions/:id` — `204`

Soft delete. If attached to a statement, logs a `late_removal` adjustment.

## Transfers

### `POST /transfers` — `201`

Requires an `Idempotency-Key` header (see Conventions).

```json
{
  "from_type": "account", "from_id": "uuid",
  "to_type": "credit_card", "to_id": "uuid",
  "amount_minor_units": 20000,
  "from_currency": "CAD", "to_currency": "CAD",
  "date": "2026-07-30",
  "description": "Card payment"
}
```

If `from_currency != to_currency`, `fx_rate` is required (or server-resolved from the FxRate
cache) and the destination amount is converted accordingly.

### `GET /transfers` — `200`

Query params: `account_id` / `credit_card_id` (matches either side), `date_from`, `date_to`.

### `GET /transfers/:id` — `200`

### `DELETE /transfers/:id` — `204`

Soft delete.

## Statements

Statements are generated by a scheduled job, not created via API. The API surfaces them
read-only, plus the payment action.

### `GET /credit-cards/:id/statements` — `200`

Lists statements for a card, most recent first.

### `GET /statements/:id` — `200`

```json
{
  "id": "uuid",
  "credit_card_id": "uuid",
  "cycle_start_date": "2026-06-16",
  "cycle_end_date": "2026-07-15",
  "due_date": "2026-08-05",
  "total_amount_minor_units": 128400,
  "paid_amount_minor_units": null,
  "status": "closed",
  "created_at": "..."
}
```

### `GET /statements/:id/transactions` — `200`

Lists the transactions currently attached to this statement.

### `GET /statements/:id/adjustments` — `200`

Lists the `StatementAdjustment` audit trail for this statement (empty array if never amended).

### `POST /statements/:id/mark-paid` — `200`

```json
{ "paid_amount_minor_units": 128400 }
```

Flips status to `paid` and snapshots `paid_amount_minor_units`. Does not lock the statement
against future adjustments — late transactions can still attach later (see Business Rules). Kept
as an explicit action rather than a `PATCH` since marking paid is a meaningful business event
(distinct from the plain field-level status change used for household member approval).

## Reports

All report endpoints are computed live (no precomputation in MVP) and return amounts converted to
the caller's `reporting_currency`.

### `GET /reports/net-worth` — `200`

Sum of all account balances minus all credit card balances (owed), converted.

### `GET /reports/spending-by-category` — `200`

Query params: `date_from`, `date_to`, `household_id` (optional, to include shared spending).
Returns totals per category with child rollups.

### `GET /reports/household-shared-spending` — `200`

Query params: `household_id`, `date_from`, `date_to`. Breaks down shared-transaction totals per
member (visible only to approved members of that household).

### `GET /reports/statement-history` — `200`

Query params: `credit_card_id`. Returns statement totals over time — useful for a simple trend
view before any charting is built on the frontend.

## Out of scope for this contract (handled internally, not exposed)

- `FxRate` cache population — internal scheduled job hitting the external FX API; not user-facing
  in MVP.
- Statement generation — internal scheduled job triggered by `close_day`; no manual "generate"
  endpoint in MVP (may add a dev-only trigger later for local testing).

## Changelog

- **2026-08-03** — Migrated this doc from Notion ("API Contract (MVP)") into the repo, so it can't
  silently drift from CLAUDE.md or the code. The Notion page now points here.
- **2026-08-03** — Documented the fixed pagination `page_size` allowlist (`10/25/50/100/200`) and
  added the Ordering convention. *Reasoning:* both are now actually enforced by
  `internal/shared/pagination` and the account list query (TALLY-137/138); the old "max 200"
  wording undersold what the API actually validates.
- **2026-08-03** — `POST /accounts`: removed `icon` from the request body; `PATCH /accounts/:id`:
  removed `icon` from independently-updatable fields. *Reasoning:* matches the domain model change
  making `icon` server-derived from `type` (see Requirements & Domain Model changelog).
- **2026-08-03** — `POST /credit-cards`: noted `color` and `icon` are both required. *Reasoning:*
  matches the schema change making both columns `NOT NULL`.
- **2026-08-03** — Problem `type` values changed from hyphenated (`validation-error`, `not-found`,
  `missing-idempotency-key`) to underscored (`validation_error`, `not_found`,
  `missing_idempotency_key`). *Reasoning:* the doc was ahead of the code — `internal/apperr` and
  `internal/transport/http/errors.go` already emit underscored values, and changing the running
  code's wire format is a bigger, riskier move than fixing the spec to match it (this was flagged
  as a known drift in TALLY-138 and left unresolved there).
- **2026-08-03** — Added a Currencies convention noting the MVP-only `CAD`/`USD`/`BRL` allowlist.
  *Reasoning:* matches `internal/shared/currency.Code`; nothing previously said currency was a
  closed set rather than open multi-currency support.
- **2026-08-03** — Validation errors documented as `400` instead of `422` (pagination allowlist,
  currency allowlist, and the RFC 9457 example all updated; example `title` corrected to
  `"Bad Request"` to match `http.StatusText(400)`, which is what `internal/transport/http/errors.go`
  actually returns via `WriteError`). *Reasoning:* `400` is what the code has always sent for
  `apperr.KindValidation`, and `400` is a defensible RFC 9457/HTTP semantics choice on its own —
  not worth changing the running behavior just to match the spec. Note this leaves TALLY-137's
  written acceptance criteria ("invalid input → `422`") inconsistent with this doc; the ticket
  should be corrected to say `400` rather than treated as still requiring `422`.
