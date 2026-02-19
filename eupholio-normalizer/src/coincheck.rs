use chrono::{DateTime, Utc};
use csv::StringRecord;
use eupholio_core::event::Event;
use regex::Regex;
use rust_decimal::Decimal;
use std::collections::HashMap;
use std::str::FromStr;
use std::sync::OnceLock;

const OP_COMPLETED_TRADING_CONTRACTS: &str = "Completed trading contracts";
const TIME_LAYOUT: &str = "%Y-%m-%d %H:%M:%S %z";
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

enum RowOutcome {
    Event(Event),
    Unsupported(String),
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
    let header_index = build_header_index(&headers);

    let mut events = Vec::new();
    let mut diagnostics = Vec::new();

    for (i, result) in rdr.records().enumerate() {
        let row = i + 2;
        let record = result.map_err(|e| format!("row {}: invalid csv row: {}", row, e))?;

        if record.iter().all(|v| v.trim().is_empty()) {
            continue;
        }

        match map_row(&header_index, &record) {
            Ok(RowOutcome::Event(event)) => events.push(event),
            Ok(RowOutcome::Unsupported(reason)) => diagnostics.push(NormalizeDiagnostic { row, reason }),
            Err(e) => return Err(format!("row {}: {}", row, e)),
        }
    }

    Ok(NormalizeResult {
        events,
        diagnostics,
    })
}

fn map_row(index: &HashMap<String, usize>, row: &StringRecord) -> Result<RowOutcome, String> {
    let id = get(index, row, "id")?;
    let operation = get(index, row, "operation")?;

    if operation != OP_COMPLETED_TRADING_CONTRACTS {
        return Ok(RowOutcome::Unsupported(format!(
            "unsupported operation: operation='{}', id='{}'",
            sanitize_diagnostic_value(operation),
            sanitize_diagnostic_value(id)
        )));
    }

    let amount = parse_decimal(get(index, row, "amount")?)?;
    let trading_currency = get(index, row, "trading_currency")?;
    let fee = parse_optional_decimal(get(index, row, "fee")?)?.unwrap_or(Decimal::ZERO);
    let ts = parse_datetime(get(index, row, "time")?)?;

    let (rate, quote, base) = parse_rate_pair(get(index, row, "comment")?)?;
    if base != "JPY" {
        return Err(format!(
            "unsupported payment asset '{}', only JPY is supported",
            base
        ));
    }

    let event = if trading_currency == quote {
        Event::Acquire {
            id: format!("{}:acquire", id),
            asset: quote.to_string(),
            qty: amount,
            jpy_cost: (rate * amount) + fee,
            ts,
        }
    } else if trading_currency == base {
        Event::Dispose {
            id: format!("{}:dispose", id),
            asset: quote.to_string(),
            qty: amount / rate,
            jpy_proceeds: amount - fee,
            ts,
        }
    } else {
        return Err(format!(
            "invalid trading currency '{}' for pair {}_{}",
            trading_currency, quote, base
        ));
    };

    Ok(RowOutcome::Event(event))
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

fn parse_optional_decimal(s: &str) -> Result<Option<Decimal>, String> {
    let trimmed = s.trim();
    if trimmed.is_empty() {
        return Ok(None);
    }
    parse_decimal(trimmed).map(Some)
}

fn parse_datetime(s: &str) -> Result<DateTime<Utc>, String> {
    let dt = DateTime::parse_from_str(s, TIME_LAYOUT)
        .map_err(|e| format!("invalid datetime '{}': {}", s, e))?;
    Ok(dt.with_timezone(&Utc))
}

fn parse_rate_pair(comment: &str) -> Result<(Decimal, String, String), String> {
    static COMMENT_RATE_PAIR_RE: OnceLock<Regex> = OnceLock::new();
    let re = COMMENT_RATE_PAIR_RE.get_or_init(|| {
        Regex::new(r"Rate: ([0-9]+(?:\.[0-9]+)?), Pair: ([0-9a-z]+)_([0-9a-z]+)")
            .expect("regex should compile")
    });

    let caps = re
        .captures(comment)
        .ok_or_else(|| format!("failed to parse comment: {}", sanitize_diagnostic_value(comment)))?;
    let rate = parse_decimal(&caps[1])?;
    let quote = caps[2].to_ascii_uppercase();
    let base = caps[3].to_ascii_uppercase();
    Ok((rate, quote, base))
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
