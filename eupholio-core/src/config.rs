use std::collections::HashMap;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CostMethod {
    MovingAverage,
    TotalAverage,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RoundingMode {
    HalfUp,
    Down,
    HalfEven,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RoundingTiming {
    ReportOnly,
    PerEvent,
    PerYear,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct RoundRule {
    pub scale: u32,
    pub mode: RoundingMode,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct RoundingPolicy {
    pub currency: HashMap<String, RoundRule>,
    pub unit_price: RoundRule,
    pub quantity: RoundRule,
    pub timing: RoundingTiming,
}

impl Default for RoundingPolicy {
    fn default() -> Self {
        let mut currency = HashMap::new();
        currency.insert(
            "JPY".to_string(),
            RoundRule {
                scale: 0,
                mode: RoundingMode::HalfUp,
            },
        );
        Self {
            currency,
            unit_price: RoundRule {
                scale: 8,
                mode: RoundingMode::HalfUp,
            },
            quantity: RoundRule {
                scale: 8,
                mode: RoundingMode::HalfUp,
            },
            timing: RoundingTiming::ReportOnly,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Config {
    pub method: CostMethod,
    pub tax_year: i32,
    pub rounding: RoundingPolicy,
}
