---
name: test-writing
description: Use this skill when the user asks to write tests, add test coverage, or implement the "Definition of Done" test scenario for a Linear ticket in this Go project. Distinguishes service-layer unit tests (fake repository, no DB) from repository-layer integration tests (real Dockerized Postgres), per this project's testing conventions.
---

# Test Writing

Match tests to the layer being tested — this project deliberately separates fast, DB-free
service-layer tests from real-DB repository-layer tests (see CLAUDE.md §8). Don't blur the two.

## Step 1: identify what's being tested

- **Service method** (`internal/<domain>/service.go`) → unit test with a fake repository.
- **Repository implementation** (`internal/platform/postgres/`) → integration test against real
  Postgres (via the Docker Compose instance from Epic 0).
- **HTTP handler** (`internal/transport/http/`) → can go either way depending on what's being
  verified: request parsing/response shape (fake service, no DB) vs. full end-to-end behavior
  (real DB, real service). Prefer fake-service tests for handlers; reserve full end-to-end tests
  for the small number of scenarios that genuinely need to prove the whole stack works together
  (see the Epic 0 reference vertical slice for the pattern).

## Step 2: service-layer unit tests

- Write or reuse a hand-written fake implementing the domain's repository interface — in-memory,
  no `sqlc`, no `pgx`, no real DB connection.
- Use table-driven tests as the default shape:
  ```go
  func TestCreateAccount(t *testing.T) {
      tests := []struct {
          name    string
          input   CreateAccountInput
          wantErr string
      }{
          {name: "valid account", input: ..., wantErr: ""},
          {name: "missing currency", input: ..., wantErr: "currency is required"},
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) { ... })
      }
  }
  ```
- Cover the happy path AND every edge case named in the Linear ticket's Acceptance Criteria —
  don't stop at the happy path if the ticket names specific failure/edge behaviors.
- For money-related logic (FX conversion, balance computation, statement totals), include at
  least one test case that would fail if float arithmetic were used instead of integer minor
  units (e.g. a rounding-boundary case) — this project has had real bugs hide behind
  float-shaped test cases that happened to round cleanly.

## Step 3: repository-layer integration tests

- Run against the real Dockerized Postgres — assume `make db-up && make migrate-up` has been run
  (or wire test setup to do this automatically if a test harness for this doesn't exist yet).
- Test the actual SQL: constraint enforcement (unique, foreign key, not-null), soft-delete
  filtering (a soft-deleted row must not appear in a subsequent `List`/`GetByID` call), and any
  Postgres-specific behavior (e.g. `numeric` precision for FX rates).
- Clean up test data between tests (transaction rollback per test, or truncate between runs) —
  don't let integration tests leak state into each other.

## Step 4: cross-check against the ticket

If a Linear ticket is available (ID or pasted content), confirm every bullet under its
"Acceptance criteria" and "Definition of done" sections maps to at least one test. If the ticket
names a specific scenario ("a test seeds a known combination of transactions/transfers and
asserts the computed balance against a manually calculated expected value"), that scenario should
exist as a literal test case, not just be "covered in spirit" by a more generic test.

## Common gaps to check for before declaring tests complete

- Ownership/authorization boundary tests (does the test suite include a case where the resource
  belongs to a different user?).
- Soft-delete exclusion (is there a test confirming a soft-deleted record doesn't leak into
  results?).
- The specific edge cases named in this project's tickets (e.g. `close_day = 31` in a short
  month, idempotency key replay, FX rate rounding boundary, category parent-cycle detection).
