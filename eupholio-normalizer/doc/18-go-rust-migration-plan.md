# 18. Go -> Rust migration plan (normalizer + core path)

This plan defines a safe, staged replacement of legacy Go paths with Rust-based adapters + core.

## Principles

- No big-bang replacement.
- Keep rollback path at every stage.
- Compare outputs before switching defaults.

## Stages

### Stage 0: Foundations (done/in progress)

- `eupholio-core` stabilized (calculation + tests + CI)
- `eupholio-normalizer` separated as dedicated crate
- phase-1 Bittrex adapter + fixtures + smoke tests

### Stage 1: Parallel adapter expansion

- Add next live-source adapter (target: coincheck)
- For each adapter:
  - mapping doc
  - fixture pair (raw + normalized)
  - smoke test into `calculate`
  - diagnostics for unsupported rows

### Stage 2: Dual-run comparison

- For selected sources, run:
  - legacy Go path output
  - Rust normalizer + core output
- Diff key fields:
  - realized pnl
  - carry summary
  - diagnostics counts
- Record known acceptable deltas (if any)

### Stage 3: Controlled default switch

- Enable Rust path behind per-source switch/flag
- Roll out source-by-source
- Keep Go fallback available during soak period

### Stage 4: Go path retirement

- Remove source paths with stable Rust replacement
- Keep historical fixtures and comparison docs
- Simplify CI accordingly

## Rollback rules

Immediate rollback to Go path if:

- mismatch beyond accepted tolerance
- unhandled row-type spikes
- performance regression beyond agreed threshold

## Definition of done (per source)

- mapping doc approved
- fixture coverage includes happy + failure path
- dual-run diff reviewed
- switch enabled with no critical alerts during soak
