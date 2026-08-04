# Product Roadmap

> Migrated from Notion on 2026-08-03. CLAUDE.md §11 keeps a condensed "don't build this yet" list
> derived from this file — if a phase here changes, update §11 to match.

This roadmap sequences work after the MVP core (Auth/Households, Accounts/Cards,
Transactions/Transfers, Statements & FX, basic Reports) by learning value, not just feature
demand. Items are grouped into phases; within a phase, order is flexible.

## MVP baseline additions (cheap now, expensive later)

- **Structured logging** — Go's `log/slog`, with a request/job correlation ID attached to every
  log line. Unlike metrics/tracing/job queues (Phase 7), this isn't a scale concern — it's about
  being able to answer "what happened and why" from day one, especially for unattended background
  jobs (statement generation, FX rate fetch) where nobody's watching a terminal when things break.

## Phase 2 — Consistency & Scheduled Work

- Cached/denormalized balances (deliberate cache-invalidation exercise, contrasted with the
  computed-on-the-fly MVP approach)
- Statement/due-date reminders (notifications) — natural extension of the scheduled-job
  infrastructure already built for statement generation

## Phase 3 — Collaboration

- **Household Settlement** — members set a split type (equal/percentage/income-based); at the end
  of a period (since household creation or since the last settlement), the app calculates all
  shared transactions and determines who owes whom — ideally reduced to the minimum number of
  payments (a debt-simplification/graph problem)

## Phase 4 — Personalization & Control

- Transaction splitting across multiple categories
- Credit limit enforcement
- Budgets, with threshold alerts

## Phase 5 — Automation

- Recurring/scheduled transactions (rent, subscriptions)
- Bank sync (Plaid-like integrations) — aspirational; a significant regulatory/security
  undertaking, more "if you want to go there" than a firm commitment

## Phase 6 — Insight & Forecasting

- Spending trend analysis / forecasting (time-series exercise)
- Anomaly detection (unusual spending flags)
- Report precomputation/materialized views — revisit only once real data volume justifies it,
  rather than building it preemptively

## Phase 7 — Scale & Reliability

- Metrics and distributed tracing (a full observability stack, beyond the baseline structured
  logging already in place from the MVP)
- Background job queue (replacing simple cron with a real message broker) — a good showcase for
  Go's concurrency primitives
- Read replicas / sharding — mostly a design/discussion exercise; unlikely to be genuinely needed
  at personal-app scale, but worth reasoning through deliberately

## Frontend track (parallel, not sequential)

- Angular UI — can start any time after the MVP API is stable enough to consume; doesn't need to
  wait for later backend phases

## Changelog

- **2026-08-03** — Migrated this doc from Notion ("Product Roadmap") into the repo; no content
  changes. The Notion page now points here.
