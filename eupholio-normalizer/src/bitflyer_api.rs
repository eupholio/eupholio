use chrono::{DateTime, Utc};
use eupholio_core::event::Event;
use hmac::{Hmac, Mac};
use rust_decimal::Decimal;
use serde::Deserialize;
use sha2::Sha256;

use crate::bitflyer::{NormalizeDiagnostic, NormalizeResult};

type HmacSha256 = Hmac<Sha256>;

#[derive(Debug, Clone, Deserialize, PartialEq)]
pub struct Execution {
    pub id: i64,
    pub side: String,
    pub price: Decimal,
    pub size: Decimal,
    pub exec_date: DateTime<Utc>,
    #[serde(default)]
    pub commission: Option<Decimal>,
}

pub fn sign_request(secret: &str, timestamp: &str, method: &str, path: &str, body: &str) -> String {
    let payload = format!("{}{}{}{}", timestamp, method, path, body);
    let mut mac = HmacSha256::new_from_slice(secret.as_bytes()).expect("hmac init should succeed");
    mac.update(payload.as_bytes());
    let bytes = mac.finalize().into_bytes();
    hex::encode(bytes)
}

pub fn normalize_executions(
    executions: &[Execution],
    product_code: &str,
) -> Result<NormalizeResult, String> {
    let (base_asset, quote_asset) = split_product(product_code)?;
    if quote_asset != "JPY" {
        return Err(format!(
            "unsupported quote asset '{}', only JPY is supported in phase-1",
            quote_asset
        ));
    }

    let mut events = Vec::new();
    let mut diagnostics = Vec::new();

    for (i, ex) in executions.iter().enumerate() {
        let row = i + 1;
        if ex.size <= Decimal::ZERO {
            return Err(format!("row {}: size must be > 0, got {}", row, ex.size));
        }
        if ex.price <= Decimal::ZERO {
            return Err(format!("row {}: price must be > 0, got {}", row, ex.price));
        }

        let fee_base = ex.commission.unwrap_or(Decimal::ZERO).abs();
        let fee_jpy = fee_base * ex.price;
        let jpy_total = ex.price * ex.size;
        let side = ex.side.to_ascii_uppercase();

        match side.as_str() {
            "BUY" => {
                let net_qty = ex.size - fee_base;
                if net_qty <= Decimal::ZERO {
                    return Err(format!(
                        "row {}: buy qty must be > 0 after fee, got {}",
                        row, net_qty
                    ));
                }
                events.push(Event::Acquire {
                    id: format!("bfexec-{}:acquire", ex.id),
                    asset: base_asset.to_string(),
                    qty: net_qty,
                    jpy_cost: jpy_total + fee_jpy,
                    ts: ex.exec_date,
                });
            }
            "SELL" => {
                events.push(Event::Dispose {
                    id: format!("bfexec-{}:dispose", ex.id),
                    asset: base_asset.to_string(),
                    qty: ex.size,
                    jpy_proceeds: jpy_total - fee_jpy,
                    ts: ex.exec_date,
                });
            }
            other => diagnostics.push(NormalizeDiagnostic {
                row,
                reason: format!(
                    "unsupported side: side='{}', execution_id='{}'",
                    other, ex.id
                ),
            }),
        }
    }

    Ok(NormalizeResult {
        events,
        diagnostics,
    })
}

fn split_product(product_code: &str) -> Result<(String, String), String> {
    let mut parts = product_code.split('_');
    let base = parts
        .next()
        .ok_or_else(|| format!("invalid product_code '{}': missing base", product_code))?;
    let quote = parts
        .next()
        .ok_or_else(|| format!("invalid product_code '{}': missing quote", product_code))?;
    if parts.next().is_some() {
        return Err(format!("invalid product_code '{}'", product_code));
    }
    Ok((
        base.trim().to_ascii_uppercase(),
        quote.trim().to_ascii_uppercase(),
    ))
}
