
/* Master Data Tables */

DROP TABLE IF EXISTS symbols;

CREATE TABLE symbols (
    symbol CHAR(10) PRIMARY KEY,
    `name` VARCHAR(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS market_price;

CREATE TABLE market_price (
    source VARCHAR(20),
    currency CHAR(10) NOT NULL,
    `time` DATETIME NOT NULL,
    base_currency CHAR(10) NOT NULL,
    price DECIMAL(20, 10) NOT NULL,
    PRIMARY KEY (source, base_currency, currency, `time`)
);
