#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CostMethod {
    MovingAverage,
    TotalAverage,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Config {
    pub method: CostMethod,
    pub tax_year: i32,
}
