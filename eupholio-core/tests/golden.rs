use std::collections::HashMap;

use chrono::{TimeZone, Utc};
use eupholio_core::{
    calculate, calculate_total_average_with_carry,
    config::{Config, CostMethod, RoundingPolicy, RoundingTiming},
    event::{Event, TransferDirection},
    report::{CarryIn, Warning},
};
use rust_decimal::Decimal;
use rust_decimal_macros::dec;
use serde::Deserialize;

fn ts(y: i32, m: u32, d: u32) -> chrono::DateTime<Utc> {
    Utc.with_ymd_and_hms(y, m, d, 0, 0, 0).unwrap()
}

#[test]
fn case1_moving_average_vs_total_average() {
    let events = vec![
        Event::Acquire {
            id: "a1".into(),
            asset: "BTC".into(),
            qty: dec!(1),
            jpy_cost: dec!(3000000),
            ts: ts(2026, 1, 1),
        },
        Event::Acquire {
            id: "a2".into(),
            asset: "BTC".into(),
            qty: dec!(1),
            jpy_cost: dec!(5000000),
            ts: ts(2026, 1, 2),
        },
        Event::Dispose {
            id: "d1".into(),
            asset: "BTC".into(),
            qty: dec!(1),
            jpy_proceeds: dec!(6000000),
            ts: ts(2026, 1, 3),
        },
    ];

    let moving = calculate(
        Config {
            method: CostMethod::MovingAverage,
            tax_year: 2026,
            rounding: RoundingPolicy::default(),
        },
        &events,
    );
    let total = calculate(
        Config {
            method: CostMethod::TotalAverage,
            tax_year: 2026,
            rounding: RoundingPolicy::default(),
        },
        &events,
    );

    assert_eq!(moving.realized_pnl_jpy, dec!(2000000));
    assert_eq!(total.realized_pnl_jpy, dec!(2000000));
    assert_eq!(moving.positions["BTC"].qty, dec!(1));
    assert_eq!(total.positions["BTC"].qty, dec!(1));
}

#[test]
fn case2_transfer_mixed_no_realized_from_transfer() {
    let events = vec![
        Event::Acquire {
            id: "a1".into(),
            asset: "ETH".into(),
            qty: dec!(10),
            jpy_cost: dec!(100000),
            ts: ts(2026, 2, 1),
        },
        Event::Transfer {
            id: "t1".into(),
            asset: "ETH".into(),
            qty: dec!(2),
            direction: TransferDirection::Out,
            ts: ts(2026, 2, 2),
        },
        Event::Dispose {
            id: "d1".into(),
            asset: "ETH".into(),
            qty: dec!(3),
            jpy_proceeds: dec!(45000),
            ts: ts(2026, 2, 3),
        },
    ];

    let moving = calculate(
        Config {
            method: CostMethod::MovingAverage,
            tax_year: 2026,
            rounding: RoundingPolicy::default(),
        },
        &events,
    );

    assert_eq!(moving.realized_pnl_jpy, dec!(15000));
    assert_eq!(moving.positions["ETH"].qty, dec!(5));
}

#[test]
fn case3_crypto_crypto_as_acquire_plus_dispose() {
    let events = vec![
        Event::Acquire {
            id: "a1".into(),
            asset: "BTC".into(),
            qty: dec!(1),
            jpy_cost: dec!(4000000),
            ts: ts(2026, 3, 1),
        },
        Event::Dispose {
            id: "d1".into(),
            asset: "BTC".into(),
            qty: dec!(0.5),
            jpy_proceeds: dec!(2500000),
            ts: ts(2026, 3, 2),
        },
        Event::Acquire {
            id: "a2".into(),
            asset: "ETH".into(),
            qty: dec!(10),
            jpy_cost: dec!(2500000),
            ts: ts(2026, 3, 2),
        },
    ];

    let moving = calculate(
        Config {
            method: CostMethod::MovingAverage,
            tax_year: 2026,
            rounding: RoundingPolicy::default(),
        },
        &events,
    );

    assert_eq!(moving.realized_pnl_jpy, dec!(500000));
    assert_eq!(moving.positions["BTC"].qty, dec!(0.5));
    assert_eq!(moving.positions["ETH"].qty, dec!(10));
}

#[test]
fn case4_total_average_with_carry_in() {
    let events = vec![
        Event::Acquire {
            id: "a1".into(),
            asset: "BTC".into(),
            qty: dec!(1),
            jpy_cost: dec!(6000000),
            ts: ts(2026, 1, 5),
        },
        Event::Dispose {
            id: "d1".into(),
            asset: "BTC".into(),
            qty: dec!(1),
            jpy_proceeds: dec!(7000000),
            ts: ts(2026, 2, 1),
        },
    ];

    let mut carry_in = HashMap::new();
    carry_in.insert(
        "BTC".into(),
        CarryIn {
            qty: dec!(2),
            cost: dec!(8000000),
        },
    );

    let report = calculate_total_average_with_carry(2026, &events, &carry_in);

    // avg = (8,000,000 + 6,000,000) / (2 + 1) = 4,666,666.666...
    // realized raw = 7,000,000 - 1 * avg = 2,333,333.333...
    // report_only + JPY(0桁, half_up) で 2,333,333 へ丸め
    assert_eq!(report.realized_pnl_jpy, dec!(2333333));
    assert_eq!(report.positions["BTC"].qty, dec!(2));
}

#[derive(Debug, Deserialize)]
struct PerEventFixture {
    tax_year: i32,
    method: String,
    rounding: PerEventFixtureRounding,
    events: Vec<Event>,
    expected: PerEventFixtureExpected,
}

#[derive(Debug, Deserialize)]
struct PerEventFixtureRounding {
    currency: HashMap<String, eupholio_core::config::RoundRule>,
    unit_price: eupholio_core::config::RoundRule,
    quantity: eupholio_core::config::RoundRule,
}

#[derive(Debug, Deserialize)]
struct PerEventFixtureExpected {
    report_only_realized_pnl_jpy: Decimal,
    per_event_realized_pnl_jpy: Decimal,
}

fn assert_per_event_fixture(path: &str) {
    let fixture_path = format!("{}/tests/{}", env!("CARGO_MANIFEST_DIR"), path);
    let fixture_raw = std::fs::read_to_string(fixture_path).unwrap();
    let fixture: PerEventFixture = serde_json::from_str(&fixture_raw).unwrap();
    let method = match fixture.method.as_str() {
        "moving_average" => CostMethod::MovingAverage,
        "total_average" => CostMethod::TotalAverage,
        v => panic!("unknown method: {v}"),
    };

    let base_rounding = RoundingPolicy {
        currency: fixture.rounding.currency,
        unit_price: fixture.rounding.unit_price,
        quantity: fixture.rounding.quantity,
        timing: RoundingTiming::ReportOnly,
    };

    let report_only = calculate(
        Config {
            method,
            tax_year: fixture.tax_year,
            rounding: base_rounding.clone(),
        },
        &fixture.events,
    );

    let per_event = calculate(
        Config {
            method,
            tax_year: fixture.tax_year,
            rounding: RoundingPolicy {
                timing: RoundingTiming::PerEvent,
                ..base_rounding
            },
        },
        &fixture.events,
    );

    assert_eq!(
        report_only.realized_pnl_jpy,
        fixture.expected.report_only_realized_pnl_jpy
    );
    assert_eq!(
        per_event.realized_pnl_jpy,
        fixture.expected.per_event_realized_pnl_jpy
    );
    assert_ne!(report_only.realized_pnl_jpy, per_event.realized_pnl_jpy);
}

#[derive(Debug, Deserialize)]
struct PerYearFixture {
    tax_year: i32,
    rounding: PerEventFixtureRounding,
    events: Vec<Event>,
    expected: PerYearFixtureExpected,
}

#[derive(Debug, Deserialize)]
struct PerYearFixtureExpected {
    report_only_realized_pnl_jpy: Decimal,
    per_year_realized_pnl_jpy: Decimal,
}

fn assert_per_year_fixture(path: &str) {
    let fixture_path = format!("{}/tests/{}", env!("CARGO_MANIFEST_DIR"), path);
    let fixture_raw = std::fs::read_to_string(fixture_path).unwrap();
    let fixture: PerYearFixture = serde_json::from_str(&fixture_raw).unwrap();

    let base_rounding = RoundingPolicy {
        currency: fixture.rounding.currency,
        unit_price: fixture.rounding.unit_price,
        quantity: fixture.rounding.quantity,
        timing: RoundingTiming::ReportOnly,
    };

    let report_only = calculate(
        Config {
            method: CostMethod::TotalAverage,
            tax_year: fixture.tax_year,
            rounding: base_rounding.clone(),
        },
        &fixture.events,
    );

    let per_year = calculate(
        Config {
            method: CostMethod::TotalAverage,
            tax_year: fixture.tax_year,
            rounding: RoundingPolicy {
                timing: RoundingTiming::PerYear,
                ..base_rounding
            },
        },
        &fixture.events,
    );

    assert_eq!(
        report_only.realized_pnl_jpy,
        fixture.expected.report_only_realized_pnl_jpy
    );
    assert_eq!(
        per_year.realized_pnl_jpy,
        fixture.expected.per_year_realized_pnl_jpy
    );
    assert_ne!(report_only.realized_pnl_jpy, per_year.realized_pnl_jpy);
}

#[test]
fn case5_per_event_fixture_moving_average_differs_from_report_only() {
    assert_per_event_fixture("fixtures/per_event_moving_difference.json");
}

#[test]
fn case6_per_event_fixture_total_average_differs_from_report_only() {
    assert_per_event_fixture("fixtures/per_event_total_difference.json");
}

#[test]
fn case7_per_year_fixture_total_average_differs_from_report_only() {
    assert_per_year_fixture("fixtures/per_year_total_difference.json");
}

#[test]
fn case8_year_mismatch_is_excluded_consistently_across_methods() {
    let events = vec![
        Event::Acquire {
            id: "a_mismatch".into(),
            asset: "BTC".into(),
            qty: dec!(1),
            jpy_cost: dec!(100),
            ts: ts(2025, 12, 31),
        },
        Event::Acquire {
            id: "a1".into(),
            asset: "BTC".into(),
            qty: dec!(1),
            jpy_cost: dec!(120),
            ts: ts(2026, 1, 1),
        },
        Event::Dispose {
            id: "d1".into(),
            asset: "BTC".into(),
            qty: dec!(1),
            jpy_proceeds: dec!(150),
            ts: ts(2026, 1, 2),
        },
    ];

    let moving = calculate(
        Config {
            method: CostMethod::MovingAverage,
            tax_year: 2026,
            rounding: RoundingPolicy::default(),
        },
        &events,
    );

    let total = calculate(
        Config {
            method: CostMethod::TotalAverage,
            tax_year: 2026,
            rounding: RoundingPolicy::default(),
        },
        &events,
    );

    assert_eq!(moving.realized_pnl_jpy, dec!(30));
    assert_eq!(total.realized_pnl_jpy, dec!(30));
    assert_eq!(moving.positions["BTC"].qty, dec!(0));
    assert_eq!(total.positions["BTC"].qty, dec!(0));

    assert!(moving.diagnostics.iter().any(|w| matches!(
        w,
        Warning::YearMismatch {
            event_year: 2025,
            tax_year: 2026
        }
    )));
    assert!(total.diagnostics.iter().any(|w| matches!(
        w,
        Warning::YearMismatch {
            event_year: 2025,
            tax_year: 2026
        }
    )));
}
