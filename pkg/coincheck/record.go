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

import (
	"regexp"
	"strings"
	"time"

	"github.com/ericlagergren/decimal"
)

const recordTimeFormat = "2006-01-02 15:04:05 +0900"

type Record []string

func (r Record) Get(id ColumnID) string {
	return r[int(id)]
}

func (r Record) ID() string {
	return r.Get(ID)
}

func (r Record) Time() (time.Time, error) {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	t := r.Get(Time)
	return time.ParseInLocation(recordTimeFormat, t, jst)
}

func (r Record) Operation() string {
	return r.Get(Operation)
}

func (r Record) Amount() *decimal.Big {
	return r.parseNullDecimal(Amount)
}

func (r Record) TradingCurrency() string {
	return r.Get(TradingCurrency)
}

func (r Record) Price() *decimal.Big {
	return r.parseNullDecimal(Price)
}

func (r Record) OriginalCurrency() string {
	return r.Get(OriginalCurrency)
}

func (r Record) Fee() *decimal.Big {
	return r.parseNullDecimal(Fee)
}

func (r Record) Comment() string {
	return r.Get(Comment)
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

var limitOrderRe = regexp.MustCompile(`Rate: ([0-9]+\.[0-9]+), Pair: ([0-9a-z]+)_([0-9a-z]+)`)

func parseLimitOrder(comment string) (rate *decimal.Big, quote, base string, ok bool) {
	matched := limitOrderRe.FindAllStringSubmatch(comment, -1)
	if len(matched) == 1 {
		m := matched[0]
		rate, ok = new(decimal.Big).SetString(m[1])
		quote = strings.ToUpper(m[2])
		base = strings.ToUpper(m[3])
	}
	return
}

func parseCompletedTradingContracts(comment string) (rate *decimal.Big, quote, base string, ok bool) {
	return parseLimitOrder(comment)
}

var sentRe = regexp.MustCompile("Address: ([0-9a-zA-Z]+)")

func parseSent(comment string) (address string, ok bool) {
	matched := sentRe.FindAllStringSubmatch(comment, -1)
	if len(matched) == 1 {
		m := matched[0]
		address = strings.ToUpper(m[1])
		ok = true
	}
	return
}
