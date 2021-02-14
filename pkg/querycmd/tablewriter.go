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

package querycmd

import (
	"io"
	"sort"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/ericlagergren/decimal"

	"github.com/eupholio/eupholio/models"
)

type TableWriter struct {
	writer *tablewriter.Table
}

func NewTableWriter(w io.Writer) *TableWriter {
	return &TableWriter{
		writer: tablewriter.NewWriter(w),
	}
}

func (t *TableWriter) Write(value interface{}) error {
	switch v := value.(type) {
	case models.EventSlice:
		t.PrintEvents(v)
	case models.BalanceSlice:
		t.PrintBalances(v)
	}
	return nil
}

func (t *TableWriter) PrintEvents(es models.EventSlice) {
	t.writer.SetHeader([]string{
		"Year", "Currency", "Type", "Quantity", "Fiat", "TID",
	})
	for _, e := range es {
		if e.Quantity.Sign() == 0 {
			continue
		}
		t.writer.Append([]string{
			e.Time.String(),
			e.Currency,
			e.Type,
			e.Quantity.String(),
			e.BaseQuantity.RoundToInt().String() + " " + e.BaseCurrency,
			strconv.Itoa(e.TransactionID),
		})
	}
	t.writer.Render()
}

func (t *TableWriter) PrintBalances(bs models.BalanceSlice) {
	t.writer.SetHeader([]string{
		"Year", "Currency", "Beginning", "Open qty", "Close qty", "Quantity", "Price", "Profit",
	})
	profit := decimal.New(0, 0)
	sort.SliceStable(bs, func(i, j int) bool {
		return bs[i].Currency < bs[j].Currency
	})
	for _, b := range bs {
		if b.Quantity.Sign() == 0 && b.Profit.Sign() == 0 {
			continue
		}
		t.writer.Append([]string{
			strconv.Itoa(b.Year),
			b.Currency,
			b.BeginningQuantity.Round(8).String(),
			b.OpenQuantity.Round(8).String(),
			b.CloseQuantity.Round(8).String(),
			b.Quantity.Round(8).String(),
			b.Price.Round(8).String(),
			b.Profit.Round(8).String(),
		})
		switch b.Currency {
		case "JPY":
		default:
			profit.Add(profit, b.Profit.Big)
		}
	}
	t.writer.SetFooter([]string{
		"Total", "", "", "", "", "", "", profit.String(),
	})
	t.writer.Render()
}
