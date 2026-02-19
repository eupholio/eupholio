use eupholio_core::config::{Config, CostMethod};
use eupholio_core::event::Event;
use eupholio_normalizer::bitflyer::normalize_transaction_history_csv;

#[test]
fn bitflyer_smoke_source_to_normalized_to_calculate() {
    let raw = include_str!("fixtures/normalizer/bitflyer_transaction_history_smoke.csv");
    let expected_raw = include_str!("fixtures/normalizer/bitflyer_transaction_history_smoke.normalized.json");

    let normalized = normalize_transaction_history_csv(raw).expect("normalization should succeed");
    assert_eq!(normalized.diagnostics.len(), 1);
    assert!(normalized.diagnostics[0]
        .reason
        .contains("unsupported trade type"));

    let expected: Vec<Event> = serde_json::from_str(expected_raw).expect("fixture json should be valid");
    assert_eq!(normalized.events, expected);

    let report = eupholio_core::calculate(
        Config {
            method: CostMethod::MovingAverage,
            tax_year: 2026,
            rounding: Default::default(),
        },
        &normalized.events,
    );

    assert_eq!(report.realized_pnl_jpy.to_string(), "1981");
}

#[test]
fn bitflyer_errors_on_missing_header_and_non_jpy_payment() {
    let missing = "取引日時,通貨,取引種別\n2026/01/01 00:00:00,BTC/JPY,買い\n";
    assert!(normalize_transaction_history_csv(missing)
        .expect_err("missing header should fail")
        .contains("missing required header"));

    let non_jpy = "取引日時,通貨,取引種別,取引価格,通貨1,通貨1数量,手数料,通貨1の対円レート,通貨2,通貨2数量,自己・媒介,注文 ID,備考\n\
2026/01/01 00:00:00,BTC/USD,買い,500000,BTC,0.01,-0.00001,500000,USD,-5000,媒介,bf-order-004,\n";
    assert!(normalize_transaction_history_csv(non_jpy)
        .expect_err("non-jpy should fail")
        .contains("only JPY is supported"));
}
