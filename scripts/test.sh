#!/bin/bash

set -x
set -e

FORMAT=${FORMAT:-"csv"}
METHOD=${METHOD:-"wam"}

CONFIG="go run ./cmd/config"
ETL="go run ./cmd/etl"
QUERY="go run ./cmd/query"

${CONFIG} costmethod --year 2008 --method ${METHOD}

${ETL} import bf testdata/bitflyer/TradeHistory.csv --overwrite
${ETL} import coincheck testdata/coincheck/*.csv
${ETL} import bittrex testdata/bittrex/BittrexOrderHistory_*.csv
${ETL} import bittrex --filetype deposit testdata/bittrex/BittrexDeposit.csv
${ETL} import bittrex --filetype withdraw testdata/bittrex/BittrexWithdraw.csv
${ETL} import poloniex testdata/poloniex/*.csv
${ETL} import cryptact --filetype custom testdata/cryptact/Custom.csv
${ETL} translate

for i in 2017 2018 2019 2020 2021; do
  ${ETL} calculate --year $i
done

for i in 2017 2018 2019 2020 2021; do
  ${QUERY} transaction --year $i --format ${FORMAT} > transaction$i.csv;
  ${QUERY} balance --year $i --format ${FORMAT} > balance$i.csv;
done
