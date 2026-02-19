# PR Ready Summary (core-rs -> main)

Last updated: 2026-02-19 (Asia/Tokyo)
Branch: `core-rs`

## 1) Scope

This PR line introduces/finishes the Rust accounting core track for rounding timing and parity hardening, including:

- rounding timing implementation in core (`per_event`, `per_year`)
- validation warning code extension (`EVENT_YEAR_MISMATCH`)
- parity fixture expansion + strict CI parity checks
- docs updates for CLI/rounding/PR prep

## 2) Key commits grouped by theme

### A. Core logic (rounding timing)
- `37021b6` feat(rounding): implement per_event timing in calculation core
- `d7f6127` feat(core-rs): implement per_year rounding at yearly finalization

### B. Validation / quality gates
- `41e3fcf` feat(validate): add EVENT_YEAR_MISMATCH warning code
- `ec40567` test: add per-event golden/parity fixtures showing timing differences
- `2690a6b` ci: strict-check rounding showcase parity fixtures

### C. Documentation / PR prep
- `4ebc202` docs(core): expand README with cli usage and parity workflow
- `33c4fdd` docs(roadmap): define per_event/per_year rounding implementation plan
- `1040e6a` docs(pr): add core-rs to main pull request prep notes
- `84b09e5` docs(cli): add side-by-side rounding timing examples and validate warning note

## 3) Verification commands and results

Commands expected by repo/CI:

```bash
cd eupholio-core
cargo test --all-targets

cd ..
scripts/compare_go_rust.py
```

### Local execution status in this environment

- `cargo test --all-targets` -> **NOT RUN** (`cargo: command not found`)
- `scripts/compare_go_rust.py` -> **NOT RUN** (requires local Go/Rust toolchains)

### CI parity/tests status interpretation

- Workflow `.github/workflows/rust-core.yml` includes:
  - `cargo test --all-targets`
  - `scripts/compare_go_rust.py`
- Therefore PR gate relies on CI green for final pass/fail in this environment.

## 4) Branch cleanliness / parity readiness

- Current branch: `core-rs`
- HEAD: `84b09e5`
- Merge base vs `main`: `dfadf12`
- Commits ahead of `main`: 9
- Working tree: clean (`git status -sb` => `## core-rs`)

## 5) Known limitations / follow-ups

- Exchange-specific normalizers are still outside `eupholio-core` scope.
- Parity validation is fixture-based (not exhaustive over all historical data shapes).
- Final confidence depends on CI environment where both Go and Rust toolchains are available.
- `EVENT_YEAR_MISMATCH` is a warning-oriented guard; policy escalation (warning/error) can be revisited.

## 6) PR note (concise)

- No history rewrite performed.
- This step only prepares PR-facing summary and checks branch readiness metadata.
