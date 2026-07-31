---
name: code-review
description: Use this skill when the user asks to review code, review a PR, review a diff, or check code before committing/merging in this Go project. Covers Go idioms, Light DDD layering adherence, and money/security correctness — weighted equally, since a stylistically clean handler that mishandles money is a worse bug than an ugly one that's correct.
---

# Code Review

Review the given code (diff, PR, or file set) against three equally-weighted categories. Don't
skip a category because the diff is small — a one-line change can still violate a layering rule
or a money-handling rule.

## 1. Go Idioms & Style

- Errors wrapped with context (`fmt.Errorf("doing X: %w", err)`), not swallowed or returned bare.
- No naked `panic()` in request-handling or job code paths outside of genuinely unrecoverable
  startup failures.
- Receiver naming, package naming, and file organization follow standard Go conventions (short
  receiver names, lowercase package names, no stutter like `account.AccountService`).
- `context.Context` is the first parameter and is actually propagated through to DB calls, not
  just accepted and ignored.
- No unstructured `fmt.Println`/`log.Println` in request or job code — see CLAUDE.md §7.

## 2. Light DDD Layering Adherence

Check against CLAUDE.md §3. Specifically:

- **Handlers** (`internal/transport/http/`) only parse requests, call a service method, and write
  responses. Flag any business logic (validation beyond basic shape-checking, conditional
  branching on domain rules, direct DB queries) found in a handler.
- **Services** (`internal/<domain>/service.go`) depend on the **repository interface**, never on
  `sqlc`-generated types or a concrete Postgres struct directly. If a service file imports
  anything from `internal/platform/postgres`, that's a layering violation — flag it.
- **Repository interfaces** are owned by the domain package; **implementations** live in
  `internal/platform/postgres`. Flag a repository interface defined in the wrong package, or
  business logic (e.g. balance computation) implemented inside a `postgres` repository method
  instead of in the service layer.
- **Entities** are constructed via a constructor function that enforces invariants, not built as
  a bare struct literal from arbitrary external input. Flag any code path that bypasses the
  constructor (e.g. handler code building an entity struct directly instead of calling
  `NewAccount(...)`).

## 3. Money & Security Correctness

Check against CLAUDE.md §4, §5, §6.

- **No floats for money.** Any `float32`/`float64` touching an amount, balance, or rate is a hard
  flag — money fields are `int64` (minor units) or `numeric` (FX rates) only.
- **Minor-unit exponent isn't assumed to be 2.** If code multiplies/divides by a hardcoded 100
  anywhere in currency conversion or display logic, flag it — the exponent must be looked up per
  currency.
- **FX snapshot immutability.** Any code path that recalculates `fx_rate_to_reporting` or
  `converted_amount_minor_units` on an existing transaction (rather than only setting them once
  at creation) is a serious flag — this breaks historical report stability.
- **Idempotency key enforcement** on `POST /transactions` and `POST /transfers` specifically — a
  missing check here is a serious flag, not a nitpick.
- **RFC 9457 error format** — flag any handler returning a bare string, a non-problem+json error
  body, or a raw Go error/panic trace to the client.
- **Ownership checks** — every handler/service method that reads or mutates a resource by ID
  must verify the resource belongs to the authenticated caller (or, for household-shared reads,
  belongs to an approved household member). Flag any query that trusts a client-supplied ID
  without an ownership/membership check.
- **Soft delete correctness** — flag any `DELETE` that physically removes a row from a table
  listed in CLAUDE.md §5.6, and flag any query elsewhere that forgot to filter out
  `deleted_at IS NOT NULL` rows.

## Output format

Structure the review as:

```
## Summary
[1-2 sentences: overall assessment]

## Must Fix
- [category] file:line — issue, and why it matters

## Should Fix
- [category] file:line — issue

## Nitpicks
- file:line — minor style/naming

## What's Good
- [call out at least one thing done well, specifically — not generic praise]
```

Categorize each finding with its section (`Go idiom`, `DDD layering`, `Money/security`) so the
person can see whether one category is systematically weaker than the others across the diff.
