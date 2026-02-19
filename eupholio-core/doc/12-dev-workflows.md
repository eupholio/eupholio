# 12. Practical Dev Workflows

This note captures reusable loops for day-to-day core development and PR review in this repository.

## eupholio-core-dev loop

Use this for normal feature/fix work inside `eupholio-core`.

1. **Sync and branch**
   - `git fetch origin`
   - `git switch <base>`
   - `git pull --ff-only`
   - `git switch -c <topic-branch>`
2. **Implement in small commits**
   - Prefer one concern per commit (logic, tests, docs).
3. **Run focused checks first**
   - `cd eupholio-core && cargo test --all-targets`
4. **Run repo-level checks before PR**
   - `go test ./...`
   - `go test ./test/integration/...`
   - `scripts/check_validation_codes.py`
   - `scripts/compare_go_rust.py`
5. **Prepare PR**
   - Rebase on latest base branch.
   - Keep PR body short: motivation, change summary, validation run.

## github-pr-review-loop

Use this when addressing review comments quickly and safely.

1. **Triage comments**
   - Group comments into: must-fix, optional, clarification-only.
2. **Patch by group**
   - Apply related changes together so each commit has a clear intent.
3. **Re-run impacted checks only, then full required set**
   - Start with nearby tests, then run mandatory checks from `AGENTS.md`.
4. **Reply with evidence**
   - For each resolved thread, reference exactly what changed and where.
5. **Final pass**
   - `git diff --stat origin/<base>...HEAD`
   - Confirm docs/fixtures stayed aligned with behavior changes.

## Quick commands

```bash
# core-only checks
cd eupholio-core
cargo test --all-targets

# full mandatory checks from repo root
go test ./...
go test ./test/integration/...
scripts/check_validation_codes.py
scripts/compare_go_rust.py
```
