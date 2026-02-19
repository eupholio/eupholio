# Go parity testing

A script is provided to compare PnL results between the Go implementation (`mam`/`wam`) and the Rust implementation.

## Prerequisites

- Go toolchain
- Rust toolchain

## Run

```bash
cd eupholio
scripts/compare_go_rust.py
```

## Current fixtures

- `parity_fixture_case1.json` (basic buy/sell)
- `parity_fixture_case3.json` (crypto-to-crypto decomposition)
- `parity_fixture_transfer.json` (mixed Transfer events)
- `parity_fixture_fractional.json` (fractional precision)
- `parity_fixture_carry_in.json` (cross-year carry-over, total average)
- `parity_fixture_per_event_moving.json` (visualizing per_event rounding differences: moving)
- `parity_fixture_per_event_total.json` (visualizing per_event rounding differences: total)
- `parity_fixture_per_year_total.json` (visualizing per_year rounding differences: total)

## Judgment criteria

- PnL is compared as `Decimal` (tiny differences are tolerated)
- `check_moving` / `check_total` can be toggled per case
- `parity_fixture_per_event_*` / `parity_fixture_per_year_*` are not pass/fail targets for Go parity; they are fixtures used to lock timing differences on the Rust side
