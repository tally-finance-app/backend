---
name: linear-ac-check
description: Use this skill when the user asks to check code against a Linear ticket, verify a ticket is done, or confirm acceptance criteria are met before marking a story complete. Requires either a Linear ticket ID/URL (fetched via the Linear MCP connector if available) or pasted ticket content.
---

# Linear Acceptance-Criteria Check

Every ticket in this project's Linear backlog follows a fixed template (see CLAUDE.md §10):
Background, User story, Scope, Acceptance criteria, Out of scope, Dependencies, Definition of
done. This skill cross-checks actual code against that ticket, section by section — it is not a
general code review (use the `code-review` skill for that); it is specifically "does the code do
what this ticket says it should."

## Step 1: get the ticket content

- If a Linear MCP connector is available, fetch the ticket by ID (e.g. `TALLY-109`) or URL.
- Otherwise, ask the person to paste the ticket's full description.
- Do not proceed on a vague summary of the ticket from memory — the acceptance criteria are
  specific enough that paraphrasing from memory will miss details.

## Step 2: work through each section

**Scope** — confirm every listed endpoint/component actually exists in the codebase. Flag
anything listed in Scope that has no corresponding code.

**Acceptance criteria** — go through every bullet individually. For each one:
- State whether it's met, partially met, or not met.
- Point to the specific code (file/function) that satisfies it, or explain what's missing.
- Pay special attention to bullets describing edge cases or specific numeric/behavioral examples
  (e.g. "a card with `close_day = 31` in a 30-day month") — these are the bullets most likely to
  be silently unmet even when the happy path works.

**Out of scope** — confirm the implementation didn't creep into building any of these. If it did,
that's worth flagging even though it's "extra work done," since unplanned scope has a way of
introducing untested edge cases and diverging from the documented design.

**Dependencies** — confirm anything listed as a dependency actually exists and is used correctly
(e.g. if a ticket depends on the RFC 9457 error helper, confirm the code actually calls that
helper rather than rolling its own error response).

**Definition of done** — this is usually a specific named test scenario. Confirm that exact test
exists (not just "a test that's similar in spirit") and passes.

## Step 3: summarize

```
## TALLY-XXX: <title>

### Acceptance Criteria
- [met/partial/not met] <criterion> — <evidence or gap>
...

### Out-of-Scope Creep
- [none | list anything built beyond scope]

### Definition of Done
- [met/not met] <specific test scenario> — <evidence>

### Verdict
[Ready to mark done / Not yet — needs: ...]
```

Don't recommend marking a ticket done if any Acceptance Criteria bullet is unmet or the named
Definition of Done test doesn't exist — "mostly done" is not "done" for this checklist's purpose.
