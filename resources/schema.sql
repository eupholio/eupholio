
/* Configuration Tables */

DROP TABLE IF EXISTS config;

CREATE TABLE config (
    id INT NOT NULL,
    year INT NOT NULL,
    cost_method VARCHAR(20) NOT NULL,
    PRIMARY KEY (id, year)
);

/* Transaction Tables */

DROP TABLE IF EXISTS transactions;

CREATE TABLE transactions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    `time` DATETIME NOT NULL,
    wallet_code VARCHAR(10) NOT NULL,
    wallet_tid INT NOT NULL,
    `description` VARCHAR(255) NOT NULL,
    INDEX(`time`)
);

DROP TABLE IF EXISTS `event`;

CREATE TABLE `event` (
    id INT PRIMARY KEY AUTO_INCREMENT,
    transaction_id INT NOT NULL,
    `time` DATETIME NOT NULL,
    `type` CHAR(10) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    quantity DECIMAL(20, 10) NOT NULL,
    base_currency VARCHAR(10) NOT NULL,
    base_quantity DECIMAL(20, 10) NOT NULL,
    INDEX (time),
    INDEX (currency, time),
    INDEX (transaction_id)
);

DROP TABLE IF EXISTS `entry`;

CREATE TABLE `entry` (
    id INT PRIMARY KEY,
    transaction_id INT NOT NULL,
    `time` DATETIME NOT NULL,
    `type` CHAR(10) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    quantity DECIMAL(20, 10) NOT NULL,
    position DECIMAL(20, 10) NOT NULL,
    fiat_currency VARCHAR(10) NOT NULL,
    fiat_quantity DECIMAL(20, 10) NOT NULL,
    commission DECIMAL(20, 10) DEFAULT NULL,
    price DECIMAL(20, 10) DEFAULT NULL,
    INDEX (time),
    INDEX (currency, time),
    INDEX (transaction_id)
);

DROP TABLE IF EXISTS balance;

CREATE TABLE balance (
    id INT PRIMARY KEY AUTO_INCREMENT,
    year INT NOT NULL,
    currency VARCHAR(10) NOT NULL,
    beginning_quantity DECIMAL(20, 10) NOT NULL,
    open_quantity DECIMAL(20, 10) NOT NULL,
    close_quantity DECIMAL(20, 10) NOT NULL,
    price DECIMAL(20, 10) NOT NULL,
    quantity DECIMAL(20, 10) NOT NULL,
    profit DECIMAL(20, 10) NOT NULL
);

DROP TABLE IF EXISTS method;

CREATE TABLE method (
    year INT PRIMARY KEY,
    method CHAR(10) NOT NULL
);

/* Wallet Tables */

/* Bitflyer */

DROP TABLE IF EXISTS bf_transactions;

CREATE TABLE bf_transactions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    tr_date DATETIME NOT NULL,
    currency VARCHAR(10) NOT NULL,
    tr_type INT(1) NOT NULL,
    tr_price DECIMAL(20, 10) NOT NULL,
    currency1 VARCHAR(10) NOT NULL,
    currency1_quantity DECIMAL(20, 10) NOT NULL,
    fee DECIMAL(20, 10) NOT NULL,
    currency1_jpy_rate DECIMAL(20, 10),
    currency2 VARCHAR(10),
    currency2_quantity DECIMAL(20, 10) NOT NULL,
    deal_type INT(1),
    order_id VARCHAR(100) NOT NULL,
    remarks VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS bf_orders;

CREATE TABLE bf_orders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    order_id VARCHAR(100) NOT NULL,
    transaction_id INT NOT NULL,
    tr_date DATETIME NOT NULL
);

DROP TABLE IF EXISTS coincheck_history;

CREATE TABLE coincheck_history (
    id INT PRIMARY KEY AUTO_INCREMENT,
    id_code VARCHAR(16) NOT NULL,
    `time` DATETIME NOT NULL,
    operation VARCHAR(30) NOT NULL,
    amount DECIMAL(20, 10) NOT NULL,
    trading_currency VARCHAR(10) NOT NULL,
    price DECIMAL(20, 10),
    original_currency VARCHAR(10),
    fee DECIMAL(20, 10),
    comment VARCHAR(255) NOT NULL,
    UNIQUE(id_code),
    INDEX(`time`)
);

/* Bittrex */

DROP TABLE IF EXISTS bittrex_order_history;

CREATE TABLE bittrex_order_history (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) NOT NULL,
    exchange VARCHAR(20) NOT NULL,
    `timestamp` DATETIME NOT NULL,
    order_type INT NOT NULL,
    `limit` DECIMAL(20, 10) NOT NULL,
    quantity DECIMAL(20, 10) NOT NULL,
    quantity_remaining DECIMAL(20, 10) NOT NULL,
    commission DECIMAL(20, 10) NOT NULL,
    price DECIMAL(20, 10) NOT NULL,
    price_per_unit DECIMAL(20, 10) NOT NULL,
    is_conditional TINYINT(1) NOT NULL,
    `condition` VARCHAR(100),
    condition_target DECIMAL(20, 10),
    immediate_or_cancel TINYINT(1) NOT NULL,
    `closed` DATETIME NOT NULL,
    time_in_force_type_id INT NOT NULL,
    time_in_force TEXT,
    INDEX (uuid),
    INDEX (`timestamp`)
);

DROP TABLE IF EXISTS bittrex_deposit_history;

CREATE TABLE bittrex_deposit_history (
    id INT PRIMARY KEY AUTO_INCREMENT,
    `timestamp` DATETIME NOT NULL,
    currency VARCHAR(10) NOT NULL,
    quantity DECIMAL(20, 10) NOT NULL,
    `status` VARCHAR(10) NOT NULL,
    INDEX (`timestamp`)
);

DROP TABLE IF EXISTS bittrex_withdraw_history;

CREATE TABLE bittrex_withdraw_history (
    id INT PRIMARY KEY AUTO_INCREMENT,
    `timestamp` DATETIME NOT NULL,
    currency VARCHAR(10) NOT NULL,
    quantity DECIMAL(20, 10) NOT NULL,
    `status` VARCHAR(10) NOT NULL,
    INDEX (`timestamp`)
);

/* Poloniex */

DROP TABLE IF EXISTS poloniex_trades;

CREATE TABLE poloniex_trades (
    id INT PRIMARY KEY AUTO_INCREMENT,
    `date` DATETIME NOT NULL,
    market VARCHAR(20) NOT NULL,
    `type` VARCHAR(10) NOT NULL,
    price DECIMAL(20, 10) NOT NULL,
    amount DECIMAL(20, 10) NOT NULL,
    total DECIMAL(20, 10) NOT NULL,
    fee VARCHAR(20) NOT NULL,
    order_number BIGINT NOT NULL,
    base_total_less_fee DECIMAL(20, 10) NOT NULL,
    quote_total_less_fee DECIMAL(20, 10) NOT NULL,
    fee_currency VARCHAR(10) NOT NULL,
    fee_total DECIMAL(20, 10) NOT NULL,
    INDEX (`date`)
);

DROP TABLE IF EXISTS poloniex_deposits;

CREATE TABLE poloniex_deposits (
    id INT PRIMARY KEY AUTO_INCREMENT,
    `date` DATETIME NOT NULL,
    currency VARCHAR(10) NOT NULL,
    amount DECIMAL(20, 10) NOT NULL,
    `address` VARCHAR(100) NOT NULL,
    `status` VARCHAR(100) NOT NULL,
    INDEX (`date`)
);

DROP TABLE IF EXISTS poloniex_withdrawals;

CREATE TABLE poloniex_withdrawals (
    id INT PRIMARY KEY AUTO_INCREMENT,
    `date` DATETIME NOT NULL,
    currency VARCHAR(10) NOT NULL,
    amount DECIMAL(20, 10) NOT NULL,
    fee_deducted DECIMAL(20, 10) NOT NULL,
    amount_minus_fee DECIMAL(20, 10) NOT NULL,
    `address` VARCHAR(100) NOT NULL,
    `status` VARCHAR(100) NOT NULL,
    INDEX (`date`)
);

DROP TABLE IF EXISTS poloniex_distributions;

CREATE TABLE poloniex_distributions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    `date` DATETIME NOT NULL,
    currency VARCHAR(10) NOT NULL,
    amount DECIMAL(20, 10) NOT NULL,
    wallet VARCHAR(100) NOT NULL,
    INDEX (`date`)
);

/* Cryptact */

DROP TABLE IF EXISTS cryptact_custom;

CREATE TABLE cryptact_custom (
    id INT PRIMARY KEY AUTO_INCREMENT,
    `timestamp` DATETIME NOT NULL,
    `action` VARCHAR(20) NOT NULL,
    `source` VARCHAR(100) NOT NULL,
    base VARCHAR(10) NOT NULL,
    volume DECIMAL(20, 10) NOT NULL,
    price DECIMAL(20, 10),
    `counter` VARCHAR(10) NOT NULL,
    fee DECIMAL(20, 10) NOT NULL,
    fee_ccy VARCHAR(10) NOT NULL,
    INDEX (`timestamp`)
);

