pub mod config;
pub mod engine;
pub mod event;
pub mod report;

use std::collections::HashMap;

use config::{Config, CostMethod};
use event::Event;
use report::{CarryIn, Report};

pub fn calculate(config: Config, events: &[Event]) -> Report {
    match config.method {
        CostMethod::MovingAverage => engine::moving_average::run(events, config.tax_year),
        CostMethod::TotalAverage => engine::total_average::run(events, config.tax_year),
    }
}

pub fn calculate_total_average_with_carry(
    tax_year: i32,
    events: &[Event],
    carry_in: &HashMap<String, CarryIn>,
) -> Report {
    engine::total_average::run_with_carry(events, tax_year, carry_in)
}
