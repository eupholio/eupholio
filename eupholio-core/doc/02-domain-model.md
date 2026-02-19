# Domain Model

## Config

- `method`: `MovingAverage` or `TotalAverage`
- `tax_year`: 対象年

## Event

- `Acquire { asset, qty, jpy_cost, ts }`
- `Dispose { asset, qty, jpy_proceeds, ts }`
- `Income { asset, qty, jpy_value, ts }`
- `Transfer { asset, qty, direction, ts }`

補足:
- `id` を持ち、重複イベント検知に利用
- すべて正規化済み前提（JPY値が確定している）

## Report

- `positions`: 資産ごとの残数量・平均単価
- `realized_pnl_jpy`: 実現損益合計
- `income_jpy`: 所得相当（Income由来）
- `yearly_summary`: 総平均時の年次サマリ
- `diagnostics`: 警告

## Warning

- `DuplicateEventId`
- `NegativePosition`
- `YearMismatch`（`event.ts` の年が `tax_year` と不一致。警告を出しつつ当該イベントは計算から除外）
- `YearBoundaryCarry`
