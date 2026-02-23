# Cryptact custom CSV action support matrix (current)

This document summarizes which `Action` values are currently mapped by
`eupholio-normalizer/src/cryptact.rs`, and what constraints apply.

Scope note:
- This is a **current-state matrix** for operator/developer reference.
- `counter != JPY` is currently treated as unsupported for all actions.

## Supported actions

| Action | Mapped Event | Main constraints | Notes |
| --- | --- | --- | --- |
| `BUY` | `Event::Acquire` | `price > 0`; `volume > 0`; `fee_ccy` must be `JPY` or `Base` | If `fee_ccy == Base`, acquired qty becomes `volume - fee` |
| `SELL` | `Event::Dispose` | `price > 0`; `volume > 0`; `fee_ccy` must be `JPY` or `Base` | If `fee_ccy == Base`, disposed qty becomes `volume + fee` |
| `PAY` | `Event::Dispose` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` | Missing price is allowed (`jpy_proceeds = 0`) |
| `MINING` | `Event::Income` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` | Missing price is allowed (`jpy_value = 0`) |
| `SENDFEE` | `Event::Transfer(Out)` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` |  |
| `BONUS` | `Event::Income` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` | Missing price is allowed (`jpy_value = 0`) |
| `LENDING` | `Event::Income` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` | Missing price is allowed (`jpy_value = 0`) |
| `STAKING` | `Event::Income` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` | Missing price is allowed (`jpy_value = 0`) |
| `TIP` | `Event::Dispose` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` | Missing price is allowed (`jpy_proceeds = 0`) |
| `LOSS` | `Event::Dispose` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` | Always `jpy_proceeds = 0` |
| `REDUCE` | `Event::Transfer(Out)` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` |  |
| `LEND` | `Event::Transfer(Out)` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` | phase-5 |
| `RECOVER` | `Event::Transfer(In)` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` | phase-5 |
| `BORROW` | `Event::Transfer(In)` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` | phase-5 |
| `RETURN` | `Event::Transfer(Out)` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` | phase-5 |
| `DEFIFEE` | `Event::Dispose` | `fee == 0`; `fee_ccy == JPY`; `volume > 0` | Always `jpy_proceeds = 0` |

## Known unsupported action

| Action | Outcome | Reason |
| --- | --- | --- |
| `CASH` | diagnostic (`Unsupported`) | `CASH is not supported in rust-core Event model yet` |

## Diagnostic vs hard error behavior

- **Unsupported/diagnostic** (`NormalizeResult.diagnostics`):
  - unknown action
  - non-JPY `counter`
  - unsupported fee currency combinations (depends on action)
  - `CASH`
- **Hard error** (`Err(...)`):
  - `volume <= 0`
  - `fee < 0`
  - required positive `price` missing/invalid for `BUY`/`SELL`
  - non-zero fee on actions that require `fee == 0`

## Reference

- Mapper implementation: `eupholio-normalizer/src/cryptact.rs`
- Coverage tests: `eupholio-normalizer/tests/cryptact_poc.rs`
