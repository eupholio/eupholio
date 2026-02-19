#!/usr/bin/env python3
import json
import os
import pathlib
import shutil
import subprocess
import sys
from decimal import Decimal

ROOT = pathlib.Path(__file__).resolve().parents[1]
GO = os.environ.get("GO_BIN") or shutil.which("go")


def rust_input(fixture, method):
    events = []
    for i, e in enumerate(fixture["events"], 1):
        base = {
            "id": e.get("id", f"e{i}"),
            "asset": e["asset"],
            "qty": e["qty"],
            "ts": e.get("ts", "2026-01-01T00:00:00Z"),
        }
        if e["type"] == "Acquire":
            base["type"] = "Acquire"
            base["jpy_cost"] = e["jpy_cost"]
        elif e["type"] == "Dispose":
            base["type"] = "Dispose"
            base["jpy_proceeds"] = e["jpy_proceeds"]
        elif e["type"] == "Income":
            base["type"] = "Income"
            base["jpy_value"] = e["jpy_value"]
        else:
            continue
        events.append(base)
    out = {"method": method, "tax_year": fixture["tax_year"], "events": events}
    if method == "total_average" and fixture.get("carry_in"):
        out["carry_in"] = fixture["carry_in"]
    if fixture.get("rounding"):
        out["rounding"] = fixture["rounding"]
    return out


def run_go(fixture):
    if not GO:
        raise RuntimeError("go binary not found (set GO_BIN or install Go in PATH)")
    p = subprocess.run(
        [GO, "run", str(ROOT / "scripts/go_cost_compare.go")],
        input=json.dumps(fixture).encode(),
        cwd=ROOT,
        capture_output=True,
        check=True,
    )
    return json.loads(p.stdout)


def run_rust(inp):
    cmd = '. $HOME/.cargo/env && cargo run --quiet --bin eupholio-core-cli'
    p = subprocess.run(
        ["bash", "-lc", cmd],
        input=json.dumps(inp).encode(),
        cwd=ROOT / "eupholio-core",
        capture_output=True,
    )
    if p.returncode != 0:
        return {
            "error": p.stderr.decode().strip() or f"rust cli failed: exit={p.returncode}",
            "exit_code": p.returncode,
        }
    return json.loads(p.stdout)


def approx_equal(a, b, eps=Decimal("0.000000001")):
    return abs(Decimal(str(a)) - Decimal(str(b))) <= eps


def jp_report_round(v):
    return Decimal(str(v)).quantize(Decimal("1"))


def pick(go_outs, method):
    key = "mam" if method == "moving_average" else "wam"
    for o in go_outs:
        if o["method"] == key:
            return o
    raise RuntimeError("method not found")


def check_expectation(expectation, rust_m, rust_t):
    if not expectation:
        return None, None

    moving_expected = expectation.get("moving_realized_pnl_jpy")
    total_expected = expectation.get("total_realized_pnl_jpy")

    moving_ok = (
        approx_equal(rust_m["realized_pnl_jpy"], moving_expected)
        if moving_expected is not None
        else None
    )
    total_ok = (
        approx_equal(rust_t["realized_pnl_jpy"], total_expected)
        if total_expected is not None
        else None
    )

    return moving_ok, total_ok


def compare_case(path):
    fixture = json.loads(path.read_text())
    go_outs = run_go(fixture)

    check_moving = fixture.get("check_moving", True)
    check_total = fixture.get("check_total", True)
    expectation = fixture.get("expectation") or {}

    need_moving = check_moving or expectation.get("moving_realized_pnl_jpy") is not None
    need_total = check_total or expectation.get("total_realized_pnl_jpy") is not None

    rust_m = run_rust(rust_input(fixture, "moving_average")) if need_moving else None
    rust_t = run_rust(rust_input(fixture, "total_average")) if need_total else None

    go_m = pick(go_outs, "moving_average")
    go_t = pick(go_outs, "total_average")

    print(f"== {path.name} ==")
    print(
        "moving: go=",
        go_m["realized_pnl_jpy"],
        " rust=",
        rust_m["realized_pnl_jpy"] if rust_m and "realized_pnl_jpy" in rust_m else "(skipped)",
    )
    print(
        "total : go=",
        go_t["realized_pnl_jpy"],
        " rust=",
        rust_t["realized_pnl_jpy"] if rust_t and "realized_pnl_jpy" in rust_t else "(skipped)",
    )

    moving_error = rust_m.get("error") if rust_m else None
    total_error = rust_t.get("error") if rust_t else None

    moving_equal = None
    if check_moving and rust_m and "realized_pnl_jpy" in rust_m:
        moving_equal = approx_equal(jp_report_round(go_m["realized_pnl_jpy"]), rust_m["realized_pnl_jpy"])

    total_equal = None
    if check_total and rust_t and "realized_pnl_jpy" in rust_t:
        total_equal = approx_equal(jp_report_round(go_t["realized_pnl_jpy"]), rust_t["realized_pnl_jpy"])

    exp_moving_ok = None
    moving_expected = expectation.get("moving_realized_pnl_jpy")
    if moving_expected is not None and rust_m and "realized_pnl_jpy" in rust_m:
        exp_moving_ok = approx_equal(rust_m["realized_pnl_jpy"], moving_expected)

    exp_total_ok = None
    total_expected = expectation.get("total_realized_pnl_jpy")
    if total_expected is not None and rust_t and "realized_pnl_jpy" in rust_t:
        exp_total_ok = approx_equal(rust_t["realized_pnl_jpy"], total_expected)

    return {
        "case": path.name,
        "moving_equal": moving_equal,
        "total_equal": total_equal,
        "moving_expected_ok": exp_moving_ok,
        "total_expected_ok": exp_total_ok,
        "moving_error": moving_error,
        "total_error": total_error,
    }


def has_failures(result):
    checks = [
        result.get("moving_equal"),
        result.get("total_equal"),
        result.get("moving_expected_ok"),
        result.get("total_expected_ok"),
    ]
    errors = [result.get("moving_error"), result.get("total_error")]
    return any(v is False for v in checks) or any(e is not None for e in errors)


if __name__ == "__main__":
    cases = [
        ROOT / "scripts/parity_fixture_case1.json",
        ROOT / "scripts/parity_fixture_case3.json",
        ROOT / "scripts/parity_fixture_transfer.json",
        ROOT / "scripts/parity_fixture_fractional.json",
        ROOT / "scripts/parity_fixture_carry_in.json",
        ROOT / "scripts/parity_fixture_per_event_moving.json",
        ROOT / "scripts/parity_fixture_per_event_total.json",
        ROOT / "scripts/parity_fixture_per_year_total.json",
    ]
    results = [compare_case(c) for c in cases]
    print("summary:", json.dumps(results, ensure_ascii=False))

    if any(has_failures(r) for r in results):
        sys.exit(1)
