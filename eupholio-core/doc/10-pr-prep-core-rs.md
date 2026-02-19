# PR Prep (core-rs -> main)

This branch is `core-rs`. It introduces the Rust core rebuild in incremental steps.

## Change summary

### Core implementation
- Implemented `moving_average` / `total_average`
- Added carry-in support (`calculate_total_average_with_carry_and_rounding`)
- Applied report-only rounding

### CLI
- Introduced subcommands (`calc`, `validate`, `version`)
- Added external rounding injection (`rounding`)
- Structured `validate` issue codes (centrally managed via enums)
- Added `EVENT_YEAR_MISMATCH` warning

### Quality
- Golden tests
- CLI end-to-end tests
- Extended Go/Rust parity script and fixtures
- Added GitHub Actions workflow (`rust-core.yml`)

### Docs
- Organized documentation from `doc/01` to `doc/10`
- Added rounding policy / validation codes / normalizer interface docs

## PR body template (draft)

### What
- Add Rust-based `eupholio-core` with switchable cost methods and carry-in support.
- Introduce subcommand-based CLI and structured validation issue codes.
- Add parity and CI scaffolding.

### Why
- Isolate the accounting engine from exchange-specific ingestion.
- Prepare for wasm/local execution and safer migration from Go.

### Notes
- `per_event` rounding timing is implemented (golden/parity fixtures added to lock in behavior differences).
- `per_year` rounding timing is implemented for TotalAverage yearly finalization (non per-event).
- `report_only` is the default.

### Verification
- `cargo test` passed
- `scripts/compare_go_rust.py` passed on fixture set

## Recommended merge strategy
- Squash merge is recommended (commit history is extensive)
- Suggested PR title: `feat: introduce rust eupholio-core with cli/validation/parity foundation`
