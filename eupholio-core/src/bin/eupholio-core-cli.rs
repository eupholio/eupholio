use std::{collections::{HashMap, HashSet}, io::{self, Read}};

use clap::{Parser, Subcommand};
use eupholio_core::{
    calculate, calculate_total_average_with_carry_and_rounding,
    config::{Config, CostMethod, RoundingPolicy, RoundingTiming},
    event::Event,
    report::{CarryIn, Report},
};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize)]
struct Input {
    method: String,
    tax_year: i32,
    events: Vec<Event>,
    #[serde(default)]
    carry_in: HashMap<String, CarryIn>,
    #[serde(default)]
    rounding: Option<RoundingPolicy>,
}

#[derive(Debug, Serialize)]
struct ValidationResult {
    ok: bool,
    errors: Vec<String>,
    warnings: Vec<String>,
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
    /// Validate JSON input from stdin
    Validate,
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

fn read_input() -> Result<Input, Box<dyn std::error::Error>> {
    let mut raw = String::new();
    io::stdin().read_to_string(&mut raw)?;
    let input: Input = serde_json::from_str(&raw)?;
    Ok(input)
}

fn run(input: Input) -> Result<Report, io::Error> {
    let method = parse_method(&input.method)
        .map_err(|e| io::Error::new(io::ErrorKind::InvalidInput, e))?;
    let rounding = input.rounding.unwrap_or_default();

    let report = match method {
        CostMethod::MovingAverage => calculate(
            Config {
                method,
                tax_year: input.tax_year,
                rounding,
            },
            &input.events,
        ),
        CostMethod::TotalAverage => {
            if input.carry_in.is_empty() {
                calculate(
                    Config {
                        method,
                        tax_year: input.tax_year,
                        rounding,
                    },
                    &input.events,
                )
            } else {
                calculate_total_average_with_carry_and_rounding(
                    input.tax_year,
                    &input.events,
                    &input.carry_in,
                    rounding,
                )
            }
        }
    };

    Ok(report)
}

fn validate_input(input: &Input) -> ValidationResult {
    let mut errors = Vec::new();
    let mut warnings = Vec::new();

    let method = parse_method(&input.method);
    if let Err(e) = &method {
        errors.push(e.clone());
    }

    if input.events.is_empty() {
        warnings.push("events is empty".to_string());
    }

    if input.tax_year < 2000 || input.tax_year > 3000 {
        warnings.push(format!("tax_year looks unusual: {}", input.tax_year));
    }

    if method.as_ref().ok() == Some(&CostMethod::MovingAverage) && !input.carry_in.is_empty() {
        warnings.push("carry_in is ignored for moving_average".to_string());
    }

    for (asset, c) in &input.carry_in {
        if c.qty < Decimal::ZERO {
            errors.push(format!("carry_in qty must be >= 0 (asset={asset})"));
        }
        if c.cost < Decimal::ZERO {
            errors.push(format!("carry_in cost must be >= 0 (asset={asset})"));
        }
        if c.qty == Decimal::ZERO && c.cost > Decimal::ZERO {
            warnings.push(format!("carry_in has cost without qty (asset={asset})"));
        }
    }

    if let Some(rounding) = &input.rounding {
        if let Some(jpy) = rounding.currency.get("JPY") {
            if jpy.scale > 18 {
                errors.push("rounding.currency.JPY.scale must be <= 18".to_string());
            }
        }
        if rounding.unit_price.scale > 18 {
            errors.push("rounding.unit_price.scale must be <= 18".to_string());
        }
        if rounding.quantity.scale > 18 {
            errors.push("rounding.quantity.scale must be <= 18".to_string());
        }
        if rounding.timing != RoundingTiming::ReportOnly {
            warnings.push("rounding.timing other than report_only is not fully implemented yet".to_string());
        }
    }

    let mut ids = HashSet::new();
    for e in &input.events {
        let id = e.id().to_string();
        if !ids.insert(id.clone()) {
            errors.push(format!("duplicate event id: {id}"));
        }

        match e {
            Event::Acquire { qty, jpy_cost, .. } => {
                if *qty <= Decimal::ZERO {
                    errors.push(format!("Acquire qty must be > 0 (id={})", e.id()));
                }
                if *jpy_cost < Decimal::ZERO {
                    errors.push(format!("Acquire jpy_cost must be >= 0 (id={})", e.id()));
                }
            }
            Event::Dispose {
                qty,
                jpy_proceeds,
                ..
            } => {
                if *qty <= Decimal::ZERO {
                    errors.push(format!("Dispose qty must be > 0 (id={})", e.id()));
                }
                if *jpy_proceeds < Decimal::ZERO {
                    errors.push(format!("Dispose jpy_proceeds must be >= 0 (id={})", e.id()));
                }
            }
            Event::Income { qty, jpy_value, .. } => {
                if *qty <= Decimal::ZERO {
                    errors.push(format!("Income qty must be > 0 (id={})", e.id()));
                }
                if *jpy_value < Decimal::ZERO {
                    errors.push(format!("Income jpy_value must be >= 0 (id={})", e.id()));
                }
            }
            Event::Transfer { qty, .. } => {
                if *qty <= Decimal::ZERO {
                    errors.push(format!("Transfer qty must be > 0 (id={})", e.id()));
                }
            }
        }
    }

    ValidationResult {
        ok: errors.is_empty(),
        errors,
        warnings,
    }
}

fn run_calc() -> Result<(), Box<dyn std::error::Error>> {
    let input = read_input()?;
    let report = run(input)?;

    println!("{}", serde_json::to_string_pretty(&report)?);
    Ok(())
}

fn run_validate() -> Result<(), Box<dyn std::error::Error>> {
    let input = read_input()?;
    let result = validate_input(&input);
    println!("{}", serde_json::to_string_pretty(&result)?);
    if result.ok {
        Ok(())
    } else {
        Err(io::Error::new(io::ErrorKind::InvalidInput, "validation failed").into())
    }
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();

    match cli.command.unwrap_or(Commands::Calc) {
        Commands::Calc => run_calc(),
        Commands::Validate => run_validate(),
        Commands::Version => {
            println!("{}", env!("CARGO_PKG_VERSION"));
            Ok(())
        }
    }
}
