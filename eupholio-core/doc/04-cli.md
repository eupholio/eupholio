# CLI

バイナリ:
- `eupholio-core-cli`

標準入力JSONを読み、計算結果 `Report` JSON を標準出力へ返します。

## サブコマンド

- `calc` : JSON入力を計算してReportを出力
- `validate` : JSON入力を検証（不正なら非0終了）
  - `issues[]` に `code` / `level` / `message` を返却
- `version` : CLIバージョン表示

`calc` は互換のため、サブコマンド省略時のデフォルトでも実行されます。

## 実行

```bash
cd eupholio-core
cat input.json | cargo run --quiet --bin eupholio-core-cli -- calc
# 互換: calc 省略でも可
cat input.json | cargo run --quiet --bin eupholio-core-cli

# バリデーションのみ
cat input.json | cargo run --quiet --bin eupholio-core-cli -- validate
```

## 入力（moving_average）

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

## 入力（total_average + carry_in + rounding override）

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
