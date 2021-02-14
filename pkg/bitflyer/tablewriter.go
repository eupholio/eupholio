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
	"io"
	"strconv"

	"github.com/olekukonko/tablewriter"

	"github.com/eupholio/eupholio/models"
)

var idTrTypeMap = map[int]string{
	TrTypeBuy:      "BUY",
	TrTypeSell:     "SELL",
	TrTypeReceive:  "RECEIVE",
	TrTypeDeposit:  "DEPOSIT",
	TrTypeTransfer: "TRANSFER",
	TrTypeFee:      "FEE",
}

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
	case models.BFTransactionSlice:
		t.PrintTransactions(v)
	}
	return nil
}

func (t *TableWriter) PrintTransactionsShort(slice models.BFTransactionSlice) {
	t.writer.SetHeader([]string{
		"Date", "Type", "Price", "C1", "Qty", "Fee", "JPY", "C2", "Qty",
	})
	for _, entry := range slice {
		t.writer.Append([]string{
			entry.TRDate.Format("2006/01/02 15:04:05"),
			idTrTypeMap[entry.TRType],
			entry.TRPrice.String(),
			entry.Currency1,
			entry.Currency1Quantity.String(),
			entry.Fee.String(),
			entry.Currency1JpyRate.String(),
			entry.Currency2.String,
			entry.Currency2Quantity.String(),
		})
	}
	t.writer.Render()
}

func (t *TableWriter) PrintTransactions(slice models.BFTransactionSlice) {
	t.writer.SetHeader([]string{
		"ID", "Date", "Currency", "Type", "Price", "Currency1", "Qty", "Fee", "JPY", "Currency2", "Qty", "Deal", "OrderID", "Remarks",
	})
	for _, entry := range slice {
		t.writer.Append([]string{
			strconv.Itoa(entry.ID),
			entry.TRDate.String(),
			entry.Currency,
			strconv.Itoa(entry.TRType),
			entry.TRPrice.String(),
			entry.Currency1,
			entry.Currency1Quantity.String(),
			entry.Fee.String(),
			entry.Currency1JpyRate.String(),
			entry.Currency2.String,
			entry.Currency2Quantity.String(),
			strconv.Itoa(entry.DealType.Int),
			entry.OrderID,
			entry.Remarks.String,
		})
	}
	t.writer.Render()
}
