use chrono::{TimeZone, Utc};
use eupholio_core::{
    calculate,
    config::{Config, CostMethod, RoundRule, RoundingMode, RoundingPolicy, RoundingTiming},
    event::Event,
};
use rust_decimal_macros::dec;
use std::collections::HashMap;

fn ts(y: i32, m: u32, d: u32) -> chrono::DateTime<Utc> {
    Utc.with_ymd_and_hms(y, m, d, 0, 0, 0).unwrap()
}

#[test]
fn per_year_rounding_keeps_carry_out_cost_coherent_with_qty_and_avg() {
    let events = vec![
        Event::Acquire {
            id: "a1".into(),
            asset: "BTC".into(),
            qty: dec!(2.4),
            jpy_cost: dec!(241.44),
            ts: ts(2026, 1, 1),
        },
        Event::Dispose {
            id: "d1".into(),
            asset: "BTC".into(),
            qty: dec!(1.0),
            jpy_proceeds: dec!(150),
            ts: ts(2026, 1, 2),
        },
    ];

    let mut currency = HashMap::new();
    currency.insert(
        "JPY".to_string(),
        RoundRule {
            scale: 0,
            mode: RoundingMode::HalfUp,
        },
    );

    let report = calculate(
        Config {
            method: CostMethod::TotalAverage,
            tax_year: 2026,
            rounding: RoundingPolicy {
                currency,
                unit_price: RoundRule {
                    scale: 0,
                    mode: RoundingMode::HalfUp,
                },
                quantity: RoundRule {
                    scale: 0,
                    mode: RoundingMode::HalfUp,
                },
                timing: RoundingTiming::PerYear,
            },
        },
        &events,
    );

    let summary = &report
        .yearly_summary
        .as_ref()
        .unwrap()
        .by_asset["BTC"];

    assert_eq!(summary.average_cost_per_unit, dec!(101));
    assert_eq!(summary.carry_out_qty, dec!(1));
    assert_eq!(summary.carry_out_cost, dec!(101));
    assert_eq!(summary.carry_out_cost, summary.carry_out_qty * summary.average_cost_per_unit);
}
