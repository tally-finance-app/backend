# Tally — Personal Finance Tracker

Self-contained project reference for Claude Code. This file is meant to make Notion/Linear lookups unnecessary for day-to-day work — if something here goes stale, fix it here first.

Source of truth: this file is self-contained by design (Claude Code shouldn't need to fetch Notion mid-session). If this file and the Notion Requirements & Domain Model doc ever disagree, Notion is canonical — update this file to match, then note the fix in your next commit. Check for drift periodically, not just when something visibly breaks.

## 1. Purpose & Goals

A multi-user, multi-currency personal finance tracker with household expense sharing and credit
card statement management. Built as a **deliberate learning project** for system design and
data structures/algorithms — prefer the more instructive design over the fastest one to ship,
within reason. A genuinely usable app is a secondary outcome, not the driver of scope decisions.

## 2. Stack

- **Language:** Go (backend)
- **Database:** Postgres, accessed via raw SQL + `sqlc` (typed codegen) + `pgx` driver.
  Deliberately not an ORM — index usage, join cost, and query shape should stay visible.
- **Migrations:** `golang-migrate`
- **HTTP router:** `chi`
- **Frontend (future phase, not yet started):** Angular + TypeScript, consuming the REST API.
- **Git host:** GitHub. CI is GitHub Actions.
- **Task tracking:** Linear (project "Go Backend", team "Tally Finance App").
  Specs live in Notion; this file duplicates what matters day-to-day so Claude Code doesn't need
  to fetch either.

## 3. Architecture — Light DDD

Tactical DDD patterns without full ceremony. Solo project — the overhead of full DDD (formal
aggregates, domain events, CQRS, bounded-context mapping) mainly pays off coordinating teams, not
solo learning. What we DO use:

- **Entities with enforced invariants** — constructor functions, not raw public structs anyone
  can build in an invalid state.
- **Repository interfaces owned by the domain package**, implemented separately in
  `internal/platform/postgres`. This is dependency inversion — business logic depends on an
  interface, never on `sqlc`-generated code or `pgx` directly.
- **A service/use-case layer** separate from HTTP handlers — handlers do request
  parsing/response writing only; all business logic lives in the service layer, which is
  unit-testable with a fake repository (no real DB needed for service-layer tests).

### Package layout

```
internal/
  account/            model.go, repository.go [interface], service.go
  creditcard/
  transaction/
  transfer/
  category/
  household/
  statement/
  fx/
  user/
  platform/postgres/  repository implementations, sqlc-generated code
  transport/http/      handlers, routing, middleware
cmd/
  api/                main.go — HTTP server
  jobs/                main.go — statement generation, FX caching
```

**Every domain package should look like this** (the "reference vertical slice," built first as
Epic 0):

```
internal/<domain>/
  model.go       — entity + constructor enforcing invariants
  repository.go  — interface only (Create, GetByID, List, Update, SoftDelete, etc.)
  service.go     — use-case logic, depends only on the repository interface
```

Postgres implementation goes in `internal/platform/postgres/<domain>_repository.go`.
HTTP handlers go in `internal/transport/http/<domain>_handler.go`.

## 4. Money & Currency Rules

- **All amounts are integers in the currency's minor unit** (cents), stored as `bigint`, never a
  float. Field names always end in `_minor_units`.
- **Not all currencies use 2 decimal places** — JPY/KRW use 0, some currencies use 3. Look up the
  minor-unit exponent per currency; never assume ×100.
- **FX rates** are stored as `numeric(20,10)`, never float.
- **Rounding rule:** round-half-up to the nearest minor unit, applied consistently everywhere a
  conversion happens. Don't reinvent this per call site — use the shared conversion helper.
- **FX snapshotting:** a transaction/transfer's `fx_rate_to_reporting` and
  `converted_amount_minor_units` are computed ONCE at creation time, from the cached `FxRate`
  table for that exact date — never recalculated later even if the cache updates. This is what
  keeps historical reports stable.

## 5. Key Business Rules (the non-obvious ones)

1. **Household sharing is opt-in per transaction, not per account.** Accounts/cards are never
   jointly owned. A transaction is visible to other household members only if
   `shared_to_household = true` AND the viewer is an `approved` member of that household.
2. **Balances are computed live**, not cached, in the MVP. `balance = initial_balance +
sum(transactions) + sum(transfers in) - sum(transfers out)`. Caching is a deliberate Phase 2
   exercise — don't pre-build it now.
3. **Categories are fully user-owned**, not shared/global. Each user gets a cloned, translated
   set at registration (from an internal seed template) and can freely edit everything about any
   of them, including seeded ones. There is no "system category" concept anywhere in the code.
4. **Statement totals are mutable, but every change after generation is audited.** A late
   transaction (dated inside an already-closed or even already-paid statement's cycle) still
   attaches to that statement, recalculates its total, and logs a `StatementAdjustment` row
   (append-only). This applies even to `paid` statements — `paid_amount_minor_units` is a frozen
   snapshot of what was actually paid; `total_amount_minor_units` keeps moving with reality.
5. **Statement cycle assignment uses `transaction_date`, never posting date.**
6. **Soft delete** (`deleted_at`) on Account, CreditCard, Transaction, Transfer, Category — never
   hard-delete anything with potential downstream references. A delete that would break
   referential integrity (e.g. an account with transactions) returns `409 CONFLICT`, not a
   cascading delete.
7. **Idempotency keys are required** on `POST /transactions` and `POST /transfers` — these are
   the two endpoints where a duplicate write genuinely corrupts data. Every other endpoint does
   not require one.

## 6. API Conventions

- **Errors:** RFC 9457 Problem Details (`Content-Type: application/problem+json`) everywhere.
  `type` is the stable, machine-readable identifier (`validation-error`, `not-found`,
  `unauthorized`, `forbidden`, `conflict`) — never match on `title`/`detail` text.
- **Pagination:** every list endpoint takes `?page=1&page_size=50` (max 200), responds with
  `{ "data": [...], "page", "page_size", "total" }`.
- **Status codes:** `200` GET/PATCH, `201` POST-creates, `204` DELETE/logout.
- **Auth:** Bearer token in `Authorization` header. Passwords hashed with bcrypt/argon2 — never
  hand-rolled.

## 7. Logging

Structured logging via `log/slog` only — no `fmt.Println`/`log.Println` in request or job code.
Every HTTP request and every scheduled job run gets its own correlation ID, present on every log
line for that operation.

## 8. Testing Conventions

- **Service-layer unit tests**: use a hand-written fake implementing the repository interface —
  no real database. This is what proves the dependency-inversion pattern actually decouples
  business logic from Postgres.
- **Repository-layer integration tests**: run against the real Dockerized Postgres — these are
  the tests that actually exercise SQL, constraints, and Postgres-specific behavior. They must
  `t.Skip` (never `t.Fatal`) when `DATABASE_URL` is unset or `-short` is passed, so `make test`
  stays green on a machine with no Postgres. Run them with `make test-integration`.
- **Table-driven tests** are the default style for anything with multiple input/output cases.
- Every Linear ticket's "Definition of Done" names a specific test scenario — write that test,
  don't just write _a_ test.

## 9. Commands

```
make db-up            # start local Postgres (Docker Compose)
make db-down           # stop and remove (including volume)
make migrate-up         # apply migrations
make migrate-down       # roll back one migration
make migrate-verify      # up -> down -> up, proving the down migrations work
make generate             # regenerate sqlc code (after editing a queries.sql)
make verify-generate       # fail if committed generated code is stale
make build                  # go build ./...
make vet                     # go vet ./...
make lint                     # golangci-lint run
make test                      # unit tests only, no database needed
make test-integration           # everything, incl. real Postgres (needs db-up first)
make ci                          # what CI runs, end to end
make tools                        # build the pinned dev tools into ./bin
make hooks                         # install the lefthook git hooks (once per clone)
```

**Development tools live in a separate `tools/` module.** `sqlc`, `golangci-lint`, and `lefthook`
are pinned in `tools/go.mod` and built into `./bin` (gitignored) by `make tools`; every target that
needs one declares it as a prerequisite, so they build on demand. Never `go install` them, and never
add them to the root `go.mod`.

The split is not cosmetic: minimal version selection unifies the whole module graph, so when sqlc
was a root tool dependency **its** requirement silently chose the application's `pgx` version. A
separate module makes that impossible. For the same reason, do **not** add `tools/` to a `go.work`
workspace — workspaces re-unify MVS across modules and would undo it.

Tools are invoked as `./bin/<tool>`, not `go tool <tool>` — `go -C tools tool sqlc ...` would run
with `tools/` as the working directory, where `sqlc.yaml` and `migrations/` don't exist.

Git hooks are managed by `lefthook.yml`: `pre-commit` formats and lints staged Go files and checks
sqlc freshness, and `commit-msg` enforces §10 (blocks a `Co-Authored-By` trailer, warns on a missing
`TALLY-<n>` reference). There is deliberately **no** `pre-push` hook — `main` is branch-protected,
so build/vet/test are covered by CI on the PR.

**sqlc-generated code is committed**, not gitignored: a fresh clone must build without sqlc
installed, and a migration that silently changes a Go type should be visible in code review.
`make verify-generate` (`sqlc diff`) is the CI guard that keeps it current. Never hand-edit
anything under `internal/platform/postgres/generated/`, and never put a hand-written file there.

## 10. Git & Linear Workflow

- Branch naming follows Linear's convention: `<username>/tally-<issue-number>-<slug>` (Linear
  generates this automatically per issue — copy it from the issue rather than inventing your own).
- Commit messages and PR descriptions should reference the Linear ticket ID (e.g. `TALLY-109`).
- Every Linear story follows this template — match it when writing PR descriptions or checking
  work against a ticket: **Background, User story, Scope, Acceptance criteria, Out of scope,
  Dependencies, Definition of done.**
- Commits should **not** include a `Co-Authored-By` trailer for Claude.

## 11. Roadmap Awareness

Things deliberately **NOT** in this MVP — don't build them "while you're in there," even if it
seems convenient:

- Cached/denormalized balances (Phase 2 — deliberate cache-invalidation learning exercise)
- Household Settlement / expense splitting (Phase 3)
- User-customizable... wait, categories already are (see §5.3) — but transaction splitting,
  credit limit enforcement, budgets (Phase 4)
- Recurring transactions, bank sync (Phase 5)
- Spending forecasting, anomaly detection, report precomputation (Phase 6)
- Full observability (metrics/tracing), background job queue, read replicas (Phase 7)

## 12. Available Skills

See `.claude/skills/` — code-review, test-writing, linear-ac-check, sql-migration-review,
commit-pr-writer, endpoint-scaffold. Each has its own SKILL.md with detailed instructions.
