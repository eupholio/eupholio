# 15. Phase-1 normalizer mapping: Bittrex OrderHistory CSV

This document freezes the phase-1 bootstrap mapping for one existing source already present in the legacy Eupholio repo:

- Source: `pkg/bittrex` extractor/translator path
- Raw format: Bittrex `OrderHistory` CSV
- Scope: `LIMIT_BUY` and `LIMIT_SELL` happy-path rows only

## Conversion target

`eupholio-core::event::Event`

## Row-level rules

- `OrderType = LIMIT_BUY` -> `Event::Acquire`
- `OrderType = LIMIT_SELL` -> `Event::Dispose`
- other `OrderType` -> skipped with explicit diagnostic (`unsupported order type`)

## Field mapping table

| Source column | Acquire mapping (`LIMIT_BUY`) | Dispose mapping (`LIMIT_SELL`) | Notes |
| --- | --- | --- | --- |
| `Uuid` | `id = "{Uuid}:acquire"` | `id = "{Uuid}:dispose"` | Stable, reproducible event ID |
| `Exchange` | `asset = split(Exchange, "-")[1]` | `asset = split(Exchange, "-")[1]` | Uses trading currency from `PAYMENT-TRADING` |
| `Quantity` | `qty = Quantity` | `qty = Quantity` | Decimal as-is |
| `Price` | used in cost formula | used in proceeds formula | Treated as JPY amount in this phase-1 fixture |
| `Commission` | used in cost formula | used in proceeds formula | Decimal as-is |
| `Closed` | `ts = Closed` | `ts = Closed` | Parsed with `%m/%d/%Y %I:%M:%S %p`, assumed UTC |

## JPY formulas

- Acquire (`LIMIT_BUY`): `jpy_cost = Price + Commission`
- Dispose (`LIMIT_SELL`): `jpy_proceeds = Price - Commission`

## Known limits (intentionally phase-1)

- No transfer/income rows yet
- No FX conversion in normalizer (fixture uses `JPY-*` pair)
- CSV parser is intentionally minimal and strict for required columns
