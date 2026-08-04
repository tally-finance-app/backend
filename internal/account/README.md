# Account — reference vertical slice

Account is the template every other domain module (`creditcard`, `transaction`, `transfer`,
`category`, `household`, `statement`, `fx`) copies. If you're starting a new epic, copy this
shape rather than inventing a new one — see `.claude/skills/endpoint-scaffold` for the automated
version of "copy this."

## The layers, in dependency order

```
model.go        Account entity + NewAccount() constructor enforcing invariants
                 (non-empty name, valid type/currency, icon derived from type — never
                 independently settable, so it can't drift from type).

repository.go    Repository interface ONLY: Create, GetByID, List, Count, Update,
                 SoftDelete. Owned by this package, not by postgres. Nothing in here
                 imports sqlc or pgx.

service.go       Service — use-case logic (CreateAccount, GetAccount, ListAccounts).
                 Depends only on the Repository interface above. This is what makes
                 service_test.go possible without a database: tests construct
                 Service with a hand-written fake Repository instead of the real
                 Postgres one.
```

Then, outside this package:

```
internal/platform/postgres/account_repository.go
                 sqlc-backed implementation of account.Repository. Only this file
                 (plus account_convert.go, account_queries.sql) knows sqlc/pgx exist.

internal/transport/http/account_handler.go
                 AccountHandler — parses the HTTP request, calls the Service, writes
                 the HTTP response. No business logic here: validation and
                 persistence both happen in service.go/model.go, not the handler.

cmd/api/main.go  Wires it together: pgxpool -> postgres.NewAccountRepository ->
                 account.NewService -> tallyhttp.NewAccountHandler -> NewRouter.
```

## Why this shape

Each arrow is a dependency-inversion point: `service.go` never imports
`internal/platform/postgres`, and `account_handler.go` never imports it either — both depend on
the `Repository` interface declared here. That's what lets `service_test.go` unit-test business
logic with zero database, while `internal/platform/postgres/account_queries_test.go` separately
proves the real SQL works, against real Postgres. See CLAUDE.md §8 for the testing conventions
this maps to, and §3 for the layering rationale.

## What this slice deliberately does NOT cover

Per TALLY-137 (reference vertical slice) vs. Epic 3 (the full Accounts feature): this only proves
`POST /accounts`, `GET /accounts`, `GET /accounts/:id`. Update, soft-delete, and balance
calculation already have repository/model support (see `repository.go`, `Update`/`SoftDelete`)
but no service methods or HTTP handlers yet — that's Epic 3's job, reusing this exact shape.
