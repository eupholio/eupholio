# Normalizer Interface (Draft)

Objective: absorb exchange-specific differences outside `eupholio-core`, and pass only normalized Events into the core.

## Layer Separation

- Normalizer layer: CSV/API input -> normalized Event
- Core layer: Event -> PnL calculation

## Normalizer Responsibilities

1. Parse exchange-specific formats
2. Determine JPY values including fees / net of fees
3. Split crypto-crypto transactions into Acquire + Dispose
4. Ensure Event ID uniqueness
5. Normalize timezone (UTC)

## Contract Passed to Core

- `Event::Acquire` uses `jpy_cost` (fee-inclusive)
- `Event::Dispose` uses `jpy_proceeds` (fee-deducted)
- `Event::Income` uses `jpy_value`
- `Transfer` carries movement data that must not generate PnL

## Recommended I/F (Concept)

```rust
trait Normalizer {
    fn normalize(&self, raw: &[u8]) -> Result<Vec<Event>, NormalizeError>;
}
```

In actual implementation, this does not need to be fixed as a trait; CLI/adapter-level implementations are also acceptable.

## Target Exchanges (Expected)

- bitFlyer
- Coinbase
- SBI VC Trade

## Validation

Normalizer outputs must always pass through `eupholio-core-cli validate`.
