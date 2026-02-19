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
fn cli_validate_per_year_total_average_no_timing_warning() {
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
        .stdout(predicate::str::contains("ROUNDING_TIMING_NOT_FULLY_IMPLEMENTED").not())
        .stdout(
            predicate::str::contains("ROUNDING_PER_YEAR_UNSUPPORTED_FOR_MOVING_AVERAGE").not(),
        );
}

#[test]
fn cli_validate_ng_per_year_for_moving_average() {
    let input = r#"{
      "method":"moving_average",
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
        .failure()
        .stdout(predicate::str::contains(
            "ROUNDING_PER_YEAR_UNSUPPORTED_FOR_MOVING_AVERAGE",
        ));
}

#[test]
fn cli_calc_ng_per_year_for_moving_average() {
    let input = r#"{
      "method":"moving_average",
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
    cmd.arg("calc")
        .write_stdin(input)
        .assert()
        .failure()
        .stderr(predicate::str::contains(
            "rounding.timing=per_year is not supported for method=moving_average",
        ));
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

#[test]
fn cli_version_works() {
    let mut cmd = Command::new(assert_cmd::cargo::cargo_bin!("eupholio-core-cli"));
    cmd.arg("version")
        .assert()
        .success()
        .stdout(predicate::str::contains("0.1.0"));
}

#[test]
fn cli_calc_default_subcommand_works() {
    let input = r#"{
      "method":"moving_average",
      "tax_year":2026,
      "events":[
        {"type":"Acquire","id":"a1","asset":"BTC","qty":"1","jpy_cost":"3000000","ts":"2026-01-01T00:00:00Z"}
      ]
    }"#;

    let mut cmd = Command::new(assert_cmd::cargo::cargo_bin!("eupholio-core-cli"));
    cmd.write_stdin(input)
        .assert()
        .success()
        .stdout(predicate::str::contains("\"positions\""));
}

#[test]
fn cli_validate_warn_and_error_for_carry_in_moving_average() {
    let input = r#"{
      "method":"moving_average",
      "tax_year":2026,
      "carry_in":{"BTC":{"qty":"0","cost":"1"}},
      "events":[
        {"type":"Acquire","id":"a1","asset":"BTC","qty":"1","jpy_cost":"1","ts":"2026-01-01T00:00:00Z"}
      ]
    }"#;

    let mut cmd = Command::new(assert_cmd::cargo::cargo_bin!("eupholio-core-cli"));
    cmd.arg("validate")
        .write_stdin(input)
        .assert()
        .success()
        .stdout(predicate::str::contains("CARRY_IN_IGNORED_FOR_MOVING"))
        .stdout(predicate::str::contains("CARRY_IN_COST_WITH_ZERO_QTY"));
}

#[test]
fn cli_calc_ng_invalid_method() {
    let input = r#"{
      "method":"unknown_method",
      "tax_year":2026,
      "events":[]
    }"#;

    let mut cmd = Command::new(assert_cmd::cargo::cargo_bin!("eupholio-core-cli"));
    cmd.arg("calc")
        .write_stdin(input)
        .assert()
        .failure()
        .stderr(predicate::str::contains("unsupported method"));
}
