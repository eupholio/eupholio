# Domain Model

## Config

- `method`: `MovingAverage` or `TotalAverage`
- `tax_year`: target year

## Event

- `Acquire { asset, qty, jpy_cost, ts }`
- `Dispose { asset, qty, jpy_proceeds, ts }`
- `Income { asset, qty, jpy_value, ts }`
- `Transfer { asset, qty, direction, ts }`

Notes:
- Includes an `id`, used for duplicate event detection.
- All events are assumed to be pre-normalized (JPY values are finalized).

## Report

- `positions`: remaining quantity and average unit cost per asset
- `realized_pnl_jpy`: total realized profit/loss
- `income_jpy`: income-equivalent amount (from `Income` events)
- `yearly_summary`: yearly summary for the total-average method
- `diagnostics`: warnings

## Warning

- `DuplicateEventId`
- `NegativePosition`
- `YearMismatch` (the year in `event.ts` differs from `tax_year`; a warning is issued and the event is excluded from calculation)
- `YearBoundaryCarry`
