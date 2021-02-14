/*
 * Eupholio - A portfolio tracker tool for cryptocurrency
 * Copyright (C) 2021 Kiyoshi Nakao
 *
 * This file is part of Eupholio.
 *
 * Eupholio is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * Eupholio is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Eupholio.  If not, see <http://www.gnu.org/licenses/>.
 */

package coincheck

const (
	WalletCode = "COINCHECK"
	FiatCode = "JPY"
)

const (
	OperationReceived = "Received"
	OperationLimitOrder = "Limit Order"
	OperationCompletedTradingContracts = "Completed trading contracts"
	OperationSent = "Sent"
	OperationBankWithdrawal = "Bank Withdrawal"
	OperationCancelLimitOrder = "Cancel Limit Order"
)

type ColumnID int

// Column Index
const (
	ID ColumnID = iota
	Time
	Operation
	Amount
	TradingCurrency
	Price
	OriginalCurrency
	Fee
	Comment
)

const (
	IDColumn               = "id"
	TimeColumn             = "time"
	OperationColumn        = "operation"
	AmountColumn           = "amount"
	TradingCurrencyColumn  = "trading_currency"
	PriceColumn            = "price"
	OriginalCurrencyColumn = "original_currency"
	FeeColumn              = "fee"
	CommentColumn          = "comment"
)

var columnNames = []string{
	IDColumn,
	TimeColumn,
	OperationColumn,
	AmountColumn,
	TradingCurrencyColumn,
	PriceColumn,
	OriginalCurrencyColumn,
	FeeColumn,
	CommentColumn,
}

var columnNamesSet map[string]ColumnID

func init() {
	columnNamesSet = make(map[string]ColumnID)
	for i, cn := range columnNames {
		columnNamesSet[cn] = ColumnID(i)
	}
}
