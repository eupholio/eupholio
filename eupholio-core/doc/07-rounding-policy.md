# Rounding Policy (External Config)

In `eupholio-core`, rounding rules are not hard-coded in the core; they are designed to be injected via external configuration.

## Objective

- Support differences across countries, tax regimes, and operational policies
- Allow switching compatibility modes (rounding behavior from existing implementations)
- Preserve internal calculation precision and adjust only at final output

## Proposed Japan Default

- Realized PnL (JPY): scale=0, mode=half_up
- Income (JPY): scale=0, mode=half_up
- Average unit price: scale=8, mode=half_up
- Quantity: scale=8, mode=half_up
- timing: report_only

## Example JSON Config

```json
{
  "jurisdiction": "JP",
  "rounding": {
    "currency": {
      "JPY": { "scale": 0, "mode": "half_up" }
    },
    "unit_price": { "scale": 8, "mode": "half_up" },
    "quantity": { "scale": 8, "mode": "half_up" },
    "timing": "report_only"
  }
}
```

## Timing

- `report_only`: Do not round intermediate calculations; round only at output time
- `per_event`: Round after each event is processed
- `per_year`: Round at annual closing/aggregation

## Implementation Policy (Phased Rollout)

1. Introduce config structs (use JP defaults when unspecified)
2. Apply rounding in the report output layer (`report_only`)
3. Add `per_event` / `per_year` for compatibility use cases

## Implementation Tasks (Next Phase)

### per_event

- Purpose: apply rounding after each event application
- Implemented:
  - Added fixtures that reproduce results different from `report_only` for the same input
  - In `validate`, do not emit the `ROUNDING_TIMING_NOT_FULLY_IMPLEMENTED` warning when `per_event` is specified
  - Added 2 `per_event` cases to golden tests

### per_year

- Purpose: apply rounding when annual aggregation is completed
- Implemented:
  - In TotalAverage, apply rounding when finalizing `yearly_summary.by_asset` (not per event)
  - `realized_pnl_jpy` is calculated by summing annually rounded values per asset
  - Applied similarly for annual aggregation with carry-in
  - Added `per_year` fixtures to `compare_go_rust.py` (for visualization)
- Constraints (explicit):
  - `timing=per_year` is not supported when `method=moving_average`
  - CLI `validate` returns `ROUNDING_PER_YEAR_UNSUPPORTED_FOR_MOVING_AVERAGE` as an error
  - CLI `calc` also rejects this condition as an input error (no implicit fallback)
