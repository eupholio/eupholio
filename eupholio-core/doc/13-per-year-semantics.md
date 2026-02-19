# 13. Per-year semantics lock

This note locks current `per_year` behavior in `eupholio-core` to avoid interpretation drift.

## Scope

- Applies to `method=total_average` with `rounding.timing=per_year`.
- `method=moving_average` + `timing=per_year` is not supported by CLI/input validation (rejected there); in the core library `calculate` path it is accepted but treated as `rounding.timing=report_only`.

## Locked behavior

1. Event processing is performed without per-event rounding.
2. Rounding is applied when finalizing yearly summary values.
3. For each asset summary:
   - `average_cost_per_unit` is rounded by `rounding.unit_price`.
   - `carry_out_qty` is rounded by `rounding.quantity`.
   - `carry_out_cost` is computed from the rounded fields and then rounded by JPY rule:
     - `carry_out_cost = round_jpy(rounded_carry_out_qty * rounded_average_cost_per_unit)`
4. Invariant target in yearly summary:
   - `carry_out_cost == round_jpy(carry_out_qty * average_cost_per_unit)` (using `rounding.currency["JPY"]`, defaulting to scale=0 HalfUp).

## Related paths

- `src/engine/total_average.rs`
- `src/lib.rs` (`report_only` rounding coherence path)
- `tests/per_year_carry_consistency.rs`
