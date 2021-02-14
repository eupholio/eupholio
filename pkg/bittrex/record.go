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

import (
	"strconv"
	"strings"
	"time"

	"github.com/ericlagergren/decimal"
)

const recordTimeFormat = "1/2/2006 3:04:05 PM"

type Record []string

func (r Record) Get(id ColumnID) string {
	return r[int(id)]
}

func (r Record) UUID() string {
	return r.Get(UUID)
}

func (r Record) Exchange() string {
	return r.Get(Exchange)
}

func (r Record) TimeStamp() (time.Time, error) {
	t := r.Get(TimeStamp)
	return time.Parse(recordTimeFormat, t)
}

func (r Record) OrderType() int {
	switch r.Get(OrderType) {
	case LimitSell:
		return OrderTypeLimitSell
	case LimitBuy:
		return OrderTypeLimitBuy
	}
	return OrderTypeUnknown
}

func (r Record) Limit() *decimal.Big {
	return r.parseNullDecimal(Limit)
}

func (r Record) Quantity() *decimal.Big {
	return r.parseNullDecimal(Quantity)
}

func (r Record) QuantityRemaining() *decimal.Big {
	return r.parseNullDecimal(Quantity)
}

func (r Record) Commission() *decimal.Big {
	return r.parseNullDecimal(Commission)
}

func (r Record) Price() *decimal.Big {
	return r.parseNullDecimal(Price)
}

func (r Record) PricePerUnit() *decimal.Big {
	return r.parseNullDecimal(PricePerUnit)
}

func (r Record) IsConditional() bool {
	c := r.Get(IsConditional)
	return c == "True"
}

func (r Record) Condition() string {
	return r.Get(Condition)
}

func (r Record) ConditionTarget() *decimal.Big {
	return r.parseNullDecimal(ConditionTarget)
}

func (r Record) ImmediateOrCancel() bool {
	c := r.Get(ImmediateOrCancel)
	return c == "True"
}

func (r Record) Closed() (time.Time, error) {
	t := r.Get(Closed)
	return time.Parse(recordTimeFormat, t)
}

func (r Record) TimeInForceTypeID() (int, error) {
	return strconv.Atoi(r.Get(TimeInForceTypeID))
}

func (r Record) TimeInForce() string {
	return r.Get(TimeInForce)
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

func MakeRecords(head []string, rows [][]string) ([]Record, error) {
	m := make(map[int]ColumnID)
	for i, k := range head {
		columnID := columnNamesSet[k]
		m[i] = columnID
	}
	var ret []Record
	for _, row := range rows {
		sorted := make([]string, len(columnNames))
		for i, col := range row {
			columnID := m[i]
			sorted[columnID] = col
		}
		ret = append(ret, sorted)
	}
	return ret, nil
}
