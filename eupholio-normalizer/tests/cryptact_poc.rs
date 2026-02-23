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
fn cryptact_normalize_buy_with_base_fee_reduces_acquired_qty() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,BUY,bitFlyer,ETH,2,300000,JPY,0.01,ETH,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.diagnostics.len(), 0);
    assert_eq!(got.events.len(), 1);

    match &got.events[0] {
        Event::Acquire { qty, jpy_cost, .. } => {
            assert_eq!(*qty, d("1.99"));
            assert_eq!(*jpy_cost, d("600000"));
        }
        other => panic!("unexpected event: {other:?}"),
    }
}

#[test]
fn cryptact_normalize_sell_with_base_fee_increases_disposed_qty() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,SELL,bitFlyer,ETH,2,300000,JPY,0.01,ETH,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.diagnostics.len(), 0);
    assert_eq!(got.events.len(), 1);

    match &got.events[0] {
        Event::Dispose {
            qty, jpy_proceeds, ..
        } => {
            assert_eq!(*qty, d("2.01"));
            assert_eq!(*jpy_proceeds, d("600000"));
        }
        other => panic!("unexpected event: {other:?}"),
    }
}

#[test]
fn cryptact_normalize_buy_missing_price_is_error() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,BUY,bitFlyer,BTC,0.1,,JPY,120,JPY,
"#;

    let err = normalize_custom_csv(csv).expect_err("missing price should fail");
    assert!(err.contains("price must be provided for BUY"));
}

#[test]
fn cryptact_normalize_header_lookup_is_case_insensitive() {
    let csv = r#"timestamp,action,source,base,volume,price,counter,fee,feeccy,comment
2026/1/2 12:00:00,BUY,bitFlyer,BTC,0.1,6000000,JPY,120,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("lowercase headers should parse");
    assert_eq!(got.events.len(), 1);
}

#[test]
fn cryptact_normalize_ids_are_unique_per_row() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,BUY,bitFlyer,BTC,0.1,6000000,JPY,120,JPY,
2026/1/2 12:00:00,BUY,bitFlyer,BTC,0.2,6100000,JPY,0,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.events.len(), 2);
    assert_ne!(got.events[0].id(), got.events[1].id());
}

#[test]
fn cryptact_normalize_unsupported_action_to_diagnostic() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,RETURN,bitFlyer,BTC,0.1,0,JPY,0,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.events.len(), 0);
    assert_eq!(got.diagnostics.len(), 1);
    assert!(got.diagnostics[0].reason.contains("unsupported action"));
}

#[test]
fn cryptact_normalize_pay_to_dispose() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,PAY,bitFlyer,BTC,0.01,500000,JPY,0,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.events.len(), 1);
    match &got.events[0] {
        Event::Dispose {
            asset,
            qty,
            jpy_proceeds,
            ..
        } => {
            assert_eq!(asset, "BTC");
            assert_eq!(*qty, d("0.01"));
            assert_eq!(*jpy_proceeds, d("5000"));
        }
        other => panic!("unexpected event: {other:?}"),
    }
}

#[test]
fn cryptact_normalize_pay_missing_price_uses_zero_proceeds() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,PAY,bitFlyer,BTC,0.01,,JPY,0,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.events.len(), 1);
    match &got.events[0] {
        Event::Dispose { jpy_proceeds, .. } => assert_eq!(*jpy_proceeds, d("0")),
        other => panic!("unexpected event: {other:?}"),
    }
}

#[test]
fn cryptact_normalize_mining_to_income() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,MINING,bitFlyer,ETH,1,,JPY,0,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.events.len(), 1);
    match &got.events[0] {
        Event::Income {
            asset,
            qty,
            jpy_value,
            ..
        } => {
            assert_eq!(asset, "ETH");
            assert_eq!(*qty, d("1"));
            assert_eq!(*jpy_value, d("0"));
        }
        other => panic!("unexpected event: {other:?}"),
    }
}

#[test]
fn cryptact_normalize_pay_and_mining_non_jpy_fee_ccy_to_diagnostics() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,PAY,bitFlyer,BTC,0.01,500000,JPY,0,BTC,
2026/1/3 12:00:00,MINING,bitFlyer,ETH,1,1000,JPY,0,ETH,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.events.len(), 0);
    assert_eq!(got.diagnostics.len(), 2);
}

#[test]
fn cryptact_normalize_pay_nonzero_fee_errors() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,PAY,bitFlyer,BTC,0.01,500000,JPY,10,JPY,
"#;

    let err = normalize_custom_csv(csv).expect_err("non-zero PAY fee should fail");
    assert!(err.contains("fee must be 0 for PAY"));
}

#[test]
fn cryptact_normalize_sendfee_nonzero_fee_errors() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,SENDFEE,bitFlyer,BTC,0.0001,,JPY,10,JPY,
"#;

    let err = normalize_custom_csv(csv).expect_err("non-zero sendfee fee should fail");
    assert!(err.contains("fee must be 0 for SENDFEE"));
}

#[test]
fn cryptact_normalize_sendfee_to_transfer_out() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,SENDFEE,bitFlyer,BTC,0.0001,,JPY,0,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.events.len(), 1);
    match &got.events[0] {
        Event::Transfer {
            asset,
            qty,
            direction,
            ..
        } => {
            assert_eq!(asset, "BTC");
            assert_eq!(*qty, d("0.0001"));
            assert_eq!(*direction, eupholio_core::event::TransferDirection::Out);
        }
        other => panic!("unexpected event: {other:?}"),
    }
}

#[test]
fn cryptact_normalize_bonus_lending_staking_to_income() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,BONUS,bitFlyer,BTC,0.01,,JPY,0,JPY,
2026/1/3 12:00:00,LENDING,bitFlyer,ETH,1,1000,JPY,0,JPY,
2026/1/4 12:00:00,STAKING,bitFlyer,SOL,2,500,JPY,0,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.diagnostics.len(), 0);
    assert_eq!(got.events.len(), 3);

    match &got.events[0] {
        Event::Income {
            asset,
            qty,
            jpy_value,
            ..
        } => {
            assert_eq!(asset, "BTC");
            assert_eq!(*qty, d("0.01"));
            assert_eq!(*jpy_value, d("0"));
        }
        other => panic!("unexpected event[0]: {other:?}"),
    }
    match &got.events[1] {
        Event::Income {
            asset,
            qty,
            jpy_value,
            ..
        } => {
            assert_eq!(asset, "ETH");
            assert_eq!(*qty, d("1"));
            assert_eq!(*jpy_value, d("1000"));
        }
        other => panic!("unexpected event[1]: {other:?}"),
    }
    match &got.events[2] {
        Event::Income {
            asset,
            qty,
            jpy_value,
            ..
        } => {
            assert_eq!(asset, "SOL");
            assert_eq!(*qty, d("2"));
            assert_eq!(*jpy_value, d("1000"));
        }
        other => panic!("unexpected event[2]: {other:?}"),
    }
}

#[test]
fn cryptact_normalize_tip_to_dispose() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,TIP,bitFlyer,BTC,0.0001,6000000,JPY,0,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.events.len(), 1);
    match &got.events[0] {
        Event::Dispose {
            qty, jpy_proceeds, ..
        } => {
            assert_eq!(*qty, d("0.0001"));
            assert_eq!(*jpy_proceeds, d("600"));
        }
        other => panic!("unexpected event: {other:?}"),
    }
}

#[test]
fn cryptact_normalize_phase4_non_jpy_fee_ccy_to_diagnostics() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,BONUS,bitFlyer,BTC,0.01,,JPY,0,BTC,
2026/1/3 12:00:00,TIP,bitFlyer,ETH,0.1,1000,JPY,0,ETH,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.events.len(), 0);
    assert_eq!(got.diagnostics.len(), 2);
}

#[test]
fn cryptact_normalize_phase4_nonzero_fee_errors() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,LOSS,bitFlyer,BTC,0.1,,JPY,1,JPY,
"#;

    let err = normalize_custom_csv(csv).expect_err("non-zero phase-4 fee should fail");
    assert!(err.contains("fee must be 0 for LOSS in phase-4"));
}

#[test]
fn cryptact_normalize_loss_and_reduce() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,LOSS,bitFlyer,BTC,0.1,,JPY,0,JPY,
2026/1/3 12:00:00,REDUCE,bitFlyer,ETH,1,,JPY,0,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.events.len(), 2);

    match &got.events[0] {
        Event::Dispose {
            qty, jpy_proceeds, ..
        } => {
            assert_eq!(*qty, d("0.1"));
            assert_eq!(*jpy_proceeds, d("0"));
        }
        other => panic!("unexpected loss event: {other:?}"),
    }

    match &got.events[1] {
        Event::Transfer { qty, direction, .. } => {
            assert_eq!(*qty, d("1"));
            assert_eq!(*direction, eupholio_core::event::TransferDirection::Out);
        }
        other => panic!("unexpected reduce event: {other:?}"),
    }
}

#[test]
fn cryptact_normalize_phase5_lend_recover_borrow_return_defifee() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,LEND,bitFlyer,BTC,0.1,,JPY,0,JPY,
2026/1/3 12:00:00,RECOVER,bitFlyer,BTC,0.05,,JPY,0,JPY,
2026/1/4 12:00:00,BORROW,bitFlyer,ETH,1,,JPY,0,JPY,
2026/1/5 12:00:00,RETURN,bitFlyer,ETH,1,,JPY,0,JPY,
2026/1/6 12:00:00,DEFIFEE,bitFlyer,BNB,0.01,,JPY,0,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.diagnostics.len(), 0);
    assert_eq!(got.events.len(), 5);

    assert!(matches!(
        &got.events[0],
        Event::Transfer {
            direction: eupholio_core::event::TransferDirection::Out,
            ..
        }
    ));
    assert!(matches!(
        &got.events[1],
        Event::Transfer {
            direction: eupholio_core::event::TransferDirection::In,
            ..
        }
    ));
    assert!(matches!(
        &got.events[2],
        Event::Transfer {
            direction: eupholio_core::event::TransferDirection::In,
            ..
        }
    ));
    assert!(matches!(
        &got.events[3],
        Event::Transfer {
            direction: eupholio_core::event::TransferDirection::Out,
            ..
        }
    ));
    assert!(matches!(
        &got.events[4],
        Event::Dispose { jpy_proceeds, .. } if *jpy_proceeds == d("0")
    ));
}

#[test]
fn cryptact_normalize_phase5_cash_to_diagnostic() {
    let csv = r#"Timestamp,Action,Source,Base,Volume,Price,Counter,Fee,FeeCcy,Comment
2026/1/2 12:00:00,CASH,manual,JPY,0,0,JPY,2000,JPY,
"#;

    let got = normalize_custom_csv(csv).expect("should parse");
    assert_eq!(got.events.len(), 0);
    assert_eq!(got.diagnostics.len(), 1);
    assert!(got.diagnostics[0].reason.contains("CASH is not supported"));
}

fn d(v: &str) -> Decimal {
    Decimal::from_str(v).expect("valid decimal")
}
