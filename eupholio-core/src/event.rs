use chrono::{DateTime, Datelike, Utc};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TransferDirection {
    In,
    Out,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum Event {
    Acquire {
        id: String,
        asset: String,
        qty: Decimal,
        jpy_cost: Decimal,
        ts: DateTime<Utc>,
    },
    Dispose {
        id: String,
        asset: String,
        qty: Decimal,
        jpy_proceeds: Decimal,
        ts: DateTime<Utc>,
    },
    Income {
        id: String,
        asset: String,
        qty: Decimal,
        jpy_value: Decimal,
        ts: DateTime<Utc>,
    },
    Transfer {
        id: String,
        asset: String,
        qty: Decimal,
        direction: TransferDirection,
        ts: DateTime<Utc>,
    },
}

impl Event {
    pub fn id(&self) -> &str {
        match self {
            Event::Acquire { id, .. }
            | Event::Dispose { id, .. }
            | Event::Income { id, .. }
            | Event::Transfer { id, .. } => id,
        }
    }

    pub fn year(&self) -> i32 {
        match self {
            Event::Acquire { ts, .. }
            | Event::Dispose { ts, .. }
            | Event::Income { ts, .. }
            | Event::Transfer { ts, .. } => ts.year(),
        }
    }
}

