# 16. Normalizer fixture policy

Policy for maintaining `eupholio-normalizer/tests/fixtures` as the adapter set grows.

## Goals

- Make regressions reproducible and easy to debug.
- Protect privacy and sensitive trading information.
- Keep fixtures small, reviewable, and deterministic.

## Directory structure

- `tests/fixtures/<source>/...`
  - raw source sample (`*.csv` / `*.json`)
  - expected normalized events (`*.normalized.json`)
  - optional notes (`README.md`) for source-specific caveats

Example:

- `tests/fixtures/normalizer/bittrex_order_history_smoke.csv`
- `tests/fixtures/normalizer/bittrex_order_history_smoke.normalized.json`

## Naming convention

`<source>_<scenario>.<ext>`

- `source`: exchange/export identifier (`bittrex_order_history`, etc.)
- `scenario`: concise behavior target (`smoke`, `partial_fill`, `quoted_amount`, `missing_column`)

## Privacy / anonymization rules

Never commit raw personal exports as-is.

Before adding fixture data:

1. Replace user-identifying IDs with stable placeholders.
2. Replace asset symbols if needed (keep structural meaning).
3. Shift timestamps consistently (preserve ordering and day boundaries when relevant).
4. Scale or perturb amounts where possible while preserving edge behavior.
5. Remove unused columns that could leak account metadata.

## What to include

Prefer minimal rows needed to prove one behavior.

Good fixture targets:

- happy path buy/sell
- unsupported order type diagnosis
- quoted thousands-separator values
- missing required column/field
- malformed datetime/decimal
- duplicate source IDs (if applicable)

## Test expectations

Each fixture-backed test should assert at least one of:

- normalized events exactly match expected JSON
- diagnostics are explicit (no silent drop)
- normalized events can flow into `eupholio_core::calculate`
- stable key report fields for smoke/e2e scenarios

## PR checklist for fixture changes

- [ ] scenario purpose is clear from filename
- [ ] anonymization applied and reviewed
- [ ] expected normalized output updated
- [ ] regression test added/updated
- [ ] `cargo test` passes at workspace root

## Maintenance loop

When production parsing fails:

1. reduce to minimal reproducible sample
2. anonymize
3. add fixture + failing test
4. implement fix
5. keep fixture as permanent regression guard
