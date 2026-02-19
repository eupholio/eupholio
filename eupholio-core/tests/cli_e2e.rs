use assert_cmd::Command;
use predicates::prelude::*;

#[test]
fn cli_calc_works() {
    let input = r#"{
      "method":"moving_average",
      "tax_year":2026,
      "events":[
        {"type":"Acquire","id":"a1","asset":"BTC","qty":"1","jpy_cost":"3000000","ts":"2026-01-01T00:00:00Z"},
        {"type":"Dispose","id":"d1","asset":"BTC","qty":"0.5","jpy_proceeds":"2000000","ts":"2026-02-01T00:00:00Z"}
      ]
    }"#;

    let mut cmd = Command::new(assert_cmd::cargo::cargo_bin!("eupholio-core-cli"));
    cmd.arg("calc")
        .write_stdin(input)
        .assert()
        .success()
        .stdout(predicate::str::contains("\"realized_pnl_jpy\""));
}

#[test]
fn cli_validate_ok() {
    let input = r#"{
      "method":"total_average",
      "tax_year":2026,
      "carry_in":{"BTC":{"qty":"2","cost":"8000000"}},
      "events":[
        {"type":"Acquire","id":"a1","asset":"BTC","qty":"1","jpy_cost":"6000000","ts":"2026-01-05T00:00:00Z"}
      ]
    }"#;

    let mut cmd = Command::new(assert_cmd::cargo::cargo_bin!("eupholio-core-cli"));
    cmd.arg("validate")
        .write_stdin(input)
        .assert()
        .success()
        .stdout(predicate::str::contains("\"ok\": true"));
}

#[test]
fn cli_validate_ng_duplicate_id() {
    let input = r#"{
      "method":"moving_average",
      "tax_year":2026,
      "events":[
        {"type":"Acquire","id":"dup","asset":"BTC","qty":"1","jpy_cost":"3000000","ts":"2026-01-01T00:00:00Z"},
        {"type":"Dispose","id":"dup","asset":"BTC","qty":"0.5","jpy_proceeds":"2000000","ts":"2026-02-01T00:00:00Z"}
      ]
    }"#;

    let mut cmd = Command::new(assert_cmd::cargo::cargo_bin!("eupholio-core-cli"));
    cmd.arg("validate")
        .write_stdin(input)
        .assert()
        .failure()
        .stdout(predicate::str::contains("DUPLICATE_EVENT_ID"));
}

#[test]
fn cli_validate_ng_rounding_scale() {
    let input = r#"{
      "method":"moving_average",
      "tax_year":2026,
      "rounding": {
        "currency": {"JPY": {"scale": 99, "mode": "half_up"}},
        "unit_price": {"scale": 8, "mode": "half_up"},
        "quantity": {"scale": 8, "mode": "half_up"},
        "timing": "report_only"
      },
      "events":[
        {"type":"Acquire","id":"a1","asset":"BTC","qty":"1","jpy_cost":"3000000","ts":"2026-01-01T00:00:00Z"}
      ]
    }"#;

    let mut cmd = Command::new(assert_cmd::cargo::cargo_bin!("eupholio-core-cli"));
    cmd.arg("validate")
        .write_stdin(input)
        .assert()
        .failure()
        .stdout(predicate::str::contains("ROUNDING_JPY_SCALE_TOO_LARGE"));
}

#[test]
fn cli_validate_per_year_no_timing_warning() {
    let input = r#"{
      "method":"total_average",
      "tax_year":2026,
      "rounding": {
        "currency": {"JPY": {"scale": 0, "mode": "half_up"}},
        "unit_price": {"scale": 8, "mode": "half_up"},
        "quantity": {"scale": 8, "mode": "half_up"},
        "timing": "per_year"
      },
      "events":[
        {"type":"Acquire","id":"a1","asset":"BTC","qty":"1","jpy_cost":"3000000","ts":"2026-01-01T00:00:00Z"}
      ]
    }"#;

    let mut cmd = Command::new(assert_cmd::cargo::cargo_bin!("eupholio-core-cli"));
    cmd.arg("validate")
        .write_stdin(input)
        .assert()
        .success()
        .stdout(predicate::str::contains("ROUNDING_TIMING_NOT_FULLY_IMPLEMENTED").not());
}

#[test]
fn cli_validate_warn_event_year_mismatch() {
    let input = r#"{
      "method":"moving_average",
      "tax_year":2026,
      "events":[
        {"type":"Acquire","id":"a1","asset":"BTC","qty":"1","jpy_cost":"3000000","ts":"2025-12-31T00:00:00Z"}
      ]
    }"#;

    let mut cmd = Command::new(assert_cmd::cargo::cargo_bin!("eupholio-core-cli"));
    cmd.arg("validate")
        .write_stdin(input)
        .assert()
        .success()
        .stdout(predicate::str::contains("EVENT_YEAR_MISMATCH"));
}

#[test]
fn cli_validate_per_event_no_timing_warning() {
    let input = r#"{
      "method":"moving_average",
      "tax_year":2026,
      "rounding": {
        "currency": {"JPY": {"scale": 0, "mode": "half_up"}},
        "unit_price": {"scale": 8, "mode": "half_up"},
        "quantity": {"scale": 8, "mode": "half_up"},
        "timing": "per_event"
      },
      "events":[
        {"type":"Acquire","id":"a1","asset":"BTC","qty":"1","jpy_cost":"3000000","ts":"2026-01-01T00:00:00Z"}
      ]
    }"#;

    let mut cmd = Command::new(assert_cmd::cargo::cargo_bin!("eupholio-core-cli"));
    cmd.arg("validate")
        .write_stdin(input)
        .assert()
        .success()
        .stdout(predicate::str::contains("ROUNDING_TIMING_NOT_FULLY_IMPLEMENTED").not());
}
