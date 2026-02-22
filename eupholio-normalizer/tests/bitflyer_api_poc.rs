use chrono::{DateTime, Utc};
use eupholio_core::event::Event;
use eupholio_normalizer::bitflyer_api::{
    build_executions_path, filter_executions_by_time, normalize_executions, sign_request,
    Execution, FetchOptions,
};

#[test]
fn bitflyer_api_sign_request_is_deterministic() {
    let sign = sign_request(
        "secret",
        "1700000000",
        "GET",
        "/v1/me/getexecutions?product_code=BTC_JPY&count=100",
        "",
    );
    assert_eq!(
        sign,
        "3a8a2a40c61b15c6014722eee597a2a111f35d22ba2999e4bc04310a946f9719"
    );
}

#[test]
fn bitflyer_api_normalize_executions_to_events() {
    let raw = r#"[
      {"id": 1001, "side": "BUY", "price": "1000000", "size": "0.01", "exec_date": "2026-01-01T00:00:00Z", "commission": "0.0001"},
      {"id": 1002, "side": "SELL", "price": "1200000", "size": "0.005", "exec_date": "2026-01-02T00:00:00Z", "commission": "0.0001"},
      {"id": 1003, "side": "OTHER", "price": "1", "size": "1", "exec_date": "2026-01-03T00:00:00Z"}
    ]"#;

    let executions: Vec<Execution> = serde_json::from_str(raw).expect("json should parse");
    let normalized =
        normalize_executions(&executions, "btc_jpy").expect("normalization should work");

    assert_eq!(normalized.events.len(), 2);
    assert_eq!(normalized.diagnostics.len(), 1);
    assert!(normalized.diagnostics[0]
        .reason
        .contains("unsupported side"));

    match &normalized.events[0] {
        Event::Acquire {
            id,
            asset,
            qty,
            jpy_cost,
            ..
        } => {
            assert_eq!(id, "bfexec-1001:acquire");
            assert_eq!(asset, "BTC");
            assert_eq!(qty.to_string(), "0.0099");
            assert_eq!(jpy_cost.to_string(), "10100.0000");
        }
        _ => panic!("expected acquire"),
    }

    match &normalized.events[1] {
        Event::Dispose {
            id,
            asset,
            qty,
            jpy_proceeds,
            ..
        } => {
            assert_eq!(id, "bfexec-1002:dispose");
            assert_eq!(asset, "BTC");
            assert_eq!(qty.to_string(), "0.005");
            assert_eq!(jpy_proceeds.to_string(), "5880.0000");
        }
        _ => panic!("expected dispose"),
    }
}

#[test]
fn bitflyer_api_non_jpy_quote_is_rejected() {
    let executions: Vec<Execution> = serde_json::from_str("[]").unwrap();
    assert!(normalize_executions(&executions, "BTC_USD")
        .expect_err("non-jpy should fail")
        .contains("only JPY is supported"));
}

#[test]
fn bitflyer_api_build_executions_path_with_paging() {
    let path = build_executions_path(&FetchOptions {
        product_code: "BTC_JPY".to_string(),
        count: 100,
        before: Some(123),
        after: Some(45),
    });

    assert_eq!(
        path,
        "/v1/me/getexecutions?product_code=BTC_JPY&count=100&before=123&after=45"
    );
}

#[test]
fn bitflyer_api_filter_executions_by_time_window() {
    let raw = r#"[
      {"id": 1, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-01-01T00:00:00Z"},
      {"id": 2, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-02-01T00:00:00Z"},
      {"id": 3, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-03-01T00:00:00Z"}
    ]"#;
    let executions: Vec<Execution> = serde_json::from_str(raw).unwrap();

    let since = DateTime::parse_from_rfc3339("2026-01-15T00:00:00Z")
        .unwrap()
        .with_timezone(&Utc);
    let until = DateTime::parse_from_rfc3339("2026-02-15T00:00:00Z")
        .unwrap()
        .with_timezone(&Utc);

    let filtered = filter_executions_by_time(&executions, Some(since), Some(until));
    assert_eq!(filtered.len(), 1);
    assert_eq!(filtered[0].id, 2);
}
