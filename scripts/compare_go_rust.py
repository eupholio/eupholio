#!/usr/bin/env python3
import json, subprocess, pathlib
from decimal import Decimal

ROOT = pathlib.Path(__file__).resolve().parents[1]
GO = "/home/kinakao/.local/go/bin/go"


def rust_input(fixture, method):
    events = []
    for i, e in enumerate(fixture["events"], 1):
        base = {"id": f"e{i}", "asset": e["asset"], "qty": e["qty"], "ts": "2026-01-01T00:00:00Z"}
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
    return out


def run_go(fixture):
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
        check=True,
    )
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


def compare_case(path):
    fixture = json.loads(path.read_text())
    go_outs = run_go(fixture)
    rust_m = run_rust(rust_input(fixture, "moving_average"))
    rust_t = run_rust(rust_input(fixture, "total_average"))

    go_m = pick(go_outs, "moving_average")
    go_t = pick(go_outs, "total_average")

    print(f"== {path.name} ==")
    print("moving: go=", go_m["realized_pnl_jpy"], " rust=", rust_m["realized_pnl_jpy"])
    print("total : go=", go_t["realized_pnl_jpy"], " rust=", rust_t["realized_pnl_jpy"])

    check_moving = fixture.get("check_moving", True)
    check_total = fixture.get("check_total", True)

    moving_equal = approx_equal(jp_report_round(go_m["realized_pnl_jpy"]), rust_m["realized_pnl_jpy"])
    total_equal = approx_equal(jp_report_round(go_t["realized_pnl_jpy"]), rust_t["realized_pnl_jpy"])

    return {
        "case": path.name,
        "moving_equal": moving_equal if check_moving else None,
        "total_equal": total_equal if check_total else None,
    }


if __name__ == "__main__":
    cases = [
        ROOT / "scripts/parity_fixture_case1.json",
        ROOT / "scripts/parity_fixture_case3.json",
        ROOT / "scripts/parity_fixture_transfer.json",
        ROOT / "scripts/parity_fixture_fractional.json",
        ROOT / "scripts/parity_fixture_carry_in.json",
    ]
    results = [compare_case(c) for c in cases]
    print("summary:", json.dumps(results, ensure_ascii=False))
