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
    let mut report = match config.method {
        CostMethod::MovingAverage => engine::moving_average::run(events, config.tax_year),
        CostMethod::TotalAverage => engine::total_average::run(events, config.tax_year),
    };
    apply_rounding(&mut report, &config.rounding);
    report
}

pub fn calculate_total_average_with_carry(
    tax_year: i32,
    events: &[Event],
    carry_in: &HashMap<String, CarryIn>,
) -> Report {
    let mut report = engine::total_average::run_with_carry(events, tax_year, carry_in);
    apply_rounding(&mut report, &RoundingPolicy::default());
    report
}

fn apply_rounding(report: &mut Report, policy: &RoundingPolicy) {
    if policy.timing != RoundingTiming::ReportOnly {
        return;
    }

    let jpy_rule = policy
        .currency
        .get("JPY")
        .copied()
        .unwrap_or(RoundRule {
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
            y.carry_out_cost = round_by_rule(y.carry_out_cost, jpy_rule);
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
