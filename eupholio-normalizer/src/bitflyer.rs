use chrono::{DateTime, Utc};
use csv::StringRecord;
use eupholio_core::event::Event;
use rust_decimal::Decimal;
use std::collections::HashMap;
use std::str::FromStr;

const OP_BUY_JP: &str = "買い";
const OP_SELL_JP: &str = "売り";
const OP_BUY_EN: &str = "BUY";
const OP_SELL_EN: &str = "SELL";
const MAX_DIAGNOSTIC_VALUE_LEN: usize = 64;

const HEADERS_JP: &[&str] = &[
    "取引日時",
    "通貨",
    "取引種別",
    "取引価格",
    "通貨1",
    "通貨1数量",
    "手数料",
    "通貨1の対円レート",
    "通貨2",
    "通貨2数量",
    "自己・媒介",
    "注文 ID",
    "備考",
];

const HEADERS_EN: &[&str] = &[
    "Trade Date",
    "Product",
    "Trade Type",
    "Traded Price",
    "Currency 1",
    "Amount (Currency 1)",
    "Fee",
    "JPY Rate (Currency 1)",
    "Currency 2",
    "Amount (Currency 2)",
    "Counter Party",
    "Order ID",
    "Details",
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

pub fn normalize_transaction_history_csv(raw: &str) -> Result<NormalizeResult, String> {
    let mut rdr = csv::ReaderBuilder::new()
        .has_headers(true)
        .trim(csv::Trim::All)
        .from_reader(raw.as_bytes());

    let headers = rdr
        .headers()
        .map_err(|e| format!("invalid csv header: {}", e))?
        .clone();

    let lang_headers = detect_headers(&headers)?;
    let idx = build_header_index(&headers);

    let mut events = Vec::new();
    let mut diagnostics = Vec::new();

    for (i, rec) in rdr.records().enumerate() {
        let row = i + 2;
        let rec = rec.map_err(|e| format!("row {}: invalid csv row: {}", row, e))?;
        if rec.iter().all(|v| v.trim().is_empty()) {
            continue;
        }

        match map_row(&idx, &rec, lang_headers) {
            Ok(Some(event)) => events.push(event),
            Ok(None) => diagnostics.push(NormalizeDiagnostic {
                row,
                reason: format!(
                    "unsupported trade type: trade_type='{}', order_id='{}'",
                    sanitize(get(&idx, &rec, lang_headers[2])?),
                    sanitize(get(&idx, &rec, lang_headers[11])?)
                ),
            }),
            Err(e) => return Err(format!("row {}: {}", row, e)),
        }
    }

    Ok(NormalizeResult { events, diagnostics })
}

fn map_row(
    idx: &HashMap<String, usize>,
    row: &StringRecord,
    h: &[&str],
) -> Result<Option<Event>, String> {
    let ts = parse_datetime(get(idx, row, h[0])?)?;
    let trade_type = get(idx, row, h[2])?;
    let asset1 = get(idx, row, h[4])?.to_ascii_uppercase();
    let qty1 = parse_decimal(get(idx, row, h[5])?)?;
    let fee = parse_decimal(get(idx, row, h[6])?)?;
    let rate_jpy = parse_decimal(get(idx, row, h[7])?)?;
    let asset2 = get(idx, row, h[8])?.to_ascii_uppercase();
    let qty2 = parse_decimal(get(idx, row, h[9])?)?;
    let order_id = get(idx, row, h[11])?;

    let fee_jpy = fee.abs() * rate_jpy;

    if trade_type == OP_BUY_JP || trade_type == OP_BUY_EN {
        if asset2 != "JPY" {
            return Err(format!("unsupported payment asset '{}', only JPY is supported", asset2));
        }
        let net_qty = qty1 + fee;
        if net_qty <= Decimal::ZERO {
            return Err(format!("buy qty must be > 0 after fee, got {}", net_qty));
        }
        return Ok(Some(Event::Acquire {
            id: format!("{}:acquire", order_id),
            asset: asset1,
            qty: net_qty,
            jpy_cost: qty2.abs() + fee_jpy,
            ts,
        }));
    }

    if trade_type == OP_SELL_JP || trade_type == OP_SELL_EN {
        if asset2 != "JPY" {
            return Err(format!("unsupported payment asset '{}', only JPY is supported", asset2));
        }
        if qty1 <= Decimal::ZERO {
            return Err(format!("sell qty must be > 0, got {}", qty1));
        }
        return Ok(Some(Event::Dispose {
            id: format!("{}:dispose", order_id),
            asset: asset1,
            qty: qty1,
            jpy_proceeds: qty2.abs() - fee_jpy,
            ts,
        }));
    }

    Ok(None)
}

fn detect_headers(headers: &StringRecord) -> Result<&'static [&'static str], String> {
    if validate_headers(headers, HEADERS_JP).is_ok() {
        return Ok(HEADERS_JP);
    }
    validate_headers(headers, HEADERS_EN)?;
    Ok(HEADERS_EN)
}

fn validate_headers(headers: &StringRecord, expected: &[&str]) -> Result<(), String> {
    for h in expected {
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

fn get<'a>(idx: &HashMap<String, usize>, row: &'a StringRecord, key: &str) -> Result<&'a str, String> {
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

fn parse_datetime(s: &str) -> Result<DateTime<Utc>, String> {
    let with_zone = format!("{} +0900", s);
    let dt = DateTime::parse_from_str(&with_zone, "%Y/%m/%d %H:%M:%S %z")
        .or_else(|_| DateTime::parse_from_rfc3339(s))
        .map_err(|e| format!("invalid datetime '{}': {}", s, e))?;
    Ok(dt.with_timezone(&Utc))
}

fn sanitize(s: &str) -> String {
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
