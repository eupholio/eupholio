# 23. bitFlyer API対応 PoC 仕様（phase-1）

## 目的

CSVエクスポート前提の運用に加えて、bitFlyer APIから直接データ取得し、既存ノーマライザーへ流せる土台を作る。

本PoCは「まず動く最小構成」を重視し、対象は**約定履歴（現物）**に限定する。

---

## スコープ（phase-1）

### 対象
- 取引所: bitFlyer
- データ種別: 約定履歴（Trade / Execution）
- マーケット: Spot（例: `BTC_JPY`）

### 非対象（phase-2以降）
- 入出金履歴
- CFD/先物系
- 建玉・ポジション情報
- 複数取引所の同時統合

---

## 参照API（候補）

> エンドポイント基底: `https://api.bitflyer.com/v1/`

Private API（要認証）から約定系を取得する。
実際の使用APIは実装時に最終確認するが、PoCでは以下方針:

- 約定履歴取得（product_code / count / before / after でページング）
- 必要に応じて注文情報APIを補助的に参照

※ Public API（ticker/board）は本PoCでは使用しない。

---

## 認証仕様（必須）

bitFlyer Private API認証ヘッダ:

- `ACCESS-KEY`
- `ACCESS-TIMESTAMP` (Unix timestamp)
- `ACCESS-SIGN`

署名:
- `ACCESS-SIGN = HMAC-SHA256(secret, timestamp + method + path + body)`

要件:
- API key/secret は環境変数から読む
- ログに secret を絶対に出さない
- 署名失敗・401系は明確にエラー分類する

---

## データ変換方針

APIレスポンスを**中間フォーマット**へ正規化し、既存 `bitflyer.rs` のCSV行相当へマップする。

### 中間フォーマット（PoC）
- `executed_at`
- `side` (BUY/SELL)
- `base_asset`
- `base_qty`
- `quote_asset` (JPY想定)
- `quote_qty`
- `fee_jpy`（取得可能なら）
- `order_id`（または execution_id から合成ID）

### 既存ロジック接続
最終的には既存の `Event::Acquire / Event::Dispose` マッピングに接続する。

---

## ページング・取得戦略

- `count` は固定上限（例: 100）
- `before/after` で過去分を辿る
- 無限ループ防止で最大ページ数を設定
- APIレート制限を考慮し、短いsleepまたはリトライ間隔を入れる

---

## エラー設計（PoC）

### Hard error
- 認証失敗
- APIレスポンス形式不正
- 必須フィールド欠損
- JPY建て以外（phase-1では非対応として失敗）

### Diagnostic（継続可能）
- side不正
- 未対応フィールド構成
- 1行単位での変換不能

---

## CLI想定（PoC）

実行イメージ:

```bash
cargo run -p eupholio-normalizer -- \
  --source bitflyer-api \
  --product-code BTC_JPY \
  --since 2025-01-01T00:00:00Z \
  --until 2025-12-31T23:59:59Z
```

環境変数:
- `BITFLYER_API_KEY`
- `BITFLYER_API_SECRET`

---

## セキュリティ・運用ルール

- 秘密情報は環境変数のみ（ファイル平文保存禁止）
- ログに署名素材（timestamp+path+body）を出しすぎない
- リトライは指数バックオフ（最大回数制限あり）

---

## 受け入れ条件（phase-1 Done）

- 指定期間の約定履歴をAPIから取得できる
- 既存bitFlyer正規化イベント（Acquire/Dispose）に変換できる
- 最低1つの固定fixtureで再現テストが通る
- 認証失敗/レート制限/フォーマット異常の主要エラーパスがテストされる

---

## 次フェーズ候補

1. 入出金API対応
2. JPY以外のクオート資産対応
3. Coincheck APIアダプタへ横展開
4. API取得結果のスナップショットテスト自動化
