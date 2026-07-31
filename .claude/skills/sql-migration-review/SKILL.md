---
name: sql-migration-review
description: Use this skill when the user asks to review a database migration, review a sqlc query file, or check SQL before running it against the project's Postgres database. Covers migration safety (up/down correctness, data types matching the ER Diagram) and query correctness (parameter binding, index usage, N+1 risk).
---

# SQL & Migration Review

This project uses `golang-migrate` for schema migrations and `sqlc` for typed query generation —
deliberately not an ORM, specifically so SQL stays visible and reviewable (CLAUDE.md §2). This
skill is where that visibility actually gets used.

## Migrations (`golang-migrate`)

- **Up/down pairs**: every `up` migration has a corresponding `down` that actually reverses it —
  flag a migration with no down file, or a down file that doesn't fully undo the up (e.g. drops a
  table but doesn't restore a dropped column's data type on rollback).
- **Data types match CLAUDE.md §4**: money fields are `bigint` (minor units), never `numeric` or
  `float` for amounts; FX rates are `numeric(20,10)`. Flag any deviation.
- **Soft-delete columns**: Account, CreditCard, Transaction, Transfer, Category should each have
  a nullable, indexed `deleted_at timestamp`. Flag a new table in this set that's missing it.
- **Foreign keys and constraints**: check that foreign keys exist where the ER Diagram implies a
  relationship, and that unique constraints exist where the domain requires uniqueness (e.g.
  `User.email`).
- **Indexes**: flag missing indexes on columns that will obviously be queried/filtered on
  frequently — `Transaction.transaction_date`, `Transaction.source_id`,
  `Transaction.category_id`, `FxRate(currency_pair, date)`, `Statement.credit_card_id`. Also flag
  indexes that are clearly redundant (e.g. indexing a column that's already covered by a
  composite index's leading column).
- **Destructive migrations**: a migration that drops a column or table containing data should be
  flagged for explicit confirmation before running — even in a personal/dev-only project, this is
  the right habit to build.
- **Idempotency of migration numbering**: confirm the migration doesn't reuse a version number
  already used, which would break `golang-migrate`'s ordering.

## sqlc queries

- **Named parameters, not string concatenation**: every dynamic value must be a `sqlc.arg()` or
  positional parameter — flag any string-built SQL that could allow injection, even though sqlc
  makes this rare by design.
- **N+1 risk**: flag a service-layer pattern that calls a single-row query in a loop where a
  single batched/joined query would work instead (e.g. fetching each transaction's category
  individually instead of joining).
- **Ownership filtering in the query itself**: for any query fetching a user-scoped resource
  (Account, CreditCard, Transaction, Category, Transfer), confirm the `WHERE` clause filters by
  the owning `user_id` (or household membership, for shared-transaction queries) at the SQL
  level — don't rely on filtering happening only in application code after fetching, since that's
  both slower and a security footgun if a check is ever accidentally skipped.
- **Soft-delete filtering**: every `SELECT` that isn't explicitly meant to include soft-deleted
  rows should filter `WHERE deleted_at IS NULL` — flag queries that forgot this.
- **Minor-units and numeric handling**: confirm query parameter types match the Go types used for
  money (`int64` for minor units, not `float64`), and that `sqlc`'s generated types for
  `numeric` columns (FX rates) aren't being silently coerced to `float64` in application code.

## Output format

```
## Migrations
- [flag/ok] file — issue or confirmation

## Queries
- [flag/ok] file:query name — issue or confirmation

## Summary
[Ready to apply/merge, or specific blockers]
```
