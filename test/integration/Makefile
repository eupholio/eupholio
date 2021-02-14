
ETL=../../bin/etl
QUERY=../../bin/query
METHOD=mam

test:
	make -C ../.. db-init MYSQL_PORT=3307
	go test -v -count 1 ./...

run_test:
	make -C ../.. db-init
	$(ETL) load yahoofinance historical_price ../testdata/BTC-JPY.csv
	$(ETL) load yahoofinance historical_price ../testdata/ETH-JPY.csv
	$(ETL) load yahoofinance historical_price ../testdata/BTC-USD.csv
	$(ETL) load yahoofinance historical_price ../testdata/ETH-USD.csv
	# etl
	make -C ../.. db-clear
	$(ETL) import bf ../testdata/TradeHistory.csv --overwrite
	$(ETL) import coincheck ../testdata/01_100000_2017_1.csv --overwrite
	$(ETL) import bittrex ../testdata/BittrexOrderHistory.csv --overwrite
	$(ETL) import bittrex --filetype deposit ../testdata/BittrexDeposit.csv --overwrite
	$(ETL) import bittrex --filetype withdraw ../testdata/BittrexWithdraw.csv --overwrite
	$(ETL) translate
	# calc
	$(ETL) calculate --year 2017 --method ${METHOD}
	$(QUERY) transaction --year 2017
	$(QUERY) balance --year 2017
	$(ETL) calculate --year 2018 --method ${METHOD}
	$(QUERY) transaction --year 2018
	$(QUERY) balance --year 2018
	$(ETL) calculate --year 2019 --method ${METHOD}
	$(QUERY) transaction --year 2019
	$(QUERY) balance --year 2019
