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

package bitflyer

const WalletCode = "BF"
const FiatCode = "JPY"

const (
	TrTypeUnknown int = iota
	TrTypeBuy
	TrTypeSell
	TrTypeReceive
	TrTypeDeposit
	TrTypeWithdraw
	TrTypeTransfer
	TrTypeFee
)

type ColumnID int

const (
	TrDate ColumnID = iota
	Currency
	TrType
	TrPrice
	Currency1
	Currency1Quantity
	Fee
	Currency1JPYRate
	Currency2
	Currency2Quantity
	DealType
	OrderID
	Remarks
	numOfColumns
)

const (
	En = "en"
	Jp = "jp"
)

var columnNames = map[string][]string{
	En: {
		"Trade Date",
		"Product",
		"Trade Type",
		"Traded Price",
		"Currency 1",
		"Amount (Currency 1)",
		"Fee",
		"JPY Rate (Currency 1)",
		"Currency 2",
		"Amount (Currency 2)",
		"Counter Party",
		"Order ID",
		"Details",
	},
	Jp: {
		"取引日時",
		"通貨",
		"取引種別",
		"取引価格",
		"通貨1",
		"通貨1数量",
		"手数料",
		"通貨1の対円レート",
		"通貨2",
		"通貨2数量",
		"自己・媒介",
		"注文 ID",
		"備考",
	},
}

var columnNamesSet = make(map[string]map[string]ColumnID)

func init() {
	for k, l := range columnNames {
		columnNamesSet[k] = make(map[string]ColumnID)
		for i, cn := range l {
			columnNamesSet[k][cn] = ColumnID(i)
		}
	}
}
