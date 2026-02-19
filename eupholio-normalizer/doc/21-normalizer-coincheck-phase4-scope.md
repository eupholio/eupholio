# 21. Coincheck normalizer phase-4 scope

This phase extends Coincheck normalization beyond phase-3 transfer support.

## Goal

Add minimal, explicit handling for fiat transfer operations while keeping tax-event semantics deterministic.

## Target operations (phase-4)

- `Withdrawal` (JPY) -> `Event::Transfer { direction: TransferDirection::Out }`
- `Deposit` (JPY) -> `Event::Transfer { direction: TransferDirection::In }`

## Scope rules

- Only JPY fiat transfer rows are in scope for this phase.
- Non-JPY fiat/asset variants remain unsupported and must emit diagnostics.
- Existing operations from phase-2/3 remain unchanged.

## Mapping draft

- `id` ->
  - `id = "{id}:fiat_transfer_in"` for `Deposit`
  - `id = "{id}:fiat_transfer_out"` for `Withdrawal`
- `trading_currency` -> `asset` (must be `JPY` in this phase)
- `amount` -> `qty` (must be `> 0`)
- `time` -> `ts` (UTC-normalized)

## Validation

Hard errors:
- invalid/missing required headers
- invalid datetime
- invalid decimal amount
- non-positive qty for in-scope operations

Diagnostics:
- unsupported operation rows
- out-of-scope currency for phase-4 fiat mapping

## Test plan

1. fixture: `coincheck_history_fiat_transfers_smoke.csv`
2. fixture: expected normalized events JSON
3. tests:
   - Deposit/Withdrawal mapped as Transfer In/Out
   - non-JPY Deposit/Withdrawal diagnosed
   - existing phase-2/3 mappings unchanged

## Exit criteria

- deterministic mapping and validation documented
- fixture + smoke coverage added
- all existing coincheck tests remain green
