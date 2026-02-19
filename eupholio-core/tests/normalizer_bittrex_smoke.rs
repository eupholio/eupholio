use eupholio_core::config::{Config, CostMethod};
use eupholio_core::event::Event;
use eupholio_core::normalizer::bittrex::normalize_order_history_csv;

#[test]
fn bittrex_smoke_source_to_normalized_to_calculate() {
    let raw = include_str!("fixtures/normalizer/bittrex_order_history_smoke.csv");
    let expected_raw = include_str!("fixtures/normalizer/bittrex_order_history_smoke.normalized.json");

    let normalized = normalize_order_history_csv(raw).expect("normalization should succeed");
    assert_eq!(normalized.diagnostics.len(), 1, "unsupported rows are diagnosed");
    assert_eq!(normalized.diagnostics[0].row, 4);
    assert_eq!(normalized.diagnostics[0].reason, "unsupported order type");

    let expected: Vec<Event> = serde_json::from_str(expected_raw).expect("fixture json should be valid");
    assert_eq!(normalized.events, expected, "normalized output should match fixture");

    let report = eupholio_core::calculate(
        Config {
            method: CostMethod::MovingAverage,
            tax_year: 2026,
            rounding: Default::default(),
        },
        &normalized.events,
    );

    assert_eq!(report.realized_pnl_jpy.to_string(), "198800");
    assert_eq!(report.positions.len(), 1);
    let btc = report.positions.get("BTC").expect("btc position should exist");
    assert_eq!(btc.qty.to_string(), "0.0");
}
