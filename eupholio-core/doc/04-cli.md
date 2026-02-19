# CLI

バイナリ:
- `eupholio-core-cli`

標準入力JSONを読み、計算結果 `Report` JSON を標準出力へ返します。

## サブコマンド

- `calc` : JSON入力を計算してReportを出力
- `validate` : JSON入力を検証（errorがあれば非0終了）
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

## rounding.timing 差分例（同一最小入力で比較）

以下は `tests/fixtures/per_year_total_difference.json` と同じ入力（`method=total_average`）を使い、`timing` だけを切り替えた比較です。

### 共通入力

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

`timing` を `report_only` / `per_event` / `per_year` に変更して `calc` を実行。

### 期待差分（抜粋）

| timing | realized_pnl_jpy (report) | yearly_summary.by_asset.BTC.realized_pnl_jpy | yearly_summary.by_asset.ETH.realized_pnl_jpy |
|---|---:|---:|---:|
| `report_only` | `1` | `0` | `0` |
| `per_event` | `0` | `0` | `0` |
| `per_year` | `0` | `0` | `0` |

補足:
- `report_only=1`, `per_year=0` は fixture (`per_year_total_difference.json`) で固定済み。
- 同一入力で `per_event` も `0` となり、このケースでは `per_year` と同値。

## validate の timing warning 挙動（現状）

現行CLIでは、`rounding.timing=per_event` / `per_year` を指定しても
`ROUNDING_TIMING_NOT_FULLY_IMPLEMENTED` は返しません（warningなし）。

- 回帰テスト: `tests/cli_e2e.rs`
  - `cli_validate_per_event_no_timing_warning`
  - `cli_validate_per_year_no_timing_warning`

一方で、`validate` は timing 以外の warning/error（例: `EVENT_YEAR_MISMATCH`, `DUPLICATE_EVENT_ID`）は通常どおり返します。
