# Overview

## 目的

`eupholio-core` は、暗号資産の **JPYベース損益計算エンジン** です。

- サーバ機能やDBアクセスは持たない
- 価格取得やFX換算は行わない
- 正規化済みイベント入力から、損益・在庫を計算する

## スコープ

- `CostMethod` 切替
  - `MovingAverage`
  - `TotalAverage`
- 年次（tax_year）を伴う集計
- 繰越（carry-in）を使った総平均計算
- CLIによるJSON入出力

## 非スコープ

- 取引所API接続
- 取引所CSV個別パース
- 価格取得API
- UI / DB

## 設計原則

1. コアは帳簿計算のみ
2. 入力は正規化済み（JPY値を含む）
3. 浮動小数点は使わない（`rust_decimal`）
4. エンジン切替は enum 分岐で行う（trait依存を最小化）
