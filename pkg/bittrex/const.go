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

package bittrex

const WalletCode = "BITTREX"

const (
	LimitSell = "LIMIT_SELL"
	LimitBuy  = "LIMIT_BUY"
)

const (
	OrderTypeUnknown int = iota
	OrderTypeLimitSell
	OrderTypeLimitBuy
)

// CokumnID is index type for columns
type ColumnID int

// Column Index
const (
	UUID ColumnID = iota
	Exchange
	TimeStamp
	OrderType
	Limit
	Quantity
	QuantityRemaining
	Commission
	Price
	PricePerUnit
	IsConditional
	Condition
	ConditionTarget
	ImmediateOrCancel
	Closed
	TimeInForceTypeID
	TimeInForce
)

// Column names
const (
	UUIDColumn              = "Uuid"
	ExchangeColumn          = "Exchange"
	TimeStampColumn         = "TimeStamp"
	OrderTypeColumn         = "OrderType"
	LimitColumn             = "Limit"
	QuantityColumn          = "Quantity"
	QuantityRemainingColumn = "QuantityRemaining"
	CommissionColumn        = "Commission"
	PriceColumn             = "Price"
	PricePerUnitColumn      = "PricePerUnit"
	IsConditionalColumn     = "IsConditional"
	ConditionColumn         = "Condition"
	ConditionTargetColumn   = "ConditionTarget"
	ImmediateOrCancelColumn = "ImmediateOrCancel"
	ClosedColumn            = "Closed"
	TimeInForceTypeIDColumn = "TimeInForceTypeId"
	TimeInForceColumn       = "TimeInForce"
)

var columnNames = []string{
	UUIDColumn,
	ExchangeColumn,
	TimeStampColumn,
	OrderTypeColumn,
	LimitColumn,
	QuantityColumn,
	QuantityRemainingColumn,
	CommissionColumn,
	PriceColumn,
	PricePerUnitColumn,
	IsConditionalColumn,
	ConditionColumn,
	ConditionTargetColumn,
	ImmediateOrCancelColumn,
	ClosedColumn,
	TimeInForceTypeIDColumn,
	TimeInForceColumn,
}

var columnNamesSet map[string]ColumnID

func init() {
	columnNamesSet = make(map[string]ColumnID)
	for i, cn := range columnNames {
		columnNamesSet[cn] = ColumnID(i)
	}
}
