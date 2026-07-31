---
name: commit-pr-writer
description: Use this skill when the user asks to write a commit message, write a PR description, or prepare a pull request for this project. Ties commit/PR content back to the relevant Linear ticket's structure (Background, Acceptance Criteria, Definition of Done).
---

# Commit & PR Writer

Commits and PRs in this project reference their Linear ticket ID and should make it easy for
future-you (or a reviewer) to see which Acceptance Criteria are actually addressed — not just a
generic summary of what changed.

## Commit messages

Format:
```
<TICKET-ID>: <concise imperative summary>

<optional body: what changed and why, if not obvious from the summary>
```

Example:
```
TALLY-109: add POST /transactions with idempotency key support

Implements transaction creation with FX rate snapshotting and the
statement-attachment path for late entries. Idempotency-Key header
required per Epic 10's shared middleware.
```

- Imperative mood ("add", not "added" or "adds").
- Reference the ticket ID at the start, not buried in the body.
- If the commit is part of a larger ticket (not the final commit), say so ("part of
  TALLY-109: ...") rather than implying the ticket is fully done.

## PR descriptions

Structure:

```markdown
## TALLY-XXX: <ticket title>

<one-paragraph summary of what this PR does>

### Acceptance Criteria Addressed
- [x] <criterion from the ticket, copied verbatim or near-verbatim>
- [x] <criterion>
- [ ] <criterion not yet addressed, if this PR is partial — be explicit about what's left>

### Definition of Done
- [x/ ] <the specific test scenario named in the ticket> — <link to the test or brief note>

### Out of Scope (per ticket)
<brief note confirming nothing from the ticket's Out of Scope section crept in>

### Notes for the reviewer
<anything non-obvious: a deliberate deviation from the ticket, an edge case handled differently
than originally planned, a follow-up ticket that should be filed>
```

- Pull the Acceptance Criteria list directly from the Linear ticket (fetch via Linear MCP if
  available, or ask the person to paste it) — don't paraphrase from memory, since the specific
  wording of edge-case criteria matters.
- If this PR only partially completes a ticket, say so explicitly in both the summary and the
  checklist — don't let an unchecked box speak for itself without a note explaining why.
- Cross-check against the `linear-ac-check` skill's output if it's been run for this ticket — the
  PR description should not claim criteria are met if that check found gaps.
