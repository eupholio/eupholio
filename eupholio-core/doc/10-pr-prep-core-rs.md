# PR Prep (core-rs -> main)

このブランチは `core-rs`。Rustコア再構築を段階的に追加した。

## 変更サマリ

### Core implementation
- `moving_average` / `total_average` 実装
- carry-in 対応 (`calculate_total_average_with_carry_and_rounding`)
- report-only 丸め適用

### CLI
- サブコマンド化 (`calc`, `validate`, `version`)
- `rounding` 外部注入対応
- `validate` の issue code 体系化（enum一元管理）
- `EVENT_YEAR_MISMATCH` warning 追加

### Quality
- golden tests
- CLI e2e tests
- Go/Rust parity scriptとfixture拡張
- GitHub Actions workflow (`rust-core.yml`)

### Docs
- `doc/01`〜`doc/10` まで体系化
- rounding policy / validation codes / normalizer interface を追加

## PR本文テンプレ（草案）

### What
- Add Rust-based `eupholio-core` with switchable cost methods and carry-in support.
- Introduce subcommand-based CLI and structured validation issue codes.
- Add parity and CI scaffolding.

### Why
- Isolate accounting engine from exchange-specific ingestion.
- Prepare for wasm/local execution and safer migration from Go.

### Notes
- `per_event` rounding timing is implemented (golden/parity fixtures added for差分固定).
- `per_year` rounding timing is still planned / partially implemented.
- `report_only` is default.

### Verification
- `cargo test` passed
- `scripts/compare_go_rust.py` passed on fixture set

## 推奨マージ方針
- squash merge 推奨（履歴が多いため）
- PRタイトル案: `feat: introduce rust eupholio-core with cli/validation/parity foundation`
