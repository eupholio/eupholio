use std::collections::{HashMap, HashSet};

use rust_decimal::Decimal;

use crate::event::{Event, TransferDirection};
use crate::report::{Position, Report, Warning};

pub fn run(events: &[Event], tax_year: i32) -> Report {
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
    }

    Report {
        positions,
        realized_pnl_jpy: realized,
        income_jpy: income,
        yearly_summary: None,
        diagnostics,
    }
}
