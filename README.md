# Tally — Personal Finance Tracker

[![CI](https://github.com/tally-finance-app/backend/actions/workflows/ci.yml/badge.svg)](https://github.com/tally-finance-app/backend/actions/workflows/ci.yml)

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

- Go 1.26+ (see `go` directive in `go.mod`)
- Docker + Docker Compose (or Podman)

That's it. `sqlc`, `golangci-lint`, and `lefthook` are pinned in a separate `tools/` module and
built into `./bin` by `make tools` (which every target needing them depends on), so CI and your
machine run identical versions with nothing to install. `golang-migrate` is used as a library by
`cmd/migrate`, so its CLI isn't needed either.

Tools are deliberately kept in their own module so their dependency graphs can't influence the
versions the application builds against — see `tools/go.mod` for the details.

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

# Enable git hooks (format, lint, and codegen freshness on commit)
make hooks

# Run the API server
go run ./cmd/api
```

sqlc-generated query code is committed, so no codegen step is needed for a first run — you only
run `make generate` after editing a `queries.sql` file.

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
make migrate-verify     # up -> down -> up, proving the down migrations work
make generate            # regenerate sqlc code (after editing a queries.sql)
make verify-generate      # fail if committed generated code is stale
make build                 # go build ./...
make vet                    # go vet ./...
make lint                    # golangci-lint run
make test                     # unit tests only, no database needed
make test-integration          # everything, incl. real Postgres (needs db-up first)
make ci                         # what CI runs, end to end
make tools                        # build the pinned dev tools into ./bin
make hooks                         # install the lefthook git hooks
```

### Git hooks

[lefthook](https://github.com/evilmartians/lefthook) (pinned in `tools/go.mod`, config in
`lefthook.yml`) runs on commit once you've run `make hooks`:

| Hook | Does |
|---|---|
| `pre-commit` | formats staged Go files and re-stages them, lints the new changes, verifies committed sqlc output isn't stale |
| `commit-msg` | blocks a `Co-Authored-By` trailer; warns if no `TALLY-<n>` reference |

There is deliberately no `pre-push` hook: `main` is branch-protected, so every change arrives via a
PR and CI is the gate for build/vet/test. Bypass a hook with `git commit --no-verify`.

## Testing

- **Service-layer unit tests** use a hand-written fake repository — no database required, fast.
- **Repository-layer integration tests** run against the real Dockerized Postgres, exercising
  actual SQL and constraints.

Integration tests skip automatically when `DATABASE_URL` is unset, so `make test` is always safe
to run without Postgres.

```bash
make test                 # fast, no database

make db-up && make migrate-up
make test-integration      # includes repository-layer tests
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
