# 17. Next adapter survey (parallel prep)

This note records the next source candidate after phase-1 Bittrex.

## Survey result

Primary candidate: **Coincheck trade history CSV** (`pkg/coincheck/extractor.go`)

Why:

- Existing extractor path already in repo (`pkg/coincheck`)
- Single CSV source shape (simpler than multi-file Poloniex flow)
- Likely active exchange data in current operations
- Good bridge from frozen-reference adapter (Bittrex) to live-source adapter

## Other candidates and trade-offs

- **bitFlyer** (`pkg/bitflyer`)
  - Pros: active source
  - Cons: bilingual/format variation handling appears heavier
- **Poloniex** (`pkg/poloniex/*`)
  - Pros: broad event coverage (trade/deposit/withdraw/distribution)
  - Cons: multi-file ingestion complexity is high for phase-2
- **Cryptact** (`pkg/cryptact`)
  - Pros: aggregator-style format
  - Cons: semantics less exchange-native for adapter contract validation

## Proposed scope for next adapter (coincheck)

- phase-2 minimal events: `Acquire`, `Dispose` first
- explicit diagnostics for unsupported row categories
- deterministic ID scheme documented
- fixture set:
  - `coincheck_*_smoke.csv`
  - `coincheck_*_smoke.normalized.json`
  - at least one malformed/missing-header fixture

## Exit criteria

- source -> normalized parity fixture is stable
- normalized -> `eupholio_core::calculate` smoke path passes
- unsupported rows are diagnosable and non-silent
