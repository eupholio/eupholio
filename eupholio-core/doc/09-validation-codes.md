# Validation Codes

`eupholio-core-cli validate` が返す `issues[].code` 一覧。

## Error codes

- `INVALID_METHOD`
  - `method` が `moving_average` / `total_average` 以外
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
- `DUPLICATE_EVENT_ID`
  - イベントID重複
- `ACQUIRE_QTY_NON_POSITIVE`
  - Acquireのqtyが0以下
- `ACQUIRE_COST_NEGATIVE`
  - Acquireのjpy_costが負
- `DISPOSE_QTY_NON_POSITIVE`
  - Disposeのqtyが0以下
- `DISPOSE_PROCEEDS_NEGATIVE`
  - Disposeのjpy_proceedsが負
- `INCOME_QTY_NON_POSITIVE`
  - Incomeのqtyが0以下
- `INCOME_VALUE_NEGATIVE`
  - Incomeのjpy_valueが負
- `TRANSFER_QTY_NON_POSITIVE`
  - Transferのqtyが0以下

## Warning codes

- `EMPTY_EVENTS`
  - eventsが空
- `UNUSUAL_TAX_YEAR`
  - tax_yearが通常範囲外
- `CARRY_IN_IGNORED_FOR_MOVING`
  - moving_averageでcarry_in指定（現在無視）
- `CARRY_IN_COST_WITH_ZERO_QTY`
  - qty=0でcost>0
- `EVENT_YEAR_MISMATCH`
  - `event.ts` の年が `tax_year` と不一致

## 運用指針

- `ok=false` の場合は計算実行しない
- warningのみなら実行可能だが、ログとして保存
- 自動処理では `code` をキーに分岐し、`message` は人向け表示に使う
