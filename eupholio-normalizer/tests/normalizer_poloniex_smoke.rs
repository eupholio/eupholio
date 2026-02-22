use eupholio_core::config::{Config, CostMethod};
use eupholio_core::event::{Event, TransferDirection};
use eupholio_normalizer::poloniex::{
    normalize_deposit_history_csv, normalize_distribution_history_csv, normalize_trade_history_csv,
    normalize_withdrawal_history_csv,
};

#[test]
fn poloniex_smoke_source_to_normalized_to_calculate() {
    let trade_raw = include_str!("fixtures/normalizer/poloniex_trade_smoke.csv");
    let deposit_raw = include_str!("fixtures/normalizer/poloniex_deposit_smoke.csv");
    let withdrawal_raw = include_str!("fixtures/normalizer/poloniex_withdrawal_smoke.csv");
    let distribution_raw = include_str!("fixtures/normalizer/poloniex_distribution_smoke.csv");
    let expected_raw = include_str!("fixtures/normalizer/poloniex_trade_smoke.normalized.json");

    let mut events = Vec::new();

    let trade = normalize_trade_history_csv(trade_raw).expect("trade normalization should succeed");
    let deposit =
        normalize_deposit_history_csv(deposit_raw).expect("deposit normalization should succeed");
    let withdrawal = normalize_withdrawal_history_csv(withdrawal_raw)
        .expect("withdrawal normalization should succeed");
    let distribution = normalize_distribution_history_csv(distribution_raw)
        .expect("distribution normalization should succeed");

    assert!(trade.diagnostics.is_empty());
    assert!(deposit.diagnostics.is_empty());
    assert!(withdrawal.diagnostics.is_empty());
    assert!(distribution.diagnostics.is_empty());

    events.extend(trade.events);
    events.extend(deposit.events);
    events.extend(withdrawal.events);
    events.extend(distribution.events);

    let expected: Vec<Event> =
        serde_json::from_str(expected_raw).expect("fixture json should be valid");
    assert_eq!(events, expected, "normalized output should match fixture");

    let report = eupholio_core::calculate(
        Config {
            method: CostMethod::MovingAverage,
            tax_year: 2026,
            rounding: Default::default(),
        },
        &events,
    );

    assert_eq!(report.realized_pnl_jpy.to_string(), "9375");
}

#[test]
fn poloniex_unsupported_trade_type_is_diagnosed() {
    let raw = "Date,Market,Category,Type,Price,Amount,Total,Fee,Order Number,Base Total Less Fee,Quote Total Less Fee,Fee Currency,Fee Total\n\
2026-01-05 10:00:00,BTC/JPY,Exchange,Margin,1000000,0.1,100000,0.001,1001,-100000,0.099,BTC,0.001\n";

    let normalized = normalize_trade_history_csv(raw).expect("normalization should succeed");
    assert_eq!(normalized.events.len(), 0);
    assert_eq!(normalized.diagnostics.len(), 1);
    assert!(normalized.diagnostics[0]
        .reason
        .contains("unsupported trade type"));
}

#[test]
fn poloniex_non_jpy_payment_is_rejected() {
    let raw = "Date,Market,Category,Type,Price,Amount,Total,Fee,Order Number,Base Total Less Fee,Quote Total Less Fee,Fee Currency,Fee Total\n\
2026-01-05 10:00:00,BTC/USDT,Exchange,Buy,1000000,0.1,100000,0.001,1001,-100000,0.099,BTC,0.001\n";

    assert!(normalize_trade_history_csv(raw)
        .expect_err("non-JPY should fail")
        .contains("only JPY is supported"));
}

#[test]
fn poloniex_distribution_bad_date_is_rejected() {
    let raw = "date,currency,amount,wallet\n2026/01/25,ETH,0.5,exchange\n";

    assert!(normalize_distribution_history_csv(raw)
        .expect_err("bad date should fail")
        .contains("invalid datetime"));
}

#[test]
fn poloniex_withdrawal_emits_fee_transfer_when_positive_fee() {
    let raw = "Date,Currency,Amount,Fee Deducted,Amount - Fee,Address,Status\n\
2026-01-20 09:00:00,BTC,0.01,0.0001,0.0099,addr2,COMPLETE\n";

    let normalized = normalize_withdrawal_history_csv(raw).expect("normalization should succeed");
    assert_eq!(normalized.events.len(), 2);

    match &normalized.events[1] {
        Event::Transfer { direction, qty, .. } => {
            assert_eq!(*direction, TransferDirection::Out);
            assert_eq!(qty.to_string(), "0.0001");
        }
        _ => panic!("expected transfer event"),
    }
}
