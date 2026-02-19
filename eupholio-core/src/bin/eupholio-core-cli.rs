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

#[derive(Debug, Clone, Copy, Serialize)]
#[serde(rename_all = "snake_case")]
enum Level {
    Error,
    Warning,
}

#[derive(Debug, Clone, Copy)]
enum ValidationCode {
    InvalidMethod,
    EmptyEvents,
    UnusualTaxYear,
    CarryInIgnoredForMoving,
    NegativeCarryInQty,
    NegativeCarryInCost,
    CarryInCostWithZeroQty,
    RoundingJpyScaleTooLarge,
    RoundingUnitPriceScaleTooLarge,
    RoundingQuantityScaleTooLarge,
    RoundingTimingNotFullyImplemented,
    EventYearMismatch,
    DuplicateEventId,
    AcquireQtyNonPositive,
    AcquireCostNegative,
    DisposeQtyNonPositive,
    DisposeProceedsNegative,
    IncomeQtyNonPositive,
    IncomeValueNegative,
    TransferQtyNonPositive,
}

impl ValidationCode {
    fn as_str(self) -> &'static str {
        match self {
            ValidationCode::InvalidMethod => "INVALID_METHOD",
            ValidationCode::EmptyEvents => "EMPTY_EVENTS",
            ValidationCode::UnusualTaxYear => "UNUSUAL_TAX_YEAR",
            ValidationCode::CarryInIgnoredForMoving => "CARRY_IN_IGNORED_FOR_MOVING",
            ValidationCode::NegativeCarryInQty => "NEGATIVE_CARRY_IN_QTY",
            ValidationCode::NegativeCarryInCost => "NEGATIVE_CARRY_IN_COST",
            ValidationCode::CarryInCostWithZeroQty => "CARRY_IN_COST_WITH_ZERO_QTY",
            ValidationCode::RoundingJpyScaleTooLarge => "ROUNDING_JPY_SCALE_TOO_LARGE",
            ValidationCode::RoundingUnitPriceScaleTooLarge => "ROUNDING_UNIT_PRICE_SCALE_TOO_LARGE",
            ValidationCode::RoundingQuantityScaleTooLarge => "ROUNDING_QUANTITY_SCALE_TOO_LARGE",
            ValidationCode::RoundingTimingNotFullyImplemented => "ROUNDING_TIMING_NOT_FULLY_IMPLEMENTED",
            ValidationCode::EventYearMismatch => "EVENT_YEAR_MISMATCH",
            ValidationCode::DuplicateEventId => "DUPLICATE_EVENT_ID",
            ValidationCode::AcquireQtyNonPositive => "ACQUIRE_QTY_NON_POSITIVE",
            ValidationCode::AcquireCostNegative => "ACQUIRE_COST_NEGATIVE",
            ValidationCode::DisposeQtyNonPositive => "DISPOSE_QTY_NON_POSITIVE",
            ValidationCode::DisposeProceedsNegative => "DISPOSE_PROCEEDS_NEGATIVE",
            ValidationCode::IncomeQtyNonPositive => "INCOME_QTY_NON_POSITIVE",
            ValidationCode::IncomeValueNegative => "INCOME_VALUE_NEGATIVE",
            ValidationCode::TransferQtyNonPositive => "TRANSFER_QTY_NON_POSITIVE",
        }
    }
}

#[derive(Debug, Serialize)]
struct ValidationIssue {
    code: &'static str,
    level: Level,
    message: String,
}

#[derive(Debug, Serialize)]
struct ValidationResult {
    ok: bool,
    errors: Vec<String>,
    warnings: Vec<String>,
    issues: Vec<ValidationIssue>,
}

impl ValidationResult {
    fn new() -> Self {
        Self {
            ok: true,
            errors: Vec::new(),
            warnings: Vec::new(),
            issues: Vec::new(),
        }
    }

    fn push_error(&mut self, code: ValidationCode, message: String) {
        self.ok = false;
        self.errors.push(message.clone());
        self.issues.push(ValidationIssue {
            code: code.as_str(),
            level: Level::Error,
            message,
        });
    }

    fn push_warning(&mut self, code: ValidationCode, message: String) {
        self.warnings.push(message.clone());
        self.issues.push(ValidationIssue {
            code: code.as_str(),
            level: Level::Warning,
            message,
        });
    }
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
    let mut out = ValidationResult::new();

    let method = parse_method(&input.method);
    if let Err(e) = &method {
        out.push_error(ValidationCode::InvalidMethod, e.clone());
    }

    if input.events.is_empty() {
        out.push_warning(ValidationCode::EmptyEvents, "events is empty".to_string());
    }

    if input.tax_year < 2000 || input.tax_year > 3000 {
        out.push_warning(
            ValidationCode::UnusualTaxYear,
            format!("tax_year looks unusual: {}", input.tax_year),
        );
    }

    if method.as_ref().ok() == Some(&CostMethod::MovingAverage) && !input.carry_in.is_empty() {
        out.push_warning(
            ValidationCode::CarryInIgnoredForMoving,
            "carry_in is ignored for moving_average".to_string(),
        );
    }

    for (asset, c) in &input.carry_in {
        if c.qty < Decimal::ZERO {
            out.push_error(
                ValidationCode::NegativeCarryInQty,
                format!("carry_in qty must be >= 0 (asset={asset})"),
            );
        }
        if c.cost < Decimal::ZERO {
            out.push_error(
                ValidationCode::NegativeCarryInCost,
                format!("carry_in cost must be >= 0 (asset={asset})"),
            );
        }
        if c.qty == Decimal::ZERO && c.cost > Decimal::ZERO {
            out.push_warning(
                ValidationCode::CarryInCostWithZeroQty,
                format!("carry_in has cost without qty (asset={asset})"),
            );
        }
    }

    if let Some(rounding) = &input.rounding {
        if let Some(jpy) = rounding.currency.get("JPY") {
            if jpy.scale > 18 {
                out.push_error(
                    ValidationCode::RoundingJpyScaleTooLarge,
                    "rounding.currency.JPY.scale must be <= 18".to_string(),
                );
            }
        }
        if rounding.unit_price.scale > 18 {
            out.push_error(
                ValidationCode::RoundingUnitPriceScaleTooLarge,
                "rounding.unit_price.scale must be <= 18".to_string(),
            );
        }
        if rounding.quantity.scale > 18 {
            out.push_error(
                ValidationCode::RoundingQuantityScaleTooLarge,
                "rounding.quantity.scale must be <= 18".to_string(),
            );
        }
        if rounding.timing == RoundingTiming::PerYear {
            out.push_warning(
                ValidationCode::RoundingTimingNotFullyImplemented,
                "rounding.timing=per_year is not fully implemented yet".to_string(),
            );
        }
    }

    let mut ids = HashSet::new();
    for e in &input.events {
        if e.year() != input.tax_year {
            out.push_warning(
                ValidationCode::EventYearMismatch,
                format!("event year {} does not match tax_year {} (id={})", e.year(), input.tax_year, e.id()),
            );
        }

        let id = e.id().to_string();
        if !ids.insert(id.clone()) {
            out.push_error(ValidationCode::DuplicateEventId, format!("duplicate event id: {id}"));
        }

        match e {
            Event::Acquire { qty, jpy_cost, .. } => {
                if *qty <= Decimal::ZERO {
                    out.push_error(
                        ValidationCode::AcquireQtyNonPositive,
                        format!("Acquire qty must be > 0 (id={})", e.id()),
                    );
                }
                if *jpy_cost < Decimal::ZERO {
                    out.push_error(
                        ValidationCode::AcquireCostNegative,
                        format!("Acquire jpy_cost must be >= 0 (id={})", e.id()),
                    );
                }
            }
            Event::Dispose {
                qty,
                jpy_proceeds,
                ..
            } => {
                if *qty <= Decimal::ZERO {
                    out.push_error(
                        ValidationCode::DisposeQtyNonPositive,
                        format!("Dispose qty must be > 0 (id={})", e.id()),
                    );
                }
                if *jpy_proceeds < Decimal::ZERO {
                    out.push_error(
                        ValidationCode::DisposeProceedsNegative,
                        format!("Dispose jpy_proceeds must be >= 0 (id={})", e.id()),
                    );
                }
            }
            Event::Income { qty, jpy_value, .. } => {
                if *qty <= Decimal::ZERO {
                    out.push_error(
                        ValidationCode::IncomeQtyNonPositive,
                        format!("Income qty must be > 0 (id={})", e.id()),
                    );
                }
                if *jpy_value < Decimal::ZERO {
                    out.push_error(
                        ValidationCode::IncomeValueNegative,
                        format!("Income jpy_value must be >= 0 (id={})", e.id()),
                    );
                }
            }
            Event::Transfer { qty, .. } => {
                if *qty <= Decimal::ZERO {
                    out.push_error(
                        ValidationCode::TransferQtyNonPositive,
                        format!("Transfer qty must be > 0 (id={})", e.id()),
                    );
                }
            }
        }
    }

    out
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
