use chrono::{DateTime, Utc};
use csv::StringRecord;
use eupholio_core::event::{Event, TransferDirection};
use rust_decimal::Decimal;
use std::collections::HashMap;
use std::str::FromStr;

const TRADE_HEADERS: &[&str] = &[
    "Date",
    "Market",
    "Category",
    "Type",
    "Price",
    "Amount",
    "Total",
    "Fee",
    "Order Number",
    "Base Total Less Fee",
    "Quote Total Less Fee",
    "Fee Currency",
    "Fee Total",
];

const DEPOSIT_HEADERS: &[&str] = &["Date", "Currency", "Amount", "Address", "Status"];
const WITHDRAWAL_HEADERS: &[&str] = &[
    "Date",
    "Currency",
    "Amount",
    "Fee Deducted",
    "Amount - Fee",
    "Address",
    "Status",
];
const DISTRIBUTION_HEADERS: &[&str] = &["date", "currency", "amount", "wallet"];

const TYPE_BUY: &str = "Buy";
const TYPE_SELL: &str = "Sell";
const MAX_DIAGNOSTIC_VALUE_LEN: usize = 64;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct NormalizeDiagnostic {
    pub row: usize,
    pub reason: String,
}

#[derive(Debug, Clone, PartialEq)]
pub struct NormalizeResult {
    pub events: Vec<Event>,
    pub diagnostics: Vec<NormalizeDiagnostic>,
}

pub fn normalize_trade_history_csv(raw: &str) -> Result<NormalizeResult, String> {
    let mut rdr = csv::ReaderBuilder::new()
        .has_headers(true)
        .trim(csv::Trim::All)
        .from_reader(raw.as_bytes());

    let headers = rdr
        .headers()
        .map_err(|e| format!("invalid csv header: {}", e))?
        .clone();
    validate_headers(&headers, TRADE_HEADERS)?;
    let idx = build_header_index(&headers);

    let mut events = Vec::new();
    let mut diagnostics = Vec::new();

    for (i, rec) in rdr.records().enumerate() {
        let row = i + 2;
        let rec = rec.map_err(|e| format!("row {}: invalid csv row: {}", row, e))?;
        if rec.iter().all(|v| v.trim().is_empty()) {
            continue;
        }

        let trade_type = get(&idx, &rec, "Type")?;
        if trade_type != TYPE_BUY && trade_type != TYPE_SELL {
            diagnostics.push(NormalizeDiagnostic {
                row,
                reason: format!(
                    "unsupported trade type: type='{}', order_number='{}'",
                    sanitize_diagnostic_value(trade_type),
                    sanitize_diagnostic_value(get(&idx, &rec, "Order Number")?)
                ),
            });
            continue;
        }

        let ts = parse_datetime(get(&idx, &rec, "Date")?, "%Y-%m-%d %H:%M:%S")?;
        let market = get(&idx, &rec, "Market")?;
        let (trading_asset, payment_asset) = parse_market(market)?;

        let price = parse_decimal(get(&idx, &rec, "Price")?)?;
        let amount = parse_decimal(get(&idx, &rec, "Amount")?)?;
        let total = parse_decimal(get(&idx, &rec, "Total")?)?;
        let base_total_less_fee = parse_decimal(get(&idx, &rec, "Base Total Less Fee")?)?.abs();
        let quote_total_less_fee = parse_decimal(get(&idx, &rec, "Quote Total Less Fee")?)?.abs();

        if payment_asset != "JPY" {
            return Err(format!(
                "row {}: unsupported payment asset '{}', only JPY is supported",
                row, payment_asset
            ));
        }
        if price <= Decimal::ZERO {
            return Err(format!("row {}: price must be > 0, got {}", row, price));
        }

        let order_number = get(&idx, &rec, "Order Number")?;

        let event = if trade_type == TYPE_BUY {
            if quote_total_less_fee <= Decimal::ZERO {
                return Err(format!(
                    "row {}: buy qty must be > 0 after fee, got {}",
                    row, quote_total_less_fee
                ));
            }
            Event::Acquire {
                id: format!("{}:acquire", order_number),
                asset: trading_asset.to_string(),
                qty: quote_total_less_fee,
                jpy_cost: total,
                ts,
            }
        } else {
            if amount <= Decimal::ZERO {
                return Err(format!("row {}: sell qty must be > 0, got {}", row, amount));
            }
            Event::Dispose {
                id: format!("{}:dispose", order_number),
                asset: trading_asset.to_string(),
                qty: amount,
                jpy_proceeds: base_total_less_fee,
                ts,
            }
        };
        events.push(event);
    }

    Ok(NormalizeResult {
        events,
        diagnostics,
    })
}

pub fn normalize_deposit_history_csv(raw: &str) -> Result<NormalizeResult, String> {
    let mut rdr = csv::ReaderBuilder::new()
        .has_headers(true)
        .trim(csv::Trim::All)
        .from_reader(raw.as_bytes());

    let headers = rdr
        .headers()
        .map_err(|e| format!("invalid csv header: {}", e))?
        .clone();
    validate_headers(&headers, DEPOSIT_HEADERS)?;
    let idx = build_header_index(&headers);

    let mut events = Vec::new();
    for (i, rec) in rdr.records().enumerate() {
        let row = i + 2;
        let rec = rec.map_err(|e| format!("row {}: invalid csv row: {}", row, e))?;
        if rec.iter().all(|v| v.trim().is_empty()) {
            continue;
        }

        let ts = parse_datetime(get(&idx, &rec, "Date")?, "%Y-%m-%d %H:%M:%S")?;
        let asset = get(&idx, &rec, "Currency")?.to_ascii_uppercase();
        let qty = parse_decimal(get(&idx, &rec, "Amount")?)?;
        if qty <= Decimal::ZERO {
            return Err(format!("row {}: deposit qty must be > 0, got {}", row, qty));
        }

        events.push(Event::Transfer {
            id: format!("poloniex-deposit-{}", row),
            asset,
            qty,
            direction: TransferDirection::In,
            ts,
        });
    }

    Ok(NormalizeResult {
        events,
        diagnostics: Vec::new(),
    })
}

pub fn normalize_withdrawal_history_csv(raw: &str) -> Result<NormalizeResult, String> {
    let mut rdr = csv::ReaderBuilder::new()
        .has_headers(true)
        .trim(csv::Trim::All)
        .from_reader(raw.as_bytes());

    let headers = rdr
        .headers()
        .map_err(|e| format!("invalid csv header: {}", e))?
        .clone();
    validate_headers(&headers, WITHDRAWAL_HEADERS)?;
    let idx = build_header_index(&headers);

    let mut events = Vec::new();
    for (i, rec) in rdr.records().enumerate() {
        let row = i + 2;
        let rec = rec.map_err(|e| format!("row {}: invalid csv row: {}", row, e))?;
        if rec.iter().all(|v| v.trim().is_empty()) {
            continue;
        }

        let ts = parse_datetime(get(&idx, &rec, "Date")?, "%Y-%m-%d %H:%M:%S")?;
        let asset = get(&idx, &rec, "Currency")?.to_ascii_uppercase();

        let amount = parse_decimal(get(&idx, &rec, "Amount")?)?;
        if amount <= Decimal::ZERO {
            return Err(format!(
                "row {}: withdrawal qty must be > 0, got {}",
                row, amount
            ));
        }

        events.push(Event::Transfer {
            id: format!("poloniex-withdrawal-{}", row),
            asset: asset.clone(),
            qty: amount,
            direction: TransferDirection::Out,
            ts,
        });

        let fee = parse_decimal(get(&idx, &rec, "Fee Deducted")?)?.abs();
        if fee > Decimal::ZERO {
            events.push(Event::Transfer {
                id: format!("poloniex-withdrawal-fee-{}", row),
                asset,
                qty: fee,
                direction: TransferDirection::Out,
                ts,
            });
        }
    }

    Ok(NormalizeResult {
        events,
        diagnostics: Vec::new(),
    })
}

pub fn normalize_distribution_history_csv(raw: &str) -> Result<NormalizeResult, String> {
    let mut rdr = csv::ReaderBuilder::new()
        .has_headers(true)
        .trim(csv::Trim::All)
        .from_reader(raw.as_bytes());

    let headers = rdr
        .headers()
        .map_err(|e| format!("invalid csv header: {}", e))?
        .clone();
    validate_headers(&headers, DISTRIBUTION_HEADERS)?;
    let idx = build_header_index(&headers);

    let mut events = Vec::new();
    for (i, rec) in rdr.records().enumerate() {
        let row = i + 2;
        let rec = rec.map_err(|e| format!("row {}: invalid csv row: {}", row, e))?;
        if rec.iter().all(|v| v.trim().is_empty()) {
            continue;
        }

        let ts = parse_date(get(&idx, &rec, "date")?)?;
        let asset = get(&idx, &rec, "currency")?.to_ascii_uppercase();
        let qty = parse_decimal(get(&idx, &rec, "amount")?)?;
        if qty <= Decimal::ZERO {
            return Err(format!(
                "row {}: distribution qty must be > 0, got {}",
                row, qty
            ));
        }

        events.push(Event::Income {
            id: format!("poloniex-distribution-{}", row),
            asset,
            qty,
            jpy_value: Decimal::ZERO,
            ts,
        });
    }

    Ok(NormalizeResult {
        events,
        diagnostics: Vec::new(),
    })
}

fn validate_headers(headers: &StringRecord, required: &[&str]) -> Result<(), String> {
    for h in required {
        if !headers.iter().any(|x| x == *h) {
            return Err(format!("missing required header {}", h));
        }
    }
    Ok(())
}

fn build_header_index(headers: &StringRecord) -> HashMap<String, usize> {
    headers
        .iter()
        .enumerate()
        .map(|(i, col)| (col.to_string(), i))
        .collect()
}

fn get<'a>(
    idx: &HashMap<String, usize>,
    row: &'a StringRecord,
    key: &str,
) -> Result<&'a str, String> {
    let i = *idx
        .get(key)
        .ok_or_else(|| format!("missing required header {}", key))?;
    row.get(i)
        .map(str::trim)
        .ok_or_else(|| format!("missing required field {}", key))
}

fn parse_decimal(s: &str) -> Result<Decimal, String> {
    let v = s.replace(',', "");
    Decimal::from_str(&v).map_err(|e| format!("invalid decimal '{}': {}", s, e))
}

fn parse_datetime(s: &str, layout: &str) -> Result<DateTime<Utc>, String> {
    let dt = DateTime::parse_from_str(&format!("{} +0000", s), &format!("{} %z", layout))
        .map_err(|e| format!("invalid datetime '{}': {}", s, e))?;
    Ok(dt.with_timezone(&Utc))
}

fn parse_date(s: &str) -> Result<DateTime<Utc>, String> {
    parse_datetime(&format!("{} 00:00:00", s), "%Y-%m-%d %H:%M:%S")
}

fn parse_market(s: &str) -> Result<(String, String), String> {
    let mut parts = s.split('/');
    let base = parts
        .next()
        .ok_or_else(|| format!("invalid market field {}", sanitize_diagnostic_value(s)))?;
    let quote = parts
        .next()
        .ok_or_else(|| format!("invalid market field {}", sanitize_diagnostic_value(s)))?;
    if parts.next().is_some() {
        return Err(format!(
            "invalid market field {}",
            sanitize_diagnostic_value(s)
        ));
    }
    Ok((
        base.trim().to_ascii_uppercase(),
        quote.trim().to_ascii_uppercase(),
    ))
}

fn sanitize_diagnostic_value(s: &str) -> String {
    let mut out = String::new();
    let mut len = 0usize;
    for c in s.chars() {
        if c.is_control() {
            continue;
        }
        out.push(c);
        len += 1;
        if len >= MAX_DIAGNOSTIC_VALUE_LEN {
            out.push('…');
            break;
        }
    }
    out
}
