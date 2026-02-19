# Normalizer Interface (Draft)

目的: 取引所ごとの差分を `eupholio-core` の外側で吸収し、コアには正規化済みEventのみを渡す。

## レイヤ分離

- Normalizer層: CSV/API入力 -> 正規化Event
- Core層: Event -> 損益計算

## Normalizerの責務

1. 取引所固有フォーマットの解析
2. 手数料込み/差引後JPY値の確定
3. crypto-cryptoを Acquire + Dispose へ分解
4. Event ID の一意化
5. タイムゾーン正規化（UTC）

## Coreへ渡す契約

- `Event::Acquire` は `jpy_cost`（手数料込み）
- `Event::Dispose` は `jpy_proceeds`（手数料差引後）
- `Event::Income` は `jpy_value`
- `Transfer` は損益を発生させない移動情報

## 推奨I/F（概念）

```rust
trait Normalizer {
    fn normalize(&self, raw: &[u8]) -> Result<Vec<Event>, NormalizeError>;
}
```

実装では trait 固定にせず、CLI/アダプタ単位でも可。

## 対象取引所（想定）

- bitFlyer
- Coinbase
- SBI VC Trade

## バリデーション

Normalizer出力は `eupholio-core-cli validate` を必ず通す。
