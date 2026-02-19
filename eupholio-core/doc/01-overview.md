# Overview

## Purpose

`eupholio-core` is a **JPY-based profit/loss calculation engine** for crypto assets.

- It does not provide server features or database access.
- It does not fetch market prices or perform FX conversion.
- It calculates profit/loss and positions from normalized event inputs.

## Scope

- `CostMethod` switching
  - `MovingAverage`
  - `TotalAverage`
- Aggregation by tax year (`tax_year`)
- Total-average calculation using carry-in balances
- JSON I/O via CLI

## Out of Scope

- Exchange API integration
- Exchange-specific CSV parsing
- Price-feed APIs
- UI / DB

## Design Principles

1. The core focuses only on ledger calculations.
2. Inputs must already be normalized (including JPY values).
3. No floating-point arithmetic (`rust_decimal` is used).
4. Engine switching is handled with enum branching (minimizing trait dependency).
