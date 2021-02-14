#!/bin/bash

make db-clear

./bin/config costmethod --year 2008 --method mam

mkdir -p history
./bin/etl import bf history/bitflyer/TradeHistory.csv
./bin/etl import coincheck history/coincheck/*.csv
./bin/etl import bittrex history/bittrex/BittrexOrderHistory_*.csv
./bin/etl import poloniex history/poloniex/*.csv

./bin/etl translate
./bin/etl calculate

./bin/query transaction --year 2020
./bin/query balance --year 2020
