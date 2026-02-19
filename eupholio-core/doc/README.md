# eupholio-core docs

このディレクトリは `eupholio-core`（Rust計算コア）の設計・運用ドキュメント置き場です。

## 目次

- [01-overview.md](./01-overview.md)
  - 目的・スコープ・非スコープ
- [02-domain-model.md](./02-domain-model.md)
  - Event/Config/Report の仕様
- [03-calculation-methods.md](./03-calculation-methods.md)
  - MovingAverage / TotalAverage の計算詳細
- [04-cli.md](./04-cli.md)
  - CLIのJSON入出力、実行例
- [05-parity-testing.md](./05-parity-testing.md)
  - Go版とのパリティ検証方法
- [06-roadmap.md](./06-roadmap.md)
  - 今後の拡張計画
- [07-rounding-policy.md](./07-rounding-policy.md)
  - 外部注入可能な丸めルール方針
- [08-normalizer-interface.md](./08-normalizer-interface.md)
  - 取引所入力をEventへ正規化するI/F草案
- [09-validation-codes.md](./09-validation-codes.md)
  - validate が返す issue code 一覧
- [10-pr-prep-core-rs.md](./10-pr-prep-core-rs.md)
  - core-rs から main へのPR準備ノート
