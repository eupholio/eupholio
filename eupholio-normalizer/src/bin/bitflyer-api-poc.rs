use std::process;

use eupholio_normalizer::bitflyer_api::{BitflyerApiClient, FetchOptions};

fn main() {
    if let Err(err) = run() {
        eprintln!("error: {err}");
        process::exit(1);
    }
}

fn run() -> Result<(), String> {
    let args: Vec<String> = std::env::args().collect();
    let product_code = args
        .get(1)
        .cloned()
        .unwrap_or_else(|| "BTC_JPY".to_string());
    let count = args
        .get(2)
        .and_then(|v| v.parse::<usize>().ok())
        .unwrap_or(100);

    let client = BitflyerApiClient::from_env()?;
    let executions = client.fetch_executions_page(&FetchOptions {
        product_code,
        count,
        before: None,
        after: None,
    })?;

    println!("fetched executions: {}", executions.len());
    for ex in executions.iter().take(3) {
        println!(
            "id={} side={} price={} size={} exec_date={}",
            ex.id, ex.side, ex.price, ex.size, ex.exec_date
        );
    }

    Ok(())
}
