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
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/eupholio"
	"github.com/eupholio/eupholio/pkg/repository"
)

const timeFormat = "2006/01/02 15:04:05"

func QueryTransactions(ctx context.Context, writer io.Writer, tx *sql.Tx, year int, loc *time.Location, baseCurrency, source string, of OutputFormat) error {
	repo := repository.New(tx, currency.Symbol(baseCurrency))

	transactions, err := eupholio.FindEntriesOfTransactions(ctx, repo, year, loc)
	if err != nil {
		return err
	}

	switch of {
	case OutputFormatTable:
		table := tablewriter.NewWriter(writer)
		table.SetHeader([]string{
			"TID",
			"Time " + loc.String(),
			"close",
			"qty",
			"cur",
			"Position",
			"Price",
			baseCurrency,
			"open",
			"qty",
			"cur",
			"Position",
			"Price",
			baseCurrency,
			"REM",
			"Desc",
		})
		for _, t := range transactions {
			id := strconv.Itoa(t.ID)
			tm := t.Time.In(loc).Format(timeFormat)
			var debt [][]string
			var credit [][]string
			rem := shortWalletCode(t.WalletCode)
			desc := t.Description
			for _, e := range t.Entries {
				currency := e.Currency
				qtyStr := e.Quantity.String()
				fa := e.FiatQuantity
				switch baseCurrency {
				case "JPY":
					fa.RoundToInt()
				}
				faStr := fa.String()
				appendDebt := func(s ...string) {
					debt = append(debt, s)
				}
				appendCredit := func(s ...string) {
					credit = append(credit, s)
				}
				poStr := e.Position.String()
				priceStr := ""
				if e.Price.Big != nil {
					priceStr = e.Price.RoundToInt().String()
				}
				switch e.Type {
				case eupholio.EntryTypeOpen:
					appendCredit("OPEN", qtyStr, currency, poStr, priceStr, faStr)
				case eupholio.EntryTypeClose:
					appendDebt("CLOSE", qtyStr, currency, poStr, priceStr, faStr)
				default:
					log.Printf("unknown: %s", e.Type)
				}
			}
			n := len(credit)
			if len(debt) > len(credit) {
				n = len(debt)
			}
			empty := []string{"", "", "", "", "", ""}
			for i := 0; i < n; i++ {
				row := []string{"", ""}
				if i == 0 {
					row = []string{id, tm}
				}
				switch {
				case i < len(debt) && i < len(credit):
					row = append(row, debt[i]...)
					row = append(row, credit[i]...)
				case i < len(debt):
					row = append(row, debt[i]...)
					row = append(row, empty...)
				case i < len(credit):
					row = append(row, empty...)
					row = append(row, credit[i]...)
				}
				if i == 0 {
					row = append(row, rem, desc)
				} else {
					row = append(row, "", "")
				}
				table.Append(row)
			}
			table.SetBorder(true)
		}
		table.Render()
	case OutputFormatCSV:
		table := csv.NewWriter(writer)
		err := table.Write([]string{
			"TID",
			"Time " + loc.String(),
			"close",
			"qty",
			"cur",
			"Position",
			baseCurrency,
			"open",
			"qty",
			"cur",
			"Position",
			baseCurrency,
			"REM",
			"Desc",
		})
		if err != nil {
			return err
		}
		for _, t := range transactions {
			id := strconv.Itoa(t.ID)
			tm := t.Time.In(loc).Format(timeFormat)
			var debt [][]string
			var credit [][]string
			rem := shortWalletCode(t.WalletCode)
			desc := t.Description
			for _, e := range t.Entries {
				currency := e.Currency
				qtyStr := e.Quantity.String()
				fa := e.FiatQuantity
				switch baseCurrency {
				case "JPY":
					fa.RoundToInt()
				}
				faStr := fa.String()
				poStr := e.Position.String()
				appendDebt := func(s ...string) {
					debt = append(debt, s)
				}
				appendCredit := func(s ...string) {
					credit = append(credit, s)
				}
				switch e.Type {
				case eupholio.EntryTypeOpen:
					appendCredit("OPEN", qtyStr, currency, poStr, faStr)
				case eupholio.EntryTypeClose:
					appendDebt("CLOSE", qtyStr, currency, poStr, faStr)
				default:
					log.Printf("unknown: %s", e.Type)
				}
			}
			n := len(credit)
			if len(debt) > len(credit) {
				n = len(debt)
			}
			empty := []string{"", "", "", "", ""}
			for i := 0; i < n; i++ {
				row := []string{"", ""}
				if i == 0 {
					row = []string{id, tm}
				}
				switch {
				case i < len(debt) && i < len(credit):
					row = append(row, debt[i]...)
					row = append(row, credit[i]...)
				case i < len(debt):
					row = append(row, debt[i]...)
					row = append(row, empty...)
				case i < len(credit):
					row = append(row, empty...)
					row = append(row, credit[i]...)
				}
				if i == 0 {
					row = append(row, rem, desc)
				} else {
					row = append(row, "", "")
				}
				err := table.Write(row)
				if err != nil {
					return err
				}
			}
		}
		table.Flush()
	default:
		return fmt.Errorf("unknown output format %s", of)
	}
	return nil
}

func shortWalletCode(walletCode string) (shortCode string) {
	switch walletCode {
	case "BITFLYER":
		shortCode = "BF"
	case "COINCHECK":
		shortCode = "CC"
	case "BITTREX":
		shortCode = "BT"
	case "BITTREX_W":
		shortCode = "BTw"
	case "BITTREX_D":
		shortCode = "BTd"
	case "POLONIEX":
		shortCode = "PO"
	case "POLONIEX_D":
		shortCode = "POd"
	case "POLONIEX_W":
		shortCode = "POw"
	case "POLONIEX_A":
		shortCode = "POa"
	case "CRYPTACT_C":
		shortCode = "CTc"
	default:
		shortCode = walletCode
	}
	return
}
