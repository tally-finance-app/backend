# Tally — Personal Finance Tracker

A multi-user, multi-currency personal finance tracker with household expense sharing and credit
card statement management, built as a hands-on system design and data structures/algorithms
learning project.

> **Why this exists:** I wanted a finance tracker that actually fit how I manage money (credit
> card statements, multiple currencies, sharing specific expenses with my household — not whole
> accounts) and couldn't find one that did all of it. Building it myself doubles as a deliberate
> exercise in backend system design: access control, scheduled jobs, historical-data integrity,
> audit trails, and the kind of edge cases (a card that closes on the 31st in a 30-day month)
> that don't show up until you actually design for them.

## Status

🚧 In active development — MVP backend first, Angular frontend to follow once the API is stable.
See [Roadmap](#documentation--planning) for what's built, what's next, and what's deliberately
deferred.

## Tech Stack

| Layer | Choice |
|---|---|
| Language | Go |
| Database | PostgreSQL |
| DB access | Raw SQL via [`sqlc`](https://sqlc.dev/) (typed codegen) + [`pgx`](https://github.com/jackc/pgx) — deliberately not an ORM |
| Migrations | [`golang-migrate`](https://github.com/golang-migrate/migrate) |
| HTTP router | [`chi`](https://github.com/go-chi/chi) |
| Frontend (future) | Angular + TypeScript |
| CI | GitHub Actions |

## Why These Choices

- **No ORM.** System design intuition — index usage, join cost, query shape — comes from seeing
  actual SQL. `sqlc` gives compile-time-safe, typed Go from hand-written queries without an ORM
  hiding what's really happening in Postgres.
- **Light DDD, not full DDD.** Entities with enforced invariants, repository interfaces owned by
  the domain, and a service layer separate from HTTP handlers — but no formal aggregates, domain
  events, or bounded-context ceremony. This is a solo project; that overhead mainly pays off
  coordinating teams.
- **Money as integers.** Every amount is stored as an integer in the currency's minor unit
  (cents), never a float — the standard fix for the classic floating-point rounding bug class in
  financial software.
- **Balances computed live, not cached — for now.** A deliberate MVP simplification; caching is
  planned as its own dedicated exercise later specifically to learn cache invalidation properly,
  rather than bolting it on prematurely.

## Architecture

```
internal/
  account/            model, repository interface, service
  creditcard/
  transaction/
  transfer/
  category/
  household/
  statement/
  fx/
  user/
  platform/postgres/   repository implementations (sqlc-generated)
  transport/http/       handlers, routing, middleware
cmd/
  api/                 HTTP server entrypoint
  jobs/                 scheduled jobs (statement generation, FX rate caching)
```

Each domain package follows the same shape: `model.go` (entity + constructor enforcing
invariants) → `repository.go` (interface only) → a Postgres implementation in
`internal/platform/postgres/` → `service.go` (business logic, depends only on the repository
interface) → an HTTP handler in `internal/transport/http/`. See `CLAUDE.md` for the full
architectural reference.

## Getting Started

### Prerequisites

- Go 1.22+
- Docker + Docker Compose
- [`sqlc`](https://docs.sqlc.dev/en/latest/overview/install.html)
- [`golang-migrate`](https://github.com/golang-migrate/migrate#installation)

### Setup

```bash
# Clone and enter the repo
git clone <this-repo-url>
cd tally

# Copy env template and fill in values
cp .env.example .env

# Start local Postgres
make db-up

# Apply migrations
make migrate-up

# Generate typed query code from SQL
make generate

# Run the API server
go run ./cmd/api
```

Server starts on the port configured in `.env`; confirm it's up with:

```bash
curl localhost:<port>/health
```

### Common commands

```bash
make db-up            # start local Postgres (Docker Compose)
make db-down           # stop and remove (including volume)
make migrate-up         # apply migrations
make migrate-down       # roll back one migration
make generate             # run sqlc generate
make build                 # go build ./...
make test                    # go test ./... (needs db-up first for integration tests)
make lint                     # golangci-lint run
```

## Testing

- **Service-layer unit tests** use a hand-written fake repository — no database required, fast.
- **Repository-layer integration tests** run against the real Dockerized Postgres, exercising
  actual SQL and constraints.

```bash
make db-up
make test
```

## API

REST API following [RFC 9457](https://www.rfc-editor.org/info/rfc9457) for error responses
(`application/problem+json`), consistent pagination across list endpoints, and
`Idempotency-Key`-protected money-moving endpoints (`POST /transactions`, `POST /transfers`).
Full contract in the [API Contract doc](#documentation--planning) below.

## Claude Code

This repo includes `CLAUDE.md` and a set of custom skills under `.claude/skills/` (code review,
test writing, SQL/migration review, Linear acceptance-criteria checking, commit/PR writing, and
domain-module scaffolding) for working with [Claude Code](https://claude.com/product/claude-code)
on this codebase.

## Documentation & Planning

Full specs, architecture decisions, and task tracking (private — linked here for my own
reference):

- [Requirements & Domain Model](https://app.notion.com/p/3ae70b529659816cafb5cef20c21f6ca)
- [ER Diagram](https://app.notion.com/p/3ae70b52965981ab8a56f1ab02e3beef)
- [API Contract](https://app.notion.com/p/3ae70b529659819ab060f13f57117bf8)
- [Product Roadmap](https://app.notion.com/p/3ae70b5296598192bc37e0b1656ec5bb)
- [Linear project (task tracking)](https://linear.app/dseagull/project/personal-finance-tracker-mvp-77c8fb089726)

## License

TBD.
