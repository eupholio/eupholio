use std::process;

use chrono::{DateTime, Utc};
use eupholio_core::config::{Config, CostMethod};
use eupholio_normalizer::bitflyer_api::{BitflyerApiClient, FetchWindowOptions};

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
    let count_per_page = args
        .get(2)
        .and_then(|v| v.parse::<usize>().ok())
        .unwrap_or(100);
    let tax_year = args
        .get(3)
        .and_then(|v| v.parse::<i32>().ok())
        .unwrap_or(2026);
    let since = parse_dt(args.get(4).map(String::as_str))?;
    let until = parse_dt(args.get(5).map(String::as_str))?;

    let opts = FetchWindowOptions {
        product_code,
        count_per_page,
        max_pages: 20,
        since,
        until,
    };

    let client = BitflyerApiClient::from_env()?;
    let normalized = client.fetch_normalize_window(&opts)?;

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

fn parse_dt(s: Option<&str>) -> Result<Option<DateTime<Utc>>, String> {
    match s {
        None => Ok(None),
        Some(v) if v.trim().is_empty() => Ok(None),
        Some(v) => DateTime::parse_from_rfc3339(v)
            .map(|dt| Some(dt.with_timezone(&Utc)))
            .map_err(|e| format!("invalid RFC3339 datetime '{}': {}", v, e)),
    }
}
