# CLI

Binary:
- `eupholio-core-cli`

Reads JSON from standard input and writes the computed `Report` JSON to standard output.

## Subcommands

- `calc`: Compute from JSON input and output a report
- `validate`: Validate JSON input (exits non-zero if there are errors)
  - Returns `code` / `level` / `message` in `issues[]`
- `version`: Show CLI version

For backward compatibility, `calc` is also executed as the default when the subcommand is omitted.

## Run

```bash
cd eupholio-core
cat input.json | cargo run --quiet --bin eupholio-core-cli -- calc
# Compatibility: calc can be omitted
cat input.json | cargo run --quiet --bin eupholio-core-cli

# Validation only
cat input.json | cargo run --quiet --bin eupholio-core-cli -- validate
```

## Input (moving_average)

```json
{
  "method": "moving_average",
  "tax_year": 2026,
  "events": [
    {"type":"Acquire","id":"a1","asset":"BTC","qty":"1","jpy_cost":"3000000","ts":"2026-01-01T00:00:00Z"},
    {"type":"Dispose","id":"d1","asset":"BTC","qty":"0.5","jpy_proceeds":"2000000","ts":"2026-02-01T00:00:00Z"}
  ]
}
```

## Input (total_average + carry_in + rounding override)

```json
{
  "method": "total_average",
  "tax_year": 2026,
  "carry_in": {
    "BTC": {"qty":"2","cost":"8000000"}
  },
  "rounding": {
    "currency": {
      "JPY": {"scale": 0, "mode": "half_up"}
    },
    "unit_price": {"scale": 8, "mode": "half_up"},
    "quantity": {"scale": 8, "mode": "half_up"},
    "timing": "report_only"
  },
  "events": [
    {"type":"Acquire","id":"a1","asset":"BTC","qty":"1","jpy_cost":"6000000","ts":"2026-01-05T00:00:00Z"},
    {"type":"Dispose","id":"d1","asset":"BTC","qty":"1","jpy_proceeds":"7000000","ts":"2026-02-01T00:00:00Z"}
  ]
}
```

## Example differences for `rounding.timing` (comparison using the same minimal input)

The following comparison uses the same input as `tests/fixtures/per_year_total_difference.json` (`method=total_average`), with only `timing` switched.

### Common input

```json
{
  "method": "total_average",
  "tax_year": 2026,
  "rounding": {
    "currency": {"JPY": {"scale": 0, "mode": "half_up"}},
    "unit_price": {"scale": 8, "mode": "half_up"},
    "quantity": {"scale": 8, "mode": "half_up"},
    "timing": "report_only"
  },
  "events": [
    {"type":"Acquire","id":"a1","asset":"BTC","qty":"1","jpy_cost":"100","ts":"2026-01-01T00:00:00Z"},
    {"type":"Dispose","id":"d1","asset":"BTC","qty":"1","jpy_proceeds":"100.49","ts":"2026-01-02T00:00:00Z"},
    {"type":"Acquire","id":"a2","asset":"ETH","qty":"1","jpy_cost":"100","ts":"2026-01-03T00:00:00Z"},
    {"type":"Dispose","id":"d2","asset":"ETH","qty":"1","jpy_proceeds":"100.49","ts":"2026-01-04T00:00:00Z"}
  ]
}
```

Run `calc` with `timing` set to `report_only` / `per_event` / `per_year`.

### Expected differences (excerpt)

| timing | realized_pnl_jpy (report) | yearly_summary.by_asset.BTC.realized_pnl_jpy | yearly_summary.by_asset.ETH.realized_pnl_jpy |
|---|---:|---:|---:|
| `report_only` | `1` | `0` | `0` |
| `per_event` | `0` | `0` | `0` |
| `per_year` | `0` | `0` | `0` |

Notes:
- `report_only=1`, `per_year=0` is fixed in the fixture (`per_year_total_difference.json`).
- With the same input, `per_event` is also `0`; in this case it matches `per_year`.

## Current timing behavior in `validate` / `calc`

- `rounding.timing=per_event`: no warning (implemented)
- `rounding.timing=per_year`:
  - `method=total_average`: no warning (implemented)
  - `method=moving_average`: returns `ROUNDING_PER_YEAR_UNSUPPORTED_FOR_MOVING_AVERAGE` as an **error**

`calc` also rejects the same condition (`moving_average + per_year`) as an input error, with no implicit fallback.

- Regression tests: `tests/cli_e2e.rs`
  - `cli_validate_per_event_no_timing_warning`
  - `cli_validate_per_year_total_average_no_timing_warning`
  - `cli_validate_ng_per_year_for_moving_average`
  - `cli_calc_ng_per_year_for_moving_average`

On the other hand, `validate` still returns non-timing warnings/errors as usual (e.g., `EVENT_YEAR_MISMATCH`, `DUPLICATE_EVENT_ID`).
