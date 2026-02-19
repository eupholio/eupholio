# 20. Coincheck normalizer phase-3 scope

This document defines the next minimal extension after phase-2 (`Completed trading contracts`).

## Goal

Add support for transfer-like rows in Coincheck history with explicit mapping and diagnostics.

## Target operations (phase-3)

- `Received` -> map to `Event::Transfer { direction: TransferDirection::In }`
- `Sent` -> map to `Event::Transfer { direction: TransferDirection::Out }`

Other operations remain out-of-scope and must continue to emit explicit diagnostics.

## Mapping draft

Required headers (same as phase-2 baseline):

- `id`, `time`, `operation`, `amount`, `trading_currency`, `fee`, `comment`
  - `fee` / `comment` remain required for compatibility with phase-2 parser, but are not used in phase-3 transfer mapping.

Mapping:

- `id` -> `id = "{id}:transfer_in" | "{id}:transfer_out"`
- `trading_currency` -> `asset` (upper-cased)
- `amount` -> `qty` (must be `> 0`)
- `time` -> `ts` (existing parser, UTC-normalized)
- `operation` -> transfer direction

## Validation and diagnostics

Hard errors:

- invalid/missing required headers
- invalid datetime
- invalid decimal amount
- non-positive transfer qty

Warnings/diagnostics:

- unsupported operation rows should still be collected as diagnostics (non-silent)

## Test plan

1. fixture: `coincheck_history_phase3_smoke.csv`
2. fixture: expected normalized events JSON
3. e2e smoke:
   - includes `Received` + `Sent` rows
   - flows into `eupholio_core::calculate`
4. error-path tests:
   - zero/negative amount
   - invalid datetime
   - unsupported operation diagnostics unchanged

## Non-goals

- bank withdrawal fiat semantics
- staking/lending-specific operation semantics
- fee reinterpretation for transfer rows

## Exit criteria

- deterministic mapping documented and implemented
- fixtures + tests added
- no regression in phase-2 tests
