use eupholio_core::event::Event;
use eupholio_normalizer::cryptact::normalize_custom_csv;
use rust_decimal::Decimal;
use std::str::FromStr;

#[test]
fn cryptact_normalize_buy_sell_jpy() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,BUY,bitFlyer,BTC,0.1,6000000,JPY,120,JPY,
2026/1/3 12:00:00,SELL,bitFlyer,BTC,0.05,6200000,JPY,100,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.diagnostics.len(), 0);
    assert_eq!(got.events.len(), 2);

    match &got.events[0] {
        Event::Acquire {
            asset,
            qty,
            jpy_cost,
            ..
        } => {
            assert_eq!(asset, "BTC");
            assert_eq!(*qty, d("0.1"));
            assert_eq!(*jpy_cost, d("600120"));
        }
        other => panic!("unexpected event: {other:?}"),
    }

    match &got.events[1] {
        Event::Dispose {
            asset,
            qty,
            jpy_proceeds,
            ..
        } => {
            assert_eq!(asset, "BTC");
            assert_eq!(*qty, d("0.05"));
            assert_eq!(*jpy_proceeds, d("309900"));
        }
        other => panic!("unexpected event: {other:?}"),
    }
}

#[test]
fn cryptact_normalize_rejects_missing_required_header() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,Comment
2026/1/2 12:00:00,BUY,bitFlyer,BTC,0.1,6000000,JPY,120,
"#;

    let err = normalize_custom_csv(csv).expect_err("missing FeeCcy should fail");
    assert!(err.contains("missing required header FeeCcy"));
}

#[test]
fn cryptact_normalize_unsupported_action_to_diagnostic() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,MINING,bitFlyer,BTC,0.1,0,JPY,0,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.events.len(), 0);
    assert_eq!(got.diagnostics.len(), 1);
    assert!(got.diagnostics[0].reason.contains("unsupported action"));
}

fn d(v: &str) -> Decimal {
    Decimal::from_str(v).expect("valid decimal")
}
