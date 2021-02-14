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
	"testing"
)

func TestExtract(t *testing.T) {
	reader := strings.NewReader(testCsv)
	trh, err := Extract(reader)
	if err != nil {
		t.Error(err)
	}
	if len(trh.Transactions) == 0 {
		t.Error("no transaction")
	}
}

func TestExtractRecords(t *testing.T) {
	reader := strings.NewReader(testCsv)

	rs, err := extractRecords(reader)
	if err != nil {
		t.Error(err)
	}
	if len(rs) == 0 {
		t.Error("no record")
	}
}

var testCsv = `"取引日時","通貨","取引種別","取引価格","通貨1","通貨1数量","手数料","通貨1の対円レート","通貨2","通貨2数量","自己・媒介","注文 ID","備考"
"2017/08/16 23:46:37","BTC/JPY","買い","454,359","BTC","0.009","-0.0000135","454,359","JPY","-4,089","媒介","JOR20170816-000006-000001",""
"2017/07/24 14:07:52","JPY","入金","0","JPY","100,000","0","0","","0","","MDP20170724-000002-000001",""
`
