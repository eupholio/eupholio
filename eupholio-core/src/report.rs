use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Position {
    pub qty: Decimal,
    pub avg_cost_jpy_per_unit: Decimal,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct CarryIn {
    pub qty: Decimal,
    pub cost: Decimal,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct YearlyAssetSummary {
    pub carry_in_qty: Decimal,
    pub carry_in_cost: Decimal,
    pub total_acquired_qty: Decimal,
    pub total_acquired_cost: Decimal,
    pub total_disposed_qty: Decimal,
    pub total_disposed_proceeds: Decimal,
    pub average_cost_per_unit: Decimal,
    pub realized_pnl_jpy: Decimal,
    pub carry_out_qty: Decimal,
    pub carry_out_cost: Decimal,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct YearlySummary {
    pub tax_year: i32,
    pub by_asset: HashMap<String, YearlyAssetSummary>,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum Warning {
    DuplicateEventId { id: String },
    NegativePosition { asset: String },
    YearMismatch { event_year: i32, tax_year: i32 },
    YearBoundaryCarry { asset: String },
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Report {
    pub positions: HashMap<String, Position>,
    pub realized_pnl_jpy: Decimal,
    pub income_jpy: Decimal,
    pub yearly_summary: Option<YearlySummary>,
    pub diagnostics: Vec<Warning>,
}
