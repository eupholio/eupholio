# 22. BitFlyer normalizer phase-1 trade mapping

This phase introduces a minimal BitFlyer adapter in `eupholio-normalizer`.

## Scope

- Supported trade types:
  - `買い` / `BUY` -> `Event::Acquire`
  - `売り` / `SELL` -> `Event::Dispose`
- Unsupported trade types are emitted as diagnostics.
- Payment asset must be `JPY` for supported trade rows.

## Mapping

Input columns (JP/EN header variants supported):
- Trade Date
- Trade Type
- Currency 1
- Amount (Currency 1)
- Fee
- JPY Rate (Currency 1)
- Currency 2
- Amount (Currency 2)
- Order ID

Event mapping:
- Buy (`買い`/`BUY`)
  - `id = "{OrderID}:acquire"`
  - `asset = Currency1`
  - `qty = Amount(Currency1) + Fee` (fee is negative in sample exports)
  - `jpy_cost = abs(Amount(Currency2)) + abs(Fee) * JPYRate`
- Sell (`売り`/`SELL`)
  - `id = "{OrderID}:dispose"`
  - `asset = Currency1`
  - `qty = Amount(Currency1)`
  - `jpy_proceeds = abs(Amount(Currency2)) - abs(Fee) * JPYRate`

## Validation

Hard errors:
- invalid/missing required headers
- invalid csv row / decimal / datetime
- non-JPY payment asset for supported buy/sell rows

Diagnostics:
- unsupported trade type rows with row number and sanitized details.
