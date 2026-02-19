use chrono::{DateTime, NaiveDateTime, Utc};
use csv::StringRecord;
use eupholio_core::event::Event;
use rust_decimal::Decimal;
use std::collections::HashMap;
use std::str::FromStr;

const ORDER_TYPE_LIMIT_BUY: &str = "LIMIT_BUY";
const ORDER_TYPE_LIMIT_SELL: &str = "LIMIT_SELL";
const CLOSED_LAYOUT: &str = "%m/%d/%Y %I:%M:%S %p";

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

pub fn normalize_order_history_csv(raw: &str) -> Result<NormalizeResult, String> {
    let mut rdr = csv::ReaderBuilder::new()
        .has_headers(true)
        .trim(csv::Trim::All)
        .from_reader(raw.as_bytes());

    let headers = rdr
        .headers()
        .map_err(|e| format!("invalid csv header: {}", e))?
        .clone();
    let header_index = build_header_index(&headers);

    let mut events = Vec::new();
    let mut diagnostics = Vec::new();

    for (i, result) in rdr.records().enumerate() {
        let row = i + 2;
        let record = result.map_err(|e| format!("row {}: invalid csv row: {}", row, e))?;

        // Skip physically empty rows.
        if record.iter().all(|v| v.trim().is_empty()) {
            continue;
        }

        match map_row_to_event(&header_index, &record) {
            Ok(Some(event)) => events.push(event),
            Ok(None) => diagnostics.push(NormalizeDiagnostic {
                row,
                reason: "unsupported order type".to_string(),
            }),
            Err(err) => {
                return Err(format!("row {}: {}", row, err));
            }
        }
    }

    Ok(NormalizeResult {
        events,
        diagnostics,
    })
}

fn map_row_to_event(index: &HashMap<String, usize>, row: &StringRecord) -> Result<Option<Event>, String> {
    let id = get(index, row, "Uuid")?;
    let exchange = get(index, row, "Exchange")?;
    let order_type = get(index, row, "OrderType")?;
    let qty = parse_decimal(get(index, row, "Quantity")?)?;
    let price = parse_decimal(get(index, row, "Price")?)?;
    let commission = parse_decimal(get(index, row, "Commission")?)?;
    let ts = parse_datetime(get(index, row, "Closed")?)?;

    let (_payment, trading) = split_exchange(exchange)?;

    let event = match order_type {
        ORDER_TYPE_LIMIT_BUY => Event::Acquire {
            id: format!("{}:acquire", id),
            asset: trading.to_string(),
            qty,
            jpy_cost: price + commission,
            ts,
        },
        ORDER_TYPE_LIMIT_SELL => Event::Dispose {
            id: format!("{}:dispose", id),
            asset: trading.to_string(),
            qty,
            jpy_proceeds: price - commission,
            ts,
        },
        _ => return Ok(None),
    };

    Ok(Some(event))
}

fn build_header_index(headers: &StringRecord) -> HashMap<String, usize> {
    headers
        .iter()
        .enumerate()
        .map(|(i, col)| (col.to_string(), i))
        .collect()
}

fn get<'a>(index: &HashMap<String, usize>, row: &'a StringRecord, key: &str) -> Result<&'a str, String> {
    let i = *index
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

fn parse_datetime(s: &str) -> Result<DateTime<Utc>, String> {
    let naive = NaiveDateTime::parse_from_str(s, CLOSED_LAYOUT)
        .map_err(|e| format!("invalid datetime '{}': {}", s, e))?;
    Ok(DateTime::<Utc>::from_naive_utc_and_offset(naive, Utc))
}

fn split_exchange(s: &str) -> Result<(&str, &str), String> {
    let mut parts = s.split('-');
    let payment = parts.next().ok_or_else(|| "missing payment asset".to_string())?;
    let trading = parts.next().ok_or_else(|| "missing trading asset".to_string())?;
    if parts.next().is_some() {
        return Err(format!("invalid exchange pair '{}'", s));
    }
    Ok((payment, trading))
}
