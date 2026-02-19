use eupholio_core::config::{Config, CostMethod};
use eupholio_core::event::Event;
use eupholio_normalizer::bittrex::normalize_order_history_csv;

#[test]
fn bittrex_smoke_source_to_normalized_to_calculate() {
    let raw = include_str!("fixtures/normalizer/bittrex_order_history_smoke.csv");
    let expected_raw = include_str!("fixtures/normalizer/bittrex_order_history_smoke.normalized.json");

    let normalized = normalize_order_history_csv(raw).expect("normalization should succeed");
    assert_eq!(normalized.diagnostics.len(), 1, "unsupported rows are diagnosed");
    assert_eq!(normalized.diagnostics[0].row, 4);
    assert_eq!(
        normalized.diagnostics[0].reason,
        "unsupported order type: OrderType='UNKNOWN', Uuid='skip-001'"
    );

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

#[test]
fn bittrex_csv_parser_handles_quoted_thousands_separator() {
    let raw = "Uuid,Exchange,OrderType,Quantity,Price,Commission,Closed\n\
q1,JPY-BTC,LIMIT_BUY,1,\"1,234\",10,01/01/2026 12:00:00 AM\n";

    let normalized = normalize_order_history_csv(raw).expect("normalization should succeed");
    assert!(normalized.diagnostics.is_empty());
    assert_eq!(normalized.events.len(), 1);

    match &normalized.events[0] {
        Event::Acquire { jpy_cost, .. } => assert_eq!(jpy_cost.to_string(), "1244"),
        _ => panic!("expected acquire event"),
    }
}

#[test]
fn bittrex_normalizer_errors_on_missing_required_header() {
    let raw = "Uuid,Exchange,OrderType,Quantity,Price,Commission\n\
q1,JPY-BTC,LIMIT_BUY,1,100,1\n";

    let err = normalize_order_history_csv(raw).expect_err("missing header should fail");
    assert!(err.contains("missing required header Closed"));
}

#[test]
fn bittrex_normalizer_errors_on_invalid_exchange_and_values() {
    let bad_exchange = "Uuid,Exchange,OrderType,Quantity,Price,Commission,Closed\n\
q1,JPY--BTC,LIMIT_BUY,1,100,1,01/01/2026 12:00:00 AM\n";
    assert!(normalize_order_history_csv(bad_exchange)
        .expect_err("invalid exchange should fail")
        .contains("invalid exchange pair"));

    let non_jpy = "Uuid,Exchange,OrderType,Quantity,Price,Commission,Closed\n\
q2,USD-BTC,LIMIT_BUY,1,100,1,01/01/2026 12:00:00 AM\n";
    assert!(normalize_order_history_csv(non_jpy)
        .expect_err("non-JPY should fail")
        .contains("only JPY is supported"));

    let bad_decimal = "Uuid,Exchange,OrderType,Quantity,Price,Commission,Closed\n\
q3,JPY-BTC,LIMIT_BUY,not-a-number,100,1,01/01/2026 12:00:00 AM\n";
    assert!(normalize_order_history_csv(bad_decimal)
        .expect_err("invalid decimal should fail")
        .contains("invalid decimal"));

    let bad_datetime = "Uuid,Exchange,OrderType,Quantity,Price,Commission,Closed\n\
q4,JPY-BTC,LIMIT_BUY,1,100,1,not-a-datetime\n";
    assert!(normalize_order_history_csv(bad_datetime)
        .expect_err("invalid datetime should fail")
        .contains("invalid datetime"));
}
