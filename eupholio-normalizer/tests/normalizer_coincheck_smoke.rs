use eupholio_core::config::{Config, CostMethod};
use eupholio_core::event::Event;
use eupholio_normalizer::coincheck::normalize_trade_history_csv;

#[test]
fn coincheck_smoke_source_to_normalized_to_calculate() {
    let raw = include_str!("fixtures/normalizer/coincheck_history_smoke.csv");
    let expected_raw = include_str!("fixtures/normalizer/coincheck_history_smoke.normalized.json");

    let normalized = normalize_trade_history_csv(raw).expect("normalization should succeed");
    assert!(normalized.diagnostics.is_empty());

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
    let btc = report.positions.get("BTC").expect("btc position should exist");
    assert_eq!(btc.qty.to_string(), "0.0");
}

#[test]
fn coincheck_parser_handles_commas_and_blank_fee() {
    let raw = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc1,2026-01-01 00:00:00 +0900,Completed trading contracts,\"1,000\",JPY,,,,\"Rate: 10000.0, Pair: btc_jpy\"\n";

    let normalized = normalize_trade_history_csv(raw).expect("normalization should succeed");
    assert!(normalized.diagnostics.is_empty());
    assert_eq!(normalized.events.len(), 1);

    match &normalized.events[0] {
        Event::Dispose {
            qty,
            jpy_proceeds,
            ..
        } => {
            assert_eq!(qty.to_string(), "0.1");
            assert_eq!(jpy_proceeds.to_string(), "1000");
        }
        _ => panic!("expected dispose event"),
    }
}

#[test]
fn coincheck_normalizer_errors_on_missing_required_header_and_invalid_rows() {
    let missing_header = "id,time,operation,amount,trading_currency,price,original_currency,fee\n\
cc1,2026-01-01 00:00:00 +0900,Completed trading contracts,1,BTC,,,0\n";
    assert!(normalize_trade_history_csv(missing_header)
        .expect_err("missing header should fail")
        .contains("missing required header comment"));

    let bad_comment = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc2,2026-01-01 00:00:00 +0900,Completed trading contracts,1,BTC,,,0,broken\n";
    assert!(normalize_trade_history_csv(bad_comment)
        .expect_err("bad comment should fail")
        .contains("failed to parse comment"));

    let non_jpy = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc3,2026-01-01 00:00:00 +0900,Completed trading contracts,1,BTC,,,0,\"Rate: 10000.0, Pair: btc_usd\"\n";
    assert!(normalize_trade_history_csv(non_jpy)
        .expect_err("non-JPY should fail")
        .contains("only JPY is supported"));

    let invalid_currency = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc4,2026-01-01 00:00:00 +0900,Completed trading contracts,1,ETH,,,0,\"Rate: 10000.0, Pair: btc_jpy\"\n";
    assert!(normalize_trade_history_csv(invalid_currency)
        .expect_err("invalid trading currency should fail")
        .contains("invalid trading currency"));
}

#[test]
fn coincheck_fee_parse_error_is_reported() {
    let bad_fee = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc5,2026-01-01 00:00:00 +0900,Completed trading contracts,1,BTC,,,not-a-number,\"Rate: 10000.0, Pair: btc_jpy\"\n";

    assert!(normalize_trade_history_csv(bad_fee)
        .expect_err("invalid fee should fail")
        .contains("invalid decimal"));
}

#[test]
fn coincheck_timestamp_timezone_crosses_tax_year_boundary() {
    let raw = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc6,2025-12-31 23:30:00 -0900,Completed trading contracts,1000,JPY,,,0,\"Rate: 10000.0, Pair: btc_jpy\"\n";

    let normalized = normalize_trade_history_csv(raw).expect("normalization should succeed");
    assert!(normalized.diagnostics.is_empty());
    assert_eq!(normalized.events.len(), 1);

    match &normalized.events[0] {
        Event::Dispose { ts, .. } => {
            assert_eq!(ts.to_rfc3339(), "2026-01-01T08:30:00+00:00");
        }
        _ => panic!("expected dispose event"),
    }
}

#[test]
fn coincheck_trading_currency_is_case_insensitive() {
    let raw = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc7,2026-01-01 00:00:00 +0900,Completed trading contracts,1,btc,,,0,\"Rate: 10000.0, Pair: btc_jpy\"\n";

    let normalized = normalize_trade_history_csv(raw).expect("normalization should succeed");
    assert!(normalized.diagnostics.is_empty());
    assert_eq!(normalized.events.len(), 1);
    match &normalized.events[0] {
        Event::Acquire { .. } => {}
        _ => panic!("expected acquire event"),
    }
}

#[test]
fn coincheck_zero_rate_is_rejected() {
    let raw = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc8,2026-01-01 00:00:00 +0900,Completed trading contracts,1000,JPY,,,0,\"Rate: 0, Pair: btc_jpy\"\n";

    assert!(normalize_trade_history_csv(raw)
        .expect_err("zero rate should fail")
        .contains("rate must be > 0"));
}

#[test]
fn coincheck_transfer_requires_positive_qty() {
    let raw = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc9,2026-01-01 00:00:00 +0900,Sent,-0.01,BTC,,,0,\"Address: 1ABC\"\n";

    assert!(normalize_trade_history_csv(raw)
        .expect_err("non-positive transfer qty should fail")
        .contains("transfer qty must be > 0"));
}

#[test]
fn coincheck_unsupported_operation_is_diagnosed() {
    let raw = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc10,2026-01-01 00:00:00 +0900,Canceled,1,BTC,,,0,\"\"\n";

    let normalized = normalize_trade_history_csv(raw).expect("normalization should succeed");
    assert_eq!(normalized.events.len(), 0);
    assert_eq!(normalized.diagnostics.len(), 1);
    assert!(normalized.diagnostics[0]
        .reason
        .contains("unsupported operation"));
}

#[test]
fn coincheck_transfer_handles_lowercase_asset_with_whitespace() {
    let raw = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc11,2026-01-01 00:00:00 +0900,Received,1.25, btc ,,,0,\"\"\n";

    let normalized = normalize_trade_history_csv(raw).expect("normalization should succeed");
    assert!(normalized.diagnostics.is_empty());
    assert_eq!(normalized.events.len(), 1);

    match &normalized.events[0] {
        Event::Transfer {
            asset,
            qty,
            direction,
            ..
        } => {
            assert_eq!(asset, "BTC");
            assert_eq!(qty.to_string(), "1.25");
            assert_eq!(*direction, eupholio_core::event::TransferDirection::In);
        }
        _ => panic!("expected transfer event"),
    }
}

#[test]
fn coincheck_transfer_empty_amount_is_rejected() {
    let raw = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc12,2026-01-01 00:00:00 +0900,Sent,,BTC,,,0,\"\"\n";

    assert!(normalize_trade_history_csv(raw)
        .expect_err("empty amount should fail")
        .contains("invalid decimal"));
}

#[test]
fn coincheck_transfer_bad_datetime_is_rejected() {
    let raw = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc13,2026/01/01 00:00:00,Received,1,BTC,,,0,\"\"\n";

    assert!(normalize_trade_history_csv(raw)
        .expect_err("invalid datetime should fail")
        .contains("invalid datetime"));
}

#[test]
fn coincheck_fiat_deposit_withdrawal_map_to_transfer() {
    let raw = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc14,2026-01-01 00:00:00 +0900,Deposit,10000,JPY,,,0,\"\"\n\
cc15,2026-01-02 00:00:00 +0900,Withdrawal,2500,JPY,,,0,\"\"\n";

    let normalized = normalize_trade_history_csv(raw).expect("normalization should succeed");
    assert!(normalized.diagnostics.is_empty());
    assert_eq!(normalized.events.len(), 2);

    match &normalized.events[0] {
        Event::Transfer {
            id,
            asset,
            direction,
            ..
        } => {
            assert_eq!(id, "cc14:fiat_transfer_in");
            assert_eq!(asset, "JPY");
            assert_eq!(*direction, eupholio_core::event::TransferDirection::In);
        }
        _ => panic!("expected transfer event"),
    }

    match &normalized.events[1] {
        Event::Transfer {
            id,
            asset,
            direction,
            ..
        } => {
            assert_eq!(id, "cc15:fiat_transfer_out");
            assert_eq!(asset, "JPY");
            assert_eq!(*direction, eupholio_core::event::TransferDirection::Out);
        }
        _ => panic!("expected transfer event"),
    }
}

#[test]
fn coincheck_non_jpy_fiat_transfer_is_diagnosed() {
    let raw = "id,time,operation,amount,trading_currency,price,original_currency,fee,comment\n\
cc16,2026-01-01 00:00:00 +0900,Deposit,10,USD,,,0,\"\"\n";

    let normalized = normalize_trade_history_csv(raw).expect("normalization should succeed");
    assert_eq!(normalized.events.len(), 0);
    assert_eq!(normalized.diagnostics.len(), 1);
    assert!(normalized.diagnostics[0]
        .reason
        .contains("unsupported fiat transfer currency"));
}
