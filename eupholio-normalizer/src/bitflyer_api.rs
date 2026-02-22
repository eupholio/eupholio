use std::env;
use std::thread;
use std::time::Duration;

use chrono::{DateTime, Utc};
use eupholio_core::event::Event;
use hmac::{Hmac, Mac};
use reqwest::blocking::Client;
use reqwest::header::{HeaderMap, HeaderValue};
use rust_decimal::Decimal;
use serde::Deserialize;
use sha2::Sha256;

use crate::bitflyer::{NormalizeDiagnostic, NormalizeResult};

type HmacSha256 = Hmac<Sha256>;

const DEFAULT_BASE_URL: &str = "https://api.bitflyer.com";
const EXECUTIONS_PATH: &str = "/v1/me/getexecutions";

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

#[derive(Debug, Clone)]
pub struct ApiCredentials {
    pub api_key: String,
    pub api_secret: String,
}

#[derive(Debug, Clone)]
pub struct FetchOptions {
    pub product_code: String,
    pub count: usize,
    pub before: Option<i64>,
    pub after: Option<i64>,
}

#[derive(Debug, Clone)]
pub struct FetchWindowOptions {
    pub product_code: String,
    pub count_per_page: usize,
    pub max_pages: usize,
    pub since: Option<DateTime<Utc>>,
    pub until: Option<DateTime<Utc>>,
}

impl Default for FetchWindowOptions {
    fn default() -> Self {
        Self {
            product_code: "BTC_JPY".to_string(),
            count_per_page: 100,
            max_pages: 20,
            since: None,
            until: None,
        }
    }
}

impl Default for FetchOptions {
    fn default() -> Self {
        Self {
            product_code: "BTC_JPY".to_string(),
            count: 100,
            before: None,
            after: None,
        }
    }
}

pub struct BitflyerApiClient {
    base_url: String,
    credentials: ApiCredentials,
    http: Client,
}

impl BitflyerApiClient {
    pub fn from_env() -> Result<Self, String> {
        let api_key =
            env::var("BITFLYER_API_KEY").map_err(|_| "BITFLYER_API_KEY is required".to_string())?;
        let api_secret = env::var("BITFLYER_API_SECRET")
            .map_err(|_| "BITFLYER_API_SECRET is required".to_string())?;

        Self::new(
            DEFAULT_BASE_URL.to_string(),
            ApiCredentials {
                api_key,
                api_secret,
            },
        )
    }

    pub fn new(base_url: String, credentials: ApiCredentials) -> Result<Self, String> {
        let http = Client::builder()
            .user_agent("eupholio-normalizer/bitflyer-api-poc")
            .timeout(Duration::from_secs(15))
            .build()
            .map_err(|e| format!("failed to build http client: {e}"))?;

        Ok(Self {
            base_url: base_url.trim_end_matches('/').to_string(),
            credentials,
            http,
        })
    }

    pub fn fetch_executions_page(&self, opts: &FetchOptions) -> Result<Vec<Execution>, String> {
        validate_product_code(&opts.product_code)?;
        let path_with_query = build_executions_path(opts);
        let url = format!("{}{}", self.base_url, path_with_query);

        let mut backoff = Duration::from_millis(300);
        let max_attempts = 3;

        for attempt in 1..=max_attempts {
            let ts = Utc::now().timestamp().to_string();
            let sign = sign_request(
                &self.credentials.api_secret,
                &ts,
                "GET",
                &path_with_query,
                "",
            );
            let headers = build_auth_headers(&self.credentials.api_key, &ts, &sign)?;

            let resp = self
                .http
                .get(&url)
                .headers(headers)
                .send()
                .map_err(|e| format!("request failed: {e}"))?;

            let status = resp.status();
            if status.is_success() {
                return resp
                    .json::<Vec<Execution>>()
                    .map_err(|e| format!("failed to parse executions response: {e}"));
            }

            if status.as_u16() == 401 || status.as_u16() == 403 {
                return Err(format!("authentication failed: status={}", status.as_u16()));
            }

            let retryable = status.as_u16() == 429 || status.is_server_error();
            if retryable && attempt < max_attempts {
                thread::sleep(backoff);
                backoff *= 2;
                continue;
            }

            return Err(format!("bitFlyer API error: status={}", status.as_u16()));
        }

        unreachable!("unreachable retry loop in fetch_executions_page")
    }

    pub fn fetch_and_normalize_page(&self, opts: &FetchOptions) -> Result<NormalizeResult, String> {
        let executions = self.fetch_executions_page(opts)?;
        normalize_executions(&executions, &opts.product_code)
    }

    pub fn fetch_executions_window(
        &self,
        opts: &FetchWindowOptions,
    ) -> Result<Vec<Execution>, String> {
        validate_fetch_window_options(opts)?;

        let mut all = Vec::new();
        let mut before: Option<i64> = None;

        for _ in 0..opts.max_pages {
            let page = self.fetch_executions_page(&FetchOptions {
                product_code: opts.product_code.clone(),
                count: opts.count_per_page,
                before,
                after: None,
            })?;

            if page.is_empty() {
                break;
            }

            let filtered = filter_executions_by_time(&page, opts.since, opts.until);
            all.extend(filtered);

            // bitFlyer `before` is an upper bound on execution id (exclusive-ish in practice);
            // stepping to oldest id keeps pagination moving toward older records.
            let oldest_id = page.iter().map(|e| e.id).min();
            before = oldest_id;

            // NOTE: we intentionally do not early-terminate by comparing timestamp ordering
            // across pages. If execution-id ordering and timestamps are not strictly correlated,
            // early termination could skip valid records inside the requested time window.
            thread::sleep(Duration::from_millis(150));
        }

        Ok(all)
    }

    pub fn fetch_and_normalize_window(
        &self,
        opts: &FetchWindowOptions,
    ) -> Result<NormalizeResult, String> {
        let executions = self.fetch_executions_window(opts)?;
        normalize_executions(&executions, &opts.product_code)
    }
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
        let side_raw = ex.side.trim();

        if side_raw.eq_ignore_ascii_case("BUY") {
            let net_qty = ex.size - fee_base;
            if net_qty <= Decimal::ZERO {
                return Err(format!(
                    "row {}: buy qty must be > 0 after fee, got {}",
                    row, net_qty
                ));
            }
            events.push(Event::Acquire {
                id: format!("bfexec-{}:acquire", ex.id),
                asset: base_asset.clone(),
                qty: net_qty,
                jpy_cost: jpy_total + fee_jpy,
                ts: ex.exec_date,
            });
        } else if side_raw.eq_ignore_ascii_case("SELL") {
            events.push(Event::Dispose {
                id: format!("bfexec-{}:dispose", ex.id),
                asset: base_asset.clone(),
                qty: ex.size,
                jpy_proceeds: jpy_total - fee_jpy,
                ts: ex.exec_date,
            });
        } else {
            diagnostics.push(NormalizeDiagnostic {
                row,
                reason: format!(
                    "unsupported side: side='{}', execution_id='{}'",
                    side_raw, ex.id
                ),
            });
        }
    }

    Ok(NormalizeResult {
        events,
        diagnostics,
    })
}

pub fn filter_executions_by_time(
    executions: &[Execution],
    since: Option<DateTime<Utc>>,
    until: Option<DateTime<Utc>>,
) -> Vec<Execution> {
    executions
        .iter()
        .filter(|e| {
            let ge_since = since.map(|s| e.exec_date >= s).unwrap_or(true);
            let le_until = until.map(|u| e.exec_date <= u).unwrap_or(true);
            ge_since && le_until
        })
        .cloned()
        .collect()
}

pub fn build_executions_path(opts: &FetchOptions) -> String {
    let product_code =
        validate_product_code(&opts.product_code).unwrap_or_else(|_| opts.product_code.clone());

    let mut q = vec![
        format!("product_code={}", product_code),
        format!("count={}", opts.count),
    ];

    if let Some(v) = opts.before {
        q.push(format!("before={}", v));
    }
    if let Some(v) = opts.after {
        q.push(format!("after={}", v));
    }

    format!("{}?{}", EXECUTIONS_PATH, q.join("&"))
}

fn build_auth_headers(api_key: &str, timestamp: &str, sign: &str) -> Result<HeaderMap, String> {
    let mut headers = HeaderMap::new();
    headers.insert(
        "ACCESS-KEY",
        HeaderValue::from_str(api_key).map_err(|e| format!("invalid ACCESS-KEY header: {e}"))?,
    );
    headers.insert(
        "ACCESS-TIMESTAMP",
        HeaderValue::from_str(timestamp)
            .map_err(|e| format!("invalid ACCESS-TIMESTAMP header: {e}"))?,
    );
    headers.insert(
        "ACCESS-SIGN",
        HeaderValue::from_str(sign).map_err(|e| format!("invalid ACCESS-SIGN header: {e}"))?,
    );
    Ok(headers)
}

fn split_product(product_code: &str) -> Result<(String, String), String> {
    let normalized = validate_product_code(product_code)?;
    let mut parts = normalized.split('_');
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

fn validate_product_code(product_code: &str) -> Result<String, String> {
    let code = product_code.trim().to_ascii_uppercase();
    if code.is_empty() {
        return Err("product_code must not be empty".to_string());
    }
    if !code
        .chars()
        .all(|c| c.is_ascii_uppercase() || c.is_ascii_digit() || c == '_')
    {
        return Err(format!(
            "invalid product_code '{}': only [A-Z0-9_] is allowed",
            product_code
        ));
    }
    Ok(code)
}

fn validate_fetch_window_options(opts: &FetchWindowOptions) -> Result<(), String> {
    if opts.count_per_page == 0 {
        return Err("count_per_page must be > 0".to_string());
    }
    if opts.max_pages == 0 {
        return Err("max_pages must be > 0".to_string());
    }
    if let (Some(since), Some(until)) = (opts.since, opts.until) {
        if since > until {
            return Err(format!(
                "invalid time window: since ({}) must be <= until ({})",
                since, until
            ));
        }
    }
    Ok(())
}
