---
name: pr-autofix-safe
description: Apply safe, low-risk fixes for GitHub PR review comments and resolve threads automatically. Use for typos, dead code removal, import/order cleanup, non-semantic docs updates, and tooling-path portability fixes; avoid semantic calculation changes without approval.
---

# PR Autofix (Safe)

Apply only low-risk fixes, push, and resolve threads.

## Allowed Auto-Fix Scope

- Typos and naming cleanup
- Import ordering / formatting nits
- Dead code removal
- Documentation wording/sync fixes
- Script portability improvements (e.g., configurable binary path)

## Blocked Scope (Require Human Approval)

- Changes to tax/cost calculation semantics
- Rounding policy behavior changes
- Data model / API contract changes
- Performance trade-off changes that alter algorithmic behavior

## Workflow

1. Fetch PR comments and unresolved threads.
2. Classify each comment (`safe` or `needs-human`).
3. If no safe items, stop and report.
4. Implement safe fixes in minimal commits.
5. Run relevant tests/checks for touched area.
6. Push branch.
7. Resolve only threads directly addressed by the commit.
8. Report:
   - commit hash(es)
   - resolved thread ids/links
   - remaining `needs-human` items

## Guardrails

- Prefer smallest possible patch.
- Do not batch unrelated fixes in one commit.
- If uncertain whether semantic impact exists, escalate to `needs-human`.
