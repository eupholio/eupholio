# Copilot Review Instructions (eupholio)

Focus review quality on accounting-core safety, deterministic behavior, and CI guardrails.

## Priority review areas

1. **Rounding behavior correctness**
   - `report_only`, `per_event`, `per_year` semantics must remain explicit and deterministic.
   - `moving_average + per_year` is intentionally unsupported and should not silently degrade.

2. **Validation code stability**
   - Validation codes are treated as stable machine-facing identifiers.
   - Any add/remove/rename must be reflected in `eupholio-core/doc/09-validation-codes.md`.

3. **Parity and reproducibility**
   - Go/Rust parity scripts and fixtures must remain CI-portable (no machine-local hardcoded paths).
   - Prefer deterministic fixture inputs and explicit expectations.

4. **Year handling policy**
   - Out-of-tax-year events should be handled consistently across methods.

## What to deprioritize

- Pure style suggestions that do not improve safety, correctness, portability, or maintainability.
- Suggestions that introduce broad semantic changes without tests.

## Expected evidence in suggestions

When suggesting logic changes, include at least one of:
- affected files/functions
- concrete failure mode
- test/fixture to add or update

## Project-specific references

- `eupholio-core/doc/04-cli.md`
- `eupholio-core/doc/07-rounding-policy.md`
- `eupholio-core/doc/09-validation-codes.md`
- `scripts/compare_go_rust.py`
- `scripts/check_validation_codes.py`
