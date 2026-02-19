pub mod config;
pub mod engine;
pub mod event;
pub mod report;

use config::{Config, CostMethod};
use event::Event;
use report::Report;

pub fn calculate(config: Config, events: &[Event]) -> Report {
    match config.method {
        CostMethod::MovingAverage => engine::moving_average::run(events, config.tax_year),
        CostMethod::TotalAverage => engine::total_average::run(events, config.tax_year),
    }
}
