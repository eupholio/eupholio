# Eupholio

Eupholio is a portfolio tracking tool for cryptocurrencies. You can load your trading history
downloaded from multiple exchanges and calculate the total amount of your portfolio.

This software is under development and not tested well so you can't use it for practical use. 

## Features

The main feature of this software is simple portfolio tracking. It just accumlates opened and 
closed positions and show the total quantity of assets. Besides position tracking Eupholio 
calculates the aquisition cost and profit of each currency in a very simple way. The result 
might not be accurate because what should be considered as "cost" is ambiguous and the historical
price data may not be presice enough (and the program may have some bugs). You can just use the 
result as approximate information of your portfolio.

- supported fiat currencies
  - JPY
- supported cost calculation methods
  - weighted avarage method
  - moving average method
- supported wallets and exchanges (margin trading is not supported)
  - Bittrex
  - Poloniex
  - BitFlyer
  - Coincheck

## How to build

```
$ make
```

## Setup

Before using this software you need to set up MySQL database. If you have installed docker, you can setup it 
with docker-compose.yml.

```
$ docker-compose up -d
```

Once MySQL server started, you can create tables.  

```
$ make db-init
```

You need to download and setup histrical market price data.

```bash
mkdir -p pricedata
./bin/etl download coingecko historical_price --dir pricedata/coingecko
./bin/etl load coingecko historical_price pricedata/coingecko/*.csv
./bin/etl download yahoofinance historical_price --dir pricedata/yahoofinance --symbol USD --fiat JPY # optional
./bin/etl download yahoofinance historical_price --dir pricedata/yahoofinance --symbol EUR --fiat JPY # optional
./bin/etl load yahoofinance historical_price pricedata/yahoofinance/*.csv # optional
```

## Usage

Now you can import your trading history files downloaded from exchanges.
Before running import command, you need to place the files somewhere.

```
make db-clear # XXX
```

```bash
mkdir -p history
./bin/etl import bf history/bitflyer/TradeHistory.csv # optional
./bin/etl import coincheck history/coincheck/*.csv # optional
./bin/etl import bittrex history/bittrex/BittrexOrderHistory_*.csv # optional
./bin/etl import poloniex history/poloniex/*.csv # optional
```

```bash
./bin/config costmethod --year 2008 --method mam
./bin/etl translate
./bin/etl calculate
```

```bash
./bin/query transaction --year 2020
./bin/query balance --year 2020
```

## TODO

- Ethereum wallet support
- DEX support

## Lisence

Copyright (c) Kiyoshi Nakao, as shown by the AUTHORS file.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
