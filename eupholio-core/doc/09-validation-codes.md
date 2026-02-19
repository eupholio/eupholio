# Validation Codes

List of `issues[].code` values returned by `eupholio-core-cli validate`.

## Error codes

- `INVALID_METHOD`
  - `method` is neither `moving_average` nor `total_average`
- `NEGATIVE_CARRY_IN_QTY`
  - `carry_in.<asset>.qty < 0`
- `NEGATIVE_CARRY_IN_COST`
  - `carry_in.<asset>.cost < 0`
- `ROUNDING_JPY_SCALE_TOO_LARGE`
  - `rounding.currency.JPY.scale > 18`
- `ROUNDING_UNIT_PRICE_SCALE_TOO_LARGE`
  - `rounding.unit_price.scale > 18`
- `ROUNDING_QUANTITY_SCALE_TOO_LARGE`
  - `rounding.quantity.scale > 18`
- `ROUNDING_PER_YEAR_UNSUPPORTED_FOR_MOVING_AVERAGE`
  - `method=moving_average` and `rounding.timing=per_year`
  - Reason: moving_average has no annual-closing rounding point; this is an explicit error to avoid implicit fallback (effectively equivalent to report_only)
- `DUPLICATE_EVENT_ID`
  - Duplicate event IDs
- `ACQUIRE_QTY_NON_POSITIVE`
  - Acquire qty is 0 or less
- `ACQUIRE_COST_NEGATIVE`
  - Acquire `jpy_cost` is negative
- `DISPOSE_QTY_NON_POSITIVE`
  - Dispose qty is 0 or less
- `DISPOSE_PROCEEDS_NEGATIVE`
  - Dispose `jpy_proceeds` is negative
- `INCOME_QTY_NON_POSITIVE`
  - Income qty is 0 or less
- `INCOME_VALUE_NEGATIVE`
  - Income `jpy_value` is negative
- `TRANSFER_QTY_NON_POSITIVE`
  - Transfer qty is 0 or less

## Warning codes

- `EMPTY_EVENTS`
  - `events` is empty
- `UNUSUAL_TAX_YEAR`
  - `tax_year` is outside the typical range
- `CARRY_IN_IGNORED_FOR_MOVING`
  - `carry_in` is specified with moving_average (currently ignored)
- `CARRY_IN_COST_WITH_ZERO_QTY`
  - cost > 0 while qty=0
- `EVENT_YEAR_MISMATCH`
  - Year of `event.ts` does not match `tax_year`

## Documentation Consistency Check

- The `code` list in implementation (`eupholio-core/src/bin/eupholio-core-cli.rs`) and this document is automatically verified in CI
- Verification script: `scripts/check_validation_codes.py`
- If only one side is changed, the `rust-core` workflow fails, so always update both together

## Operational Guidelines

- If `ok=false`, do not execute calculation
- If only warnings exist, execution is allowed, but keep them in logs
- In automated processing, branch by `code` as the key, and use `message` for human-readable display
