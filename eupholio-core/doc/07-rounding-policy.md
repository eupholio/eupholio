# Rounding Policy (External Config)

`eupholio-core` では、丸めルールをコアに焼き込まず、外部設定で注入できる形を採用する。

## 目的

- 国・税制・運用差分へ対応
- 互換モード（既存実装の丸め挙動）を切替可能にする
- 内部計算精度を落とさず、最終出力で調整する

## 日本向けデフォルト案

- 実現損益（JPY）: scale=0, mode=half_up
- 所得（JPY）: scale=0, mode=half_up
- 平均単価: scale=8, mode=half_up
- 数量: scale=8, mode=half_up
- timing: report_only

## 設定JSON例

```json
{
  "jurisdiction": "JP",
  "rounding": {
    "currency": {
      "JPY": { "scale": 0, "mode": "half_up" }
    },
    "unit_price": { "scale": 8, "mode": "half_up" },
    "quantity": { "scale": 8, "mode": "half_up" },
    "timing": "report_only"
  }
}
```

## Timing

- `report_only`: 中間計算は丸めず、出力時のみ丸め
- `per_event`: イベント処理ごとに丸め
- `per_year`: 年次締め処理で丸め

## 実装方針（段階導入）

1. 設定構造体を導入（未指定時はJPデフォルト）
2. Report出力レイヤーで丸め適用（`report_only`）
3. 互換用途で `per_event` / `per_year` を追加

## 実装タスク（次フェーズ）

### per_event

- 目的: 各イベント適用後に丸めを実施
- 実装済み:
  - 同一入力で `report_only` と異なる結果を再現できるfixtureを追加
  - `validate` で `per_event` 指定時の `ROUNDING_TIMING_NOT_FULLY_IMPLEMENTED` warning は出さない
  - Golden testに `per_event` ケースを2件追加

### per_year

- 目的: 年次集計終了時に丸めを実施
- 受け入れ基準:
  - TotalAverage の `yearly_summary` に対し年次丸めを一括適用
  - carry-in を含むケースで期待値を固定
  - `compare_go_rust.py` で `per_year` fixture比較が可能
