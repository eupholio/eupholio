# 25. AI normalizer adapter playbook

This document defines the minimum documentation/data package needed to implement a
new exchange adapter with high confidence using AI-assisted development.

Goal:
- Make adapter development repeatable
- Reduce ambiguous mapping decisions
- Prefer `diagnostic/unsupported` over silent wrong conversion

## 1) Required inputs (minimum viable package)

### A. Source schema spec (`spec/schema.md`)

For each input field, define:
- field name (exact)
- type (`string|decimal|datetime|enum|...`)
- required/optional
- normalization rule (trim, uppercase, locale decimal, timezone)
- semantic meaning

Also define:
- row identity strategy (dedup key / id components)
- allowed timestamp formats + timezone assumptions
- value domain constraints (e.g. volume > 0)

### B. Mapping rules (`spec/mapping.md`)

Action-by-action mapping table:
- predicate (how to detect row type)
- target `Event` variant
- field transforms
- validation rules
- unsupported conditions (diagnostic)
- hard error conditions (`Err`)

Rule priority must be explicit when predicates can overlap.

### C. Golden fixtures (`fixtures/golden/*.json` + raw sample CSV)

A golden fixture should include:
- input row(s)
- expected events (full payload)
- expected diagnostics (if any)
- expected hard error (if any)
- rationale (short)

Coverage categories (must include all):
- happy path (major actions)
- boundary values (0, tiny decimals, large values)
- invalid fee/price/volume combinations
- unsupported action/currency/path
- malformed date / malformed decimal

## 2) Behavioral contract (recommended defaults)

Unless exchange-specific docs require otherwise:
- unknown action => diagnostic (not panic)
- ambiguous mapping => diagnostic
- impossible numeric constraints => hard error
- preserve original row context in diagnostic reason (sanitized)

## 3) File template (copy/paste starter)

```text
new-adapter/
  spec/
    schema.md
    mapping.md
  fixtures/
    raw/
      sample-01.csv
      sample-02.csv
    golden/
      case-001-buy-jpy.json
      case-002-sell-base-fee.json
      case-003-unsupported-action.json
```

### `schema.md` minimal template

```md
# <Exchange> schema

## Timestamp
- type: datetime
- format: %Y/%m/%d %H:%M:%S
- timezone: Asia/Tokyo
- required: yes
- note: source exports local time

## Action
- type: enum
- values: BUY, SELL, ...
- required: yes
```

### `mapping.md` minimal template

```md
# <Exchange> mapping

## Rule 1: BUY
- predicate: Action == BUY
- event: Acquire
- transforms:
  - asset = Base.upper()
  - qty = Volume
  - jpy_cost = Price * Volume + Fee(if fee_ccy==JPY)
- validations:
  - volume > 0
  - price > 0
- diagnostics:
  - fee_ccy not in {JPY, Base}
- errors:
  - volume <= 0
  - price <= 0
```

## 4) AI implementation workflow

1. Freeze `schema.md` and `mapping.md`
2. Generate/curate golden fixtures
3. Ask AI to implement mapper from mapping rules only
4. Ask AI to generate tests from golden fixtures
5. Run CI and manually review diagnostics wording quality
6. Add regression fixture for every production bug

## 5) Acceptance checklist

- [ ] Every supported action has at least 1 happy-path fixture
- [ ] Every supported action has at least 1 invalid-path fixture
- [ ] Unknown action test exists
- [ ] Non-supported currency/path test exists
- [ ] Date parse failure test exists
- [ ] Decimal parse failure test exists
- [ ] Id uniqueness/row identity test exists
- [ ] Diagnostic messages include enough context (sanitized)

## 6) Notes for full automation

"Fully AI" is feasible only when the above package is maintained as source-of-truth.
When schema drift is detected (new column/action/value), block auto-merge and require
human review + fixture update.
