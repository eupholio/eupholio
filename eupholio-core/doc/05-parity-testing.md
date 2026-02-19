# Go parity testing

Go実装 (`mam`/`wam`) と Rust実装の損益を比較するスクリプトを用意しています。

## 前提

- Go toolchain
- Rust toolchain

## 実行

```bash
cd eupholio
scripts/compare_go_rust.py
```

## 現在のfixture

- `parity_fixture_case1.json`（基本売買）
- `parity_fixture_case3.json`（crypto-crypto分解）
- `parity_fixture_transfer.json`（Transfer混在）
- `parity_fixture_fractional.json`（小数精度）
- `parity_fixture_carry_in.json`（年跨ぎ繰越、総平均）
- `parity_fixture_per_event_moving.json`（per_event 丸め差分の可視化: moving）
- `parity_fixture_per_event_total.json`（per_event 丸め差分の可視化: total）

## 判定

- 損益は `Decimal` 比較（微小差許容あり）
- caseごとに `check_moving` / `check_total` を切替可能
- `parity_fixture_per_event_*` は Go parity の合否対象ではなく、Rust 側で `report_only` と `per_event` の差分を固定化するための fixture
