---
name: pr-watchdog
description: Monitor a GitHub pull request continuously using gh CLI. Use when you need to check CI/check-runs status, collect new review comments, detect unresolved threads, and report actionable deltas without reprocessing old items.
---

# PR Watchdog

Monitor one PR in a loop and report only new, actionable changes.

## Inputs

- `repo` (e.g. `eupholio/eupholio`)
- `pr` number
- optional baseline timestamp / known comment ids

## Workflow

1. Check CI state.
   - `gh pr checks <pr> --repo <repo>`
2. Pull review comments.
   - `gh api repos/<repo>/pulls/<pr>/comments`
3. Pull unresolved review threads.
   - GraphQL `reviewThreads` query; filter `isResolved=false`.
4. Compare against previous snapshot (ids/timestamps).
5. Output only deltas:
   - newly failed checks
   - newly added comments
   - unresolved thread count changes

## Triage Rules

- If all checks pass and no new comments: report `NO_CHANGE`.
- If comments are only style/typo/doc nits: classify as `safe-fix candidates`.
- If comments touch calculation semantics/API contracts: classify as `human decision required`.

## Suggested Output Format

- CI: pass/fail summary
- New comments: count + links
- Unresolved threads: count
- Next action: `none` / `run pr-autofix-safe` / `ask human`
