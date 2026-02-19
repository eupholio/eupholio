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
    return {"method": method, "tax_year": fixture["tax_year"], "events": events}


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

    return {
        "case": path.name,
        "moving_equal": Decimal(str(go_m["realized_pnl_jpy"])) == Decimal(str(rust_m["realized_pnl_jpy"])),
        "total_equal": Decimal(str(go_t["realized_pnl_jpy"])) == Decimal(str(rust_t["realized_pnl_jpy"])),
    }


if __name__ == "__main__":
    cases = [
        ROOT / "scripts/parity_fixture_case1.json",
        ROOT / "scripts/parity_fixture_case3.json",
    ]
    results = [compare_case(c) for c in cases]
    print("summary:", json.dumps(results, ensure_ascii=False))
