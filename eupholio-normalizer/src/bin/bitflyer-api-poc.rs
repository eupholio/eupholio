use std::fs;
use std::path::PathBuf;
use std::process;

use chrono::{DateTime, Utc};
use eupholio_core::config::{Config, CostMethod};
use eupholio_normalizer::bitflyer_api::{BitflyerApiClient, FetchWindowOptions};
use serde_json::json;

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
    let out_dir = args.get(6).map(PathBuf::from);

    let opts = FetchWindowOptions {
        product_code,
        count_per_page,
        max_pages: 20,
        since,
        until,
    };

    let client = BitflyerApiClient::from_env()?;
    let normalized = client.fetch_and_normalize_window(&opts)?;

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

    if let Some(dir) = out_dir {
        fs::create_dir_all(&dir).map_err(|e| format!("failed to create out dir: {e}"))?;

        let events_path = dir.join("events.json");
        let diagnostics_path = dir.join("diagnostics.json");
        let report_path = dir.join("report.json");
        let meta_path = dir.join("meta.json");

        fs::write(
            &events_path,
            serde_json::to_vec_pretty(&normalized.events)
                .map_err(|e| format!("failed to encode events json: {e}"))?,
        )
        .map_err(|e| format!("failed to write events json: {e}"))?;

        fs::write(
            &diagnostics_path,
            serde_json::to_vec_pretty(&normalized.diagnostics)
                .map_err(|e| format!("failed to encode diagnostics json: {e}"))?,
        )
        .map_err(|e| format!("failed to write diagnostics json: {e}"))?;

        fs::write(
            &report_path,
            serde_json::to_vec_pretty(&report)
                .map_err(|e| format!("failed to encode report json: {e}"))?,
        )
        .map_err(|e| format!("failed to write report json: {e}"))?;

        let meta = json!({
            "product_code": opts.product_code,
            "tax_year": tax_year,
            "count_per_page": opts.count_per_page,
            "max_pages": opts.max_pages,
            "since": opts.since,
            "until": opts.until,
            "generated_at": Utc::now(),
            "files": {
                "events": events_path,
                "diagnostics": diagnostics_path,
                "report": report_path
            }
        });

        fs::write(
            &meta_path,
            serde_json::to_vec_pretty(&meta)
                .map_err(|e| format!("failed to encode meta json: {e}"))?,
        )
        .map_err(|e| format!("failed to write meta json: {e}"))?;

        println!("exported json to {}", dir.display());
    }

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
