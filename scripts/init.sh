#!/bin/bash

set -x
set -e

LOAD="go run ./cmd/etl load"

${LOAD} coingecko historical_price pricedata/coingecko/*.csv
${LOAD} yahoofinance historical_price pricedata/yahoofinance/*.csv
${LOAD} cryptodatadownload historical_price pricedata/cryptodatadownload/*.csv
