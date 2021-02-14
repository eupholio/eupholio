#!/bin/bash

set -x
set -e

make db-init

mkdir -p pricedata
./bin/etl download coingecko historical_price --dir pricedata/coingecko
./bin/etl load coingecko historical_price pricedata/coingecko/*.csv
./bin/etl download yahoofinance historical_price --dir pricedata/yahoofinance --symbol USD --fiat JPY
./bin/etl download yahoofinance historical_price --dir pricedata/yahoofinance --symbol EUR --fiat JPY
./bin/etl load yahoofinance historical_price pricedata/yahoofinance/*.csv
