# eupholio-core

Rust implementation of the eupholio cost-calculation core.

## Current status

Implemented:
- Cost methods: `moving_average`, `total_average`
- Carry-in support for total-average
- CLI subcommands: `calc`, `validate`, `version`
- External rounding override input (`rounding`)
- Go vs Rust parity scripts
- Year mismatch handling: events whose `event.ts` year differs from `tax_year` are emitted as warnings (`YearMismatch` / `EVENT_YEAR_MISMATCH`) and excluded from calculation
- Per-year rounding consistency for total-average: `carry_out_cost` is derived from rounded `carry_out_qty × average_cost_per_unit` (then rounded by JPY rule)

Not yet fully implemented:
- Exchange-specific normalizers

## Quickstart

```bash
cd eupholio-core
cargo test
```

### Calculate

```bash
cat input.json | cargo run --quiet --bin eupholio-core-cli -- calc
# default subcommand
cat input.json | cargo run --quiet --bin eupholio-core-cli
```

### Validate

```bash
cat input.json | cargo run --quiet --bin eupholio-core-cli -- validate
```

### Version

```bash
cargo run --quiet --bin eupholio-core-cli -- version
```

## Input example (rounding override)

```json
{
  "method": "total_average",
  "tax_year": 2026,
  "carry_in": {
    "BTC": {"qty": "2", "cost": "8000000"}
  },
  "rounding": {
    "currency": {"JPY": {"scale": 0, "mode": "half_up"}},
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

## Parity check

From the repository root:

```bash
scripts/compare_go_rust.py
```

## Docs

- `doc/README.md`
- `doc/09-validation-codes.md`
