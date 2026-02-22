pub mod config;
pub mod engine;
pub mod event;
pub mod report;

use std::collections::HashMap;

use config::{Config, CostMethod, RoundRule, RoundingMode, RoundingPolicy, RoundingTiming};
use event::Event;
use report::{CarryIn, Report};
use rust_decimal::{Decimal, RoundingStrategy};

pub fn calculate(config: Config, events: &[Event]) -> Report {
    let mut effective_rounding = config.rounding.clone();

    let mut report = match (config.method, config.rounding.timing) {
        (CostMethod::MovingAverage, RoundingTiming::PerEvent) => {
            engine::moving_average::run_per_event(events, config.tax_year, &config.rounding)
        }
        (CostMethod::TotalAverage, RoundingTiming::PerEvent) => {
            engine::total_average::run_per_event(events, config.tax_year, &config.rounding)
        }
        (CostMethod::TotalAverage, RoundingTiming::PerYear) => {
            engine::total_average::run_per_year(events, config.tax_year, &config.rounding)
        }
        (CostMethod::MovingAverage, RoundingTiming::PerYear) => {
            // moving_average + per_year is unsupported in CLI validation.
            // Keep library behavior deterministic: fall back to report-only rounding
            // instead of producing unrounded output.
            effective_rounding.timing = RoundingTiming::ReportOnly;
            engine::moving_average::run(events, config.tax_year)
        }
        (CostMethod::MovingAverage, _) => engine::moving_average::run(events, config.tax_year),
        (CostMethod::TotalAverage, _) => engine::total_average::run(events, config.tax_year),
    };
    apply_rounding(&mut report, &effective_rounding);
    report
}

pub fn calculate_total_average_with_carry(
    tax_year: i32,
    events: &[Event],
    carry_in: &HashMap<String, CarryIn>,
) -> Report {
    calculate_total_average_with_carry_and_rounding(
        tax_year,
        events,
        carry_in,
        RoundingPolicy::default(),
    )
}

pub fn calculate_total_average_with_carry_and_rounding(
    tax_year: i32,
    events: &[Event],
    carry_in: &HashMap<String, CarryIn>,
    rounding: RoundingPolicy,
) -> Report {
    let mut report = match rounding.timing {
        RoundingTiming::PerEvent => {
            engine::total_average::run_with_carry_per_event(events, tax_year, carry_in, &rounding)
        }
        RoundingTiming::PerYear => {
            engine::total_average::run_with_carry_per_year(events, tax_year, carry_in, &rounding)
        }
        _ => engine::total_average::run_with_carry(events, tax_year, carry_in),
    };
    apply_rounding(&mut report, &rounding);
    report
}

fn apply_rounding(report: &mut Report, policy: &RoundingPolicy) {
    if policy.timing != RoundingTiming::ReportOnly {
        return;
    }

    let jpy_rule = policy.currency.get("JPY").copied().unwrap_or(RoundRule {
        scale: 0,
        mode: RoundingMode::HalfUp,
    });

    report.realized_pnl_jpy = round_by_rule(report.realized_pnl_jpy, jpy_rule);
    report.income_jpy = round_by_rule(report.income_jpy, jpy_rule);

    for p in report.positions.values_mut() {
        p.qty = round_by_rule(p.qty, policy.quantity);
        p.avg_cost_jpy_per_unit = round_by_rule(p.avg_cost_jpy_per_unit, policy.unit_price);
    }

    if let Some(summary) = report.yearly_summary.as_mut() {
        for y in summary.by_asset.values_mut() {
            y.carry_in_qty = round_by_rule(y.carry_in_qty, policy.quantity);
            y.carry_in_cost = round_by_rule(y.carry_in_cost, jpy_rule);
            y.total_acquired_qty = round_by_rule(y.total_acquired_qty, policy.quantity);
            y.total_acquired_cost = round_by_rule(y.total_acquired_cost, jpy_rule);
            y.total_disposed_qty = round_by_rule(y.total_disposed_qty, policy.quantity);
            y.total_disposed_proceeds = round_by_rule(y.total_disposed_proceeds, jpy_rule);
            y.average_cost_per_unit = round_by_rule(y.average_cost_per_unit, policy.unit_price);
            y.realized_pnl_jpy = round_by_rule(y.realized_pnl_jpy, jpy_rule);
            y.carry_out_qty = round_by_rule(y.carry_out_qty, policy.quantity);
            y.carry_out_cost = round_by_rule(y.carry_out_qty * y.average_cost_per_unit, jpy_rule);
        }
    }
}

fn round_by_rule(v: Decimal, rule: RoundRule) -> Decimal {
    v.round_dp_with_strategy(rule.scale, to_strategy(rule.mode))
}

fn to_strategy(mode: RoundingMode) -> RoundingStrategy {
    match mode {
        RoundingMode::HalfUp => RoundingStrategy::MidpointAwayFromZero,
        RoundingMode::Down => RoundingStrategy::ToZero,
        RoundingMode::HalfEven => RoundingStrategy::MidpointNearestEven,
    }
}
