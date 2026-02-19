use std::collections::HashMap;

use chrono::{TimeZone, Utc};
use eupholio_core::{
    calculate, calculate_total_average_with_carry,
    config::{Config, CostMethod, RoundingPolicy},
    event::{Event, TransferDirection},
    report::CarryIn,
};
use rust_decimal_macros::dec;

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
    // realized = 7,000,000 - 1 * avg
    assert!(report.realized_pnl_jpy > dec!(2333333));
    assert!(report.realized_pnl_jpy < dec!(2333334));
    assert_eq!(report.positions["BTC"].qty, dec!(2));
}
