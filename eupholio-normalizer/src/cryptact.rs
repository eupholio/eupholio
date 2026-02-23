use chrono::{DateTime, NaiveDateTime, Utc};
use csv::StringRecord;
use eupholio_core::event::Event;
use rust_decimal::Decimal;
use std::collections::HashMap;
use std::str::FromStr;

const TS_LAYOUT: &str = "%Y/%-m/%-d %H:%M:%S";
const REQUIRED_HEADERS: [&str; 10] = [
    "Timestamp",
    "Action",
    "Source",
    "Base",
    "Volume",
    "Price",
    "Counter",
    "Fee",
    "FeeCcy",
    "Comment",
];
const SUPPORTED_ACTIONS: [&str; 17] = [
    "BUY", "SELL", "PAY", "MINING", "SENDFEE", "TIP", "REDUCE", "BONUS", "LENDING",
    "STAKING", "LEND", "RECOVER", "BORROW", "RETURN", "LOSS", "CASH", "DEFIFEE",
];

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

enum RowOutcome {
    Event(Event),
    Unsupported(String),
}

pub fn normalize_custom_csv(raw: &str) -> Result<NormalizeResult, String> {
    let mut rdr = csv::ReaderBuilder::new()
        .has_headers(true)
        .trim(csv::Trim::All)
        .from_reader(raw.as_bytes());

    let headers = rdr
        .headers()
        .map_err(|e| format!("invalid csv header: {}", e))?
        .clone();
    validate_headers(&headers)?;
    let index = build_header_index(&headers);

    let mut events = Vec::new();
    let mut diagnostics = Vec::new();

    for (i, result) in rdr.records().enumerate() {
        let row = i + 2;
        let record = result.map_err(|e| format!("row {}: invalid csv row: {}", row, e))?;
        if record.iter().all(|v| v.trim().is_empty()) {
            continue;
        }

        match map_row(&index, &record) {
            Ok(RowOutcome::Event(e)) => events.push(e),
            Ok(RowOutcome::Unsupported(reason)) => {
                diagnostics.push(NormalizeDiagnostic { row, reason })
            }
            Err(e) => return Err(format!("row {}: {}", row, e)),
        }
    }

    Ok(NormalizeResult {
        events,
        diagnostics,
    })
}

fn map_row(index: &HashMap<String, usize>, row: &StringRecord) -> Result<RowOutcome, String> {
    let ts = parse_datetime(get(index, row, "Timestamp")?)?;
    let action = get(index, row, "Action")?.to_ascii_uppercase();
    let id_base = build_id_base(
        get(index, row, "Timestamp")?,
        get(index, row, "Source")?,
        get(index, row, "Base")?,
        get(index, row, "Counter")?,
    );

    let base_asset = get(index, row, "Base")?.to_ascii_uppercase();
    let qty = parse_decimal(get(index, row, "Volume")?)?;
    if qty <= Decimal::ZERO {
        return Err(format!("volume must be > 0, got {}", qty));
    }

    let counter = get(index, row, "Counter")?.to_ascii_uppercase();
    let price = parse_optional_decimal(get(index, row, "Price")?)?.unwrap_or(Decimal::ZERO);
    let fee = parse_optional_decimal(get(index, row, "Fee")?)?.unwrap_or(Decimal::ZERO);
    let fee_ccy = get(index, row, "FeeCcy")?.to_ascii_uppercase();

    if counter != "JPY" {
        return Ok(RowOutcome::Unsupported(format!(
            "unsupported counter currency: counter='{}', action='{}'",
            counter, action
        )));
    }

    match action.as_str() {
        "BUY" => {
            if fee_ccy != counter && fee_ccy != base_asset {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported BUY fee currency: fee_ccy='{}', base='{}', counter='{}'",
                    fee_ccy, base_asset, counter
                )));
            }

            let jpy_cost = if fee_ccy == counter {
                (price * qty) + fee
            } else {
                price * (qty + fee)
            };

            Ok(RowOutcome::Event(Event::Acquire {
                id: format!("{}:acquire", id_base),
                asset: base_asset,
                qty,
                jpy_cost,
                ts,
            }))
        }
        "SELL" => {
            if fee_ccy != counter && fee_ccy != base_asset {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported SELL fee currency: fee_ccy='{}', base='{}', counter='{}'",
                    fee_ccy, base_asset, counter
                )));
            }

            let jpy_proceeds = if fee_ccy == counter {
                (price * qty) - fee
            } else {
                price * (qty - fee)
            };

            Ok(RowOutcome::Event(Event::Dispose {
                id: format!("{}:dispose", id_base),
                asset: base_asset,
                qty,
                jpy_proceeds,
                ts,
            }))
        }
        _ => Ok(RowOutcome::Unsupported(format!(
            "unsupported action: action='{}' (known actions: {})",
            action,
            SUPPORTED_ACTIONS.join(",")
        ))),
    }
}

fn validate_headers(headers: &StringRecord) -> Result<(), String> {
    let found: Vec<&str> = headers.iter().map(str::trim).collect();

    for expected in REQUIRED_HEADERS {
        if !found.iter().any(|h| h.eq_ignore_ascii_case(expected)) {
            return Err(format!("missing required header {}", expected));
        }
    }

    for h in &found {
        if !REQUIRED_HEADERS.iter().any(|e| h.eq_ignore_ascii_case(e)) {
            return Err(format!("unknown header {}", h));
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
    index: &HashMap<String, usize>,
    row: &'a StringRecord,
    key: &str,
) -> Result<&'a str, String> {
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

fn parse_optional_decimal(s: &str) -> Result<Option<Decimal>, String> {
    let trimmed = s.trim();
    if trimmed.is_empty() {
        return Ok(None);
    }
    parse_decimal(trimmed).map(Some)
}

fn parse_datetime(s: &str) -> Result<DateTime<Utc>, String> {
    let dt = NaiveDateTime::parse_from_str(s, TS_LAYOUT)
        .map_err(|e| format!("invalid datetime '{}': {}", s, e))?;
    Ok(DateTime::<Utc>::from_naive_utc_and_offset(dt, Utc))
}

fn build_id_base(ts: &str, source: &str, base: &str, counter: &str) -> String {
    format!(
        "{}:{}:{}:{}",
        ts.trim(),
        source.trim(),
        base.trim().to_ascii_uppercase(),
        counter.trim().to_ascii_uppercase()
    )
}
