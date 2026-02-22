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
            base_url,
            credentials,
            http,
        })
    }

    pub fn fetch_executions_page(&self, opts: &FetchOptions) -> Result<Vec<Execution>, String> {
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

            let body = resp.text().unwrap_or_default();
            return Err(format!(
                "bitFlyer API error: status={}, body={}{}",
                status.as_u16(),
                body.chars().take(200).collect::<String>(),
                if body.len() > 200 { "…" } else { "" }
            ));
        }

        Err("unreachable retry loop".to_string())
    }

    pub fn fetch_and_normalize_page(&self, opts: &FetchOptions) -> Result<NormalizeResult, String> {
        let executions = self.fetch_executions_page(opts)?;
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

pub fn build_executions_path(opts: &FetchOptions) -> String {
    let mut q = vec![
        format!("product_code={}", opts.product_code),
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
