use std::io::{self, Read};

use eupholio_core::{
    calculate,
    config::{Config, CostMethod},
    event::Event,
};
use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize)]
struct Input {
    method: String,
    tax_year: i32,
    events: Vec<Event>,
}

#[derive(Debug, Serialize)]
struct Output {
    realized_pnl_jpy: String,
    income_jpy: String,
    diagnostics_count: usize,
}

fn parse_method(s: &str) -> Result<CostMethod, String> {
    match s {
        "moving_average" => Ok(CostMethod::MovingAverage),
        "total_average" => Ok(CostMethod::TotalAverage),
        _ => Err(format!("unsupported method: {s}")),
    }
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut raw = String::new();
    io::stdin().read_to_string(&mut raw)?;

    let input: Input = serde_json::from_str(&raw)?;
    let method = parse_method(&input.method)
        .map_err(|e| io::Error::new(io::ErrorKind::InvalidInput, e))?;

    let report = calculate(
        Config {
            method,
            tax_year: input.tax_year,
        },
        &input.events,
    );

    let output = Output {
        realized_pnl_jpy: report.realized_pnl_jpy.to_string(),
        income_jpy: report.income_jpy.to_string(),
        diagnostics_count: report.diagnostics.len(),
    };

    println!("{}", serde_json::to_string_pretty(&output)?);
    Ok(())
}
