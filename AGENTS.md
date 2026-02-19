# AGENTS.md (eupholio contributors)

Scope: this file applies to the whole repository.

## Branch workflow (core-rs track)
- Base branch for Rust core work: `core-rs`.
- Create feature branches from `core-rs` (`feat/...`, `fix/...`, `docs/...`).
- Keep PRs small and focused (logic vs docs vs fixtures/scripts when possible).
- Before opening/updating a PR, rebase on latest `core-rs` and rerun required checks.

## Mandatory checks before PR
Run from repo root unless noted:
- `go test ./...`
- `go test ./test/integration/...`
- `scripts/check_validation_codes.py`
- `cd eupholio-core && cargo test --all-targets`
- `scripts/compare_go_rust.py`

## Rounding support matrix (method × timing)
| method \ timing | report_only | per_event | per_year |
|---|---:|---:|---:|
| `moving_average` | ✅ | ✅ | ❌ (validation error: `ROUNDING_PER_YEAR_UNSUPPORTED_FOR_MOVING_AVERAGE`) |
| `total_average`  | ✅ | ✅ | ✅ |

## Validation code ↔ docs sync rules
- If you add/remove/rename a validation code in `eupholio-core/src/bin/eupholio-core-cli.rs`, update `eupholio-core/doc/09-validation-codes.md` in the same PR.
- `scripts/check_validation_codes.py` must pass; treat failures as blocking.
- Keep code strings stable once released; if unavoidable, document migration impact in PR notes.

## Docs to update when behavior changes
Update the relevant docs in the same PR:
- CLI behavior/input/output: `eupholio-core/doc/04-cli.md`
- Rounding semantics/timing: `eupholio-core/doc/07-rounding-policy.md`
- Validation issue codes: `eupholio-core/doc/09-validation-codes.md`
- Overview/status: `eupholio-core/README.md`
- If docs list changes, refresh index links in `eupholio-core/doc/README.md`

## PR checklist
- [ ] Scope is focused and branch targets `core-rs`
- [ ] Mandatory checks all pass locally
- [ ] Fixtures/golden tests updated for behavior changes
- [ ] Go↔Rust parity checked (`scripts/compare_go_rust.py`)
- [ ] Validation codes and docs synchronized
- [ ] Related docs updated (CLI/rounding/validation/README)
- [ ] PR description explains user-visible changes and any incompatibilities
