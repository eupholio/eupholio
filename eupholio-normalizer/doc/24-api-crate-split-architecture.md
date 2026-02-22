# 24. APIクレート分離アーキテクチャ案

## 方針

`eupholio-normalizer` は純粋変換ロジックに集中し、
取引所APIアクセス（認証/署名/リトライ/レート制限）は別クレートへ分離する。

---

## 想定クレート構成

- `eupholio-normalizer`
  - 入力DTO → 正規化イベント
  - API依存なし

- `eupholio-exchange-api`（新規）
  - 取引所APIクライアント（phase-1は bitFlyer）
  - 認証ヘッダ生成（HMAC-SHA256）
  - ページング/リトライ/エラー分類

- `eupholio-cli`（既存/将来）
  - API取得 + normalizer実行を接続

---

## 責務境界

### eupholio-exchange-api
- `BitflyerClient`:
  - `list_executions(...) -> Vec<ExecutionRecord>`
- `Signer`:
  - `sign(timestamp, method, path, body) -> hex`
- `RateLimitPolicy`:
  - 連打抑制
- `ApiError`:
  - Auth / RateLimited / Transport / Decode / Unexpected

### eupholio-normalizer
- `normalize_bitflyer_records(records)`
- API都合の情報を受け取らない（中間DTO化して吸収）

---

## WASM前提の設計ルール

- secretをWASMに置かない
- 署名処理はサーバー側（BFF/Backend）責務
- WASMは「取得済みデータの正規化」に専念

---

## phase-1実装ステップ

1. `eupholio-exchange-api` 新規作成（bitFlyer executionsのみ）
2. `ExecutionRecord` DTO定義
3. `eupholio-normalizer` 側にDTO入力関数を追加
4. CLIで end-to-end 実行可能にする
5. fixture + モックでテスト整備

---

## Done条件

- API層と正規化層がcrate境界で分離されている
- normalizer単体テストがAPI無しで完結
- APIクレートは秘密情報マスク/エラー分類を実装
- 将来の取引所追加が `eupholio-exchange-api` 内拡張で対応可能
