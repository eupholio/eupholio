#!/bin/bash

set -x
set -e

datasource=${1:-default}

DOWNLOAD="go run ./cmd/etl download"

case $datasource in
cryptodatadownload)
	${DOWNLOAD} cryptodatadownload historical_price --dir pricedata/cryptodatadownload
    ;;
coingecko)
	${DOWNLOAD} coingecko historical_price --dir pricedata/coingecko
    ;;
yahoofinance)
	${DOWNLOAD} yahoofinance historical_price --dir pricedata/yahoofinance
	${DOWNLOAD} yahoofinance historical_price --dir pricedata/yahoofinance --symbol USD --fiat JPY
	${DOWNLOAD} yahoofinance historical_price --dir pricedata/yahoofinance --symbol EUR --fiat JPY
    ;;
default)
	${DOWNLOAD} coingecko historical_price --dir pricedata/coingecko
	${DOWNLOAD} yahoofinance historical_price --dir pricedata/yahoofinance --symbol USD --fiat JPY
	${DOWNLOAD} yahoofinance historical_price --dir pricedata/yahoofinance --symbol EUR --fiat JPY
esac
