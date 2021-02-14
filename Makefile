
MYSQL_PORT=3306
MYSQL=mysql --defaults-extra-file=mysql.conf --ssl-mode=DISABLED -P $(MYSQL_PORT)
DATASOURCE=default

build:
	go build -o bin/config ./cmd/config
	go build -o bin/etl ./cmd/etl
	go build -o bin/query ./cmd/query

.PHONY: test
test:
	go test -v ./...

.PHONY: init
init: db-init
	./scripts/init.sh

.PHONY: db-init
db-init:
	cat resources/master.sql resources/symbols.sql | $(MYSQL) eupholio
	cat resources/schema.sql | $(MYSQL) eupholio

.PHONY: db-clear	
db-clear:
	cat resources/schema.sql | $(MYSQL) eupholio

.PHONY: gen-model
gen-model:
	sqlboiler mysql

.PHONY: download
download:
	./scripts/download.sh

.PHONY: add-copyright
add-copyright:
	copyright-header \
	--add-path pkg:cmd:test \
	--out-dir . \
	--license AGPL3 \
	--copyright-year 2021 \
	--copyright-holder 'Kiyoshi Nakao' \
	--copyright-software 'Eupholio' \
	--copyright-software-description 'A portfolio tracker tool for cryptocurrency'

.PHONY: license-check
license-check:
	go-licenses check github.com/eupholio/eupholio
