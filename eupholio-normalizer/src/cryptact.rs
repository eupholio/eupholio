use chrono::{DateTime, FixedOffset, NaiveDateTime, TimeZone, Utc};
use csv::StringRecord;
use eupholio_core::event::Event;
use rust_decimal::Decimal;
use std::collections::HashMap;
use std::str::FromStr;

const TS_LAYOUT: &str = "%Y/%-m/%-d %H:%M:%S";
const JST_OFFSET_SECS: i32 = 9 * 60 * 60;
const MAX_DIAGNOSTIC_VALUE_LEN: usize = 120;
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
    "BUY", "SELL", "PAY", "MINING", "SENDFEE", "TIP", "REDUCE", "BONUS", "LENDING", "STAKING",
    "LEND", "RECOVER", "BORROW", "RETURN", "LOSS", "CASH", "DEFIFEE",
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

        match map_row(&index, &record, row) {
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

fn map_row(
    index: &HashMap<String, usize>,
    row: &StringRecord,
    row_num: usize,
) -> Result<RowOutcome, String> {
    let ts_raw = get(index, row, "Timestamp")?;
    let ts = parse_datetime(ts_raw)?;
    let action = get(index, row, "Action")?.to_ascii_uppercase();
    let id_base = build_id_base(
        ts_raw,
        get(index, row, "Source")?,
        get(index, row, "Base")?,
        get(index, row, "Counter")?,
        row_num,
    );

    let base_asset = get(index, row, "Base")?.to_ascii_uppercase();
    let qty = parse_decimal(get(index, row, "Volume")?)?;
    if qty <= Decimal::ZERO {
        return Err(format!("volume must be > 0, got {}", qty));
    }

    let counter = get(index, row, "Counter")?.to_ascii_uppercase();
    let price_raw = get(index, row, "Price")?;
    let price_opt = parse_optional_decimal(price_raw)?;
    let fee = parse_optional_decimal(get(index, row, "Fee")?)?.unwrap_or(Decimal::ZERO);
    if fee < Decimal::ZERO {
        return Err(format!("fee must be >= 0, got {}", fee));
    }
    let fee_ccy = get(index, row, "FeeCcy")?.to_ascii_uppercase();

    if counter != "JPY" {
        return Ok(RowOutcome::Unsupported(format!(
            "unsupported counter currency: counter='{}', action='{}'",
            sanitize_diagnostic_value(&counter),
            sanitize_diagnostic_value(&action)
        )));
    }

    match action.as_str() {
        "PAY" => {
            if fee_ccy != counter {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported PAY fee currency: fee_ccy='{}', counter='{}'",
                    sanitize_diagnostic_value(&fee_ccy),
                    sanitize_diagnostic_value(&counter)
                )));
            }
            if fee != Decimal::ZERO {
                return Err(format!("fee must be 0 for PAY in phase-2, got {}", fee));
            }

            let jpy_proceeds = price_opt.map(|p| p * qty).unwrap_or(Decimal::ZERO);
            Ok(RowOutcome::Event(Event::Dispose {
                id: format!("{}:pay", id_base),
                asset: base_asset,
                qty,
                jpy_proceeds,
                ts,
            }))
        }
        "MINING" => {
            if fee_ccy != counter {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported MINING fee currency: fee_ccy='{}', counter='{}'",
                    sanitize_diagnostic_value(&fee_ccy),
                    sanitize_diagnostic_value(&counter)
                )));
            }
            if fee != Decimal::ZERO {
                return Err(format!("fee must be 0 for MINING in phase-2, got {}", fee));
            }

            let jpy_value = price_opt.map(|p| p * qty).unwrap_or(Decimal::ZERO);
            Ok(RowOutcome::Event(Event::Income {
                id: format!("{}:mining", id_base),
                asset: base_asset,
                qty,
                jpy_value,
                ts,
            }))
        }
        "SENDFEE" => {
            if fee != Decimal::ZERO {
                return Err(format!("fee must be 0 for SENDFEE, got {}", fee));
            }
            if fee_ccy != counter {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported SENDFEE fee currency: fee_ccy='{}', counter='{}'",
                    sanitize_diagnostic_value(&fee_ccy),
                    sanitize_diagnostic_value(&counter)
                )));
            }

            Ok(RowOutcome::Event(Event::Transfer {
                id: format!("{}:sendfee", id_base),
                asset: base_asset,
                qty,
                direction: eupholio_core::event::TransferDirection::Out,
                ts,
            }))
        }
        "BONUS" | "LENDING" | "STAKING" => {
            if fee_ccy != counter {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported {} fee currency: fee_ccy='{}', counter='{}'",
                    action,
                    sanitize_diagnostic_value(&fee_ccy),
                    sanitize_diagnostic_value(&counter)
                )));
            }
            if fee != Decimal::ZERO {
                return Err(format!(
                    "fee must be 0 for {} in phase-4, got {}",
                    action, fee
                ));
            }

            let jpy_value = price_opt.map(|p| p * qty).unwrap_or(Decimal::ZERO);
            Ok(RowOutcome::Event(Event::Income {
                id: format!("{}:{}", id_base, action.to_ascii_lowercase()),
                asset: base_asset,
                qty,
                jpy_value,
                ts,
            }))
        }
        "TIP" => {
            if fee_ccy != counter {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported TIP fee currency: fee_ccy='{}', counter='{}'",
                    sanitize_diagnostic_value(&fee_ccy),
                    sanitize_diagnostic_value(&counter)
                )));
            }
            if fee != Decimal::ZERO {
                return Err(format!("fee must be 0 for TIP in phase-4, got {}", fee));
            }

            let jpy_proceeds = price_opt.map(|p| p * qty).unwrap_or(Decimal::ZERO);
            Ok(RowOutcome::Event(Event::Dispose {
                id: format!("{}:tip", id_base),
                asset: base_asset,
                qty,
                jpy_proceeds,
                ts,
            }))
        }
        "LOSS" => {
            if fee_ccy != counter {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported LOSS fee currency: fee_ccy='{}', counter='{}'",
                    sanitize_diagnostic_value(&fee_ccy),
                    sanitize_diagnostic_value(&counter)
                )));
            }
            if fee != Decimal::ZERO {
                return Err(format!("fee must be 0 for LOSS in phase-4, got {}", fee));
            }

            Ok(RowOutcome::Event(Event::Dispose {
                id: format!("{}:loss", id_base),
                asset: base_asset,
                qty,
                jpy_proceeds: Decimal::ZERO,
                ts,
            }))
        }
        "REDUCE" => {
            if fee_ccy != counter {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported REDUCE fee currency: fee_ccy='{}', counter='{}'",
                    sanitize_diagnostic_value(&fee_ccy),
                    sanitize_diagnostic_value(&counter)
                )));
            }
            if fee != Decimal::ZERO {
                return Err(format!("fee must be 0 for REDUCE in phase-4, got {}", fee));
            }

            Ok(RowOutcome::Event(Event::Transfer {
                id: format!("{}:reduce", id_base),
                asset: base_asset,
                qty,
                direction: eupholio_core::event::TransferDirection::Out,
                ts,
            }))
        }
        "LEND" => {
            if fee_ccy != counter {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported LEND fee currency: fee_ccy='{}', counter='{}'",
                    sanitize_diagnostic_value(&fee_ccy),
                    sanitize_diagnostic_value(&counter)
                )));
            }
            if fee != Decimal::ZERO {
                return Err(format!("fee must be 0 for LEND in phase-5, got {}", fee));
            }
            Ok(RowOutcome::Event(Event::Transfer {
                id: format!("{}:lend", id_base),
                asset: base_asset,
                qty,
                direction: eupholio_core::event::TransferDirection::Out,
                ts,
            }))
        }
        "RECOVER" | "BORROW" | "RETURN" => {
            if fee_ccy != counter {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported {} fee currency: fee_ccy='{}', counter='{}'",
                    action,
                    sanitize_diagnostic_value(&fee_ccy),
                    sanitize_diagnostic_value(&counter)
                )));
            }
            if fee != Decimal::ZERO {
                return Err(format!("fee must be 0 for {} in phase-5, got {}", action, fee));
            }

            let direction = if action == "RECOVER" || action == "BORROW" {
                eupholio_core::event::TransferDirection::In
            } else {
                eupholio_core::event::TransferDirection::Out
            };

            Ok(RowOutcome::Event(Event::Transfer {
                id: format!("{}:{}", id_base, action.to_ascii_lowercase()),
                asset: base_asset,
                qty,
                direction,
                ts,
            }))
        }
        "DEFIFEE" => {
            if fee_ccy != counter {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported DEFIFEE fee currency: fee_ccy='{}', counter='{}'",
                    sanitize_diagnostic_value(&fee_ccy),
                    sanitize_diagnostic_value(&counter)
                )));
            }
            if fee != Decimal::ZERO {
                return Err(format!("fee must be 0 for DEFIFEE in phase-5, got {}", fee));
            }
            Ok(RowOutcome::Event(Event::Dispose {
                id: format!("{}:defifee", id_base),
                asset: base_asset,
                qty,
                jpy_proceeds: Decimal::ZERO,
                ts,
            }))
        }
        "CASH" => Ok(RowOutcome::Unsupported(
            "CASH is not supported in rust-core Event model yet".to_string(),
        )),
        "BUY" => {
            let price = required_positive_price(price_opt, &action)?;
            if fee_ccy != counter && fee_ccy != base_asset {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported BUY fee currency: fee_ccy='{}', base='{}', counter='{}'",
                    sanitize_diagnostic_value(&fee_ccy),
                    sanitize_diagnostic_value(&base_asset),
                    sanitize_diagnostic_value(&counter)
                )));
            }

            let (acquired_qty, jpy_cost) = if fee_ccy == counter {
                (qty, (price * qty) + fee)
            } else {
                let net_qty = qty - fee;
                if net_qty <= Decimal::ZERO {
                    return Err(format!(
                        "net volume after base-asset fee must be > 0, got {}",
                        net_qty
                    ));
                }
                (net_qty, price * qty)
            };

            Ok(RowOutcome::Event(Event::Acquire {
                id: format!("{}:acquire", id_base),
                asset: base_asset,
                qty: acquired_qty,
                jpy_cost,
                ts,
            }))
        }
        "SELL" => {
            let price = required_positive_price(price_opt, &action)?;
            if fee_ccy != counter && fee_ccy != base_asset {
                return Ok(RowOutcome::Unsupported(format!(
                    "unsupported SELL fee currency: fee_ccy='{}', base='{}', counter='{}'",
                    sanitize_diagnostic_value(&fee_ccy),
                    sanitize_diagnostic_value(&base_asset),
                    sanitize_diagnostic_value(&counter)
                )));
            }

            let jpy_proceeds = if fee_ccy == counter {
                (price * qty) - fee
            } else {
                price * qty
            };
            let disposed_qty = if fee_ccy == base_asset {
                qty + fee
            } else {
                qty
            };

            Ok(RowOutcome::Event(Event::Dispose {
                id: format!("{}:dispose", id_base),
                asset: base_asset,
                qty: disposed_qty,
                jpy_proceeds,
                ts,
            }))
        }
        _ => Ok(RowOutcome::Unsupported(format!(
            "unsupported action: action='{}' (known actions: {})",
            sanitize_diagnostic_value(&action),
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
        .map(|(i, col)| (normalize_header(col), i))
        .collect()
}

fn get<'a>(
    index: &HashMap<String, usize>,
    row: &'a StringRecord,
    key: &str,
) -> Result<&'a str, String> {
    let i = *index
        .get(&normalize_header(key))
        .ok_or_else(|| format!("missing required header {}", key))?;
    row.get(i)
        .map(str::trim)
        .ok_or_else(|| format!("missing required field {}", key))
}

fn normalize_header(s: &str) -> String {
    s.trim().to_ascii_lowercase()
}

fn required_positive_price(price_opt: Option<Decimal>, action: &str) -> Result<Decimal, String> {
    let price = price_opt.ok_or_else(|| format!("price must be provided for {}", action))?;
    if price <= Decimal::ZERO {
        return Err(format!("price must be > 0 for {}, got {}", action, price));
    }
    Ok(price)
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
    let jst = FixedOffset::east_opt(JST_OFFSET_SECS).ok_or("invalid JST offset")?;
    let local = jst
        .from_local_datetime(&dt)
        .single()
        .ok_or_else(|| format!("ambiguous datetime '{}' in JST", s))?;
    Ok(local.with_timezone(&Utc))
}

fn build_id_base(ts: &str, source: &str, base: &str, counter: &str, row_num: usize) -> String {
    format!(
        "{}:{}:{}:{}:{}",
        ts.trim(),
        source.trim(),
        base.trim().to_ascii_uppercase(),
        counter.trim().to_ascii_uppercase(),
        row_num
    )
}

fn sanitize_diagnostic_value(s: &str) -> String {
    let mut out = String::new();
    for c in s.chars() {
        if out.len() >= MAX_DIAGNOSTIC_VALUE_LEN {
            out.push('…');
            break;
        }
        if c.is_control() {
            out.push('�');
        } else {
            out.push(c);
        }
    }
    out
}
