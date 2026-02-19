# 14. Normalizer phase-1 bootstrap

Now that `eupholio-core` is release-level, the next milestone is end-to-end integration via a minimal normalizer path.

## Goal (phase-1)

Build one production-like path:

`exchange input (CSV/API)` -> `normalized events` -> `eupholio-core calculate` -> `report JSON`

## Scope

- Exactly **one** exchange/source in phase-1.
- Focus on happy-path records first (acquire/dispose/income/transfer needed for that source).
- Keep core logic untouched; implement conversion outside core.

## Deliverables

1. `normalizer` adapter (source-specific)
2. Event mapping table (source field -> `Event` field)
3. Fixture pair:
   - raw source sample
   - expected normalized events JSON
4. e2e smoke test:
   - run normalizer -> run `calculate` -> assert non-empty report + stable snapshot fields
5. docs update for runbook

## Acceptance criteria

- Mapping is deterministic and documented.
- Unsupported rows are skipped with explicit diagnostics (no silent drop).
- Event IDs are stable/reproducible.
- `cargo test` and existing parity checks remain green.

## Non-goals (phase-1)

- Multi-exchange unification
- full historical backfill complexity
- asset-class-specific tax semantics outside current event model

## Suggested execution order

1. Select source (single exchange/export format)
2. Freeze mapping table in doc
3. Implement adapter + diagnostics
4. Add fixtures and e2e smoke
5. PR with mapping rationale and known limits
