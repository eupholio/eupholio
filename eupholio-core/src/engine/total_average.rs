use std::collections::{HashMap, HashSet};

use rust_decimal::Decimal;

use crate::event::Event;
use crate::report::{CarryIn, Position, Report, Warning, YearlyAssetSummary, YearlySummary};

#[derive(Debug, Clone)]
struct Bucket {
    carry_in_qty: Decimal,
    carry_in_cost: Decimal,
    total_acquired_qty: Decimal,
    total_acquired_cost: Decimal,
    total_disposed_qty: Decimal,
    total_disposed_proceeds: Decimal,
}

impl Bucket {
    fn new() -> Self {
        Self {
            carry_in_qty: Decimal::ZERO,
            carry_in_cost: Decimal::ZERO,
            total_acquired_qty: Decimal::ZERO,
            total_acquired_cost: Decimal::ZERO,
            total_disposed_qty: Decimal::ZERO,
            total_disposed_proceeds: Decimal::ZERO,
        }
    }
}

pub fn run(events: &[Event], tax_year: i32) -> Report {
    run_with_carry(events, tax_year, &HashMap::new())
}

pub fn run_with_carry(events: &[Event], tax_year: i32, carry_in: &HashMap<String, CarryIn>) -> Report {
    let mut buckets: HashMap<String, Bucket> = HashMap::new();
    let mut diagnostics = Vec::new();
    let mut income = Decimal::ZERO;
    let mut seen_ids = HashSet::new();

    for (asset, c) in carry_in {
        let b = buckets.entry(asset.clone()).or_insert_with(Bucket::new);
        b.carry_in_qty = c.qty;
        b.carry_in_cost = c.cost;
        diagnostics.push(Warning::YearBoundaryCarry {
            asset: asset.clone(),
        });
    }

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
            continue;
        }

        match e {
            Event::Acquire {
                asset,
                qty,
                jpy_cost,
                ..
            } => {
                let b = buckets.entry(asset.clone()).or_insert_with(Bucket::new);
                b.total_acquired_qty += *qty;
                b.total_acquired_cost += *jpy_cost;
            }
            Event::Dispose {
                asset,
                qty,
                jpy_proceeds,
                ..
            } => {
                let b = buckets.entry(asset.clone()).or_insert_with(Bucket::new);
                b.total_disposed_qty += *qty;
                b.total_disposed_proceeds += *jpy_proceeds;
            }
            Event::Income {
                asset,
                qty,
                jpy_value,
                ..
            } => {
                income += *jpy_value;
                let b = buckets.entry(asset.clone()).or_insert_with(Bucket::new);
                b.total_acquired_qty += *qty;
                b.total_acquired_cost += *jpy_value;
            }
            Event::Transfer { .. } => {}
        }
    }

    let mut positions = HashMap::new();
    let mut by_asset = HashMap::new();
    let mut realized_total = Decimal::ZERO;

    for (asset, b) in buckets {
        let denom_qty = b.carry_in_qty + b.total_acquired_qty;
        let denom_cost = b.carry_in_cost + b.total_acquired_cost;
        let avg = if denom_qty.is_zero() {
            Decimal::ZERO
        } else {
            denom_cost / denom_qty
        };

        let realized = b.total_disposed_proceeds - b.total_disposed_qty * avg;
        realized_total += realized;

        let carry_out_qty = denom_qty - b.total_disposed_qty;
        let carry_out_cost = carry_out_qty * avg;

        if carry_out_qty < Decimal::ZERO {
            diagnostics.push(Warning::NegativePosition {
                asset: asset.clone(),
            });
        }

        positions.insert(
            asset.clone(),
            Position {
                qty: carry_out_qty,
                avg_cost_jpy_per_unit: avg,
            },
        );

        by_asset.insert(
            asset,
            YearlyAssetSummary {
                carry_in_qty: b.carry_in_qty,
                carry_in_cost: b.carry_in_cost,
                total_acquired_qty: b.total_acquired_qty,
                total_acquired_cost: b.total_acquired_cost,
                total_disposed_qty: b.total_disposed_qty,
                total_disposed_proceeds: b.total_disposed_proceeds,
                average_cost_per_unit: avg,
                realized_pnl_jpy: realized,
                carry_out_qty,
                carry_out_cost,
            },
        );
    }

    Report {
        positions,
        realized_pnl_jpy: realized_total,
        income_jpy: income,
        yearly_summary: Some(YearlySummary { tax_year, by_asset }),
        diagnostics,
    }
}
