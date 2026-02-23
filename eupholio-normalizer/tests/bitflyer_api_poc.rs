use std::io::{Read, Write};
use std::net::TcpListener;
use std::sync::{
    atomic::{AtomicUsize, Ordering},
    Arc,
};
use std::thread;
use std::time::Instant;

use chrono::{DateTime, Duration as ChronoDuration, Utc};
use eupholio_core::event::Event;
use eupholio_normalizer::bitflyer_api::{
    build_executions_path, filter_executions_by_time, normalize_executions, sign_request,
    ApiCredentials, BitflyerApiClient, Execution, FetchOptions, FetchWindowOptions,
};

#[test]
fn bitflyer_api_sign_request_is_deterministic() {
    let sign = sign_request(
        "secret",
        "1700000000",
        "GET",
        "/v1/me/getexecutions?product_code=BTC_JPY&count=100",
        "",
    );
    assert_eq!(
        sign,
        "3a8a2a40c61b15c6014722eee597a2a111f35d22ba2999e4bc04310a946f9719"
    );
}

#[test]
fn bitflyer_api_normalize_executions_to_events() {
    let raw = r#"[
      {"id": 1001, "side": "BUY", "price": "1000000", "size": "0.01", "exec_date": "2026-01-01T00:00:00Z", "commission": "0.0001"},
      {"id": 1002, "side": "SELL", "price": "1200000", "size": "0.005", "exec_date": "2026-01-02T00:00:00Z", "commission": "0.0001"},
      {"id": 1003, "side": "  oThEr  ", "price": "1", "size": "1", "exec_date": "2026-01-03T00:00:00Z"}
    ]"#;

    let executions: Vec<Execution> = serde_json::from_str(raw).expect("json should parse");
    let normalized =
        normalize_executions(&executions, "btc_jpy").expect("normalization should work");

    assert_eq!(normalized.events.len(), 2);
    assert_eq!(normalized.diagnostics.len(), 1);
    assert!(normalized.diagnostics[0]
        .reason
        .contains("side='  oThEr  '"));

    match &normalized.events[0] {
        Event::Acquire {
            id,
            asset,
            qty,
            jpy_cost,
            ..
        } => {
            assert_eq!(id, "bfexec-1001:acquire");
            assert_eq!(asset, "BTC");
            assert_eq!(qty.to_string(), "0.0099");
            assert_eq!(jpy_cost.to_string(), "10100.0000");
        }
        _ => panic!("expected acquire"),
    }

    match &normalized.events[1] {
        Event::Dispose {
            id,
            asset,
            qty,
            jpy_proceeds,
            ..
        } => {
            assert_eq!(id, "bfexec-1002:dispose");
            assert_eq!(asset, "BTC");
            assert_eq!(qty.to_string(), "0.005");
            assert_eq!(jpy_proceeds.to_string(), "5880.0000");
        }
        _ => panic!("expected dispose"),
    }
}

#[test]
fn bitflyer_api_non_jpy_quote_is_rejected() {
    let executions: Vec<Execution> = serde_json::from_str("[]").unwrap();
    assert!(normalize_executions(&executions, "BTC_USD")
        .expect_err("non-jpy should fail")
        .contains("only JPY is supported"));
}

#[test]
fn bitflyer_api_build_executions_path_with_paging() {
    let path = build_executions_path(&FetchOptions {
        product_code: "BTC_JPY".to_string(),
        count: 100,
        before: Some(123),
        after: Some(45),
    })
    .expect("path should build");

    assert_eq!(
        path,
        "/v1/me/getexecutions?product_code=BTC_JPY&count=100&before=123&after=45"
    );
}

#[test]
fn bitflyer_api_build_executions_path_rejects_zero_count() {
    let err = build_executions_path(&FetchOptions {
        product_code: "BTC_JPY".to_string(),
        count: 0,
        before: None,
        after: None,
    })
    .expect_err("count=0 should fail");
    assert!(err.contains("count must be > 0"));
}

#[test]
fn bitflyer_api_filter_executions_by_time_window() {
    let raw = r#"[
      {"id": 1, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-01-01T00:00:00Z"},
      {"id": 2, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-02-01T00:00:00Z"},
      {"id": 3, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-03-01T00:00:00Z"}
    ]"#;
    let executions: Vec<Execution> = serde_json::from_str(raw).unwrap();

    let since = DateTime::parse_from_rfc3339("2026-01-15T00:00:00Z")
        .unwrap()
        .with_timezone(&Utc);
    let until = DateTime::parse_from_rfc3339("2026-02-15T00:00:00Z")
        .unwrap()
        .with_timezone(&Utc);

    let filtered = filter_executions_by_time(&executions, Some(since), Some(until));
    assert_eq!(filtered.len(), 1);
    assert_eq!(filtered[0].id, 2);
}

#[test]
fn bitflyer_api_build_executions_path_normalizes_lowercase_product_code() {
    let path = build_executions_path(&FetchOptions {
        product_code: "btc_jpy".to_string(),
        count: 10,
        before: None,
        after: None,
    })
    .expect("lowercase product_code should be normalized");

    assert!(path.contains("product_code=BTC_JPY"));
}

#[test]
fn bitflyer_api_build_executions_path_rejects_invalid_product_code() {
    let err = build_executions_path(&FetchOptions {
        product_code: "BTC-JPY".to_string(),
        count: 10,
        before: None,
        after: None,
    })
    .expect_err("invalid chars should fail");

    assert!(err.contains("only [A-Z0-9_] is allowed"));
}

#[test]
fn bitflyer_api_build_executions_path_rejects_empty_product_code() {
    let err = build_executions_path(&FetchOptions {
        product_code: "   ".to_string(),
        count: 10,
        before: None,
        after: None,
    })
    .expect_err("empty product_code should fail");

    assert!(err.contains("must not be empty"));
}

#[test]
fn bitflyer_api_window_rejects_invalid_since_until() {
    let client = BitflyerApiClient::new(
        "https://api.bitflyer.com/".to_string(),
        ApiCredentials {
            api_key: "k".to_string(),
            api_secret: "s".to_string(),
        },
    )
    .expect("client should construct");

    let since = DateTime::parse_from_rfc3339("2026-02-01T00:00:00Z")
        .unwrap()
        .with_timezone(&Utc);
    let until = DateTime::parse_from_rfc3339("2026-01-01T00:00:00Z")
        .unwrap()
        .with_timezone(&Utc);

    let err = client
        .fetch_executions_window(&FetchWindowOptions {
            product_code: "BTC_JPY".to_string(),
            count_per_page: 100,
            max_pages: 1,
            since: Some(since),
            until: Some(until),
        })
        .expect_err("since > until should fail before network");

    assert!(err.contains("since"));
    assert!(err.contains("until"));
}

#[test]
fn bitflyer_api_window_dedupes_and_stops_on_non_decreasing_before() {
    let listener = TcpListener::bind("127.0.0.1:0").expect("bind test server");
    let addr = listener.local_addr().expect("local addr");
    let calls = Arc::new(AtomicUsize::new(0));
    let calls_bg = Arc::clone(&calls);

    let server = thread::spawn(move || {
        for stream in listener.incoming().take(2) {
            let mut stream = stream.expect("incoming stream");
            let idx = calls_bg.fetch_add(1, Ordering::SeqCst);

            let mut buf = [0_u8; 2048];
            let _ = stream.read(&mut buf);

            let body = if idx == 0 {
                r#"[
                  {"id": 100, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-01-02T00:00:00Z"},
                  {"id": 99, "side": "SELL", "price": "110", "size": "0.5", "exec_date": "2026-01-01T00:00:00Z"}
                ]"#
            } else {
                // Repeat oldest id=99 to verify de-dup + before-progress guard break.
                r#"[
                  {"id": 99, "side": "SELL", "price": "110", "size": "0.5", "exec_date": "2026-01-01T00:00:00Z"}
                ]"#
            };

            let resp = format!(
                "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                body.len(),
                body
            );
            stream.write_all(resp.as_bytes()).expect("write response");
        }
    });

    let client = BitflyerApiClient::new(
        format!("http://{}", addr),
        ApiCredentials {
            api_key: "k".to_string(),
            api_secret: "s".to_string(),
        },
    )
    .expect("client should construct");

    let got = client
        .fetch_executions_window(&FetchWindowOptions {
            product_code: "BTC_JPY".to_string(),
            count_per_page: 100,
            max_pages: 10,
            since: None,
            until: None,
        })
        .expect("window fetch should succeed");

    server.join().expect("server thread join");

    assert_eq!(
        calls.load(Ordering::SeqCst),
        2,
        "expected exactly 2 page fetches"
    );
    assert_eq!(
        got.len(),
        2,
        "duplicate execution id should be de-duplicated"
    );
    assert!(got.iter().any(|e| e.id == 100));
    assert!(got.iter().any(|e| e.id == 99));
}

#[test]
fn bitflyer_api_window_rejects_zero_count_per_page_before_network() {
    let client = BitflyerApiClient::new(
        "https://api.bitflyer.com/".to_string(),
        ApiCredentials {
            api_key: "k".to_string(),
            api_secret: "s".to_string(),
        },
    )
    .expect("client should construct");

    let err = client
        .fetch_executions_window(&FetchWindowOptions {
            product_code: "BTC_JPY".to_string(),
            count_per_page: 0,
            max_pages: 1,
            since: None,
            until: None,
        })
        .expect_err("count_per_page=0 should fail before network");

    assert!(err.contains("count_per_page must be > 0"));
}

#[test]
fn bitflyer_api_window_rejects_zero_max_pages_before_network() {
    let client = BitflyerApiClient::new(
        "https://api.bitflyer.com/".to_string(),
        ApiCredentials {
            api_key: "k".to_string(),
            api_secret: "s".to_string(),
        },
    )
    .expect("client should construct");

    let err = client
        .fetch_executions_window(&FetchWindowOptions {
            product_code: "BTC_JPY".to_string(),
            count_per_page: 100,
            max_pages: 0,
            since: None,
            until: None,
        })
        .expect_err("max_pages=0 should fail before network");

    assert!(err.contains("max_pages must be > 0"));
}

#[test]
fn bitflyer_api_filter_executions_by_time_window_is_inclusive() {
    let raw = r#"[
      {"id": 1, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-01-01T00:00:00Z"},
      {"id": 2, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-02-01T00:00:00Z"},
      {"id": 3, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-03-01T00:00:00Z"}
    ]"#;
    let executions: Vec<Execution> = serde_json::from_str(raw).unwrap();

    let since = DateTime::parse_from_rfc3339("2026-02-01T00:00:00Z")
        .unwrap()
        .with_timezone(&Utc);
    let until = DateTime::parse_from_rfc3339("2026-02-01T00:00:00Z")
        .unwrap()
        .with_timezone(&Utc);

    let filtered = filter_executions_by_time(&executions, Some(since), Some(until));
    assert_eq!(filtered.len(), 1);
    assert_eq!(filtered[0].id, 2);
}

#[test]
fn bitflyer_api_fetch_page_retries_on_5xx_then_succeeds() {
    let listener = TcpListener::bind("127.0.0.1:0").expect("bind test server");
    let addr = listener.local_addr().expect("local addr");
    let calls = Arc::new(AtomicUsize::new(0));
    let calls_bg = Arc::clone(&calls);

    let server = thread::spawn(move || {
        for stream in listener.incoming().take(3) {
            let mut stream = stream.expect("incoming stream");
            let idx = calls_bg.fetch_add(1, Ordering::SeqCst);

            let mut buf = [0_u8; 2048];
            let _ = stream.read(&mut buf);

            let (status, body) = if idx < 2 {
                ("500 Internal Server Error", "[]")
            } else {
                (
                    "200 OK",
                    r#"[{"id": 1, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-01-01T00:00:00Z"}]"#,
                )
            };

            let resp = format!(
                "HTTP/1.1 {}\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                status,
                body.len(),
                body
            );
            stream.write_all(resp.as_bytes()).expect("write response");
        }
    });

    let client = BitflyerApiClient::new(
        format!("http://{}", addr),
        ApiCredentials {
            api_key: "k".to_string(),
            api_secret: "s".to_string(),
        },
    )
    .expect("client should construct");

    let page = client
        .fetch_executions_page(&FetchOptions {
            product_code: "BTC_JPY".to_string(),
            count: 10,
            before: None,
            after: None,
        })
        .expect("should succeed after retries");

    server.join().expect("server thread join");

    assert_eq!(calls.load(Ordering::SeqCst), 3, "expected 3 attempts");
    assert_eq!(page.len(), 1);
    assert_eq!(page[0].id, 1);
}

#[test]
fn bitflyer_api_fetch_page_honors_retry_after_on_429() {
    let listener = TcpListener::bind("127.0.0.1:0").expect("bind test server");
    let addr = listener.local_addr().expect("local addr");
    let calls = Arc::new(AtomicUsize::new(0));
    let calls_bg = Arc::clone(&calls);

    let server = thread::spawn(move || {
        for stream in listener.incoming().take(2) {
            let mut stream = stream.expect("incoming stream");
            let idx = calls_bg.fetch_add(1, Ordering::SeqCst);

            let mut buf = [0_u8; 2048];
            let _ = stream.read(&mut buf);

            let (status, extra_headers, body) = if idx == 0 {
                ("429 Too Many Requests", "Retry-After: 1\r\n", "{}")
            } else {
                (
                    "200 OK",
                    "",
                    r#"[{"id": 10, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-01-01T00:00:00Z"}]"#,
                )
            };

            let resp = format!(
                "HTTP/1.1 {}\r\n{}Content-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                status,
                extra_headers,
                body.len(),
                body
            );
            stream.write_all(resp.as_bytes()).expect("write response");
        }
    });

    let client = BitflyerApiClient::new(
        format!("http://{}", addr),
        ApiCredentials {
            api_key: "k".to_string(),
            api_secret: "s".to_string(),
        },
    )
    .expect("client should construct");

    let started = Instant::now();
    let page = client
        .fetch_executions_page(&FetchOptions {
            product_code: "BTC_JPY".to_string(),
            count: 10,
            before: None,
            after: None,
        })
        .expect("should succeed after retry");
    let elapsed = started.elapsed();

    server.join().expect("server thread join");

    assert_eq!(calls.load(Ordering::SeqCst), 2, "expected 2 attempts");
    assert_eq!(page.len(), 1);
    assert_eq!(page[0].id, 10);
    let elapsed_ms = elapsed.as_millis();
    assert!(
        (500..=1500).contains(&elapsed_ms),
        "expected retry delay to be roughly 1s (500ms–1500ms), got {:?}",
        elapsed
    );
}

#[test]
fn bitflyer_api_fetch_page_honors_retry_after_http_date_on_429() {
    let listener = TcpListener::bind("127.0.0.1:0").expect("bind test server");
    let addr = listener.local_addr().expect("local addr");
    let calls = Arc::new(AtomicUsize::new(0));
    let calls_bg = Arc::clone(&calls);

    let retry_at = (Utc::now() + ChronoDuration::seconds(2))
        .format("%a, %d %b %Y %H:%M:%S GMT")
        .to_string();

    let server = thread::spawn(move || {
        for stream in listener.incoming().take(2) {
            let mut stream = stream.expect("incoming stream");
            let idx = calls_bg.fetch_add(1, Ordering::SeqCst);

            let mut buf = [0_u8; 2048];
            let _ = stream.read(&mut buf);

            let (status, extra_headers, body) = if idx == 0 {
                (
                    "429 Too Many Requests",
                    format!("Retry-After: {}\r\n", retry_at),
                    "{}".to_string(),
                )
            } else {
                (
                    "200 OK",
                    "".to_string(),
                    r#"[{"id": 11, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-01-01T00:00:00Z"}]"#
                        .to_string(),
                )
            };

            let resp = format!(
                "HTTP/1.1 {}\r\n{}Content-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                status,
                extra_headers,
                body.len(),
                body
            );
            stream.write_all(resp.as_bytes()).expect("write response");
        }
    });

    let client = BitflyerApiClient::new(
        format!("http://{}", addr),
        ApiCredentials {
            api_key: "k".to_string(),
            api_secret: "s".to_string(),
        },
    )
    .expect("client should construct");

    let started = Instant::now();
    let page = client
        .fetch_executions_page(&FetchOptions {
            product_code: "BTC_JPY".to_string(),
            count: 10,
            before: None,
            after: None,
        })
        .expect("should succeed after retry");
    let elapsed = started.elapsed();

    server.join().expect("server thread join");

    assert_eq!(calls.load(Ordering::SeqCst), 2, "expected 2 attempts");
    assert_eq!(page.len(), 1);
    assert_eq!(page[0].id, 11);
    let elapsed_ms = elapsed.as_millis();
    assert!(
        (500..=2500).contains(&elapsed_ms),
        "expected retry delay from HTTP-date Retry-After, got {:?}",
        elapsed
    );
}

#[test]
fn bitflyer_api_fetch_page_falls_back_when_retry_after_invalid() {
    let listener = TcpListener::bind("127.0.0.1:0").expect("bind test server");
    let addr = listener.local_addr().expect("local addr");
    let calls = Arc::new(AtomicUsize::new(0));
    let calls_bg = Arc::clone(&calls);

    let server = thread::spawn(move || {
        for stream in listener.incoming().take(2) {
            let mut stream = stream.expect("incoming stream");
            let idx = calls_bg.fetch_add(1, Ordering::SeqCst);

            let mut buf = [0_u8; 2048];
            let _ = stream.read(&mut buf);

            let (status, extra_headers, body) = if idx == 0 {
                ("429 Too Many Requests", "Retry-After: invalid\r\n", "{}")
            } else {
                (
                    "200 OK",
                    "",
                    r#"[{"id": 12, "side": "BUY", "price": "100", "size": "1", "exec_date": "2026-01-01T00:00:00Z"}]"#,
                )
            };

            let resp = format!(
                "HTTP/1.1 {}\r\n{}Content-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                status,
                extra_headers,
                body.len(),
                body
            );
            stream.write_all(resp.as_bytes()).expect("write response");
        }
    });

    let client = BitflyerApiClient::new(
        format!("http://{}", addr),
        ApiCredentials {
            api_key: "k".to_string(),
            api_secret: "s".to_string(),
        },
    )
    .expect("client should construct");

    let started = Instant::now();
    let page = client
        .fetch_executions_page(&FetchOptions {
            product_code: "BTC_JPY".to_string(),
            count: 10,
            before: None,
            after: None,
        })
        .expect("should succeed after fallback retry");
    let elapsed = started.elapsed();

    server.join().expect("server thread join");

    assert_eq!(calls.load(Ordering::SeqCst), 2, "expected 2 attempts");
    assert_eq!(page.len(), 1);
    assert_eq!(page[0].id, 12);
    assert!(
        elapsed.as_millis() >= 200,
        "expected fallback backoff delay when Retry-After invalid, got {:?}",
        elapsed
    );
}

#[test]
fn bitflyer_api_fetch_page_does_not_retry_on_401() {
    let listener = TcpListener::bind("127.0.0.1:0").expect("bind test server");
    let addr = listener.local_addr().expect("local addr");
    let calls = Arc::new(AtomicUsize::new(0));
    let calls_bg = Arc::clone(&calls);

    let server = thread::spawn(move || {
        for stream in listener.incoming().take(1) {
            let mut stream = stream.expect("incoming stream");
            let _ = calls_bg.fetch_add(1, Ordering::SeqCst);

            let mut buf = [0_u8; 2048];
            let _ = stream.read(&mut buf);

            let body = "{}";
            let resp = format!(
                "HTTP/1.1 401 Unauthorized\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                body.len(),
                body
            );
            stream.write_all(resp.as_bytes()).expect("write response");
        }
    });

    let client = BitflyerApiClient::new(
        format!("http://{}", addr),
        ApiCredentials {
            api_key: "k".to_string(),
            api_secret: "s".to_string(),
        },
    )
    .expect("client should construct");

    let err = client
        .fetch_executions_page(&FetchOptions {
            product_code: "BTC_JPY".to_string(),
            count: 10,
            before: None,
            after: None,
        })
        .expect_err("401 should fail immediately");

    server.join().expect("server thread join");

    assert_eq!(calls.load(Ordering::SeqCst), 1, "expected single attempt");
    assert!(err.contains("authentication failed"));
}
