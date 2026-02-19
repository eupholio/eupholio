use std::collections::HashMap;

use chrono::{TimeZone, Utc};
use eupholio_core::{
    calculate, calculate_total_average_with_carry_and_rounding,
    config::{Config, CostMethod, RoundingMode, RoundingPolicy, RoundingTiming},
    event::Event,
    report::CarryIn,
};
use rust_decimal_macros::dec;

fn ts(y: i32, m: u32, d: u32) -> chrono::DateTime<Utc> {
    Utc.with_ymd_and_hms(y, m, d, 0, 0, 0).unwrap()
}

#[test]
fn per_event_total_average_rounds_carry_in_before_first_touch() {
    let events = vec![
        Event::Acquire {
            id: "eth-a1".into(),
            asset: "ETH".into(),
            qty: dec!(1),
            jpy_cost: dec!(100),
            ts: ts(2026, 1, 1),
        },
        Event::Dispose {
            id: "btc-d1".into(),
            asset: "BTC".into(),
            qty: dec!(1),
            jpy_proceeds: dec!(102),
            ts: ts(2026, 1, 2),
        },
    ];

    let mut carry_in = HashMap::new();
    carry_in.insert(
        "BTC".to_string(),
        CarryIn {
            qty: dec!(1.004),
            cost: dec!(100.6),
        },
    );

    let report = calculate_total_average_with_carry_and_rounding(
        2026,
        &events,
        &carry_in,
        RoundingPolicy {
            quantity: eupholio_core::config::RoundRule {
                scale: 2,
                mode: RoundingMode::HalfUp,
            },
            unit_price: eupholio_core::config::RoundRule {
                scale: 2,
                mode: RoundingMode::HalfUp,
            },
            currency: HashMap::from([(
                "JPY".to_string(),
                eupholio_core::config::RoundRule {
                    scale: 0,
                    mode: RoundingMode::HalfUp,
                },
            )]),
            timing: RoundingTiming::PerEvent,
        },
    );

    // carry_in is rounded first: qty=1.00, cost=101, then dispose 1 @102 => pnl=1
    assert_eq!(report.realized_pnl_jpy, dec!(1));
}

#[test]
fn per_event_moving_average_still_rounds_realized_and_income() {
    let events = vec![
        Event::Acquire {
            id: "a1".into(),
            asset: "BTC".into(),
            qty: dec!(1),
            jpy_cost: dec!(100),
            ts: ts(2026, 1, 1),
        },
        Event::Dispose {
            id: "d1".into(),
            asset: "BTC".into(),
            qty: dec!(1),
            jpy_proceeds: dec!(100.6),
            ts: ts(2026, 1, 2),
        },
        Event::Income {
            id: "i1".into(),
            asset: "ETH".into(),
            qty: dec!(1),
            jpy_value: dec!(200.6),
            ts: ts(2026, 1, 3),
        },
    ];

    let report = calculate(
        Config {
            method: CostMethod::MovingAverage,
            tax_year: 2026,
            rounding: RoundingPolicy {
                quantity: eupholio_core::config::RoundRule {
                    scale: 2,
                    mode: RoundingMode::HalfUp,
                },
                unit_price: eupholio_core::config::RoundRule {
                    scale: 2,
                    mode: RoundingMode::HalfUp,
                },
                currency: HashMap::from([(
                    "JPY".to_string(),
                    eupholio_core::config::RoundRule {
                        scale: 0,
                        mode: RoundingMode::HalfUp,
                    },
                )]),
                timing: RoundingTiming::PerEvent,
            },
        },
        &events,
    );

    assert_eq!(report.realized_pnl_jpy, dec!(1));
    assert_eq!(report.income_jpy, dec!(201));
}
