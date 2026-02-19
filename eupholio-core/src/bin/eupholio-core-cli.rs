use std::{collections::HashMap, io::{self, Read}};

use clap::{Parser, Subcommand};
use eupholio_core::{
    calculate, calculate_total_average_with_carry,
    config::{Config, CostMethod, RoundingPolicy},
    event::Event,
    report::{CarryIn, Report},
};
use serde::Deserialize;

#[derive(Debug, Deserialize)]
struct Input {
    method: String,
    tax_year: i32,
    events: Vec<Event>,
    #[serde(default)]
    carry_in: HashMap<String, CarryIn>,
}

#[derive(Debug, Parser)]
#[command(name = "eupholio-core-cli", version, about = "Eupholio core calculator CLI")]
struct Cli {
    #[command(subcommand)]
    command: Option<Commands>,
}

#[derive(Debug, Subcommand)]
enum Commands {
    /// Calculate report from JSON provided via stdin
    Calc,
    /// Show CLI version
    Version,
}

fn parse_method(s: &str) -> Result<CostMethod, String> {
    match s {
        "moving_average" => Ok(CostMethod::MovingAverage),
        "total_average" => Ok(CostMethod::TotalAverage),
        _ => Err(format!("unsupported method: {s}")),
    }
}

fn run(input: Input) -> Result<Report, io::Error> {
    let method = parse_method(&input.method)
        .map_err(|e| io::Error::new(io::ErrorKind::InvalidInput, e))?;

    let report = match method {
        CostMethod::MovingAverage => calculate(
            Config {
                method,
                tax_year: input.tax_year,
                rounding: RoundingPolicy::default(),
            },
            &input.events,
        ),
        CostMethod::TotalAverage => {
            if input.carry_in.is_empty() {
                calculate(
                    Config {
                        method,
                        tax_year: input.tax_year,
                        rounding: RoundingPolicy::default(),
                    },
                    &input.events,
                )
            } else {
                calculate_total_average_with_carry(input.tax_year, &input.events, &input.carry_in)
            }
        }
    };

    Ok(report)
}

fn run_calc() -> Result<(), Box<dyn std::error::Error>> {
    let mut raw = String::new();
    io::stdin().read_to_string(&mut raw)?;

    let input: Input = serde_json::from_str(&raw)?;
    let report = run(input)?;

    println!("{}", serde_json::to_string_pretty(&report)?);
    Ok(())
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();

    match cli.command.unwrap_or(Commands::Calc) {
        Commands::Calc => run_calc(),
        Commands::Version => {
            println!("{}", env!("CARGO_PKG_VERSION"));
            Ok(())
        }
    }
}
