use std::collections::{HashMap, HashSet};

use rust_decimal::{Decimal, RoundingStrategy};

use crate::config::{RoundRule, RoundingMode, RoundingPolicy};
use crate::event::{Event, TransferDirection};
use crate::report::{Position, Report, Warning};

pub fn run(events: &[Event], tax_year: i32) -> Report {
    run_inner(events, tax_year, None)
}

pub fn run_per_event(events: &[Event], tax_year: i32, policy: &RoundingPolicy) -> Report {
    run_inner(events, tax_year, Some(policy))
}

fn run_inner(events: &[Event], tax_year: i32, per_event_rounding: Option<&RoundingPolicy>) -> Report {
    let mut positions: HashMap<String, Position> = HashMap::new();
    let mut realized = Decimal::ZERO;
    let mut income = Decimal::ZERO;
    let mut diagnostics = Vec::new();
    let mut seen_ids = HashSet::new();

    for e in events {
        if !seen_ids.insert(e.id().to_string()) {
            diagnostics.push(Warning::DuplicateEventId {
                id: e.id().to_string(),
            });
            continue;
        }

        if e.year() != tax_year {
            diagnostics.push(Warning::YearMismatch {
                event_year: e.year(),
                tax_year,
            });
        }

        match e {
            Event::Acquire {
                asset,
                qty,
                jpy_cost,
                ..
            } => {
                let p = positions.entry(asset.clone()).or_insert(Position {
                    qty: Decimal::ZERO,
                    avg_cost_jpy_per_unit: Decimal::ZERO,
                });
                if *qty > Decimal::ZERO {
                    let total_cost = p.qty * p.avg_cost_jpy_per_unit + *jpy_cost;
                    let new_qty = p.qty + *qty;
                    p.avg_cost_jpy_per_unit = if new_qty.is_zero() {
                        Decimal::ZERO
                    } else {
                        total_cost / new_qty
                    };
                    p.qty = new_qty;
                }
            }
            Event::Dispose {
                asset,
                qty,
                jpy_proceeds,
                ..
            } => {
                let p = positions.entry(asset.clone()).or_insert(Position {
                    qty: Decimal::ZERO,
                    avg_cost_jpy_per_unit: Decimal::ZERO,
                });
                let cost = *qty * p.avg_cost_jpy_per_unit;
                realized += *jpy_proceeds - cost;
                p.qty -= *qty;
                if p.qty < Decimal::ZERO {
                    diagnostics.push(Warning::NegativePosition {
                        asset: asset.clone(),
                    });
                }
            }
            Event::Income {
                asset,
                qty,
                jpy_value,
                ..
            } => {
                income += *jpy_value;
                let p = positions.entry(asset.clone()).or_insert(Position {
                    qty: Decimal::ZERO,
                    avg_cost_jpy_per_unit: Decimal::ZERO,
                });
                if *qty > Decimal::ZERO {
                    let total_cost = p.qty * p.avg_cost_jpy_per_unit + *jpy_value;
                    let new_qty = p.qty + *qty;
                    p.avg_cost_jpy_per_unit = if new_qty.is_zero() {
                        Decimal::ZERO
                    } else {
                        total_cost / new_qty
                    };
                    p.qty = new_qty;
                }
            }
            Event::Transfer {
                asset,
                qty,
                direction,
                ..
            } => {
                let p = positions.entry(asset.clone()).or_insert(Position {
                    qty: Decimal::ZERO,
                    avg_cost_jpy_per_unit: Decimal::ZERO,
                });
                match direction {
                    TransferDirection::In => p.qty += *qty,
                    TransferDirection::Out => p.qty -= *qty,
                }
                if p.qty < Decimal::ZERO {
                    diagnostics.push(Warning::NegativePosition {
                        asset: asset.clone(),
                    });
                }
            }
        }

        if let Some(policy) = per_event_rounding {
            apply_per_event_rounding(&mut positions, &mut realized, &mut income, policy);
        }
    }

    Report {
        positions,
        realized_pnl_jpy: realized,
        income_jpy: income,
        yearly_summary: None,
        diagnostics,
    }
}

fn apply_per_event_rounding(
    positions: &mut HashMap<String, Position>,
    realized: &mut Decimal,
    income: &mut Decimal,
    policy: &RoundingPolicy,
) {
    let jpy_rule = policy
        .currency
        .get("JPY")
        .copied()
        .unwrap_or(RoundRule {
            scale: 0,
            mode: RoundingMode::HalfUp,
        });

    *realized = round_by_rule(*realized, jpy_rule);
    *income = round_by_rule(*income, jpy_rule);

    for p in positions.values_mut() {
        p.qty = round_by_rule(p.qty, policy.quantity);
        p.avg_cost_jpy_per_unit = round_by_rule(p.avg_cost_jpy_per_unit, policy.unit_price);
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
