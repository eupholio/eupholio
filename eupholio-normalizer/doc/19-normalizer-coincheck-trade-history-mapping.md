# 19. Phase-2 normalizer mapping: Coincheck trade history CSV

This document freezes the phase-2 mapping for Coincheck trade history CSV.

- Source: `pkg/coincheck` extractor/translator path
- Raw format: Coincheck trade history CSV
- Scope: `Completed trading contracts` rows only (Acquire/Dispose)

## Conversion target

`eupholio-core::event::Event`

## Row-level rules

- `operation = Completed trading contracts` -> mapped to `Event::Acquire` or `Event::Dispose`
- other `operation` values -> skipped with explicit diagnostic (`unsupported operation`)

## Side detection

A completed-trade row contains comment metadata like:

`Rate: 5000000.0, Pair: btc_jpy`

where pair means `QUOTE_BASE`.

- If `trading_currency == QUOTE` -> `Event::Acquire` of `QUOTE`
- If `trading_currency == BASE` -> `Event::Dispose` of `QUOTE`
- Otherwise -> hard error (`invalid trading currency`)

This phase only supports `BASE = JPY`.

## Field mapping table

| Source column | Acquire mapping | Dispose mapping | Notes |
| --- | --- | --- | --- |
| `id` | `id = "{id}:acquire"` | `id = "{id}:dispose"` | Stable event ID |
| `comment` (`Pair`) | `asset = QUOTE` | `asset = QUOTE` | Upper-cased |
| `amount` | `qty = amount` | `qty = amount / rate` | Dispose amount is JPY received |
| `comment` (`Rate`) | used in cost formula | used in qty formula | Decimal rate |
| `fee` | used in cost formula | used in proceeds formula | blank fee = 0 |
| `time` | `ts = time` | `ts = time` | Parsed `%Y-%m-%d %H:%M:%S %z` then converted to UTC |

## JPY formulas

- Acquire: `jpy_cost = (rate * amount) + fee`
- Dispose: `jpy_proceeds = amount - fee`

## Known limits (intentional in phase-2)

- No `Received`, `Sent`, `Bank Withdrawal`, or cancel rows yet (diagnosed and skipped)
- Requires parsable `Rate: ..., Pair: ...` comment shape
- Supports only `*_jpy` pairs for now
