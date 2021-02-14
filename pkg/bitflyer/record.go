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

import (
	"strings"

	"github.com/ericlagergren/decimal"
)

var trTypeIDMap = map[string]int{
	"買い":   TrTypeBuy,
	"売り":   TrTypeSell,
	"受取":   TrTypeReceive,
	"入金":   TrTypeDeposit,
	"外部送付": TrTypeTransfer,
	"手数料":  TrTypeFee,
}

type Record []string

func (r Record) Get(id ColumnID) string {
	return r[int(id)]
}

func (r Record) Currency() string {
	return r.Get(Currency)
}

func (r Record) TrType() int {
	if id, ok := trTypeIDMap[r.Get(TrType)]; ok {
		return id
	} else {
		return TrTypeUnknown
	}
}

func (r Record) TrPrice() *decimal.Big {
	return r.parseNullDecimal(TrPrice)
}

func (r Record) Currency1() string {
	return r.Get(Currency1)
}

func (r Record) Currency1Quantity() *decimal.Big {
	return r.parseNullDecimal(Currency1Quantity)
}

func (r Record) Fee() *decimal.Big {
	return r.parseNullDecimal(Fee)
}

func (r Record) Currency1JPYRate() *decimal.Big {
	return r.parseNullDecimal(Currency1JPYRate)
}

func (r Record) Currency2() string {
	return r.Get(Currency2)
}

func (r Record) Currency2Quantity() *decimal.Big {
	return r.parseNullDecimal(Currency2Quantity)
}

func (r Record) DealType() int {
	switch r.Get(DealType) {
	case "自己":
		return DealTypeSelf
	case "媒介":
		return DealTypeMed
	case "":
		return DealTypeNone
	}
	return DealTypeUnkown
}

func (r Record) OrderID() string {
	return r.Get(OrderID)
}

func (r Record) Remarks() string {
	return r.Get(Remarks)
}

func (r Record) parseDecimal(id ColumnID) (*decimal.Big, bool) {
	s := r.Get(id)
	s = strings.ReplaceAll(s, ",", "")
	if b, ok := new(decimal.Big).SetString(s); ok {
		return b, true
	} else {
		return nil, false
	}
}

func (r Record) parseNullDecimal(id ColumnID) *decimal.Big {
	b, _ := r.parseDecimal(id)
	return b
}

func MakeRecords(lang string, head []string, rows [][]string) ([]Record, error) {
	m := make(map[int]ColumnID)
	for i, k := range head {
		if columnID, ok := columnNamesSet[lang][k]; ok {
			m[i] = columnID
		}
	}
	var ret []Record
	for _, row := range rows {
		sorted := make([]string, numOfColumns)
		for i, col := range row {
			columnID := m[i]
			sorted[columnID] = col
		}
		ret = append(ret, sorted)
	}
	return ret, nil
}
