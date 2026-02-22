use std::process;

use eupholio_core::config::{Config, CostMethod};
use eupholio_normalizer::bitflyer_api::{BitflyerApiClient, FetchOptions};

fn main() {
    if let Err(err) = run() {
        eprintln!("error: {err}");
        process::exit(1);
    }
}

fn run() -> Result<(), String> {
    let args: Vec<String> = std::env::args().collect();
    let product_code = args
        .get(1)
        .cloned()
        .unwrap_or_else(|| "BTC_JPY".to_string());
    let count = args
        .get(2)
        .and_then(|v| v.parse::<usize>().ok())
        .unwrap_or(100);
    let tax_year = args
        .get(3)
        .and_then(|v| v.parse::<i32>().ok())
        .unwrap_or(2026);

    let opts = FetchOptions {
        product_code,
        count,
        before: None,
        after: None,
    };

    let client = BitflyerApiClient::from_env()?;
    let normalized = client.fetch_and_normalize_page(&opts)?;

    println!(
        "normalized: events={} diagnostics={}",
        normalized.events.len(),
        normalized.diagnostics.len()
    );

    let report = eupholio_core::calculate(
        Config {
            method: CostMethod::MovingAverage,
            tax_year,
            rounding: Default::default(),
        },
        &normalized.events,
    );

    println!("realized_pnl_jpy={}", report.realized_pnl_jpy);
    println!("income_jpy={}", report.income_jpy);
    println!("positions={}", report.positions.len());

    Ok(())
}
