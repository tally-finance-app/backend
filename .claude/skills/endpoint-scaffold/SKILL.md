---
name: endpoint-scaffold
description: Use this skill when the user asks to scaffold a new domain module, start a new epic's implementation, or create the boilerplate for a new entity (e.g. "scaffold the CreditCard module"). Generates the same layered structure as the Epic 0 Account reference vertical slice, so every domain module follows an identical shape.
---

# Endpoint Scaffold

Epic 0 builds one full vertical slice (Account) specifically to be copied — every other domain
module should follow the exact same shape (CLAUDE.md §3). This skill generates that shape for a
new domain concept so the pattern doesn't drift module to module.

## Step 1: gather what's needed

Ask (if not already given):
- The domain name (e.g. `creditcard`, `transaction`) — must match an existing entry in the
  Requirements & Domain Model's field list.
- The entity's fields, pulling from the Notion ER Diagram / Requirements doc if available, or
  asking the person to paste the relevant entity section.
- Which operations are needed (Create/List/GetByID/Update/SoftDelete, or a subset — not every
  entity needs all five; check the relevant Linear epic's stories for which endpoints actually
  exist).

## Step 2: generate the layers, in this order

1. **`internal/<domain>/model.go`** — the entity struct and a constructor function
   (`New<Entity>(...)`) that validates required fields and returns an error for invalid input.
   Never expose a way to construct the entity that bypasses validation.

2. **`internal/<domain>/repository.go`** — the repository interface only. Method signatures
   should mirror the operations gathered in Step 1, taking `context.Context` first.

3. **sqlc query file** (`internal/platform/postgres/<domain>_queries.sql`, per `sqlc.yaml` —
   lives with the Postgres adapter, not the domain package, since raw SQL is an infrastructure
   concern) — hand-written SQL for each repository method, with named parameters.

4. **`internal/platform/postgres/<domain>_repository.go`** — the sqlc-backed implementation of
   the repository interface from step 2. This file should be a thin adapter — no business logic,
   just mapping between the interface and generated sqlc calls.

5. **`internal/<domain>/service.go`** — use-case logic, depending only on the repository
   interface (never importing `internal/platform/postgres` directly — that's the layering
   violation the `code-review` skill flags).

6. **`internal/transport/http/<domain>_handler.go`** — HTTP handlers calling the service. Parse
   request, call service, use the RFC 9457 error helper for any error path, write response.

7. **Route registration** — wire the new handler's routes into the router in `cmd/api/main.go`.

## Step 3: tests alongside, not after

For each layer generated, also generate its test scaffold (empty table-driven test functions with
TODO cases, at minimum) — don't generate implementation code without corresponding test
scaffolding, per this project's testing conventions (see the `test-writing` skill).

## Step 4: cross-check against the relevant Linear epic

Before finishing, check the domain's Linear epic (e.g. Epic 4 for Credit Cards) and confirm the
scaffolded endpoints match what the epic's stories actually specify — don't scaffold an operation
that isn't in scope for this domain (e.g. don't add a balance endpoint to Category, which has no
such concept), and don't skip one that a specific story requires (e.g. Credit Cards needs its own
balance endpoint, distinct from Account's, per Epic 4's stories).

## Explicitly do not

- Do not hardcode business rules (e.g. minor-unit rounding, FX lookup, statement-attachment
  logic) inline in a scaffolded handler or repository — those live in the service layer, and for
  genuinely cross-cutting rules (FX conversion, RFC 9457 errors), call the existing shared
  helper rather than reimplementing it per domain.
